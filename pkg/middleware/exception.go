package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"net/http"
	"runtime/debug"
)

func Exception(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.WithRequestId(c).Error("[exception middleware]runtime err: %v\nstack: %v", err, string(debug.Stack()))
			rp := resp.Resp{
				Code:      resp.InternalServerError,
				Data:      map[string]interface{}{},
				Msg:       resp.CustomError[resp.InternalServerError],
				RequestId: c.GetString(constant.MiddlewareRequestIdCtxKey),
			}
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
			rid := c.GetString(constant.MiddlewareRequestIdCtxKey)
			rp := resp.Resp{
				Code:      resp.InternalServerError,
				Data:      map[string]interface{}{},
				Msg:       resp.CustomError[resp.InternalServerError],
				RequestId: rid,
			}
			if item, ok := err.(resp.Resp); ok {
				rp = item
				rp.RequestId = rid
			} else {
				log.WithRequestId(c).Error("[exception middleware]runtime err: %+v", err, string(debug.Stack()))
			}
			// set json data
			c.JSON(http.StatusOK, rp)
			c.Abort()
			return
		}
	}()
	c.Next()
}
