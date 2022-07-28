package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/delay"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
)

// FindDelayExport
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Delay
// @Description FindDelayExport
// @Param params query req.DelayExportHistory true "params"
// @Router /delay/export/list [GET]
func FindDelayExport(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "FindDelayExport"))
		defer span.End()
		var r req.DelayExportHistory
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		ex := delay.NewExport(ops.exportOps...)
		list, err := ex.FindHistory(&r)
		resp.CheckErr(err)
		resp.SuccessWithPageData(list, &[]resp.DelayExportHistory{}, r.Page)
	}
}

// BatchDeleteDelayExportByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Delay
// @Description BatchDeleteDelayExportByIds
// @Param ids body req.Ids true "ids"
// @Router /delay/export/delete/batch [DELETE]
func BatchDeleteDelayExportByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "BatchDeleteDelayExportByIds"))
		defer span.End()
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		ex := delay.NewExport(ops.exportOps...)
		err := ex.DeleteHistoryByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}
