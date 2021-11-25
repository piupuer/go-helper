package fsm

import "fmt"

var (
	ErrDbNil                     = fmt.Errorf("db instance is empty")
	ErrTransitionNil             = fmt.Errorf("transition handler is empty")
	ErrEventsNil                 = fmt.Errorf("events is empty")
	ErrEventNameNil              = fmt.Errorf("event name is empty")
	ErrEventEndPointNotUnique    = fmt.Errorf("event end position is not unique or has no end position")
	ErrRepeatSubmit              = fmt.Errorf("approval record already exists")
	ErrParams                    = fmt.Errorf("illegal param")
	ErrNoPermissionApprove       = fmt.Errorf("no permission to pass the approval")
	ErrNoPermissionOrEnded       = fmt.Errorf("no permission to approve or approval ended")
	ErrNoEditLogDetailPermission = fmt.Errorf("no permission to edit log detail")
	ErrOnlySubmitterCancel       = fmt.Errorf("only the submitter can cancel")
	ErrStartedCannotCancel       = fmt.Errorf("the process is already in progress and cannot be cancelled halfway")
)
