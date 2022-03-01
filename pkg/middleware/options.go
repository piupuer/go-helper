package middleware

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

type CorsOptions struct {
	origin     string
	header     string
	expose     string
	method     string
	credential string
}

func WithCorsOrigin(s string) func(*CorsOptions) {
	return func(options *CorsOptions) {
		getCorsOptionsOrSetDefault(options).origin = s
	}
}

func WithCorsHeader(s string) func(*CorsOptions) {
	return func(options *CorsOptions) {
		getCorsOptionsOrSetDefault(options).header = s
	}
}

func WithCorsExpose(s string) func(*CorsOptions) {
	return func(options *CorsOptions) {
		getCorsOptionsOrSetDefault(options).expose = s
	}
}

func WithCorsMethod(s string) func(*CorsOptions) {
	return func(options *CorsOptions) {
		getCorsOptionsOrSetDefault(options).method = s
	}
}

func WithCorsCredential(s string) func(*CorsOptions) {
	return func(options *CorsOptions) {
		getCorsOptionsOrSetDefault(options).credential = s
	}
}

func getCorsOptionsOrSetDefault(options *CorsOptions) *CorsOptions {
	if options == nil {
		return &CorsOptions{
			origin:     constant.MiddlewareCorsOrigin,
			header:     constant.MiddlewareCorsHeaders,
			expose:     constant.MiddlewareCorsExpose,
			method:     constant.MiddlewareCorsMethods,
			credential: constant.MiddlewareCorsCredentials,
		}
	}
	return options
}

type ParamsOptions struct {
	bodyCtxKey  string
	queryCtxKey string
}

func WithParamsBodyCtxKey(key string) func(*ParamsOptions) {
	return func(options *ParamsOptions) {
		getParamsOptionsOrSetDefault(options).bodyCtxKey = key
	}
}

func WithParamsQueryCtxKey(key string) func(*ParamsOptions) {
	return func(options *ParamsOptions) {
		getParamsOptionsOrSetDefault(options).queryCtxKey = key
	}
}

func getParamsOptionsOrSetDefault(options *ParamsOptions) *ParamsOptions {
	if options == nil {
		return &ParamsOptions{
			bodyCtxKey:  constant.MiddlewareParamsBodyCtxKey,
			queryCtxKey: constant.MiddlewareParamsQueryCtxKey,
		}
	}
	return options
}

type AccessLogOptions struct {
	urlPrefix string
	detail    bool
}

func WithAccessLogUrlPrefix(prefix string) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		getAccessLogOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func WithAccessLogDetail(flag bool) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		getAccessLogOptionsOrSetDefault(options).detail = flag
	}
}

func getAccessLogOptionsOrSetDefault(options *AccessLogOptions) *AccessLogOptions {
	if options == nil {
		return &AccessLogOptions{
			urlPrefix: constant.MiddlewareUrlPrefix,
			detail:    true,
		}
	}
	return options
}

type CasbinOptions struct {
	urlPrefix      string
	getCurrentUser func(c *gin.Context) ms.User
	Enforcer       *casbin.Enforcer
	failWithCode   func(code int)
}

func WithCasbinUrlPrefix(prefix string) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		getCasbinOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func WithCasbinGetCurrentUser(fun func(c *gin.Context) ms.User) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if fun != nil {
			getCasbinOptionsOrSetDefault(options).getCurrentUser = fun
		}
	}
}

func WithCasbinEnforcer(enforcer *casbin.Enforcer) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if enforcer != nil {
			getCasbinOptionsOrSetDefault(options).Enforcer = enforcer
		}
	}
}

func WithCasbinFailWithCode(fun func(code int)) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if fun != nil {
			getCasbinOptionsOrSetDefault(options).failWithCode = fun
		}
	}
}

func ParseCasbinOptions(options ...func(*CasbinOptions)) *CasbinOptions {
	ops := getCasbinOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return ops
}

func getCasbinOptionsOrSetDefault(options *CasbinOptions) *CasbinOptions {
	if options == nil {
		options = &CasbinOptions{}
		options.urlPrefix = constant.MiddlewareUrlPrefix
		options.failWithCode = resp.FailWithCode
	}
	return options
}

type SignOptions struct {
	expire       string
	findSkipPath func(c *gin.Context) []string
	getSignUser  func(c *gin.Context, appId string) ms.SignUser
	headerKey    []string
	checkScope   bool
}

func WithSignExpire(duration string) func(*SignOptions) {
	return func(options *SignOptions) {
		getSignOptionsOrSetDefault(options).expire = duration
	}
}

func WithSignFindSkipPath(fun func(c *gin.Context) []string) func(*SignOptions) {
	return func(options *SignOptions) {
		if fun != nil {
			getSignOptionsOrSetDefault(options).findSkipPath = fun
		}
	}
}

func WithSignGetSignUser(fun func(c *gin.Context, appId string) ms.SignUser) func(*SignOptions) {
	return func(options *SignOptions) {
		if fun != nil {
			getSignOptionsOrSetDefault(options).getSignUser = fun
		}
	}
}

