package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestId(options ...func(*RequestIdOptions)) gin.HandlerFunc {
	ops := getRequestIdOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		// get from request header
		requestId := c.Request.Header.Get(ops.headerName)

		if requestId == "" {
			requestId = uuid.NewString()
		}

		// set to context
		c.Set(ops.ctxKey, requestId)

		// set to header
		c.Writer.Header().Set(ops.headerName, requestId)
		c.Next()
	}
}
