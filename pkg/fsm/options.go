package fsm

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
)

type Options struct {
	ctx        context.Context
	db         *gorm.DB
	prefix     string
	transition func(ctx context.Context, logs ...resp.FsmApprovalLog) error
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).ctx = getCtx(ctx)
	}
}

func WithDb(db *gorm.DB) func(*Options) {
	return func(options *Options) {
		if db != nil {
			getOptionsOrSetDefault(options).db = db
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
			ctx:    getCtx(nil),
			prefix: constant.FsmPrefix,
		}
	}
	return options
}

func getCtx(ctx context.Context) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	return context.WithValue(ctx, constant.LogSkipHelperCtxKey, false)
}
