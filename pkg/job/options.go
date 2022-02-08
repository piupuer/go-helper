package job

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Options struct {
	ctx            context.Context
	prefix         string
	taskNameCtxKey string
	autoRequestId  bool
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !utils.InterfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithAutoRequestId(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).autoRequestId = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			ctx:            context.Background(),
			taskNameCtxKey: constant.JobTaskNameCtxKey,
		}
	}
	return options
}

type DriverOptions struct {
	ctx    context.Context
	prefix string
}

func WithDriverCtx(ctx context.Context) func(*DriverOptions) {
	return func(options *DriverOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getDriverOptionsOrSetDefault(options).ctx = ctx
		}
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
			ctx:    context.Background(),
			prefix: constant.JobDriverPrefix,
		}
	}
	return options
}

type CronOptions struct {
	ctx context.Context
}

func WithCronCtx(ctx context.Context) func(*CronOptions) {
	return func(options *CronOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getCronOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func getCronOptionsOrSetDefault(options *CronOptions) *CronOptions {
	if options == nil {
		return &CronOptions{
			ctx: context.Background(),
		}
	}
	return options
}
