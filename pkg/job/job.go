package job

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/libi/dcron"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
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
	running             bool
	Name                string
	Expr                string
	SkipIfStillRunning  bool
	DelayIfStillRunning bool
	Func                func(ctx context.Context) error
	Wrappers            []cron.JobWrapper
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
			return nil, errors.WithStack(err)
		}
		job.redis = r
	}

	_, err := job.redis.Ping(context.Background()).Result()
	if err != nil {
		job.single = true
		job.singleTasks = make(map[string]GoodSingleTask, 0)
		log.WithRequestId(job.ops.ctx).Warn("initialize redis failed, switch singe mode, err: %+v", err)
		return &job, nil
	}

	drv, err := NewDriver(
		job.redis,
		WithDriverCtx(job.ops.ctx),
		WithDriverPrefix(job.ops.prefix),
	)
	if err != nil {
		return nil, errors.WithStack(err)
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
	return nil, errors.Errorf("invalid redis config")
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
		log.WithRequestId(g.ops.ctx).Warn("task %s already exists, skip", task.Name)
		return g
	}

	c := cron.New(cron.WithChain(g.parseWrapper(task)...))
	c.AddFunc(task.Expr, g.parseFun(task))

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
		log.WithRequestId(g.ops.ctx).Warn("task %s already exists, skip", task.Name)
		return g
	}

	c := dcron.NewDcronWithOption(
		task.Name,
		g.driver,
		dcron.WithLogger(newDCronLogger(
			WithCronCtx(g.ops.ctx),
		)),
		dcron.CronOptionChain(g.parseWrapper(task)...),
	)
	c.AddFunc(task.Name, task.Expr, g.parseFun(task))
	t := GoodDistributeTask{
		GoodTask: task,
		c:        c,
	}
	g.tasks[task.Name] = t
	return g
}

func (g GoodJob) parseWrapper(task GoodTask) []cron.JobWrapper {
	cronLogger := NewCronLogger(
		WithCronCtx(g.ops.ctx),
	)
	if task.SkipIfStillRunning {
		task.Wrappers = append(task.Wrappers, cron.SkipIfStillRunning(cronLogger))
	}
	if !task.SkipIfStillRunning && task.DelayIfStillRunning {
		task.Wrappers = append(task.Wrappers, cron.DelayIfStillRunning(cronLogger))
	}
	return task.Wrappers
}

func (g GoodJob) parseFun(task GoodTask) func() {
	return (func(task GoodTask) func() {
		return func() {
			ctx := context.Background()
			if g.ops.autoRequestId {
				ctx = query.NewRequestId(ctx)
			}
			ctx = context.WithValue(ctx, g.ops.taskNameCtxKey, task.Name)
			task.Func(ctx)
		}
	})(task)
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
					log.WithRequestId(g.ops.ctx).Warn("task %s is not running, skip", task.Name)
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
					log.WithRequestId(g.ops.ctx).Warn("task %s is not running, skip", task.Name)
				}
			}
		}
	}
}
