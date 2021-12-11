package interceptor

import (
	"context"
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
		c := context.WithValue(ctx, ops.txCtxKey, tx)
		resp, err := handler(c, r)
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		return resp, err
	}
}
