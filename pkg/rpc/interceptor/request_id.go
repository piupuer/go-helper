package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/query"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestId(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	newContext := query.NewRequestIdWithMetaData(ctx)
	resp, err := handler(newContext, r)
	md := metadata.Pairs(constant.MiddlewareRequestIdCtxKey, newContext.Value(constant.MiddlewareRequestIdCtxKey).(string))
	grpc.SendHeader(newContext, md)
	return resp, err
}
