package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

// FindDict
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description FindDict
// @Param params query req.Dict true "params"
// @Router /dict/list [GET]
func FindDict(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Dict
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
		resp.SuccessWithPageData(list, &[]resp.Dict{}, r.Page)
	}
}

// CreateDict
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description CreateDict
// @Param params body req.CreateDict true "params"
// @Router /dict/create [POST]
func CreateDict(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.CreateDict
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateDict(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// UpdateDictById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description UpdateDictById
// @Param id path uint true "id"
// @Param params body req.UpdateDict true "params"
// @Router /dict/update/{id} [PATCH]
func UpdateDictById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateDict
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateDictById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// BatchDeleteDictByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description BatchDeleteDictByIds
// @Param ids body req.Ids true "ids"
// @Router /dict/delete/batch [DELETE]
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

// FindDictData
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description FindDictData
// @Param params query req.DictData true "params"
// @Router /dict/data/list [GET]
func FindDictData(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.DictData
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
		resp.SuccessWithPageData(list, &[]resp.DictData{}, r.Page)
	}
}

// CreateDictData
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description CreateDictData
// @Param params body req.CreateDictData true "params"
// @Router /dict/data/create [POST]
func CreateDictData(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.CreateDictData
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateDictData(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// UpdateDictDataById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description UpdateDictDataById
// @Param id path uint true "id"
// @Param params body req.UpdateDictData true "params"
// @Router /dict/data/update/{id} [PATCH]
func UpdateDictDataById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateDictData
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateDictDataById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// BatchDeleteDictDataByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags Dict
// @Description BatchDeleteDictDataByIds
// @Param ids body req.Ids true "ids"
// @Router /dict/data/delete/batch [DELETE]
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
