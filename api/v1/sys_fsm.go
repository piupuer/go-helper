package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
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
	if ops.findRoleByIds == nil {
		panic("findRoleByIds is empty")
	}
	if ops.findUserByIds == nil {
		panic("findUserByIds is empty")
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
		roleIds := make([]uint, 0)
		for _, item := range list {
			roleIds = append(roleIds, item.SubmitterRoleId)
			for _, u := range item.CanApprovalRoles {
				roleIds = append(roleIds, u.Id)
			}
		}
		roles := ops.findRoleByIds(c, roleIds)
		newRoles := make([]resp.Role, len(roles))
		utils.Struct2StructByJson(roles, &newRoles)
		m1 := make(map[uint]resp.Role)
		for _, role := range newRoles {
			m1[role.Id] = role
		}
		for i, item := range list {
			list[i].SubmitterRole = m1[item.SubmitterRoleId]
			for j, u := range item.CanApprovalRoles {
				list[i].CanApprovalRoles[j] = m1[u.Id]
			}
		}
		userIds := make([]uint, 0)
		for _, item := range list {
			userIds = append(userIds, item.SubmitterUserId)
			for _, u := range item.CanApprovalUsers {
				userIds = append(userIds, u.Id)
			}
		}
		users := ops.findUserByIds(c, userIds)
		newUsers := make([]resp.User, len(users))
		utils.Struct2StructByJson(users, &newUsers)
		m2 := make(map[uint]resp.User)
		for _, user := range newUsers {
			m2[user.Id] = user
		}
		for i, item := range list {
			list[i].SubmitterUser = m2[item.SubmitterUserId]
			for j, u := range item.CanApprovalUsers {
				list[i].CanApprovalUsers[j] = m2[u.Id]
			}
		}
		resp.SuccessWithPageData(list, &[]resp.FsmApprovingLog{}, r.Page)
	}
}

func GetFsmSubmitterDetail(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getFsmSubmitterDetail == nil {
		panic("getFsmDetail is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmSubmitterDetail
		req.ShouldBind(c, &r)
		resp.SuccessWithData(ops.getFsmSubmitterDetail(c, r))
	}
}

func FsmApproveLog(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.fsmTransition == nil {
		panic("fsmTransition is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmApproveLog
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ApprovalRoleId = u.Id
		r.ApprovalUserId = u.RoleId
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		item, err := q.FsmApproveLog(r)
		resp.CheckErr(err)
		err = ops.fsmTransition(c, *item)
		resp.CheckErr(err)
		resp.SuccessWithData(item)
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
