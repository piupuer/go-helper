package req

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// request param uint
type NullUint uint

func (r *NullUint) UnmarshalJSON(data []byte) (err error) {
	str := strings.Trim(string(data), "\"")
	// skip empty str value
	if str == "null" || strings.TrimSpace(str) == "" {
		*r = NullUint(0)
		return
	}
	num, _ := strconv.ParseUint(str, 10, 64)
	*r = NullUint(uint(num))
	return
}

func (r NullUint) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", r)), nil
}

// gorm set to mysql
// driver.Value types: int64/float64/bool/[]byte/string/time.Time
func (r NullUint) Value() (driver.Value, error) {
	return int64(r), nil
}

// gorm get from mysql
func (r *NullUint) Scan(v interface{}) error {
	value, ok := v.(NullUint)
	if ok {
		*r = value
		return nil
	}
	return fmt.Errorf("can not convert %v to NullUint", v)
}
