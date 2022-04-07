package mq

import (
	"context"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
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
						e := msg.Ack(false)
						if e != nil {
							log.WithRequestId(ctx).WithError(e).Error("consume ack failed")
						}
					} else {
						e := msg.Nack(false, co.ops.nackRequeue)
						if e != nil {
							log.WithRequestId(ctx).WithError(e).Error("consume nack failed")
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
				d, e := co.consume(ctx)
				if e != nil {
					log.WithRequestId(ctx).WithError(e).Error("reconsume failed")
					return
				}
				delivery = d
			}
		}
	}()
	return nil
}

func (qu *Queue) ConsumeOne(handler func(context.Context, string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) (err error) {
	if handler == nil {
		err = errors.Errorf("handler is nil")
		return
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		err = errors.WithStack(co.Error)
		return
	}
	ctx := co.newContext(co.ops.oneCtx)
	var msg amqp.Delivery
	var ok bool
	msg, ok, err = co.consumeOne(ctx)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !ok {
		err = errors.Errorf("queue is empty, can't get one msg")
		return
	}
	if co.ops.autoAck {
		handler(ctx, co.qu.ops.name, msg)
		return
	}
	if handler(ctx, co.qu.ops.name, msg) {
		err = msg.Ack(false)
		if err != nil {
			log.WithRequestId(ctx).WithError(err).Error("consume ack failed")
		}
		return
	}
	var retryCount int32
	if v, o := msg.Headers["x-retry-count"].(int32); o {
		retryCount = v + 1
	} else {
		retryCount = 1
	}
	if co.ops.nackRetry {
		if retryCount > co.ops.nackMaxRetryCount {
			log.WithRequestId(ctx).Warn("maximum retry %d exceeded, discard data", co.ops.nackMaxRetryCount)
			err = msg.Nack(false, co.ops.nackRequeue)
			if err != nil {
				log.WithRequestId(ctx).WithError(err).Error("consume nack failed")
			}
			return
		}
		msg.Headers["x-retry-count"] = retryCount
		err = qu.ex.PublishByte(
			msg.Body,
			WithPublishHeaders(msg.Headers),
			WithPublishRouteKey(msg.RoutingKey),
		)
		if err != nil {
			log.WithRequestId(ctx).WithError(err).Error("consume republish failed")
			return
		}
	}
	err = msg.Nack(false, co.ops.nackRequeue)
	if err != nil {
		log.WithRequestId(ctx).WithError(err).Error("consume nack failed")
	}
	return
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
