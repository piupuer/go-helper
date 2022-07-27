package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/tracing"
	"strings"
)

func (my MySql) GetDictData(dictName, dictDataKey string) (rp ms.SysDictData) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GetDictData"))
	defer span.End()
	rp = my.CacheGetDictData(my.Ctx, dictName, dictDataKey)
	if rp.Id > constant.Zero {
		return
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("created_at DESC").
		Where("status = ?", constant.One).
		Find(&list)
	for _, data := range list {
		if data.Dict.Name == dictName && data.Key == dictDataKey {
			my.CacheSetDictDataItem(my.Ctx, dictName, dictDataKey, data)
			rp = data
			return
		}
	}
	return
}

func (my MySql) FindDictDataByName(name string) (rp []ms.SysDictData) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictDataByName"))
	defer span.End()
	rp = my.CacheFindDictData(my.Ctx, name)
	if len(rp) > 0 {
		return
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("sort").
		Find(&list)
	rp = make([]ms.SysDictData, 0)
	for _, data := range list {
		if data.Dict.Name == name {
			rp = append(rp, data)
		}
	}
	if len(rp) > 0 {
		my.CacheSetDictData(my.Ctx, name, rp)
	}
	return
}

func (my MySql) FindDictDataValByName(name string) (rp []string) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictDataValByName"))
	defer span.End()
	rp = my.CacheDictDataVal(my.Ctx, name)
	if len(rp) > 0 {
		return
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Find(&list)
	rp = make([]string, 0)
	for _, data := range list {
		if data.Dict.Name == name {
			rp = append(rp, data.Val)
		}
	}
	if len(rp) > 0 {
		my.CacheSetDictDataVal(my.Ctx, name, rp)
	}
	return
}

func (my MySql) FindDict(r *req.Dict) (rp []ms.SysDict) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDict"))
	defer span.End()
	rp = make([]ms.SysDict, 0)
	q := my.Tx.
		Model(&ms.SysDict{}).
		Preload("DictDatas").
		Order("created_at DESC")
	name := strings.TrimSpace(r.Name)
	if name != "" {
		q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	desc := strings.TrimSpace(r.Desc)
	if desc != "" {
		q.Where("desc = ?", desc)
	}
	if r.Status != nil {
		q.Where("status = ?", *r.Status)
	}
	my.FindWithPage(q, &r.Page, &rp)
	return
}

func (my MySql) FindDictData(r *req.DictData) (rp []ms.SysDictData) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictData"))
	defer span.End()
	rp = make([]ms.SysDictData, 0)
	q := my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("created_at DESC")
	key := strings.TrimSpace(r.Key)
	if key != "" {
		q.Where("key LIKE ?", fmt.Sprintf("%%%s%%", key))
	}
	val := strings.TrimSpace(r.Val)
	if val != "" {
		q.Where("val LIKE ?", fmt.Sprintf("%%%s%%", val))
	}
	if r.Status != nil {
		q.Where("status = ?", *r.Status)
	}
	if r.DictId != nil {
		q.Where("dict_id = ?", *r.DictId)
	}
	my.FindWithPage(q, &r.Page, &rp)
	return
}

func (my MySql) CreateDict(r *req.CreateDict) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateDict"))
	defer span.End()
	err = my.Create(r, new(ms.SysDict))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) UpdateDictById(id uint, r req.UpdateDict) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateDictById"))
	defer span.End()
	err = my.UpdateById(id, r, new(ms.SysDict))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) DeleteDictByIds(ids []uint) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteDictByIds"))
	defer span.End()
	err = my.DeleteByIds(ids, new(ms.SysDict))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) CreateDictData(r *req.CreateDictData) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateDictData"))
	defer span.End()
	err = my.Create(r, new(ms.SysDictData))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) UpdateDictDataById(id uint, r req.UpdateDictData) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateDictDataById"))
	defer span.End()
	err = my.UpdateById(id, r, new(ms.SysDictData))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) DeleteDictDataByIds(ids []uint) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteDictDataByIds"))
	defer span.End()
	err = my.DeleteByIds(ids, new(ms.SysDictData))
	my.CacheFlushDictData(my.Ctx)
	my.CacheFlushDictDataVal(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}
