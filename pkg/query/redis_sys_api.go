package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"strings"
)

func (rd Redis) FindApi(r *req.Api) []ms.SysApi {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "FindApi"))
	defer span.End()
	list := make([]ms.SysApi, 0)
	q := rd.
		Table("sys_api").
		Order("created_at DESC")
	method := strings.TrimSpace(r.Method)
	if method != "" {
		q.Where("method", "contains", method)
	}
	path := strings.TrimSpace(r.Path)
	if path != "" {
		q.Where("path", "contains", path)
	}
	category := strings.TrimSpace(r.Category)
	if category != "" {
		q.Where("category", "contains", category)
	}
	rd.FindWithPage(q, &r.Page, &list)
	return list
}

func (rd Redis) FindApiGroupByCategoryByRoleKeyword(currentRoleKeyword, roleKeyword string) (tree []resp.ApiGroupByCategory, accessIds []uint) {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "FindApiGroupByCategoryByRoleKeyword"))
	defer span.End()
	tree = make([]resp.ApiGroupByCategory, 0)
	accessIds = make([]uint, 0)
	if rd.ops.enforcer == nil {
		log.WithContext(rd.Ctx).Warn("casbin enforcer is empty")
		return
	}
	allApi := make([]ms.SysApi, 0)
	rd.
		Table("sys_api").
		Find(&allApi)
	// find all casbin by current user's role id
	currentCasbins := FindCasbinByRoleKeyword(rd.ops.enforcer, currentRoleKeyword)
	// find all casbin by current role id
	casbins := FindCasbinByRoleKeyword(rd.ops.enforcer, roleKeyword)

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
	return
}
