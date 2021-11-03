package req

import (
	"github.com/piupuer/go-helper/pkg/resp"
	"time"
)

type OperationLog struct {
	Method   string `json:"method" form:"method"`
	Path     string `json:"path" form:"path"`
	Username string `json:"username" form:"username"`
	Ip       string `json:"ip" form:"ip"`
	Status   string `json:"status" form:"status"`
	resp.Page
}

type CreateOperationLog struct {
	ApiDesc    string        `json:"apiDesc"`
	Path       string        `json:"path"`
	Method     string        `json:"method"`
	Params     string        `json:"params"`
	Body       string        `json:"body"`
	Data       string        `json:"data"`
	Status     NullUint      `json:"status"`
	Username   string        `json:"username"`
	RoleName   string        `json:"roleName"`
	Ip         string        `json:"ip"`
	IpLocation string        `json:"ipLocation"`
	Latency    time.Duration `json:"latency"`
	UserAgent  string        `json:"userAgent"`
}
