package query

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

type MySql struct {
	ops MysqlOptions
	Ctx context.Context
	Tx  *gorm.DB
	Db  *gorm.DB
}

func NewMySql(dbNoTx *gorm.DB, options ...func(*MysqlOptions)) MySql {
	ops := getMysqlOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	my := MySql{}
	tx := getTx(dbNoTx, *ops)
	rc := NewRequestId(ops.ctx, ops.requestIdCtxKey)
	my.Ctx = rc
	my.Tx = tx.WithContext(rc)
	my.Db = dbNoTx.WithContext(rc)
	return my
}

func getTx(dbNoTx *gorm.DB, ops MysqlOptions) *gorm.DB {
	tx := dbNoTx
	if ops.ctx != nil {
		method := ""
		if c, ok := ops.ctx.(*gin.Context); ok {
			if c.Request != nil {
				method = c.Request.Method
			}
		}
		if !(method == "OPTIONS" || method == "GET") {
			txValue := ops.ctx.Value(ops.txCtxKey)
			if item, ok := txValue.(*gorm.DB); ok {
				tx = item
			}
		}
	}
	return tx
}

// get one data by id
// model must be pointer
func (my MySql) GetById(id uint, model interface{}, options ...func(*MysqlReadOptions)) (err error) {
	return my.FindByColumns(id, model, options...)
}

// find data by ids
func (my MySql) FindByIds(ids []uint, model interface{}, options ...func(*MysqlReadOptions)) (err error) {
	return my.FindByColumns(ids, model, options...)
}

// find data by columns
func (my MySql) FindByColumns(ids interface{}, model interface{}, options ...func(*MysqlReadOptions)) (err error) {
	return my.FindByColumnsWithPreload(ids, model, options...)
}

// find data by columns with preload other tables
func (my MySql) FindByColumnsWithPreload(ids interface{}, model interface{}, options ...func(*MysqlReadOptions)) (err error) {
	ops := getMysqlReadOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.cache && my.ops.redis == nil {
		ops.cache = false
	}
	var newIds interface{}
	var firstId interface{}
	// check ids is array
	idsRv := reflect.ValueOf(ids)
	idsRt := reflect.TypeOf(ids)
	newIdsRv := reflect.ValueOf(newIds)
	newIdsIsArr := false
	if idsRv.Kind() == reflect.Ptr {
		return fmt.Errorf("ids cannot be pointer")
	}
	// get model val
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("model must be a pointer")
	}
	// check param/value len
	if idsRv.Kind() == reflect.Slice {
		if idsRv.Len() == 0 {
			return
		}
		// get first id
		firstId = idsRv.Index(0).Convert(idsRt.Elem()).Interface()
		if idsRv.Len() > 1 {
			// param is array, value is not array, use firstId
			if reflect.ValueOf(model).Elem().Kind() != reflect.Slice {
				newIds = firstId
			} else {
				// copy new array
				newArr := reflect.MakeSlice(reflect.TypeOf(ids), idsRv.Len(), idsRv.Len())
				reflect.Copy(newArr, idsRv)
				newIds = newArr.Interface()
				newIdsIsArr = true
			}
		} else {
			// len=0, reset value
			newIds = firstId
		}
	} else {
		firstId = ids
		newIds = ids
	}
	newIdsRv = reflect.ValueOf(newIds)

	// column not primary, value maybe array
	if ops.column != constant.QueryPrimaryKey && !newIdsIsArr && newIdsRv.Kind() != reflect.Slice && rv.Elem().Kind() == reflect.Slice {
		newIdsIsArr = true
	}
	// ids is array, model is not array, use firstId
	if newIdsIsArr && rv.Elem().Kind() != reflect.Slice {
		newIds = firstId
	}
	cacheKey := ""
	// set cache
	if ops.cache {
		structName := ""
		if rv.Elem().Kind() == reflect.Slice {
			structName = strings.ToLower(rv.Elem().Type().Elem().String())
		} else {
			structName = strings.ToLower(rv.Elem().Type().String())
		}
		preload := "preload_nothing"
		if len(ops.preloads) > 0 {
			preload = "preload_" + strings.ToLower(strings.Join(ops.preloads, "_"))
		}
		// cache key: table+preloads+key+ids+modelIsArr
		cacheKey = fmt.Sprintf("%s_%s_%s_%s_find", structName, preload, ops.column, utils.Struct2Json(newIds))
		if rv.Elem().Kind() != reflect.Slice {
			cacheKey = fmt.Sprintf("%s_%s_%s_%s_first", structName, preload, ops.column, utils.Struct2Json(newIds))
		}
		oldCache, cacheErr := my.ops.redis.Get(my.ops.ctx, cacheKey).Result()
		if cacheErr == nil {
			// model = oldCache
			crv := reflect.ValueOf(oldCache)
			if rv.Elem().Kind() == reflect.Struct && crv.Kind() == reflect.Slice {
				rv.Elem().Set(crv.Index(0))
			} else if rv.Elem().Kind() == reflect.Slice && crv.Kind() == reflect.Struct {
				// set first value
				newArr1 := reflect.MakeSlice(rv.Elem().Type(), 1, 1)
				v := newArr1.Index(0)
				v.Set(crv)
				// copy new array
				newArr2 := reflect.MakeSlice(rv.Elem().Type(), 1, 1)
				reflect.Copy(newArr2, newArr1)
				rv.Elem().Set(newArr2)
			} else if rv.Elem().Kind() == reflect.Slice && crv.Kind() == reflect.Slice {
				// copy new array
				newArr := reflect.MakeSlice(rv.Elem().Type(), crv.Len(), crv.Len())
				reflect.Copy(newArr, crv)
				rv.Elem().Set(newArr)
			} else {
				rv.Elem().Set(crv)
			}
			return
		}
	}
	query := my.Tx
	for _, preload := range ops.preloads {
		query = query.Preload(preload)
	}
	if !newIdsIsArr {
		err = query.
			Where(fmt.Sprintf("`%s` = ?", ops.column), newIds).
			First(model).Error
	} else {
		if newIdsIsArr && newIdsRv.Kind() != reflect.Slice {
			// column not primary, value maybe array
			err = query.
				Where(fmt.Sprintf("`%s` = ?", ops.column), firstId).
				Find(model).Error
		} else {
			err = query.
				Where(fmt.Sprintf("`%s` IN (?)", ops.column), newIds).
				Find(model).Error
		}
	}
	if ops.cache {
		expiration := time.Duration(ops.cacheExpire) * time.Second
		if rv.Elem().Kind() == reflect.Slice {
			// column not primary, value maybe array
			newArr := reflect.MakeSlice(rv.Elem().Type(), rv.Elem().Len(), rv.Elem().Len())
			reflect.Copy(newArr, rv.Elem())
			my.ops.redis.Set(my.ops.ctx, cacheKey, newArr.Interface(), expiration)
		} else {
			my.ops.redis.Set(my.ops.ctx, cacheKey, rv.Elem().Interface(), expiration)
		}
	}
	return
}

