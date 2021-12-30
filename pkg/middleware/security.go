package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeader(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Frame-Options", "deny")
	c.Next()
}
