package query

import (
	"github.com/piupuer/go-helper/pkg/fsm"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/pkg/errors"
)

// find finite state machine
func (my MySql) FindFsm(r *req.FsmMachine) ([]resp.FsmMachine, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsm"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	return f.FindMachine(r)
}

// find waiting approve log
func (my MySql) FindFsmApprovingLog(r *req.FsmPendingLog) ([]resp.FsmApprovingLog, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsmApprovingLog"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	return f.FindPendingLogByApprover(r)
}

// find approve log
func (my MySql) FindFsmLogTrack(r req.FsmLog) ([]resp.FsmLogTrack, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsmLogTrack"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	logs, err := f.FindLog(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return f.FindLogTrack(logs)
}

// find waiting approve log
func (my MySql) FsmApproveLog(r req.FsmApproveLog) (*resp.FsmApprovalLog, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmApproveLog"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
		fsm.WithTransition(my.ops.fsmTransition),
	)
	return f.ApproveLog(r)
}

// cancel finite state machine log by uuids
func (my MySql) FsmCancelLogByUuids(r req.FsmCancelLog) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmCancelLogByUuids"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
		fsm.WithTransition(my.ops.fsmTransition),
	)
	return f.CancelLogByUuids(r)
}

// check edit log detail permission
func (my MySql) FsmCheckEditLogDetailPermission(r req.FsmCheckEditLogDetailPermission) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmCheckEditLogDetailPermission"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	return f.CheckEditLogDetailPermission(r)
}

// create finite state machine
func (my MySql) CreateFsm(r req.FsmCreateMachine) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateFsm"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	_, err := f.CreateMachine(r)
	return err
}

// update finite state machine
func (my MySql) UpdateFsmById(id uint, r req.FsmUpdateMachine) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateFsmById"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
		fsm.WithTransition(my.ops.fsmTransition),
	)
	_, err := f.UpdateMachineById(id, r)
	return err
}

// delete finite state machine
func (my MySql) DeleteFsmByIds(ids []uint) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteFsmByIds"))
	defer span.End()
	f := fsm.New(
		fsm.WithCtx(my.Ctx),
		fsm.WithDb(my.Tx),
	)
	return f.DeleteMachineByIds(ids)
}
