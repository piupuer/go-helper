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
	// router auto transmit children
	if ops.casbin {
		cabinOps := middleware.ParseCasbinOptions(ops.casbinOps...)
		if cabinOps.Enforcer != nil {
			ops.v1Ops = append(
				ops.v1Ops,
				v1.WithDbOps(
					query.WithMysqlCasbinEnforcer(cabinOps.Enforcer),
				),
				v1.WithBinlogOps(
					query.WithRedisCasbinEnforcer(cabinOps.Enforcer),
				))
		}
	}
	if ops.redis != nil {
		ops.idempotenceOps = append(ops.idempotenceOps, middleware.WithIdempotenceRedis(ops.redis))
		ops.v1Ops = append(ops.v1Ops, v1.WithRedis(ops.redis))
	}
	ops.v1Ops = append(ops.v1Ops, v1.WithBinlog(ops.redisBinlog))
	ops.v1Ops = append(ops.v1Ops, v1.WithMessageHubOps(
		query.WithMessageHubIdempotence(ops.idempotence),
		query.WithMessageHubIdempotenceOps(ops.idempotenceOps...),
	))
	r := &Router{
		ops: *ops,
	}
	return r
}

func (rt Router) Group(path string) gin.IRoutes {
	r := rt.ops.group.Group(path)
	return r
}

func (rt Router) Casbin(path string) gin.IRoutes {
	r := rt.Group(path)
	if rt.ops.jwt {
		r.Use(middleware.Jwt(rt.ops.jwtOps...))
	}
	if rt.ops.casbin {
		r.Use(middleware.Casbin(rt.ops.casbinOps...))
	}
	return r
}

func (rt Router) Idempotence(path string) gin.IRoutes {
	r := rt.Group(path)
	if rt.ops.idempotence {
		r.Use(middleware.Idempotence(rt.ops.idempotenceOps...))
	}
	return r
}

func (rt Router) CasbinAndIdempotence(path string) gin.IRoutes {
	r := rt.Casbin(path)
	if rt.ops.jwt {
		r.Use(middleware.Jwt(rt.ops.jwtOps...))
	}
	if rt.ops.casbin {
		r.Use(middleware.Casbin(rt.ops.casbinOps...))
	}
	if rt.ops.idempotence {
		r.Use(middleware.Idempotence(rt.ops.idempotenceOps...))
	}
	return r
}
