package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/piupuer/go-helper/pkg/logger"
	"github.com/piupuer/go-helper/pkg/rpc/interceptor"
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
			ops.logger.Warn(ops.ctx, "load tls err: %v", err)
		} else {
			serverOps = append(serverOps, grpc.Creds(t))
		}
	}
	interceptors := make([]grpc.UnaryServerInterceptor, 0)
	if ops.exception {
		ops.exceptionOps = append(ops.exceptionOps, interceptor.WithExceptionLogger(ops.logger))
		interceptors = append(interceptors, interceptor.Exception(ops.exceptionOps...))
	}
	if ops.requestId {
		interceptors = append(interceptors, interceptor.RequestId(ops.requestIdOps...))
	}
	if ops.tag {
		interceptors = append(interceptors, grpc_ctxtags.UnaryServerInterceptor(ops.tagOps...))
	}
	if ops.opentracing {
		interceptors = append(interceptors, grpc_opentracing.UnaryServerInterceptor(ops.opentracingOps...))
	}
	if z, ok := ops.logger.(*logger.Logger); ok {
		interceptors = append(interceptors, grpc_zap.UnaryServerInterceptor(z.GetZapLog()))
	}
	if ops.transaction {
		interceptors = append(interceptors, interceptor.Transaction(ops.transactionOps...))
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
			err = fmt.Errorf("[grpc]load x509 key pair err: %v", err)
			return
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(ops.caPem); !ok {
			err = fmt.Errorf("[grpc]append certs from pem err: %v", err)
			return
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}), nil
	}
	return nil, fmt.Errorf("[grpc]invalid options, serverKey: %s, serverPem: %s, caPem: %s", string(ops.serverKey), string(ops.serverPem), string(ops.caPem))
}
