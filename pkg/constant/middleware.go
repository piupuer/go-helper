package constant

const (
	MiddlewareUrlPrefix                      = "api"
	MiddlewareIdempotencePrefix              = "idempotence"
	MiddlewareIdempotenceExpire              = 24
	MiddlewareIdempotenceTokenName           = "api-idempotence-token"
	MiddlewareOperationLogCtxKey             = "operation_log_response"
	MiddlewareOperationLogNotLogin           = "not login"
	MiddlewareOperationLogApiCacheKey        = "operation_log_api"
	MiddlewareOperationLogSkipPathDict       = "OperationLogSkipPath"
	MiddlewareOperationLogMaxCountBeforeSave = 100
	MiddlewareRequestIdHeaderName            = "X-Request-Id"
	MiddlewareRequestIdCtxKey                = "RequestId"
	MiddlewareTransactionTxCtxKey            = "tx"
	MiddlewareTransactionForceCommitCtxKey   = "ForceCommitTx"
)
