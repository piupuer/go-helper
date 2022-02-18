package utils

import "fmt"

type EnvOptions struct {
	obj    interface{}
	prefix string
	format func(key string, val interface{}) string
}

func WithEnvPrefix(prefix string) func(*EnvOptions) {
	return func(options *EnvOptions) {
		getOptionsOrSetDefault(options).prefix = prefix
	}
}

func WithEnvObj(i interface{}) func(*EnvOptions) {
	return func(options *EnvOptions) {
		getOptionsOrSetDefault(options).obj = i
	}
}

func WithEnvFormat(fun func(key string, val interface{}) string) func(*EnvOptions) {
	return func(options *EnvOptions) {
		if fun != nil {
			getOptionsOrSetDefault(options).format = fun
		}
	}
}

func getOptionsOrSetDefault(options *EnvOptions) *EnvOptions {
	if options == nil {
		return &EnvOptions{
			format: func(key string, val interface{}) string {
				return fmt.Sprintf("%s: %v", key, val)
			},
		}
	}
	return options
}
