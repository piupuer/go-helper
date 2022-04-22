package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type Grpc struct {
	ops   GrpcOptions
	Conn  *grpc.ClientConn
	Error error
}

func NewGrpc(uri string, options ...func(*GrpcOptions)) (gr *Grpc) {
	gr = &Grpc{}
	ops := getGrpcOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	gr.ops = *ops
	var ctl credentials.TransportCredentials
	if len(ops.clientKey) > 0 || len(ops.clientPem) > 0 || len(ops.caPem) > 0 {
		cert, err := tls.X509KeyPair(gr.ops.clientPem, gr.ops.clientKey)
		if err != nil {
			gr.Error = errors.Wrap(err, "load x509 key pair failed")
			return
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(gr.ops.caPem); !ok {
			gr.Error = errors.Errorf("append certs from pem failed")
			return
		}
		ctl = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   gr.ops.serverName,
			RootCAs:      certPool,
		})
	} else {
		ctl = insecure.NewCredentials()
	}

	streamInterceptors := make([]grpc.StreamClientInterceptor, 0)
	unaryInterceptors := make([]grpc.UnaryClientInterceptor, 0)
	// retry
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(250 * time.Millisecond)),
		grpc_retry.WithMax(5),
		grpc_retry.WithCodes(codes.Aborted, codes.NotFound, codes.Unavailable),
	}
	streamInterceptors = append(streamInterceptors, grpc_retry.StreamClientInterceptor(retryOpts...))
	unaryInterceptors = append(unaryInterceptors, grpc_retry.UnaryClientInterceptor(retryOpts...))

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(ctl),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
	}

	if len(ops.customs) > 0 {
		opts = append(opts, ops.customs...)
	}

	ctx, _ := context.WithTimeout(gr.ops.ctx, time.Duration(gr.ops.timeout)*time.Second)
	gr.Conn, gr.Error = grpc.DialContext(ctx, uri, opts...)
	return
}
