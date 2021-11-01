package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/query"
	"google.golang.org/grpc"
)

func RequestId(options ...func(*RequestIdOptions)) grpc.UnaryServerInterceptor {
	ops := getRequestIdOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newContext := query.NewRequestId(ctx, ops.ctxKey)
		resp, err := handler(newContext, req)
		return resp, err
	}
}
