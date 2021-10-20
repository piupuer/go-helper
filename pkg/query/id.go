package query

import (
	"context"
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
