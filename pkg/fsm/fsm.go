package fsm

import (
	"fmt"
	"github.com/looplab/fsm"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

type Fsm struct {
	ops     Options
	session *gorm.DB
	Error   error
}

// Migrate mysql DDL migrate rollback is not supported, Migrate before New
func Migrate(options ...func(*Options)) (err error) {
	fs := New(options...)
	if fs.Error != nil {
		return
	}
	session := fs.initSession()
	err = session.AutoMigrate(
		new(Machine),
		new(Event),
		new(User),
		new(EventSrcItemRelation),
		new(EventUserRelation),
		new(EventItem),
		new(Log),
		new(LogApprovalUserRelation),
	)
	return
}

func New(options ...func(*Options)) (fs *Fsm) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	fs = &Fsm{
		ops: *ops,
	}
	if ops.db != nil {
		fs.session = fs.initSession()
	} else {
		fs.Error = ErrDbNil
	}
	return
}

func (fs *Fsm) DeleteMachineByIds(ids []uint) {
	if fs.Error != nil {
		return
	}
	if len(ids) == 0 {
		return
	}
	machines := make([]Machine, 0)
	fs.session.
		Model(&Machine{}).
		Where("id IN (?)", ids).
		Find(&machines)
	categories := make([]uint, 0)
	for _, item := range machines {
		categories = append(categories, item.Category)
	}
	for _, item := range categories {
		fs.CancelLog(item)
	}
	fs.session.
		Where("id IN (?)", ids).
		Delete(&Machine{})
	return
}

func (fs *Fsm) CreateMachine(r req.FsmCreateMachine) (rp Machine) {
	if fs.Error != nil {
		return
	}
	var machine Machine
	utils.Struct2StructByJson(r, &machine)
	// category is unique
	var count int64
	fs.session.
		Model(&machine).
		Where("category = ?", machine.Category).
		Count(&count)
	if count > 0 {
		fs.AddError(ErrMachineCategoryAlreadyExists)
		return
	}
	// save json for query
	machine.EventsJson = utils.Struct2Json(r.Levels)
	fs.session.Create(&machine)
	// batch fsm event
	fs.batchCreateEvent(machine.Id, r.Levels)
	if fs.Error != nil {
		return
	}
	fs.findEventDesc(machine.Id)
	if fs.Error != nil {
		return
	}
	rp = machine
	return
}

func (fs *Fsm) UpdateMachineById(id uint, r req.FsmUpdateMachine) (rp Machine) {
	if fs.Error != nil {
		return
	}
	var oldMachine Machine
	fs.session.
		Model(&Machine{}).
		Where("id = ?", id).
		First(&oldMachine)
	// cancel log when machine config change
	fs.CancelLog(oldMachine.Category)
	if fs.Error != nil {
		return
	}
	levels := make([]req.FsmCreateEvent, len(r.Levels))
	copy(levels, r.Levels)
	r.Levels = make([]req.FsmCreateEvent, 0)
	eventsJson := utils.Struct2Json(levels)
	m := make(map[string]interface{}, 0)
	utils.CompareDiff2SnakeKey(oldMachine, r, &m)
	if oldMachine.EventsJson != eventsJson {
		m["events_json"] = eventsJson
	}
	fs.session.
		Model(&Machine{}).
		Where("id = ?", id).
		Updates(&m)
	// batch fsm event
	fs.batchCreateEvent(oldMachine.Id, levels)
	if fs.Error != nil {
		return
	}
	fs.findEventDesc(oldMachine.Id)
	if fs.Error != nil {
		return
	}
	rp = oldMachine
	return
}

// SubmitLog
// =======================================================
// approval log function
// =======================================================
// first submit log
func (fs *Fsm) SubmitLog(r req.FsmCreateLog) (rp []EventItem) {
	rp = make([]EventItem, 0)
	if fs.Error != nil {
		return
	}
	machine := fs.GetMachineByCategory(uint(r.Category))
	if machine.Id == constant.Zero {
		fs.AddError(ErrMachineNotFound)
		return
	}
	// check whether approval is pending
	last := fs.getLastPendingLog(req.FsmLog{
		Category: r.Category,
		Uuid:     r.Uuid,
	})
	if last.Id > constant.Zero {
		fs.AddError(ErrRepeatSubmit)
		return
	}
	startEvent := fs.getStartEvent(machine.Id)
	if fs.Error != nil {
		return
	}

	// first create log
	var l Log
	l.Category = uint(r.Category)
	l.Uuid = r.Uuid
	nextEvent := fs.getNextEvent(machine.Id, startEvent.Level)
	if fs.Error != nil {
		return
	}
	l.ProgressId = startEvent.DstId
	l.CanApprovalRoles = nextEvent.Roles
	l.CanApprovalUsers = nextEvent.Users
	l.SubmitterRoleId = r.SubmitterRoleId
	l.SubmitterUserId = r.SubmitterUserId
	l.PrevDetail = startEvent.Dst.Name
	l.Detail = nextEvent.Name.Name
	l.CurrentEventId = startEvent.Id
	l.NextEventId = nextEvent.Id
	fs.session.Create(&l)

	rp = append(rp, []EventItem{
		startEvent.Dst,
		nextEvent.Name,
	}...)
	return
}