func WithSignHeaderKey(arr ...string) func(*SignOptions) {
	return func(options *SignOptions) {
		switch len(arr) {
		case 1:
			getSignOptionsOrSetDefault(options).headerKey[0] = arr[0]
		case 2:
			getSignOptionsOrSetDefault(options).headerKey[0] = arr[0]
			getSignOptionsOrSetDefault(options).headerKey[1] = arr[1]
		case 3:
			getSignOptionsOrSetDefault(options).headerKey[0] = arr[0]
			getSignOptionsOrSetDefault(options).headerKey[1] = arr[1]
			getSignOptionsOrSetDefault(options).headerKey[2] = arr[2]
		case 4:
			getSignOptionsOrSetDefault(options).headerKey = arr
		}
	}
}

func WithSignCheckScope(flag bool) func(*SignOptions) {
	return func(options *SignOptions) {
		getSignOptionsOrSetDefault(options).checkScope = flag
	}
}

func getSignOptionsOrSetDefault(options *SignOptions) *SignOptions {
	if options == nil {
		return &SignOptions{
			expire: "60s",
			headerKey: []string{
				constant.MiddlewareSignTokenHeaderKey,
				constant.MiddlewareSignAppIdHeaderKey,
				constant.MiddlewareSignTimestampHeaderKey,
				constant.MiddlewareSignSignatureHeaderKey,
			},
			checkScope: true,
		}
	}
	return options
}

type IdempotenceOptions struct {
	redis           redis.UniversalClient
	cachePrefix     string
	expire          int
	tokenName       string
	successWithData func(...interface{})
	failWithMsg     func(format interface{}, a ...interface{})
}

func WithIdempotenceRedis(rd redis.UniversalClient) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if rd != nil {
			getIdempotenceOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithIdempotenceCachePrefix(prefix string) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		getIdempotenceOptionsOrSetDefault(options).cachePrefix = prefix
	}
}

func WithIdempotenceExpire(hours int) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if hours > 0 {
			getIdempotenceOptionsOrSetDefault(options).expire = hours
		}
	}
}

func WithIdempotenceTokenName(name string) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		getIdempotenceOptionsOrSetDefault(options).tokenName = name
	}
}

func WithIdempotenceSuccessWithData(fun func(...interface{})) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if fun != nil {
			getIdempotenceOptionsOrSetDefault(options).successWithData = fun
		}
	}
}

func WithIdempotenceFailWithMsg(fun func(format interface{}, a ...interface{})) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if fun != nil {
			getIdempotenceOptionsOrSetDefault(options).failWithMsg = fun
		}
	}
}

func ParseIdempotenceOptions(options ...func(*IdempotenceOptions)) *IdempotenceOptions {
	ops := getIdempotenceOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return ops
}

func getIdempotenceOptionsOrSetDefault(options *IdempotenceOptions) *IdempotenceOptions {
	if options == nil {
		return &IdempotenceOptions{
			cachePrefix:     constant.MiddlewareIdempotencePrefix,
			expire:          constant.MiddlewareIdempotenceExpire,
			tokenName:       constant.MiddlewareIdempotenceTokenName,
			successWithData: resp.SuccessWithData,
			failWithMsg:     resp.FailWithMsg,
		}
	}
	return options
}

type JwtOptions struct {
	realm              string
	key                string
	timeout            int
	maxRefresh         int
	tokenLookup        string
	tokenHeaderName    string
	sendCookie         bool
	cookieName         string
	privateBytes       []byte
	success            func()
	successWithData    func(...interface{})
	failWithMsg        func(format interface{}, a ...interface{})
	failWithCodeAndMsg func(code int, format interface{}, a ...interface{})
	loginPwdCheck      func(c *gin.Context, r req.LoginCheck) (userId int64, err error)
}

func WithJwtRealm(realm string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).realm = realm
	}
}

func WithJwtKey(key string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).key = key
	}
}

func WithJwtTimeout(timeout int) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).timeout = timeout
	}
}

func WithJwtMaxRefresh(maxRefresh int) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).maxRefresh = maxRefresh
	}
}

func WithJwtTokenLookup(tokenLookup string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).tokenLookup = tokenLookup
	}
}

func WithJwtTokenHeaderName(tokenHeaderName string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).tokenHeaderName = tokenHeaderName
	}
}

func WithJwtSendCookie(flag bool) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).sendCookie = flag
	}
}

func WithJwtCookieName(cookieName string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).cookieName = cookieName
	}
}

func WithJwtPrivateBytes(bs []byte) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOrSetDefault(options).privateBytes = bs
	}
}

func WithJwtSuccess(fun func()) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).success = fun
		}
	}
}

func WithJwtSuccessWithData(fun func(...interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).successWithData = fun
		}
	}
}

func WithJwtFailWithMsg(fun func(format interface{}, a ...interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).failWithMsg = fun
		}
	}
}

func WithJwtFailWithCodeAndMsg(fun func(code int, format interface{}, a ...interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).failWithCodeAndMsg = fun
		}
	}
}

