package query

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	CachePrefix               = "mysql_cache"
	CacheSuffixDictName       = "dict_name"
	CacheSuffixDictNameAndKey = "dict_name_and_key"
)

// get dict name from cache by uid
func (my MySql) CacheGetDictName(c context.Context, name string) ([]ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictName), name).Result()
		if err == nil && res != "" {
			list := make([]ms.SysDictData, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

// set dict name to cache by uid
func (my MySql) CacheSetDictName(c context.Context, name string, data []ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictName), name, utils.Struct2Json(data))
	}
}

// delete dict name
func (my MySql) CacheDeleteDictName(c context.Context, name string) {
	if my.ops.redis != nil {
		my.ops.redis.HDel(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictName), name)
	}
}

// clear dict name cache
func (my MySql) CacheFlushDictName(c context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictName))
	}
}

// get dict name and key from cache by uid
func (my MySql) CacheGetDictNameAndKey(c context.Context, name, key string) (*ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictNameAndKey), fmt.Sprintf("%s_%s", name, key)).Result()
		if err == nil && res != "" {
			item := ms.SysDictData{}
			utils.Json2Struct(res, &item)
			return &item, true
		}
	}
	return nil, false
}

// set dict name and key to cache by uid
func (my MySql) CacheSetDictNameAndKey(c context.Context, name, key string, data ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictNameAndKey), fmt.Sprintf("%s_%s", name, key), utils.Struct2Json(data))
	}
}

// delete dict name and key
func (my MySql) CacheDeleteDictNameAndKey(c context.Context, name, key string) {
	if my.ops.redis != nil {
		my.ops.redis.HDel(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictNameAndKey), fmt.Sprintf("%s_%s", name, key))
	}
}

// clear dict name and key cache
func (my MySql) CacheFlushDictNameAndKey(c context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(c, fmt.Sprintf("%s_%s", CachePrefix, CacheSuffixDictNameAndKey))
	}
}
