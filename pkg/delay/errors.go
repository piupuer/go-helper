package delay

import "fmt"

var (
	ErrDbNil              = fmt.Errorf("db instance is empty")
	ErrUuidNil            = fmt.Errorf("uuid is empty")
	ErrUuidInvalid        = fmt.Errorf("uuid is invalid")
	ErrOssSecretInvalid   = fmt.Errorf("oss id or secret is invalid")
	ErrOssBucketInvalid   = fmt.Errorf("oss bucket is invalid")
	ErrOssPutObjectFailed = fmt.Errorf("oss put object failed")
)
