package tracing

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/constant"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"strings"
)

func NewId(ctx context.Context) context.Context {
	ctx = RealCtx(ctx)
	requestId, traceId, spanId := GenId(ctx)
	if traceId != "" {
		ctx = context.WithValue(ctx, constant.MiddlewareTraceIdCtxKey, traceId)
		ctx = context.WithValue(ctx, constant.MiddlewareSpanIdCtxKey, spanId)
	} else {
		ctx = context.WithValue(ctx, constant.MiddlewareRequestIdCtxKey, requestId)
	}
	return ctx
}

func NewGinId(ctx context.Context) *gin.Context {
	keys := make(map[string]interface{})
	ctx = RealCtx(ctx)
	requestId, traceId, spanId := GenId(ctx)
	if traceId != "" {
		keys[constant.MiddlewareTraceIdCtxKey] = traceId
		keys[constant.MiddlewareSpanIdCtxKey] = spanId
	} else {
		keys[constant.MiddlewareRequestIdCtxKey] = requestId
	}
	return &gin.Context{
		Keys: keys,
	}
}

func GenId(ctx context.Context) (string, string, string) {
	ctx = RealCtx(ctx)
	requestId := RequestId(ctx)
	traceId, spanId := TraceId(ctx)
	if traceId != "" {
		requestId = traceId
	}
	// gen uuid
	if requestId == "" {
		requestId = uuid.NewString()
	}
	return requestId, traceId, spanId
}

func GetId(ctx context.Context) (string, string, string) {
	ctx = RealCtx(ctx)
	requestId := RequestId(ctx)
	traceId, spanId := TraceId(ctx)
	if traceId != "" {
		requestId = traceId
	}
	return requestId, traceId, spanId
}

func RequestId(ctx context.Context) (id string) {
	ctx = RealCtx(ctx)
	// get value from context
	requestIdValue := ctx.Value(constant.MiddlewareRequestIdCtxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		id = item
	}
	return
}

func TraceId(ctx context.Context) (traceId, spanId string) {
	ctx = RealCtx(ctx)
	span := trace.SpanFromContext(ctx).SpanContext()
	if span.IsValid() {
		traceId = span.TraceID().String()
		spanId = span.SpanID().String()
	}
	return
}

func RealCtx(ctx context.Context) context.Context {
	if interfaceIsNil(ctx) {
		return context.Background()
	}
	if c, ok := ctx.(*gin.Context); ok {
		ctx = c.Request.Context()
	}
	return ctx
}

func Name(name ...string) string {
	return strings.Join(name, ".")
}

func interfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
