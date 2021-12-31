package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (rd Redis) FindDict(r *req.Dict) []ms.SysDict {
	list := make([]ms.SysDict, 0)
	q := rd.
		Table("sys_dict").
		Preload("DictDatas").
		Order("created_at DESC")
	name := strings.TrimSpace(r.Name)
	if name != "" {
		q.Where("name", "contains", name)
	}
	desc := strings.TrimSpace(r.Desc)
	if desc != "" {
		q.Where("desc", "=", desc)
	}
	if r.Status != nil {
		q.Where("status", "=", *r.Status)
	}
	rd.FindWithPage(q, &r.Page, &list)
	return list
}

func (rd Redis) FindDictData(r *req.DictData) []ms.SysDictData {
	list := make([]ms.SysDictData, 0)
	q := rd.
		Table("sys_dict_data").
		Preload("Dict").
		Order("created_at DESC")
	key := strings.TrimSpace(r.Key)
	if key != "" {
		q.Where("key", "contains", key)
	}
	val := strings.TrimSpace(r.Val)
	if val != "" {
		q.Where("val", "contains", val)
	}
	if r.Status != nil {
		q.Where("status", "=", *r.Status)
	}
	if r.DictId != nil {
		q.Where("dict_id", "=", *r.DictId)
	}
	rd.FindWithPage(q, &r.Page, &list)
	return list
}
