package req

import "github.com/piupuer/go-helper/pkg/resp"

type Message struct {
	ToUserId uint   `json:"toUserId"`
	Title    string `json:"title" form:"title"`
	Content  string `json:"content" form:"content"`
	Type     *uint  `json:"type" form:"type"`
	Status   *uint  `json:"status" form:"status"`
	resp.Page
}

type PushMessage struct {
	FromUserId       uint
	Type             *NullUint `json:"type" form:"type" validate:"required"`
	ToUserIds        []uint    `json:"toUserIds" form:"toUserIds"`
	ToRoleIds        []uint    `json:"toRoleIds" form:"toRoleIds"`
	Title            string    `json:"title" form:"title" validate:"required"`
	Content          string    `json:"content" form:"content" validate:"required"`
	IdempotenceToken string    `json:"idempotenceToken" form:"idempotenceToken"`
}

func (s PushMessage) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Type"] = "type"
	m["Title"] = "title"
	m["Content"] = "content"
	return m
}

type MessageWs struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
