package v1

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	CacheSuffixMenuTree = "menu_tree"
)

// get menu tree from cache by uid
func CacheGetMenuTree(ctx context.Context, uid uint, ops Options) ([]resp.MenuTree, bool) {
	if ops.redis != nil {
		res, err := ops.redis.HGet(ctx, getMenuTreeCacheKeyPrefix(ops), fmt.Sprintf("%d", uid)).Result()
		if err == nil && res != "" {
			list := make([]resp.MenuTree, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

// set menu tree to cache by uid
func CacheSetMenuTree(ctx context.Context, uid uint, data []resp.MenuTree, ops Options) {
	if ops.redis != nil {
		ops.redis.HSet(ctx, getMenuTreeCacheKeyPrefix(ops), fmt.Sprintf("%d", uid), utils.Struct2Json(data))
	}
}

// clear menu tree cache
func CacheFlushMenuTree(ctx context.Context, ops Options) {
	if ops.redis != nil {
		ops.redis.Del(ctx, getMenuTreeCacheKeyPrefix(ops))
	}
}

func getMenuTreeCacheKeyPrefix(ops Options) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixMenuTree)
}
