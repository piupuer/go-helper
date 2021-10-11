package request

import "github.com/piupuer/go-helper/models"

type CreateMachineReq struct {
	Name                       string           `json:"name"`
	SubmitterName              string           `json:"submitterName"`
	SubmitterEditFields        string           `json:"submitterEditFields"`
	SubmitterConfirm           models.ReqUint   `json:"submitterConfirm"`
	SubmitterConfirmEditFields string           `json:"submitterConfirmEditFields"`
	Levels                     []CreateEventReq `json:"levels"`
}

type CreateEventReq struct {
	Name       string         `json:"name" form:"name"`
	Edit       models.ReqUint `json:"edit" form:"edit"`
	Refuse     models.ReqUint `json:"refuse" form:"refuse"`
	EditFields string         `json:"editFields" form:"editFields"`
	Roles      []uint         `json:"roles" form:"roles"`
	Users      []uint         `json:"users" form:"users"`
}

type CreateLogReq struct {
	Category        uint   `json:"category" form:"category"`
	Uuid            string `json:"uuid" form:"uuid"`
	ApprovalRoleId  uint   `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId  uint   `json:"approvalUserId" form:"approvalUserId"`
	MId             uint   `json:"MId" form:"mId"`
	SubmitterRoleId uint   `json:"submitterRoleId" form:"submitterRoleId"`
	SubmitterUserId uint   `json:"submitterUserId" form:"submitterUserId"`
}

type ApproveLogReq struct {
	Category        uint   `json:"category" form:"category"`
	Uuid            string `json:"uuid" form:"uuid"`
	ApprovalRoleId  uint   `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId  uint   `json:"approvalUserId" form:"approvalUserId"`
	MId             uint   `json:"MId" form:"mId"`
	ApprovalOpinion string `json:"approvalOpinion" form:"approvalOpinion"`
	Approved        uint   `json:"approved" form:"approved"`
}

type PermissionLogReq struct {
	Category       uint   `json:"category" form:"category"`
	Uuid           string `json:"uuid" form:"uuid"`
	ApprovalRoleId uint   `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId uint   `json:"approvalUserId" form:"approvalUserId"`
	Approved       uint   `json:"approved" form:"approved"`
}

type PendingLogReq struct {
	ApprovalRoleId uint `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId uint `json:"approvalUserId" form:"approvalUserId"`
	Category       uint `json:"category" form:"category"`
}

type LogReq struct {
	Category uint   `json:"category" form:"category"`
	Uuid     string `json:"uuid" form:"uuid"`
}
