package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (rd Redis) FindDict(req *req.Dict) []ms.SysDict {
	list := make([]ms.SysDict, 0)
	query := rd.
		Table("sys_dict").
		Preload("DictDatas").
		Order("created_at DESC")
	name := strings.TrimSpace(req.Name)
	if name != "" {
		query = query.Where("name", "contains", name)
	}
	desc := strings.TrimSpace(req.Desc)
	if desc != "" {
		query = query.Where("desc", "=", desc)
	}
	if req.Status != nil {
		query = query.Where("status", "=", *req.Status)
	}
	rd.FindWithPage(query, &req.Page, &list)
	return list
}

func (rd Redis) FindDictData(req *req.DictData) []ms.SysDictData {
	list := make([]ms.SysDictData, 0)
	query := rd.
		Table("sys_dict_data").
		Preload("Dict").
		Order("created_at DESC")
	key := strings.TrimSpace(req.Key)
	if key != "" {
		query = query.Where("key", "contains", key)
	}
	val := strings.TrimSpace(req.Val)
	if val != "" {
		query = query.Where("val", "contains", val)
	}
	attr := strings.TrimSpace(req.Attr)
	if attr != "" {
		query = query.Where("attr", "=", attr)
	}
	if req.Status != nil {
		query = query.Where("status", "=", *req.Status)
	}
	if req.DictId != nil {
		query = query.Where("dict_id", "=", *req.DictId)
	}
	rd.FindWithPage(query, &req.Page, &list)
	return list
}
