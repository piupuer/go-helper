package resp

type MessageResp struct {
	Base
	Status     uint   `json:"status"`
	ToUserId   uint   `json:"toUserId"`
	Type       uint   `json:"type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	FromUserId uint   `json:"fromUserId"`
}

type MessageWsResp struct {
	Type   string `json:"type"`
	Detail Resp   `json:"detail"`
}
