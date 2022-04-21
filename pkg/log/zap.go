package log

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type zapLog struct {
	log *zap.Logger
	ops Options
}

func newZap(ops *Options) *zapLog {
	enConfig := zap.NewProductionEncoderConfig()
	enConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	enConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(carbon.Time2Carbon(t).ToRfc3339String())
	}
	encoder := zapcore.NewConsoleEncoder(enConfig)
	if ops.json {
		enConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(enConfig)
	}
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		loggerToZapLevel(ops.level),
	)
	l := zapLog{
		log: zap.New(core),
		ops: *ops,
	}
	return &l
}

func (l *zapLog) Options() Options {
	return l.ops
}

func (l *zapLog) WithFields(fields map[string]interface{}) Interface {
	data := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		data = append(data, zap.Any(k, v))
	}
	ll := &zapLog{
		log: l.log.With(data...),
		ops: l.ops,
	}
	return ll
}

func (l *zapLog) Log(level Level, args ...interface{}) {
	ll := loggerToZapLevel(level)
	msg := fmt.Sprint(args...)
	switch ll {
	case zap.DebugLevel:
		l.log.Debug(msg)
	case zap.InfoLevel:
		l.log.Info(msg)
	case zap.WarnLevel:
		l.log.Warn(msg)
	case zap.ErrorLevel:
		l.log.Error(msg)
	case zap.FatalLevel:
		l.log.Fatal(msg)
	}
}

func (l *zapLog) Logf(level Level, format string, args ...interface{}) {
	ll := loggerToZapLevel(level)
	msg := fmt.Sprintf(format, args...)
	switch ll {
	case zap.DebugLevel:
		l.log.Debug(msg)
	case zap.InfoLevel:
		l.log.Info(msg)
	case zap.WarnLevel:
		l.log.Warn(msg)
	case zap.ErrorLevel:
		l.log.Error(msg)
	case zap.FatalLevel:
		l.log.Fatal(msg)
	}
}

func loggerToZapLevel(level Level) zapcore.Level {
	switch level {
	case TraceLevel, DebugLevel:
		return zap.DebugLevel
	case InfoLevel:
		return zap.InfoLevel
	case WarnLevel:
		return zap.WarnLevel
	case ErrorLevel:
		return zap.ErrorLevel
	case FatalLevel:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