// find data by query condition
func (my MySql) Find(query *gorm.DB, page *resp.Page, model interface{}, options ...func(*MysqlReadOptions)) (err error) {
	ops := getMysqlReadOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.cache && my.ops.redis == nil {
		ops.cache = false
	}
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		return fmt.Errorf("model must be a pointer")
	}

	countCache := false
	if page.CountCache != nil {
		countCache = *page.CountCache
	}
	if !page.NoPagination {
		if !page.SkipCount {
			fromCache := false
			// get sql by DryRun
			stmt := query.Session(&gorm.Session{DryRun: true}).Count(&page.Total).Statement
			// SQL statement as cache key
			cacheKey := my.Tx.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
			if ops.cache && countCache {
				oldCount, cacheErr := my.ops.redis.Get(my.ops.ctx, cacheKey).Result()
				if cacheErr == nil {
					total := utils.Str2Int64(oldCount)
					page.Total = total
					fromCache = true
				}
			}
			if !fromCache {
				err = query.Count(&page.Total).Error
				if ops.cache && err == nil {
					my.ops.redis.Set(my.ops.ctx, cacheKey, page.Total, time.Duration(ops.cacheExpire)*time.Second)
				}
			} else {
				my.ops.logger.Debug(my.ops.ctx, "hit count cache: %s, total: %d", cacheKey, page.Total)
			}
		}
		if page.Total > 0 || page.SkipCount {
			limit, offset := page.GetLimit()
			if page.LimitPrimary == "" {
				err = query.Limit(limit).Offset(offset).Find(model).Error
			} else {
				// parse model
				if query.Statement.Model != nil {
					err = query.Statement.Parse(query.Statement.Model)
					if err != nil {
						return
					}
				}
				err = query.Joins(
					// add LimitPrimary index before join, improve query efficiency
					fmt.Sprintf(
						"JOIN (?) AS `OFFSET_T` ON `%s`.`id` = `OFFSET_T`.`%s`",
						query.Statement.Table,
						page.LimitPrimary,
					),
					query.
						Session(&gorm.Session{}).
						Select(
							fmt.Sprintf("`%s`.`%s`", query.Statement.Table, page.LimitPrimary),
						).
						Limit(limit).
						Offset(offset),
				).Find(model).Error
			}
		}
	} else {
		// no pagination
		err = query.Find(model).Error
		if err == nil {
			page.Total = int64(rv.Elem().Len())
			page.GetLimit()
		}
	}
	page.CountCache = &countCache
	return
}

// scan data  query condition(often used to JOIN)
func (my MySql) Scan(query *gorm.DB, page *resp.Page, model interface{}) (err error) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		return fmt.Errorf("model must be a pointer")
	}

	if !page.NoPagination {
		err = query.Count(&page.Total).Error
		if err == nil && page.Total > 0 {
			limit, offset := page.GetLimit()
			err = query.Limit(limit).Offset(offset).Scan(model).Error
		}
	} else {
		err = query.Scan(model).Error
		if err == nil {
			page.Total = int64(rv.Elem().Len())
			page.GetLimit()
		}
	}
	return
}

// create data
func (my MySql) Create(req interface{}, model interface{}) (err error) {
	utils.Struct2StructByJson(req, model)
	err = my.Tx.Create(model).Error
	return
}

// update data by id
func (my MySql) UpdateById(id uint, req interface{}, model interface{}) error {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Struct) {
		return fmt.Errorf("model must be a pointer")
	}
	query := my.Tx.Model(rv.Interface()).Where("id = ?", id).First(rv.Interface())
	if query.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("can not get old record")
	}

	m := make(map[string]interface{}, 0)
	utils.CompareDiff2SnakeKey(rv.Elem().Interface(), req, &m)

	return query.Updates(&m).Error
}

// batch delete by ids
func (my MySql) DeleteByIds(ids []uint, model interface{}) (err error) {
	return my.Tx.Where("id IN (?)", ids).Delete(model).Error
}
