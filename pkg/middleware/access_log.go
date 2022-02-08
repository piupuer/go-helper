package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
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

		detail := make(map[string]interface{})
		if ops.detail {
			detail = getRequestDetail(c)
		}

		detail[constant.MiddlewareAccessLogIpLogKey] = clientIP

		l := log.WithRequestId(c).WithFields(detail)

		if reqMethod == "OPTIONS" || reqPath == fmt.Sprintf("/%s/ping", ops.urlPrefix) {
			l.Debug(
				"%s %s %d %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
			)
		} else {
			l.Info(
				"%s %s %d %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
			)
		}
	}
}
