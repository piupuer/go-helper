package rpc

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FailWithMsg(format interface{}, a ...interface{}) error {
	var f string
	switch format.(type) {
	case string:
		f = format.(string)
	case error:
		f = fmt.Sprintf("%v", format.(error))
	}
	code := codes.InvalidArgument
	if f == "" {
		code = codes.OK
	}
	return status.Errorf(code, f, a)
}
