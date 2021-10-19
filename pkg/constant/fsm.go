package constant

const (
	FsmPrefix = "tb_fsm_"
)

const (
	FsmLogStatusWaiting   uint = iota // pending approval
	FsmLogStatusApproved              // approved
	FsmLogStatusRefused               // approval rejection
	FsmLogStatusCancelled             // approval cancelled
)

const (
	FsmMsgSubmitterCancel = "submitter cancelled"
	FsmMsgEnded           = "process ended"
	FsmMsgConfigChanged   = "configuration changes"
)

const (
	FsmSuffixWaiting   = "waiting"
	FsmSuffixResubmit  = "resubmit"
	FsmSuffixSubmitted = "submitted"
	FsmSuffixApproved  = "approved"
	FsmSuffixRefused   = "refused"
	FsmSuffixConfirm   = "confirm"
	FsmSuffixConfirmed = "confirmed"
)
