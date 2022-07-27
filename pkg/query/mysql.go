package query

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"github.com/thedevsaddam/gojsonq/v2"
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

func NewMySql(options ...func(*MysqlOptions)) MySql {
	ops := getMysqlOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.db == nil {
		panic("mysql db is empty")
	}
	my := MySql{}
	rc := tracing.NewId(ops.ctx)
	my.Ctx = rc
	tx := getTx(ops.db, *ops)
	my.Tx = tx.WithContext(rc)
	my.Db = ops.db.WithContext(rc)
	my.ops = *ops
	return my
}

func getTx(dbNoTx *gorm.DB, ops MysqlOptions) *gorm.DB {
	tx := dbNoTx
	if ops.ctx != nil {
		txValue := ops.ctx.Value(constant.MiddlewareTransactionTxCtxKey)
		if item, ok := txValue.(*gorm.DB); ok {
			tx = item
		}
	}
	return tx
}

func (my MySql) ForceCommit() {
	if my.ops.ctx != nil {
		if c, ok := my.ops.ctx.(*gin.Context); ok {
			c.Set(constant.MiddlewareTransactionForceCommitCtxKey, true)
		}
	}
}

func (my MySql) GetById(id uint, model interface{}, options ...func(*MysqlReadOptions)) {
	my.FindByColumns(id, model, options...)
}

func (my MySql) FindByIds(ids []uint, model interface{}, options ...func(*MysqlReadOptions)) {
	my.FindByColumns(ids, model, options...)
}

func (my MySql) FindByColumns(ids interface{}, model interface{}, options ...func(*MysqlReadOptions)) {
	my.FindByColumnsWithPreload(ids, model, options...)
}

func (my MySql) FindByColumnsWithPreload(ids interface{}, model interface{}, options ...func(*MysqlReadOptions)) {
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
		log.WithContext(my.Ctx).Warn("ids cannot be pointer")
		return
	}
	// get model val
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		log.WithContext(my.Ctx).Warn("model must be a pointer")
		return
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
		pre := "preload_nothing"
		if len(ops.preloads) > 0 {
			pre = "preload_" + strings.ToLower(strings.Join(ops.preloads, "_"))
		}
		// cache key: table+preloads+key+ids+modelIsArr
		cacheKey = fmt.Sprintf("%s_%s_%s_%s_find", structName, pre, ops.column, utils.Struct2Json(newIds))
		if rv.Elem().Kind() != reflect.Slice {
			cacheKey = fmt.Sprintf("%s_%s_%s_%s_first", structName, pre, ops.column, utils.Struct2Json(newIds))
		}
		cacheKey = fmt.Sprintf("%s_%s", my.ops.cachePrefix, cacheKey)
		oldCache, cacheErr := my.ops.redis.Get(my.Ctx, cacheKey).Result()
		if cacheErr == nil {
			list := gojsonq.New().FromString(oldCache).Get()
			if list != nil {
				arr := false
				noSize := false
				switch list.(type) {
				case []interface{}:
					v, _ := list.([]interface{})
					if len(v) > 0 {
						arr = true
					} else {
						noSize = true
					}
				default:
					noSize = true
				}
				if !noSize {
					// size > 0 parse data otherwise get data from db
					if rv.Elem().Kind() == reflect.Struct && arr {
						utils.Struct2StructByJson(list.([]interface{})[0], model)
					} else if rv.Elem().Kind() == reflect.Slice && !arr {
						// set first value
						newArr1 := reflect.MakeSlice(rv.Elem().Type(), 1, 1)
						v := newArr1.Index(0)
						i := reflect.New(rv.Elem().Type()).Interface()
						utils.Struct2StructByJson(list.([]interface{})[0], i)
						v.Set(reflect.ValueOf(i))
						// copy new array
						newArr2 := reflect.MakeSlice(rv.Elem().Type(), 1, 1)
						reflect.Copy(newArr2, newArr1)
						rv.Elem().Set(newArr2)
					} else {
						utils.Struct2StructByJson(list, model)
					}
					return
				}
			}
		}
	}
	// clone=2, so need to set q = q.xxx
	q := my.Tx.Session(&gorm.Session{NewDB: true})
	for _, item := range ops.preloads {
		q = q.Preload(item)
	}
	if !newIdsIsArr {
		q = q.
			Where(fmt.Sprintf("`%s` = ?", ops.column), newIds).
			First(model)
	} else {
		if newIdsIsArr && newIdsRv.Kind() != reflect.Slice {
			// column not primary, value maybe array
			q = q.
				Where(fmt.Sprintf("`%s` = ?", ops.column), firstId).
				Find(model)
		} else {
			q = q.
				Where(fmt.Sprintf("`%s` IN (?)", ops.column), newIds).
				Find(model)
		}
	}
	if ops.cache {
		expiration := time.Duration(ops.cacheExpire) * time.Second
		if rv.Elem().Kind() == reflect.Slice {
			// column not primary, value maybe array
			newArr := reflect.MakeSlice(rv.Elem().Type(), rv.Elem().Len(), rv.Elem().Len())
			reflect.Copy(newArr, rv.Elem())
			my.ops.redis.Set(my.Ctx, cacheKey, utils.Struct2Json(newArr.Interface()), expiration)
		} else {
			my.ops.redis.Set(my.Ctx, cacheKey, utils.Struct2Json(rv.Elem().Interface()), expiration)
		}
	}
	return
}

