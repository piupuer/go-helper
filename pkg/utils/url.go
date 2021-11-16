package utils

import (
	"fmt"
	"net/url"
	"reflect"
)

func Struct2UrlValues(obj interface{}) url.Values {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)
	params := url.Values{}
	for i := 0; i < rv.NumField(); i++ {
		key := rt.Field(i).Tag.Get("json")
		if rt.Field(i).Type.Kind() == reflect.Struct {
			child := Struct2UrlValues(rv.Field(i).Interface())
			if key == "" {
				// no json tag
				for k, list := range child {
					for _, item := range list {
						params.Add(k, item)
					}
				}
			}

			count := 0
			for range child {
				count++
			}
			if count > 0 && key != "" {
				params.Add(key, UrlValues2Json(child))
			}
			continue
		}
		if key == "" {
			key = rt.Field(i).Name
		}
		if rv.Field(i).Kind() == reflect.Slice {
			for j := 0; j < rv.Field(i).Len(); j++ {
				v := fmt.Sprintf("%v", rv.Field(i).Index(j).Interface())
				if v != "" {
					params.Add(key+"[]", v)
				}
			}
		} else {
			v := fmt.Sprintf("%v", rv.Field(i).Interface())
			if v != "" {
				params.Add(key, v)
			}
		}
	}
	return params
}

func UrlValues2Json(data url.Values) string {
	m := make(map[string]interface{})
	for key, item := range data {
		if len(item) == 1 {
			m[key] = item[0]
		} else {
			m[key] = item
		}
	}
	return Struct2Json(m)
}
