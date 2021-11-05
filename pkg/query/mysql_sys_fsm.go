package query

import (
	"github.com/piupuer/go-helper/pkg/fsm"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
)

// find finite state machine
func (my MySql) FindFsm(r *req.FsmMachine) ([]resp.FsmMachine, error) {
	f := fsm.New(my.Tx)
	return f.FindMachine(r)
}

// find waiting approve log
func (my MySql) FindFsmApprovingLog(r *req.FsmPendingLog) ([]resp.FsmApprovingLog, error) {
	f := fsm.New(my.Tx)
	return f.FindPendingLogByApprover(r)
}

// create finite state machine
func (my MySql) CreateFsm(r req.FsmCreateMachine) error {
	f := fsm.New(my.Tx)
	_, err := f.CreateMachine(r)
	return err
}

// update finite state machine
func (my MySql) UpdateFsmById(id uint, r req.FsmUpdateMachine) error {
	f := fsm.New(my.Tx)
	_, err := f.UpdateMachineById(id, r)
	return err
}

// delete finite state machine
func (my MySql) DeleteFsmByIds(ids []uint) error {
	f := fsm.New(my.Tx)
	return f.DeleteMachineByIds(ids)
}
