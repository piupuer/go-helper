package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"io/ioutil"
	"time"
)

type Grpc struct {
	uri      string
	insecure bool
	ops      GrpcOptions
}

func NewGrpc(uri string, options ...func(*GrpcOptions)) *Grpc {
	var gr Grpc
	gr.uri = uri
	ops := getGrpcOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.ClientKey == "" || ops.ClientPem == "" || ops.CaPem == "" {
		gr.insecure = true
	}
	gr.ops = *ops
	return &gr
}

func (gr *Grpc) Conn() (*grpc.ClientConn, error) {
	var ctl credentials.TransportCredentials
	if !gr.insecure {
		cert, err := tls.LoadX509KeyPair(gr.ops.ClientPem, gr.ops.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("[grpc]load x509 key pair err: %v", err)
		}
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(gr.ops.CaPem)
		if err != nil {
			return nil, fmt.Errorf("[grpc]read ca.pem err: %v", err)
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("[grpc]append certs from pem err: %v", err)
		}
		ctl = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   gr.ops.ServerName,
			RootCAs:      certPool,
		})
	}
	var option grpc.DialOption
	if ctl != nil {
		option = grpc.WithTransportCredentials(ctl)
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
