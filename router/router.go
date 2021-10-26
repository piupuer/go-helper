package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/piupuer/go-helper/api/v1"
	"github.com/piupuer/go-helper/pkg/middleware"
	"github.com/piupuer/go-helper/pkg/query"
)

type Router struct {
	ops Options
}

func NewRouter(options ...func(*Options)) *Router {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.group == nil {
		panic("group is empty")
	}
	if ops.logger != nil {
		ops.jwtOps = append(ops.jwtOps, middleware.WithJwtLogger(ops.logger))
		ops.v1Ops = append(ops.v1Ops, v1.WithDbOps(
			query.WithMysqlLogger(ops.logger),
		))
		ops.v1Ops = append(ops.v1Ops, v1.WithBinlogOps(
			query.WithRedisLogger(ops.logger),
		))
	}
	if ops.redis != nil {
		ops.idempotenceOps = append(ops.idempotenceOps, middleware.WithIdempotenceRedis(ops.redis))
		ops.v1Ops = append(ops.v1Ops, v1.WithRedis(ops.redis))
	}
	ops.v1Ops = append(ops.v1Ops, v1.WithBinlog(ops.redisBinlog))
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
