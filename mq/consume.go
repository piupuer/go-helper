package mq

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/logger"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type Consume struct {
	qu    *Queue
	ops   ConsumeOptions
	Error error
}

func (qu *Queue) Consume(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) error {
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		return co.Error
	}
	ctx := co.newContext(nil)
	delivery, err := co.consume(ctx)
	if err != nil {
		return err
	}
	go func() {
		for {
			go func() {
				for msg := range delivery {
					if co.ops.AutoAck {
						handler(ctx, co.qu.ops.Name, msg)
						continue
					}
					if handler(ctx, co.qu.ops.Name, msg) {
						err := msg.Ack(false)
						if err != nil {
							co.qu.ex.rb.ops.logger.Error(ctx, "consume ack err: %v", err)
						}
					} else {
						err := msg.Nack(false, co.ops.NackRequeue)
						if err != nil {
							co.qu.ex.rb.ops.logger.Error(ctx, "consume nack err: %v", err)
						}
					}
				}
			}()
			// wait connection lost
			if err = <-co.qu.ex.rb.lostCh; err != nil {
				for {
					err = co.qu.ex.rb.reconnect(ctx)
					if err == nil {
						break
					}
				}
				if co.ops.NewRequestIdWhenConnectionLost {
					ctx = co.newContext(nil)
				}
				d, err := co.consume(ctx)
				if err != nil {
					co.qu.ex.rb.ops.logger.Error(ctx, "reconsume err: %v", err)
					return
				}
				delivery = d
			}
		}
	}()
	return nil
}

func (qu *Queue) ConsumeOne(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) error {
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		return co.Error
	}
	ctx := co.newContext(co.ops.oneCtx)
	msg, ok, err := co.consumeOne(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("queue is empty, can't get one msg")
	}
	if co.ops.AutoAck {
		handler(ctx, co.qu.ops.Name, msg)
		return nil
	}
	if handler(ctx, co.qu.ops.Name, msg) {
		err := msg.Ack(false)
		if err != nil {
			co.qu.ex.rb.ops.logger.Error(ctx, "consume ack err: %v", err)
		}
	} else {
		err := msg.Nack(false, co.ops.NackRequeue)
		if err != nil {
			co.qu.ex.rb.ops.logger.Error(ctx, "consume nack err: %v", err)
		}
	}
	return nil
}

func (qu *Queue) beforeConsume(options ...func(*ConsumeOptions)) *Consume {
	var co Consume
	if qu.Error != nil {
		co.Error = qu.Error
		return &co
	}
	ops := getConsumeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	co.ops = *ops
	co.qu = qu
	return &co
}

func (co *Consume) consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	channel, err := co.qu.ex.rb.getChannel(ctx)
	if err != nil {
		return nil, err
	}
	// 消费者流控, 防止数据库爆库
	// 消息的消费需要配合Qos
	err = channel.Qos(
		// 每次队列只消费一个消息, 这个消息处理不完服务器不会发送第二个消息过来
		// 当前消费者一次能接受的最大消息数量
		co.ops.QosPrefetchCount,
		// 服务器传递的最大容量
		co.ops.QosPrefetchSize,
		// 如果为true 对channel可用 false则只对当前队列可用
		co.ops.QosGlobal,
	)
	if err != nil {
		return nil, err
	}
	return channel.Consume(
		co.qu.ops.Name,
		co.ops.Consumer,
		co.ops.AutoAck,
		co.ops.Exclusive,
		co.ops.NoLocal,
		co.ops.NoWait,
		co.ops.Args,
	)
}

func (co *Consume) consumeOne(ctx context.Context) (amqp.Delivery, bool, error) {
	channel, err := co.qu.ex.rb.getChannel(ctx)
	var msg amqp.Delivery
	if err != nil {
		return msg, false, err
	}

	return channel.Get(
		co.qu.ops.Name,
		co.ops.AutoAck,
	)
}

func (co *Consume) newContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if co.ops.AutoRequestId {
		ctx = context.WithValue(ctx, logger.RequestIdContextKey, uuid.NewV4().String())
	}
	return ctx
}
