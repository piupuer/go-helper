package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
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
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "Idempotence"))
		defer span.End()
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

// GetIdempotenceToken
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *Base
// @Description IdempotenceToken
// @Router /base/idempotenceToken [GET]
func GetIdempotenceToken(options ...func(*IdempotenceOptions)) gin.HandlerFunc {
	ops := getIdempotenceOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "GetIdempotenceToken"))
		defer span.End()
		ops.successWithData(GenIdempotenceToken(c, *ops))
	}
}

// generate token by redis
func GenIdempotenceToken(ctx context.Context, ops IdempotenceOptions) string {
	token := uuid.NewString()
	if ops.redis != nil {
		ops.redis.Set(ctx, fmt.Sprintf("%s_%s", ops.cachePrefix, token), true, time.Duration(ops.expire)*time.Hour)
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotence is invalid")
	}
	return token
}

// check token by exec redis lua script
func CheckIdempotenceToken(ctx context.Context, token string, ops IdempotenceOptions) bool {
	if ops.redis != nil {
		res, err := ops.redis.Eval(ctx, lua, []string{fmt.Sprintf("%s_%s", ops.cachePrefix, token)}).Result()
		if err != nil || res != "1" {
			return false
		}
	} else {
		log.WithContext(ctx).Warn("please enable redis, otherwise the idempotence is invalid")
	}
	return true
}
