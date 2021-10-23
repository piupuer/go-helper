package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	v1 "github.com/piupuer/go-helper/api/v1"
	"github.com/piupuer/go-helper/pkg/middleware"
)

type Options struct {
	redis          redis.UniversalClient
	redisService   bool
	group          *gin.RouterGroup
	jwt            bool
	jwtOps         []func(*middleware.JwtOptions)
	casbin         bool
	casbinOps      []func(*middleware.CasbinOptions)
	idempotence    bool
	idempotenceOps []func(*middleware.IdempotenceOptions)
	v1Ops          []func(options *v1.Options)
}

func WithGroup(group *gin.RouterGroup) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).group = group
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithRedisService(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).redisService = flag
	}
}

func WithJwt(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).jwt = flag
	}
}

func WithJwtOps(ops ...func(*middleware.JwtOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).jwtOps = append(getOptionsOrSetDefault(options).jwtOps, ops...)
	}
}

func WithCasbin(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).casbin = flag
	}
}

func WithCasbinOps(ops ...func(*middleware.CasbinOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).casbinOps = append(getOptionsOrSetDefault(options).casbinOps, ops...)
	}
}

func WithIdempotence(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).idempotence = flag
	}
}

func WithIdempotenceOps(ops ...func(*middleware.IdempotenceOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).idempotenceOps = append(getOptionsOrSetDefault(options).idempotenceOps, ops...)
	}
}

func WithV1Ops(ops ...func(options *v1.Options)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).v1Ops = append(getOptionsOrSetDefault(options).v1Ops, ops...)
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			redisService: false,
			jwt:          true,
			casbin:       true,
			idempotence:  true,
		}
	}
	return options
}
