package req

import "github.com/piupuer/go-helper/pkg/resp"

type Dict struct {
	Name   string    `json:"name" form:"name"`
	Desc   string    `json:"desc" form:"desc"`
	Status *NullUint `json:"status" form:"status"`
	resp.Page
}

type CreateDict struct {
	Name   string    `json:"name" validate:"required"`
	Desc   string    `json:"desc" validate:"required"`
	Status *NullUint `json:"status"`
}

func (s CreateDict) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Name"] = "dict name"
	m["Desc"] = "dict description"
	return m
}

type UpdateDict struct {
	Name   *string   `json:"name"`
	Desc   *string   `json:"desc"`
	Status *NullUint `json:"status"`
}

type DictData struct {
	DictId *NullUint `json:"dictId" form:"dictId"`
	Key    string    `json:"key" form:"key"`
	Val    string    `json:"val" form:"val"`
	Status *NullUint `json:"status" form:"status"`
	resp.Page
}

type CreateDictData struct {
	Key      string    `json:"key" validate:"required"`
	Val      string    `json:"val" validate:"required"`
	Addition string    `json:"addition"`
	Sort     *NullUint `json:"sort"`
	Status   *NullUint `json:"status"`
	DictId   uint      `json:"dictId" validate:"required"`
}

func (s CreateDictData) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Key"] = "key"
	m["Val"] = "value"
	m["DictId"] = "dict id"
	return m
}

type UpdateDictData struct {
	Key      *string   `json:"key"`
	Val      *string   `json:"val"`
	Addition *string   `json:"addition"`
	Sort     *NullUint `json:"sort"`
	Status   *NullUint `json:"status"`
	DictId   *NullUint `json:"dictId"`
}
