package binlog

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
)

type Options struct {
	logger        *logger.Wrapper
	ctx           context.Context
	dsn           *mysql.Config
	db            *gorm.DB
	redis         redis.UniversalClient
	ignores       []string
	models        []interface{}
	serverId      uint32
	executionPath string
	binlogPos     string
}

func WithLogger(l *logger.Wrapper) func(*Options) {
	return func(options *Options) {
		if l != nil {
			getOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !utils.InterfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
			options.logger = options.logger.WithRequestId(ctx)
		}
	}
}

func WithDsn(dsn *mysql.Config) func(*Options) {
	return func(options *Options) {
		if dsn != nil {
			getOptionsOrSetDefault(options).dsn = dsn
		}
	}
}

func WithDb(db *gorm.DB) func(*Options) {
	return func(options *Options) {
		if db != nil {
			getOptionsOrSetDefault(options).db = db
		}
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithIgnore(ignores ...string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).ignores = append(getOptionsOrSetDefault(options).ignores, ignores...)
	}
}

func WithModels(models ...interface{}) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).models = append(getOptionsOrSetDefault(options).models, models...)
	}
}

func WithServerId(serverId uint32) func(*Options) {
	return func(options *Options) {
		if serverId > 0 {
			getOptionsOrSetDefault(options).serverId = serverId
		}
	}
}

func WithExecutionPath(p string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).executionPath = p
	}
}

func WithBinlogPos(key string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlogPos = key
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			logger:        logger.NewWrapper(logger.New()),
			ignores:       []string{},
			serverId:      100,
			executionPath: "mysqldump",
			binlogPos:     "mysql_binlog_pos",
		}
	}
	return options
}
