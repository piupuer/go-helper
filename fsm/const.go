package fsm

import "fmt"

const (
	LogStatusWaiting   uint = iota // 待处理
	LogStatusApproved              // 通过
	LogStatusRefused               // 拒绝
	LogStatusCancelled             // 取消
)

var (
	ErrDbNil                  = fmt.Errorf("数据库实例不能为空")
	ErrEventsNil              = fmt.Errorf("未设置状态机事件")
	ErrEventNameNil           = fmt.Errorf("事件名称不能为空")
	ErrRepeatSubmit           = fmt.Errorf("审批记录已存在")
	ErrStatus                 = fmt.Errorf("审批状态不合法")
	ErrNoPermissionApprove    = fmt.Errorf("无权限通过审批")
	ErrNoPermissionRefuse     = fmt.Errorf("无权限拒绝审批")
	ErrNoPermissionOrEnded    = fmt.Errorf("无权限审批或审批已结束")
	ErrOnlySubmitterCancel    = fmt.Errorf("只有提交人才能取消")
	ErrStartedCannotCancel    = fmt.Errorf("流程已在进行中, 不得中途取消")
	ErrEventEndPointNotUnique = fmt.Errorf("事件结束位置不唯一或没有结束位置")
)

const (
	MsgSubmitterCancel = "提交人手动取消"
	MsgEnded           = "流程已结束"
	MsgConfigChanged   = "配置发生变化"
)

const (
	SuffixWaiting   = "待审批"
	SuffixResubmit  = "待提交"
	SuffixSubmitted = "已提交"
	SuffixApproved  = "已通过"
	SuffixRefused   = "已拒绝"
	SuffixConfirm   = "待确认"
	SuffixConfirmed = "已确认"
)
