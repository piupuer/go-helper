package job

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	"net/http"
	"os"
	"testing"
	"time"
)

func Run(ctx context.Context) error {
	fmt.Println("ctx", ctx.Value(constant.MiddlewareRequestIdCtxKey), ctx.Value(constant.JobTaskNameCtxKey), time.Now(), "start")
	time.Sleep(13 * time.Second)
	http.Get(fmt.Sprintf("http://127.0.0.1/api/ping?key=%d&pid=%d", time.Now().Unix(), os.Getpid()))
	fmt.Println("ctx", ctx.Value(constant.MiddlewareRequestIdCtxKey), ctx.Value(constant.JobTaskNameCtxKey), time.Now(), "end")
	return nil
}

var uri1 = "redis://127.0.0.1:6379"
var uri2 = "redis-sentinel://:123456@127.0.0.1:6179,127.0.0.1:6180,127.0.0.1:6181?master=prod"

func TestNew(t *testing.T) {
	job, err := New(
		Config{
			RedisUri: uri1,
		},
		WithAutoRequestId(true),
	)
	if err != nil {
		panic(err)
	}
	job.AddTask(GoodTask{
		Name: "work",
		Expr: "@every 2s",
		Func: func(ctx context.Context) error {
			return Run(ctx)
		},
		// SkipIfStillRunning: true,
		DelayIfStillRunning: true,
	}).Start()

	time.Sleep(1 * time.Second)
	job.AddTask(GoodTask{
		Name: "work2",
		Expr: "@every 1s",
		Func: func(ctx context.Context) error {
			return Run(ctx)
		},
	}).Start()

	time.Sleep(5 * time.Second)
	job.Stop("work")
}
