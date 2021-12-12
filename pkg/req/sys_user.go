package req

import "github.com/golang-module/carbon"

type UserStatus struct {
	Username   string `json:"username" form:"username"`
	Wrong      int
	Locked     uint
	LockExpire int64
}

type UserNeedCaptcha struct {
	Wrong int
}

type UserNeedResetPwd struct {
	First         uint                `json:"first"`
	LastLoginTime carbon.ToDateString `json:"lastLoginTime"`
}

type ResetUserPwd struct {
	Username    string `json:"username" form:"username"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type LoginCheck struct {
	Username      string `json:"username" form:"username"`
	Password      string `json:"password" form:"password"`
	CaptchaId     string `json:"captchaId" form:"captchaId"`
	CaptchaAnswer string `json:"captchaAnswer" form:"captchaAnswer"`
}
