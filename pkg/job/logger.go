package job

import (
	"github.com/robfig/cron/v3"
	"strings"
)

type dcronLogger struct {
	ops CronOptions
}

func newDCronLogger(options ...func(*CronOptions)) *dcronLogger {
	ops := getCronOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &dcronLogger{ops: *ops}
}

func (c dcronLogger) Printf(format string, args ...interface{}) {
	if strings.HasPrefix(format, dcronInfoPrefix) {
		c.ops.logger.Info(append([]interface{}{strings.TrimPrefix(format, dcronInfoPrefix)}, args...)...)
	} else if strings.HasPrefix(format, dcronErrorPrefix) {
		c.ops.logger.Error(append([]interface{}{strings.TrimPrefix(format, dcronErrorPrefix)}, args...)...)
	}
}

type CronLogger struct {
	ops CronOptions
}

func (cl CronLogger) Printf(msg string, args ...interface{}) {
	cl.ops.logger.Info(append([]interface{}{msg}, args...)...)
}

func NewCronLogger(options ...func(*CronOptions)) cron.Logger {
	ops := getCronOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return cron.VerbosePrintfLogger(CronLogger{ops: *ops})
}
