package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
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
	return nil, gorm.ErrRecordNotFound
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

func (my MySql) FindDict(req *req.Dict) []ms.SysDict {
	list := make([]ms.SysDict, 0)
	query := my.Tx.
		Model(&ms.SysDict{}).
		Preload("DictDatas").
		Order("created_at DESC")
	name := strings.TrimSpace(req.Name)
	if name != "" {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	desc := strings.TrimSpace(req.Desc)
	if desc != "" {
		query = query.Where("desc = ?", desc)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	my.FindWithPage(query, &req.Page, &list)
	return list
}

func (my MySql) FindDictData(req *req.DictData) []ms.SysDictData {
	list := make([]ms.SysDictData, 0)
	query := my.Tx.
		Model(&ms.SysDictData{}).
		Preload("Dict").
		Order("created_at DESC")
	key := strings.TrimSpace(req.Key)
	if key != "" {
		query = query.Where("key LIKE ?", fmt.Sprintf("%%%s%%", key))
	}
	val := strings.TrimSpace(req.Val)
	if val != "" {
		query = query.Where("val LIKE ?", fmt.Sprintf("%%%s%%", val))
	}
	attr := strings.TrimSpace(req.Attr)
	if attr != "" {
		query = query.Where("attr = ?", attr)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.DictId != nil {
		query = query.Where("dict_id = ?", *req.DictId)
	}
	my.FindWithPage(query, &req.Page, &list)
	return list
}

func (my MySql) CreateDict(req *req.CreateDict) (err error) {
	err = my.Create(req, new(ms.SysDict))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) UpdateDictById(id uint, req req.UpdateDict) (err error) {
	err = my.UpdateById(id, req, new(ms.SysDict))
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

func (my MySql) CreateDictData(req *req.CreateDictData) (err error) {
	err = my.Create(req, new(ms.SysDictData))
	my.CacheFlushDictName(my.Ctx)
	my.CacheFlushDictNameAndKey(my.Ctx)
	return
}

func (my MySql) UpdateDictDataById(id uint, req req.UpdateDictData) (err error) {
	err = my.UpdateById(id, req, new(ms.SysDictData))
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
