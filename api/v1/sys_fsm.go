package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

// FindFsm
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description FindFsm
// @Param params query req.FsmMachine true "params"
// @Router /fsm/list [GET]
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

// FindFsmApprovingLog
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description FindFsmApprovingLog
// @Param params query req.FsmPendingLog true "params"
// @Router /fsm/approving/list [GET]
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
		r.ApprovalRoleId = u.RoleId
		r.ApprovalUserId = u.Id
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

// FindFsmLogTrack
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description FindFsmLogTrack
// @Param params query req.FsmLog true "params"
// @Router /fsm/log/track [GET]
func FindFsmLogTrack(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		var r req.FsmLog
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		item, err := q.FindFsmLogTrack(r)
		resp.CheckErr(err)
		resp.SuccessWithData(item)
	}
}

// GetFsmSubmitterDetail
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description GetFsmSubmitterDetail
// @Param params query req.FsmSubmitterDetail true "params"
// @Router /fsm/submitter/detail [GET]
func GetFsmSubmitterDetail(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getFsmSubmitterDetail == nil {
		panic("getFsmSubmitterDetail is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmSubmitterDetail
		req.ShouldBind(c, &r)
		resp.SuccessWithData(ops.getFsmSubmitterDetail(c, r))
	}
}

// UpdateFsmSubmitterDetail
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description UpdateFsmSubmitterDetail
// @Param params body req.UpdateFsmSubmitterDetail true "params"
// @Router /fsm/submitter/detail [PATCH]
func UpdateFsmSubmitterDetail(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	if ops.updateFsmSubmitterDetail == nil {
		panic("updateFsmSubmitterDetail is empty")
	}
	return func(c *gin.Context) {
		var r req.UpdateFsmSubmitterDetail
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		r.Parse()
		err := q.FsmCheckEditLogDetailPermission(req.FsmCheckEditLogDetailPermission{
			Category:       r.Category,
			Uuid:           r.Uuid,
			ApprovalRoleId: u.RoleId,
			ApprovalUserId: u.Id,
			Fields:         r.Keys,
			Approver:       true,
		})
		resp.CheckErr(err)
		err = ops.updateFsmSubmitterDetail(c, r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// FsmApproveLog
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description FsmApproveLog
// @Param params query req.FsmApproveLog true "params"
// @Router /fsm/approve [PATCH]
func FsmApproveLog(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmApproveLog
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ApprovalRoleId = u.RoleId
		r.ApprovalUserId = u.Id
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		item, err := q.FsmApproveLog(r)
		resp.CheckErr(err)
		resp.SuccessWithData(item)
	}
}

// FsmCancelLogByUuids
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description FsmCancelLogByUuids
// @Param params query req.FsmCancelLog true "params"
// @Router /fsm/cancel [PATCH]
func FsmCancelLogByUuids(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.FsmCancelLog
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		r.ApprovalRoleId = u.RoleId
		r.ApprovalUserId = u.Id
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.FsmCancelLogByUuids(r)
		resp.CheckErr(err)
		resp.Success()
	}
}

// CreateFsm
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description CreateFsm
// @Param params body req.FsmCreateMachine true "params"
// @Router /fsm/create [POST]
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

// UpdateFsmById
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description UpdateFsmById
// @Param id path uint true "id"
// @Param params body req.FsmUpdateMachine true "params"
// @Router /fsm/update/{id} [PATCH]
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

// BatchDeleteFsmByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Fsm
// @Description BatchDeleteFsmByIds
// @Param ids body req.Ids true "ids"
// @Router /fsm/delete/batch [DELETE]
func BatchDeleteFsmByIds(options ...func(*Options)) gin.HandlerFunc {
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
