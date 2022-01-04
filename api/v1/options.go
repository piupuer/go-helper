package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/oss"
	"github.com/piupuer/go-helper/pkg/query"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

type Options struct {
	logger                     logger.Interface
	binlog                     bool
	binlogOps                  []func(options *query.RedisOptions)
	dbOps                      []func(options *query.MysqlOptions)
	redis                      redis.UniversalClient
	cachePrefix                string
	operationAllowedToDelete   bool
	getCurrentUser             func(c *gin.Context) ms.User
	findRoleKeywordByRoleIds   func(c *gin.Context, roleIds []uint) []string
	findRoleByIds              func(c *gin.Context, roleIds []uint) []ms.Role
	findUserByIds              func(c *gin.Context, userIds []uint) []ms.User
	beforeResetUserPwd         func(c *gin.Context, r *req.ResetUserPwd) error
	getUserLoginStatus         func(c *gin.Context, r *req.UserStatus) error
	getFsmSubmitterDetail      func(c *gin.Context, detail req.FsmSubmitterDetail) []resp.FsmSubmitterDetail
	updateFsmSubmitterDetail   func(c *gin.Context, detail req.UpdateFsmSubmitterDetail) error
	uploadSaveDir              string
	uploadSingleMaxSize        int64
	uploadMergeConcurrentCount int
	uploadMinio                *oss.MinioOss
	uploadMinioBucket          string
	MessageHub                 bool
	messageHubOps              []func(options *query.MessageHubOptions)
}

func WithLogger(l logger.Interface) func(*Options) {
	return func(options *Options) {
		if l != nil {
			getOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithLoggerLevel(level logger.Level) func(*Options) {
	return func(options *Options) {
		l := options.logger
		if options.logger == nil {
			l = getOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogLevel(level)
	}
}

func WithBinlog(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlog = flag
	}
}

func WithBinlogOps(ops ...func(options *query.RedisOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlogOps = append(getOptionsOrSetDefault(options).binlogOps, ops...)
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

func WithCachePrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).cachePrefix = prefix
	}
}

func WithOperationAllowedToDelete(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).operationAllowedToDelete = flag
	}
}

func WithGetCurrentUser(fun func(c *gin.Context) ms.User) func(*Options) {
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

func WithFindRoleByIds(fun func(c *gin.Context, roleIds []uint) []ms.Role) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findRoleByIds = fun
		}
	}
}

func WithFindUserByIds(fun func(c *gin.Context, userIds []uint) []ms.User) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findUserByIds = fun
		}
	}
}

func WithBeforeResetUserPwd(fun func(c *gin.Context, r *req.ResetUserPwd) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).beforeResetUserPwd = fun
		}
	}
}

func WithGetUserLoginStatus(fun func(c *gin.Context, r *req.UserStatus) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getUserLoginStatus = fun
		}
	}
}

func WithFsmGetFsmSubmitterDetail(fun func(c *gin.Context, detail req.FsmSubmitterDetail) []resp.FsmSubmitterDetail) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getFsmSubmitterDetail = fun
		}
	}
}

func WithFsmUpdateFsmSubmitterDetail(fun func(c *gin.Context, detail req.UpdateFsmSubmitterDetail) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).updateFsmSubmitterDetail = fun
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

func WithMessageHub(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).MessageHub = flag
	}
}

func WithMessageHubOps(ops ...func(options *query.MessageHubOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).messageHubOps = append(getOptionsOrSetDefault(options).messageHubOps, ops...)
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			logger:                     logger.DefaultLogger(),
			binlog:                     false,
			cachePrefix:                "v1_cache",
			operationAllowedToDelete:   true,
			uploadSaveDir:              "upload",
			uploadSingleMaxSize:        32,
			uploadMergeConcurrentCount: 10,
			MessageHub:                 true,
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
	if ops.binlog {
		if ops.redis != nil {
			ops.binlogOps = append(ops.binlogOps, query.WithRedisClient(ops.redis))
		}
		query.NewRedis(ops.binlogOps...)
	}
	if ops.redis != nil {
		ops.dbOps = append(ops.dbOps, query.WithMysqlRedis(ops.redis))
	}
	query.NewMySql(ops.dbOps...)
	return ops
}

func (ops *Options) addCtx(ctx context.Context) {
	if ops.binlog {
		ops.binlogOps = append(ops.binlogOps, query.WithRedisCtx(ctx))
	}
	ops.dbOps = append(ops.dbOps, query.WithMysqlCtx(ctx))
}
