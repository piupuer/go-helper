package delay

import "fmt"

var (
	ErrDbNil                         = fmt.Errorf("db instance is empty")
	ErrUuidNil                       = fmt.Errorf("uuid is empty")
	ErrUuidInvalid                   = fmt.Errorf("uuid is invalid")
	ErrOssSecretInvalid              = fmt.Errorf("oss id or secret is invalid")
	ErrOssBucketInvalid              = fmt.Errorf("oss bucket is invalid")
	ErrOssPutObjectFailed            = fmt.Errorf("oss put object failed")
	ErrRedisNil                      = fmt.Errorf("redis is empty")
	ErrRedisInvalid                  = fmt.Errorf("redis is invalid")
	ErrExprInvalid                   = fmt.Errorf("expr is invalid")
	ErrSaveCron                      = fmt.Errorf("save cron failed")
	ErrHttpCallbackTimeout           = fmt.Errorf("http callback timeout")
	ErrHttpCallback                  = fmt.Errorf("http callback err")
	ErrHttpCallbackInvalidStatusCode = fmt.Errorf("http callback invalid status code")
)
