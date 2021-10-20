package query

import (
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

type Statement struct {
	DB              *Redis
	Schema          *schema.Schema
	ReflectValue    reflect.Value
	Model           interface{}
	Table           string
	Dest            interface{}
	preloads        []searchPreload
	whereConditions []whereCondition
	orderConditions []orderCondition
	limit           int
	offset          int
	first           bool
	count           bool
	json            bool
	jsonStr         string
}

type searchPreload struct {
	schema string
}

// like jsonq Where
type whereCondition struct {
	key  string
	cond string
	val  interface{}
}

// like jsonq SortBy
type orderCondition struct {
	property string
	asc      bool
}

func (stmt *Statement) Preload(schema string) *Statement {
	var preloads []searchPreload
	for _, preload := range stmt.preloads {
		if preload.schema != schema {
			preloads = append(preloads, preload)
		}
	}
	preloads = append(preloads, searchPreload{schema})
	stmt.preloads = preloads
	return stmt
}

func (stmt *Statement) FromString(str string) *Statement {
	stmt.jsonStr = str
	stmt.json = true
	return stmt
}

func (stmt *Statement) Where(key, cond string, val interface{}) *Statement {
	// multiple conditions split by .
	keys := strings.Split(key, ".")
	newKeys := make([]string, 0)
	for _, item := range keys {
		// redis key is camel case
		newKeys = append(newKeys, utils.CamelCaseLowerFirst(item))
	}
	key = strings.Join(newKeys, ".")
	m1 := map[string]interface{}{
		"key": val,
	}
	var m2 map[string]interface{}
	utils.Struct2StructByJson(m1, &m2)
	v := m2["key"]
	arr, ok := v.([]interface{})
	if ok {
		newArr1 := make([]string, 0)
		newArr2 := make([]float64, 0)
		newArr3 := make([]int, 0)
		// json types: []string/[]float64/[]int
		for _, item := range arr {
			switch item.(type) {
			case string:
				newArr1 = append(newArr1, item.(string))
			case float64:
				newArr2 = append(newArr2, item.(float64))
			case int:
				newArr3 = append(newArr3, item.(int))
			}
		}
		if len(newArr1) > 0 {
			v = newArr1
		} else if len(newArr2) > 0 {
			v = newArr2
		} else if len(newArr3) > 0 {
			v = newArr3
		}
	}

	var whereConditions []whereCondition
	// old condition
	for _, condition := range stmt.whereConditions {
		if condition.key != key {
			whereConditions = append(whereConditions, condition)
		}
	}
	whereConditions = append(whereConditions, whereCondition{key, cond, v})
	stmt.whereConditions = whereConditions
	return stmt
}

func (stmt *Statement) Order(key string) *Statement {
	key = strings.ToLower(key)
	// multiple conditions split by space
	fields := strings.Split(key, " ")
	property := key
	asc := true
	// The order has been set when len = 2
	if len(fields) == 2 && strings.TrimSpace(fields[1]) == "desc" {
		property = fields[0]
		asc = false
	}
	property = utils.CamelCaseLowerFirst(property)

	var orderConditions []orderCondition
	for _, condition := range stmt.orderConditions {
		if condition.property != key {
			orderConditions = append(orderConditions, condition)
		}
	}
	orderConditions = append(orderConditions, orderCondition{property, asc})
	stmt.orderConditions = orderConditions
	return stmt
}

func (stmt *Statement) Parse(value interface{}) (err error) {
	if stmt.DB.cacheStore == nil {
		stmt.DB.cacheStore = &sync.Map{}
	}
	if stmt.DB.namingStrategy == nil {
		stmt.DB.namingStrategy = schema.NamingStrategy{}
	}
	if stmt.Schema, err = schema.Parse(value, stmt.DB.cacheStore, stmt.DB.namingStrategy); err == nil && stmt.Table == "" {
		if tables := strings.Split(stmt.Schema.Table, "."); len(tables) == 2 {
			stmt.Table = tables[1]
			return
		}

		stmt.Table = stmt.Schema.Table
	}
	return err
}
