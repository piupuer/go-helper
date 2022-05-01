package delay

import (
	"fmt"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	i := 0
	for i < 10 {
		qu := NewQueue()
		fmt.Println(qu)
		err := qu.Cron(
			WithQueueTaskUuid("order2"),
			WithQueueTaskName("task2"),
			WithQueueTaskExpr("@every 1s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order3"),
			WithQueueTaskName("task3"),
			WithQueueTaskExpr("@every 1m"),
		)
		fmt.Println(err)
		i++
	}

	go func() {
		time.Sleep(500 * time.Second)
		qu := NewQueue()
		qu.Remove("order2")
	}()

	ch := make(chan int)
	<-ch
}
