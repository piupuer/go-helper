package interceptor

import (
	"github.com/piupuer/go-helper/pkg/constant"
	"gorm.io/gorm"
)

type ExceptionOptions struct {
}

func getExceptionOptionsOrSetDefault(options *ExceptionOptions) *ExceptionOptions {
	if options == nil {
		return &ExceptionOptions{}
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

type AccessLogOptions struct {
	detail bool
}

func WithAccessLogDetail(flag bool) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		getAccessLogOptionsOrSetDefault(options).detail = flag
	}
}

func getAccessLogOptionsOrSetDefault(options *AccessLogOptions) *AccessLogOptions {
	if options == nil {
		return &AccessLogOptions{
			detail: true,
		}
	}
	return options
}
