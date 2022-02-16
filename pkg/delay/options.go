package delay

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
	"strings"
)

type ExportOptions struct {
	ctx       context.Context
	dbNoTx    *gorm.DB
	machineId string
	tbPrefix  string
	objPrefix string
	key       string
	secret    string
	endpoint  string
	bucket    string
	expire    int64
}

func WithExportCtx(ctx context.Context) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).ctx = getExportCtx(ctx)
	}
}

func WithExportDbNoTx(db *gorm.DB) func(*ExportOptions) {
	return func(options *ExportOptions) {
		if db != nil {
			getExportOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func WithExportMachineId(id string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).machineId = id
	}
}

func WithExportTbPrefix(prefix string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).tbPrefix = prefix
	}
}

func WithExportObjPrefix(prefix string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).objPrefix = prefix
	}
}

func WithExportKey(key string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).key = key
	}
}

func WithExportSecret(secret string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).secret = secret
	}
}

func WithExportEndpoint(endpoint string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		if !strings.HasSuffix(endpoint, constant.DelayExportEndPointSuffix) {
			endpoint = endpoint + constant.DelayExportEndPointSuffix
		}
		getExportOptionsOrSetDefault(options).endpoint = endpoint
	}
}

func WithExportBucket(bucket string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).bucket = bucket
	}
}

func WithExportExpire(min int64) func(*ExportOptions) {
	return func(options *ExportOptions) {
		if min > 0 {
			getExportOptionsOrSetDefault(options).expire = min
		}
	}
}

func getExportOptionsOrSetDefault(options *ExportOptions) *ExportOptions {
	if options == nil {
		return &ExportOptions{
			ctx:       getExportCtx(nil),
			tbPrefix:  constant.DelayExportTbPrefix,
			objPrefix: constant.DelayExportObjPrefix,
			machineId: fmt.Sprintf("%d", constant.One),
			endpoint:  "oss-cn-shenzhen" + constant.DelayExportEndPointSuffix,
			bucket:    "piupuer",
			expire:    constant.DelayExportObjExpire,
		}
	}
	return options
}

func getExportCtx(ctx context.Context) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	return context.WithValue(ctx, constant.LogSkipHelperCtxKey, false)
}
