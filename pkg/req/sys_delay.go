package req

import "github.com/piupuer/go-helper/pkg/resp"

type DelayExportHistory struct {
	Name     string    `json:"name" form:"name"`
	Category string    `json:"category" form:"category"`
	End      *NullUint `json:"end" form:"end"`
	resp.Page
}
