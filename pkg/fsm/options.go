package fsm

import (
	"context"
	"strings"
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
		if !strings.HasSuffix(prefix, "_") {
			prefix = prefix + "_"
		}
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			prefix: "tb_fsm_",
		}
	}
	return options
}

type MigrateOptions struct {
}
