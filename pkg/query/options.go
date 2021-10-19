package query

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/middleware"
)

const (
	primaryKey  = "id"
	cacheExpire = 86400
)

type MysqlOptions struct {
	logger          logger.Interface
	redis           redis.UniversalClient
	ctx             context.Context
	txCtxKey        string
	requestIdCtxKey string
}

func WithMysqlLogger(l logger.Interface) func(*MysqlOptions) {
	return func(options *MysqlOptions) {
		if l != nil {
			getMysqlOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithMysqlLoggerLevel(level logger.Level) func(*MysqlOptions) {
	return func(options *MysqlOptions) {
		l := options.logger
		if options.logger == nil {
			l = getMysqlOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithMysqlCtx(ctx context.Context) func(*MysqlOptions) {
	return func(options *MysqlOptions) {
		if ctx != nil {
			getMysqlOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func getMysqlOptionsOrSetDefault(options *MysqlOptions) *MysqlOptions {
	if options == nil {
		return &MysqlOptions{
			logger:          logger.DefaultLogger(),
			ctx:             context.Background(),
			txCtxKey:        middleware.TransactionTxCtxKey,
			requestIdCtxKey: middleware.RequestIdCtxKey,
		}
	}
	return options
}

type MysqlReadOptions struct {
	preloads    []string
	cache       bool
	cacheExpire int
}

func WithMySqlReadPreload(preloads ...string) func(*MysqlReadOptions) {
	return func(options *MysqlReadOptions) {
		options.preloads = append(options.preloads, preloads...)
	}
}

func WithMySqlReadCache(options *MysqlReadOptions) {
	options.cache = true
}

func WithMySqlReadCacheExpire(seconds int) func(*MysqlReadOptions) {
	return func(options *MysqlReadOptions) {
		if seconds > 0 {
			options.cacheExpire = seconds
		}
	}
}

func getMysqlReadOptionsOrSetDefault(options *MysqlReadOptions) *MysqlReadOptions {
	if options == nil {
		return &MysqlReadOptions{
			preloads:    []string{},
			cacheExpire: cacheExpire,
		}
	}
	return options
}
