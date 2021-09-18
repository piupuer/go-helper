package logger

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	RequestIdContextKey = "RequestId"
)

var (
	sourceDir = ""
)

func init() {
	// get runtime root
	_, file, _, _ := runtime.Caller(0)
	sourceDir = strings.TrimSuffix(file, fmt.Sprintf("%slogger%slogger.go", string(os.PathSeparator), string(os.PathSeparator)))
}

// zap logger for gorm
type Logger struct {
	Config
	log                                            *zap.Logger
	normalStr, traceStr, traceErrStr, traceWarnStr string
}

type Config struct {
	logger.Config
	LineNumPrefix string
	LineNumLevel  int
}

// New logger like gorm2
func New(zapLogger *zap.Logger, config Config) *Logger {
	var (
		normalStr    = "%v%s "
		traceStr     = "%v%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%v%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%v%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		normalStr = logger.Cyan + "%v" + logger.Blue + "%s " + logger.Reset
		traceStr = logger.Cyan + "%v" + logger.Blue + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Cyan + "%v" + logger.Blue + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.Cyan + "%v" + logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}

	if config.LineNumLevel <= 0 {
		config.LineNumLevel = 3
	}

	l := &Logger{
		log:          zapLogger,
		Config:       config,
		normalStr:    normalStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
	return l
}

// LogMode gorm log mode
// LogMode log mode
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Debug print info
func (l Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.log.Core().Enabled(zapcore.DebugLevel) {
		requestId := l.getRequestId(ctx)
		l.log.Sugar().Debugf(l.normalStr+format, append([]interface{}{requestId, l.removePrefix(utils.FileWithLineNum(), fileWithLineNum())}, args...)...)
	}
}

// Info print info
func (l Logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.log.Core().Enabled(zapcore.InfoLevel) {
		requestId := l.getRequestId(ctx)
		l.log.Sugar().Infof(l.normalStr+format, append([]interface{}{requestId, l.removePrefix(utils.FileWithLineNum(), fileWithLineNum())}, args...)...)
	}
}

// Warn print warn messages
func (l Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.log.Core().Enabled(zapcore.WarnLevel) {
		requestId := l.getRequestId(ctx)
		l.log.Sugar().Warnf(l.normalStr+format, append([]interface{}{requestId, l.removePrefix(utils.FileWithLineNum(), fileWithLineNum())}, args...)...)
	}
}

// Error print error messages
func (l Logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.log.Core().Enabled(zapcore.ErrorLevel) {
		requestId := l.getRequestId(ctx)
		l.log.Sugar().Errorf(l.normalStr+format, append([]interface{}{requestId, l.removePrefix(utils.FileWithLineNum(), fileWithLineNum())}, args...)...)
	}
}

// Trace print sql message
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.log.Core().Enabled(zapcore.DPanicLevel) || l.LogLevel <= logger.Silent {
		return
	}
	lineNum := l.removePrefix(utils.FileWithLineNum(), fileWithLineNum())
	elapsed := time.Since(begin)
	elapsedF := float64(elapsed.Nanoseconds()) / 1e6
	sql, rows := fc()
	row := "-"
	if rows > -1 {
		row = fmt.Sprintf("%d", rows)
	}
	requestId := l.getRequestId(ctx)
	switch {
	case l.log.Core().Enabled(zapcore.ErrorLevel) && err != nil && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		l.log.Error(fmt.Sprintf(l.traceErrStr, requestId, lineNum, err, elapsedF, row, sql))
	case l.log.Core().Enabled(zapcore.WarnLevel) && elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.log.Warn(fmt.Sprintf(l.traceWarnStr, requestId, lineNum, slowLog, elapsedF, row, sql))
	case l.log.Core().Enabled(zapcore.DebugLevel):
		l.log.Debug(fmt.Sprintf(l.traceStr, requestId, lineNum, elapsedF, row, sql))
	case l.log.Core().Enabled(zapcore.InfoLevel):
		l.log.Info(fmt.Sprintf(l.traceStr, requestId, lineNum, elapsedF, row, sql))
	}
}

func (l Logger) GetZapLog() *zap.Logger {
	return l.log
}

func (l Logger) getRequestId(ctx context.Context) string {
	var v interface{}
	vi := reflect.ValueOf(ctx)
	if vi.Kind() == reflect.Ptr {
		if !vi.IsNil() {
			v = ctx.Value(RequestIdContextKey)
		}
	}
	requestId := ""
	if v != nil {
		requestId = fmt.Sprintf("%v ", v)
	}
	return requestId
}

func fileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, sourceDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func (l Logger) removePrefix(s1 string, s2 string) string {
	res1 := l.removeBaseDir(s1)
	res2 := l.removeBaseDir(s2)
	if len(res1) < len(res2) && !strings.HasPrefix(s1, sourceDir) {
		return res1
	}
	return res2
}

func (l Logger) removeBaseDir(s string) string {
	if strings.HasPrefix(s, sourceDir) {
		s = strings.TrimPrefix(s, path.Dir(sourceDir)+"/")
	}
	if strings.HasPrefix(s, l.LineNumPrefix) {
		s = strings.TrimPrefix(s, l.LineNumPrefix)
	}
	arr := strings.Split(s, "@")
	if len(arr) == 2 {
		s = fmt.Sprintf("%s@%s", l.getParentDir(arr[0], l.LineNumLevel), arr[1])
	}
	return s
}

func (l Logger) getParentDir(dir string, index int) string {
	d, b := filepath.Split(filepath.Clean(dir))
	parentDir := ""
	if index > 0 {
		parentDir = l.getParentDir(d, index-1)
	}
	if parentDir != "" {
		return fmt.Sprintf("%s%s%s", parentDir, string(os.PathSeparator), b)
	}
	return b
}