// ApproveLog start approve log
func (fs *Fsm) ApproveLog(r req.FsmApproveLog) (rp resp.FsmApprovalLog) {
	if fs.Error != nil {
		return
	}
	machine := fs.GetMachineByCategory(uint(r.Category))
	if fs.Error != nil {
		return
	}
	approved := uint(r.Approved)
	rp = resp.FsmApprovalLog{
		Uuid:     r.Uuid,
		Category: uint(r.Category),
	}
	// check current user/role permission
	oldLog := fs.CheckLogPermission(req.FsmPermissionLog{
		Category:       r.Category,
		Uuid:           r.Uuid,
		ApprovalRoleId: r.ApprovalRoleId,
		ApprovalUserId: r.ApprovalUserId,
		Approved:       approved,
	})
	if fs.Error != nil {
		return
	}

	// submitter cancel
	if approved == constant.FsmLogStatusCancelled {
		m := make(map[string]interface{}, 0)
		m["approved"] = constant.FsmLogStatusCancelled
		m["approval_role_id"] = r.ApprovalRoleId
		m["approval_user_id"] = r.ApprovalUserId
		m["approval_opinion"] = r.ApprovalOpinion
		m["next_event_id"] = constant.Zero
		m["detail"] = constant.FsmMsgSubmitterCancel
		rp.Cancel = constant.One
		fs.session.
			Model(&Log{}).
			Where("id = ?", oldLog.Id).
			Updates(&m)
		return
	}

	desc := fs.findEventDesc(machine.Id)
	if fs.Error != nil {
		return
	}
	// create fsm instance from current progress
	f := fsm.NewFSM(oldLog.Progress.Name, desc, nil)

	transitions := f.AvailableTransitions()
	eventName := ""
	for _, transition := range transitions {
		match := false
		switch approved {
		case constant.FsmLogStatusApproved:
			if strings.HasSuffix(transition, constant.FsmSuffixWaiting) || strings.HasSuffix(transition, constant.FsmSuffixResubmit) || strings.HasSuffix(transition, constant.FsmSuffixConfirm) {
				match = true
			}
		case constant.FsmLogStatusRefused:
			if strings.HasSuffix(transition, constant.FsmSuffixWaiting) {
				match = true
			}
		}
		if match {
			eventName = transition
			break
		}
	}

	if eventName == "" {
		fs.AddError(errors.Wrap(ErrParams, "approved"))
		return
	}
	nextName := getNextItemName(approved, eventName)
	f.SetState(nextName)
	event := fs.getEvent(machine.Id, eventName)
	if fs.Error != nil {
		return
	}
	progressItem := fs.getEventItemByName(nextName)
	if progressItem.Id == constant.Zero {
		fs.AddError(gorm.ErrRecordNotFound)
		return
	}
	var newLog Log
	newLog.Category = uint(r.Category)
	newLog.Uuid = r.Uuid
	newLog.SubmitterRoleId = oldLog.SubmitterRoleId
	newLog.SubmitterUserId = oldLog.SubmitterUserId
	newLog.PrevDetail = nextName
	newLog.CurrentEventId = event.Id
	if len(f.AvailableTransitions()) != 0 {
		// bind next approver
		var nextEvent Event
		if approved == constant.FsmLogStatusApproved {
			nextEvent = fs.getNextEvent(machine.Id, event.Level)
		} else {
			nextEvent = fs.getPrevEvent(machine.Id, event.Level)
		}
		if fs.Error != nil {
			return
		}
		// no users/roles, maybe submitter resubmit/confirm
		noUser := false
		if len(nextEvent.Roles) == 0 && len(nextEvent.Users) == 0 {
			noUser = true
			if strings.HasSuffix(nextEvent.Name.Name, constant.FsmSuffixConfirm) {
				rp.Confirm = constant.One
			} else {
				rp.Resubmit = constant.One
			}
		}
		newLog.ProgressId = progressItem.Id
		newLog.NextEventId = nextEvent.Id
		if rp.Resubmit == constant.One {
			newLog.Resubmit = constant.One
		}
		if rp.Confirm == constant.One {
			newLog.Confirm = constant.One
		}
		if noUser {
			newLog.CanApprovalRoles = []Role{
				{
					Id: oldLog.SubmitterRoleId,
				},
			}
			newLog.CanApprovalUsers = []User{
				{
					Id: oldLog.SubmitterUserId,
				},
			}
		} else {
			newLog.CanApprovalRoles = nextEvent.Roles
			newLog.CanApprovalUsers = nextEvent.Users
		}
		newLog.Detail = nextEvent.Name.Name
	} else {
		rp.End = constant.One
		newLog.Approved = constant.FsmLogStatusApproved
		newLog.Detail = constant.FsmMsgEnded
	}
	fs.session.Create(&newLog)
	m := make(map[string]interface{}, 0)
	m["approved"] = constant.FsmLogStatusApproved
	if approved == constant.FsmLogStatusRefused {
		m["approved"] = constant.FsmLogStatusRefused
	}
	m["approval_role_id"] = r.ApprovalRoleId
	m["approval_user_id"] = r.ApprovalUserId
	m["approval_opinion"] = r.ApprovalOpinion
	// update oldLog approved
	fs.session.
		Model(&Log{}).
		Where("id = ?", oldLog.Id).
		Updates(&m)
	// status transition
	if fs.ops.transition == nil {
		log.WithContext(fs.ops.ctx).Warn("%s", ErrTransitionNil)
		return
	}
	fs.AddError(fs.ops.transition(fs.ops.ctx, rp))
	return
}

