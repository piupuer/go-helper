package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/constant"
)

func RequestId(c *gin.Context) {
	// get from request header
	requestId := c.Request.Header.Get(constant.MiddlewareRequestIdHeaderName)

	if requestId == "" {
		requestId = uuid.NewString()
	}

	// set to context
	c.Set(constant.MiddlewareRequestIdCtxKey, requestId)

	// set to header
	c.Writer.Header().Set(constant.MiddlewareRequestIdHeaderName, requestId)
	c.Next()
}
