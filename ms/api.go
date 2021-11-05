package ms

type User struct {
	M
	Username        string `json:"username"`
	Mobile          string `json:"mobile"`
	Nickname        string `json:"nickname"`
	RoleId          uint   `json:"roleId"`
	RoleSort        uint   `json:"roleSort"`
	RoleKeyword     string `json:"roleKeyword"`
	PathRoleId      uint   `json:"pathRoleId"`
	PathRoleKeyword string `json:"pathRoleKeyword"`
}

type Role struct {
	M
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
}
