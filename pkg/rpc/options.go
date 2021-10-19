package rpc

import (
	"github.com/piupuer/go-helper/pkg/constant"
	"io/ioutil"
)

type GrpcOptions struct {
	serverName  string
	caPem       []byte
	clientPem   []byte
	clientKey   []byte
	timeout     int
	healthCheck bool
}

func WithGrpcServerName(name string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).serverName = name
	}
}

func WithGrpcCaPem(caPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).caPem = caPem
	}
}

func WithGrpcClientPem(clientPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).clientPem = clientPem
	}
}

func WithGrpcClientKey(clientKey []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).clientKey = clientKey
	}
}

func WithGrpcCaPemFile(caPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(caPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).caPem = bs
		}
	}
}

func WithGrpcClientPemFile(clientPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).clientPem = bs
		}
	}
}

func WithGrpcClientKeyFile(clientKey string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientKey)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).clientKey = bs
		}
	}
}

func WithGrpcTimeout(second int) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if second > 0 {
			getGrpcOptionsOrSetDefault(options).timeout = second
		}
	}
}

func WithGrpcHealthCheck(options *GrpcOptions) {
	getGrpcOptionsOrSetDefault(options).healthCheck = true
}

func getGrpcOptionsOrSetDefault(options *GrpcOptions) *GrpcOptions {
	if options == nil {
		return &GrpcOptions{
			timeout: constant.GrpcTimeout,
		}
	}
	return options
}
