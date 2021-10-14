package rpc

import (
	"fmt"
	"testing"
)

func TestNewGrpc(t *testing.T) {
	conn, err :=
		NewGrpc(
			"127.0.0.1:9004",
			WithGrpcTimeout(20),
			WithGrpcHealthCheck,
			WithGrpcServerName("grpc.com"),
			WithGrpcCaPemFile("ca.pem"),
			WithGrpcClientKeyFile("client.key"),
			WithGrpcClientPemFile("client.pem"),
		).Conn()
	fmt.Println(conn, err)
}
