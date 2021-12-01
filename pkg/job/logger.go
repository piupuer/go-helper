package job

import (
	"context"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/robfig/cron/v3"
	"strings"
)

type dcronLogger struct {
	l logger.Interface
}

func (c dcronLogger) Printf(format string, args ...interface{}) {
	ctx := context.Background()
	if strings.HasPrefix(format, dcronInfoPrefix) {
		c.l.Info(ctx, strings.TrimPrefix(format, dcronInfoPrefix), args...)
	} else if strings.HasPrefix(format, dcronErrorPrefix) {
		c.l.Error(ctx, strings.TrimPrefix(format, dcronErrorPrefix), args...)
	}
}

type CronLogger struct {
	ops CronOptions
}

func (cl CronLogger) Printf(msg string, args ...interface{}) {
	cl.ops.logger.Info(cl.ops.ctx, msg, args...)
}

func NewCronLogger(options ...func(*CronOptions)) cron.Logger {
	ops := getCronOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return cron.VerbosePrintfLogger(CronLogger{ops: *ops})
}
