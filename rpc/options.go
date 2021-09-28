package rpc

type GrpcOptions struct {
	ServerName  string
	CaPem       string
	ClientPem   string
	ClientKey   string
	Timeout     int
	HealthCheck bool
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
