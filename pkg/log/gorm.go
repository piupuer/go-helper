package log

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type gormLogger struct {
	Config
	normalStr, traceStr, traceErrStr, traceWarnStr string
}

func NewDefaultGormLogger() logger.Interface {
	return NewGormLogger(Config{
		gorm: logger.Config{
			SlowThreshold: 200,
		},
	})
}

func NewGormLogger(config Config) logger.Interface {
	var (
		normalStr    = "%v%s "
		traceStr     = "%v%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%v%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%v%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.gorm.Colorful {
		normalStr = logger.Cyan + "%v" + logger.Blue + "%s " + logger.Reset
		traceStr = logger.Cyan + "%v" + logger.Blue + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Cyan + "%v" + logger.Blue + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.Cyan + "%v" + logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}

	l := gormLogger{
		Config:       config,
		normalStr:    normalStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
	return &l
}

func (l *gormLogger) getLogger(ctx context.Context) Interface {
	requestId := getRequestId(ctx)
	if requestId != "" {
		return New().WithFields(map[string]interface{}{
			constant.MiddlewareRequestIdCtxKey: requestId,
		})
	}
	return New()
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.gorm.LogLevel = level
	return &newLogger
}

func (l gormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Info {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			"lineNum": lineNum,
		})
		log.Logf(InfoLevel, l.normalStr+format, args...)
	}
}

func (l gormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Warn {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			"lineNum": lineNum,
		})
		log.Logf(WarnLevel, l.normalStr+format, args...)
	}
}

func (l gormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Error {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			"lineNum": lineNum,
		})
		log.Logf(ErrorLevel, l.normalStr+format, args...)
	}
}

func (l gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.gorm.LogLevel > logger.Silent {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		elapsed := time.Since(begin)
		elapsedF := float64(elapsed.Nanoseconds()) / 1e6
		sql, rows := fc()
		row := "-"
		if rows > -1 {
			row = fmt.Sprintf("%d", rows)
		}
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			"lineNum": lineNum,
		})
		switch {
		case l.gorm.LogLevel >= logger.Error && err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			log.Logf(TraceLevel, l.traceErrStr, err, elapsedF, row, sql)
		case l.gorm.LogLevel >= logger.Warn && l.gorm.SlowThreshold > 0 && elapsed > l.gorm.SlowThreshold:
			slowLog := fmt.Sprintf("SLOW SQL >= %v", 200)
			log.Logf(TraceLevel, l.traceErrStr, err, slowLog, elapsedF, row, sql)
		case l.gorm.LogLevel == logger.Info:
			log.Logf(TraceLevel, l.traceErrStr, err, elapsedF, row, sql)
		}
	}
}
