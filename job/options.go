package job

import (
	"context"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	glogger "gorm.io/gorm/logger"
	"time"

	"os"
)

type Options struct {
	logger        glogger.Interface
	ctx           context.Context
	prefix        string
	AutoRequestId bool
}

func WithLogger(l glogger.Interface) func(*Options) {
	return func(options *Options) {
		if l != nil {
			getOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithLoggerLevel(level glogger.LogLevel) func(*Options) {
	return func(options *Options) {
		l := options.logger
		if options.logger == nil {
			l = getOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogMode(level)
	}
}

func WithContext(ctx context.Context) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).ctx = ctx
	}
}

func WithPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithAutoRequestId(options *Options) {
	getOptionsOrSetDefault(options).AutoRequestId = true
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		enConfig := zap.NewProductionEncoderConfig()
		enConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(carbon.Time2Carbon(t).ToRfc3339String())
		}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(enConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
			zapcore.DebugLevel,
		)
		l := zap.New(core)
		return &Options{
			logger: logger.New(
				l,
				logger.Config{
					LineNumLevel: 2,
					Config: glogger.Config{
						Colorful: true,
					},
				},
			),
		}
	}
	return options
}

type DriverOptions struct {
	logger glogger.Interface
	ctx    context.Context
	prefix string
}

func WithDriverLogger(l glogger.Interface) func(*DriverOptions) {
	return func(options *DriverOptions) {
		if l != nil {
			getDriverOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithDriverLoggerLevel(level glogger.LogLevel) func(*DriverOptions) {
	return func(options *DriverOptions) {
		l := options.logger
		if options.logger == nil {
			l = getDriverOptionsOrSetDefault(options).logger
		}
		options.logger = l.LogMode(level)
	}
}

func WithDriverContext(ctx context.Context) func(*DriverOptions) {
	return func(options *DriverOptions) {
		getDriverOptionsOrSetDefault(options).ctx = ctx
	}
}

func WithDriverPrefix(prefix string) func(*DriverOptions) {
	return func(options *DriverOptions) {
		getDriverOptionsOrSetDefault(options).prefix = prefix
	}
}

func getDriverOptionsOrSetDefault(options *DriverOptions) *DriverOptions {
	if options == nil {
		enConfig := zap.NewProductionEncoderConfig()
		enConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(carbon.Time2Carbon(t).ToRfc3339String())
		}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(enConfig),
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
			zapcore.DebugLevel,
		)
		l := zap.New(core)
		return &DriverOptions{
			logger: logger.New(
				l,
				logger.Config{
					LineNumLevel: 2,
					Config: glogger.Config{
						Colorful: true,
					},
				},
			),
		}
	}
	return options
}
