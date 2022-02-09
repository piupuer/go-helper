package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"io/ioutil"
	"net/http"
	"strings"
)

func Params(c *gin.Context) {
	getBody(c)
	getQuery(c)
	c.Next()
}

func getBody(c *gin.Context) (rp string) {
	if v := c.GetString(constant.MiddlewareParamsBodyCtxKey); v != "" {
		rp = v
		return
	}
	reqMethod := c.Request.Method
	// read body
	var body []byte
	if reqMethod == http.MethodPost || reqMethod == http.MethodPut || reqMethod == http.MethodPatch {
		var err error
		body, err = ioutil.ReadAll(c.Request.Body)
		if err == nil {
			// write back to gin request body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
	}
	if len(body) == 0 {
		rp = constant.MiddlewareParamsNullBody
	} else {
		rp = string(body)
	}
	c.Set(constant.MiddlewareParamsBodyCtxKey, rp)
	return
}

func getQuery(c *gin.Context) (rp string) {
	if v := c.GetString(constant.MiddlewareParamsQueryCtxKey); v != "" {
		rp = v
		return
	}
	rp = c.Request.URL.RawQuery
	c.Set(constant.MiddlewareParamsQueryCtxKey, rp)
	return
}

func getResp(c *gin.Context) (rp string) {
	if v := c.GetString(constant.MiddlewareParamsRespCtxKey); v != "" {
		rp = v
		return
	}
	if w, ok := c.Writer.(*accessWriter); ok {
		rp = w.body.String()
		c.Set(constant.MiddlewareParamsRespCtxKey, rp)
	}
	return
}

func getRequestDetail(c *gin.Context) (rp map[string]interface{}) {
	rp = make(map[string]interface{})
	rp[constant.MiddlewareParamsRespLogKey] = strings.ReplaceAll(getResp(c),"\"", "'")
	rp[constant.MiddlewareParamsQueryLogKey] = strings.ReplaceAll(getQuery(c),"\"", "'")
	rp[constant.MiddlewareParamsBodyLogKey] = strings.ReplaceAll(getBody(c),"\"", "'")
	return
}
