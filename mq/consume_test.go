package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestQueue_Consume(t *testing.T) {
	ex := NewRabbit(uri).
		Exchange(
			WithExchangeName("ex1"),
		)
	if ex.Error != nil {
		panic(ex.Error)
	}
	qu := ex.Queue(
		WithQueueName("q1"),
		WithQueueSkipDeclare,
		WithQueueSkipBind,
	)
	if qu.Error != nil {
		panic(qu.Error)
	}

	err := qu.Consume(handler)
	if err != nil {
		panic(err)
	}

	ch := make(chan int)
	<-ch
}

func handler(q string, delivery amqp.Delivery) bool {
	fmt.Println(time.Now(), q, delivery.Exchange)
	return true
}
