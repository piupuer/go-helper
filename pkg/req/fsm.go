package req

type FsmCreateMachine struct {
	Name                       string           `json:"name"`
	SubmitterName              string           `json:"submitterName"`
	SubmitterEditFields        string           `json:"submitterEditFields"`
	SubmitterConfirm           NullUint         `json:"submitterConfirm"`
	SubmitterConfirmEditFields string           `json:"submitterConfirmEditFields"`
	Levels                     []FsmCreateEvent `json:"levels"`
}

type FsmCreateEvent struct {
	Name       string   `json:"name" form:"name"`
	Edit       NullUint `json:"edit" form:"edit"`
	Refuse     NullUint `json:"refuse" form:"refuse"`
	EditFields string   `json:"editFields" form:"editFields"`
	Roles      []uint   `json:"roles" form:"roles"`
	Users      []uint   `json:"users" form:"users"`
}

type FsmCreateLog struct {
	Category        NullUint `json:"category" form:"category"`
	Uuid            string   `json:"uuid" form:"uuid"`
	MachineId       uint     `json:"machineId" form:"machineId"`
	SubmitterRoleId uint     `json:"submitterRoleId" form:"submitterRoleId"`
	SubmitterUserId uint     `json:"submitterUserId" form:"submitterUserId"`
}

type FsmApproveLog struct {
	Category        NullUint `json:"category" form:"category"`
	Uuid            string   `json:"uuid" form:"uuid"`
	ApprovalRoleId  uint     `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId  uint     `json:"approvalUserId" form:"approvalUserId"`
	MachineId       uint     `json:"machineId" form:"machineId"`
	ApprovalOpinion string   `json:"approvalOpinion" form:"approvalOpinion"`
	Approved        NullUint `json:"approved" form:"approved"`
}

type FsmPermissionLog struct {
	Category       NullUint `json:"category" form:"category"`
	Uuid           string   `json:"uuid" form:"uuid"`
	ApprovalRoleId uint     `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId uint     `json:"approvalUserId" form:"approvalUserId"`
	Approved       uint     `json:"approved" form:"approved"`
}

type FsmPendingLog struct {
	ApprovalRoleId uint     `json:"approvalRoleId" form:"approvalRoleId"`
	ApprovalUserId uint     `json:"approvalUserId" form:"approvalUserId"`
	Category       NullUint `json:"category" form:"category"`
}

type FsmLog struct {
	Category NullUint `json:"category" form:"category"`
	Uuid     string   `json:"uuid" form:"uuid"`
}

type FsmMachine struct {
	Name             string    `json:"name" form:"name"`
	SubmitterName    string    `json:"submitterName" form:"submitterName"`
	SubmitterConfirm *NullUint `json:"submitterConfirm" form:"submitterConfirm"`
}
