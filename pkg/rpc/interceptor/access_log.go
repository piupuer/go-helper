package interceptor

import (
	"context"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"time"
)

type errResp struct {
	Error errCode `json:"error"`
}

type errCode struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

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

		rp, err := handler(ctx, req)

		endTime := time.Now()
		// calc request exec time
		execTime := endTime.Sub(startTime).String()

		fullMethod := info.FullMethod
		addr := ""
		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}
		code := status.Code(err).String()
		var e errResp
		utils.Struct2StructByJson(rp, &e)
		if err != nil {
			ops.logger.Error(
				ctx,
				"%s %s %s RPC code: '%s', RPC err: '%v'",
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
					utils.Struct2Json(rp),
				)
			}
			if e.Error.Code == resp.Ok {
				ops.logger.Info(
					ctx,
					"%s %s %s RPC code: '%s', APP code: '%d'",
					fullMethod,
					execTime,
					addr,
					code,
					e.Error.Code,
				)
			} else {
				ops.logger.Error(
					ctx,
					"%s %s %s RPC code: '%s', APP code: '%d', APP err: '%s'",
					fullMethod,
					execTime,
					addr,
					code,
					e.Error.Code,
					e.Error.Msg,
				)
			}
		}
		return rp, err
	}
}
