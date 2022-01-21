package query

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	CacheSuffixDictDataList    = "dict_data_list"
	CacheSuffixDictDataValList = "dict_data_val_list"
	CacheSuffixDictDataItem    = "dict_data_item"
)

func (my MySql) CacheDictDataList(ctx context.Context, name string) ([]ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataListCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			list := make([]ms.SysDictData, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

func (my MySql) CacheSetDictDataList(ctx context.Context, name string, data []ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(ctx, getDictDataListCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataList(ctx context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(ctx, getDictDataListCacheKey(my.ops))
	}
}

func (my MySql) CacheDictDataValList(ctx context.Context, name string) ([]string, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataValListCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			list := make([]string, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

func (my MySql) CacheSetDictDataValList(ctx context.Context, name string, data []string) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(ctx, getDictDataValListCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataValList(ctx context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(ctx, getDictDataValListCacheKey(my.ops))
	}
}

func (my MySql) CacheDictDataItem(ctx context.Context, name, key string) (*ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(ctx, getDictDataItemCacheKey(my.ops), fmt.Sprintf("%s_%s", name, key)).Result()
		if err == nil && res != "" {
			item := ms.SysDictData{}
			utils.Json2Struct(res, &item)
			return &item, true
		}
	}
	return nil, false
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

func getDictDataListCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictDataList)
}

func getDictDataValListCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictDataValList)
}

func getDictDataItemCacheKey(ops MysqlOptions) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixDictDataItem)
}
