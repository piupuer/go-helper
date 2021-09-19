package job

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func Run(ctx context.Context) error {
	fmt.Println(time.Now(), "running context: ", ctx)
	http.Get(fmt.Sprintf("http://127.0.0.1/api/ping?key=%d&pid=%d", time.Now().Unix(), os.Getpid()))
	return nil
}

var uri1 = "redis://:123456@127.0.0.1:6379"
var uri2 = "redis-sentinel://:123456@127.0.0.1:6179,127.0.0.1:6180,127.0.0.1:6181?master=prod"

func TestNew(t *testing.T) {
	job, err := New(Config{
		RedisUri: uri2,
	})
	if err != nil {
		panic(err)
	}
	job.AddTask(GoodTask{
		Name:    "work",
		Expr:    "@every 10s",
		Func: func(ctx context.Context) error {
			return Run(ctx)
		},
	}).Start()

	time.Sleep(30 * time.Second)
	job.AddTask(GoodTask{
		Name:    "work2",
		Expr:    "@every 5s",
		Func: func(ctx context.Context) error {
			return Run(ctx)
		},
	}).Start()

	time.Sleep(15 * time.Second)
	job.Stop("work")

	ch := make(chan int, 0)
	<-ch
}
