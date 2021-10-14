package fsm

import "fmt"

const (
	LogStatusWaiting   uint = iota // pending approval
	LogStatusApproved              // approved
	LogStatusRefused               // approval rejection
	LogStatusCancelled             // approval cancelled
)

var (
	ErrDbNil                  = fmt.Errorf("db instance is empty")
	ErrEventsNil              = fmt.Errorf("events is empty")
	ErrEventNameNil           = fmt.Errorf("event name is empty")
	ErrEventEndPointNotUnique = fmt.Errorf("event end position is not unique or has no end position")
	ErrRepeatSubmit           = fmt.Errorf("approval record already exists")
	ErrStatus                 = fmt.Errorf("illegal approval status")
	ErrNoPermissionApprove    = fmt.Errorf("no permission to pass the approval")
	ErrNoPermissionRefuse     = fmt.Errorf("no permission to refuse approval")
	ErrNoPermissionOrEnded    = fmt.Errorf("no permission to approve or approval ended")
	ErrOnlySubmitterCancel    = fmt.Errorf("only the submitter can cancel")
	ErrStartedCannotCancel    = fmt.Errorf("the process is already in progress and cannot be cancelled halfway")
)

const (
	MsgSubmitterCancel = "submitter cancelled"
	MsgEnded           = "process ended"
	MsgConfigChanged   = "configuration changes"
)

const (
	SuffixWaiting   = "waiting"
	SuffixResubmit  = "resubmit"
	SuffixSubmitted = "submitted"
	SuffixApproved  = "approved"
	SuffixRefused   = "refused"
	SuffixConfirm   = "confirm"
	SuffixConfirmed = "confirmed"
)
