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

func WithSkipLumber(options *Options) {
	getOptionsOrSetDefault(options).lumber = false
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
