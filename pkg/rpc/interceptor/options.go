package interceptor

import (
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
	dbNoTx *gorm.DB
}

func WithTransactionDbNoTx(db *gorm.DB) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		if db != nil {
			getTransactionOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func getTransactionOptionsOrSetDefault(options *TransactionOptions) *TransactionOptions {
	if options == nil {
		return &TransactionOptions{}
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
