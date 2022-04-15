package listen

import (
	"context"
	"github.com/piupuer/go-helper/pkg/rpc"
	"github.com/piupuer/go-helper/pkg/utils"
	"google.golang.org/grpc"
	"net/http"
)

type HttpOptions struct {
	ctx       context.Context
	host      string
	port      int
	pprofPort int
	urlPrefix string
	proName   string
	handler   http.Handler
	exit      func()
}

func WithHttpCtx(ctx context.Context) func(*HttpOptions) {
	return func(options *HttpOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getHttpOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithHttpHost(s string) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).host = s
	}
}

func WithHttpPort(i int) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).port = i
	}
}

func WithHttpPprofPort(i int) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).pprofPort = i
	}
}

func WithHttpUrlPrefix(s string) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).urlPrefix = s
	}
}

func WithHttpProName(s string) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).proName = s
	}
}

func WithHttpHandler(h http.Handler) func(*HttpOptions) {
	return func(options *HttpOptions) {
		getHttpOptionsOrSetDefault(options).handler = h
	}
}

func WithHttpExit(f func()) func(*HttpOptions) {
	return func(options *HttpOptions) {
		if f != nil {
			getHttpOptionsOrSetDefault(options).exit = f
		}
	}
}

func getHttpOptionsOrSetDefault(options *HttpOptions) *HttpOptions {
	if options == nil {
		return &HttpOptions{
			ctx:       context.Background(),
			host:      "0.0.0.0",
			port:      8080,
			urlPrefix: "api",
			proName:   "project",
		}
	}
	return options
}

type GrpcOptions struct {
	ctx       context.Context
	host      string
	port      int
	proName   string
	serverOps []func(*rpc.GrpcServerOptions)
	register  func(g *grpc.Server)
	exit      func()
}

func WithGrpcCtx(ctx context.Context) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getGrpcOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithGrpcHost(s string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).host = s
	}
}

func WithGrpcPort(i int) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).port = i
	}
}

func WithGrpcProName(s string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).proName = s
	}
}

func WithGrpcServerOps(o ...func(*rpc.GrpcServerOptions)) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).serverOps = o
	}
}

func WithGrpcRegister(fun func(g *grpc.Server)) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if fun != nil {
			getGrpcOptionsOrSetDefault(options).register = fun
		}
	}
}

func WithGrpcExit(f func()) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if f != nil {
			getGrpcOptionsOrSetDefault(options).exit = f
		}
	}
}

func getGrpcOptionsOrSetDefault(options *GrpcOptions) *GrpcOptions {
	if options == nil {
		return &GrpcOptions{
			ctx:     context.Background(),
			host:    "0.0.0.0",
			port:    9090,
			proName: "project",
		}
	}
	return options
}
