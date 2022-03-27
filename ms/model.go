package ms

import "github.com/golang-module/carbon"

type M struct {
	Id        uint                    `gorm:"primaryKey;comment:auto increment id" json:"id"`
	CreatedAt carbon.ToDateTimeString `gorm:"comment:create time" json:"createdAt"`
	UpdatedAt carbon.ToDateTimeString `gorm:"comment:update time" json:"updatedAt"`
	DeletedAt DeletedAt               `gorm:"index:idx_deleted_at;comment:soft delete time" json:"deletedAt"`
}
