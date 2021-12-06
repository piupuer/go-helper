package query

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/metadata"
)

func NewRequestId(ctx context.Context, ctxKey string) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	requestId := ""
	// get value from context
	requestIdValue := ctx.Value(ctxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		requestId = item
	}
	// gen uuid
	if requestId == "" {
		uuid4 := uuid.NewV4()
		requestId = uuid4.String()
	}
	return context.WithValue(ctx, ctxKey, requestId)
}

func NewRequestIdWithMetaData(ctx context.Context, ctxKey string) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	requestId := ""
	// get value from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		arr := md.Get(ctxKey)
		if len(arr) == 1 {
			requestId = arr[0]
		}
	} else {
		md = metadata.MD{}
	}
	// get value from context
	requestIdValue := ctx.Value(ctxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		requestId = item
	}
	// gen uuid
	if requestId == "" {
		uuid4 := uuid.NewV4()
		requestId = uuid4.String()
	}
	md.Set(ctxKey, requestId)
	ctx = metadata.NewIncomingContext(ctx, md)
	return context.WithValue(ctx, ctxKey, requestId)
}

func NewRequestIdReturnGinCtx(ctx context.Context, ctxKey string) *gin.Context {
	c := NewRequestId(ctx, ctxKey)
	keys := make(map[string]interface{})
	keys[ctxKey] = c.Value(ctxKey)
	return &gin.Context{
		Keys: keys,
	}
}
