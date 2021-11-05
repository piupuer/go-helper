package resp

import "github.com/golang-module/carbon"

type FsmApprovalLog struct {
	Uuid     string `json:"uuid"`
	Category uint   `json:"category"`
	End      bool   `json:"end"`      // is ended?
	Confirm  bool   `json:"confirm"`  // is waiting submitter confirm?
	Resubmit bool   `json:"resubmit"` // is waiting submitter resubmit?
	Cancel   bool   `json:"cancel"`   // is submitter canceled?
}

type FsmApprovingLog struct {
	Base
	Uuid             string `json:"uuid"`
	Category         uint   `json:"category"`
	SubmitterRoleId  uint   `json:"submitterRoleId"`
	SubmitterRole    Role   `json:"submitterRole"`
	SubmitterUserId  uint   `json:"submitterUserId"`
	SubmitterUser    User   `json:"submitterUser"`
	PrevDetail       string `json:"prevDetail"`
	Detail           string `json:"detail"`
	CanApprovalRoles []Role `json:"canApprovalRoles"`
	CanApprovalUsers []User `json:"canApprovalUsers"`
}

type FsmLogTrack struct {
	CreatedAt carbon.ToDateTimeString `json:"createdAt"`
	UpdatedAt carbon.ToDateTimeString `json:"updatedAt"`
	Name      string                  `json:"name"`
	Opinion   string                  `json:"opinion"`
	Status    uint                    `json:"status"`
	End       bool                    `json:"end"`
	Cancel    bool                    `json:"cancel"`
	Resubmit  bool                    `json:"resubmit"`
	Confirm   bool                    `json:"confirm"`
}

type FsmMachine struct {
	Base
	Category                   uint   `json:"category"`
	Name                       string `json:"name"`
	SubmitterName              string `json:"submitterName"`
	SubmitterEditFields        string `json:"submitterEditFields"`
	SubmitterConfirm           uint   `json:"submitterConfirm"`
	SubmitterConfirmEditFields string `json:"submitterConfirmEditFields"`
	EventsJson                 string `json:"eventsJson"`
}
