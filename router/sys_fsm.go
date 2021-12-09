package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Fsm() {
	router1 := rt.Casbin("/fsm")
	router2 := rt.CasbinAndIdempotence("/fsm")
	router1.GET("/list", v1.FindFsm(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateFsm(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateFsmById(rt.ops.v1Ops...))
	router1.GET("/approving/list", v1.FindFsmApprovingLog(rt.ops.v1Ops...))
	router1.GET("/log/track", v1.FindFsmLogTrack(rt.ops.v1Ops...))
	router1.GET("/submitter/detail", v1.GetFsmSubmitterDetail(rt.ops.v1Ops...))
	router1.PATCH("/submitter/detail", v1.UpdateFsmSubmitterDetail(rt.ops.v1Ops...))
	router1.PATCH("/approve", v1.FsmApproveLog(rt.ops.v1Ops...))
	router1.PATCH("/cancel", v1.FsmCancelLogByUuids(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteFsmByIds(rt.ops.v1Ops...))
}
