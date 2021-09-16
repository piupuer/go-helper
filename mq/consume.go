package mq

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Consume struct {
	qu    *Queue
	ops   ConsumeOptions
	Error error
}

func (qu *Queue) Consume(handler func(string, amqp.Delivery) bool, options ...func(*ConsumeOptions)) error {
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}
	co := qu.beforeConsume(options...)
	if co.Error != nil {
		return co.Error
	}
	delivery, err := co.consume()
	if err != nil {
		return err
	}
	go func() {
		for {
			go func() {
				for msg := range delivery {
					if co.ops.AutoAck {
						handler(qu.ops.Name, msg)
						continue
					}
					if handler(qu.ops.Name, msg) {
						err := msg.Ack(false)
						if err != nil {
						}
					} else {
						err := msg.Nack(false, true)
						if err != nil {
						}
					}
				}
			}()
			// wait connection lost
			if err = <-qu.ex.rb.lostCh; err != nil {
				for {
					err = qu.ex.rb.reconnect()
					if err == nil {
						break
					}
				}
				d, err := co.consume()
				if err != nil {
					return
				}
				delivery = d
			}
		}
	}()
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

func (co *Consume) consume() (<-chan amqp.Delivery, error) {
	channel, err := co.qu.ex.rb.getChannel()
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
