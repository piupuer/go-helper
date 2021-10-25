package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

func FindMachine(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.MachineReq
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysMachine, 0)
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list = rd.FindMachine(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindMachine(&r)
		}
		resp.SuccessWithPageData(list, []resp.MachineResp{}, r.Page)
	}
}

func CreateMachine(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.CreateMachineReq
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.Create(r, new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateMachineById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateMachineReq
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateById(id, r, new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}

func ConnectMachineById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.ConnectMachine(id)
		resp.CheckErr(err)
		resp.Success()
	}
}

func BatchDeleteMachineByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteByIds(r.Uints(), new(ms.SysMachine))
		resp.CheckErr(err)
		resp.Success()
	}
}
