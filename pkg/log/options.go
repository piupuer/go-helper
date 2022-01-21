package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
)

type Options struct {
	ctx      context.Context
	level    Level
	category string
	json     bool
	lineNum  bool
	output   io.Writer
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !interfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithLevel(level Level) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).level = level
	}
}

func WithCategory(s string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).category = s
	}
}

func WithJson(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).json = flag
	}
}

func WithLineNum(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lineNum = flag
	}
}

func WithOutput(output io.Writer) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).output = output
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			ctx:     context.Background(),
			level:   Level(logrus.DebugLevel),
			lineNum: true,
			json:    false,
		}
	}
	return options
}
