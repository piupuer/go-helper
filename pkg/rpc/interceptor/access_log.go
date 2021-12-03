package interceptor

import (
	"context"
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
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		if ops.detail {
			ops.logger.Info(
				ctx,
				"req: %s",
				utils.Struct2Json(req),
			)
		}

		resp, err := handler(ctx, req)

		endTime := time.Now()

		fullMethod := info.FullMethod

		addr := ""
		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}
		code := status.Code(err).String()
		// calc request exec time
		execTime := endTime.Sub(startTime).String()
		if err != nil {
			ops.logger.Error(
				ctx,
				"%s %d %s %s %v",
				fullMethod,
				execTime,
				addr,
				code,
				err,
			)
		} else {
			if ops.detail {
				ops.logger.Info(
					ctx,
					"resp: %s",
					utils.Struct2Json(resp),
				)
			}
			ops.logger.Info(
				ctx,
				"%s %d %s %s",
				fullMethod,
				execTime,
				addr,
				code,
			)
		}
		return resp, err
	}
}
