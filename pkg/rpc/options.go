package rpc

import (
	"context"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/rpc/interceptor"
	"github.com/piupuer/go-helper/pkg/utils"
	"io/ioutil"
)

type GrpcOptions struct {
	serverName  string
	caPem       []byte
	clientPem   []byte
	clientKey   []byte
	timeout     int
	healthCheck bool
}

func WithGrpcServerName(name string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).serverName = name
	}
}

func WithGrpcCaPem(caPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).caPem = caPem
	}
}

func WithGrpcClientPem(clientPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).clientPem = clientPem
	}
}

func WithGrpcClientKey(clientKey []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).clientKey = clientKey
	}
}

func WithGrpcCaPemFile(caPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(caPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).caPem = bs
		}
	}
}

func WithGrpcClientPemFile(clientPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).clientPem = bs
		}
	}
}

func WithGrpcClientKeyFile(clientKey string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientKey)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).clientKey = bs
		}
	}
}

func WithGrpcTimeout(second int) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if second > 0 {
			getGrpcOptionsOrSetDefault(options).timeout = second
		}
	}
}

func WithGrpcHealthCheck(flag bool) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).healthCheck = flag
	}
}

func getGrpcOptionsOrSetDefault(options *GrpcOptions) *GrpcOptions {
	if options == nil {
		return &GrpcOptions{
			timeout: constant.GrpcTimeout,
		}
	}
	return options
}

type GrpcServerOptions struct {
	logger         logger.Interface
	ctx            context.Context
	tls            bool
	tlsOps         []func(*GrpcServerTlsOptions)
	exception      bool
	exceptionOps   []func(*interceptor.ExceptionOptions)
	requestId      bool
	requestIdOps   []func(*interceptor.RequestIdOptions)
	transaction    bool
	transactionOps []func(*interceptor.TransactionOptions)
	tag            bool
	tagOps         []grpc_ctxtags.Option
	opentracing    bool
	opentracingOps []grpc_opentracing.Option
	healthCheck    bool
	reflection     bool
}

func WithGrpcServerLogger(l logger.Interface) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		if l != nil {
			getGrpcServerOptionsOrSetDefault(options).logger = l
		}
	}
}

func WithGrpcServerCtx(ctx context.Context) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getGrpcServerOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithGrpcServerTls(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).tls = flag
	}
}

func WithGrpcServerTlsOps(ops ...func(*GrpcServerTlsOptions)) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).tlsOps = append(getGrpcServerOptionsOrSetDefault(options).tlsOps, ops...)
	}
}

func WithGrpcServerException(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).exception = flag
	}
}

func WithGrpcServerExceptionOps(ops ...func(*interceptor.ExceptionOptions)) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).exceptionOps = append(getGrpcServerOptionsOrSetDefault(options).exceptionOps, ops...)
	}
}

func WithGrpcServerRequestId(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).requestId = flag
	}
}

func WithGrpcServerRequestIdOps(ops ...func(*interceptor.RequestIdOptions)) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).requestIdOps = append(getGrpcServerOptionsOrSetDefault(options).requestIdOps, ops...)
	}
}

func WithGrpcServerTransaction(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).transaction = flag
	}
}

func WithGrpcServerTransactionOps(ops ...func(*interceptor.TransactionOptions)) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).transactionOps = append(getGrpcServerOptionsOrSetDefault(options).transactionOps, ops...)
	}
}

func WithGrpcServerTag(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).tag = flag
	}
}

func WithGrpcServerTagOps(ops ...grpc_ctxtags.Option) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).tagOps = append(getGrpcServerOptionsOrSetDefault(options).tagOps, ops...)
	}
}

func WithGrpcServerOpentracing(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).opentracing = flag
	}
}

func WithGrpcServerOpentracingOps(ops ...grpc_opentracing.Option) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).opentracingOps = append(getGrpcServerOptionsOrSetDefault(options).opentracingOps, ops...)
	}
}

func WithGrpcServerHealthCheck(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).healthCheck = flag
	}
}

func WithGrpcServerReflection(flag bool) func(*GrpcServerOptions) {
	return func(options *GrpcServerOptions) {
		getGrpcServerOptionsOrSetDefault(options).reflection = flag
	}
}

func getGrpcServerOptionsOrSetDefault(options *GrpcServerOptions) *GrpcServerOptions {
	if options == nil {
		return &GrpcServerOptions{
			logger:      logger.DefaultLogger(),
			ctx:         context.Background(),
			tls:         true,
			exception:   true,
			requestId:   true,
			tag:         true,
			opentracing: true,
			transaction: true,
		}
	}
	return options
}

type GrpcServerTlsOptions struct {
	caPem     []byte
	serverPem []byte
	serverKey []byte
}

func WithGrpcServerTlsCaPem(bs []byte) func(*GrpcServerTlsOptions) {
	return func(options *GrpcServerTlsOptions) {
		getGrpcServerTlsOptionsOrSetDefault(options).caPem = bs
	}
}

func WithGrpcServerTlsServerPem(bs []byte) func(*GrpcServerTlsOptions) {
	return func(options *GrpcServerTlsOptions) {
		getGrpcServerTlsOptionsOrSetDefault(options).serverPem = bs
	}
}

func WithGrpcServerTlsServerKey(bs []byte) func(*GrpcServerTlsOptions) {
	return func(options *GrpcServerTlsOptions) {
		getGrpcServerTlsOptionsOrSetDefault(options).serverKey = bs
	}
}

func getGrpcServerTlsOptionsOrSetDefault(options *GrpcServerTlsOptions) *GrpcServerTlsOptions {
	if options == nil {
		return &GrpcServerTlsOptions{}
	}
	return options
}
