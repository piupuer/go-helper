package utils

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/shopspring/decimal"
	"reflect"
)

// CompareDiff compare oldStruct/newStruct to update
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
			case map[string]interface{}, []interface{}:
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
					case carbon.DateTime:
						t := carbon.Parse(rv.String())
						// skip zero time
						if !t.IsZero() {
							m3[k1] = t
						}
					default:
						// rv convert to e
						if rv.CanConvert(realT) {
							e.Set(rv.Convert(realT))
						} else {
							// call SetString method
							v := reflect.New(realT)
							m := v.MethodByName("SetString")
							if m.IsValid() {
								switch v1.(type) {
								case string:
									// v1 type is string
									m.Call([]reflect.Value{rv})
								default:
									// convert v1 to string
									m.Call([]reflect.Value{reflect.ValueOf(fmt.Sprintf("%v", v1))})
								}
							} else {
								log.Warn("%s's type %s missing SetString method, convert ignored", k1, realT.Name())
							}
							e = v.Elem()
						}
						m3[k1] = e.Interface()
					}
					break
				}
			}
		}
	}
	*update = m3
}

// CompareDiff2SnakeKey compare oldStruct/newStruct to update, map key to snake
func CompareDiff2SnakeKey(oldStruct interface{}, newStruct interface{}, update *map[string]interface{}) {
	m1 := make(map[string]interface{}, 0)
	m2 := make(map[string]interface{}, 0)
	CompareDiff(oldStruct, newStruct, &m1)
	for key, item := range m1 {
		m2[SnakeCase(key)] = item
	}
	*update = m2
}
