package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/tracing"
)

func SecurityHeader(c *gin.Context) {
	ctx := tracing.RealCtx(c)
	_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "SecurityHeader"))
	defer span.End()
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Frame-Options", "deny")
	c.Next()
}
