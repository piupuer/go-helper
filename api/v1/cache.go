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

// CacheGetMenuTree get menu tree from cache by uid
func CacheGetMenuTree(ctx context.Context, uid uint, ops Options) (rp []resp.MenuTree, exists bool) {
	rp = make([]resp.MenuTree, 0)
	if ops.redis != nil {
		res, err := ops.redis.HGet(ctx, getMenuTreeCacheKeyPrefix(ops), fmt.Sprintf("%d", uid)).Result()
		if err == nil && res != "" {
			utils.Json2Struct(res, &rp)
			if len(rp) > 0 {
				exists = true
			}
			return
		}
	}
	return
}

// CacheSetMenuTree set menu tree to cache by uid
func CacheSetMenuTree(ctx context.Context, uid uint, data []resp.MenuTree, ops Options) {
	if ops.redis != nil {
		ops.redis.HSet(ctx, getMenuTreeCacheKeyPrefix(ops), fmt.Sprintf("%d", uid), utils.Struct2Json(data))
	}
}

// CacheFlushMenuTree clear menu tree cache
func CacheFlushMenuTree(ctx context.Context, ops Options) {
	if ops.redis != nil {
		ops.redis.Del(ctx, getMenuTreeCacheKeyPrefix(ops))
	}
}

func getMenuTreeCacheKeyPrefix(ops Options) string {
	return fmt.Sprintf("%s_%s", ops.cachePrefix, CacheSuffixMenuTree)
}
