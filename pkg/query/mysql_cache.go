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

func (my MySql) CacheDictDataList(c context.Context, name string) ([]ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(c, getDictDataListCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			list := make([]ms.SysDictData, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

func (my MySql) CacheSetDictDataList(c context.Context, name string, data []ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(c, getDictDataListCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataList(c context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(c, getDictDataListCacheKey(my.ops))
	}
}

func (my MySql) CacheDictDataValList(c context.Context, name string) ([]string, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(c, getDictDataValListCacheKey(my.ops), name).Result()
		if err == nil && res != "" {
			list := make([]string, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

func (my MySql) CacheSetDictDataValList(c context.Context, name string, data []string) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(c, getDictDataValListCacheKey(my.ops), name, utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataValList(c context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(c, getDictDataValListCacheKey(my.ops))
	}
}

func (my MySql) CacheDictDataItem(c context.Context, name, key string) (*ms.SysDictData, bool) {
	if my.ops.redis != nil {
		res, err := my.ops.redis.HGet(c, getDictDataItemCacheKey(my.ops), fmt.Sprintf("%s_%s", name, key)).Result()
		if err == nil && res != "" {
			item := ms.SysDictData{}
			utils.Json2Struct(res, &item)
			return &item, true
		}
	}
	return nil, false
}

func (my MySql) CacheSetDictDataItem(c context.Context, name, key string, data ms.SysDictData) {
	if my.ops.redis != nil {
		my.ops.redis.HSet(c, getDictDataItemCacheKey(my.ops), fmt.Sprintf("%s_%s", name, key), utils.Struct2Json(data))
	}
}

func (my MySql) CacheFlushDictDataItem(c context.Context) {
	if my.ops.redis != nil {
		my.ops.redis.Del(c, getDictDataItemCacheKey(my.ops))
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
