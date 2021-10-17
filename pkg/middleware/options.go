package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/resp"
	"gorm.io/gorm"
	"strings"
)

const (
	urlPrefix               = "api"
	roleKey                 = "admin"
	idempotencePrefix       = "idempotence_"
	idempotenceTokenName    = "api-idempotence-token"
	operationLogCtxKey      = "operation_log_response"
	operationLogNotLogin    = "not login"
	operationLogApiCacheKey = "OPERATION_LOG_API"
	requestIdHeaderName     = "X-Request-Id"
	requestIdCtxKey         = "RequestId"
	transactionTxCtxKey     = "tx"
)

type AccessLogOptions struct {
	logger    logger.Interface
	urlPrefix string
}

func WithAccessLogLogger(l logger.Interface) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		if l != nil {
			getAccessLogOptionsOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithAccessLogLoggerLevel(level logger.Level) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		l := options.logger
		if options.logger == nil {
			l = getAccessLogOptionsOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithAccessLogUrlPrefix(prefix string) func(*AccessLogOptions) {
	return func(options *AccessLogOptions) {
		getAccessLogOptionsOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func getAccessLogOptionsOptionsOrSetDefault(options *AccessLogOptions) *AccessLogOptions {
	if options == nil {
		return &AccessLogOptions{
			logger:    logger.DefaultLogger(),
			urlPrefix: urlPrefix,
		}
	}
	return options
}

type CasbinOptions struct {
	urlPrefix    string
	roleKey      func(c *gin.Context) string
	enforcer     *casbin.Enforcer
	failWithCode func(code int)
}

func WithCasbinUrlPrefix(prefix string) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		getCasbinOptionsOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func WithCasbinRoleKey(fun func(c *gin.Context) string) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if fun != nil {
			getCasbinOptionsOptionsOrSetDefault(options).roleKey = fun
		}
	}
}

func WithCasbinEnforcer(enforcer *casbin.Enforcer) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if enforcer != nil {
			getCasbinOptionsOptionsOrSetDefault(options).enforcer = enforcer
		}
	}
}

func WithCasbinFailWithCode(fun func(code int)) func(*CasbinOptions) {
	return func(options *CasbinOptions) {
		if fun != nil {
			getCasbinOptionsOptionsOrSetDefault(options).failWithCode = fun
		}
	}
}

func getCasbinOptionsOptionsOrSetDefault(options *CasbinOptions) *CasbinOptions {
	if options == nil {
		return &CasbinOptions{
			urlPrefix:    urlPrefix,
			failWithCode: resp.FailWithCode,
			roleKey: func(c *gin.Context) string {
				return roleKey
			},
		}
	}
	return options
}

type ExceptionOptions struct {
	logger             logger.Interface
	operationLogCtxKey string
}

func WithExceptionLogger(l logger.Interface) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		if l != nil {
			getExceptionOptionsOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithExceptionLoggerLevel(level logger.Level) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		l := options.logger
		if options.logger == nil {
			l = getExceptionOptionsOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithExceptionOperationLogCtxKey(key string) func(*ExceptionOptions) {
	return func(options *ExceptionOptions) {
		getExceptionOptionsOptionsOrSetDefault(options).operationLogCtxKey = key
	}
}

func getExceptionOptionsOptionsOrSetDefault(options *ExceptionOptions) *ExceptionOptions {
	if options == nil {
		return &ExceptionOptions{
			logger:             logger.DefaultLogger(),
			operationLogCtxKey: operationLogCtxKey,
		}
	}
	return options
}

type IdempotenceOptions struct {
	redis           redis.UniversalClient
	prefix          string
	tokenName       string
	successWithData func(interface{})
	failWithMsg     func(format interface{}, a ...interface{})
}

func WithIdempotenceRedis(rd redis.UniversalClient) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if rd != nil {
			getIdempotenceOptionsOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithIdempotencePrefix(prefix string) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		getIdempotenceOptionsOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithIdempotenceTokenName(name string) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		getIdempotenceOptionsOptionsOrSetDefault(options).tokenName = name
	}
}

func WithIdempotenceSuccessWithData(fun func(interface{})) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if fun != nil {
			getIdempotenceOptionsOptionsOrSetDefault(options).successWithData = fun
		}
	}
}