// CancelLog cancel log by category(it is applicable to the automatic cancellation of records to be approved when the approval configuration changes)
func (fs *Fsm) CancelLog(category uint) {
	if fs.Error != nil {
		return
	}
	m := make(map[string]interface{}, 0)
	m["approved"] = constant.FsmLogStatusCancelled
	m["next_event_id"] = constant.Zero
	m["detail"] = constant.FsmMsgConfigChanged
	q := fs.session.
		Model(&Log{}).
		Where("category = ?", category).
		Where("approved = ?", constant.FsmLogStatusWaiting)
	oldLogs := make([]Log, 0)
	q.Find(&oldLogs)
	list := make([]resp.FsmApprovalLog, 0)
	for i, l := 0, len(oldLogs); i < l; i++ {
		list = append(list, resp.FsmApprovalLog{
			Uuid:     oldLogs[i].Uuid,
			Category: oldLogs[i].Category,
			Cancel:   constant.One,
		})
	}
	q.Updates(&m)
	if fs.ops.transition == nil {
		log.WithContext(fs.ops.ctx).Warn("%s", ErrTransitionNil)
		return
	}
	// status transition
	fs.AddError(fs.ops.transition(fs.ops.ctx, list...))
	return
}

func (fs *Fsm) CancelLogByUuids(r req.FsmCancelLog) {
	if fs.Error != nil {
		return
	}
	if len(r.Uuids) == 0 {
		fs.AddError(errors.Wrap(ErrParams, "uuids"))
		return
	}
	m := make(map[string]interface{}, 0)
	m["approved"] = constant.FsmLogStatusCancelled
	m["approval_role_id"] = r.ApprovalRoleId
	m["approval_user_id"] = r.ApprovalUserId
	m["next_event_id"] = constant.Zero
	m["detail"] = constant.FsmMsgManualCancel
	q := fs.session.
		Model(&Log{}).
		Where("uuid IN (?)", r.Uuids).
		Where("approved = ?", constant.FsmLogStatusWaiting)
	oldLogs := make([]Log, 0)
	q.Find(&oldLogs)
	if len(oldLogs) == 0 {
		fs.AddError(ErrNoPermissionOrEnded)
		return
	}
	list := make([]resp.FsmApprovalLog, 0)
	for i, l := 0, len(oldLogs); i < l; i++ {
		list = append(list, resp.FsmApprovalLog{
			Uuid:     oldLogs[i].Uuid,
			Category: oldLogs[i].Category,
			Cancel:   constant.One,
		})
	}
	q.Updates(&m)
	// status transition
	if fs.ops.transition == nil {
		log.WithContext(fs.ops.ctx).Warn("%s", ErrTransitionNil)
		return
	}
	fs.AddError(fs.ops.transition(fs.ops.ctx, list...))
	return
}

