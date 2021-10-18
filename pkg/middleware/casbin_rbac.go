package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/resp"
	"strings"
	"sync"
)

func Casbin(options ...func(*CasbinOptions)) gin.HandlerFunc {
	ops := getCasbinOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.enforcer == nil {
		panic("casbin enforcer handler is empty")
	}
	return func(c *gin.Context) {
		// get role.key as subject
		sub := ops.roleKey(c)
		// request path as object
		obj := strings.Replace(c.Request.URL.Path, "/"+ops.urlPrefix, "", 1)
		// request method as action
		act := c.Request.Method
		if !check(sub, obj, act, *ops) {
			ops.failWithCode(resp.Forbidden)
			return
		}
		c.Next()
	}
}

var checkLock sync.Mutex

func check(sub, obj, act string, ops CasbinOptions) bool {
	checkLock.Lock()
	defer checkLock.Unlock()
	pass, _ := ops.enforcer.Enforce(sub, obj, act)
	return pass
}
