package fsm

import "fmt"

var (
	ErrDbNil                        = fmt.Errorf("db instance is empty")
	ErrTransitionNil                = fmt.Errorf("transition handler is empty")
	ErrEventsNil                    = fmt.Errorf("events is empty")
	ErrEventNameNil                 = fmt.Errorf("event name is empty")
	ErrEventEndPointNotUnique       = fmt.Errorf("duplicate level name")
	ErrRepeatSubmit                 = fmt.Errorf("repeat submit")
	ErrParams                       = fmt.Errorf("illegal param")
	ErrNoPermissionApprove          = fmt.Errorf("no permission to approve")
	ErrNoPermissionOrEnded          = fmt.Errorf("no permission to approve or approval ended")
	ErrNoEditLogDetailPermission    = fmt.Errorf("no permission to edit detail")
	ErrOnlySubmitterCancel          = fmt.Errorf("only the submitter can cancel")
	ErrStartedCannotCancel          = fmt.Errorf("the flow is started and cannot cancel")
	ErrMachineNotFound              = fmt.Errorf("config not found")
	ErrMachineCategoryAlreadyExists = fmt.Errorf("category already exists")
)
