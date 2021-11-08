package middleware

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
	"strings"
)

type AccessLogOptions struct {
	logger    logger.Interface
	urlPrefix string
}

func WithAccessLogLogger(l logger.Interface) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		if l != nil {
			getAccessLogOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithAccessLogUrlPrefix(prefix string) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		getAccessLogOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func getAccessLogOptionsOrSetDefault(options *AccessLogOptions) *AccessLogOptions {
	if options == nil {
		return &AccessLogOptions{
			logger:    logger.DefaultLogger(),
			urlPrefix: constant.MiddlewareUrlPrefix,
		}
	}
	return options
}

type CasbinOptions struct {
	logger         logger.Interface
	urlPrefix      string
	getCurrentUser func(c *gin.Context) ms.User
	Enforcer       *casbin.Enforcer
	failWithCode   func(code int)
}

func WithCasbinLogger(l logger.Interface) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if l != nil {
			getCasbinOptionsOrSetDefault(options).logger = l
		}
	}
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
		options.logger = logger.DefaultLogger()
		options.urlPrefix = constant.MiddlewareUrlPrefix
		options.failWithCode = resp.FailWithCode
	}
	return options
}

type ExceptionOptions struct {
	logger             logger.Interface
	operationLogCtxKey string
	requestIdCtxKey    string
}

func WithExceptionLogger(l logger.Interface) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		if l != nil {
			getExceptionOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithExceptionOperationLogCtxKey(key string) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		getExceptionOptionsOrSetDefault(options).operationLogCtxKey = key
	}
}

func getExceptionOptionsOrSetDefault(options *ExceptionOptions) *ExceptionOptions {
	if options == nil {
		return &ExceptionOptions{
			logger:             logger.DefaultLogger(),
			operationLogCtxKey: constant.MiddlewareOperationLogCtxKey,
			requestIdCtxKey:    constant.MiddlewareRequestIdCtxKey,
		}
	}
	return options
}

type IdempotenceOptions struct {
	logger          logger.Interface
	redis           redis.UniversalClient
	prefix          string
	expire          int
	tokenName       string
	successWithData func(...interface{})
	failWithMsg     func(format interface{}, a ...interface{})
}

func WithIdempotenceLogger(l logger.Interface) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if l != nil {
			getIdempotenceOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithIdempotenceRedis(rd redis.UniversalClient) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if rd != nil {
			getIdempotenceOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithIdempotencePrefix(prefix string) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		getIdempotenceOptionsOrSetDefault(options).prefix = prefix
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
			logger:          logger.DefaultLogger(),
			prefix:          constant.MiddlewareIdempotencePrefix,
			expire:          constant.MiddlewareIdempotenceExpire,
			tokenName:       constant.MiddlewareIdempotenceTokenName,
			successWithData: resp.SuccessWithData,
			failWithMsg:     resp.FailWithMsg,
		}
	}
	return options
}

type JwtOptions struct {
	logger             logger.Interface
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
	loginPwdCheck      func(c *gin.Context, username, password string) (userId int64, pass bool)
}

func WithJwtLogger(l logger.Interface) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if l != nil {
			getJwtOptionsOrSetDefault(options).logger = l
		}
	}
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

func WithJwtLoginPwdCheck(fun func(c *gin.Context, username, password string) (userId int64, pass bool)) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOrSetDefault(options).loginPwdCheck = fun
		}
	}
}

func getJwtOptionsOrSetDefault(options *JwtOptions) *JwtOptions {
	if options == nil {
		return &JwtOptions{
			logger:             logger.DefaultLogger(),
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
			loginPwdCheck: func(c *gin.Context, username, password string) (userId int64, pass bool) {
				return 0, true
			},
		}
	}
	return options
}

type OperationLogOptions struct {
	logger                 logger.Interface
	redis                  redis.UniversalClient
	ctxKey                 string
	apiCacheKey            string
	urlPrefix              string
	realIpKey              string
	skipGetOrOptionsMethod bool
	skipPaths              []string
	singleFileMaxSize      int64
	getUserInfo            func(c *gin.Context) (username, roleName string)
	save                   func(c *gin.Context, list []OperationRecord)
	maxCountBeforeSave     int
	findApi                func(c *gin.Context) []OperationApi
}

