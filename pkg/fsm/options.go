package fsm

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Options struct {
	ctx        context.Context
	prefix     string
	transition func(ctx context.Context, logs ...resp.FsmApprovalLog) error
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
			ctx:    context.Background(),
			prefix: constant.FsmPrefix,
		}
	}
	return options
}
