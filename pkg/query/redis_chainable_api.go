package query

import (
	"regexp"
	"strings"
)

var tableRegexp = regexp.MustCompile(`(?i).+? AS (\w+)\s*(?:$|,)`)

// set json str
func (rd Redis) FromString(str string) *Redis {
	ins := rd.Session()
	ins.Statement.FromString(str)
	ins.Statement.Dest = nil
	return ins
}

// set table name
func (rd Redis) Table(name string, args ...interface{}) (ins *Redis) {
	ins = rd.Session()
	if strings.Contains(name, " ") || strings.Contains(name, "`") || len(args) > 0 {
		if results := tableRegexp.FindStringSubmatch(name); len(results) == 2 {
			ins.Statement.Table = rd.ops.namingStrategy.TableName(results[1])
			return
		}
	}

	ins.Statement.Table = rd.ops.namingStrategy.TableName(name)
	return
}

// preload column
func (rd Redis) Preload(column string) *Redis {
	return rd.Session().Statement.Preload(column).DB
}

// where condition
func (rd Redis) Where(key, cond string, val interface{}) *Redis {
	return rd.Session().Statement.Where(key, cond, val).DB
}

// sort condition
func (rd Redis) Order(key string) *Redis {
	return rd.Session().Statement.Order(key).DB
}

// limit condition
func (rd Redis) Limit(limit int) *Redis {
	ins := rd.Session()
	ins.Statement.limit = limit
	return ins
}

// offset condition
func (rd Redis) Offset(offset int) *Redis {
	ins := rd.Session()
	ins.Statement.offset = offset
	return ins
}
