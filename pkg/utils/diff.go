package utils

import (
	"github.com/golang-module/carbon"
	"github.com/shopspring/decimal"
	"reflect"
)

// compare oldStruct/newStruct to update
func CompareDiff(oldStruct interface{}, newStruct interface{}, update *map[string]interface{}) {
	// json to map
	m1 := make(map[string]interface{}, 0)
	m2 := make(map[string]interface{}, 0)
	m3 := make(map[string]interface{}, 0)
	Struct2StructByJson(newStruct, &m1)
	Struct2StructByJson(oldStruct, &m2)
	for k1, v1 := range m1 {
		for k2, v2 := range m2 {
			switch v1.(type) {
			// skip complex structure
			case map[string]interface{}:
				continue
			}
			rv := reflect.ValueOf(v1)
			if rv.Kind() != reflect.Invalid {
				// different value
				if k1 == k2 && v1 != v2 {
					t := reflect.TypeOf(oldStruct)
					key := CamelCase(k1)
					var fieldType reflect.Type
					oldStructV := reflect.ValueOf(oldStruct)
					// distinguish between map and structure
					if oldStructV.Kind() == reflect.Map {
						mapV := oldStructV.MapIndex(reflect.ValueOf(k1))
						if !mapV.IsValid() {
							break
						}
						fieldType = mapV.Type()
					} else if oldStructV.Kind() == reflect.Struct {
						structField, ok := t.FieldByName(key)
						if !ok {
							break
						}
						fieldType = structField.Type
					} else {
						// oldStruct not right skip
						break
					}
					realT := fieldType
					// pointer need to get element
					if fieldType.Kind() == reflect.Ptr {
						realT = fieldType.Elem()
					}
					e := reflect.New(realT).Elem()
					// decimal.Decimal/carbon.ToDateTimeString can not use Convert
					switch e.Interface().(type) {
					case decimal.Decimal:
						d, _ := decimal.NewFromString(rv.String())
						m3[k1] = d
					case carbon.ToDateTimeString:
						t := carbon.Parse(rv.String())
						// skip zero time
						if !t.IsZero() {
							m3[k1] = t
						}
					default:
						// rv convert to e
						e.Set(rv.Convert(realT))
						m3[k1] = e.Interface()
					}
					break
				}
			}
		}
	}
	*update = m3
}

// compare oldStruct/newStruct to update, map key to snake
func CompareDiff2SnakeKey(oldStruct interface{}, newStruct interface{}, update *map[string]interface{}) {
	m1 := make(map[string]interface{}, 0)
	m2 := make(map[string]interface{}, 0)
	CompareDiff(oldStruct, newStruct, &m1)
	for key, item := range m1 {
		m2[SnakeCase(key)] = item
	}
	*update = m2
}
