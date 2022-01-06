package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Params(options ...func(*ParamsOptions)) gin.HandlerFunc {
	ops := getParamsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		// read body
		var body []byte
		body, err := ioutil.ReadAll(c.Request.Body)
		if err == nil {
			// write back to gin request body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		c.Set(constant.MiddlewareParamsBodyCtxKey, string(body))
		c.Set(constant.MiddlewareParamsQueryCtxKey, c.Request.URL.Query())
		f := make(url.Values)
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			c.Request.ParseForm()
			if c.Request.PostForm != nil {
				f = c.Request.PostForm
			}
		}
		c.Set(constant.MiddlewareParamsFormCtxKey, f)
	}
}
