package fsm

var (
	ErrDbEmpty                  = "go-helper.fsm.error.db"
	ErrTransitionEmpty          = "go-helper.fsm.error.transition"
	ErrLevelsEmpty              = "go-helper.fsm.error.levels"
	ErrDuplicateLevelName       = "go-helper.fsm.error.level-name"
	ErrLevelNameEmpty           = "go-helper.fsm.error.duplicate-level-name"
	ErrRepeatSubmit             = "go-helper.fsm.error.repeat-submit"
	ErrParams                   = "go-helper.fsm.error.illegal-param"
	ErrNoPermissionApprove      = "go-helper.fsm.error.no-permission-approve"
	ErrNoPermissionOrEnded      = "go-helper.fsm.error.no-permission-or-ended"
	ErrNoPermissionEdit         = "go-helper.fsm.error.no-permission-edit"
	ErrOnlySubmitterCancel      = "go-helper.fsm.error.only-submitter-can-cancel"
	ErrStartedCannotCancel      = "go-helper.fsm.error.started-cannot-cancel"
	ErrMachineNotFound          = "go-helper.fsm.error.machine-not-found"
	ErrDuplicateMachineCategory = "go-helper.fsm.error.duplicate-machine-category"
)
