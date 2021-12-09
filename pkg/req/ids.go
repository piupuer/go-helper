package req

import (
	"github.com/piupuer/go-helper/pkg/utils"
)

type Ids struct {
	Ids string `json:"ids" form:"ids"` // id array string, split by comma
}

func (id Ids) Uints() []uint {
	return utils.Str2UintArr(id.Ids)
}

func (id Ids) Ints() []int {
	return utils.Str2IntArr(id.Ids)
}

func (id Ids) Int64s() []int64 {
	return utils.Str2Int64Arr(id.Ids)
}

type IdsStr string

func (s IdsStr) Uints() []uint {
	return utils.Str2UintArr(string(s))
}

func (s IdsStr) Ints() []int {
	return utils.Str2IntArr(string(s))
}

func (s IdsStr) Int64s() []int64 {
	return utils.Str2Int64Arr(string(s))
}
