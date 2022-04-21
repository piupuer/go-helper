package resp

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/utils"
)

type Time struct {
	CreatedAt carbon.DateTime `json:"createdAt" swaggertype:"string" example:"2019-01-01 00:00:00"` // create time
	UpdatedAt carbon.DateTime `json:"updatedAt" swaggertype:"string" example:"2019-01-01 00:00:00"` // update time
}

// base fields(like Id/CreatedAt/UpdatedAt common fields)
type Base struct {
	Id uint `json:"id"` // primary key
	Time
}

// http resp structure
type Resp struct {
	Code      int         `json:"code" enums:"201,401,403,405,500"`                         // response code
	Data      interface{} `json:"data" swaggertype:"string" example:"{}"`                   // response data if code=201
	Msg       string      `json:"msg" example:"success"`                                    // response msg
	RequestId string      `json:"requestId" example:"4cb6e3f6-1f52-4fba-9b7d-e65098600f02"` // request id
}

// array data page info
type Page struct {
	PageNum      uint   `json:"pageNum" form:"pageNum"`           // current page
	PageSize     uint   `json:"pageSize" form:"pageSize"`         // page per count
	Total        int64  `json:"total"`                            // all data count
	NoPagination bool   `json:"noPagination" form:"noPagination"` // query all data
	CountCache   *bool  `json:"countCache" form:"countCache"`     // use count cache
	SkipCount    bool   `json:"skipCount" form:"skipCount"`       // not use 'SELECT count(*) FROM ...' before 'SELECT * FROM ...'
	LimitPrimary string `json:"-"`                                // When there is a large amount of data, limit is optimized by specifying a field (the field is usually self incremented ID or indexed), which can improve the query efficiency (if it is not transmitted, it will not be optimized)
}

// array data page with list
type PageData struct {
	Page
	List interface{} `json:"list"`
}

// calc limit/offset
func (s *Page) GetLimit() (int, int) {
	var pageSize int64
	var pageNum int64
	total := s.Total
	// PageSize must be greater than 0
	if s.PageSize < constant.MinPageSize || s.PageSize > constant.MaxPageSize {
		pageSize = constant.PageSize
	} else {
		pageSize = int64(s.PageSize)
	}
	// PageNum must be greater than 0
	if s.PageNum < constant.MinPageNum {
		pageNum = constant.MinPageNum
	} else {
		pageNum = int64(s.PageNum)
	}

	// calc maxPageNum
	maxPageNum := total/pageSize + 1
	if total%pageSize == 0 {
		maxPageNum = total / pageSize
	}
	// maxPageNum must be greater than 0
	if maxPageNum < constant.MinPageNum {
		maxPageNum = constant.MinPageNum
	}
	// pageNum must be less than or equal to total
	if total > 0 && pageNum > total {
		pageNum = maxPageNum
	}

	limit := pageSize
	offset := limit * (pageNum - 1)
	// PageNum less than 1 is set as page 1 data
	if s.PageNum < 1 {
		offset = 0
	}

	// PageNum greater than maxPageNum is set as empty data: offset=last
	if total > 0 && int64(s.PageNum) > maxPageNum {
		pageNum = maxPageNum + 1
		offset = limit * maxPageNum
	}

	s.PageNum = uint(pageNum)
	s.PageSize = uint(pageSize)
	if s.NoPagination {
		s.PageSize = uint(total)
	}
	// gorm v2 interface is int
	return int(limit), int(offset)
}

func GetResult(code int, data interface{}, format interface{}, a ...interface{}) Resp {
	var f string
	switch format.(type) {
	case string:
		f = format.(string)
	case error:
		f = fmt.Sprintf("%v", format.(error))
	}
	return Resp{
		Code: code,
		Data: data,
		Msg:  fmt.Sprintf(f, a...),
	}
}

func GetSuccess() Resp {
	return GetResult(Ok, map[string]interface{}{}, CustomError[Ok])
}

func GetSuccessWithData(data ...interface{}) Resp {
	switch len(data) {
	case 1:
		return GetResult(Ok, data[0], CustomError[Ok])
	case 2:
		utils.Struct2StructByJson(data[0], data[1])
		return GetResult(Ok, data[1], CustomError[Ok])
	}
	return GetSuccess()
}

func GetSuccessWithPageData(real, brief interface{}, page Page) Resp {
	utils.Struct2StructByJson(real, brief)
	var rp PageData
	rp.Page = page
	rp.List = brief
	return GetResult(Ok, rp, CustomError[Ok])
}

func GetFailWithMsg(format interface{}, a ...interface{}) Resp {
	return GetResult(NotOk, map[string]interface{}{}, format, a...)
}

func GetFailWithCode(code int) Resp {
	// default NotOk
	msg := CustomError[NotOk]
	if val, ok := CustomError[code]; ok {
		msg = val
	}
	return GetResult(code, map[string]interface{}{}, msg)
}

func GetFailWithCodeAndMsg(code int, format interface{}, a ...interface{}) Resp {
	return GetResult(code, map[string]interface{}{}, format, a...)
}

func Success() {
	panic(GetSuccess())
}

func SuccessWithData(data ...interface{}) {
	panic(GetSuccessWithData(data...))
}

func SuccessWithPageData(real, brief interface{}, page Page) {
	panic(GetSuccessWithPageData(real, brief, page))
}

func FailWithMsg(format interface{}, a ...interface{}) {
	panic(GetFailWithMsg(format, a...))
}

func FailWithCode(code int) {
	panic(GetFailWithCode(code))
}

func FailWithCodeAndMsg(code int, format interface{}, a ...interface{}) {
	panic(GetFailWithCodeAndMsg(code, format, a...))
}

func CheckErr(format interface{}, a ...interface{}) {
	var f string
	switch format.(type) {
	case string:
		f = format.(string)
	case error:
		f = fmt.Sprintf("%v", format.(error))
	}
	if f != "" {
		FailWithMsg(f, a...)
	}
}
