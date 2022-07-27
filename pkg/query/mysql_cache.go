package query

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	CacheSuffixDictData     = "dict_data_list"
	CacheSuffixDictDataVal  = "dict_data_val_list"
	CacheSuffixDictDataItem = "dict_data_item"
)

func (my MySql) CacheFindDictData(ctx context.Context, name string) (list []ms.SysDictData) {
	list = make([]ms.SysDictData, 0)
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			utils.Json2Struct(res, &list)
			return
		}
	}
	return
}

func (my MySql) CacheSetDictData(ctx context.Context, name string, data []ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(ctx, getDictDataCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictData(ctx context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(ctx, getDictDataCacheKey(my.ops))
	}
}

func (my MySql) CacheDictDataVal(ctx context.Context, name string) (list []string) {
	list = make([]string, 0)
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataValCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			utils.Json2Struct(res, &list)
			return
		}
	}
	return
}

func (my MySql) CacheSetDictDataVal(ctx context.Context, name string, data []string) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(ctx, getDictDataValCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataVal(ctx context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(ctx, getDictDataValCacheKey(my.ops))
	}
}

func (my MySql) CacheGetDictData(ctx context.Context, name, key string) (item ms.SysDictData) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataItemCacheKey(my.ops), fmt.Sprintf("%s_%s", name, key)).Result()
		if err == nil && res != "" {
			utils.Json2Struct(res, &item)
			return
		}
	}
	return
}

func (my MySql) CacheSetDictDataItem(ctx context.Context, name, key string, data ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(ctx, getDictDataItemCacheKey(my.ops), fmt.Sprintf("%s_%s", name, key), utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataItem(ctx context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(ctx, getDictDataItemCacheKey(my.ops))
	}
}

func getDictDataCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictData)
}

func getDictDataValCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictDataVal)
}

func getDictDataItemCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictDataItem)
}
