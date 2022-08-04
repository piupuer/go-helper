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
	FsmMsgSubmitterCancel = "go-helper.fsm.msg.submitter-cancel"
	FsmMsgEnded           = "go-helper.fsm.msg.ended"
	FsmMsgConfigChanged   = "go-helper.fsm.msg.config-changed"
	FsmMsgManualCancel    = "go-helper.fsm.msg.manual-cancel"
)

const (
	FsmSuffixWaiting   = "go-helper.fsm.suffix.waiting"
	FsmSuffixResubmit  = "go-helper.fsm.suffix.resubmit"
	FsmSuffixSubmitted = "go-helper.fsm.suffix.submitted"
	FsmSuffixApproved  = "go-helper.fsm.suffix.approved"
	FsmSuffixRefused   = "go-helper.fsm.suffix.refused"
	FsmSuffixConfirm   = "go-helper.fsm.suffix.confirm"
	FsmSuffixConfirmed = "go-helper.fsm.suffix.confirmed"
)
