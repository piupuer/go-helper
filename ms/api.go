package ms

type User struct {
	M
	Username        string `json:"username"`
	Mobile          string `json:"mobile"`
	Nickname        string `json:"nickname"`
	RoleId          uint   `json:"roleId"`
	RoleName        string `json:"roleName"`
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

type SignUser struct {
	M
	AppId     string      `json:"appId"`
	AppSecret string      `json:"appSecret"`
	Scopes    []SignScope `json:"scopes"`
	Status    uint        `json:"status"`
}

type SignScope struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
