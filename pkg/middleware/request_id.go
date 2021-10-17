package middleware

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func RequestId(options ...func(*RequestIdOptions)) gin.HandlerFunc {
	ops := getRequestIdOptionsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		// get from request header
		requestId := c.Request.Header.Get(ops.headerName)

		if requestId == "" {
			uuid4 := uuid.NewV4()
			requestId = uuid4.String()
		}

		// set to context
		c.Set(ops.ctxKey, requestId)

		// set to header
		c.Writer.Header().Set(ops.headerName, requestId)
		c.Next()
	}
}
