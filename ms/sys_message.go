package ms

import "github.com/golang-module/carbon/v2"

const (
	// message type
	SysMessageTypeOneToOne     uint   = 0
	SysMessageTypeOneToMany    uint   = 1
	SysMessageTypeSystem       uint   = 2
	SysMessageTypeOneToOneStr  string = "one2one"
	SysMessageTypeOneToManyStr string = "one2more"
	SysMessageTypeSystemStr    string = "system"

	// message status
	SysMessageLogStatusUnRead     uint   = 0
	SysMessageLogStatusRead       uint   = 1
	SysMessageLogStatusDeleted    uint   = 2
	SysMessageLogStatusUnReadStr  string = "unread"
	SysMessageLogStatusReadStr    string = "read"
	SysMessageLogStatusDeletedStr string = "deleted"
)

var (
	SysMessageTypeConst = map[uint]string{
		SysMessageTypeOneToOne:  SysMessageTypeOneToOneStr,
		SysMessageTypeOneToMany: SysMessageTypeOneToManyStr,
		SysMessageTypeSystem:    SysMessageTypeSystemStr,
	}
	SysMessageLogStatusConst = map[uint]string{
		SysMessageLogStatusUnRead:  SysMessageLogStatusUnReadStr,
		SysMessageLogStatusRead:    SysMessageLogStatusReadStr,
		SysMessageLogStatusDeleted: SysMessageLogStatusDeletedStr,
	}
)

type SysMessage struct {
	M
	FromUserId uint             `gorm:"comment:sender user id" json:"fromUserId"`
	Title      string           `gorm:"comment:title" json:"title"`
	Content    string           `gorm:"comment:content" json:"content"`
	Type       uint             `gorm:"type:tinyint;default:0;comment:type(0: one2one, 1: one2more, 2: system(one2all))" json:"type"`
	RoleId     uint             `gorm:"comment:role id" json:"roleId"`
	ExpiredAt  *carbon.DateTime `gorm:"comment:expire time" json:"expiredAt"`
}

type SysMessageLog struct {
	M
	ToUserId  uint       `gorm:"comment:receiver user id" json:"toUserId"`
	MessageId uint       `gorm:"comment:message id" json:"messageId"`
	Message   SysMessage `gorm:"foreignKey:MessageId" json:"message"`
	Status    uint       `gorm:"type:tinyint;default:0;comment:status(0: unread, 1: read, 2: deleted)" json:"status"`
}
