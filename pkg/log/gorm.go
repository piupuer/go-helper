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
	normalStr, normalErrStr, slowStr, slowErrStr string
}

func NewDefaultGormLogger() logger.Interface {
	return NewGormLogger(Config{
		gorm: logger.Config{
			SlowThreshold: 200 * time.Millisecond,
		},
	})
}

func NewGormLogger(config Config) logger.Interface {
	var (
		normalStr    = "[%.3fms] [rows:%v] %s"
		slowStr      = "[%.3fms(slow)] [rows:%v] %s"
		normalErrStr = "%s\n[%.3fms] [rows:%v] %s"
		slowErrStr   = "%s\n[%.3fms(slow)] [rows:%v] %s"
	)

	if config.gorm.Colorful {
		normalStr = logger.Green + "[%.3fms] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		slowStr = logger.Yellow + "[%.3fms(slow)] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		normalErrStr = logger.RedBold + "%s\n" + logger.Reset + logger.Green + "[%.3fms] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		slowErrStr = logger.RedBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms(slow)] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}

	l := gormLogger{
		Config:       config,
		normalStr:    normalStr,
		slowStr:      slowStr,
		normalErrStr: normalErrStr,
		slowErrStr:   slowErrStr,
	}
	return &l
}

func (l *gormLogger) getLogger(ctx context.Context) Interface {
	return DefaultWrapper.WithRequestId(ctx).log.WithFields(DefaultWrapper.WithRequestId(ctx).fields)
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
			constant.LogLineNumKey: lineNum,
		})
		log.Logf(InfoLevel, format, args...)
	}
}

func (l gormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Warn {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			constant.LogLineNumKey: lineNum,
		})
		log.Logf(WarnLevel, format, args...)
	}
}

func (l gormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Error {
		lineNum := removePrefix(utils.FileWithLineNum(), fileWithLineNum(), l.ops)
		log := l.getLogger(ctx).WithFields(map[string]interface{}{
			constant.LogLineNumKey: lineNum,
		})
		log.Logf(ErrorLevel, format, args...)
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
			constant.LogLineNumKey: lineNum,
		})
		switch {
		case l.gorm.LogLevel >= logger.Error && err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			if l.gorm.SlowThreshold > 0 && elapsed > l.gorm.SlowThreshold {
				log.Logf(ErrorLevel, l.slowErrStr, err, elapsedF, row, sql)
			} else {
				log.Logf(ErrorLevel, l.normalErrStr, err, elapsedF, row, sql)
			}
		case l.gorm.LogLevel >= logger.Warn && l.gorm.SlowThreshold > 0 && elapsed > l.gorm.SlowThreshold:
			log.Logf(WarnLevel, l.slowStr, elapsedF, row, sql)
		case l.gorm.LogLevel == logger.Info:
			log.Logf(InfoLevel, l.normalStr, elapsedF, row, sql)
		}
	}
}
