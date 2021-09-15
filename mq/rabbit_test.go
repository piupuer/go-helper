package mq

import (
	"testing"
)

const uri = "amqp://admin:admin@127.0.0.1:5672/gsgc"

func TestNewRabbitMq(t *testing.T) {
	rb := NewRabbit(uri)
	if rb.Error != nil {
		panic(rb.Error)
	}
	ch := rb.Channel(
		WithChannelQosPrefetchCount(5),
	)
	if ch.Error != nil {
		panic(ch.Error)
	}

	ex := ch.Exchange(
		WithExchangeName("ex1"),
	)
	if ex.Error != nil {
		panic(ex.Error)
	}
	err := ex.QueueWithDeadLetter(
		WithQueueName("q1"),
		WithQueueRouteKey("rt1"),
		WithQueueDeadLetterName("dl-ex"),
		WithQueueDeadLetterKey("dlr"),
		WithQueueMessageTTL(30000),
	).Error
	if err != nil {
		panic(err)
	}
	
	err = ex.QueueWithDeadLetter(
		WithQueueName("q2"),
		WithQueueRouteKey("rt2"),
		WithQueueDeadLetterName("dl-ex"),
		WithQueueDeadLetterKey("dlr"),
		WithQueueMessageTTL(30000),
	).Error
	if err != nil {
		panic(err)
	}

	err = ch.Exchange(
		WithExchangeName("ex2"),
	).Queue(
		WithQueueName("q3"),
		WithQueueRouteKey("rt3"),
	).Error
	if err != nil {
		panic(err)
	}

	err = ch.Exchange(
		WithExchangeName("dl-ex"),
	).Queue(
		WithQueueName("dlq"),
		WithQueueRouteKey("dlr"),
	).Error
}
