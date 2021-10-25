package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/resp"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

// redis lua script(read => delete => get delete flag)
const (
	lua string = `
local current = redis.call('GET', KEYS[1])
if current == false then
    return '-1';
end
local del = redis.call('DEL', KEYS[1])
if del == 1 then
     return '1';
else
     return '0';
end
`
)

func Idempotence(options ...func(*IdempotenceOptions)) gin.HandlerFunc {
	ops := ParseIdempotenceOptions(options...)
	if ops.redis == nil {
		panic("idempotence redis is empty")
	}
	return func(c *gin.Context) {
		// read token from header at first
		token := c.Request.Header.Get(ops.tokenName)
		if token == "" {
			token, _ = c.Cookie(ops.tokenName)
		}
		token = strings.TrimSpace(token)
		if token == "" {
			ops.failWithMsg(resp.IdempotenceTokenEmptyMsg)
		}
		// check token
		if !CheckIdempotenceToken(c, token, *ops) {
			ops.failWithMsg(resp.IdempotenceTokenInvalidMsg)
		}
		c.Next()
	}
}

func GetIdempotenceToken(options ...func(*IdempotenceOptions)) gin.HandlerFunc {
	ops := getIdempotenceOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.redis == nil {
		panic("idempotence redis is empty")
	}
	return func(c *gin.Context) {
		ops.successWithData(GenIdempotenceToken(c, *ops))
	}
}

// generate token by redis
func GenIdempotenceToken(c context.Context, ops IdempotenceOptions) string {
	token := uuid.NewV4().String()
	ops.redis.Set(c, ops.prefix+token, true, time.Duration(ops.expire)*time.Hour)
	return token
}

// check token by exec redis lua script
func CheckIdempotenceToken(c context.Context, token string, ops IdempotenceOptions) bool {
	res, err := ops.redis.Eval(c, lua, []string{ops.prefix + token}).Result()
	if err != nil || res != "1" {
		return false
	}
	return true
}
