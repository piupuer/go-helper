package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
)

type Options struct {
	ctx           context.Context
	level         Level
	output        io.Writer
	category      string
	json          bool
	lineNum       bool
	lineNumPrefix string
	lineNumLevel  int
	keepSourceDir bool
	keepVersion   bool
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

func WithOutput(output io.Writer) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).output = output
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

func WithKeepSourceDir(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).keepSourceDir = flag
	}
}

func WithKeepVersion(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).keepVersion = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			ctx:          context.Background(),
			level:        Level(logrus.DebugLevel),
			lineNum:      true,
			lineNumLevel: 1,
			keepVersion:  true,
		}
	}
	return options
}