func WithJwtLoginPwdCheck(fun func(c *gin.Context, r req.LoginCheck) (userId int64, err error)) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).loginPwdCheck = fun
		}
	}
}

func getJwtOptionsOrSetDefault(options *JwtOptions) *JwtOptions {
	if options == nil {
		return &JwtOptions{
			realm:              "my jwt",
			key:                "my secret",
			timeout:            24,
			maxRefresh:         168,
			tokenLookup:        "header: Authorization, query: token, cookie: jwt",
			tokenHeaderName:    "Bearer",
			success:            resp.Success,
			successWithData:    resp.SuccessWithData,
			failWithMsg:        resp.FailWithMsg,
			failWithCodeAndMsg: resp.FailWithCodeAndMsg,
			loginPwdCheck: func(c *gin.Context, r req.LoginCheck) (userId int64, err error) {
				return 0, errors.Errorf(resp.LoginCheckErrorMsg)
			},
		}
	}
	return options
}

type OperationLogOptions struct {
	redis                  redis.UniversalClient
	cachePrefix            string
	urlPrefix              string
	realIpKey              string
	skipGetOrOptionsMethod bool
	findSkipPath           func(c *gin.Context) []string
	singleFileMaxSize      int64
	getCurrentUser         func(c *gin.Context) ms.User
	save                   func(c *gin.Context, list []OperationRecord)
	maxCountBeforeSave     int
	findApi                func(c *gin.Context) []OperationApi
}

func WithOperationLogRedis(rd redis.UniversalClient) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if rd != nil {
			getOperationLogOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithOperationLogCachePrefix(prefix string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).cachePrefix = prefix
	}
}

func WithOperationLogUrlPrefix(prefix string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func WithOperationLogRealIpKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).realIpKey = key
	}
}

func WithOperationLogSkipGetOrOptionsMethod(flag bool) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).skipGetOrOptionsMethod = flag
	}
}

func WithOperationLogFindSkipPath(fun func(c *gin.Context) []string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOrSetDefault(options).findSkipPath = fun
		}
	}
}

func WithOperationLogSingleFileMaxSize(size int64) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if size >= 0 {
			getOperationLogOptionsOrSetDefault(options).singleFileMaxSize = size
		}
	}
}

func WithOperationLogGetCurrentUser(fun func(c *gin.Context) ms.User) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOrSetDefault(options).getCurrentUser = fun
		}
	}
}

func WithOperationLogSave(fun func(c *gin.Context, list []OperationRecord)) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOrSetDefault(options).save = fun
		}
	}
}

func WithOperationLogSaveMaxCount(count int) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if count > 0 {
			getOperationLogOptionsOrSetDefault(options).maxCountBeforeSave = count
		}
	}
}

func WithOperationLogFindApi(fun func(c *gin.Context) []OperationApi) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOrSetDefault(options).findApi = fun
		}
	}
}

func getOperationLogOptionsOrSetDefault(options *OperationLogOptions) *OperationLogOptions {
	if options == nil {
		options = &OperationLogOptions{}
		options.cachePrefix = constant.MiddlewareOperationLogApiCacheKey
		options.urlPrefix = constant.MiddlewareUrlPrefix
		options.maxCountBeforeSave = constant.MiddlewareOperationLogMaxCountBeforeSave
		options.singleFileMaxSize = 100
		options.getCurrentUser = func(c *gin.Context) ms.User {
			return ms.User{}
		}
		options.save = func(c *gin.Context, list []OperationRecord) {
			log.WithRequestId(c).Warn("operation log save is empty")
		}
		options.findApi = func(c *gin.Context) []OperationApi {
			log.WithRequestId(c).Warn("operation log findApi is empty")
			return []OperationApi{}
		}
	}
	return options
}

type RateOptions struct {
	redis    redis.UniversalClient
	maxLimit int64
}

func WithRateRedis(rd redis.UniversalClient) func(*RateOptions) {
	return func(options *RateOptions) {
		if rd != nil {
			getRateOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithRateMaxLimit(limit int64) func(*RateOptions) {
	return func(options *RateOptions) {
		getRateOptionsOrSetDefault(options).maxLimit = limit
	}
}

func getRateOptionsOrSetDefault(options *RateOptions) *RateOptions {
	if options == nil {
		return &RateOptions{
			maxLimit: 200,
		}
	}
	return options
}

type TransactionOptions struct {
	dbNoTx *gorm.DB
}

func WithTransactionDbNoTx(db *gorm.DB) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		if db != nil {
			getTransactionOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func getTransactionOptionsOrSetDefault(options *TransactionOptions) *TransactionOptions {
	if options == nil {
		return &TransactionOptions{}
	}
	return options
}

type PrintRouterOptions struct {
	ctx context.Context
}

func WithPrintRouterCtx(ctx context.Context) func(*PrintRouterOptions) {
	return func(options *PrintRouterOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getPrintRouterOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func getPrintRouterOptionsOrSetDefault(options *PrintRouterOptions) *PrintRouterOptions {
	if options == nil {
		return &PrintRouterOptions{
			ctx: context.Background(),
		}
	}
	return options
}