func WithIdempotenceFailWithMsg(fun func(format interface{}, a ...interface{})) func(*IdempotenceOptions) {
	return func(options *IdempotenceOptions) {
		if fun != nil {
			getIdempotenceOptionsOptionsOrSetDefault(options).failWithMsg = fun
		}
	}
}

func getIdempotenceOptionsOptionsOrSetDefault(options *IdempotenceOptions) *IdempotenceOptions {
	if options == nil {
		return &IdempotenceOptions{
			prefix:          idempotencePrefix,
			tokenName:       idempotenceTokenName,
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
	successWithData    func(interface{})
	failWithMsg        func(format interface{}, a ...interface{})
	failWithCodeAndMsg func(code int, format interface{}, a ...interface{})
	loginPwdCheck      func(c *gin.Context, username, password string) (userId int64, pass bool)
}

func WithJwtLogger(l logger.Interface) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if l != nil {
			getJwtOptionsOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithJwtLoggerLevel(level logger.Level) func(*JwtOptions) {
	return func(options *JwtOptions) {
		l := options.logger
		if options.logger == nil {
			l = getJwtOptionsOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithJwtRealm(realm string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).realm = realm
	}
}

func WithJwtKey(key string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).key = key
	}
}

func WithJwtTimeout(timeout int) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).timeout = timeout
	}
}

func WithJwtMaxRefresh(maxRefresh int) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).maxRefresh = maxRefresh
	}
}

func WithJwtTokenLookup(tokenLookup string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).tokenLookup = tokenLookup
	}
}

func WithJwtTokenHeaderName(tokenHeaderName string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).tokenHeaderName = tokenHeaderName
	}
}

func WithJwtSendCookie(options *JwtOptions) {
	getJwtOptionsOptionsOrSetDefault(options).sendCookie = true
}

func WithJwtCookieName(cookieName string) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).cookieName = cookieName
	}
}

func WithJwtPrivateBytes(bs []byte) func(*JwtOptions) {
	return func(options *JwtOptions) {
		getJwtOptionsOptionsOrSetDefault(options).privateBytes = bs
	}
}

func WithJwtSuccess(fun func()) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOptionsOrSetDefault(options).success = fun
		}
	}
}

func WithJwtSuccessWithData(fun func(interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOptionsOrSetDefault(options).successWithData = fun
		}
	}
}

func WithJwtFailWithMsg(fun func(format interface{}, a ...interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOptionsOrSetDefault(options).failWithMsg = fun
		}
	}
}

func WithJwtFailWithCodeAndMsg(fun func(code int, format interface{}, a ...interface{})) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOptionsOrSetDefault(options).failWithCodeAndMsg = fun
		}
	}
}

func WithJwtLoginPwdCheck(fun func(c *gin.Context, username, password string) (userId int64, pass bool)) func(*JwtOptions) {
	return func(options *JwtOptions) {
		if fun != nil {
			getJwtOptionsOptionsOrSetDefault(options).loginPwdCheck = fun
		}
	}
}

func getJwtOptionsOptionsOrSetDefault(options *JwtOptions) *JwtOptions {
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
	findApi                func(c *gin.Context) []OperationApi
}

func WithOperationLogLogger(l logger.Interface) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if l != nil {
			getOperationLogOptionsOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithOperationLogLoggerLevel(level logger.Level) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		l := options.logger
		if options.logger == nil {
			l = getOperationLogOptionsOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithOperationLogRedis(rd redis.UniversalClient) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if rd != nil {
			getOperationLogOptionsOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithOperationLogCtxKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOptionsOrSetDefault(options).ctxKey = key
	}
}

func WithOperationLogApiCacheKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOptionsOrSetDefault(options).apiCacheKey = key
	}
}

