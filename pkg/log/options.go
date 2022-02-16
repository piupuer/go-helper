package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

type FileWithLineNumOptions struct {
	skipGorm   bool
	skipHelper bool
}

func WithSkipGorm(flag bool) func(*FileWithLineNumOptions) {
	return func(options *FileWithLineNumOptions) {
		getFileWithLineNumOptionsOrSetDefault(options).skipGorm = flag
	}
}

func WithSkipHelper(flag bool) func(*FileWithLineNumOptions) {
	return func(options *FileWithLineNumOptions) {
		getFileWithLineNumOptionsOrSetDefault(options).skipHelper = flag
	}
}

func getFileWithLineNumOptionsOrSetDefault(options *FileWithLineNumOptions) *FileWithLineNumOptions {
	if options == nil {
		return &FileWithLineNumOptions{}
	}
	return options
}

type Options struct {
	level          Level
	output         io.Writer
	category       string
	json           bool
	lineNum        bool
	lineNumPrefix  string
	lineNumLevel   int
	lineNumSource  bool
	lineNumVersion bool
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

func WithLineNumSource(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lineNumSource = flag
	}
}

func WithLineNumVersion(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).lineNumVersion = flag
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			level:          Level(logrus.DebugLevel),
			lineNum:        true,
			lineNumLevel:   1,
			lineNumVersion: true,
		}
	}
	return options
}
