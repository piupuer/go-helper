package delay

import "github.com/piupuer/go-helper/ms"

// ExportHistory save export file history
type ExportHistory struct {
	ms.M
	Uuid     string `gorm:"index:idx_uuid,unique;:comment:'uuid'" json:"uuid"`
	Category string `gorm:"comment:'custom category'" json:"category"`
	Name     string `gorm:"comment:'display name'" json:"name"`
	Progress string `gorm:"comment:'process progress'" json:"progress"`
	End      uint   `gorm:"type:tinyint(1);default:0;comment:'0: pending, 1: end)'" json:"end"`
	Url      string `gorm:"comment:'cloud file url'" json:"url"`
}
