package query

import (
	"fmt"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/captcha"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"time"
)

func (my MySql) GetUserStatus(r req.UserStatus) (rp resp.UserStatus) {
	timestamp := time.Now().Unix()
	flag := my.UserNeedCaptcha(req.UserNeedCaptcha{
		Wrong: r.Wrong,
	})
	if flag {
		rp.Captcha = my.GetCaptcha()
	}
	if r.Locked == constant.One && (r.LockExpire == 0 || timestamp < r.LockExpire) {
		rp.Locked = r.Locked
	}
	return
}

func (my MySql) UserNeedCaptcha(r req.UserNeedCaptcha) (flag bool) {
	d := my.GetDictData(constant.UserLoginDict, constant.UserLoginCaptcha)
	if d.Val != "" {
		if r.Wrong >= utils.Str2Int(d.Val) {
			flag = true
		}
	}
	return
}

func (my MySql) GetCaptcha() (rp resp.Captcha) {
	rp.Id, rp.Img = captcha.New(
		captcha.WithRedis(my.ops.redis),
		captcha.WithCtx(my.Ctx),
	).Get()
	return rp
}

func (my MySql) VerifyCaptcha(r req.LoginCheck) bool {
	return captcha.New(
		captcha.WithRedis(my.ops.redis),
		captcha.WithCtx(my.Ctx),
	).Verify(r.CaptchaId, r.CaptchaAnswer)
}

func (my MySql) UserNeedResetPwd(r req.UserNeedResetPwd) (flag bool) {
	if r.First == constant.One {
		d1 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdFirstLogin)
		if d1.Val == fmt.Sprintf("%v", constant.One) {
			flag = true
			return
		}
	}
	if !r.LastLoginTime.IsZero() {
		d2 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdAfterSomeTime)
		if d2.Val != "" {
			var t carbon.Carbon
			now := carbon.Now()
			switch d2.Addition {
			case constant.UserResetPwdAfterSomeTimeAdditionDuration:
				t = now.SubDuration(d2.Val)
			case constant.UserResetPwdAfterSomeTimeAdditionMonth:
				t = now.SubMonths(utils.Str2Int(d2.Val))
			case constant.UserResetPwdAfterSomeTimeAdditionYear:
				t = now.SubYears(utils.Str2Int(d2.Val))
			}
			if t.Gt(r.LastLoginTime.Carbon) {
				flag = true
			}
		}
	}
	return
}

func (my MySql) ResetUserPwd(r req.ResetUserPwd) error {
	pass, msg := my.CheckWeakPwd(r.NewPassword)
	if !pass {
		if msg == "" {
			msg = resp.WeakPassword
		} else {
			msg = fmt.Sprintf("%s: %s", resp.WeakPassword, msg)
		}
		return errors.Errorf(msg)
	}
	return my.Tx.
		Table(my.Tx.NamingStrategy.TableName("sys_user")).
		Where("username = ?", r.Username).
		Update("password", utils.GenPwd(r.NewPassword)).
		Error
}

func (my MySql) CheckWeakPwd(pwd string) (pass bool, msg string) {
	// check weak password
	d1 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdWeakLen)
	if d1.Val != "" {
		if len(pwd) < utils.Str2Int(d1.Val) {
			msg = d1.Addition
			return
		}
	}
	d2 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdWeakContainsChinese)
	if d2.Val != "" {
		if utils.StrContainsChinese(pwd) {
			msg = d2.Addition
			return
		}
	}
	d3 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdWeakCaseSensitive)
	if d3.Val == fmt.Sprintf("%d", constant.One) {
		if strings.ToLower(pwd) == pwd {
			msg = d3.Addition
			return
		}
	}
	d4 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdWeakSpecialChar)
	if d4.Val != "" {
		matched, _ := regexp.MatchString(d4.Val, pwd)
		if !matched {
			msg = d4.Addition
			return
		}
	}
	d5 := my.GetDictData(constant.UserResetPwdDict, constant.UserResetPwdWeakContinuousNum)
	if d5.Val != "" {
		if utils.StrContainsContinuousNum(pwd) >= utils.Str2Int(d5.Val) {
			msg = d5.Addition
			return
		}
	}
	pass = true
	return
}
