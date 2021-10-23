package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/query"
)

type Options struct {
	cache                    bool
	cacheOps                 []func(options *query.RedisOptions)
	dbOps                    []func(options *query.MysqlOptions)
	redis                    redis.UniversalClient
	operationAllowedToDelete bool
	getCurrentUser           func(c *gin.Context) ms.CurrentUser
	findRoleKeywordByRoleIds func(roleIds []uint) []string
}

func WithCacheOps(ops ...func(options *query.RedisOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).cacheOps = append(getOptionsOrSetDefault(options).cacheOps, ops...)
	}
}

func WithDbOps(ops ...func(options *query.MysqlOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).dbOps = append(getOptionsOrSetDefault(options).dbOps, ops...)
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			cache: false,
		}
	}
	return options
}

func ParseOptions(options ...func(*Options)) *Options {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return ops
}

func (ops *Options) addCtx(ctx context.Context) {
	if ops.cache {
		ops.cacheOps = append(ops.cacheOps, query.WithRedisCtx(ctx))
	}
	ops.dbOps = append(ops.dbOps, query.WithMysqlCtx(ctx))
}
