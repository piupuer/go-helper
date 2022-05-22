package delay

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-module/carbon/v2"
	"github.com/hibiken/asynq"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"net"
	"net/http"
	"strings"
	"time"
)

type Queue struct {
	ops       QueueOptions
	redis     redis.UniversalClient
	redisOpt  asynq.RedisConnOpt
	lock      nxLock
	client    *asynq.Client
	inspector *asynq.Inspector
	Error     error
}

type periodTask struct {
	Expr      string `json:"expr"` // cron expr github.com/robfig/cron/v3
	Name      string `json:"name"`
	Uid       string `json:"uid"`
	Payload   string `json:"payload"`
	Next      int64  `json:"next"`      // next schedule unix timestamp
	Processed int64  `json:"processed"` // run times
}

type periodTaskHandler struct {
	qu Queue
}

type Task struct {
	Name    string `json:"name"`
	Uid     string `json:"uid"`
	Payload string `json:"payload"`
}

func (p periodTaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	ctx = tracing.NewId(ctx)
	task := Task{
		Name:    t.Type(),
		Uid:     t.ResultWriter().TaskID(),
		Payload: string(t.Payload()),
	}
	if p.qu.ops.handler != nil {
		err = p.qu.ops.handler(ctx, task)
	} else if p.qu.ops.callback != "" {
		err = p.httpCallback(ctx, task)
	} else {
		log.
			WithContext(ctx).
			WithFields(map[string]interface{}{
				"Task": utils.Struct2Json(task),
			}).
			Info("no task handler")
	}
	// save processed count
	p.qu.processed(task.Uid)
	return
}

func (p periodTaskHandler) httpCallback(ctx context.Context, task Task) (err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	body := utils.Struct2Json(task)
	var r *http.Request
	r, _ = http.NewRequest(http.MethodPost, p.qu.ops.callback, bytes.NewReader([]byte(body)))
	r.Header.Add("Content-Type", gin.MIMEJSON)
	var res *http.Response
	res, err = client.Do(r)
	if e, ok := err.(net.Error); ok && e.Timeout() {
		log.
			WithContext(ctx).
			WithFields(map[string]interface{}{
				"Task": body,
			}).
			WithError(err).
			Error(ErrHttpCallbackTimeout)
		err = ErrHttpCallbackTimeout
		return
	}
	if err != nil {
		log.
			WithContext(ctx).
			WithFields(map[string]interface{}{
				"Task": body,
			}).
			WithError(err).
			Error(ErrHttpCallback)
		err = ErrHttpCallback
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.
			WithContext(ctx).
			WithFields(map[string]interface{}{
				"Task":       body,
				"StatusCode": res.StatusCode,
			}).
			Error(ErrHttpCallbackInvalidStatusCode)
		err = ErrHttpCallbackInvalidStatusCode
	}
	return
}

type nxLock struct {
	key   string
	redis redis.UniversalClient
}

func (n nxLock) Lock() (ok bool) {
	ok, _ = n.redis.SetNX(context.Background(), n.key, true, 10*time.Second).Result()
	return
}

func (n nxLock) Unlock() {
	n.redis.Del(context.Background(), n.key)
	return
}

// NewQueue delay queue implemented by asynq: https://github.com/hibiken/asynq
func NewQueue(options ...func(*QueueOptions)) (qu *Queue) {
	ops := getQueueOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	qu = &Queue{}
	if ops.redisUri == "" {
		qu.Error = errors.WithStack(ErrRedisNil)
		return
	}
	rs, err := asynq.ParseRedisURI(ops.redisUri)
	if err != nil {
		qu.Error = errors.WithStack(ErrRedisInvalid)
		return
	}
	rd := rs.MakeRedisClient().(redis.UniversalClient)
	client := asynq.NewClient(rs)
	inspector := asynq.NewInspector(rs)
	// initialize redis lock
	lock := nxLock{
		key:   ops.redisPeriodKey + ".lock",
		redis: rd,
	}
	// initialize server
	srv := asynq.NewServer(
		rs,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				ops.name: 10,
			},
		},
	)
	go func() {
		var h periodTaskHandler
		h.qu = *qu
		if e := srv.Run(h); e != nil {
			log.WithError(err).Error("run task handler failed")
		}
	}()
	qu.ops = *ops
	qu.redis = rd
	qu.redisOpt = rs
	qu.lock = lock
	qu.client = client
	qu.inspector = inspector
	// initialize scanner
	go func() {
		for {
			time.Sleep(time.Second)
			qu.scan()
		}
	}()
	if qu.ops.clearArchived > 0 {
		// initialize clear archived
		go func() {
			for {
				time.Sleep(time.Duration(qu.ops.clearArchived) * time.Second)
				qu.clearArchived()
			}
		}()
	}
	return
}

func (qu Queue) Once(options ...func(*QueueTaskOptions)) (err error) {
	ops := getQueueTaskOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.WithStack(ErrUuidNil)
		return
	}
	t := asynq.NewTask(ops.name+".once", []byte(ops.payload), asynq.TaskID(ops.uid))
	taskOpts := []asynq.Option{
		asynq.Queue(qu.ops.name),
		asynq.MaxRetry(qu.ops.maxRetry),
	}
	if ops.retention > 0 {
		taskOpts = append(taskOpts, asynq.Retention(time.Duration(ops.retention)*time.Second))
	} else {
		taskOpts = append(taskOpts, asynq.Retention(time.Duration(qu.ops.retention)*time.Second))
	}
	if ops.in != nil {
		taskOpts = append(taskOpts, asynq.ProcessIn(*ops.in))
	} else if ops.at != nil {
		taskOpts = append(taskOpts, asynq.ProcessAt(*ops.at))
	} else if ops.now {
		taskOpts = append(taskOpts, asynq.ProcessIn(time.Second))
	}
	_, err = qu.client.Enqueue(t, taskOpts...)
	return
}

