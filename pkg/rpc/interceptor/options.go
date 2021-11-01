package interceptor

import (
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"gorm.io/gorm"
)

type ExceptionOptions struct {
	logger logger.Interface
}

func WithExceptionLogger(l logger.Interface) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		if l != nil {
			getExceptionOptionsOrSetDefault(options).logger = l
		}
	}
}

func getExceptionOptionsOrSetDefault(options *ExceptionOptions) *ExceptionOptions {
	if options == nil {
		return &ExceptionOptions{
			logger: logger.DefaultLogger(),
		}
	}
	return options
}

type TransactionOptions struct {
	dbNoTx   *gorm.DB
	txCtxKey string
}

func WithTransactionDbNoTx(db *gorm.DB) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		if db != nil {
			getTransactionOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func WithTransactionTxCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOrSetDefault(options).txCtxKey = key
	}
}

func getTransactionOptionsOrSetDefault(options *TransactionOptions) *TransactionOptions {
	if options == nil {
		return &TransactionOptions{
			txCtxKey: constant.MiddlewareTransactionTxCtxKey,
		}
	}
	return options
}

type RequestIdOptions struct {
	ctxKey string
}

func WithRequestIdCtxKey(key string) func(*RequestIdOptions) {
	return func(options *RequestIdOptions) {
		getRequestIdOptionsOrSetDefault(options).ctxKey = key
	}
}

func getRequestIdOptionsOrSetDefault(options *RequestIdOptions) *RequestIdOptions {
	if options == nil {
		return &RequestIdOptions{
			ctxKey: constant.MiddlewareRequestIdCtxKey,
		}
	}
	return options
}
