package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"net/http"
	"net/url"
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
		// id
		appId := c.Request.Header.Get(ops.headerKey[0])
		if appId == "" {
			ops.logger.Warn(c, resp.InvalidSignIdMsg)
			abort(c, *ops, resp.InvalidSignIdMsg)
			return
		}
		// timestamp
		timestamp := c.Request.Header.Get(ops.headerKey[1])
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
		// token
		token := c.Request.Header.Get(ops.headerKey[2])
		if token == "" {
			ops.logger.Warn(c, resp.InvalidSignTokenMsg)
			abort(c, *ops, resp.InvalidSignTokenMsg)
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
		if ops.checkScope {
			reqMethod := c.Request.Method
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

		// set params
		contentType := c.ContentType()
		params := make(url.Values)
		switch contentType {
		case binding.MIMEJSON:
			params.Set(ops.valKey[0], c.GetString(ops.ctxKey[0]))
		case binding.MIMEPOSTForm:
			params.Set(ops.valKey[1], c.GetString(ops.ctxKey[1]))
		}
		params.Set(ops.valKey[2], c.GetString(ops.ctxKey[2]))
		// verify signature
		if !verifySign(*ops, u.AppSecret, token, timestamp, params) {
			ops.logger.Warn(c, "%s: %s", resp.IllegalSignTokenMsg, token)
			abort(c, *ops, "%s: %s", resp.IllegalSignTokenMsg, token)
			return
		}
		c.Next()
	}
}

func verifySign(ops SignOptions, secret, sign, timestamp string, params url.Values) (flag bool) {
	p, err := url.QueryUnescape(params.Encode())
	if err != nil {
		return
	}
	b := bytes.NewBuffer(nil)
	b.WriteString(ops.separator)
	b.WriteString(timestamp)
	b.WriteString(ops.separator)
	b.WriteString(p)
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(b.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	fmt.Println(digest)
	flag = digest == sign
	return
}

func abort(c *gin.Context, ops SignOptions, format interface{}, a ...interface{}) {
	rp := resp.GetFailWithMsg(format, a...)
	rp.RequestId = c.GetString(ops.requestIdCtxKey)
	c.JSON(http.StatusForbidden, rp)
	c.Abort()
}