func WithOperationLogLogger(l logger.Interface) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if l != nil {
			getOperationLogOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithOperationLogRedis(rd redis.UniversalClient) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if rd != nil {
			getOperationLogOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithOperationLogCtxKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).ctxKey = key
	}
}

func WithOperationLogApiCacheKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).apiCacheKey = key
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

func WithOperationLogSkipPaths(paths ...string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOrSetDefault(options).skipPaths = append(getOperationLogOptionsOrSetDefault(options).skipPaths, paths...)
	}
}

func WithOperationLogSingleFileMaxSize(size int64) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if size >= 0 {
			getOperationLogOptionsOrSetDefault(options).singleFileMaxSize = size
		}
	}
}

func WithOperationLogGetUserInfo(fun func(c *gin.Context) (username, roleName string)) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOrSetDefault(options).getUserInfo = fun
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
		options.logger = logger.DefaultLogger()
		options.ctxKey = constant.MiddlewareOperationLogCtxKey
		options.apiCacheKey = constant.MiddlewareOperationLogApiCacheKey
		options.urlPrefix = constant.MiddlewareUrlPrefix
		options.maxCountBeforeSave = constant.MiddlewareOperationLogMaxCountBeforeSave
		options.singleFileMaxSize = 100
		options.getUserInfo = func(c *gin.Context) (username, roleName string) {
			return constant.MiddlewareOperationLogNotLogin, constant.MiddlewareOperationLogNotLogin
		}
		options.save = func(c *gin.Context, list []OperationRecord) {
			options.logger.Warn(c, "operation log save is empty")
		}
		options.findApi = func(c *gin.Context) []OperationApi {
			options.logger.Warn(c, "operation log findApi is empty")
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

type RequestIdOptions struct {
	headerName string
	ctxKey     string
}

func WithRequestIdHeaderName(name string) func(*RequestIdOptions) {
	return func(options *RequestIdOptions) {
		getRequestIdOptionsOrSetDefault(options).headerName = name
	}
}

func WithRequestIdCtxKey(key string) func(*RequestIdOptions) {
	return func(options *RequestIdOptions) {
		getRequestIdOptionsOrSetDefault(options).ctxKey = key
	}
}

func getRequestIdOptionsOrSetDefault(options *RequestIdOptions) *RequestIdOptions {
	if options == nil {
		return &RequestIdOptions{
			headerName: constant.MiddlewareRequestIdHeaderName,
			ctxKey:     constant.MiddlewareRequestIdCtxKey,
		}
	}
	return options
}

type TransactionOptions struct {
	dbNoTx             *gorm.DB
	requestIdCtxKey    string
	txCtxKey           string
	operationLogCtxKey string
}

func WithTransactionDbNoTx(db *gorm.DB) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		if db != nil {
			getTransactionOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func WithTransactionRequestIdCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOrSetDefault(options).requestIdCtxKey = key
	}
}

func WithTransactionTxCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOrSetDefault(options).txCtxKey = key
	}
}

func WithTransactionOperationLogCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOrSetDefault(options).operationLogCtxKey = key
	}
}

func getTransactionOptionsOrSetDefault(options *TransactionOptions) *TransactionOptions {
	if options == nil {
		return &TransactionOptions{
			requestIdCtxKey:    constant.MiddlewareRequestIdCtxKey,
			txCtxKey:           constant.MiddlewareTransactionTxCtxKey,
			operationLogCtxKey: constant.MiddlewareOperationLogCtxKey,
		}
	}
	return options
}

type PrintRouterOptions struct {
	logger logger.Interface
	ctx    context.Context
}

func WithPrintRouterLogger(l logger.Interface) func(*PrintRouterOptions) {
	return func(options *PrintRouterOptions) {
		if l != nil {
			getPrintRouterOptionsOrSetDefault(options).logger = l
		}
	}
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
			logger: logger.DefaultLogger(),
			ctx:    context.Background(),
		}
	}
	return options
}