func (qu Queue) Cron(options ...func(*QueueTaskOptions)) (err error) {
	ops := getQueueTaskOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.uid == "" {
		err = errors.WithStack(ErrUuidNil)
		return
	}
	var next int64
	next, err = getNext(ops.expr, 0)
	if err != nil {
		err = errors.WithStack(ErrExprInvalid)
		return
	}
	t := periodTask{
		Expr:    ops.expr,
		Name:    ops.name + ".cron",
		Uid:     ops.uid,
		Payload: ops.payload,
		Next:    next,
	}
	_, err = qu.redis.HSet(context.Background(), qu.ops.redisPeriodKey, ops.uid, utils.Struct2Json(t)).Result()
	if err != nil {
		err = errors.WithStack(ErrSaveCron)
		return
	}
	return
}

func (qu Queue) Remove(uid string) (err error) {
	var ok bool
	for {
		ok = qu.lock.Lock()
		if ok {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	defer qu.lock.Unlock()
	qu.redis.HDel(context.Background(), qu.ops.redisPeriodKey, uid)

	err = qu.inspector.DeleteTask(qu.ops.name, uid)
	return
}

func (qu Queue) processed(uid string) {
	var ok bool
	for {
		ok = qu.lock.Lock()
		if ok {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	defer qu.lock.Unlock()
	ctx := context.Background()
	t, e := qu.redis.HGet(ctx, qu.ops.redisPeriodKey, uid).Result()
	if e == nil || e != redis.Nil {
		var item periodTask
		utils.Json2Struct(t, &item)
		item.Processed++
		qu.redis.HSet(ctx, qu.ops.redisPeriodKey, uid, utils.Struct2Json(item))
	}
	return
}

func (qu Queue) scan() {
	ctx := context.Background()
	ok := qu.lock.Lock()
	if !ok {
		return
	}
	defer qu.lock.Unlock()
	m, _ := qu.redis.HGetAll(ctx, qu.ops.redisPeriodKey).Result()
	p := qu.redis.Pipeline()
	ops := qu.ops
	for _, v := range m {
		var item periodTask
		utils.Json2Struct(v, &item)
		next, _ := getNext(item.Expr, item.Next)
		t := asynq.NewTask(item.Name, []byte(item.Payload), asynq.TaskID(item.Uid))
		taskOpts := []asynq.Option{
			asynq.Queue(ops.name),
			asynq.MaxRetry(ops.maxRetry),
		}
		diff := next - item.Next
		if diff > 10 {
			retention := diff / 3
			if diff > 600 {
				// max retention 10min
				retention = 600
			}
			// set retention avoid repeat in short time
			taskOpts = append(taskOpts, asynq.Retention(time.Duration(retention)*time.Second))
		}
		taskOpts = append(taskOpts, asynq.ProcessAt(time.Unix(item.Next, 0)))
		_, err := qu.client.Enqueue(t, taskOpts...)
		// enqueue success, update next
		if err == nil {
			item.Next = next
			p.HSet(ctx, qu.ops.redisPeriodKey, item.Uid, utils.Struct2Json(item))
		}
	}
	// batch save to cache
	p.Exec(ctx)
	return
}

func (qu Queue) clearArchived() {
	list, err := qu.inspector.ListArchivedTasks(qu.ops.name, asynq.Page(1), asynq.PageSize(100))
	if err != nil {
		return
	}
	ctx := context.Background()
	for _, item := range list {
		last := carbon.Time2Carbon(item.LastFailedAt)
		if !last.IsZero() && item.Retried < item.MaxRetry {
			continue
		}
		uid := item.ID
		var flag bool
		if strings.HasSuffix(item.Type, ".cron") {
			// cron task
			t, e := qu.redis.HGet(ctx, qu.ops.redisPeriodKey, uid).Result()
			if e == nil || e != redis.Nil {
				var task periodTask
				utils.Json2Struct(t, &task)
				next, _ := getNext(task.Expr, task.Next)
				diff := next - task.Next
				if diff <= 60 {
					if carbon.Now().Gt(last.AddMinutes(5)) {
						flag = true
					}
				} else if diff <= 600 {
					if carbon.Now().Gt(last.AddMinutes(30)) {
						flag = true
					}
				} else if diff <= 3600 {
					if carbon.Now().Gt(last.AddHours(2)) {
						flag = true
					}
				} else {
					if carbon.Now().Gt(last.AddHours(5)) {
						flag = true
					}
				}
			}
		} else {
			// once task, has failed for more than 5 minutes
			if carbon.Now().Gt(last.AddMinutes(5)) {
				flag = true
			}
		}
		if flag {
			qu.inspector.DeleteTask(qu.ops.name, uid)
		}
	}
}

func getNext(expr string, timestamp int64) (next int64, err error) {
	var schedule cron.Schedule
	schedule, err = cron.ParseStandard(expr)
	if err != nil {
		return
	}
	t := time.Now()
	if timestamp > 0 {
		t = time.Unix(timestamp, 0)
	}
	next = schedule.Next(t).Unix()
	return
}
