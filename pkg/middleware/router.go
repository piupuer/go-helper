package middleware

import "github.com/piupuer/go-helper/pkg/logger"

func PrintRouter(options ...func(*PrintRouterOptions)) func(httpMethod string, absolutePath string, handlerName string, nuHandlers int) {
	ops := getPrintRouterOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		logger.WithRequestId(ops.ctx).Debug("[gin-route] %-6s %-40s --> %s (%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	}
}
