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

func GetMenuTree(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		u := ops.getCurrentUser(c)
		oldCache, ok := CacheGetMenuTree(c, u.Id, *ops)
		if ok {
			resp.SuccessWithData(oldCache)
			return
		}

		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		list, err := q.GetMenuTree(u.RoleId)
		resp.CheckErr(err)
		var rp []resp.MenuTreeResp
		utils.Struct2StructByJson(list, &rp)
		CacheSetMenuTree(c, u.Id, rp, *ops)
		resp.SuccessWithData(rp)
	}
}

func FindMenuByRoleId(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		id := req.UintId(c)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		list := make([]ms.SysMenu, 0)
		ids := make([]uint, 0)
		var err error
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list, ids, err = rd.FindMenuByRoleId(u.RoleId, u.RoleSort, id)
		default:
			my := query.NewMySql(ops.dbOps...)
			list, ids, err = my.FindMenuByRoleId(u.RoleId, u.RoleSort, id)
		}
		resp.CheckErr(err)
		var rp resp.MenuTreeWithAccessResp
		rp.AccessIds = ids
		utils.Struct2StructByJson(list, &rp.List)
		resp.SuccessWithData(rp)
	}
}

func FindMenu(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		list := make([]ms.SysMenu, 0)
		switch ops.cache {
		case true:
			rd := query.NewRedis(ops.cacheOps...)
			list = rd.FindMenu(u.RoleId, u.RoleSort)
		default:
			my := query.NewMySql(ops.dbOps...)
			list = my.FindMenu(u.RoleId, u.RoleSort)
		}
		var rp []resp.MenuTreeResp
		utils.Struct2StructByJson(list, &rp)
		resp.SuccessWithData(rp)
	}
}

func CreateMenu(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		u := ops.getCurrentUser(c)
		var r req.CreateMenuReq
		req.ShouldBind(c, &r)
		req.Validate(c, r, r.FieldTrans())
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateMenu(u.RoleId, u.RoleSort, &r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateMenuById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.UpdateMenuReq
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateById(id, r, new(ms.SysMenu))
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateMenuByRoleId(options ...func(*Options)) gin.HandlerFunc {
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
		err := q.UpdateMenuByRoleId(u.RoleId, u.RoleSort, u.PathRoleId, r)
		resp.CheckErr(err)
		CacheFlushMenuTree(c, *ops)
		resp.Success()
	}
}

func BatchDeleteMenuByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteByIds(r.Uints(), new(ms.SysMenu))
		resp.CheckErr(err)
		resp.Success()
	}
}
