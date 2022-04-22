package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

func AccessLog(options ...func(*AccessLogOptions)) grpc.UnaryServerInterceptor {
	ops := getAccessLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		rp, err := handler(ctx, r)

		endTime := time.Now()
		// calc request exec time
		execTime := endTime.Sub(startTime).String()

		fullMethod := info.FullMethod
		addr := ""
		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}
		code := status.Code(err).String()

		detail := make(map[string]interface{})
		if ops.detail {
			detail = getRequestDetail(utils.Struct2Json(r), utils.Struct2Json(rp))
		}

		detail[constant.MiddlewareAccessLogIpLogKey] = addr

		l := log.
			WithContext(ctx).
			WithFields(detail)
		if err != nil {
			l.Error(
				"%s %s %s %s %v",
				fullMethod,
				execTime,
				addr,
				code,
				err,
			)
		} else {
			l.Info(
				"%s %s %s %s",
				fullMethod,
				execTime,
				addr,
				code,
			)
		}
		return rp, err
	}
}

func getRequestDetail(d1, d2 string) (rp map[string]interface{}) {
	rp = make(map[string]interface{})
	rp[constant.MiddlewareParamsBodyLogKey] = trim(d1)
	rp[constant.MiddlewareParamsRespLogKey] = trim(d2)
	return
}

func trim(s string) string {
	s = compact(s)
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	if len(s) > 500 {
		s = fmt.Sprintf("%s......omitted......%s", s[0:250], s[len(s)-250:len(s)])
	}
	return s
}

func compact(s string) string {
	got := new(bytes.Buffer)
	if err := json.Compact(got, []byte(s)); err != nil {
		return s
	}
	return got.String()
}
