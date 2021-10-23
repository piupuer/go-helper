package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/piupuer/go-helper/api/v1"
)

func (rt Router) Dict() gin.IRoutes {
	router1 := rt.Casbin("/dict")
	router2 := rt.CasbinAndIdempotence("/dict")
	router1.GET("/list", v1.FindDict(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateDict(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateDictById(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteDictByIds(rt.ops.v1Ops...))
	router1.GET("/data/list", v1.FindDictData(rt.ops.v1Ops...))
	router2.POST("/data/create", v1.CreateDictData(rt.ops.v1Ops...))
	router1.PATCH("/data/update/:id", v1.UpdateDictDataById(rt.ops.v1Ops...))
	router1.DELETE("/data/delete/batch", v1.BatchDeleteDictDataByIds(rt.ops.v1Ops...))
	return rt.ops.group
}
