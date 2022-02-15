package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Delay() {
	router1 := rt.Casbin("/delay")
	router1.GET("/export/list", v1.FindDelayExport(rt.ops.v1Ops...))
	router1.DELETE("/export/delete/batch", v1.BatchDeleteDelayExportByIds(rt.ops.v1Ops...))
}
