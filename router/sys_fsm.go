package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Fsm() {
	router1 := rt.Casbin("/fsm")
	router2 := rt.CasbinAndIdempotence("/fsm")
	router1.GET("/list", v1.FindFsm(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateFsm(rt.ops.v1Ops...))
	router1.GET("/approving/list", v1.FindFsmApprovingLog(rt.ops.v1Ops...))
}
