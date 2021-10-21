package query

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/thedevsaddam/gojsonq/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

// query redis json value like gorm v2
type Redis struct {
	ops        RedisOptions
	redis      redis.UniversalClient
	Ctx        context.Context
	Error      error
	clone      int
	Statement  *Statement
	cacheStore *sync.Map
}

func NewRedis(client redis.UniversalClient, options ...func(*RedisOptions)) Redis {
	ops := getRedisOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if client == nil {
		panic("redis client is empty")
	}
	if ops.namingStrategy == nil {
		panic("redis namingStrategy is empty")
	}
	rd := Redis{
		ops:   *ops,
		redis: client,
		clone: 1,
	}
	rc := NewRequestId(ops.ctx, ops.requestIdCtxKey)
	rd.Ctx = rc
	return rd
}

// add error to db
func (rd *Redis) AddError(err error) error {
	if rd.Error == nil {
		rd.Error = err
	} else if err != nil {
		rd.Error = fmt.Errorf("%v; %w", rd.Error, err)
	}
	return rd.Error
}

// get new instance
func (rd Redis) Session() *Redis {
	if rd.clone > 0 {
		tx := &Redis{
			ops:   rd.ops,
			redis: rd.redis,
			Ctx:   rd.Ctx,
		}

		if rd.clone == 1 {
			// clone with new statement
			tx.Statement = &Statement{
				DB: tx,
			}
		} else {
			// gorm clone>1 for shared conditions, refer to https://gorm.io/zh_CN/docs/session.html#WithConditions
			// to reduce complexity, this step is omitted
		}

		return tx
	}

	return &rd
}

// check table name
func (rd Redis) check() bool {
	ins := rd.Session()
	// check table name when json is false
	if !ins.Statement.json && strings.TrimSpace(ins.Statement.Table) == "" {
		rd.Error = fmt.Errorf("invalid table name: '%s'", ins.Statement.Table)
		return false
	}
	return true
}

func (rd Redis) beforeQuery(db *Redis) *Redis {
	stmt := db.Statement
	if stmt.Model == nil {
		stmt.Model = stmt.Dest
	} else if stmt.Dest == nil {
		stmt.Dest = stmt.Model
	}

	if stmt.Model != nil && !stmt.count {
		if err := stmt.Parse(stmt.Model); err != nil && (!errors.Is(err, schema.ErrUnsupportedDataType) || (stmt.Table == "")) {
			db.AddError(err)
		}
	}

	if stmt.Dest != nil {
		stmt.ReflectValue = reflect.ValueOf(stmt.Dest)
		for stmt.ReflectValue.Kind() == reflect.Ptr {
			stmt.ReflectValue = stmt.ReflectValue.Elem()
		}
		if !stmt.ReflectValue.IsValid() {
			db.AddError(fmt.Errorf("invalid value"))
		}
	}
	return db
}

// find data from redis
func (rd *Redis) findByTableName(tableName string) *gojsonq.JSONQ {
	jsonStr := ""
	if !rd.Statement.json {
		cacheKey := fmt.Sprintf("%s_%s", rd.ops.database, tableName)
		var err error
		str, err := rd.redis.Get(rd.Ctx, cacheKey).Result()
		rd.ops.logger.Debug(rd.Ctx, "[query redis]read %s", tableName)
		if err != nil {
			rd.ops.logger.Debug(rd.Ctx, "[query redis]read %s err: %v", tableName, err)
		} else {
			// decompress
			jsonStr = utils.DeCompressStrByZlib(str)
		}
		if jsonStr == "" {
			jsonStr = "[]"
		}
	} else {
		jsonStr = rd.Statement.jsonStr
	}
	query := rd.jsonQuery(jsonStr)
	var nullList interface{}
	list := query.Get()
	if rd.Statement.first {
		// get first data
		switch list.(type) {
		case []interface{}:
			v, _ := list.([]interface{})
			if len(v) > 0 {
				list = v[0]
			} else {
				// set null
				list = nullList
			}
		}
	}
	// select count if dest is int64, skip query struct and preload
	if _, ok := rd.Statement.Dest.(*int64); !ok {
		utils.Struct2StructByJson(list, rd.Statement.Dest)
		if list != nil {
			rd.processPreload()
		} else {
			rd.AddError(gorm.ErrRecordNotFound)
		}
	}

	return query
}

// new jsonq query
func (rd Redis) jsonQuery(str string) *gojsonq.JSONQ {
	query := gojsonq.New().FromString(str)
	// add where
	for _, condition := range rd.Statement.whereConditions {
		query = query.Where(condition.key, condition.cond, condition.val)
	}
	// add order
	for _, condition := range rd.Statement.orderConditions {
		if condition.asc {
			query = query.SortBy(condition.property)
		} else {
			query = query.SortBy(condition.property, "desc")
		}
	}
	// add limit/offset
	query.Limit(rd.Statement.limit)
	query.Offset(rd.Statement.offset)
	return query
}
