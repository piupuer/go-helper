package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
)

// FindApi
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description FindApi
// @Param params query req.Api true "params"
// @Router /api/list [GET]
func FindApi(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "FindApi"))
		defer span.End()
		var r req.Api
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysApi, 0)
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list = rd.FindApi(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindApi(&r)
		}
		resp.SuccessWithPageData(list, &[]resp.Api{}, r.Page)
	}
}

// FindApiGroupByCategoryByRoleKeyword
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description FindApiGroupByCategoryByRoleKeyword
// @Param id path uint true "id"
// @Param params query req.Api true "params"
// @Router /api/all/category/{id} [GET]
func FindApiGroupByCategoryByRoleKeyword(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "FindApiGroupByCategoryByRoleKeyword"))
		defer span.End()
		var r req.Api
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		list := make([]resp.ApiGroupByCategory, 0)
		ids := make([]uint, 0)
		var err error
		switch ops.binlog {
		case true:
			rd := query.NewRedis(ops.binlogOps...)
			list, ids, err = rd.FindApiGroupByCategoryByRoleKeyword(u.RoleKeyword, u.PathRoleKeyword)
		default:
			my := query.NewMySql(ops.dbOps...)
			list, ids, err = my.FindApiGroupByCategoryByRoleKeyword(u.RoleKeyword, u.PathRoleKeyword)
		}
		resp.CheckErr(err)
		var rp resp.ApiTreeWithAccess
		rp.AccessIds = ids
		utils.Struct2StructByJson(list, &rp.List)
		resp.SuccessWithData(rp)
	}
}

// CreateApi
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description CreateApi
// @Param params body req.CreateApi true "params"
// @Router /api/create [POST]
func CreateApi(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.findRoleKeywordByRoleIds == nil {
		panic("findRoleKeywordByRoleIds is empty")
	}
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "CreateApi"))
		defer span.End()
		var r req.CreateApi
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		r.RoleKeywords = ops.findRoleKeywordByRoleIds(c, r.RoleIds)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateApi(&r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// UpdateApiById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description UpdateApiById
// @Param id path uint true "id"
// @Param params body req.UpdateApi true "params"
// @Router /api/update/{id} [PATCH]
func UpdateApiById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "UpdateApiById"))
		defer span.End()
		var r req.UpdateApi
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateApiById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// UpdateApiByRoleId
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description UpdateApiByRoleId
// @Param id path uint true "id"
// @Param params body req.UpdateMenuIncrementalIds true "params"
// @Router /api/role/update/{id} [PATCH]
func UpdateApiByRoleId(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "UpdateApiByRoleId"))
		defer span.End()
		var r req.UpdateMenuIncrementalIds
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		if u.RoleId == u.PathRoleId {
			if u.RoleSort == constant.Zero && len(r.Delete) > 0 {
				resp.CheckErr("cannot remove super admin privileges")
			} else if u.RoleSort != constant.Zero {
				resp.CheckErr("cannot change your permissions")
			}
		}

		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateApiByRoleKeyword(u.PathRoleKeyword, r)
		resp.CheckErr(err)
		CacheFlushMenuTree(c, *ops)
		resp.Success()
	}
}

// BatchDeleteApiByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Api
// @Description BatchDeleteApiByIds
// @Param ids body req.Ids true "ids"
// @Router /api/delete/batch [DELETE]
func BatchDeleteApiByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "BatchDeleteApiByIds"))
		defer span.End()
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteApiByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}
