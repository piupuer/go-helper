package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

// GetUserStatus
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Base
// @Description GetUserStatus
// @Param params body req.UserStatus true "params"
// @Router /base/user/status [POST]
func GetUserStatus(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getUserLoginStatus == nil {
		panic("getUserLoginStatus is empty")
	}
	return func(c *gin.Context) {
		var r req.UserStatus
		req.ShouldBind(c, &r)
		err := ops.getUserLoginStatus(c, &r)
		resp.CheckErr(err)
		ops.addCtx(c)
		my := query.NewMySql(ops.dbOps...)
		resp.SuccessWithData(my.GetUserStatus(r))
	}
}

// ResetUserPwd
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Base
// @Description ResetUserPwd
// @Param params body req.ResetUserPwd true "params"
// @Router /base/user/reset [PATCH]
func ResetUserPwd(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		var r req.ResetUserPwd
		req.ShouldBind(c, &r)
		u := ops.getCurrentUser(c)
		if u.RoleSort != constant.Zero && r.Username != u.Username {
			resp.CheckErr(resp.ForbiddenMsg)
		}
		if ops.beforeResetUserPwd != nil {
			err := ops.beforeResetUserPwd(c, &r)
			resp.CheckErr(err)
		}
		ops.addCtx(c)
		my := query.NewMySql(ops.dbOps...)
		err := my.ResetUserPwd(r)
		resp.CheckErr(err)
		resp.Success()
	}
}
