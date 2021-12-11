package router

import (
	v1 "github.com/piupuer/go-helper/api/v1"
	"github.com/piupuer/go-helper/pkg/middleware"
)

func (rt Router) Base() {
	if rt.ops.jwt {
		router1 := rt.ops.group.Group("/base")
		router2 := rt.Casbin("/base")
		router1.POST("/user/status", v1.GetUserStatus(rt.ops.v1Ops...))
		router1.POST("/login", middleware.JwtLogin(rt.ops.jwtOps...))
		router1.POST("/logout", middleware.JwtLogout(rt.ops.jwtOps...))
		router1.POST("/refreshToken", middleware.JwtRefresh(rt.ops.jwtOps...))
		if rt.ops.idempotence {
			// need login
			router2.GET("/idempotenceToken", middleware.GetIdempotenceToken(rt.ops.idempotenceOps...))
			router2.POST("/user/reset/pwd", v1.ResetUserPwd(rt.ops.v1Ops...))
		}
	}
}
