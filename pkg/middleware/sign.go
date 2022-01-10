package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func Sign(options ...func(*SignOptions)) gin.HandlerFunc {
	ops := getSignOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.getSignUser == nil {
		panic("getSignUser is empty")
	}
	return func(c *gin.Context) {
		// token
		token := c.Request.Header.Get(ops.headerKey[0])
		if token == "" {
			ops.logger.Warn(c, resp.InvalidSignTokenMsg)
			abort(c, *ops, resp.InvalidSignTokenMsg)
			return
		}
		list := strings.Split(token, ",")
		re := regexp.MustCompile(`"[\D\d].*"`)
		var appId, timestamp, signature string
		for _, item := range list {
			ms := re.FindAllString(item, -1)
			if len(ms) == 1 {
				if strings.HasPrefix(item, ops.headerKey[1]) {
					appId = strings.Trim(ms[0], `"`)
				} else if strings.HasPrefix(item, ops.headerKey[2]) {
					timestamp = strings.Trim(ms[0], `"`)
				} else if strings.HasPrefix(item, ops.headerKey[3]) {
					signature = strings.Trim(ms[0], `"`)
				}
			}
		}
		if appId == "" {
			ops.logger.Warn(c, resp.InvalidSignIdMsg)
			abort(c, *ops, resp.InvalidSignIdMsg)
			return
		}
		if timestamp == "" {
			ops.logger.Warn(c, resp.InvalidSignTimestampMsg)
			abort(c, *ops, resp.InvalidSignTimestampMsg)
			return
		}
		// compare timestamp
		now := carbon.Now()
		t := carbon.CreateFromTimestamp(utils.Str2Int64(timestamp))
		if t.AddDuration(ops.expire).Lt(now) {
			ops.logger.Warn(c, "%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			abort(c, *ops, "%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			return
		}
		// query user by app id
		u := ops.getSignUser(c, appId)
		if u.AppSecret == "" {
			ops.logger.Warn(c, "%s: %s", resp.IllegalSignIdMsg, appId)
			abort(c, *ops, "%s: %s", resp.IllegalSignIdMsg, appId)
			return
		}
		if u.Status == constant.Zero {
			ops.logger.Warn(c, "%s: %s", resp.UserDisabledMsg, appId)
			abort(c, *ops, "%s: %s", resp.UserDisabledMsg, appId)
			return
		}
		// scope
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		if ops.checkScope {
			reqPath := c.Request.URL.Path
			exists := false
			for _, item := range u.Scopes {
				if item.Method == reqMethod && item.Path == reqPath {
					exists = true
					break
				}
			}
			if !exists {
				ops.logger.Warn(c, "%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				abort(c, *ops, "%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				return
			}
		}

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
			body = []byte("{}")
		}
		// verify signature
		if !verifySign(*ops, u.AppSecret, signature, reqMethod, reqUri, timestamp, string(body)) {
			ops.logger.Warn(c, "%s: %s", resp.IllegalSignTokenMsg, token)
			abort(c, *ops, "%s: %s", resp.IllegalSignTokenMsg, token)
			return
		}
		c.Next()
	}
}

func verifySign(ops SignOptions, secret, signature, method, uri, timestamp, body string) (flag bool) {
	b := bytes.NewBuffer(nil)
	b.WriteString(method)
	b.WriteString(ops.separator)
	b.WriteString(uri)
	b.WriteString(ops.separator)
	b.WriteString(timestamp)
	b.WriteString(ops.separator)
	b.WriteString(utils.JsonWithSort(body))
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(b.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	flag = digest == signature
	return
}

func abort(c *gin.Context, ops SignOptions, format interface{}, a ...interface{}) {
	rp := resp.GetFailWithMsg(format, a...)
	rp.RequestId = c.GetString(ops.requestIdCtxKey)
	c.JSON(http.StatusForbidden, rp)
	c.Abort()
}
