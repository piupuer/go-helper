package req

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	enTranslations "gopkg.in/go-playground/validator.v9/translations/en"
	cnTranslations "gopkg.in/go-playground/validator.v9/translations/zh"
)

type ValidateOptions struct {
	validator  *validator.Validate
	translator ut.Translator
}

func WithValidateValidator(v *validator.Validate) func(*ValidateOptions) {
	return func(options *ValidateOptions) {
		if v != nil {
			getValidateOptionsOrSetDefault(options).validator = v
		}
	}
}

func WithValidateTranslator(trans ut.Translator) func(*ValidateOptions) {
	return func(options *ValidateOptions) {
		if trans != nil {
			getValidateOptionsOrSetDefault(options).translator = trans
		}
	}
}

func WithValidateCn(options *ValidateOptions) {
	options = getValidateOptionsOrSetDefault(options)
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	validate := validator.New()

	_ = cnTranslations.RegisterDefaultTranslations(validate, trans)
	options.validator = validate
	options.translator = trans
}

func getValidateOptionsOrSetDefault(options *ValidateOptions) *ValidateOptions {
	if options == nil {
		e := en.New()
		uni := ut.New(e, e)
		trans, _ := uni.GetTranslator("en")
		v := validator.New()
		_ = enTranslations.RegisterDefaultTranslations(v, trans)
		return &ValidateOptions{
			validator:  v,
			translator: trans,
		}
	}
	return options
}