// CheckLogPermission
// =======================================================
// query function
// =======================================================
// check verify whether the current user/role has permission to approve
func (fs *Fsm) CheckLogPermission(r req.FsmPermissionLog) (rp Log) {
	if fs.Error != nil {
		return
	}
	// check whether approval is pending
	last := fs.getLastPendingLog(req.FsmLog{
		Category: r.Category,
		Uuid:     r.Uuid,
	})
	if last.Id == constant.Zero {
		fs.AddError(ErrNoPermissionOrEnded)
		return
	}
	if r.Approved == constant.FsmLogStatusCancelled {
		if last.SubmitterRoleId != r.ApprovalRoleId && last.SubmitterUserId != r.ApprovalUserId {
			fs.AddError(ErrOnlySubmitterCancel)
			return
		} else {
			if last.CurrentEvent.Level > constant.Zero {
				fs.AddError(ErrStartedCannotCancel)
				return
			}
			rp = last
			return
		}
	}
	roles := make([]uint, 0)
	users := make([]uint, 0)
	for _, role := range last.CanApprovalRoles {
		roles = append(roles, role.Id)
	}
	for _, user := range last.CanApprovalUsers {
		users = append(users, user.Id)
	}
	if !utils.Contains(roles, r.ApprovalRoleId) && !utils.Contains(users, r.ApprovalUserId) {
		fs.AddError(ErrNoPermissionApprove)
		return
	}
	rp = last
	return
}

// CheckEditLogDetailPermission check verify whether the current user/role has permission to edit log detail
func (fs *Fsm) CheckEditLogDetailPermission(r req.FsmCheckEditLogDetailPermission) {
	if fs.Error != nil {
		return
	}
	// check whether approval is pending
	last := fs.getLastPendingLog(req.FsmLog{
		Category: r.Category,
		Uuid:     r.Uuid,
	})
	if fs.Error != nil {
		fs.AddError(ErrNoEditLogDetailPermission)
		return
	}
	submitter := false
	confirm := false
	if last.SubmitterRoleId == r.ApprovalRoleId && last.SubmitterUserId == r.ApprovalUserId {
		submitter = true
	}
	if r.Approver {
		submitter = false
	}
	if submitter && last.NextEventId == 0 {
		confirm = true
	}
	edit := false
	editFields := ""
	if submitter || confirm {
		var machine Machine
		machine = fs.GetMachineByCategory(uint(r.Category))
		if fs.Error != nil {
			return
		}
		edit = true
		if submitter {
			editFields = machine.SubmitterEditFields
		} else {
			editFields = machine.SubmitterConfirmEditFields
		}
	} else {
		edit = last.NextEvent.Edit == constant.One
		editFields = last.NextEvent.EditFields
	}

	if !edit {
		fs.AddError(ErrNoEditLogDetailPermission)
		return
	}
	// split permission fields
	fields := strings.Split(utils.SnakeCase(editFields), ",")
	if len(fields) > 0 {
		for _, f := range r.Fields {
			if !utils.Contains(fields, utils.SnakeCase(f)) {
				fs.AddError(errors.Wrap(ErrNoEditLogDetailPermission, f))
				return
			}
		}
	}
	return
}

// GetMachineByCategory get machine by category
func (fs *Fsm) GetMachineByCategory(category uint) (machine Machine) {
	if fs.Error != nil {
		return
	}
	fs.session.
		Model(&Machine{}).
		Where("category = ?", category).
		First(&machine)
	return
}