func WithOperationLogUrlPrefix(prefix string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOptionsOrSetDefault(options).urlPrefix = strings.Trim(prefix, "/")
	}
}

func WithOperationLogRealIpKey(key string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOptionsOrSetDefault(options).realIpKey = key
	}
}

func WithOperationLogSkipGetOrOptionsMethod(options *OperationLogOptions) {
	getOperationLogOptionsOptionsOrSetDefault(options).skipGetOrOptionsMethod = true
}

func WithOperationLogSkipPaths(paths ...string) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		getOperationLogOptionsOptionsOrSetDefault(options).skipPaths = append(getOperationLogOptionsOptionsOrSetDefault(options).skipPaths, paths...)
	}
}

func WithOperationLogSingleFileMaxSize(size int64) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if size >= 0 {
			getOperationLogOptionsOptionsOrSetDefault(options).singleFileMaxSize = size
		}
	}
}

func WithOperationLogGetUserInfo(fun func(c *gin.Context) (username, roleName string)) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOptionsOrSetDefault(options).getUserInfo = fun
		}
	}
}

func WithOperationLogSave(fun func(c *gin.Context, list []OperationRecord)) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOptionsOrSetDefault(options).save = fun
		}
	}
}

func WithOperationLogFindApi(fun func(c *gin.Context) []OperationApi) func(*OperationLogOptions) {
	return func(options *OperationLogOptions) {
		if fun != nil {
			getOperationLogOptionsOptionsOrSetDefault(options).findApi = fun
		}
	}
}

func getOperationLogOptionsOptionsOrSetDefault(options *OperationLogOptions) *OperationLogOptions {
	if options == nil {
		l := logger.DefaultLogger()
		return &OperationLogOptions{
			logger:      l,
			ctxKey:      operationLogCtxKey,
			apiCacheKey: operationLogApiCacheKey,
			urlPrefix:   urlPrefix,
			getUserInfo: func(c *gin.Context) (username, roleName string) {
				return operationLogNotLogin, operationLogNotLogin
			},
			singleFileMaxSize: 100,
			save: func(c *gin.Context, list []OperationRecord) {
				l.Warn(c, "operation log save handler is not config")
			},
			findApi: func(c *gin.Context) []OperationApi {
				l.Warn(c, "operation log find api handler is not config")
				return []OperationApi{}
			},
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
			getRateOptionsOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithRateMaxLimit(limit int64) func(*RateOptions) {
	return func(options *RateOptions) {
		getRateOptionsOptionsOrSetDefault(options).maxLimit = limit
	}
}

func getRateOptionsOptionsOrSetDefault(options *RateOptions) *RateOptions {
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
		getRequestIdOptionsOptionsOrSetDefault(options).headerName = name
	}
}

func WithRequestIdCtxKey(key string) func(*RequestIdOptions) {
	return func(options *RequestIdOptions) {
		getRequestIdOptionsOptionsOrSetDefault(options).ctxKey = key
	}
}

func getRequestIdOptionsOptionsOrSetDefault(options *RequestIdOptions) *RequestIdOptions {
	if options == nil {
		return &RequestIdOptions{
			headerName: requestIdHeaderName,
			ctxKey:     requestIdCtxKey,
		}
	}
	return options
}

type TransactionOptions struct {
	dbNoTx          *gorm.DB
	requestIdCtxKey string
	txCtxKey        string
}

func WithTransactionDbNoTx(db *gorm.DB) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		if db != nil {
			getTransactionOptionsOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func WithTransactionRequestIdCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOptionsOrSetDefault(options).requestIdCtxKey = key
	}
}

func WithTransactionTxCtxKey(key string) func(*TransactionOptions) {
	return func(options *TransactionOptions) {
		getTransactionOptionsOptionsOrSetDefault(options).txCtxKey = key
	}
}

func getTransactionOptionsOptionsOrSetDefault(options *TransactionOptions) *TransactionOptions {
	if options == nil {
		return &TransactionOptions{
			requestIdCtxKey: requestIdCtxKey,
			txCtxKey:        transactionTxCtxKey,
		}
	}
	return options
}