func (my MySql) FindWithPage(q *gorm.DB, page *resp.Page, model interface{}, options ...func(*MysqlReadOptions)) {
	ops := getMysqlReadOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.cache && my.ops.redis == nil {
		ops.cache = false
	}
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(my.Ctx).Warn("model must be a pointer")
		return
	}

	countCache := false
	if page.CountCache != nil {
		countCache = *page.CountCache
	}
	if !page.NoPagination {
		if !page.SkipCount {
			fromCache := false
			// get sql by DryRun
			stmt := q.Session(&gorm.Session{DryRun: true}).Count(&page.Total).Statement
			// SQL statement as cache key
			cacheKey := my.Tx.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
			if ops.cache && countCache {
				oldCount, cacheErr := my.ops.redis.Get(my.Ctx, cacheKey).Result()
				if cacheErr == nil {
					total := utils.Str2Int64(oldCount)
					page.Total = total
					fromCache = true
				}
			}
			if !fromCache {
				q.Count(&page.Total)
				if ops.cache && page.Total > 0 {
					my.ops.redis.Set(my.Ctx, cacheKey, page.Total, time.Duration(ops.cacheExpire)*time.Second)
				}
			} else {
				log.WithContext(my.Ctx).Debug("hit count cache: %s, total: %d", cacheKey, page.Total)
			}
		}
		if page.Total > 0 || page.SkipCount {
			limit, offset := page.GetLimit()
			if page.LimitPrimary == "" {
				q.Limit(limit).Offset(offset).Find(model)
			} else {
				// parse model
				if q.Statement.Model != nil {
					err := q.Statement.Parse(q.Statement.Model)
					if err != nil {
						log.WithContext(my.Ctx).WithError(err).Warn("parse model failed")
						return
					}
				}
				q.Joins(
					// add LimitPrimary index before join, improve q efficiency
					fmt.Sprintf(
						"JOIN (?) AS `OFFSET_T` ON `%s`.`id` = `OFFSET_T`.`%s`",
						q.Statement.Table,
						page.LimitPrimary,
					),
					q.
						Session(&gorm.Session{}).
						Select(
							fmt.Sprintf("`%s`.`%s`", q.Statement.Table, page.LimitPrimary),
						).
						Limit(limit).
						Offset(offset),
				).Find(model)
			}
		}
	} else {
		// no pagination
		q.Find(model)
		page.Total = int64(rv.Elem().Len())
		page.GetLimit()
	}
	page.CountCache = &countCache
	return
}

func (my MySql) FindWithSimplePage(q *gorm.DB, page *resp.Page, model interface{}) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(my.Ctx).Warn("model must be a pointer")
		return
	}
	countCache := false
	if page.CountCache != nil {
		countCache = *page.CountCache
	}
	if !page.NoPagination {
		if !page.SkipCount {
			q.Count(&page.Total)
		}
		if page.Total > 0 || page.SkipCount {
			limit, offset := page.GetLimit()
			q.Limit(limit).Offset(offset).Find(model)
		}
	} else {
		// no pagination
		q.Find(model)
		page.Total = int64(rv.Elem().Len())
		page.GetLimit()
	}
	page.CountCache = &countCache
}

func (my MySql) ScanWithPage(q *gorm.DB, page *resp.Page, model interface{}) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(my.Ctx).Warn("model must be a pointer")
		return
	}

	if !page.NoPagination {
		q.Count(&page.Total)
		if page.Total > 0 {
			limit, offset := page.GetLimit()
			q.Limit(limit).Offset(offset).Scan(model)
		}
	} else {
		q.Scan(model)
		page.Total = int64(rv.Elem().Len())
		page.GetLimit()
	}
	return
}

func (my MySql) Create(r interface{}, model interface{}) (err error) {
	i := model
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Slice {
		mv := reflect.Indirect(reflect.ValueOf(model))
		if mv.Kind() == reflect.Struct {
			slice := reflect.MakeSlice(reflect.SliceOf(mv.Type()), 0, 0)
			arr := reflect.New(slice.Type())
			i = arr.Interface()
		} else if mv.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(mv.Type(), 0, 0)
			arr := reflect.New(slice.Type())
			i = arr.Interface()
		}
	}
	utils.Struct2StructByJson(r, i)
	err = my.Tx.Create(i).Error
	return
}

func (my MySql) UpdateById(id uint, r interface{}, model interface{}) error {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Struct) {
		return errors.Errorf("model must be a pointer")
	}
	q := my.Tx.Model(rv.Interface()).Where("id = ?", id).First(rv.Interface())
	if errors.Is(q.Error, gorm.ErrRecordNotFound) {
		return errors.Errorf("can not get old record")
	}

	m := make(map[string]interface{}, 0)
	utils.CompareDiff2SnakeKey(rv.Elem().Interface(), r, &m)

	return q.Updates(&m).Error
}

func (my MySql) DeleteByIds(ids []uint, model interface{}) (err error) {
	return my.Tx.Where("id IN (?)", ids).Delete(model).Error
}
