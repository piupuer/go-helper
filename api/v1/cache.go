package v1

import (
	"context"
	"fmt"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
)

const (
	CacheSuffix         = "v1_cache"
	CacheSuffixMenuTree = "menu_tree"
)

// get menu tree from cache by uid
func CacheGetMenuTree(c context.Context, uid uint, ops Options) ([]resp.MenuTreeResp, bool) {
	if ops.redis != nil {
		res, err := ops.redis.HGet(c, fmt.Sprintf("%s_%s", CacheSuffix, CacheSuffixMenuTree), fmt.Sprintf("%d", uid)).Result()
		if err == nil && res != "" {
			list := make([]resp.MenuTreeResp, 0)
			utils.Json2Struct(res, &list)
			return list, true
		}
	}
	return nil, false
}

// set menu tree to cache by uid
func CacheSetMenuTree(c context.Context, uid uint, data []resp.MenuTreeResp, ops Options) {
	if ops.redis != nil {
		ops.redis.HSet(c, fmt.Sprintf("%s_%s", CacheSuffix, CacheSuffixMenuTree), fmt.Sprintf("%d", uid), utils.Struct2Json(data))
	}
}

// clear menu tree cache
func CacheFlushMenuTree(c context.Context, ops Options) {
	if ops.redis != nil {
		ops.redis.Del(c, fmt.Sprintf("%s_%s", CacheSuffix, CacheSuffixMenuTree))
	}
}
