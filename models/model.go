package models

import (
	"github.com/golang-module/carbon"
)

const (
	Zero uint = iota
	One
	Two
	Three
	Four
	Five
)

// 由于gorm提供的base model没有json tag, 使用自定义
type Model struct {
	Id        uint                    `gorm:"primaryKey;comment:'自增编号'" json:"id"`
	CreatedAt carbon.ToDateTimeString `gorm:"comment:'创建时间'" json:"createdAt"`
	UpdatedAt carbon.ToDateTimeString `gorm:"comment:'更新时间'" json:"updatedAt"`
	DeletedAt DeletedAt               `gorm:"index:idx_deleted_at;comment:'删除时间(软删除)'" json:"deletedAt"`
}

// 响应结构体基础字段封装(如Id/CreatedAt/UpdatedAt等较常用字段)
type P struct {
	Id        uint                    `json:"id"`
	CreatedAt carbon.ToDateTimeString `json:"createdAt"`
	UpdatedAt carbon.ToDateTimeString `json:"updatedAt"`
}
