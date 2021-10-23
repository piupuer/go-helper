package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/piupuer/go-helper/api/v1"
)

func (rt Router) Api() gin.IRoutes {
	router1 := rt.Casbin("/api")
	router2 := rt.CasbinAndIdempotence("/api")
	router1.GET("/list", v1.FindApi(rt.ops.v1Ops...))
	router1.GET("/all/category/:id", v1.FindApiGroupByCategoryByRoleKeyword(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateApi(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateApiById(rt.ops.v1Ops...))
	router1.PATCH("/role/update/:id", v1.UpdateApiByRoleKeyword(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteApiByIds(rt.ops.v1Ops...))
	return rt.ops.group
}
