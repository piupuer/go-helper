package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

type Grpc struct {
	ops   GrpcOptions
	uri   string
	ctl   credentials.TransportCredentials
	Error error
}

func NewGrpc(uri string, options ...func(*GrpcOptions)) *Grpc {
	var gr Grpc
	gr.uri = uri
	ops := getGrpcOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	gr.ops = *ops
	if len(ops.clientKey) > 0 || len(ops.clientPem) > 0 || len(ops.caPem) > 0 {
		cert, err := tls.X509KeyPair(gr.ops.clientPem, gr.ops.clientKey)
		if err != nil {
			gr.Error = errors.Wrap(err, "load x509 key pair failed")
			return &gr
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(gr.ops.caPem); !ok {
			gr.Error = errors.Errorf("append certs from pem failed")
			return &gr
		}
		gr.ctl = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   gr.ops.serverName,
			RootCAs:      certPool,
		})
	}
	return &gr
}

func (gr Grpc) Conn() (*grpc.ClientConn, error) {
	if gr.Error != nil {
		return nil, gr.Error
	}
	var option grpc.DialOption
	if gr.ctl != nil {
		option = grpc.WithTransportCredentials(gr.ctl)
	} else {
		option = grpc.WithInsecure()
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(gr.ops.timeout)*time.Second)
	conn, err := grpc.DialContext(
		ctx,
		gr.uri,
		option,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "dial %s failed", gr.uri)
	}
	if gr.ops.healthCheck {
		// health check
		client := grpc_health_v1.NewHealthClient(conn)
		h, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return nil, errors.Wrapf(err, "health check %s failed", gr.uri)
		}
		if h.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return nil, errors.Wrapf(err, "health check %s not SERVING", gr.uri)
		}
	}
	return conn, nil
}
