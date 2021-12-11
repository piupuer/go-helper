package resp

type UserStatus struct {
	Captcha string `json:"captcha"`
	Locked  uint   `json:"locked"`
}
