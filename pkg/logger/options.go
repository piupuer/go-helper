package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	level         zapcore.Level
	colorful      bool
	lineNumPrefix string
	lineNumLevel  int
	keepSourceDir bool
	lumber        bool
	lumberOps     LumberjackOption
}

type LumberjackOption struct {
	lumberjack.Logger
	LogPath string
}

func WithLevel(level zapcore.Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = level
	}
}

func WithColorful(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).colorful = flag
	}
}

func WithLineNumPrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lineNumPrefix = prefix
	}
}

func WithLineNumLevel(level int) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lineNumLevel = level
	}
}

func WithSkipLumber(options *Options) {
	getOptionsOrSetDefault(options).lumber = false
}

func WithKeepSourceDir(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).keepSourceDir = flag
	}
}

func WithLumberjackOption(option LumberjackOption) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lumberOps = option
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			level:         zapcore.DebugLevel,
			lineNumLevel:  2,
			keepSourceDir: true,
			lumber:        true,
			lumberOps: LumberjackOption{
				Logger: lumberjack.Logger{
					MaxSize:    50,
					MaxAge:     30,
					MaxBackups: 100,
					LocalTime:  true,
					Compress:   true,
				},
				LogPath: "logs",
			},
		}
	}
	return options
}
