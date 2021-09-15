package mq

import (
	"go-helper/examples"
	"testing"
)

func TestExchange_PublishProto(t *testing.T) {
	ch := NewRabbit(uri).Channel(
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
