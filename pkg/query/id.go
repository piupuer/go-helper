package query

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/utils"
	"google.golang.org/grpc/metadata"
)

func NewRequestId(ctx context.Context) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	requestId := ""
	// get value from context
	requestIdValue := ctx.Value(constant.MiddlewareRequestIdCtxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		requestId = item
	}
	// gen uuid
	if requestId == "" {
		requestId = uuid.NewString()
	}
	return context.WithValue(ctx, constant.MiddlewareRequestIdCtxKey, requestId)
}

func NewRequestIdWithMetaData(ctx context.Context) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	requestId := ""
	// get value from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		arr := md.Get(constant.MiddlewareRequestIdCtxKey)
		if len(arr) == 1 {
			requestId = arr[0]
		}
	} else {
		md = metadata.MD{}
	}
	// get value from context
	requestIdValue := ctx.Value(constant.MiddlewareRequestIdCtxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		requestId = item
	}
	// gen uuid
	if requestId == "" {
		requestId = uuid.NewString()
	}
	md.Set(constant.MiddlewareRequestIdCtxKey, requestId)
	ctx = metadata.NewIncomingContext(ctx, md)
	return context.WithValue(ctx, constant.MiddlewareRequestIdCtxKey, requestId)
}

func NewRequestIdReturnGinCtx(ctx context.Context) *gin.Context {
	c := NewRequestId(ctx)
	keys := make(map[string]interface{})
	keys[constant.MiddlewareRequestIdCtxKey] = c.Value(constant.MiddlewareRequestIdCtxKey)
	return &gin.Context{
		Keys: keys,
	}
}
