package fsm

import (
	"github.com/piupuer/go-helper/models"
)

// 状态机(存储整个事件列表)
type Machine struct {
	models.Model
	Name                       string  `gorm:"comment:'状态机名称'" json:"name"`
	SubmitterName              string  `gorm:"comment:'提交者名称(可以是某个角色的统称)'" json:"submitterName"`
	SubmitterEditFields        string  `gorm:"comment:'提交者可编辑哪些字段'" json:"submitterEditFields"`
	SubmitterConfirm           uint    `gorm:"type:tinyint(1);default:0;comment:'提交者是否需要确认(0: 否, 1: 是)'" json:"submitterConfirm"`
	SubmitterConfirmEditFields string  `gorm:"comment:'提交者确认时可编辑哪些字段'" json:"submitterConfirmEditFields"`
	EventsJson                 string  `gorm:"comment:'事件json字符串(方便前端查询)'" json:"eventsJson"`
	Events                     []Event `gorm:"foreignKey:MId" json:"events"`
}

// 状态机事件
type Event struct {
	models.Model
	MId        uint        `gorm:"index:idx_m_id_sort,unique;comment:'状态机编号'" json:"mId"`
	M          Machine     `gorm:"foreignKey:MId" json:"m"`
	Sort       uint        `gorm:"index:idx_m_id_sort,unique;comment:'状态排序'" json:"sort"`
	Level      uint        `gorm:"comment:'等级(中间层级的通过和拒绝应该是同一level, 设定该值方便查找上下级)'" json:"level"`
	NameId     uint        `gorm:"comment:'状态名称编号'" json:"name"`
	Name       EventItem   `gorm:"foreignKey:NameId" json:"nameId"`
	Src        []EventItem `gorm:"many2many:event_src_item_relation;" json:"src"`
	DstId      uint        `gorm:"comment:'状态结束编号'" json:"dstId"`
	Dst        EventItem   `gorm:"foreignKey:DstId" json:"dst"`
	Edit       uint        `gorm:"type:tinyint(1);default:1;comment:'可编辑(0: 否, 1: 是)'" json:"edit"`
	EditFields string      `gorm:"comment:'可编辑字段列表(逗号隔开, edit=1时有效, 如果字段为空, 表示可编辑所有字段)'" json:"editFields"`
	Refuse     uint        `gorm:"type:tinyint(1);default:1;comment:'可拒绝(0: 否, 1: 是)'" json:"refuse"`
	Roles      []Role      `gorm:"many2many:event_role_relation;comment:'审批角色列表'" json:"roles"`
	Users      []User      `gorm:"many2many:event_user_relation;comment:'审批人列表'" json:"users"`
}

// 状态机审批人
type User struct {
	Id uint `gorm:"primaryKey;comment:'自增编号'" json:"id"`
}

// 状态机审批角色
type Role struct {
	Id uint `gorm:"primaryKey;comment:'自增编号'" json:"id"`
}

// 状态事件与源item关联关系
type EventSrcItemRelation struct {
	EventId     uint `json:"eventId"`
	EventItemId uint `json:"eventItemId"`
}

// 状态事件与角色关联关系
type EventRoleRelation struct {
	EventId uint `json:"eventId"`
	RoleId  uint `json:"roleId"`
}

// 状态事件与user关联关系
type EventUserRelation struct {
	EventId uint `json:"eventId"`
	UserId  uint `json:"userId"`
}

// 状态机事件项(存储具体名称)
type EventItem struct {
	Id   uint   `gorm:"primaryKey;comment:'自增编号'" json:"id"`
	Name string `gorm:"index:idx_name,unique;comment:'状态名称'" json:"name"`
}

// 状态机日志(存储每一条审批记录)
type Log struct {
	models.Model
	Category         uint      `gorm:"default:1;comment:'审批种类'" json:"category"`
	Uuid             string    `gorm:"comment:'唯一序号(同一次审批只能出现一次)'" json:"uuid"`
	Approved         uint      `gorm:"type:tinyint(1);default:0;comment:'状态(0:待审批,1:审批通过,2:审批拒绝,3:提交人取消)'" json:"approved"`
	ProgressId       uint      `gorm:"comment:'当前进度编号(关联事件名称)'" json:"progressId"`
	Progress         EventItem `gorm:"foreignKey:ProgressId" json:"progress"`
	SubmitterRoleId  uint      `gorm:"comment:'提交角色编号'" json:"submitterRoleId"`
	SubmitterUserId  uint      `gorm:"comment:'提交人编号'" json:"submitterUserId"`
	ApprovalRoleId   uint      `gorm:"comment:'审批角色编号'" json:"approvalRoleId"`
	ApprovalUserId   uint      `gorm:"comment:'审批人编号'" json:"approvalUserId"`
	ApprovalOpinion  string    `gorm:"comment:'审批意见'" json:"approvalOpinion"`
	PrevDetail       string    `gorm:"comment:'上一操作详情(如xxx审批通过/yyy审批拒绝)'" json:"prevDetail"`
	Detail           string    `gorm:"comment:'当前操作详情(如xxx审批通过/yyy审批拒绝)'" json:"detail"`
	CurrentEventId   uint      `gorm:"comment:'当前事件编号'" json:"currentEventId"`
	CurrentEvent     Event     `gorm:"foreignKey:CurrentEventId;comment:'当前事件'" json:"currentEvent"`
	NextEventId      uint      `gorm:"comment:'下一事件编号'" json:"nextEventId"`
	NextEvent        Event     `gorm:"foreignKey:NextEventId;comment:'下一事件'" json:"nextEvent"`
	CanApprovalRoles []Role    `gorm:"many2many:log_approval_role_relation;comment:'可参与审批的角色列表'" json:"canApprovalRoles"`
	CanApprovalUsers []User    `gorm:"many2many:log_approval_user_relation;comment:'可参与审批的用户列表'" json:"canApprovalUsers"`
}

// 可审批角色列表
type LogApprovalRoleRelation struct {
	LogId  uint `json:"logId"`
	RoleId uint `json:"roleId"`
}

// 可审批人列表
type LogApprovalUserRelation struct {
	LogId  uint `json:"logId"`
	UserId uint `json:"userId"`
}
