package logger

import (
	"context"
)

var defaultWrapper Wrapper

func init() {
	defaultWrapper = Wrapper{
		log:    New(),
		fields: map[string]interface{}{},
	}
}

func NewDefaultWrapper() *Wrapper {
	return &defaultWrapper
}

func Trace(args ...interface{}) {
	defaultWrapper.Trace(args...)
}

func Debug(args ...interface{}) {
	defaultWrapper.Debug(args...)
}

func Info(args ...interface{}) {
	defaultWrapper.Info(args...)
}

func Warn(args ...interface{}) {
	defaultWrapper.Warn(args...)
}

func Error(args ...interface{}) {
	defaultWrapper.Error(args...)
}

func Fatal(args ...interface{}) {
	defaultWrapper.Fatal(args...)
}

func WithError(err error) *Wrapper {
	return defaultWrapper.WithError(err)
}

func WithField(k string, v interface{}) *Wrapper {
	return defaultWrapper.WithFields(map[string]interface{}{
		k: v,
	})
}

func WithFields(fields map[string]interface{}) *Wrapper {
	return defaultWrapper.WithFields(fields)
}

func WithRequestId(ctx context.Context) *Wrapper {
	return defaultWrapper.WithRequestId(ctx)
}
