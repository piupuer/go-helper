package mq

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"time"
)

type Publish struct {
	ex    *Exchange
	ops   PublishOptions
	msg   amqp.Publishing
	Error error
}

// publish grpc proto msg
func (ex *Exchange) PublishProto(m proto.Message, options ...func(*PublishOptions)) error {
	if m == nil {
		return fmt.Errorf("msg is nil")
	}
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		return pu.Error
	}
	pu.msg.Body = b
	err = pu.publish()
	if err != nil {
		return err
	}
	return nil
}

// publish str msg
func (ex *Exchange) PublishJson(m string, options ...func(*PublishOptions)) error {
	if m == "" {
		return fmt.Errorf("msg is empty")
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		return pu.Error
	}
	pu.msg.Body = []byte(m)
	err := pu.publish()
	if err != nil {
		return err
	}
	return nil
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
	ctx := pu.ops.ctx
	if len(ops.RouteKeys) == 0 {
		ex.rb.ops.logger.Error(ctx, "route key is empty")
		pu.Error = fmt.Errorf("route key is empty")
		return &pu
	}
	if ops.DeliveryMode <= 0 || ops.DeliveryMode > amqp.Persistent {
		ops.DeliveryMode = amqp.Persistent
	}
	pu.ops = *ops
	msg := amqp.Publishing{
		DeliveryMode: ops.DeliveryMode,
		Timestamp:    time.Now(),
		ContentType:  ops.ContentType,
		Headers:      ops.Headers,
	}
	pu.msg = msg
	pu.ex = ex
	return &pu
}

func (pu *Publish) publish() error {
	ctx := pu.ops.ctx
	ch, err := pu.ex.rb.getChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Close()
	count := len(pu.ops.RouteKeys)

	// set publisher confirm
	if err := ch.Confirm(false); err != nil {
		pu.ex.rb.ops.logger.Error(ctx, "set publisher confirm err: %v", err)
		return err
	}
	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, count))
	returnCh := ch.NotifyReturn(make(chan amqp.Return, count))

	for i := 0; i < count; i++ {
		err := ch.Publish(
			pu.ex.ops.Name,
			pu.ops.RouteKeys[i],
			pu.ops.Mandatory,
			pu.ops.Immediate,
			pu.msg,
		)
		if err != nil {
			pu.ex.rb.ops.logger.Error(ctx, "publish err: %v", err)
			return err
		}
	}
	timeout := time.Duration(pu.ex.rb.ops.Timeout) * time.Second
	timer := time.NewTimer(timeout)
	index := 0
	for {
		select {
		case c := <-confirmCh:
			if !c.Ack {
				pu.ex.rb.ops.logger.Error(ctx, "publish confirm err: %v", err)
				return fmt.Errorf("delivery tag: %d", c.DeliveryTag)
			}
			index++
		case r := <-returnCh:
			pu.ex.rb.ops.logger.Error(ctx, "publish return err: reply code: %d, reply text: %s", err)
			return fmt.Errorf("reply code: %d, reply text: %s", r.ReplyCode, r.ReplyText)
		case <-timer.C:
			pu.ex.rb.ops.logger.Warn(ctx, "publish timeout: %ds, the connection may have been disconnected", pu.ex.rb.ops.Timeout)
			return fmt.Errorf("publish timeout: %ds", pu.ex.rb.ops.Timeout)
		}
		if index == count {
			break
		}
	}
	return nil
}
