package mq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type Rabbit struct {
	dsn              string           // connection url
	connLock         int32            // lock when create connect
	conn             *amqp.Connection // connection instance
	lost             bool             // connection is lost
	channelLostCount int              // can't get channel count
	lostCh           chan error       // When the connection is lost, an error is sent to this channel
	ops              RabbitOptions
	Error            error
}

type Exchange struct {
	rb    *Rabbit
	ops   ExchangeOptions
	Error error
}

type Queue struct {
	ex    *Exchange
	ops   QueueOptions
	Error error
}

func NewRabbit(dsn string, options ...func(*RabbitOptions)) *Rabbit {
	var rb Rabbit
	rb.dsn = dsn
	rb.lost = true
	rb.lostCh = make(chan error)
	ops := getRabbitOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	rb.ops = *ops
	ctx := rb.ops.ctx
	err := rb.connect(ctx)
	if err != nil {
		rb.Error = err
		return &rb
	}
	go func() {
		quit := make(chan os.Signal)
		// kill (no param) default send syscall.SIGTERM
		// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
		// kill -2 is syscall.SIGINT
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		rb.ops.logger.Warn(ctx, "process is exiting")
		if rb.conn != nil {
			rb.conn.Close()
		}
	}()

	return &rb
}

// connect mq
func (rb *Rabbit) connect(ctx context.Context) error {
	if !rb.lost {
		return nil
	}
	v := atomic.LoadInt32(&rb.connLock)
	if v == 1 {
		rb.ops.logger.Warn(ctx, "the connection is creating")
		return fmt.Errorf("the connection is creating")
	}
	if !atomic.CompareAndSwapInt32(&rb.connLock, 0, 1) {
		rb.ops.logger.Warn(ctx, "the connection is creating")
		return fmt.Errorf("the connection is creating")
	}
	defer atomic.AddInt32(&rb.connLock, -1)
	conn, err := dialWithTimeout(rb.dsn, 5)
	if err != nil {
		rb.ops.logger.Error(ctx, "dial err: %v", err)
		return err
	}
	rb.conn = conn
	rb.lost = false

	go func() {
		connLost := rb.conn.NotifyClose(make(chan *amqp.Error))
		select {
		case err := <-connLost:
			// If the connection close is triggered by the Server, a reconnection takes place
			if err != nil && err.Server {
				rb.ops.logger.Warn(ctx, "connection is lost")
				rb.lost = true
				rb.lostCh <- err
			}
		}
	}()
	return nil
}

// get a channel
func (rb *Rabbit) getChannel(ctx context.Context) (*amqp.Channel, error) {
	if rb.channelLostCount > rb.ops.ChannelMaxLostCount {
		rb.ops.logger.Warn(ctx, "get channel failed %d retries, connection maybe lost", rb.channelLostCount)
		rb.lost = true
	}
	if rb.lost == true {
		err := rb.reconnect(ctx)
		if err != nil {
			return nil, err
		}
	}
	channel, err := rb.conn.Channel()
	if err != nil {
		rb.channelLostCount++
		return nil, err
	}
	rb.channelLostCount = 0
	return channel, nil
}

// reconnect mq
func (rb *Rabbit) reconnect(ctx context.Context) error {
	if !rb.lost {
		return nil
	}
	interval := time.Duration(rb.ops.ReconnectInterval) * time.Second
	retryCount := 0
	var err error
	for {
		if atomic.LoadInt32(&rb.connLock) == 1 {
			rb.ops.logger.Warn(ctx, "the connection is creating, please wait")
			return fmt.Errorf("the connection is creating")
		}
		time.Sleep(interval)
		err = rb.connect(ctx)
		if err == nil {
			return nil
		} else {
			retryCount++
		}
		if retryCount >= rb.ops.ReconnectMaxRetryCount {
			break
		}
	}
	rb.ops.logger.Warn(ctx, "unable to connect after %d retries, last err: %v", retryCount, err)
	return fmt.Errorf("unable to reconnect")
}

// bind a exchange
func (rb *Rabbit) Exchange(options ...func(*ExchangeOptions)) *Exchange {
	ex := rb.beforeExchange(options...)
	if ex.Error != nil {
		return ex
	}
	// the exchange will be declared
	if ex.ops.Declare {
		err := ex.declare()
		if err != nil {
			ex.Error = err
			return ex
		}
	}
	return ex
}

