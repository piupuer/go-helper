package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
			ops.logger.Debug(
				c,
				"%s %s %d %s %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
				clientIP,
			)
		} else {
			ops.logger.Info(
				c,
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
