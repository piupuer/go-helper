package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

func AccessLog(options ...func(*AccessLogOptions)) grpc.UnaryServerInterceptor {
	ops := getAccessLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		logger.WithRequestId(ctx).Info(
			ctx,
			"req: %s",
			utils.Struct2Json(r),
		)

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
		if err != nil {
			logger.WithRequestId(ctx).Error(
				"%s %s %s RPC code: '%s', RPC err: '%v'",
				fullMethod,
				execTime,
				addr,
				code,
				err,
			)
		} else {
			if ops.detail {
				logger.WithRequestId(ctx).Info(
					"RPC code: '%s', resp: %s",
					code,
					utils.Struct2Json(rp),
				)
			} else {
				logger.WithRequestId(ctx).Info(
					"RPC code: '%s'",
					code,
				)
			}
		}
		return rp, err
	}
}
