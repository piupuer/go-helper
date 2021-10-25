package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/oss"
	"github.com/piupuer/go-helper/pkg/query"
)

type Options struct {
	cache                      bool
	cacheOps                   []func(options *query.RedisOptions)
	dbOps                      []func(options *query.MysqlOptions)
	redis                      redis.UniversalClient
	operationAllowedToDelete   bool
	getCurrentUser             func(c *gin.Context) ms.CurrentUser
	findRoleKeywordByRoleIds   func(c *gin.Context, roleIds []uint) []string
	uploadSaveDir              string
	uploadSingleMaxSize        int64
	uploadMergeConcurrentCount int
	uploadMinio                *oss.MinioOss
	uploadMinioBucket          string
}

func WithCache(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).cache = flag
	}
}

func WithCacheOps(ops ...func(options *query.RedisOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).cacheOps = append(getOptionsOrSetDefault(options).cacheOps, ops...)
	}
}

func WithDbOps(ops ...func(options *query.MysqlOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).dbOps = append(getOptionsOrSetDefault(options).dbOps, ops...)
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithOperationAllowedToDelete(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).operationAllowedToDelete = flag
	}
}

func WithGetCurrentUser(fun func(c *gin.Context) ms.CurrentUser) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getCurrentUser = fun
		}
	}
}

func WithFindRoleKeywordByRoleIds(fun func(c *gin.Context, roleIds []uint) []string) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findRoleKeywordByRoleIds = fun
		}
	}
}

func WithUploadSaveDir(dir string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uploadSaveDir = dir
	}
}

func WithUploadSingleMaxSize(size int64) func(*Options) {
	return func(options *Options) {
		if size > 0 {
			getOptionsOrSetDefault(options).uploadSingleMaxSize = size
		}
	}
}

func WithUploadMergeConcurrentCount(count int) func(*Options) {
	return func(options *Options) {
		if count > 1 {
			getOptionsOrSetDefault(options).uploadMergeConcurrentCount = count
		}
	}
}

func WithUploadMinio(minio *oss.MinioOss) func(*Options) {
	return func(options *Options) {
		if minio != nil {
			getOptionsOrSetDefault(options).uploadMinio = minio
		}
	}
}

func WithUploadMinioBucket(bucket string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uploadMinioBucket = bucket
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			cache: false,
		}
	}
	return options
}

func ParseOptions(options ...func(*Options)) *Options {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// check ops
	if ops.cache {
		query.NewRedis(ops.cacheOps...)
	}
	query.NewMySql(ops.dbOps...)
	return ops
}

func (ops *Options) addCtx(ctx context.Context) {
	if ops.cache {
		ops.cacheOps = append(ops.cacheOps, query.WithRedisCtx(ctx))
	}
	ops.dbOps = append(ops.dbOps, query.WithMysqlCtx(ctx))
}
