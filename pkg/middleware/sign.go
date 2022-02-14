package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
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
		if ops.findSkipPath != nil {
			list := ops.findSkipPath(c)
			for _, item := range list {
				if strings.Contains(c.Request.URL.Path, item) {
					c.Next()
					return
				}
			}
		}
		// token
		token := c.Request.Header.Get(ops.headerKey[0])
		if token == "" {
			log.WithRequestId(c).Warn(resp.InvalidSignTokenMsg)
			abort(c, resp.InvalidSignTokenMsg)
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
			log.WithRequestId(c).Warn(resp.InvalidSignIdMsg)
			abort(c, resp.InvalidSignIdMsg)
			return
		}
		if timestamp == "" {
			log.WithRequestId(c).Warn(resp.InvalidSignTimestampMsg)
			abort(c, resp.InvalidSignTimestampMsg)
			return
		}
		// compare timestamp
		now := carbon.Now()
		t := carbon.CreateFromTimestamp(utils.Str2Int64(timestamp))
		if t.AddDuration(ops.expire).Lt(now) {
			log.WithRequestId(c).Warn("%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			abort(c, "%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			return
		}
		// query user by app id
		u := ops.getSignUser(c, appId)
		if u.AppSecret == "" {
			log.WithRequestId(c).Warn("%s: %s", resp.IllegalSignIdMsg, appId)
			abort(c, "%s: %s", resp.IllegalSignIdMsg, appId)
			return
		}
		if u.Status == constant.Zero {
			log.WithRequestId(c).Warn("%s: %s", resp.UserDisabledMsg, appId)
			abort(c, "%s: %s", resp.UserDisabledMsg, appId)
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
				log.WithRequestId(c).Warn("%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				abort(c, "%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				return
			}
		}

		// verify signature
		if !verifySign(u.AppSecret, signature, reqMethod, reqUri, timestamp, getBody(c)) {
			log.WithRequestId(c).Warn("%s: %s", resp.IllegalSignTokenMsg, token)
			abort(c, "%s: %s", resp.IllegalSignTokenMsg, token)
			return
		}
		c.Next()
	}
}

func verifySign(secret, signature, method, uri, timestamp, body string) (flag bool) {
	b := bytes.NewBuffer(nil)
	b.WriteString(method)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(uri)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(timestamp)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(utils.JsonWithSort(body))
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(b.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	flag = digest == signature
	return
}

func abort(c *gin.Context, format interface{}, a ...interface{}) {
	rp := resp.GetFailWithMsg(format, a...)
	rp.RequestId = c.GetString(constant.MiddlewareRequestIdCtxKey)
	c.JSON(http.StatusForbidden, rp)
	c.Abort()
}
