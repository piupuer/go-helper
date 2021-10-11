package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// 请求字符串转ReqUint
type ReqUint uint

func (r *ReqUint) UnmarshalJSON(data []byte) (err error) {
	str := strings.Trim(string(data), "\"")
	// ""空值不进行解析
	if str == "null" || strings.TrimSpace(str) == "" {
		*r = ReqUint(0)
		return
	}
	num, _ := strconv.ParseUint(str, 10, 64)
	*r = ReqUint(uint(num))
	return
}

func (r ReqUint) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", r)), nil
}

// gorm 写入 mysql 时调用
// driver.Value可取值int64/float64/bool/[]byte/string/time.Time
func (r ReqUint) Value() (driver.Value, error) {
	return int64(r), nil
}

// gorm 检出 mysql 时调用
func (r *ReqUint) Scan(v interface{}) error {
	value, ok := v.(ReqUint)
	if ok {
		*r = value
		return nil
	}
	return fmt.Errorf("can not convert %v to ReqUint", v)
}

