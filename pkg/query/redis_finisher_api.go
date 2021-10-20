package query

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/resp"
	"reflect"
)

// gorm.Find
func (rd Redis) Find(dest interface{}) *Redis {
	ins := rd.Session()
	if !ins.check() {
		return ins
	}
	ins.Statement.Dest = dest
	ins.Statement.Model = dest
	ins.Statement.count = false
	ins.beforeQuery(ins).findByTableName(ins.Statement.Table)
	return ins
}

// like gorm.First
func (rd Redis) First(dest interface{}) *Redis {
	ins := rd.Session()
	ins.Statement.limit = 1
	ins.Statement.first = true
	ins.Find(dest)
	return ins
}

// like gorm.Count
func (rd Redis) Count(count *int64) *Redis {
	ins := rd.Session()
	ins.Statement.Dest = count
	ins.Statement.count = true
	if !ins.check() {
		*count = 0
	}
	ins.Statement.Dest = count
	*count = int64(ins.beforeQuery(ins).findByTableName(ins.Statement.Table).Count())
	return ins
}

func (rd Redis) FindWithPage(query *Redis, page *resp.Page, model interface{}) (err error) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		return fmt.Errorf("model must be a pointer")
	}

	if !page.NoPagination {
		err = query.Count(&page.Total).Error
		if err == nil && page.Total > 0 {
			limit, offset := page.GetLimit()
			err = query.Limit(limit).Offset(offset).Find(model).Error
		}
	} else {
		err = query.Find(model).Error
		if err == nil {
			page.Total = int64(rv.Elem().Len())
			page.GetLimit()
		}
	}
	return
}
