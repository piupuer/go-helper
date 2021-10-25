package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

func FindApi(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.ApiReq
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		list := make([]ms.SysApi, 0)
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list = rd.FindApi(&r)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindApi(&r)
		}
		resp.SuccessWithPageData(list, []resp.ApiResp{}, r.Page)
	}
}

func FindApiGroupByCategoryByRoleKeyword(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.ApiReq
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		list := make([]resp.ApiGroupByCategoryResp, 0)
		ids := make([]uint, 0)
		var err error
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list, ids, err = rd.FindApiGroupByCategoryByRoleKeyword(u.RoleKeyword, u.PathRoleKeyword)
		default:
			my := query.NewMySql(ops.dbOps...)
			list, ids, err = my.FindApiGroupByCategoryByRoleKeyword(u.RoleKeyword, u.PathRoleKeyword)
		}
		resp.CheckErr(err)
		var rp resp.ApiTreeWithAccessResp
		rp.AccessIds = ids
		utils.Struct2StructByJson(list, &rp.List)
		resp.SuccessWithData(rp)
	}
}

func CreateApi(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.findRoleKeywordByRoleIds == nil {
		panic("findRoleKeywordByRoleIds is empty")
	}
	return func(c *gin.Context) {
		var r req.CreateApiReq
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

func UpdateApiById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateApiReq
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateApiById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateApiByRoleKeyword(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.UpdateMenuIncrementalIdsReq
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

func BatchDeleteApiByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteApiByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}