func (fs *Fsm) FindMachine(r *req.FsmMachine) (rp []resp.FsmMachine) {
	rp = make([]resp.FsmMachine, 0)
	if fs.Error != nil {
		return
	}
	list := make([]Machine, 0)
	q := fs.session.Model(&Machine{})
	name := strings.TrimSpace(r.Name)
	if r.Category != nil {
		q.Where("category = ?", *r.Category)
	}
	if name != "" {
		q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	submitterName := strings.TrimSpace(r.SubmitterName)
	if submitterName != "" {
		q.Where("submitter_name LIKE ?", fmt.Sprintf("%%%s%%", submitterName))
	}
	if r.SubmitterConfirm != nil {
		q.Where("submitter_confirm = ?", *r.SubmitterConfirm)
	}
	page := &r.Page
	countCache := false
	if page.CountCache != nil {
		countCache = *page.CountCache
	}
	if !page.NoPagination {
		if !page.SkipCount {
			q.Count(&page.Total)
		}
		if page.Total > 0 || page.SkipCount {
			limit, offset := page.GetLimit()
			q.Limit(limit).Offset(offset).Find(&list)
		}
	} else {
		// no pagination
		q.Find(&list)
		page.Total = int64(len(list))
		page.GetLimit()
	}
	page.CountCache = &countCache
	utils.Struct2StructByJson(list, &rp)
	return
}

// FindLog find logs
func (fs *Fsm) FindLog(r req.FsmLog) (rp []Log) {
	rp = make([]Log, 0)
	if fs.Error != nil {
		return
	}
	fs.session.
		Preload("CurrentEvent").
		Preload("CurrentEvent.Roles").
		Preload("CurrentEvent.Users").
		Preload("NextEvent").
		Preload("NextEvent.Roles").
		Preload("NextEvent.Users").
		Preload("NextEvent.Name").
		Preload("CanApprovalRoles").
		Preload("CanApprovalUsers").
		Where("category = ?", r.Category).
		Where("uuid = ?", r.Uuid).
		Find(&rp)
	return
}

// FindLogTrack find log tracks
func (fs *Fsm) FindLogTrack(logs []Log) (rp []resp.FsmLogTrack) {
	rp = make([]resp.FsmLogTrack, 0)
	if fs.Error != nil {
		return
	}
	if len(logs) == 0 {
		return
	}
	l := len(logs)
	for i, item := range logs {
		prevApproved := constant.FsmLogStatusWaiting
		prevCancel := constant.Zero
		prevOpinion := ""
		end := constant.Zero
		cancel := constant.Zero
		if item.Approved == constant.FsmLogStatusCancelled {
			cancel = constant.One
		}
		if i > 0 {
			prevApproved = logs[i-1].Approved
			if logs[i-1].Approved == constant.FsmLogStatusCancelled {
				prevCancel = constant.One
			}
			prevOpinion = logs[i-1].ApprovalOpinion
		}
		if i == l-1 && item.NextEventId == constant.Zero {
			end = constant.One
		}
		if end == constant.One || cancel == constant.One {
			rp = append(rp, resp.FsmLogTrack{
				Time: resp.Time{
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.UpdatedAt,
				},
				Name:    item.PrevDetail,
				Opinion: prevOpinion,
				Status:  prevApproved,
				Cancel:  prevCancel,
			}, resp.FsmLogTrack{
				Time: resp.Time{
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.UpdatedAt,
				},
				Name:    item.Detail,
				Opinion: item.ApprovalOpinion,
				Status:  item.Approved,
				End:     end,
				Cancel:  cancel,
			})
		} else {
			rp = append(rp, resp.FsmLogTrack{
				Time: resp.Time{
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.UpdatedAt,
				},
				Name:    item.PrevDetail,
				Opinion: prevOpinion,
				Status:  prevApproved,
				End:     end,
				Cancel:  cancel,
			})
		}
		if i == l-1 && item.Approved == constant.FsmLogStatusWaiting {
			rp = append(rp, resp.FsmLogTrack{
				Name:     logs[i].Detail,
				Resubmit: item.Resubmit,
				Confirm:  item.Confirm,
			})
		}
	}
	return
}

// FindPendingLogByApprover get the pending approval list of a approver
func (fs *Fsm) FindPendingLogByApprover(r *req.FsmPendingLog) (rp []resp.FsmApprovingLog) {
	rp = make([]resp.FsmApprovingLog, 0)
	if fs.Error != nil {
		return
	}
	// get user relation
	logIds1 := make([]uint, 0)
	fs.session.
		Model(&LogApprovalUserRelation{}).
		Where("user_id = ?", r.ApprovalUserId).
		Pluck("log_id", &logIds1)
	// get role relation
	logIds2 := make([]uint, 0)
	fs.session.
		Model(&LogApprovalRoleRelation{}).
		Where("role_id = ?", r.ApprovalRoleId).
		Pluck("log_id", &logIds2)
	list := make([]Log, 0)
	ids := append(logIds1, logIds2...)
	if len(ids) > 0 {
		q := fs.session.
			Model(&Log{}).
			Preload("CanApprovalRoles").
			Preload("CanApprovalUsers").
			Where("approved = ?", constant.FsmLogStatusWaiting).
			Where("id IN (?)", ids)
		if uint(r.Category) > constant.Zero {
			q.Where("category = ?", r.Category)
		}
		page := &r.Page
		countCache := false
		if page.CountCache != nil {
			countCache = *page.CountCache
		}
		if !page.NoPagination {
			if !page.SkipCount {
				q.Count(&page.Total)
			}
			if page.Total > 0 || page.SkipCount {
				limit, offset := page.GetLimit()
				q.Limit(limit).Offset(offset).Find(&list)
			}
		} else {
			// no pagination
			q.Find(&list)
			page.Total = int64(len(list))
			page.GetLimit()
		}
		page.CountCache = &countCache
	}
	utils.Struct2StructByJson(list, &rp)
	return
}

// =======================================================
// private function
// =======================================================
// get last pending log, err will be returned when it does not exist
func (fs *Fsm) getLastPendingLog(r req.FsmLog) (rp Log) {
	fs.session.
		Preload("CanApprovalRoles").
		Preload("CanApprovalUsers").
		Preload("Progress").
		Preload("CurrentEvent").
		Preload("NextEvent").
		Where("category = ?", r.Category).
		Where("uuid = ?", r.Uuid).
		Where("approved = ?", constant.FsmLogStatusWaiting).
		First(&rp)
	return
}

func (fs *Fsm) getEvent(machineId uint, name string) (rp Event) {
	if fs.Error != nil {
		return
	}
	var events []Event
	fs.session.
		Preload("Name").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Find(&events)
	for _, event := range events {
		if event.Name.Name == name {
			rp = event
			return
		}
	}
	fs.AddError(gorm.ErrRecordNotFound)
	return
}

func (fs *Fsm) getEventItemByName(name string) (rp EventItem) {
	fs.session.
		Where("name = ?", name).
		First(&rp)
	return
}

func (fs *Fsm) getStartEvent(machineId uint) (rp Event) {
	fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Where("sort = ?", constant.Zero).
		First(&rp)
	return
}

func (fs *Fsm) getPrevEvent(machineId uint, level uint) (rp Event) {
	events := make([]Event, 0)
	fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("machine_id = ?", machineId).
		Where("level = ?", level-1).
		Order("sort").
		Find(&events)
	for _, event := range events {
		if strings.HasSuffix(event.Name.Name, constant.FsmSuffixWaiting) || strings.HasSuffix(event.Name.Name, constant.FsmSuffixResubmit) {
			rp = event
			return
		}
	}
	return
}

func (fs *Fsm) getNextEvent(machineId uint, level uint) (rp Event) {
	events := make([]Event, 0)
	fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("machine_id = ?", machineId).
		Where("level = ?", level+1).
		Order("sort").
		Find(&events)
	for _, event := range events {
		if strings.HasSuffix(event.Name.Name, constant.FsmSuffixWaiting) || strings.HasSuffix(event.Name.Name, constant.FsmSuffixConfirm) {
			rp = event
			return
		}
	}
	return
}

func (fs *Fsm) getEndEvent(machineId uint) (rp Event) {
	fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Order("sort DESC").
		First(&rp)
	return
}

func (fs *Fsm) findEventDesc(machineId uint) (rp []fsm.EventDesc) {
	events := make([]Event, 0)
	list := make([]fsm.EventDesc, 0)
	fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Order("sort").
		Find(&events)
	for _, event := range events {
		var src []string
		for _, item := range event.Src {
			src = append(src, item.Name)
		}
		list = append(list, fsm.EventDesc{
			Name: event.Name.Name,
			Src:  src,
			Dst:  event.Dst.Name,
		})
	}
	if fs.AddError(checkEvent(list)) != nil {
		return
	}
	rp = list
	return
}

// batch create events
// example(two approval level):
// []req.FsmCreateEvent{
//   {
//     Name: "L1",
//   }
//   {
//     Name: "L2",
//   }
// }
// automatic generation by event sort(it is assumed that the SubmitterName=L0) finite state machine table:
// Machine.SubmitConfirm = false
// Current            / Source                    / Destination
// L0 waiting submit  / L1 refused                / L0 submitted
// L1 waiting approve / L0 submitted , L2 refused / L1 approved
// L1 waiting refuse  / L0 submitted              / L1 refused
// L2 waiting approve / L1 approved               / L2 approved
// L2 waiting refuse  / L1 approved               / L2 refused
// end
//
// Machine.SubmitConfirm = true
// Current            / Source                    / Destination
// L0 waiting submit  / L1 refused                / L0 submitted
// L1 waiting approve / L0 submitted , L2 refused / L1 approved
// L1 waiting refuse  / L0 submitted              / L1 refused
// L2 waiting approve / L1 approved               / L2 approved
// L2 waiting refuse  / L1 approved               / L2 refused
// L0 waiting confirm / L2 approved               / L0 confirmed
// end
func (fs *Fsm) batchCreateEvent(machineId uint, r []req.FsmCreateEvent) {
	if fs.Error != nil {
		return
	}
	if len(r) == 0 {
		fs.AddError(ErrEventsNil)
		return
	}
	// clear old machine
	fs.session.
		Unscoped().
		Where("machine_id = ?", machineId).
		Delete(&Event{})

	var machine Machine
	fs.session.
		Model(&Machine{}).
		Where("id = ?", machineId).
		First(&machine)
	if machine.Id == constant.Zero {
		fs.AddError(gorm.ErrRecordNotFound)
		return
	}

	// save event names and event desc
	names := make([]string, 0)
	desc := make([]fsm.EventDesc, 0)
	// save event level for sort setup
	levels := make(map[string]uint, 0)

	// L0 waiting submit / L1 refused / L0 submitted
	l0Name := fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixResubmit)
	l0Srcs := []string{
		fmt.Sprintf("%s %s", r[0].Name, constant.FsmSuffixRefused),
	}
	l0Dst := fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixSubmitted)
	names = append(names, l0Name)
	names = append(names, l0Srcs...)
	names = append(names, l0Dst)
	desc = append(desc, fsm.EventDesc{
		Name: l0Name,
		Src:  l0Srcs,
		Dst:  l0Dst,
	})
	levels[l0Name] = 0
	levels[l0Srcs[0]] = 0
	levels[l0Dst] = 0

	l := len(r)
	for i := 0; i < l; i++ {
		// approve
		// L1 waiting approve / L0 submitted , L2 refused / L1 approved
		// L2 waiting approve / L1 approved               / L2 approved
		li1Name := fmt.Sprintf("%s %s", r[i].Name, constant.FsmSuffixWaiting)
		li1Srcs := make([]string, 0)
		if i > 0 {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", r[i-1].Name, constant.FsmSuffixApproved))
		} else {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixSubmitted))
		}
		li1Dst := fmt.Sprintf("%s %s", r[i].Name, constant.FsmSuffixApproved)
		if i+1 < l {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", r[i+1].Name, constant.FsmSuffixRefused))
		}
		names = append(names, li1Name)
		names = append(names, li1Srcs...)
		names = append(names, li1Dst)
		desc = append(desc, fsm.EventDesc{
			Name: li1Name,
			Src:  li1Srcs,
			Dst:  li1Dst,
		})
		levels[li1Name] = uint(i + 1)
		levels[li1Dst] = uint(i + 1)

		// refuse
		// L1 waiting refuse / L0 submitted / L1 refused
		// L2 waiting refuse / L1 approved  / L2 refused
		li2Name := fmt.Sprintf("%s %s", r[i].Name, constant.FsmSuffixWaiting)
		li2Srcs := make([]string, 0)
		if i == 0 {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixSubmitted))
		} else {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", r[i-1].Name, constant.FsmSuffixApproved))
			if i+1 < l {
				li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", r[i+1].Name, constant.FsmSuffixRefused))
			}
		}
		li2Dst := fmt.Sprintf("%s %s", r[i].Name, constant.FsmSuffixRefused)
		names = append(names, li2Name)
		names = append(names, li2Srcs...)
		names = append(names, li2Dst)
		desc = append(desc, fsm.EventDesc{
			Name: li2Name,
			Src:  li2Srcs,
			Dst:  li2Dst,
		})
		levels[li2Name] = uint(i + 1)
		levels[li2Dst] = uint(i + 1)
	}
	if machine.SubmitterConfirm == constant.One {
		// L0 waiting confirm / L2 approved / L0 confirmed
		l0Name := fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixConfirm)
		l0Srcs := []string{
			fmt.Sprintf("%s %s", r[l-1].Name, constant.FsmSuffixApproved),
		}
		l0Dst := fmt.Sprintf("%s %s", machine.SubmitterName, constant.FsmSuffixConfirmed)
		names = append(names, l0Name)
		names = append(names, l0Srcs...)
		names = append(names, l0Dst)
		desc = append(desc, fsm.EventDesc{
			Name: l0Name,
			Src:  l0Srcs,
			Dst:  l0Dst,
		})
		levels[l0Name] = uint(l + 1)
		levels[l0Dst] = uint(l + 1)
	}

	// remove repeat name
	names = utils.RemoveRepeat(names)
	oldItems := make([]EventItem, 0)
	fs.session.
		Where("name IN (?)", names).
		Find(&oldItems)
	items := make([]EventItem, 0)
	for _, name := range names {
		exists := false
		for _, item := range oldItems {
			if name == item.Name {
				exists = true
			}
		}
		if !exists {
			items = append(items, EventItem{
				Name: name,
			})
		}
	}
	if len(items) > 0 {
		fs.session.Create(&items)
	}
	newItems := make([]EventItem, 0)
	fs.session.
		Where("name IN (?)", names).
		Find(&newItems)
	events := make([]Event, 0)
	for i, d := range desc {
		nameId := uint(0)
		dstId := uint(0)
		src := make([]EventItem, 0)
		for _, item := range newItems {
			if item.Name == d.Name {
				nameId = item.Id
			}
			if item.Name == d.Dst {
				dstId = item.Id
			}
			if utils.Contains(d.Src, item.Name) {
				src = append(src, item)
			}
		}

		// default: no edit permission
		edit := constant.Zero
		editFields := ""
		roles := make([]Role, 0)
		users := make([]User, 0)
		if i == 0 {
			// submitter has edit permission
			edit = constant.One
			editFields = machine.SubmitterEditFields
		} else if i < len(desc)-1 || machine.SubmitterConfirm == constant.Zero {
			// machine middle levels
			index := (i+1)/2 - 1
			edit = uint(r[index].Edit)
			editFields = r[index].EditFields
			// find roles/users
			roles = fs.findRole(r[index].Roles.Uints())
			users = fs.findUser(r[index].Users.Uints())
		} else if i == len(desc)-1 && machine.SubmitterConfirm == constant.One {
			// save submitter confirm edit fields
			edit = constant.One
			editFields = machine.SubmitterConfirmEditFields
		}

		events = append(events, Event{
			MachineId:  machineId,
			Sort:       uint(i),
			Level:      levels[d.Name],
			NameId:     nameId,
			Src:        src,
			DstId:      dstId,
			Edit:       edit,
			EditFields: editFields,
			Roles:      roles,
			Users:      users,
		})
	}
	if len(events) > 0 {
		fs.session.Create(&events)
	}
	return
}

