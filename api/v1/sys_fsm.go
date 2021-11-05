package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

func FindFsm(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.FsmMachine
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		list, err := q.FindFsm(&r)
		resp.CheckErr(err)
		resp.SuccessWithPageData(list, &[]resp.FsmMachine{}, r.Page)
	}
}

func FindFsmApprovingLog(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmPendingLog
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ApprovalRoleId = u.Id
		r.ApprovalUserId = u.RoleId
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		list, err := q.FindFsmApprovingLog(&r)
		resp.CheckErr(err)
		resp.SuccessWithPageData(list, &[]resp.FsmApprovingLog{}, r.Page)
	}
}

func CreateFsm(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.FsmCreateMachine
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.CreateFsm(r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func UpdateFsmById(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.FsmUpdateMachine
		req.ShouldBind(c, &r)
		id := req.UintId(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.UpdateFsmById(id, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

func DeleteFsmByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteFsmByIds(r.Uints())
		resp.CheckErr(err)
		resp.Success()
	}
}
