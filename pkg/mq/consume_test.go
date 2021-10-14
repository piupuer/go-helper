package mq

import (
	"context"
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

	err := qu.Consume(
		handler,
		WithConsumeAutoRequestId,
	)
	if err != nil {
		panic(err)
	}

	ch := make(chan int)
	<-ch
}

func handler(ctx context.Context, q string, delivery amqp.Delivery) bool {
	fmt.Println(ctx, q, delivery.Exchange)
	return true
}

func TestQueue_ConsumeOne(t *testing.T) {
	rb := NewRabbit(uri)
	if rb.Error != nil {
		panic(rb.Error)
	}
	for {
		time.Sleep(10 * time.Second)
		err := rb.
			Exchange(
				WithExchangeName("ex1"),
			).Queue(
			WithQueueName("q1"),
			WithQueueSkipDeclare,
			WithQueueSkipBind,
		).ConsumeOne(
			handler,
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}
