package req

import "github.com/piupuer/go-helper/pkg/resp"

type MenuReq struct {
	Name       string `json:"name" form:"name"`
	Title      string `json:"title" form:"title"`
	Path       string `json:"path" form:"path"`
	Component  string `json:"component" form:"component"`
	Redirect   string `json:"redirect"`
	Status     *uint  `json:"status" form:"status"`
	Visible    *uint  `json:"visible" form:"visible"`
	Breadcrumb *uint  `json:"breadcrumb" form:"breadcrumb"`
	resp.Page
}

type CreateMenuReq struct {
	Name       string   `json:"name" validate:"required"`
	Title      string   `json:"title"`
	Icon       string   `json:"icon"`
	Path       string   `json:"path"`
	Redirect   string   `json:"redirect"`
	Component  string   `json:"component"`
	Permission string   `json:"permission"`
	Sort       NullUint `json:"sort"`
	Status     NullUint `json:"status"`
	Visible    NullUint `json:"visible"`
	Breadcrumb NullUint `json:"breadcrumb"`
	ParentId   NullUint `json:"parentId"`
}

func (s CreateMenuReq) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Name"] = "name"
	return m
}

type UpdateMenuReq struct {
	Name       *string   `json:"name"`
	Title      *string   `json:"title"`
	Icon       *string   `json:"icon"`
	Path       *string   `json:"path"`
	Redirect   *string   `json:"redirect"`
	Component  *string   `json:"component"`
	Permission *string   `json:"permission"`
	Sort       *NullUint `json:"sort"`
	Status     *NullUint `json:"status"`
	Visible    *NullUint `json:"visible"`
	Breadcrumb *NullUint `json:"breadcrumb"`
	ParentId   *NullUint `json:"parentId"`
}

type UpdateMenuIncrementalIdsReq struct {
	Create []uint `json:"create"`
	Delete []uint `json:"delete"`
}
