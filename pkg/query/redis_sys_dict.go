package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (rd Redis) FindDict(req *req.Dict) []ms.SysDict {
	list := make([]ms.SysDict, 0)
	q := rd.
		Table("sys_dict").
		Preload("DictDatas").
		Order("created_at DESC")
	name := strings.TrimSpace(req.Name)
	if name != "" {
		q.Where("name", "contains", name)
	}
	desc := strings.TrimSpace(req.Desc)
	if desc != "" {
		q.Where("desc", "=", desc)
	}
	if req.Status != nil {
		q.Where("status", "=", *req.Status)
	}
	rd.FindWithPage(q, &req.Page, &list)
	return list
}

func (rd Redis) FindDictData(req *req.DictData) []ms.SysDictData {
	list := make([]ms.SysDictData, 0)
	q := rd.
		Table("sys_dict_data").
		Preload("Dict").
		Order("created_at DESC")
	key := strings.TrimSpace(req.Key)
	if key != "" {
		q.Where("key", "contains", key)
	}
	val := strings.TrimSpace(req.Val)
	if val != "" {
		q.Where("val", "contains", val)
	}
	attr := strings.TrimSpace(req.Attr)
	if attr != "" {
		q.Where("attr", "=", attr)
	}
	if req.Status != nil {
		q.Where("status", "=", *req.Status)
	}
	if req.DictId != nil {
		q.Where("dict_id", "=", *req.DictId)
	}
	rd.FindWithPage(q, &req.Page, &list)
	return list
}
