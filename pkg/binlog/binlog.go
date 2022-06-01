package binlog

import (
	"context"
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/golang-module/carbon/v2"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// NewMysqlBinlog listen mysql binlog by siddontang/go-mysql
func NewMysqlBinlog(options ...func(*Options)) (err error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	if ops.db == nil {
		err = errors.Errorf("binlog db is empty")
		return
	}
	ops.db = ops.db.WithContext(ops.ctx)
	if ops.dsn == nil {
		err = errors.Errorf("binlog dsn is empty")
		return
	}
	if ops.redis == nil {
		err = errors.Errorf("binlog redis is empty")
		return
	}

	l := len(ops.models)
	tableNames := make([]string, l)
	for i := 0; i < l; i++ {
		tableNames[i] = getTableNameFromModel(*ops, ops.models[i])
	}
	// gen config
	cfg := canal.NewDefaultConfig()
	cfg.Addr = ops.dsn.Addr
	cfg.User = ops.dsn.User
	cfg.Password = ops.dsn.Passwd
	cfg.Flavor = "mysql"
	// cluster server id(random setting of single machine)
	cfg.ServerID = ops.serverId
	cfg.HeartbeatPeriod = 200 * time.Millisecond
	// use binlog only
	cfg.Dump.ExecutionPath = ""
	start(MysqlBinlog{
		ops:    *ops,
		cfg:    cfg,
		tables: tableNames,
	})
	return
}

type MysqlBinlog struct {
	ops    Options
	cfg    *canal.Config
	tables []string
}

func start(ins MysqlBinlog) {
	c, err := canal.NewCanal(ins.cfg)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	// event handler
	c.SetEventHandler(&EventHandler{
		ops:    ins.ops,
		tables: ins.tables,
	})
	// refresh cache before run
	err = refresh(ins.ops, ins.tables)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	// run from the last position
	pos, _ := c.GetMasterPos()
	go c.RunFrom(pos)
	go currentPos(ins, c)
}

func getTableNameFromModel(ops Options, model interface{}) (tableName string) {
	v := reflect.ValueOf(model)
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	m := v.MethodByName("TableName")
	if m.IsValid() {
		res := m.Call([]reflect.Value{})
		s, ok := res[0].Interface().(string)
		if ok {
			tableName = s
		}
	}
	if tableName == "" {
		tableName = ops.db.NamingStrategy.TableName(reflect.New(t).Elem().Type().Name())
	}
	return
}

// clear old redis cache
func refresh(ops Options, tableNames []string) error {
	for i, table := range tableNames {
		cacheKey := fmt.Sprintf("%s_%s", ops.dsn.DBName, table)
		// get old rows
		oldRows, err := getRows(ops, table, ops.models[i])
		if err != nil {
			continue
		}
		newRows := make([]map[string]interface{}, 0)
		for _, oldRow := range oldRows {
			row := make(map[string]interface{}, 0)
			for key, item := range oldRow {
				// gorm result map is camel case
				row[utils.CamelCaseLowerFirst(key)] = item
			}
			newRows = append(newRows, row)
		}
		// compress by zlib
		compress, err := utils.CompressStrByZlib(utils.Struct2Json(newRows))
		if err != nil {
			return errors.WithStack(err)
		}
		// set to redis
		err = ops.redis.Set(ops.ctx, cacheKey, compress, 0).Err()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func getRows(ops Options, table string, model interface{}) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)
	rows, err := ops.db.Table(table).Rows()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	mt := reflect.TypeOf(model).Elem()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		rows.Scan(columnPointers...)

		item := make(map[string]interface{}, 0)
		for i, colName := range cols {
			val := columns[i]
			var s interface{}
			// get model field
			field, exists := mt.FieldByName(utils.CamelCase(colName))
			if exists && val != nil {
				switch val.(type) {
				case time.Time:
					local := carbon.Time2Carbon(val.(time.Time))
					s = local.String()
				case []uint8:
					vs := string(val.([]uint8))
					k := field.Type.Kind()
					if field.Type.Kind() == reflect.Ptr {
						// if ptr field get real elem
						k = field.Type.Elem().Kind()
					}
					switch k {
					case reflect.Uint:
						s = utils.Str2Uint(vs)
					case reflect.Int:
						f, _ := strconv.Atoi(vs)
						s = f
					case reflect.Int64:
						f, _ := strconv.ParseInt(vs, 10, 64)
						s = f
					case reflect.Float64:
						f, _ := strconv.ParseFloat(vs, 64)
						s = f
					case reflect.Float32:
						f, _ := strconv.ParseFloat(vs, 32)
						s = f
					default:
						s = vs
					}
				}
			} else {
				s = nil
			}
			item[colName] = s
		}
		list = append(list, item)
	}
	return list, nil
}

