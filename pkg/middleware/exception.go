package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/pkg/errors"
	"net/http"
	"runtime/debug"
)

func Exception(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.WithContext(c).WithError(errors.Errorf("%v", err)).Error("runtime exception, stack: %s", string(debug.Stack()))
			rp := resp.Resp{
				Code: resp.InternalServerError,
				Data: map[string]interface{}{},
				Msg:  resp.CustomError[resp.InternalServerError],
			}
			rp.RequestId, _, _ = tracing.GetId(c)
			// set json data
			c.JSON(http.StatusOK, rp)
			c.Abort()
			return
		}
	}()
	c.Next()
}

func ExceptionWithNoTransaction(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			rp := resp.Resp{
				Code: resp.InternalServerError,
				Data: map[string]interface{}{},
				Msg:  resp.CustomError[resp.InternalServerError],
			}
			if item, ok := err.(resp.Resp); ok {
				rp = item
			} else {
				log.WithContext(c).WithError(errors.Errorf("%v", err)).Error("runtime exception, stack: %s", string(debug.Stack()))
			}
			rp.RequestId, _, _ = tracing.GetId(c)
			// set json data
			c.JSON(http.StatusOK, rp)
			c.Abort()
			return
		}
	}()
	c.Next()
}
