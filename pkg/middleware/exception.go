package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/resp"
	"net/http"
	"runtime/debug"
)

func Exception(options ...func(*ExceptionOptions)) gin.HandlerFunc {
	ops := getExceptionOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ops.logger.Error(c, "[exception middleware]runtime err: %v\nstack: %v", err, string(debug.Stack()))
				rp := resp.Resp{
					Code:      resp.InternalServerError,
					Data:      map[string]interface{}{},
					Msg:       resp.CustomError[resp.InternalServerError],
					RequestId: c.GetString(ops.requestIdCtxKey),
				}
				// set json data
				c.JSON(http.StatusOK, rp)
				if ops.operationLogCtxKey != "" {
					// set operation log key to context, It may be used OperationLog
					c.Set(ops.operationLogCtxKey, rp)
				}
				c.Abort()
				return
			}
		}()
		c.Next()
	}
}
