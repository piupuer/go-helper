package mq

import (
	"context"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/streadway/amqp"
	"github.com/thoas/go-funk"
)

type RabbitOptions struct {
	reconnectInterval      int
	reconnectMaxRetryCount int
	channelMaxLostCount    int
	timeout                int
	logger                 logger.Interface
	ctx                    context.Context
}

func WithReconnectInterval(second int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if second > 0 {
			getRabbitOptionsOrSetDefault(options).reconnectInterval = second
		}
	}
}

func WithReconnectMaxRetryCount(count int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if count > 0 {
			getRabbitOptionsOrSetDefault(options).reconnectMaxRetryCount = count
		}
	}
}

func WithChannelMaxLostCount(count int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if count > 0 {
			getRabbitOptionsOrSetDefault(options).channelMaxLostCount = count
		}
	}
}

func WithTimeout(second int) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if second > 0 {
			getRabbitOptionsOrSetDefault(options).timeout = second
		}
	}
}

func WithLogger(l logger.Interface) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		if l != nil {
			getRabbitOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithLoggerLevel(level logger.Level) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		l := options.logger
		if options.logger == nil {
			l = getRabbitOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithContext(ctx context.Context) func(*RabbitOptions) {
	return func(options *RabbitOptions) {
		getRabbitOptionsOrSetDefault(options).ctx = ctx
	}
}

func getRabbitOptionsOrSetDefault(options *RabbitOptions) *RabbitOptions {
	if options == nil {
		return &RabbitOptions{
			timeout:                10,
			reconnectMaxRetryCount: 3,
			channelMaxLostCount:    5,
			reconnectInterval:      5,
			logger:                 logger.DefaultLogger(),
		}
	}
	return options
}

type ExchangeOptions struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       amqp.Table
	declare    bool
	namePrefix string
}

func WithExchangeName(name string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).name = name
	}
}

func WithExchangeKind(kind string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).kind = kind
	}
}

func WithExchangeDurable(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).durable = flag
	}
}

func WithExchangeAutoDelete(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).autoDelete = flag
	}
}

func WithExchangeInternal(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).internal = flag
	}
}

func WithExchangeNoWait(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).noWait = flag
	}
}

func WithExchangeArgs(args amqp.Table) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).args = args
	}
}

func WithExchangeDeclare(flag bool) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).declare = flag
	}
}

func WithExchangeNamePrefix(prefix string) func(*ExchangeOptions) {
	return func(options *ExchangeOptions) {
		getExchangeOptionsOrSetDefault(options).namePrefix = prefix
	}
}

func getExchangeOptionsOrSetDefault(options *ExchangeOptions) *ExchangeOptions {
	if options == nil {
		return &ExchangeOptions{
			kind:    amqp.ExchangeDirect,
			durable: true,
			declare: true,
		}
	}
	return options
}

type QueueOptions struct {
	name           string
	routeKeys      []string
	durable        bool
	autoDelete     bool
	exclusive      bool
	noWait         bool
	args           amqp.Table
	bindArgs       amqp.Table
	declare        bool
	bind           bool
	deadLetterName string
	deadLetterKey  string
	messageTTL     int32
	namePrefix     string
}

func WithQueueName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).name = name
	}
}

func WithQueueRouteKeys(keys ...string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).routeKeys = append(getQueueOptionsOrSetDefault(options).routeKeys, keys...)
	}
}

func WithQueueDurable(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).durable = flag
	}
}

func WithQueueAutoDelete(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).autoDelete = flag
	}
}

func WithQueueExclusive(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).exclusive = flag
	}
}

func WithQueueNoWait(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).noWait = flag
	}
}

func WithQueueArgs(args amqp.Table) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).args = args
	}
}

func WithQueueBindArgs(args amqp.Table) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).bindArgs = args
	}
}

func WithQueueDeclare(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).declare = flag
	}
}

func WithQueueBind(flag bool) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).bind = flag
	}
}

func WithQueueDeadLetterName(name string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).deadLetterName = name
	}
}

func WithQueueDeadLetterKey(key string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).deadLetterKey = key
	}
}

func WithQueueMessageTTL(ttl int32) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).messageTTL = ttl
	}
}

func getQueueOptionsOrSetDefault(options *QueueOptions) *QueueOptions {
	if options == nil {
		return &QueueOptions{
			durable: true,
			declare: true,
			bind:    true,
		}
	}
	return options
}

type PublishOptions struct {
	routeKeys    []string
	contentType  string
	headers      amqp.Table
	deliveryMode uint8
	mandatory    bool
	immediate    bool
	expiration   string
	ctx          context.Context
}

func WithPublishOptionsContentType(contentType string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).contentType = contentType
	}
}

func WithPublishOptionsHeaders(headers amqp.Table) func(*PublishOptions) {
	return func(options *PublishOptions) {
		getPublishOptionsOrSetDefault(options).headers = headers
	}
}

func WithPublishRouteKey(key string) func(*PublishOptions) {
	return func(options *PublishOptions) {
		d := getPublishOptionsOrSetDefault(options)
		keys := d.routeKeys
		if !funk.ContainsString(keys, key) {
			d.routeKeys = append(keys, key)
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
			contentType: "text/plain",
			ctx:         context.Background(),
		}
	}
	return options
}

type ConsumeOptions struct {
	qosPrefetchCount               int
	qosPrefetchSize                int
	qosGlobal                      bool
	consumer                       string
	autoAck                        bool
	exclusive                      bool
	noLocal                        bool
	noWait                         bool
	args                           amqp.Table
	nackRequeue                    bool
	autoRequestId                  bool
	newRequestIdWhenConnectionLost bool
	oneCtx                         context.Context
}

func WithConsumeQosPrefetchCount(prefetchCount int) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).qosPrefetchCount = prefetchCount
	}
}

func WithConsumeQosPrefetchSize(prefetchSize int) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).qosPrefetchSize = prefetchSize
	}
}

func WithConsumeQosGlobal(options *ConsumeOptions) {
	getConsumeOptionsOrSetDefault(options).qosGlobal = true
}

func WithConsumeConsumer(consumer string) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).consumer = consumer
	}
}

func WithConsumeAutoAck(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).autoAck = flag
	}
}

func WithConsumeExclusive(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).exclusive = flag
	}
}

func WithConsumeNoLocal(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).noLocal = flag
	}
}

func WithConsumeNoWait(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).noWait = flag
	}
}

func WithConsumeArgs(args amqp.Table) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).args = args
	}
}

func WithConsumeNackRequeue(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).nackRequeue = flag
	}
}

func WithConsumeAutoRequestId(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).autoRequestId = flag
	}
}

func WithConsumeNewRequestIdWhenConnectionLost(flag bool) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).newRequestIdWhenConnectionLost = flag
	}
}

func WithConsumeOneContext(ctx context.Context) func(*ConsumeOptions) {
	return func(options *ConsumeOptions) {
		getConsumeOptionsOrSetDefault(options).oneCtx = ctx
	}
}

func getConsumeOptionsOrSetDefault(options *ConsumeOptions) *ConsumeOptions {
	if options == nil {
		return &ConsumeOptions{
			qosPrefetchCount: 2,
			consumer:         "any",
		}
	}
	return options
}
