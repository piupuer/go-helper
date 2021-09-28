package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

type Grpc struct {
	uri   string
	ctl   credentials.TransportCredentials
	ops   GrpcOptions
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
	if len(ops.ClientKey) > 0 || len(ops.ClientPem) > 0 || len(ops.CaPem) > 0 {
		cert, err := tls.X509KeyPair(gr.ops.ClientPem, gr.ops.ClientKey)
		if err != nil {
			gr.Error = fmt.Errorf("[grpc]load x509 key pair err: %v", err)
			return &gr
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(gr.ops.CaPem); !ok {
			gr.Error = fmt.Errorf("[grpc]append certs from pem err: %v", err)
			return &gr
		}
		gr.ctl = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   gr.ops.ServerName,
			RootCAs:      certPool,
		})
	}
	return &gr
}

func (gr Grpc) Conn() (*grpc.ClientConn, error) {
	var option grpc.DialOption
	if gr.ctl != nil {
		option = grpc.WithTransportCredentials(gr.ctl)
	} else {
		option = grpc.WithInsecure()
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(gr.ops.Timeout)*time.Second)
	conn, err := grpc.DialContext(
		ctx,
		gr.uri,
		option,
	)
	if err != nil {
		return nil, fmt.Errorf("[grpc]dial %s err: %v", gr.uri, err)
	}
	if gr.ops.HealthCheck {
		// health check
		client := grpc_health_v1.NewHealthClient(conn)
		h, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			return nil, fmt.Errorf("[grpc]health check %s err: %v", gr.uri, err)
		}
		if h.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return nil, fmt.Errorf("[grpc]health check %s not SERVING: %v", gr.uri, h.Status)
		}
	}
	return conn, nil
}
