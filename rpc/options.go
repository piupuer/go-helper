package rpc

import "io/ioutil"

type GrpcOptions struct {
	ServerName  string
	CaPem       []byte
	ClientPem   []byte
	ClientKey   []byte
	Timeout     int
	HealthCheck bool
}

func WithGrpcServerName(name string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).ServerName = name
	}
}

func WithGrpcCaPem(caPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).CaPem = caPem
	}
}

func WithGrpcClientPem(clientPem []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).ClientPem = clientPem
	}
}

func WithGrpcClientKey(clientKey []byte) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		getGrpcOptionsOrSetDefault(options).ClientKey = clientKey
	}
}

func WithGrpcCaPemFile(caPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(caPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).CaPem = bs
		}
	}
}

func WithGrpcClientPemFile(clientPem string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientPem)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).ClientPem = bs
		}
	}
}

func WithGrpcClientKeyFile(clientKey string) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		bs, err := ioutil.ReadFile(clientKey)
		if err == nil {
			getGrpcOptionsOrSetDefault(options).ClientKey = bs
		}
	}
}

func WithGrpcTimeout(second int) func(*GrpcOptions) {
	return func(options *GrpcOptions) {
		if second > 0 {
			getGrpcOptionsOrSetDefault(options).Timeout = second
		}
	}
}

func WithGrpcHealthCheck(options *GrpcOptions) {
	getGrpcOptionsOrSetDefault(options).HealthCheck = true
}

func getGrpcOptionsOrSetDefault(options *GrpcOptions) *GrpcOptions {
	if options == nil {
		return &GrpcOptions{
			Timeout:     10,
			HealthCheck: false,
		}
	}
	return options
}
