package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	uuid "github.com/satori/go.uuid"
	"time"
)

const prefix = "job-cron:"

type RedisClientDriver struct {
	client  redis.UniversalClient
	timeout time.Duration
	Key     string
	ops     DriverOptions
}

func NewDriver(client redis.UniversalClient, options ...func(*DriverOptions)) (*RedisClientDriver, error) {
	ops := getDriverOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &RedisClientDriver{
		client: client,
		ops:    *ops,
	}, nil
}

func (rd *RedisClientDriver) Ping() error {
	if _, err := rd.do("SET", "ping", "pong"); err != nil {
		return err
	}
	return nil
}

func (rd *RedisClientDriver) getKeyPre(serviceName string) string {
	p := prefix
	if rd.ops.prefix != "" {
		p = rd.ops.prefix
	}
	return fmt.Sprintf("%s%s:", p, serviceName)
}

func (rd *RedisClientDriver) SetTimeout(timeout time.Duration) {
	rd.timeout = timeout
}

func (rd *RedisClientDriver) SetHeartBeat(nodeID string) {
	go rd.heartBeat(nodeID)
}

func (rd *RedisClientDriver) heartBeat(nodeID string) {
	key := nodeID
	tickers := time.NewTicker(rd.timeout / 2)
	for range tickers.C {
		keyExist, err := rd.do("EXPIRE", key, int(rd.timeout/time.Second))
		if err != nil {
			rd.ops.logger.Warn(rd.ops.ctx, "redis expire err: %v", err)
			continue
		}
		if keyExist == int64(0) {
			if err := rd.registerServiceNode(nodeID); err != nil {
				rd.ops.logger.Warn(rd.ops.ctx, "register service node err: %v", err)
			}
		}
	}
}

func (rd *RedisClientDriver) GetServiceNodeList(serviceName string) ([]string, error) {
	mathStr := fmt.Sprintf("%s*", rd.getKeyPre(serviceName))
	return rd.scan(mathStr)
}

// RegisterServiceNode  register a service node
func (rd *RedisClientDriver) RegisterServiceNode(serviceName string) (nodeID string, err error) {
	nodeID = rd.randNodeID(serviceName)
	if err := rd.registerServiceNode(nodeID); err != nil {
		return "", err
	}
	return nodeID, nil
}

func (rd *RedisClientDriver) randNodeID(serviceName string) (nodeID string) {
	return rd.getKeyPre(serviceName) + uuid.NewV4().String()
}

func (rd *RedisClientDriver) registerServiceNode(nodeID string) error {
	_, err := rd.do("SETEX", nodeID, int(rd.timeout/time.Second), nodeID)
	return err
}

func (rd *RedisClientDriver) do(command string, params ...interface{}) (interface{}, error) {
	args := make([]interface{}, 0)
	args = append(args, command)
	args = append(args, params...)
	return rd.client.Do(context.Background(), args...).Result()
}

func (rd *RedisClientDriver) scan(matchStr string) ([]string, error) {
	cursor := "0"
	ret := make([]string, 0)
	for {
		reply, err := rd.do("scan", cursor, "match", matchStr)
		if err != nil {
			return nil, err
		}
		if r, ok := reply.([]interface{}); ok && len(r) == 2 {
			cursor = r[0].(string)

			list := r[1].([]interface{})
			for _, item := range list {
				ret = append(ret, item.(string))
			}
			if cursor == "0" {
				break
			}
		} else {
			return nil, errors.New("redis scan resp struct error")
		}
	}
	return ret, nil
}
