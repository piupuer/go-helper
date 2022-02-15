package req

import "github.com/piupuer/go-helper/pkg/resp"

type DelayExportHistory struct {
	Category string    `json:"category" form:"category"`
	End      *NullUint `json:"end" form:"end"`
	resp.Page
}
