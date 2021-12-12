package binlog

import (
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// listen mysql binlog by siddontang/go-mysql
func NewMysqlBinlog(options ...func(*Options)) error {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}

	if ops.db == nil {
		return errors.Errorf("binlog db is empty")
	}
	if ops.dsn == nil {
		return errors.Errorf("binlog dsn is empty")
	}
	if ops.redis == nil {
		return errors.Errorf("binlog redis is empty")
	}

	l := len(ops.models)
	tableNames := make([]string, l)
	for i := 0; i < l; i++ {
		t := reflect.ValueOf(ops.models[i]).Type()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		tableNames[i] = ops.db.NamingStrategy.TableName(reflect.New(t).Elem().Type().Name())
	}
	// gen config
	cfg := canal.NewDefaultConfig()
	cfg.Addr = ops.dsn.Addr
	cfg.User = ops.dsn.User
	cfg.Password = ops.dsn.Passwd
	cfg.Flavor = "mysql"
	// cluster server id(random setting of single machine)
	cfg.ServerID = ops.serverId
	// mysqldump path
	cfg.Dump.ExecutionPath = ops.executionPath
	// target database
	cfg.Dump.TableDB = ops.dsn.DBName
	// target tables
	cfg.Dump.Tables = tableNames

	c, err := canal.NewCanal(cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	// add ignore tables
	c.AddDumpIgnoreTables(cfg.Dump.TableDB, ops.ignores...)
	// event handler
	c.SetEventHandler(&EventHandler{
		ops: *ops,
	})
	// refresh cache before run
	err = refresh(*ops, tableNames)
	if err != nil {
		return errors.WithStack(err)
	}
	// run from the last position
	pos, _ := c.GetMasterPos()
	go c.RunFrom(pos)
	return nil
}

type EventHandler struct {
	canal.DummyEventHandler
	ops Options
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

// row change event
func (eh *EventHandler) OnRow(e *canal.RowsEvent) error {
	if utils.Contains(eh.ops.ignores, e.Table.Name) {
		return nil
	}
	defer func() {
		if err := recover(); err != nil {
			eh.ops.logger.Error(eh.ops.ctx, "[binlog row change]runtime err: %+v\nstack: %v", err, string(debug.Stack()))
			return
		}
	}()
	RowChange(eh.ops, e)
	eh.ops.logger.Debug(eh.ops.ctx, "[binlog row change]%s %v", e.Action, e.Rows)
	return nil
}

// ddl event
func (eh *EventHandler) OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	database := string(queryEvent.Schema)
	sql := strings.ToLower(string(queryEvent.Query))
	dropReg := regexp.MustCompile("drop table `(.+?)`")
	if dropReg != nil {
		// get drop table sql
		if m := dropReg.FindAllStringSubmatch(sql, -1); len(m) == 1 {
			table := strings.Trim(m[0][1], "`")
			cacheKey := fmt.Sprintf("%s_%s", database, table)
			err := eh.ops.redis.Del(eh.ops.ctx, cacheKey).Err()
			if err != nil {
				eh.ops.logger.Error(eh.ops.ctx, "[binlog ddl]drop table %s sync to redis err: %+v", table, err)
			} else {
				eh.ops.logger.Debug(eh.ops.ctx, "[binlog ddl]drop table %s success", table)
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
			err := eh.ops.redis.Del(eh.ops.ctx, cacheKey).Err()
			if err != nil {
				eh.ops.logger.Error(eh.ops.ctx, "[binlog ddl]truncate table %s sync to redis err: %+v", table, err)
			} else {
				eh.ops.logger.Debug(eh.ops.ctx, "[binlog ddl]truncate table %s success", table)
			}
		}
	}
	return nil
}

// pos change event
func (eh *EventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	defer func() {
		if err := recover(); err != nil {
			eh.ops.logger.Error(eh.ops.ctx, "[binlog pos change]runtime err: %v\nstack: %v", err, string(debug.Stack()))
			return
		}
	}()
	PosChange(eh.ops, pos)
	eh.ops.logger.Debug(eh.ops.ctx, "[binlog pos change]%s %v %t", pos, set, force)
	return nil
}

func (eh EventHandler) String() string {
	return "EventHandler"
}