type EventHandler struct {
	canal.DummyEventHandler
	ops    Options
	tables []string
}

// OnRow row change event
func (eh EventHandler) OnRow(e *canal.RowsEvent) error {
	if !utils.Contains(eh.tables, e.Table.Name) {
		return nil
	}
	ctx := tracing.NewId(nil)
	defer func() {
		if err := recover(); err != nil {
			log.
				WithContext(ctx).
				WithError(errors.Errorf("%v", err)).
				WithFields(map[string]interface{}{
					"Table":  e.Table.Name,
					"Action": e.Action,
					"Pos":    e.Header.LogPos,
				}).
				Error("runtime exception, stack: %v", string(debug.Stack()))
			return
		}
	}()
	RowChange(ctx, eh.ops, e)
	return nil
}

// OnDDL ddl event
func (eh EventHandler) OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) (err error) {
	ctx := tracing.NewId(nil)
	database := string(queryEvent.Schema)
	sql := strings.ToLower(string(queryEvent.Query))
	dropReg := regexp.MustCompile("drop table `(.+?)`")
	if dropReg != nil {
		// get drop table sql
		if m := dropReg.FindAllStringSubmatch(sql, -1); len(m) == 1 {
			table := strings.Trim(m[0][1], "`")
			cacheKey := fmt.Sprintf("%s_%s", database, table)
			err = eh.ops.redis.Del(ctx, cacheKey).Err()
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("drop table %s sync to redis failed", table)
			} else {
				log.WithContext(ctx).Info("drop table %s success", table)
			}
		}
	}
	if strings.Contains(sql, "truncate table") {
		table := ""
		arr := strings.Split(sql, " ")
		l := len(arr)
		for i, item := range arr {
			if item == "table" && i < l {
				table = strings.Trim(arr[i+1], "`")
			}
		}
		if table != "" {
			cacheKey := fmt.Sprintf("%s_%s", database, table)
			err = eh.ops.redis.Del(ctx, cacheKey).Err()
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("truncate table %s sync to redis failed", table)
			} else {
				log.WithContext(ctx).Info("truncate table %s success", table)
			}
		}
	}
	return
}

// OnPosSynced pos change event
func (eh EventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) (err error) {
	ctx := tracing.NewId(nil)
	defer func() {
		if e := recover(); e != nil {
			log.
				WithContext(ctx).
				WithError(errors.Errorf("%v", e)).
				WithFields(map[string]interface{}{
					"Name": pos.Name,
					"Pos":  pos.Pos,
				}).
				Error("runtime exception, stack: %s", string(debug.Stack()))
			return
		}
	}()
	err = eh.ops.redis.Set(ctx, fmt.Sprintf("%s_%s", eh.ops.dsn.DBName, eh.ops.binlogPos), utils.Struct2Json(pos), 0).Err()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("save pos failed")
	}
	return
}

func (eh EventHandler) String() string {
	return "EventHandler"
}

// show current position to know it is running
func currentPos(ins MysqlBinlog, c *canal.Canal) {
	for {
		fmt.Println(runtime.NumGoroutine())
		if c.Ctx().Err() == context.Canceled {
			log.
				WithContext(ins.ops.ctx).
				Info("canal exited, will restart after 10s")
			// restart after 10s
			time.Sleep(10 * time.Second)
			log.
				WithContext(ins.ops.ctx).
				Info("canal restarting...")
			start(ins)
			log.
				WithContext(ins.ops.ctx).
				Info("canal started")
			break
		}
		pos := c.SyncedPosition()
		log.
			WithContext(ins.ops.ctx).
			WithFields(map[string]interface{}{
				"Name": pos.Name,
				"Pos":  pos.Pos,
			}).Info("current position")
		time.Sleep(30 * time.Second)
	}
}
