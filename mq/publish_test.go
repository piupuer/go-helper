package mq

import (
	"go-helper/examples"
	"testing"
)

func TestExchange_PublishProto(t *testing.T) {
	rb := NewRabbit(
		uri,
		WithChannelQosPrefetchCount(5),
	)
	if rb.Error != nil {
		panic(rb.Error)
	}
	ex := rb.Exchange(
		WithExchangeName("ex1"),
		WithExchangeSkipDeclare,
	)
	if ex.Error != nil {
		panic(ex.Error)
	}

	var mqPb examples.Msg
	mqPb.Name = "hello"
	err := ex.PublishProto(
		&mqPb,
		WithPublishRouteKey("rt1"),
		WithPublishRouteKey("rt2"),
	).Error
	if err != nil {
		panic(err)
	}
}
