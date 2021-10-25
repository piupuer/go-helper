package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Menu() {
	router1 := rt.Casbin("/menu")
	router2 := rt.CasbinAndIdempotence("/menu")
	router1.GET("/tree", v1.GetMenuTree(rt.ops.v1Ops...))
	router1.GET("/all/:id", v1.FindMenuByRoleId(rt.ops.v1Ops...))
	router1.GET("/list", v1.FindMenu(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateMenu(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateMenuById(rt.ops.v1Ops...))
	router1.PATCH("/role/update/:id", v1.UpdateMenuByRoleId(rt.ops.v1Ops...))
	router2.DELETE("/delete/batch", v1.BatchDeleteMenuByIds(rt.ops.v1Ops...))
}
