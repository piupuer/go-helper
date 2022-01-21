package log

import "gorm.io/gorm/logger"

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

func (l Level) Enabled(lvl Level) bool {
	return l >= lvl
}

func (l Level) LevelToGorm() logger.LogLevel {
	switch l {
	case FatalLevel, ErrorLevel:
		return logger.Error
	case WarnLevel:
		return logger.Warn
	case InfoLevel, DebugLevel, TraceLevel:
		return logger.Info
	default:
		return logger.Silent
	}
}
