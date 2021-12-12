package resp

type UserStatus struct {
	Captcha Captcha `json:"captcha"`
	Locked  uint    `json:"locked"`
}

type User struct {
	Base
	Username    string `json:"username"`
	Mobile      string `json:"mobile"`
	Nickname    string `json:"nickname"`
	RoleId      uint   `json:"roleId"`
	RoleSort    uint   `json:"roleSort"`
	RoleKeyword string `json:"roleKeyword"`
}

type Role struct {
	Base
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
}

type Captcha struct {
	Id  string `json:"id"`
	Img string `json:"img"`
}
