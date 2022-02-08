package constant

const (
	MiddlewareUrlPrefix                      = "api"
	MiddlewareIdempotencePrefix              = "idempotence"
	MiddlewareIdempotenceExpire              = 24
	MiddlewareIdempotenceTokenName           = "api-idempotence-token"
	MiddlewareOperationLogCtxKey             = "ResponseBody"
	MiddlewareOperationLogNotLogin           = "not login"
	MiddlewareOperationLogApiCacheKey        = "operation_log_api"
	MiddlewareOperationLogSkipPathDict       = "OperationLogSkipPath"
	MiddlewareOperationLogMaxCountBeforeSave = 100
	MiddlewareRequestIdHeaderName            = "X-Request-Id"
	MiddlewareRequestIdCtxKey                = "RequestId"
	MiddlewareTransactionTxCtxKey            = "tx"
	MiddlewareTransactionForceCommitCtxKey   = "ForceCommitTx"
	MiddlewareJwtUserCtxKey                  = "user"
	MiddlewareSignSeparator                  = "|"
	MiddlewareSignTokenHeaderKey             = "X-Sign-Token"
	MiddlewareSignAppIdHeaderKey             = "appid"
	MiddlewareSignTimestampHeaderKey         = "timestamp"
	MiddlewareSignSignatureHeaderKey         = "signature"
	MiddlewareAccessLogIpLogKey              = "Ip"
	MiddlewareParamsQueryCtxKey              = "ParamsQuery"
	MiddlewareParamsBodyCtxKey               = "ParamsBody"
	MiddlewareParamsNullBody                 = "{}"
	MiddlewareParamsQueryLogKey              = "Query"
	MiddlewareParamsBodyLogKey               = "Body"
	MiddlewareParamsRespLogKey               = "Resp"
)
