package log

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	"gorm.io/gorm/logger"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	logDir    = ""
	helperDir = ""
)

func init() {
	// get runtime root
	_, file, _, _ := runtime.Caller(0)
	logDir = regexp.MustCompile(`logger\.go`).ReplaceAllString(file, "")
	helperDir = regexp.MustCompile(`pkg.log.logger\.go`).ReplaceAllString(file, "")
}

// Interface logger interface
type Interface interface {
	Options() Options
	WithFields(fields map[string]interface{}) Interface
	Log(level Level, v ...interface{})
	Logf(level Level, format string, v ...interface{})
}

type Config struct {
	ops  Options
	gorm logger.Config
}

func New(options ...func(*Options)) (l Interface) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	switch ops.category {
	case constant.LogCategoryZap:
		l = newZap(ops)
	case constant.LogCategoryLogrus:
		l = newLogrus(ops)
	default:
		l = newLogrus(ops)
	}
	return l
}

func fileWithLineNum(ops Options, options ...func(*FileWithLineNumOptions)) string {
	lineOps := getFileWithLineNumOptionsOrSetDefault(nil)
	for _, f := range options {
		f(lineOps)
	}
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if lineOps.skipGorm && (strings.Contains(file, "gorm.io")) {
			continue
		}
		if lineOps.skipHelper && (strings.Contains(file, helperDir)) {
			continue
		}
		if ok && (!strings.HasPrefix(file, logDir) || strings.HasSuffix(file, "_test.go")) && !strings.Contains(file, "src/runtime") {
			return removeBaseDir(file+":"+strconv.FormatInt(int64(line), 10), ops)
		}
	}

	return ""
}

func removeBaseDir(s string, ops Options) string {
	sep := string(os.PathSeparator)
	if !ops.lineNumSource && strings.HasPrefix(s, helperDir) {
		s = strings.TrimPrefix(s, path.Dir(strings.TrimSuffix(helperDir, sep))+sep)
	}
	if strings.HasPrefix(s, ops.lineNumPrefix) {
		s = strings.TrimPrefix(s, ops.lineNumPrefix)
	}
	arr := strings.Split(s, "@")
	if len(arr) == 2 {
		arr1 := strings.Split(arr[0], sep)
		arr2 := strings.Split(arr[1], sep)
		if ops.lineNumLevel > 0 {
			if ops.lineNumLevel < len(arr1) {
				arr1 = arr1[len(arr1)-ops.lineNumLevel:]
			}
		}
		if !ops.lineNumVersion {
			arr2 = arr2[1:]
		}
		s1 := strings.Join(arr1, sep)
		s2 := strings.Join(arr2, sep)
		if !ops.lineNumVersion {
			s = fmt.Sprintf("%s%s%s", s1, sep, s2)
		} else {
			s = fmt.Sprintf("%s@%s", s1, s2)
		}
	}
	return s
}
