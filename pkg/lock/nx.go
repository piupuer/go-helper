package lock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type NxLock struct {
	Redis      redis.UniversalClient
	Key        string
	Expiration time.Duration
}

func (nl NxLock) Lock() (ok bool) {
	if nl.Redis == nil {
		return
	}
	if nl.Key == "" {
		return
	}
	if nl.Expiration == 0 {
		nl.Expiration = time.Minute
	}
	ok, _ = nl.Redis.SetNX(context.Background(), nl.Key, 1, nl.Expiration).Result()
	return
}

func (nl NxLock) Unlock() {
	if nl.Redis == nil {
		return
	}
	if nl.Key == "" {
		return
	}
	nl.Redis.Del(context.Background(), nl.Key)
	return
}
