package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/utils"
	"net/http"
	"strings"
)

var methods = []string{
	http.MethodOptions,
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

var methodStr = strings.Join(methods, ", ")

func Cors(c *gin.Context) {
	method := c.Request.Method
	if !utils.Contains(methods, method) {
		c.Status(http.StatusMethodNotAllowed)
		c.Abort()
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, api-idempotence-token")
	c.Header("Access-Control-Allow-Methods", methodStr)
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")

	// skip OPTIONS
	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	}
	c.Next()
}
