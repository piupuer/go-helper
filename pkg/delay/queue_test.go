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
		// add cron tasks
		err := qu.Cron(
			WithQueueTaskUuid("order1"),
			WithQueueTaskName("task1"),
			WithQueueTaskExpr("@every 5s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order2"),
			WithQueueTaskName("task2"),
			WithQueueTaskExpr("@every 10s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order3"),
			WithQueueTaskName("task3"),
			WithQueueTaskExpr("@every 15s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order4"),
			WithQueueTaskName("task4"),
			WithQueueTaskExpr("@every 20s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order5"),
			WithQueueTaskName("task5"),
			WithQueueTaskExpr("@every 40s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order6"),
			WithQueueTaskName("task6"),
			WithQueueTaskExpr("@every 80s"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order7"),
			WithQueueTaskName("task7"),
			WithQueueTaskExpr("@every 100m"),
		)
		err = qu.Cron(
			WithQueueTaskUuid("order8"),
			WithQueueTaskName("task8"),
			WithQueueTaskExpr("0 0 28,29,30,31 * ?"),
		)
		fmt.Println(err)
		i++
	}

	go func() {
		qu := NewQueue()
		// add once task
		qu.Once(
			WithQueueTaskUuid("once.order"),
			WithQueueTaskName("once.task"),
			WithQueueTaskAt(time.Now().Add(time.Duration(240)*time.Hour)),
		)
	}()
	go func() {
		time.Sleep(5 * time.Minute)
		// remove task
		qu := NewQueue()
		qu.Remove("once.order")
	}()

	ch := make(chan int)
	<-ch
}
