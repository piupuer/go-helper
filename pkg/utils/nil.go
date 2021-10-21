package utils

import "reflect"

func InterfaceIsNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return i == nil
}
