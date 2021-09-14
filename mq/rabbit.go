package mq

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type Rabbit struct {
	dsn   string
	conn  *amqp.Connection
	Error error
}

type Channel struct {
	c     *amqp.Channel
	Error error
}

type Exchange struct {
	c     *amqp.Channel
	ops   ExchangeOptions
	Error error
}

type Queue struct {
	ops   QueueOptions
	Error error
}

func NewRabbit(dsn string) *Rabbit {
	var rb Rabbit
	conn, err := dialWithTimeout(dsn, 5)
	if err != nil {
		rb.Error = err
		return &rb
	}
	rb.conn = conn
	return &rb
}

// get a channel from the connection
func (rb *Rabbit) Channel(options ...func(*ChannelOptions)) *Channel {
	var ch Channel
	if rb.Error != nil {
		ch.Error = rb.Error
		return &ch
	}
	if rb.conn == nil {
		ch.Error = fmt.Errorf("invaild connection")
		return &ch
	}
	c, err := rb.conn.Channel()
	if err != nil {
		ch.Error = err
		return &ch
	}
	ops := &ChannelOptions{}
	for _, f := range options {
		f(ops)
	}
	// set qos
	err = c.Qos(ops.QosPrefetchCount, ops.QosPrefetchSize, ops.QosGlobal)
	if err != nil {
		ch.Error = err
		return &ch
	}
	ch.c = c
	return &ch
}

// bind a exchange
func (ch *Channel) Exchange(options ...func(*ExchangeOptions)) *Exchange {
	ex := ch.beforeExchange(options...)
	if ex.Error != nil {
		return ex
	}
	err := ex.declare(ch.c)
	if err != nil {
		ex.Error = err
		return ex
	}
	ex.c = ch.c
	return ex
}

// before bind exchange
func (ch *Channel) beforeExchange(options ...func(*ExchangeOptions)) *Exchange {
	var ex Exchange
	if ch.Error != nil {
		ex.Error = ch.Error
		return &ex
	}
	ops := getExchangeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.Name == "" {
		ex.Error = fmt.Errorf("exchange name is empty")
		return &ex
	}
	switch ops.Kind {
	case amqp.ExchangeDirect:
	case amqp.ExchangeFanout:
	case amqp.ExchangeTopic:
	case amqp.ExchangeHeaders:
	default:
		ex.Error = fmt.Errorf("invaild exchange kind: %s", ops.Kind)
		return &ex
	}
	ex.ops = *ops
	return &ex
}

// bind a queue
func (ex *Exchange) Queue(options ...func(*QueueOptions)) *Queue {
	qu := ex.beforeQueue(options...)
	if qu.Error != nil {
		return qu
	}
	if _, err := ex.c.QueueDeclare(
		qu.ops.Name,
		qu.ops.Durable,
		qu.ops.AutoDelete,
		qu.ops.Exclusive,
		qu.ops.NoWait,
		qu.ops.Args,
	); err != nil {
		qu.Error = fmt.Errorf("failed to declare %s: %v", qu.ops.Name, err)
		return qu
	}

	for _, key := range qu.ops.RouteKeys {
		if err := ex.c.QueueBind(
			qu.ops.Name,
			key,
			ex.ops.Name,
			qu.ops.NoWait,
			qu.ops.Args,
		); err != nil {
			qu.Error = fmt.Errorf("failed to declare bind queue, queue: %s, key: %s, exchange: %s, err: %v", qu.ops.Name, key, ex.ops.Name, err)
			return qu
		}
	}
	return qu
}

// bind a dead letter queue
func (ex *Exchange) QueueWithDeadLetter(options ...func(*QueueOptions)) *Queue {
	ops := getQueueOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	args := make(amqp.Table)
	if ops.Args != nil {
		args = ops.Args
	}

	if ops.DeadLetterName == "" {
		var qu Queue
		qu.Error = fmt.Errorf("dead letter name is empty")
		return &qu
	}
	args["x-dead-letter-exchange"] = ops.DeadLetterName
	if ops.DeadLetterKey != "" {
		args["x-dead-letter-routing-key"] = ops.DeadLetterKey
	}
	ops.Args = args
	return ex.Queue(func(options *QueueOptions) {
		*options = *ops
	})
}

// before bind queue
func (ex *Exchange) beforeQueue(options ...func(*QueueOptions)) *Queue {
	var qu Queue
	if ex.Error != nil {
		qu.Error = ex.Error
		return &qu
	}
	ops := getQueueOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	args := make(amqp.Table)
	if ops.Args != nil {
		args = ops.Args
	}
	if ops.MessageTTL > 0 {
		args["x-message-ttl"] = ops.MessageTTL
		ops.Args = args
	}
	if ops.Name == "" {
		qu.Error = fmt.Errorf("queue name is empty")
		return &qu
	}
	qu.ops = *ops
	return &qu
}

// declare exchange
func (ex *Exchange) declare(c *amqp.Channel) error {
	prefix := ""
	if ex.ops.NamePrefix != "" {
		prefix = ex.ops.NamePrefix
	}
	ex.ops.Name = prefix + ex.ops.Name
	if err := c.ExchangeDeclare(
		ex.ops.Name,
		ex.ops.Kind,
		ex.ops.Durable,
		ex.ops.AutoDelete,
		ex.ops.Internal,
		ex.ops.NoWait,
		ex.ops.Args,
	); err != nil {
		return fmt.Errorf("failed declare exchange %s(%s): %v", ex.ops.Name, ex.ops.Kind, err)
	}
	return nil
}

// dial rabbitmq with timeout(seconds)
func dialWithTimeout(dsn string, timeout int64) (*amqp.Connection, error) {
	conn, err := amqp.DialConfig(dsn, amqp.Config{
		Dial: amqp.DefaultDial(time.Duration(timeout) * time.Second),
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}
