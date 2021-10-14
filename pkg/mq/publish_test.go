package mq

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/examples"
	"github.com/piupuer/go-helper/pkg/logger"
	"testing"
	"time"
)

func TestExchange_PublishProto(t *testing.T) {
	rb := NewRabbit(
		uri,
		WithReconnectMaxRetryCount(3),
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

	for {
		time.Sleep(100 * time.Millisecond)
		go func() {
			ctx := context.WithValue(context.Background(), logger.RequestIdContextKey, "send-1")
			var mqPb examples.Msg
			mqPb.Name = "hello1"
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt1"),
				WithPublishRouteKey("rt2"),
				WithPublishContext(ctx),
			)
			fmt.Println(time.Now(), "send 1 end", err)
		}()
		go func() {
			ctx := context.WithValue(context.Background(), logger.RequestIdContextKey, "send-2")
			var mqPb examples.Msg
			mqPb.Name = "hello2"
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt2"),
				WithPublishContext(ctx),
			)
			fmt.Println(time.Now(), "send 2 end", err)
		}()
		go func() {
			ctx := context.WithValue(context.Background(), logger.RequestIdContextKey, "send-3")
			var mqPb examples.Msg
			mqPb.Name = "hello3"
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt2"),
				WithPublishContext(ctx),
			)
			fmt.Println(time.Now(), "send 3 end", err)
		}()
		fmt.Println()
	}

}

func TestExchange_PublishProto2(t *testing.T) {
	rb := NewRabbit(
		uri,
		WithReconnectMaxRetryCount(3),
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

	for {
		time.Sleep(time.Second)
		var mqPb examples.Msg
		mqPb.Name = "hello1"
		err := ex.PublishProto(
			&mqPb,
			WithPublishRouteKey("rt1"),
		)
		fmt.Println(time.Now(), "send end", err)
	}
}
