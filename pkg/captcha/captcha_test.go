package captcha

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"testing"
	"time"
)

var rd redis.UniversalClient

func init() {
	rd, _ = parseRedisURI("redis://127.0.0.1:6379/0")
}

func parseRedisURI(uri string) (redis.UniversalClient, error) {
	var opt asynq.RedisConnOpt
	var err error
	if uri != "" {
		opt, err = asynq.ParseRedisURI(uri)
		if err != nil {
			return nil, err
		}
		return opt.MakeRedisClient().(redis.UniversalClient), nil
	}
	return nil, errors.Errorf("invalid redis config")
}

func TestCaptcha_Verify(t *testing.T) {
	id, img := New(
		WithRedis(rd),
	).Get()
	fmt.Println(id, img)
	time.Sleep(2000)
	fmt.Println(New(
		WithRedis(rd),
	).Verify(id, "1234"))

}
