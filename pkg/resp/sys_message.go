package resp

type Message struct {
	Base
	Status       uint   `json:"status"`
	ToUserId     uint   `json:"toUserId"`
	ToUsername   string `json:"toUsername"`
	ToNickname   string `json:"toNickname"`
	Type         uint   `json:"type"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	FromUserId   uint   `json:"fromUserId"`
	FromUsername string `json:"fromUsername"`
	FromNickname string `json:"fromNickname"`
}

type MessageWs struct {
	Type   string `json:"type"`
	Detail Resp   `json:"detail"`
}
