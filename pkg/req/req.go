package req

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/resp"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

// bind request param
func ShouldBind(c *gin.Context, req interface{}) {
	err := c.ShouldBind(req)
	if err != nil {
		resp.FailWithMsg("%s: %v", resp.InvalidParameterMsg, err)
	}
}

// validate request param
func Validate(c *gin.Context, req interface{}, trans map[string]string, options ...func(*ValidateOptions)) {
	ops := getValidateOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	err := validate(ops.validator.Struct(req), trans, *ops)
	if err != nil {
		resp.FailWithMsg("%s: %v", resp.IllegalParameterMsg, err)
	}
}

func validate(err error, custom map[string]string, ops ValidateOptions) (e error) {
	if err == nil {
		return
	}
	errs := err.(validator.ValidationErrors)
	for _, item := range errs {
		tranStr := item.Translate(ops.translator)
		names := strings.Split(item.Namespace(), ".")
		// deep names
		if len(names) > 1 {
			if v, ok := custom[strings.Join(names[1:], ".")]; ok {
				return fmt.Errorf(strings.Replace(tranStr, item.Field(), v, 1))
			}
		}
		// check whether it is in custom
		if v, ok := custom[item.Field()]; ok {
			return fmt.Errorf(strings.Replace(tranStr, item.Field(), v, 1))
		} else {
			return fmt.Errorf(tranStr)
		}
	}
	return
}
