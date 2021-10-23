package fsm

import "github.com/piupuer/go-helper/ms"

// finite state machine
type Machine struct {
	ms.M
	Name                       string  `gorm:"comment:'fsm name'" json:"name"`
	SubmitterName              string  `gorm:"comment:'submitter username or role name'" json:"submitterName"`
	SubmitterEditFields        string  `gorm:"comment:'submitter can edit fields'" json:"submitterEditFields"`
	SubmitterConfirm           uint    `gorm:"type:tinyint(1);default:0;comment:'submitter confirm(0: no, 1: yes)'" json:"submitterConfirm"`
	SubmitterConfirmEditFields string  `gorm:"comment:'submitter can edit fields when confirm'" json:"submitterConfirmEditFields"`
	EventsJson                 string  `gorm:"comment:'event json str'" json:"eventsJson"`
	Events                     []Event `gorm:"foreignKey:MachineId" json:"events"`
}

// fsm event
type Event struct {
	ms.M
	MachineId  uint        `gorm:"index:idx_m_id_sort,unique;" json:"machineId"`
	Machine    Machine     `gorm:"foreignKey:MachineId" json:"machine"`
	Sort       uint        `gorm:"index:idx_m_id_sort,unique;comment:'sort by level'" json:"sort"`
	Level      uint        `gorm:"comment:'level for query'" json:"level"`
	NameId     uint        `gorm:"comment:'current event'" json:"name"`
	Name       EventItem   `gorm:"foreignKey:NameId" json:"nameId"`
	Src        []EventItem `gorm:"many2many:event_src_item_relation;" json:"src"`
	DstId      uint        `gorm:"comment:'destination event'" json:"dstId"`
	Dst        EventItem   `gorm:"foreignKey:DstId" json:"dst"`
	Edit       uint        `gorm:"type:tinyint(1);default:1;comment:'approver can edit(0: no, 1: yes)'" json:"edit"`
	EditFields string      `gorm:"comment:'approver can edit fields(split by comma, can edit all field if it empty, edit=1 take effect)'" json:"editFields"`
	Refuse     uint        `gorm:"type:tinyint(1);default:1;comment:'approver can refuse(0: no, 1: yes)'" json:"refuse"`
	Roles      []Role      `gorm:"many2many:event_role_relation;comment:'approver role ids'" json:"roles"`
	Users      []User      `gorm:"many2many:event_user_relation;comment:'approver user ids'" json:"users"`
}

// fsm event user
type User struct {
	Id uint `gorm:"primaryKey;comment:'primary key'" json:"id"`
}

// fsm event role
type Role struct {
	Id uint `gorm:"primaryKey;comment:'primary key'" json:"id"`
}

type EventSrcItemRelation struct {
	EventId     uint `json:"eventId"`
	EventItemId uint `json:"eventItemId"`
}

type EventRoleRelation struct {
	EventId uint `json:"eventId"`
	RoleId  uint `json:"roleId"`
}

type EventUserRelation struct {
	EventId uint `json:"eventId"`
	UserId  uint `json:"userId"`
}

// fsm event item(save event name)
type EventItem struct {
	Id   uint   `gorm:"primaryKey;comment:'primary key'" json:"id"`
	Name string `gorm:"index:idx_name,unique;comment:'fsm event name'" json:"name"`
}

// fsm log(save every operation)
type Log struct {
	ms.M
	Category         uint      `gorm:"default:1;comment:'custom category(>0)'" json:"category"`
	Uuid             string    `gorm:"comment:'unique str'" json:"uuid"`
	Approved         uint      `gorm:"type:tinyint(1);default:0;comment:'approval status'" json:"approved"`
	ProgressId       uint      `gorm:"comment:'current progress'" json:"progressId"`
	Progress         EventItem `gorm:"foreignKey:ProgressId" json:"progress"`
	SubmitterRoleId  uint      `gorm:"comment:'custom submitter role id'" json:"submitterRoleId"`
	SubmitterUserId  uint      `gorm:"comment:'custom submitter user id'" json:"submitterUserId"`
	ApprovalRoleId   uint      `gorm:"comment:'approver role id'" json:"approvalRoleId"`
	ApprovalUserId   uint      `gorm:"comment:'approver user id'" json:"approvalUserId"`
	ApprovalOpinion  string    `gorm:"comment:'approver approval opinion'" json:"approvalOpinion"`
	PrevDetail       string    `gorm:"comment:'last approver detail'" json:"prevDetail"`
	Detail           string    `gorm:"comment:'current approver detail'" json:"detail"`
	CurrentEventId   uint      `gorm:"comment:'current event id'" json:"currentEventId"`
	CurrentEvent     Event     `gorm:"foreignKey:CurrentEventId;comment:'current event'" json:"currentEvent"`
	NextEventId      uint      `gorm:"comment:'next event id'" json:"nextEventId"`
	NextEvent        Event     `gorm:"foreignKey:NextEventId;comment:'next event'" json:"nextEvent"`
	CanApprovalRoles []Role    `gorm:"many2many:log_approval_role_relation;comment:'can approve roles'" json:"canApprovalRoles"`
	CanApprovalUsers []User    `gorm:"many2many:log_approval_user_relation;comment:'can approve users'" json:"canApprovalUsers"`
}

type LogApprovalRoleRelation struct {
	LogId  uint `json:"logId"`
	RoleId uint `json:"roleId"`
}

type LogApprovalUserRelation struct {
	LogId  uint `json:"logId"`
	UserId uint `json:"userId"`
}
