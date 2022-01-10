package utils

import (
	"encoding/json"
	"fmt"
)

func Struct2Json(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("[Struct2Json]can not convert: %+v\n", err)
	}
	return string(str)
}

func Json2Struct(str string, obj interface{}) {
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		fmt.Printf("[Json2Struct]can not convert: %+v\n", err)
	}
}

// struct2 must be pointer
func Struct2StructByJson(struct1 interface{}, struct2 interface{}) {
	jsonStr := Struct2Json(struct1)
	Json2Struct(jsonStr, struct2)
}

// sort json keys
func JsonWithSort(str string) string {
	m := make(map[string]interface{})
	Json2Struct(str, &m)
	return Struct2Json(m)
}
