package mq

import (
	"context"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/logger"
	"github.com/streadway/amqp"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	glogger "gorm.io/gorm/logger"
	"time"

	"os"
)

type RabbitOptions struct {
	ReconnectInterval      int
	ReconnectMaxRetryCount int
	ChannelMaxLostCount    int
	Timeout                int
	logger                 glogger.Interface
	ctx                    context.Context
}

func WithReconnectInterval(second int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if second > 0 {
			getRabbitOptionsOrSetDefault(options).ReconnectInterval = second
		}
	}
}

func WithReconnectMaxRetryCount(count int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if count > 0 {
			getRabbitOptionsOrSetDefault(options).ReconnectMaxRetryCount = count
		}
	}
}

func WithChannelMaxLostCount(count int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if count > 0 {
			getRabbitOptionsOrSetDefault(options).ChannelMaxLostCount = count
		}
	}
}

func WithTimeout(second int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if second > 0 {
			getRabbitOptionsOrSetDefault(options).Timeout = second
		}
	}
}

func WithLogger(l glogger.Interface) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if l != nil {
			getRabbitOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithLoggerLevel(level glogger.LogLevel) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		l := options.logger
		if options.logger == nil {
			l = getRabbitOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogMode(level)
	}
}

func WithContext(ctx context.Context) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		getRabbitOptionsOrSetDefault(options).ctx = ctx
	}
}

func getRabbitOptionsOrSetDefault(options *RabbitOptions) *RabbitOptions {
	if options == nil {
		enConfig := zap.NewProductionEncoderConfig()
		enConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(carbon.Time2Carbon(t).ToRfc3339String())
		}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(enConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
			zapcore.DebugLevel,
		)
		l := zap.New(core)
		return &RabbitOptions{
			Timeout:                10,
			ReconnectMaxRetryCount: 3,
			ChannelMaxLostCount:    5,
			ReconnectInterval:      5,
			logger: logger.New(
				l,
				logger.Config{
					LineNumLevel: 2,
					Config: glogger.Config{
						Colorful: true,
					},
				},
			),
		}
	}
	return options
}

type ExchangeOptions struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
	Declare    bool
	NamePrefix string
}

func WithExchangeName(name string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).Name = name
	}
}

func WithExchangeKind(kind string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).Kind = kind
	}
}

func WithExchangeDurable(options *ExchangeOptions) {
	getExchangeOptionsOrSetDefault(options).Durable = true
}

func WithExchangeAutoDelete(options *ExchangeOptions) {
	getExchangeOptionsOrSetDefault(options).AutoDelete = true
}

func WithExchangeInternal(options *ExchangeOptions) {
	getExchangeOptionsOrSetDefault(options).Internal = true
}

func WithExchangeNoWait(options *ExchangeOptions) {
	getExchangeOptionsOrSetDefault(options).NoWait = true
}

func WithExchangeArgs(args amqp.Table) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).Args = args
	}
}

func WithExchangeSkipDeclare(options *ExchangeOptions) {
	getExchangeOptionsOrSetDefault(options).Declare = false
}

func WithExchangeNamePrefix(prefix string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).NamePrefix = prefix
	}
}

func getExchangeOptionsOrSetDefault(options *ExchangeOptions) *ExchangeOptions {
	if options == nil {
		return &ExchangeOptions{
			Kind:    amqp.ExchangeDirect,
			Durable: true,
			Declare: true,
		}
	}
	return options
}

type QueueOptions struct {
	Name           string
	RouteKeys      []string
	Durable        bool
	AutoDelete     bool
	Exclusive      bool
	NoWait         bool
	Args           amqp.Table
	BindArgs       amqp.Table
	Declare        bool
	Bind           bool
	DeadLetterName string
	DeadLetterKey  string
	MessageTTL     int32
	NamePrefix     string
}

func WithQueueName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).Name = name
	}
}

func WithQueueRouteKey(key string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		d := getQueueOptionsOrSetDefault(options)
		keys := d.RouteKeys
		if !funk.ContainsString(keys, key) {
			d.RouteKeys = append(keys, key)
		}
	}
}

func WithQueueSkipDeclare(options *QueueOptions) {
	getQueueOptionsOrSetDefault(options).Declare = false
}

func WithQueueSkipBind(options *QueueOptions) {
	getQueueOptionsOrSetDefault(options).Bind = false
}

func WithQueueDeadLetterName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).DeadLetterName = name
	}
}

func WithQueueDeadLetterKey(key string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).DeadLetterKey = key
	}
}

func WithQueueMessageTTL(ttl int32) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).MessageTTL = ttl
	}
}

func getQueueOptionsOrSetDefault(options *QueueOptions) *QueueOptions {
	if options == nil {
		return &QueueOptions{
			Durable: true,
			Declare: true,
			Bind:    true,
		}
	}
	return options
}

type PublishOptions struct {
	RouteKeys    []string
	ContentType  string
	Headers      amqp.Table
	DeliveryMode uint8
	Mandatory    bool
	Immediate    bool
	Expiration   string
	ctx          context.Context
}

func WithPublishOptionsContentType(contentType string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).ContentType = contentType
	}
}

func WithPublishOptionsHeaders(headers amqp.Table) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).Headers = headers
	}
}

func WithPublishRouteKey(key string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		d := getPublishOptionsOrSetDefault(options)
		keys := d.RouteKeys
		if !funk.ContainsString(keys, key) {
			d.RouteKeys = append(keys, key)
		}
	}
}

func WithPublishContext(ctx context.Context) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).ctx = ctx
	}
}

func getPublishOptionsOrSetDefault(options *PublishOptions) *PublishOptions {
	if options == nil {
		return &PublishOptions{
			ContentType: "text/plain",
			ctx:         context.Background(),
		}
	}
	return options
}

type ConsumeOptions struct {
	QosPrefetchCount               int
	QosPrefetchSize                int
	QosGlobal                      bool
	Consumer                       string
	AutoAck                        bool
	Exclusive                      bool
	NoLocal                        bool
	NoWait                         bool
	Args                           amqp.Table
	NackRequeue                    bool
	AutoRequestId                  bool
	NewRequestIdWhenConnectionLost bool
	oneCtx                         context.Context
}

func WithConsumeQosPrefetchCount(prefetchCount int) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).QosPrefetchCount = prefetchCount
	}
}

func WithConsumeQosPrefetchSize(prefetchSize int) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).QosPrefetchSize = prefetchSize
	}
}

func WithConsumeQosGlobal(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).QosGlobal = true
}

func WithConsumeConsumer(consumer string) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).Consumer = consumer
	}
}

func WithConsumeAutoAck(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).AutoAck = true
}

func WithConsumeExclusive(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).Exclusive = true
}

func WithConsumeNoLocal(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).NoLocal = true
}

func WithConsumeNoWait(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).NoWait = true
}

func WithConsumeArgs(args amqp.Table) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).Args = args
	}
}

func WithConsumeNackRequeue(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).NackRequeue = true
}

func WithConsumeAutoRequestId(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).AutoRequestId = true
}

func WithConsumeNewRequestIdWhenConnectionLost(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).NewRequestIdWhenConnectionLost = true
}

func WithConsumeOneContext(ctx context.Context) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).oneCtx = ctx
	}
}

func getConsumeOptionsOrSetDefault(options *ConsumeOptions) *ConsumeOptions {
	if options == nil {
		return &ConsumeOptions{
			QosPrefetchCount: 2,
			Consumer:         "any",
		}
	}
	return options
}
