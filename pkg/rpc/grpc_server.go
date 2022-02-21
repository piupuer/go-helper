package rpc

import (
	"crypto/tls"
	"crypto/x509"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/rpc/interceptor"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	ops GrpcServerOptions
}

func NewGrpcServer(options ...func(*GrpcServerOptions)) *grpc.Server {
	ops := getGrpcServerOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	serverOps := make([]grpc.ServerOption, 0)
	if ops.tls {
		t, err := NewGrpcServerTls(ops.tlsOps...)
		if err != nil {
			log.WithRequestId(ops.ctx).WithError(err).Warn("load tls failed")
		} else {
			serverOps = append(serverOps, grpc.Creds(t))
		}
	}
	so := make([]grpc.ServerOption, 0)
	// interceptor options
	if ops.requestId {
		so = append(so, grpc.ChainUnaryInterceptor(interceptor.RequestId))
	}
	if ops.accessLog {
		so = append(so, grpc.ChainUnaryInterceptor(interceptor.AccessLog(ops.accessLogOps...)))
	}
	if ops.tag {
		so = append(so, grpc.ChainUnaryInterceptor(grpc_ctxtags.UnaryServerInterceptor(ops.tagOps...)))
	}
	if ops.opentracing {
		so = append(so, grpc.ChainUnaryInterceptor(grpc_opentracing.UnaryServerInterceptor(ops.opentracingOps...)))
	}
	if ops.exception {
		so = append(so, grpc.ChainUnaryInterceptor(interceptor.Exception(ops.exceptionOps...)))
	}
	if ops.transaction {
		so = append(so, grpc.ChainUnaryInterceptor(interceptor.Transaction(ops.transactionOps...)))
	}
	// custom options
	if len(ops.customs) > 0 {
		so = append(so, ops.customs...)
	}
	for _, item := range so {
		serverOps = append(serverOps, item)
	}
	srv := grpc.NewServer(serverOps...)
	if ops.healthCheck {
		hs := health.NewServer()
		hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(srv, hs)
	}
	if ops.reflection {
		reflection.Register(srv)
	}
	return srv
}

func NewGrpcServerTls(options ...func(*GrpcServerTlsOptions)) (t credentials.TransportCredentials, err error) {
	ops := getGrpcServerTlsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if len(ops.serverKey) > 0 || len(ops.serverPem) > 0 || len(ops.caPem) > 0 {
		var cert tls.Certificate
		cert, err = tls.X509KeyPair(ops.serverPem, ops.serverKey)
		if err != nil {
			err = errors.Wrap(err, "load x509 key pair failed")
			return
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(ops.caPem); !ok {
			err = errors.Errorf("append certs from pem failed")
			return
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}), nil
	}
	return nil, errors.Errorf("invalid options, serverKey: %s, serverPem: %s, caPem: %s", string(ops.serverKey), string(ops.serverPem), string(ops.caPem))
}
