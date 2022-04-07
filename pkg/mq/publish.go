package mq

import (
	"github.com/golang/protobuf/proto"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"time"
)

type Publish struct {
	ex    *Exchange
	ops   PublishOptions
	msg   amqp.Publishing
	Error error
}

// PublishProto publish grpc proto msg
func (ex *Exchange) PublishProto(m proto.Message, options ...func(*PublishOptions)) (err error) {
	var b []byte
	b, err = proto.Marshal(m)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = ex.PublishByte(b, options...)
	return
}

// PublishJson publish str msg
func (ex *Exchange) PublishJson(m string, options ...func(*PublishOptions)) (err error) {
	err = ex.PublishByte([]byte(m), options...)
	return
}

// PublishByte publish byte msg
func (ex *Exchange) PublishByte(m []byte, options ...func(*PublishOptions)) (err error) {
	if len(m) == 0 {
		err = errors.Errorf("msg is empty")
		return
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		err = errors.WithStack(pu.Error)
		return
	}
	pu.msg.Body = m
	err = pu.publish()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (ex *Exchange) beforePublish(options ...func(*PublishOptions)) *Publish {
	var pu Publish
	if ex.Error != nil {
		pu.Error = ex.Error
		return &pu
	}
	ops := getPublishOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if len(ops.routeKeys) == 0 {
		pu.Error = errors.Errorf("route key is empty")
		return &pu
	}
	if ops.deadLetter {
		if ops.deadLetterFirstQueue == "" {
			pu.Error = errors.Errorf("dead letter first queue is empty")
			return &pu
		}
		if _, ok := ops.headers["x-retry-count"].(int32); !ok {
			ops.headers["x-retry-count"] = 0
		}
		ops.headers["x-first-death-queue"] = ops.deadLetterFirstQueue
	}
	if ops.deliveryMode <= 0 || ops.deliveryMode > amqp.Persistent {
		ops.deliveryMode = amqp.Persistent
	}
	pu.ops = *ops
	msg := amqp.Publishing{
		DeliveryMode: ops.deliveryMode,
		Timestamp:    time.Now(),
		ContentType:  ops.contentType,
		Headers:      ops.headers,
	}
	pu.msg = msg
	pu.ex = ex
	return &pu
}

func (pu *Publish) publish() error {
	ctx := pu.ops.ctx
	ch, err := pu.ex.rb.getChannel(ctx)
	if err != nil {
		return errors.Wrap(err, "get channel failed")
	}
	defer ch.Close()
	count := len(pu.ops.routeKeys)

	// set publisher confirm
	if err = ch.Confirm(false); err != nil {
		return errors.Wrap(err, "set publisher confirm failed")
	}
	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, count))
	returnCh := ch.NotifyReturn(make(chan amqp.Return, count))

	for i := 0; i < count; i++ {
		err = ch.Publish(
			pu.ex.ops.name,
			pu.ops.routeKeys[i],
			pu.ops.mandatory,
			pu.ops.immediate,
			pu.msg,
		)
		if err != nil {
			return errors.Wrap(err, "publish failed")
		}
	}
	timeout := time.Duration(pu.ex.rb.ops.timeout) * time.Second
	timer := time.NewTimer(timeout)
	index := 0
	for {
		select {
		case c := <-confirmCh:
			if !c.Ack {
				return errors.Errorf("publish confirm faled, delivery tag: %d", c.DeliveryTag)
			}
			index++
		case r := <-returnCh:
			log.WithRequestId(ctx).Error("publish return err: reply code: %d, reply text: %s, please check exchange name or route key", r.ReplyCode, r.ReplyText)
			return errors.Errorf("reply code: %d, reply text: %s", r.ReplyCode, r.ReplyText)
		case <-timer.C:
			log.WithRequestId(ctx).Warn("publish timeout: %ds, the connection may have been disconnected", pu.ex.rb.ops.timeout)
			return errors.Errorf("publish timeout: %ds", pu.ex.rb.ops.timeout)
		}
		if index == count {
			break
		}
	}
	return nil
}
