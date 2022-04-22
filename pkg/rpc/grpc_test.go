package rpc

import (
	"fmt"
	"testing"
)

func TestNewGrpc(t *testing.T) {
	gr :=
		NewGrpc(
			"127.0.0.1:9004",
			WithGrpcTimeout(20),
		)
	fmt.Println(gr.Conn, gr.Error)
}
