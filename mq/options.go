package mq

import (
	"github.com/streadway/amqp"
	"github.com/thoas/go-funk"
)

type ChannelOptions struct {
	QosPrefetchCount int
	QosPrefetchSize  int
	QosGlobal        bool
}

func WithChannelQosPrefetchCount(prefetchCount int) func(*ChannelOptions) {
	return func(options *ChannelOptions) {
		getChannelOptionsOrSetDefault(options).QosPrefetchCount = prefetchCount
	}
}

func WithChannelQosPrefetchSize(prefetchSize int) func(*ChannelOptions) {
	return func(options *ChannelOptions) {
		getChannelOptionsOrSetDefault(options).QosPrefetchSize = prefetchSize
	}
}

func WithChannelQosGlobal(options *ChannelOptions) {
	getChannelOptionsOrSetDefault(options).QosGlobal = true
}

func getChannelOptionsOrSetDefault(options *ChannelOptions) *ChannelOptions {
	if options == nil {
		return &ChannelOptions{
			QosPrefetchCount: 2,
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
	DeadLetterName string
	DeadLetterKey  string
	MessageTTL     int32
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

func getPublishOptionsOrSetDefault(options *PublishOptions) *PublishOptions {
	if options == nil {
		return &PublishOptions{
			ContentType: "text/plain",
		}
	}
	return options
}
