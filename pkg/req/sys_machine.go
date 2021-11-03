package req

import "github.com/piupuer/go-helper/pkg/resp"

type Machine struct {
	Id        uint   `json:"id" form:"id"`
	Host      string `json:"host" form:"host"`
	SshPort   int    `json:"sshPort" form:"sshPort"`
	Version   string `json:"version" form:"version"`
	Name      string `json:"name" form:"name"`
	Arch      string `json:"arch" form:"arch"`
	Cpu       string `json:"cpu" form:"cpu"`
	Memory    string `json:"memory" form:"memory"`
	Disk      string `json:"disk" form:"disk"`
	LoginName string `json:"loginName" form:"loginName"`
	LoginPwd  string `json:"loginPwd" form:"loginPwd"`
	Status    *uint  `json:"status" form:"status"`
	Remark    string `json:"remark" form:"remark"`
	resp.Page
}

type CreateMachine struct {
	Host      string   `json:"host" validate:"required"`
	SshPort   NullUint `json:"sshPort" validate:"required"`
	Version   string   `json:"version"`
	Name      string   `json:"name"`
	Arch      string   `json:"arch"`
	Cpu       string   `json:"cpu"`
	Memory    string   `json:"memory"`
	Disk      string   `json:"disk"`
	LoginName string   `json:"loginName" validate:"required"`
	LoginPwd  string   `json:"loginPwd" validate:"required"`
	Status    NullUint `json:"status"`
	Remark    string   `json:"remark"`
}

type MachineShellWs struct {
	Host      string   `json:"host" form:"host"`
	SshPort   NullUint `json:"sshPort" form:"sshPort"`
	LoginName string   `json:"loginName" form:"loginName"`
	LoginPwd  string   `json:"loginPwd" form:"loginPwd"`
	InitCmd   string   `json:"initCmd" form:"initCmd"`
	Cols      NullUint `json:"cols" form:"cols"`
	Rows      NullUint `json:"rows" form:"rows"`
}

func (s CreateMachine) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Host"] = "host ip/address"
	m["SshPort"] = "ssh port"
	m["LoginName"] = "login name"
	m["LoginPwd"] = "login password"
	return m
}

type UpdateMachine struct {
	Host      *string   `json:"host"`
	SshPort   *NullUint `json:"sshPort"`
	Version   *string   `json:"version"`
	Name      *string   `json:"name"`
	Arch      *string   `json:"arch"`
	Cpu       *string   `json:"cpu"`
	Memory    *string   `json:"memory"`
	Disk      *string   `json:"disk"`
	LoginName *string   `json:"loginName"`
	LoginPwd  *string   `json:"loginPwd"`
	Status    *NullUint `json:"status"`
	Remark    *string   `json:"remark"`
}
