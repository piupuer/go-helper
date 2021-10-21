package constant

const (
	MiddlewareUrlPrefix                      = "api"
	MiddlewareRoleKey                        = "admin"
	MiddlewareIdempotencePrefix              = "idempotence_"
	MiddlewareIdempotenceExpire              = 24
	MiddlewareIdempotenceTokenName           = "api-idempotence-token"
	MiddlewareOperationLogCtxKey             = "operation_log_response"
	MiddlewareOperationLogNotLogin           = "not login"
	MiddlewareOperationLogApiCacheKey        = "OPERATION_LOG_API"
	MiddlewareOperationLogMaxCountBeforeSave = 100
	MiddlewareRequestIdHeaderName            = "X-Request-Id"
	MiddlewareRequestIdCtxKey                = "RequestId"
	MiddlewareTransactionTxCtxKey            = "tx"
)
