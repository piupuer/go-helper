package query

import (
	"github.com/piupuer/go-helper/pkg/fsm"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

// find finite state machine
func (my MySql) FindFsm(r *req.FsmMachine) ([]resp.FsmMachine, error) {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
	)
	return f.FindMachine(r)
}

// find waiting approve log
func (my MySql) FindFsmApprovingLog(r *req.FsmPendingLog) ([]resp.FsmApprovingLog, error) {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
	)
	return f.FindPendingLogByApprover(r)
}

// find waiting approve log
func (my MySql) FsmApproveLog(r req.FsmApproveLog) (*resp.FsmApprovalLog, error) {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
		fsm.WithTransition(my.ops.fsmTransition),
	)
	return f.ApproveLog(r)
}

func (my MySql) FsmCheckEditLogDetailPermission(r req.FsmCheckEditLogDetailPermission) error {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
	)
	return f.CheckEditLogDetailPermission(r)
}

// create finite state machine
func (my MySql) CreateFsm(r req.FsmCreateMachine) error {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
	)
	_, err := f.CreateMachine(r)
	return err
}

// update finite state machine
func (my MySql) UpdateFsmById(id uint, r req.FsmUpdateMachine) error {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
		fsm.WithTransition(my.ops.fsmTransition),
	)
	_, err := f.UpdateMachineById(id, r)
	return err
}

// delete finite state machine
func (my MySql) DeleteFsmByIds(ids []uint) error {
	f := fsm.New(
		my.Tx,
		fsm.WithLogger(my.ops.logger),
		fsm.WithCtx(my.Ctx),
	)
	return f.DeleteMachineByIds(ids)
}
