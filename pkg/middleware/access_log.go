package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

		if reqMethod == "OPTIONS" || reqPath == fmt.Sprintf("/%s/ping", ops.urlPrefix) {
			log.WithRequestId(c).Debug(
				"%s %s %d %s %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
				clientIP,
			)
		} else {
			log.WithRequestId(c).Info(
				"%s %s %d %s %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
				clientIP,
			)
		}
	}
}
