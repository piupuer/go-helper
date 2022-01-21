package log

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	utils2 "github.com/piupuer/go-helper/pkg/utils"
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
	helperDir = regexp.MustCompile(`go-helper.pkg.log.logger\.go`).ReplaceAllString(file, "")
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

func getRequestId(ctx context.Context) (id string) {
	if utils2.InterfaceIsNil(ctx) {
		return
	}
	// get value from context
	requestIdValue := ctx.Value(constant.MiddlewareRequestIdCtxKey)
	if item, ok := requestIdValue.(string); ok && item != "" {
		id = item
	}
	return
}

func fileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, logDir) || strings.HasSuffix(file, "_test.go")) && !strings.Contains(file, "src/runtime") {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func removePrefix(s1 string, s2 string) string {
	res1 := removeBaseDir(s1)
	res2 := removeBaseDir(s2)
	if strings.HasPrefix(s1, logDir) {
		return res2
	}
	f1 := len(res1) <= len(res2)
	f2 := strings.HasPrefix(s1, logDir)
	// src/runtime may be in go routine
	if strings.Contains(res2, "src/runtime") || (f1 || !f1 && f2) {
		return res1
	}
	return res2
}

func removeBaseDir(s string) string {
	sep := string(os.PathSeparator)
	if strings.HasPrefix(s, helperDir) {
		s = strings.TrimPrefix(s, path.Dir(helperDir)+"/")
	}
	arr := strings.Split(s, "@")
	if len(arr) == 2 {
		arr1 := strings.Split(arr[0], sep)
		arr2 := strings.Split(arr[1], sep)
		if len(arr1) > 3 {
			arr1 = arr1[len(arr1)-3:]
		}
		// arr2 = arr2[1:]
		s1 := strings.Join(arr1, sep)
		s2 := strings.Join(arr2, sep)
		// s = fmt.Sprintf("%s%s%s", s1, sep, s2)
		s = fmt.Sprintf("%s@%s", s1, s2)
	}
	return s
}
