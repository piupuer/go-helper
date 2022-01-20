package mq

import (
	"context"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
)

type Consume struct {
	qu    *Queue
	ops   ConsumeOptions
	Error error
}

func (qu *Queue) Consume(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) error {
	if handler == nil {
		return errors.Errorf("handler is nil")
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		return errors.WithStack(co.Error)
	}
	ctx := co.newContext(nil)
	delivery, err := co.consume(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	go func() {
		for {
			go func() {
				for msg := range delivery {
					if co.ops.autoAck {
						handler(ctx, co.qu.ops.name, msg)
						continue
					}
					if handler(ctx, co.qu.ops.name, msg) {
						err := msg.Ack(false)
						if err != nil {
							logger.WithRequestId(ctx).Error("consume ack err: %+v", errors.WithStack(err))
						}
					} else {
						err := msg.Nack(false, co.ops.nackRequeue)
						if err != nil {
							logger.WithRequestId(ctx).Error("consume nack err: %+v", errors.WithStack(err))
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
					} else {
						time.Sleep(time.Duration(co.qu.ex.rb.ops.reconnectInterval) * time.Second)
					}
				}
				if co.ops.newRequestIdWhenConnectionLost {
					ctx = co.newContext(nil)
				}
				d, err := co.consume(ctx)
				if err != nil {
					logger.WithRequestId(ctx).Error("reconsume err: %+v", errors.WithStack(err))
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
		return errors.Errorf("handler is nil")
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		return errors.WithStack(co.Error)
	}
	ctx := co.newContext(co.ops.oneCtx)
	msg, ok, err := co.consumeOne(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	if !ok {
		return errors.Errorf("queue is empty, can't get one msg")
	}
	if co.ops.autoAck {
		handler(ctx, co.qu.ops.name, msg)
		return nil
	}
	if handler(ctx, co.qu.ops.name, msg) {
		err := msg.Ack(false)
		if err != nil {
			logger.WithRequestId(ctx).Error("consume ack err: %+v", err)
		}
	} else {
		err := msg.Nack(false, co.ops.nackRequeue)
		if err != nil {
			logger.WithRequestId(ctx).Error("consume nack err: %+v", err)
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
		return nil, errors.WithStack(err)
	}
	// set channel qos
	err = channel.Qos(
		co.ops.qosPrefetchCount,
		co.ops.qosPrefetchSize,
		co.ops.qosGlobal,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return channel.Consume(
		co.qu.ops.name,
		co.ops.consumer,
		co.ops.autoAck,
		co.ops.exclusive,
		co.ops.noLocal,
		co.ops.noWait,
		co.ops.args,
	)
}

func (co *Consume) consumeOne(ctx context.Context) (amqp.Delivery, bool, error) {
	channel, err := co.qu.ex.rb.getChannel(ctx)
	var msg amqp.Delivery
	if err != nil {
		return msg, false, errors.WithStack(err)
	}

	return channel.Get(
		co.qu.ops.name,
		co.ops.autoAck,
	)
}

func (co *Consume) newContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if co.ops.autoRequestId {
		ctx = context.WithValue(ctx, constant.MiddlewareRequestIdCtxKey, uuid.NewString())
	}
	return ctx
}
