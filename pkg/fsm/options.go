package fsm

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Options struct {
	logger     logger.Interface
	prefix     string
	ctx        context.Context
	transition func(ctx context.Context, logs ...resp.FsmApprovalLog) error
}

func WithLogger(l logger.Interface) func(*Options) {
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
		}
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithTransition(fun func(ctx context.Context, logs ...resp.FsmApprovalLog) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).transition = fun
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			logger: logger.DefaultLogger(),
			prefix: constant.FsmPrefix,
		}
	}
	return options
}
