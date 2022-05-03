package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/tracing"
)

func RequestId(c *gin.Context) {
	ctx := tracing.RealCtx(c)
	_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "RequestId"))
	defer span.End()
	requestId, _, _ := tracing.GetId(c)
	if requestId == "" {
		c.Request = c.Request.WithContext(tracing.NewId(c))
	}
	c.Next()
}