func (fs *Fsm) findUser(ids []uint) []User {
	users := make([]User, 0)
	fs.session.
		Model(&User{}).
		Where("id IN (?)", ids).
		Find(&users)
	oldIds := make([]uint, 0)
	newUsers := make([]User, 0)
	for _, user := range users {
		oldIds = append(oldIds, user.Id)
	}
	for _, id := range ids {
		if !utils.Contains(oldIds, id) {
			newUsers = append(newUsers, User{Id: id})
		}
	}
	if len(newUsers) > 0 {
		fs.session.Create(&newUsers)
		users = append(users, newUsers...)
	}
	return users
}

func (fs *Fsm) findRole(ids []uint) []Role {
	roles := make([]Role, 0)
	fs.session.
		Model(&Role{}).
		Where("id IN (?)", ids).
		Find(&roles)
	oldIds := make([]uint, 0)
	newRoles := make([]Role, 0)
	for _, user := range roles {
		oldIds = append(oldIds, user.Id)
	}
	for _, id := range ids {
		if !utils.Contains(oldIds, id) {
			newRoles = append(newRoles, Role{Id: id})
		}
	}
	if len(newRoles) > 0 {
		fs.session.Create(&newRoles)
		roles = append(roles, newRoles...)
	}
	return roles
}

func (fs *Fsm) initSession() *gorm.DB {
	namingStrategy := schema.NamingStrategy{
		TablePrefix:   fs.ops.prefix,
		SingularTable: true,
	}
	session := fs.ops.db.WithContext(fs.ops.ctx).Session(&gorm.Session{})
	session.NamingStrategy = namingStrategy
	return session
}

