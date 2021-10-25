package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Machine() {
	router1 := rt.Casbin("/machine")
	router2 := rt.CasbinAndIdempotence("/machine")
	router1.GET("/list", v1.FindMachine(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateMachine(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateMachineById(rt.ops.v1Ops...))
	router1.PATCH("/connect/:id", v1.ConnectMachineById(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteMachineByIds(rt.ops.v1Ops...))
}
