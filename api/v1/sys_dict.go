package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

func FindDict(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.DictReq
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysDict, 0)
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list = rd.FindDict(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindDict(&r)
		}
		resp.SuccessWithPageData(list, []resp.DictResp{}, r.Page)
	}
}

func CreateDict(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.CreateDictReq
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateDict(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateDictById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateDictReq
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateDictById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func BatchDeleteDictByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteDictByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}

func FindDictData(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.DictDataReq
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysDictData, 0)
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list = rd.FindDictData(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindDictData(&r)
		}
		resp.SuccessWithPageData(list, []resp.DictDataResp{}, r.Page)
	}
}

func CreateDictData(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.CreateDictDataReq
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateDictData(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateDictDataById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateDictDataReq
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateDictDataById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func BatchDeleteDictDataByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteDictDataByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}
