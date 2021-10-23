package ms

type CurrentUser struct {
	UserId          uint   `json:"userId"`
	RoleId          uint   `json:"roleId"`
	RoleSort        uint   `json:"roleSort"`
	RoleKeyword     string `json:"roleKeyword"`
	PathRoleId      uint   `json:"pathRoleId"`
	PathRoleKeyword string `json:"pathRoleKeyword"`
}
