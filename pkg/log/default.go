package log

import "context"

var DefaultWrapper *Wrapper

func init() {
	DefaultWrapper = &Wrapper{
		log:    New(),
		fields: map[string]interface{}{},
	}
}

func NewDefaultWrapper() *Wrapper {
	return DefaultWrapper
}

func Trace(args ...interface{}) {
	DefaultWrapper.Trace(args...)
}

func Debug(args ...interface{}) {
	DefaultWrapper.Debug(args...)
}

func Info(args ...interface{}) {
	DefaultWrapper.Info(args...)
}

func Warn(args ...interface{}) {
	DefaultWrapper.Warn(args...)
}

func Error(args ...interface{}) {
	DefaultWrapper.Error(args...)
}

func Fatal(args ...interface{}) {
	DefaultWrapper.Fatal(args...)
}

func WithError(err error) *Wrapper {
	return DefaultWrapper.WithError(err)
}

func WithField(k string, v interface{}) *Wrapper {
	return DefaultWrapper.WithFields(map[string]interface{}{
		k: v,
	})
}

func WithFields(fields map[string]interface{}) *Wrapper {
	return DefaultWrapper.WithFields(fields)
}

func WithRequestId(ctx context.Context) *Wrapper {
	return DefaultWrapper.WithRequestId(ctx)
}