func (fs *Fsm) AddError(err error) error {
	if fs.Error == nil {
		fs.Error = err
	} else if err != nil && !errors.Is(err, fs.Error) {
		fs.Error = fmt.Errorf("%v; %w", fs.Error, err)
	}
	return fs.Error
}

func getNextItemName(approved uint, eventName string) string {
	name := eventName
	if strings.HasSuffix(eventName, constant.FsmSuffixWaiting) {
		if approved == constant.FsmLogStatusRefused {
			name = strings.TrimSuffix(eventName, constant.FsmSuffixWaiting) + constant.FsmSuffixRefused
		} else {
			name = strings.TrimSuffix(eventName, constant.FsmSuffixWaiting) + constant.FsmSuffixApproved
		}
	}
	if strings.HasSuffix(eventName, constant.FsmSuffixResubmit) {
		name = strings.TrimSuffix(eventName, constant.FsmSuffixResubmit) + constant.FsmSuffixSubmitted
	}
	if strings.HasSuffix(eventName, constant.FsmSuffixConfirm) {
		name = strings.TrimSuffix(eventName, constant.FsmSuffixConfirm) + constant.FsmSuffixConfirmed
	}
	return name
}

// check that the state machine is valid(traverse each event, only one end position)
func checkEvent(desc []fsm.EventDesc) (err error) {
	names := make([]string, 0)
	for _, item := range desc {
		names = append(names, item.Dst)
		names = append(names, item.Src...)
	}
	names = utils.RemoveRepeat(names)
	if len(names) == 0 {
		err = ErrEventNameNil
		return
	}

	f := fsm.NewFSM(names[0], desc, nil)
	// save end count
	endCount := 0
	for _, item := range names {
		f.SetState(item)
		if len(f.AvailableTransitions()) == 0 {
			endCount++
		}
	}
	if endCount != 1 {
		err = ErrEventEndPointNotUnique
		return
	}
	return
}
