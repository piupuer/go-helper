package query

import (
	"github.com/piupuer/go-helper/pkg/fsm"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
)

// FindFsm find finite state machine
func (my MySql) FindFsm(r *req.FsmMachine) []resp.FsmMachine {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsm"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	return f.FindMachine(r)
}

// FindFsmApprovingLog find waiting approve log
func (my MySql) FindFsmApprovingLog(r *req.FsmPendingLog) []resp.FsmApprovingLog {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsmApprovingLog"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	return f.FindPendingLogByApprover(r)
}

// FindFsmLogTrack find approve log
func (my MySql) FindFsmLogTrack(r req.FsmLog) []resp.FsmLogTrack {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindFsmLogTrack"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	logs := f.FindLog(r)
	return f.FindLogTrack(logs)
}

// FsmSubmitLog submit log
func (my MySql) FsmSubmitLog(r req.FsmCreateLog) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmSubmitLog"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.SubmitLog(r)
	return f.Error
}

// FsmApproveLog approve log
func (my MySql) FsmApproveLog(r req.FsmApproveLog) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmApproveLog"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.ApproveLog(r)
	return f.Error
}

// FsmCancelLogByUuids cancel finite state machine log by uuids
func (my MySql) FsmCancelLogByUuids(r req.FsmCancelLog) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmCancelLogByUuids"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.CancelLogByUuids(r)
	return f.Error
}

// FsmCheckEditLogDetailPermission check edit log detail permission
func (my MySql) FsmCheckEditLogDetailPermission(r req.FsmCheckEditLogDetailPermission) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FsmCheckEditLogDetailPermission"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.CheckEditLogDetailPermission(r)
	return f.Error
}

// CreateFsm create finite state machine
func (my MySql) CreateFsm(r req.FsmCreateMachine) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateFsm"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.CreateMachine(r)
	return f.Error
}

// UpdateFsmById update finite state machine
func (my MySql) UpdateFsmById(id uint, r req.FsmUpdateMachine) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateFsmById"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.UpdateMachineById(id, r)
	return f.Error
}

// DeleteFsmByIds delete finite state machine
func (my MySql) DeleteFsmByIds(ids []uint) error {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteFsmByIds"))
	defer span.End()
	my.ops.fsmOps = append(my.ops.fsmOps, fsm.WithCtx(my.Ctx), fsm.WithDb(my.Tx))
	f := fsm.New(my.ops.fsmOps...)
	f.DeleteMachineByIds(ids)
	return f.Error
}
