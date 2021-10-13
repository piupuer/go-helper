package response

type ApprovalLogResp struct {
	End             bool `json:"end"`             // 是否已结束
	WaitingConfirm  bool `json:"waitingConfirm"`  // 待确认
	WaitingResubmit bool `json:"waitingResubmit"` // 待重新提交
	Cancel          bool `json:"cancel"`          // 已取消
}

type EventResp struct {
	Id         uint   `json:"id"`
	Sort       uint   `json:"sort"`
	Level      uint   `json:"level"`
	Edit       uint   `json:"edit"`
	EditFields string `json:"editFields"`
	Refuse     uint   `json:"refuse"`
	Roles      []uint `json:"roles"`
	Users      []uint `json:"users"`
}

type LogTrackResp struct {
	Name    string `json:"name"`
	Opinion string `json:"opinion"`
	End     bool   `json:"end"`
	Cancel  bool   `json:"cancel"`
}

type MachineResp struct {
	Name                       string      `json:"name"`
	SubmitterName              string      `json:"submitterName"`
	SubmitterEditFields        string      `json:"submitterEditFields"`
	SubmitterConfirm           uint        `json:"submitterConfirm"`
	SubmitterConfirmEditFields string      `json:"submitterConfirmEditFields"`
	EventsJson                 string      `json:"eventsJson"`
	Events                     []EventResp `json:"events"`
}
