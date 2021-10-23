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
