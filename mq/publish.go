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
func (ex *Exchange) PublishProto(m proto.Message, options ...func(*PublishOptions)) *Exchange {
	if m == nil {
		ex.Error = fmt.Errorf("msg is nil")
		return ex
	}
	b, err := proto.Marshal(m)
	if err != nil {
		ex.Error = err
		return ex
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		ex.Error = pu.Error
		return ex
	}
	pu.msg.Body = b
	err = pu.publish()
	if err != nil {
		ex.Error = err
		return ex
	}
	return ex
}

// publish str msg
func (ex *Exchange) PublishJson(m string, options ...func(*PublishOptions)) *Exchange {
	if m == "" {
		ex.Error = fmt.Errorf("msg is empty")
		return ex
	}
	pu := ex.beforePublish(options...)
	if pu.Error != nil {
		ex.Error = pu.Error
		return ex
	}
	pu.msg.Body = []byte(m)
	err := pu.publish()
	if err != nil {
		ex.Error = err
		return ex
	}
	return ex
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
	if len(ops.RouteKeys) == 0 {
		pu.Error = fmt.Errorf("route key is empty")
		return &pu
	}
	if ops.DeliveryMode <= 0 || ops.DeliveryMode > amqp.Persistent {
		ops.DeliveryMode = amqp.Persistent
	}
	// enable publisher confirm
	if err := ex.c.Confirm(false); err != nil {
		pu.Error = err
		return &pu
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
	for _, key := range pu.ops.RouteKeys {
		err := pu.ex.c.Publish(
			pu.ex.ops.Name,
			key,
			pu.ops.Mandatory,
			pu.ops.Immediate,
			pu.msg,
		)
		select {
		case ntf := <-pu.ex.c.NotifyPublish(make(chan amqp.Confirmation, 1)):
			if !ntf.Ack {
				return fmt.Errorf("delivery tag: %d", ntf.DeliveryTag)
			}
		case ch := <-pu.ex.c.NotifyReturn(make(chan amqp.Return)):
			return fmt.Errorf("reply code: %d, reply text: %s", ch.ReplyCode, ch.ReplyText)
		case <-time.After(10 * time.Second):
			return fmt.Errorf("connect timeout")
		}
		if err != nil {
			return err
		}
	}
	return nil
}
