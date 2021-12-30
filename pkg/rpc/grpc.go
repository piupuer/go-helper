package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
	} else {
		gr.ctl = insecure.NewCredentials()
	}
	if ops.healthCheck {
		err := gr.HealthCheck(WithGrpcHealthCheckCtx(ops.ctx))
		if err != nil {
			gr.Error = err
		}
	}
	return &gr
}

func (gr Grpc) Conn() (conn *grpc.ClientConn, err error) {
	if gr.Error != nil {
		err = gr.Error
		return
	}
	var option grpc.DialOption
	option = grpc.WithTransportCredentials(gr.ctl)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(gr.ops.timeout)*time.Second)
	conn, err = grpc.DialContext(
		ctx,
		gr.uri,
		option,
	)
	return
}

func (gr Grpc) HealthCheck(options ...func(*GrpcHealthCheckOptions)) (err error) {
	if gr.Error != nil {
		err = gr.Error
		return
	}
	ops := getGrpcHealthCheckOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// get connection
	var conn *grpc.ClientConn
	conn, err = gr.Conn()
	if err != nil {
		return
	}
	defer conn.Close()
	// health check
	client := grpc_health_v1.NewHealthClient(conn)
	var h *grpc_health_v1.HealthCheckResponse
	ctx, _ := context.WithTimeout(ops.ctx, time.Duration(gr.ops.timeout)*time.Second)
	h, err = client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		gr.ops.logger.Warn(ctx, "health check %s failed: %v", gr.uri, err)
		return errors.Wrapf(err, "health check %s failed", gr.uri)
	}
	if h.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		gr.ops.logger.Warn(ctx, "health check %s not SERVING: %v", gr.uri, h.Status)
		return errors.Errorf("health check %s not SERVING", gr.uri)
	}
	return
}
