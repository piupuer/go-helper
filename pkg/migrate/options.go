package migrate

import (
	"context"
	"embed"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Options struct {
	ctx         context.Context
	driver      string
	uri         string
	lockName    string
	changeTable string
	fs          embed.FS
	fsRoot      string
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !utils.InterfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithDriver(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).driver = s
	}
}

func WithUri(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uri = s
	}
}

func WithLockName(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lockName = s
	}
}

func WithChangeTable(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).changeTable = s
	}
}

func WithFs(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fs = fs
	}
}

func WithFsRoot(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fsRoot = s
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			driver:      "mysql",
			uri:         "root:root@tcp(127.0.0.1:4306)/gin_web?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=UTC&timeout=10000ms",
			lockName:    "MigrationLock",
			changeTable: "schema_migrations",
		}
	}
	return options
}
