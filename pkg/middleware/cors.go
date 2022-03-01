package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/utils"
	"net/http"
	"strings"
)

func Cors(options ...func(*CorsOptions)) gin.HandlerFunc {
	ops := getCorsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		methods := strings.Split(ops.method, ",")
		if !utils.Contains(methods, method) {
			c.Status(http.StatusMethodNotAllowed)
			c.Abort()
			return
		}
		c.Header("Access-Control-Allow-Origin", ops.origin)
		c.Header("Access-Control-Allow-Headers", ops.header)
		c.Header("Access-Control-Allow-Methods", ops.method)
		c.Header("Access-Control-Expose-Headers", ops.expose)
		c.Header("Access-Control-Allow-Credentials", ops.credential)

		// skip OPTIONS
		if method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
