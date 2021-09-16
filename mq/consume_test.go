package mq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"testing"
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

func handler(ctx context.Context,q string, delivery amqp.Delivery) bool {
	fmt.Println(ctx, q, delivery.Exchange)
	return true
}
