package captcha

import (
	"github.com/mojocn/base64Captcha"
)

type Captcha struct {
	ops   Options
	store base64Captcha.Store
	c     *base64Captcha.Captcha
}

func New(options ...func(*Options)) *Captcha {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	ca := &Captcha{
		ops: *ops,
	}
	ca.store = NewStore(options...)
	ca.c = base64Captcha.NewCaptcha(base64Captcha.DefaultDriverDigit, ca.store)
	return ca
}

func (ca Captcha) Get() (id, img string) {
	var err error
	id, img, err = ca.c.Generate()
	if err != nil {
		ca.ops.logger.Warn("get captcha failed: %v", err)
	}
	return
}

func (ca Captcha) Verify(id, answer string) bool {
	if answer == "" {
		return false
	}
	return ca.c.Verify(id, answer, true)
}
