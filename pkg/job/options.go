package job

import (
	"context"
	"github.com/piupuer/go-helper/pkg/logger"
)

type Options struct {
	logger        logger.Interface
	ctx           context.Context
	prefix        string
	AutoRequestId bool
}

func WithLogger(l logger.Interface) func(*Options) {
	return func(options *Options) {
		if l != nil {
			getOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithLoggerLevel(level logger.Level) func(*Options) {
	return func(options *Options) {
		l := options.logger
		if options.logger == nil {
			l = getOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
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

func WithAutoRequestId(options *Options) {
	getOptionsOrSetDefault(options).AutoRequestId = true
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			logger: logger.DefaultLogger(),
		}
	}
	return options
}

type DriverOptions struct {
	logger logger.Interface
	ctx    context.Context
	prefix string
}

func WithDriverLogger(l logger.Interface) func(*DriverOptions) {
	return func(options *DriverOptions) {
		if l != nil {
			getDriverOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithDriverLoggerLevel(level logger.Level) func(*DriverOptions) {
	return func(options *DriverOptions) {
		l := options.logger
		if options.logger == nil {
			l = getDriverOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithDriverContext(ctx context.Context) func(*DriverOptions) {
	return func(options *DriverOptions) {
		getDriverOptionsOrSetDefault(options).ctx = ctx
	}
}

func WithDriverPrefix(prefix string) func(*DriverOptions) {
	return func(options *DriverOptions) {
		getDriverOptionsOrSetDefault(options).prefix = prefix
	}
}

func getDriverOptionsOrSetDefault(options *DriverOptions) *DriverOptions {
	if options == nil {
		return &DriverOptions{
			logger: logger.DefaultLogger(),
		}
	}
	return options
}
