package query

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
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

func WithMysqlRedis(rd redis.UniversalClient) func(*MysqlOptions) {
	return func(options *MysqlOptions) {
		if rd != nil {
			getMysqlOptionsOrSetDefault(options).redis = rd
		}
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
			txCtxKey:        constant.MiddlewareTransactionTxCtxKey,
			requestIdCtxKey: constant.MiddlewareRequestIdCtxKey,
		}
	}
	return options
}

type MysqlReadOptions struct {
	preloads    []string
	cache       bool
	cacheExpire int
	column      string
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

func WithMySqlReadColumn(column string) func(*MysqlReadOptions) {
	return func(options *MysqlReadOptions) {
		if column != "" {
			options.column = column
		}
	}
}

func getMysqlReadOptionsOrSetDefault(options *MysqlReadOptions) *MysqlReadOptions {
	if options == nil {
		return &MysqlReadOptions{
			preloads:    []string{},
			cacheExpire: constant.QueryCacheExpire,
			column:      constant.QueryPrimaryKey,
		}
	}
	return options
}
