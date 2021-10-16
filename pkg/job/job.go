package job

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/libi/dcron"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
)

const (
	dcronInfoPrefix  = "INFO: "
	dcronErrorPrefix = "ERR: "
)

type Config struct {
	RedisUri    string
	RedisClient redis.UniversalClient
}

type GoodJob struct {
	lock        sync.Mutex
	redis       redis.UniversalClient
	driver      *RedisClientDriver
	tasks       map[string]GoodDistributeTask
	single      bool
	singleTasks map[string]GoodSingleTask
	ops         Options
	Error       error
}

type GoodTask struct {
	running bool
	Name    string
	Expr    string
	Func    func(ctx context.Context) error
}

type GoodSingleTask struct {
	GoodTask
	c *cron.Cron
}

type GoodDistributeTask struct {
	GoodTask
	c *dcron.Dcron
}

func New(cfg Config, options ...func(*Options)) (*GoodJob, error) {
	// init fields
	job := GoodJob{}
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	job.ops = *ops
	if cfg.RedisClient != nil {
		job.redis = cfg.RedisClient
	} else {
		if cfg.RedisUri == "" {
			cfg.RedisUri = "redis://127.0.0.1:6379/0"
		}
		r, err := ParseRedisURI(cfg.RedisUri)
		if err != nil {
			return nil, err
		}
		job.redis = r
	}

	_, err := job.redis.Ping(context.Background()).Result()
	if err != nil {
		job.single = true
		job.singleTasks = make(map[string]GoodSingleTask, 0)
		job.ops.logger.Warn(job.ops.ctx, "initialize redis failed, switch singe mode, err: %v", err)
		return &job, nil
	}

	drv, err := NewDriver(
		job.redis,
		WithDriverLogger(job.ops.logger),
		WithDriverContext(job.ops.ctx),
		WithDriverPrefix(job.ops.prefix),
	)
	if err != nil {
		return nil, err
	}
	job.driver = drv
	job.tasks = make(map[string]GoodDistributeTask, 0)
	return &job, nil
}

func ParseRedisURI(uri string) (redis.UniversalClient, error) {
	var opt asynq.RedisConnOpt
	var err error
	if uri != "" {
		opt, err = asynq.ParseRedisURI(uri)
		if err != nil {
			return nil, err
		}
		return opt.MakeRedisClient().(redis.UniversalClient), nil
	}
	return nil, fmt.Errorf("invalid redis config")
}

func (g *GoodJob) AddTask(task GoodTask) *GoodJob {
	if g.Error != nil {
		return g
	}
	if g.single {
		return g.addSingleTask(task)
	}
	return g.addDistributeTask(task)
}

func (g *GoodJob) addSingleTask(task GoodTask) *GoodJob {
	if g.Error != nil {
		return g
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if _, ok := g.singleTasks[task.Name]; ok {
		g.ops.logger.Warn(g.ops.ctx, "task %s already exists, skip", task.Name)
		return g
	}

	c := cron.New()
	j := job{
		AutoRequestId: g.ops.AutoRequestId,
		Func:          task.Func,
	}
	c.AddJob(task.Expr, j)

	t := GoodSingleTask{
		GoodTask: task,
		c:        c,
	}
	g.singleTasks[task.Name] = t
	return g
}

func (g *GoodJob) addDistributeTask(task GoodTask) *GoodJob {
	if g.Error != nil {
		return g
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if _, ok := g.tasks[task.Name]; ok {
		g.ops.logger.Warn(g.ops.ctx, "task %s already exists, skip", task.Name)
		return g
	}

	c := dcron.NewDcronWithOption(
		task.Name,
		g.driver,
		dcron.WithLogger(&cronLogger{
			g.ops.logger,
		}),
	)
	fun := (func(task GoodTask) func() {
		return func() {
			ctx := context.Background()
			if g.ops.AutoRequestId {
				ctx = context.WithValue(ctx, logger.RequestIdContextKey, uuid.NewV4().String())
			}
			task.Func(ctx)
		}
	})(task)
	c.AddFunc(task.Name, task.Expr, fun)
	t := GoodDistributeTask{
		GoodTask: task,
		c:        c,
	}
	g.tasks[task.Name] = t
	return g
}

func (g *GoodJob) Start() {
	if g.Error != nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.single {
		for _, task := range g.singleTasks {
			if !task.running {
				task.c.Start()
				task.running = true
				g.singleTasks[task.Name] = task
			}
		}
	} else {
		for _, task := range g.tasks {
			if !task.running {
				task.c.Start()
				task.running = true
				g.tasks[task.Name] = task
			}
		}
	}
}

// stop all task in current node(task still running in other node)
func (g *GoodJob) StopAll() {
	if g.Error != nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.single {
		for _, task := range g.singleTasks {
			if task.running {
				task.c.Stop()
				task.running = false
				g.singleTasks[task.Name] = task
			}
		}
	} else {
		for _, task := range g.tasks {
			if task.running {
				task.c.Stop()
				task.running = false
				g.tasks[task.Name] = task
			}
		}
	}
}

// stop task in current node(task still running in other node)
func (g *GoodJob) Stop(taskName string) {
	if g.Error != nil {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.single {
		for _, task := range g.singleTasks {
			if task.Name == taskName {
				if task.running {
					task.c.Stop()
					task.running = false
					g.singleTasks[task.Name] = task
					delete(g.singleTasks, taskName)
					break
				} else {
					g.ops.logger.Warn(g.ops.ctx, "task %s is not running, skip", task.Name)
				}
			}
		}
	} else {
		for _, task := range g.tasks {
			if task.Name == taskName {
				if task.running {
					task.c.Stop()
					task.running = false
					g.tasks[task.Name] = task
					delete(g.tasks, taskName)
					break
				} else {
					g.ops.logger.Warn(g.ops.ctx, "task %s is not running, skip", task.Name)
				}
			}
		}
	}
}

type cronLogger struct {
	l logger.Interface
}

func (c cronLogger) Printf(format string, args ...interface{}) {
	ctx := context.Background()
	if strings.HasPrefix(format, dcronInfoPrefix) {
		c.l.Info(ctx, strings.TrimPrefix(format, dcronInfoPrefix), args...)
	} else if strings.HasPrefix(format, dcronErrorPrefix) {
		c.l.Error(ctx, strings.TrimPrefix(format, dcronErrorPrefix), args...)
	}
}

type job struct {
	AutoRequestId bool
	Func          func(ctx context.Context) error
}

func (j job) Run() {
	if j.Func != nil {
		ctx := context.Background()
		if j.AutoRequestId {
			ctx = context.WithValue(ctx, logger.RequestIdContextKey, uuid.NewV4().String())
		}
		j.Func(ctx)
	}
}
