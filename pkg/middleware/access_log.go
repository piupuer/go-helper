package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/utils"
	"time"
)

func AccessLog(options ...func(*AccessLogOptions)) gin.HandlerFunc {
	ops := getAccessLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		// calc request exec time
		execTime := endTime.Sub(startTime).String()

		reqMethod := c.Request.Method
		reqPath := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		var rp string
		if ops.detail {
			data, ok := c.Get(constant.MiddlewareOperationLogCtxKey)
			if ok {
				rp = utils.Struct2Json(data)
			}
		}

		if reqMethod == "OPTIONS" || reqPath == fmt.Sprintf("/%s/ping", ops.urlPrefix) {
			if !ops.detail {
				log.WithRequestId(c).Debug(
					"%s %s %d %s %s",
					reqMethod,
					reqPath,
					statusCode,
					execTime,
					clientIP,
				)
			} else {
				log.WithRequestId(c).Debug(
					"%s %s %d %s %s resp: `%s`",
					reqMethod,
					reqPath,
					statusCode,
					execTime,
					clientIP,
					rp,
				)
			}
		} else {
			if !ops.detail {
				log.WithRequestId(c).Info(
					"%s %s %d %s %s",
					reqMethod,
					reqPath,
					statusCode,
					execTime,
					clientIP,
				)
			} else {
				log.WithRequestId(c).Info(
					"%s %s %d %s %s resp: `%s`",
					reqMethod,
					reqPath,
					statusCode,
					execTime,
					clientIP,
					rp,
				)
			}
		}
	}
}
