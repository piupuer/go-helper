package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

func (my MySql) GetDictData(dictName, dictDataKey string) ms.SysDictData {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GetDictData"))
	defer span.End()
	dict, err := my.GetDictDataWithErr(dictName, dictDataKey)
	if err != nil || dict == nil {
		return ms.SysDictData{}
	}
	return *dict
}

func (my MySql) GetDictDataWithErr(dictName, dictDataKey string) (*ms.SysDictData, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GetDictDataWithErr"))
	defer span.End()
	oldCache, ok := my.CacheDictDataItem(my.Ctx, dictName, dictDataKey)
	if ok {
		return oldCache, nil
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
			return &data, nil
		}
	}
	return nil, errors.WithStack(gorm.ErrRecordNotFound)
}

func (my MySql) FindDictDataByName(name string) ([]ms.SysDictData, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictDataByName"))
	defer span.End()
	oldCache, ok := my.CacheDictDataList(my.Ctx, name)
	if ok {
		return oldCache, nil
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("sort").
		Find(&list)
	newList := make([]ms.SysDictData, 0)
	for _, data := range list {
		if data.Dict.Name == name {
			newList = append(newList, data)
		}
	}
	if len(newList) > 0 {
		my.CacheSetDictDataList(my.Ctx, name, newList)
	}
	return newList, nil
}

func (my MySql) FindDictDataValByName(name string) []string {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictDataValByName"))
	defer span.End()
	oldCache, ok := my.CacheDictDataValList(my.Ctx, name)
	if ok {
		return oldCache
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Find(&list)
	newList := make([]string, 0)
	for _, data := range list {
		if data.Dict.Name == name {
			newList = append(newList, data.Val)
		}
	}
	if len(newList) > 0 {
		my.CacheSetDictDataValList(my.Ctx, name, newList)
	}
	return newList
}

func (my MySql) FindDict(r *req.Dict) []ms.SysDict {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDict"))
	defer span.End()
	list := make([]ms.SysDict, 0)
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
	my.FindWithPage(q, &r.Page, &list)
	return list
}

func (my MySql) FindDictData(r *req.DictData) []ms.SysDictData {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindDictData"))
	defer span.End()
	list := make([]ms.SysDictData, 0)
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
	my.FindWithPage(q, &r.Page, &list)
	return list
}

func (my MySql) CreateDict(r *req.CreateDict) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateDict"))
	defer span.End()
	err = my.Create(r, new(ms.SysDict))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) UpdateDictById(id uint, r req.UpdateDict) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateDictById"))
	defer span.End()
	err = my.UpdateById(id, r, new(ms.SysDict))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) DeleteDictByIds(ids []uint) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteDictByIds"))
	defer span.End()
	err = my.DeleteByIds(ids, new(ms.SysDict))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) CreateDictData(r *req.CreateDictData) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateDictData"))
	defer span.End()
	err = my.Create(r, new(ms.SysDictData))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) UpdateDictDataById(id uint, r req.UpdateDictData) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateDictDataById"))
	defer span.End()
	err = my.UpdateById(id, r, new(ms.SysDictData))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}

func (my MySql) DeleteDictDataByIds(ids []uint) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteDictDataByIds"))
	defer span.End()
	err = my.DeleteByIds(ids, new(ms.SysDictData))
	my.CacheFlushDictDataList(my.Ctx)
	my.CacheFlushDictDataValList(my.Ctx)
	my.CacheFlushDictDataItem(my.Ctx)
	return
}
