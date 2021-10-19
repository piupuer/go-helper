package fsm

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
)

type Options struct {
	prefix string
	ctx    context.Context
}

func WithContext(ctx context.Context) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).ctx = ctx
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix: constant.FsmPrefix,
		}
	}
	return options
}

type MigrateOptions struct {
}
