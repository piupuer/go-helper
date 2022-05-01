package query

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"github.com/thedevsaddam/gojsonq/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

// q redis json value like gorm v2
type Redis struct {
	ops        RedisOptions
	Ctx        context.Context
	Error      error
	clone      int
	Statement  *Statement
	cacheStore *sync.Map
}

func NewRedis(options ...func(*RedisOptions)) Redis {
	ops := getRedisOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.redis == nil {
		if ops.redisUri != "" {
			var err error
			ops.redis, err = ParseRedisURI(ops.redisUri)
			if err != nil {
				panic(err)
			}
		} else {
			panic("redis client is empty")
		}
	}
	if ops.namingStrategy == nil {
		panic("redis namingStrategy is empty")
	}
	rd := Redis{
		ops:   *ops,
		clone: 1,
	}
	rc := tracing.NewId(ops.ctx)
	rd.Ctx = rc
	return rd
}

func ParseRedisURI(uri string) (client redis.UniversalClient, err error) {
	var opt asynq.RedisConnOpt
	if uri != "" {
		opt, err = asynq.ParseRedisURI(uri)
		if err != nil {
			return
		}
		client = opt.MakeRedisClient().(redis.UniversalClient)
		return
	}
	err = errors.Errorf("invalid redis config")
	return
}

// add error to db
func (rd *Redis) AddError(err error) error {
	if rd.Error == nil {
		rd.Error = errors.WithStack(err)
	} else if err != nil {
		rd.Error = errors.Wrapf(rd.Error, "%v", err)
	}
	return rd.Error
}

// create new db session
func (rd *Redis) Session() *Redis {
	return &Redis{
		ops:       rd.ops,
		Ctx:       rd.Ctx,
		Statement: rd.Statement,
		Error:     rd.Error,
		clone:     1,
	}
}

// get new instance
func (rd Redis) getInstance() *Redis {
	if rd.clone > 0 {
		tx := &Redis{
			ops: rd.ops,
			Ctx: rd.Ctx,
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
	ins := rd.getInstance()
	// check table name when json is false
	if !ins.Statement.json && strings.TrimSpace(ins.Statement.Table) == "" {
		rd.Error = errors.Errorf("invalid table name: '%s'", ins.Statement.Table)
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
		str, err := rd.ops.redis.Get(rd.Ctx, cacheKey).Result()
		log.WithContext(rd.Ctx).Debug("[q redis]read %s", tableName)
		if err != nil {
			log.WithContext(rd.Ctx).WithError(err).Warn("[q redis]read %s failed", tableName)
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
	q := rd.jsonQuery(jsonStr)
	var nullList interface{}
	list := q.Get()
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
	// select count if dest is int64, skip q struct and preload
	if _, ok := rd.Statement.Dest.(*int64); !ok {
		utils.Struct2StructByJson(list, rd.Statement.Dest)
		if list != nil {
			rd.processPreload()
		} else {
			rd.AddError(errors.WithStack(gorm.ErrRecordNotFound))
		}
	}

	return q
}

// new jsonq q
func (rd Redis) jsonQuery(str string) *gojsonq.JSONQ {
	q := gojsonq.New().FromString(str)
	// add where
	for _, condition := range rd.Statement.whereConditions {
		q.Where(condition.key, condition.cond, condition.val)
	}
	// add order
	for _, condition := range rd.Statement.orderConditions {
		if condition.asc {
			q.SortBy(condition.property)
		} else {
			q.SortBy(condition.property, "desc")
		}
	}
	// add limit/offset
	q.Limit(rd.Statement.limit)
	q.Offset(rd.Statement.offset)
	return q
}
