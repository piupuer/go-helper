package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Upload() {
	router1 := rt.Casbin("/upload")
	router1.GET("/file", v1.UploadFileChunkExists(rt.ops.v1Ops...))
	router1.POST("/file", v1.UploadFile(rt.ops.v1Ops...))
	router1.POST("/merge", v1.UploadMerge(rt.ops.v1Ops...))
	router1.POST("/unzip", v1.UploadUnZip(rt.ops.v1Ops...))
}
