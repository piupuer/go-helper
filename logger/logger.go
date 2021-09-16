package logger

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Logger interface {
	Debug(ctx context.Context, format string, args ...interface{})
	Info(ctx context.Context, format string, args ...interface{})
	Warn(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// LogLevel
type LogLevel int

const (
	Debug LogLevel = iota + 1
	Info
	Warn
	Error
)

const RequestIdContextKey = "RequestId"

type Config struct {
	Level LogLevel
}

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

var sourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	sourceDir = regexp.MustCompile(`logger\.go`).ReplaceAllString(file, "")
}

func New(writer Writer, config Config) Logger {
	l := &logger{
		Config:    config,
		Writer:    writer,
		normalStr: "%s%s%s ",
	}
	return l
}

type logger struct {
	Config
	Writer
	normalStr string
}

func (l logger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= Debug {
		requestId := getRequestId(ctx)
		l.Printf(l.normalStr+format, append([]interface{}{"DEBUG", requestId, fileWithLineNum()}, args...)...)
	}
}

func (l logger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= Info {
		requestId := getRequestId(ctx)
		l.Printf(l.normalStr+format, append([]interface{}{"INFO", requestId, fileWithLineNum()}, args...)...)
	}
}

func (l logger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= Warn {
		requestId := getRequestId(ctx)
		l.Printf(l.normalStr+format, append([]interface{}{"WARN", requestId, fileWithLineNum()}, args...)...)
	}
}

func (l logger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.Level <= Error {
		requestId := getRequestId(ctx)
		l.Printf(l.normalStr+format, append([]interface{}{"ERROR", requestId, fileWithLineNum()}, args...)...)
	}
}

func getRequestId(ctx context.Context) string {
	var v interface{}
	vi := reflect.ValueOf(ctx)
	if vi.Kind() == reflect.Ptr {
		if !vi.IsNil() {
			v = ctx.Value(RequestIdContextKey)
		}
	}
	requestId := " "
	if v != nil {
		requestId = fmt.Sprintf(" %v ", v)
	}
	return requestId
}

func fileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, sourceDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}
