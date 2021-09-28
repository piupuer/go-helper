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
		).Conn()
	fmt.Println(conn, err)
}
