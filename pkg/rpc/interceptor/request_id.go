package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestId(options ...func(*RequestIdOptions)) grpc.UnaryServerInterceptor {
	ops := getRequestIdOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newContext := query.NewRequestIdWithMetaData(ctx, ops.ctxKey)
		resp, err := handler(newContext, r)
		md := metadata.Pairs(ops.ctxKey, newContext.Value(ops.ctxKey).(string))
		grpc.SendHeader(newContext, md)
		return resp, err
	}
}
