package response

import "github.com/piupuer/go-helper/models"

type ApprovalLogResp struct {
	End             bool `json:"end"`             // 是否已结束
	WaitingConfirm  bool `json:"waitingConfirm"`  // 待确认
	WaitingResubmit bool `json:"waitingResubmit"` // 待重新提交
	Cancel          bool `json:"cancel"`          // 已取消
}

type LogTrackResp struct {
	Name    string `json:"name"`
	Opinion string `json:"opinion"`
	End     bool   `json:"end"`
	Cancel  bool   `json:"cancel"`
}

type MachineResp struct {
	models.P
	Name                       string `json:"name"`
	SubmitterName              string `json:"submitterName"`
	SubmitterEditFields        string `json:"submitterEditFields"`
	SubmitterConfirm           uint   `json:"submitterConfirm"`
	SubmitterConfirmEditFields string `json:"submitterConfirmEditFields"`
	EventsJson                 string `json:"eventsJson"`
}
