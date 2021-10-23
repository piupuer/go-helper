package router

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/middleware"
)

type Router struct {
	ops Options
}

func NewRouter(options ...func(*Options)) *Router {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	r := &Router{
		ops: *ops,
	}
	return r
}

// get casbin middleware router
func (rt Router) Casbin(path string) gin.IRoutes {
	r := rt.ops.group.Group(path)
	if rt.ops.jwt {
		r.Use(middleware.Jwt(rt.ops.jwtOps...))
	}
	if rt.ops.casbin {
		r.Use(middleware.Casbin(rt.ops.casbinOps...))
	}
	return r
}

// get casbin and idempotence middleware router
func (rt Router) CasbinAndIdempotence(path string) gin.IRoutes {
	r := rt.Casbin(path)
	if rt.ops.idempotence {
		r.Use(middleware.Idempotence(rt.ops.idempotenceOps...))
	}
	return r
}
