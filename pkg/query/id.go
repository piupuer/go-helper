package query

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/utils"
	uuid "github.com/satori/go.uuid"
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

func NewRequestIdReturnGinCtx(ctx context.Context, ctxKey string) *gin.Context {
	c := NewRequestId(ctx, ctxKey)
	keys := make(map[string]interface{})
	keys[ctxKey] = c.Value(ctxKey)
	return &gin.Context{
		Keys: keys,
	}
}
