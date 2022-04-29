package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
)

// FindMachine
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Machine
// @Description FindMachine
// @Param params query req.Machine true "params"
// @Router /machine/list [GET]
func FindMachine(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "FindMachine"))
		defer span.End()
		var r req.Machine
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysMachine, 0)
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list = rd.FindMachine(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindMachine(&r)
		}
		resp.SuccessWithPageData(list, &[]resp.Machine{}, r.Page)
	}
}

// CreateMachine
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Machine
// @Description CreateMachine
// @Param params body req.CreateMachine true "params"
// @Router /machine/create [POST]
func CreateMachine(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "CreateMachine"))
		defer span.End()
		var r req.CreateMachine
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.Create(r, new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}

// UpdateMachineById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Machine
// @Description UpdateMachineById
// @Param id path uint true "id"
// @Param params body req.UpdateMachine true "params"
// @Router /machine/update/{id} [PATCH]
func UpdateMachineById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "UpdateMachineById"))
		defer span.End()
		var r req.UpdateMachine
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateById(id, r, new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}

// ConnectMachineById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Machine
// @Description ConnectMachineById
// @Param id path uint true "id"
// @Router /machine/connect/{id} [PATCH]
func ConnectMachineById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "ConnectMachineById"))
		defer span.End()
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.ConnectMachine(id)
		resp.CheckErr(err)
		resp.Success()
	}
}

// BatchDeleteMachineByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Machine
// @Description BatchDeleteMachineByIds
// @Param ids body req.Ids true "ids"
// @Router /machine/delete/batch [DELETE]
func BatchDeleteMachineByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "BatchDeleteMachineByIds"))
		defer span.End()
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteByIds(r.Uints(), new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}
