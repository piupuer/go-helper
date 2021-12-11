package req

import "github.com/golang-module/carbon"

type UserStatus struct {
	Username string `json:"username"`
	Wrong    int
	Locked   uint
}

type UserNeedResetPwd struct {
	First         uint                `json:"first"`
	LastLoginTime carbon.ToDateString `json:"lastLoginTime"`
}

type ResetUserPwd struct {
	Username    string `json:"username" form:"username"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}
