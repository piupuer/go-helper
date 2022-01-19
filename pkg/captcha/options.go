package captcha

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Options struct {
	logger *logger.Wrapper
	ctx    context.Context
	redis  redis.UniversalClient
	prefix string
	expire int
}

func WithLogger(l *logger.Wrapper) func(*Options) {
	return func(options *Options) {
		if l != nil {
			getOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !utils.InterfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
			options.logger = options.logger.WithRequestId(ctx)
		}
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithExpire(min int) func(*Options) {
	return func(options *Options) {
		if min > 0 {
			getOptionsOrSetDefault(options).expire = min
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			logger: logger.NewDefaultWrapper(),
			ctx:    context.Background(),
			prefix: constant.CaptchaPrefix,
			expire: 10,
		}
	}
	return options
}
