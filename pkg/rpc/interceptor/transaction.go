package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"google.golang.org/grpc"
)

func Transaction(options ...func(*TransactionOptions)) grpc.UnaryServerInterceptor {
	ops := getTransactionOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.dbNoTx == nil {
		panic("dbNoTx is empty")
	}
	return func(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tx := ops.dbNoTx.Begin()
		c := context.WithValue(ctx, constant.MiddlewareTransactionTxCtxKey, tx)
		resp, err := handler(c, r)
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		return resp, err
	}
}
