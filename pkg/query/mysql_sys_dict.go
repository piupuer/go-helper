package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

func (my MySql) GetDictData(dictName, dictDataKey string) ms.SysDictData {
	dict, err := my.GetDictDataWithErr(dictName, dictDataKey)
	if err != nil || dict == nil {
		return ms.SysDictData{}
	}
	return *dict
}

func (my MySql) GetDictDataWithErr(dictName, dictDataKey string) (*ms.SysDictData, error) {
	oldCache, ok := my.CacheGetDictNameAndKey(my.Ctx, dictName, dictDataKey)
	if ok {
		return oldCache, nil
	}
	list := make([]ms.SysDictData, 0)
	my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("created_at DESC").
		Find(&list)
	for _, data := range list {
		if data.Dict.Name == dictName && data.Key == dictDataKey {
			my.CacheSetDictNameAndKey(my.Ctx, dictName, dictDataKey, data)
			return &data, nil
		}
	}
	return nil, errors.WithStack(gorm.ErrRecordNotFound)
}

func (my MySql) FindDictDataByName(name string) ([]ms.SysDictData, error) {
	oldCache, ok := my.CacheGetDictName(my.Ctx, name)
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
	my.CacheSetDictName(my.Ctx, name, newList)
	return newList, nil
}

func (my MySql) FindDict(r *req.Dict) []ms.SysDict {
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
	attr := strings.TrimSpace(r.Attr)
	if attr != "" {
		q.Where("attr = ?", attr)
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
	err = my.Create(r, new(ms.SysDict))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) UpdateDictById(id uint, r req.UpdateDict) (err error) {
	err = my.UpdateById(id, r, new(ms.SysDict))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) DeleteDictByIds(ids []uint) (err error) {
	err = my.DeleteByIds(ids, new(ms.SysDict))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) CreateDictData(r *req.CreateDictData) (err error) {
	err = my.Create(r, new(ms.SysDictData))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) UpdateDictDataById(id uint, r req.UpdateDictData) (err error) {
	err = my.UpdateById(id, r, new(ms.SysDictData))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) DeleteDictDataByIds(ids []uint) (err error) {
	err = my.DeleteByIds(ids, new(ms.SysDictData))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}
