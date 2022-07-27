package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

func (my MySql) FindApi(r *req.Api) []ms.SysApi {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindApi"))
	defer span.End()
	list := make([]ms.SysApi, 0)
	q := my.Tx.
		Model(&ms.SysApi{}).
		Order("created_at DESC")
	method := strings.TrimSpace(r.Method)
	if method != "" {
		q.Where("method LIKE ?", fmt.Sprintf("%%%s%%", method))
	}
	path := strings.TrimSpace(r.Path)
	if path != "" {
		q.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	category := strings.TrimSpace(r.Category)
	if category != "" {
		q.Where("category LIKE ?", fmt.Sprintf("%%%s%%", category))
	}
	my.FindWithPage(q, &r.Page, &list)
	return list
}

func (my MySql) FindApiGroupByCategoryByRoleKeyword(currentRoleKeyword, roleKeyword string) (tree []resp.ApiGroupByCategory, accessIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindApiGroupByCategoryByRoleKeyword"))
	defer span.End()
	tree = make([]resp.ApiGroupByCategory, 0)
	accessIds = make([]uint, 0)
	if my.ops.enforcer == nil {
		log.WithContext(my.Ctx).Warn("casbin enforcer is empty")
		return
	}
	allApi := make([]ms.SysApi, 0)
	// find all api
	my.Tx.Find(&allApi)
	// find all casbin by current user's role id
	currentCasbins := FindCasbinByRoleKeyword(my.ops.enforcer, currentRoleKeyword)
	// find all casbin by current role id
	casbins := FindCasbinByRoleKeyword(my.ops.enforcer, roleKeyword)

	newApi := make([]ms.SysApi, 0)
	for _, api := range allApi {
		path := api.Path
		method := api.Method
		for _, currentCasbin := range currentCasbins {
			// have permission
			if path == currentCasbin.V1 && method == currentCasbin.V2 {
				newApi = append(newApi, api)
				break
			}
		}
	}

	// group by category
	for _, api := range newApi {
		category := api.Category
		path := api.Path
		method := api.Method
		access := false
		for _, casbin := range casbins {
			// have permission
			if path == casbin.V1 && method == casbin.V2 {
				access = true
				break
			}
		}
		// add to access ids
		if access {
			accessIds = append(accessIds, api.Id)
		}
		// generate api tree
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

func (my MySql) CreateApi(r *req.CreateApi) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateApi"))
	defer span.End()
	api := new(ms.SysApi)
	err = my.Create(r, new(ms.SysApi))
	if err != nil {
		return errors.WithStack(err)
	}
	if len(r.RoleKeywords) > 0 {
		// generate casbin rules
		cs := make([]ms.SysRoleCasbin, 0)
		for _, keyword := range r.RoleKeywords {
			cs = append(cs, ms.SysRoleCasbin{
				Keyword: keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
		}
		_, err = my.BatchCreateRoleCasbin(cs)
	}
	return
}

func (my MySql) UpdateApiById(id uint, r req.UpdateApi) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateApiById"))
	defer span.End()
	var api ms.SysApi
	q := my.Tx.Model(&api).Where("id = ?", id).First(&api)
	if errors.Is(q.Error, gorm.ErrRecordNotFound) {
		return errors.WithStack(gorm.ErrRecordNotFound)
	}

	m := make(map[string]interface{}, 0)
	utils.CompareDiff2SnakeKey(api, r, &m)

	oldApi := api
	err = q.Updates(m).Error

	// get diff fields
	diff := make(map[string]interface{}, 0)
	utils.CompareDiff2SnakeKey(oldApi, api, &diff)

	path, ok1 := diff["path"]
	method, ok2 := diff["method"]
	if (ok1 && path != "") || (ok2 && method != "") {
		// path/method change, the caspin rule needs to be updated
		oldCasbins := my.FindRoleCasbin(ms.SysRoleCasbin{
			Path:   oldApi.Path,
			Method: oldApi.Method,
		})
		if len(oldCasbins) > 0 {
			keywords := make([]string, 0)
			for _, oldCasbin := range oldCasbins {
				keywords = append(keywords, oldCasbin.Keyword)
			}
			// delete old rules
			my.BatchDeleteRoleCasbin(oldCasbins)
			// create new rules
			newCasbins := make([]ms.SysRoleCasbin, 0)
			for _, keyword := range keywords {
				newCasbins = append(newCasbins, ms.SysRoleCasbin{
					Keyword: keyword,
					Path:    api.Path,
					Method:  api.Method,
				})
			}
			_, err = my.BatchCreateRoleCasbin(newCasbins)
		}
	}
	return
}

func (my MySql) UpdateApiByRoleKeyword(keyword string, r req.UpdateMenuIncrementalIds) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateApiByRoleKeyword"))
	defer span.End()
	if len(r.Delete) > 0 {
		deleteApis := make([]ms.SysApi, 0)
		my.Tx.
			Where("id IN (?)", r.Delete).
			Find(&deleteApis)
		cs := make([]ms.SysRoleCasbin, 0)
		for _, api := range deleteApis {
			cs = append(cs, ms.SysRoleCasbin{
				Keyword: keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
		}
		_, err = my.BatchDeleteRoleCasbin(cs)
	}
	if len(r.Create) > 0 {
		createApis := make([]ms.SysApi, 0)
		my.Tx.
			Where("id IN (?)", r.Create).
			Find(&createApis)
		cs := make([]ms.SysRoleCasbin, 0)
		for _, api := range createApis {
			cs = append(cs, ms.SysRoleCasbin{
				Keyword: keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
		}
		_, err = my.BatchCreateRoleCasbin(cs)

	}
	return
}

func (my MySql) DeleteApiByIds(ids []uint) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteApiByIds"))
	defer span.End()
	var list []ms.SysApi
	q := my.Tx.Where("id IN (?)", ids).Find(&list)
	casbins := make([]ms.SysRoleCasbin, 0)
	for _, api := range list {
		casbins = append(casbins, my.FindRoleCasbin(ms.SysRoleCasbin{
			Path:   api.Path,
			Method: api.Method,
		})...)
	}
	// delete old rules
	my.BatchDeleteRoleCasbin(casbins)
	return q.Delete(&ms.SysApi{}).Error
}