// before bind exchange
func (rb *Rabbit) beforeExchange(options ...func(*ExchangeOptions)) *Exchange {
	ctx := rb.ops.ctx
	var ex Exchange
	if rb.Error != nil {
		ex.Error = rb.Error
		return &ex
	}
	ops := getExchangeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.Name == "" {
		ex.Error = fmt.Errorf("exchange name is empty")
		rb.ops.logger.Error(ctx, "exchange name is empty")
		return &ex
	}
	switch ops.Kind {
	case amqp.ExchangeDirect:
	case amqp.ExchangeFanout:
	case amqp.ExchangeTopic:
	case amqp.ExchangeHeaders:
	default:
		ex.Error = fmt.Errorf("invalid exchange kind: %s", ops.Kind)
		rb.ops.logger.Error(ctx, "invalid exchange kind: %s", ops.Kind)
		return &ex
	}
	prefix := ""
	if ops.NamePrefix != "" {
		prefix = ops.NamePrefix
	}
	ops.Name = prefix + ops.Name
	ex.ops = *ops
	ex.rb = rb
	return &ex
}

// bind a queue
func (ex *Exchange) Queue(options ...func(*QueueOptions)) *Queue {
	qu := ex.beforeQueue(options...)
	if qu.Error != nil {
		return qu
	}
	if qu.ops.Declare {
		err := qu.declare()
		if err != nil {
			qu.Error = err
			return qu
		}
	}
	if qu.ops.Bind {
		err := qu.bind()
		if err != nil {
			qu.Error = err
			return qu
		}
	}
	return qu
}

// bind a dead letter queue
func (ex *Exchange) QueueWithDeadLetter(options ...func(*QueueOptions)) *Queue {
	ctx := ex.rb.ops.ctx
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
		ex.rb.ops.logger.Error(ctx, "dead letter name is empty")
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
	ctx := ex.rb.ops.ctx
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
		ex.rb.ops.logger.Error(ctx, "queue name is empty")
		qu.Error = fmt.Errorf("queue name is empty")
		return &qu
	}
	prefix := ""
	if ops.NamePrefix != "" {
		prefix = ops.NamePrefix
	}
	ops.Name = prefix + ops.Name
	qu.ops = *ops
	qu.ex = ex
	return &qu
}

// declare exchange
func (ex *Exchange) declare() error {
	ctx := ex.rb.ops.ctx
	ch, err := ex.rb.getChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(
		ex.ops.Name,
		ex.ops.Kind,
		ex.ops.Durable,
		ex.ops.AutoDelete,
		ex.ops.Internal,
		ex.ops.NoWait,
		ex.ops.Args,
	); err != nil {
		ex.rb.ops.logger.Error(ctx, "failed declare exchange %s(%s): %v", ex.ops.Name, ex.ops.Kind, err)
		return fmt.Errorf("failed declare exchange %s(%s)", ex.ops.Name, ex.ops.Kind)
	}
	return nil
}

// declare queue
func (qu *Queue) declare() error {
	ctx := qu.ex.rb.ops.ctx
	ch, err := qu.ex.rb.getChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Close()
	if _, err := ch.QueueDeclare(
		qu.ops.Name,
		qu.ops.Durable,
		qu.ops.AutoDelete,
		qu.ops.Exclusive,
		qu.ops.NoWait,
		qu.ops.Args,
	); err != nil {
		qu.ex.rb.ops.logger.Error(ctx, "failed to declare %s: %v", qu.ops.Name, err)
		return fmt.Errorf("failed to declare %s", qu.ops.Name)
	}
	return nil
}

// bind queue
func (qu *Queue) bind() error {
	ctx := qu.ex.rb.ops.ctx
	ch, err := qu.ex.rb.getChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Close()
	for _, key := range qu.ops.RouteKeys {
		if err := ch.QueueBind(
			qu.ops.Name,
			key,
			qu.ex.ops.Name,
			qu.ops.NoWait,
			qu.ops.Args,
		); err != nil {
			qu.ex.rb.ops.logger.Error(ctx, "failed to declare bind queue, queue: %s, key: %s, exchange: %s, err: %v", qu.ops.Name, key, qu.ex.ops.Name, err)
			return fmt.Errorf("failed to declare bind queue, queue: %s, key: %s, exchange: %s", qu.ops.Name, key, qu.ex.ops.Name)
		}
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
