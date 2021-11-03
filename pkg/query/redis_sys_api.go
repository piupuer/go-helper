package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"strings"
)

func (rd Redis) FindApi(req *req.Api) []ms.SysApi {
	list := make([]ms.SysApi, 0)
	q := rd.
		Table("sys_api").
		Order("created_at DESC")
	method := strings.TrimSpace(req.Method)
	if method != "" {
		q.Where("method", "contains", method)
	}
	path := strings.TrimSpace(req.Path)
	if path != "" {
		q.Where("path", "contains", path)
	}
	category := strings.TrimSpace(req.Category)
	if category != "" {
		q.Where("category", "contains", category)
	}
	rd.FindWithPage(q, &req.Page, &list)
	return list
}

// find all api group by api category
func (rd Redis) FindApiGroupByCategoryByRoleKeyword(currentRoleKeyword, roleKeyword string) ([]resp.ApiGroupByCategory, []uint, error) {
	tree := make([]resp.ApiGroupByCategory, 0)
	accessIds := make([]uint, 0)
	allApi := make([]ms.SysApi, 0)
	rd.
		Table("sys_api").
		Find(&allApi)
	// find all casbin by current user's role id
	currentCasbins, err := FindCasbinByRoleKeyword(rd.ops.enforcer, currentRoleKeyword)
	// find all casbin by current role id
	casbins, err := FindCasbinByRoleKeyword(rd.ops.enforcer, roleKeyword)
	if err != nil {
		return tree, accessIds, err
	}

	newApi := make([]ms.SysApi, 0)
	for _, api := range allApi {
		path := api.Path
		method := api.Method
		for _, currentCasbin := range currentCasbins {
			if path == currentCasbin.V1 && method == currentCasbin.V2 {
				newApi = append(newApi, api)
				break
			}
		}
	}

	for _, api := range newApi {
		category := api.Category
		path := api.Path
		method := api.Method
		access := false
		for _, casbin := range casbins {
			if path == casbin.V1 && method == casbin.V2 {
				access = true
				break
			}
		}
		if access {
			accessIds = append(accessIds, api.Id)
		}
		existIndex := -1
		children := make([]resp.Api, 0)
		for index, leaf := range tree {
			if leaf.Category == category {
				children = leaf.Children
				existIndex = index
				break
			}
		}
		var item resp.Api
		utils.Struct2StructByJson(api, &item)
		item.Title = fmt.Sprintf("%s %s[%s]", item.Desc, item.Path, item.Method)
		children = append(children, item)
		if existIndex != -1 {
			tree[existIndex].Children = children
		} else {
			tree = append(tree, resp.ApiGroupByCategory{
				Title:    category + " group",
				Category: category,
				Children: children,
			})
		}
	}
	return tree, accessIds, err
}
