package fsm

import (
	"fmt"
	"github.com/looplab/fsm"
	"github.com/piupuer/go-helper/models"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

type Fsm struct {
	session *gorm.DB
	ops     Options
	Error   error
}

// mysql DDL migrate rollback is not supported, Migrate before New
func Migrate(db *gorm.DB, options ...func(*Options)) error {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	session := initSession(db, ops.prefix)
	return session.AutoMigrate(
		new(Machine),
		new(Event),
		new(User),
		new(EventSrcItemRelation),
		new(EventUserRelation),
		new(EventItem),
		new(Log),
		new(LogApprovalUserRelation),
	)
}

func New(tx *gorm.DB, options ...func(*Options)) *Fsm {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	fs := &Fsm{
		ops: *ops,
	}
	if tx != nil {
		fs.session = initSession(tx, ops.prefix)
	} else {
		fs.Error = ErrDbNil
	}
	return fs
}

func (fs Fsm) CreateMachine(r req.FsmCreateMachine) (*Machine, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var machine Machine
	utils.Struct2StructByJson(r, &machine)
	// save json for query
	machine.EventsJson = utils.Struct2Json(r.Levels)
	err := fs.session.Create(&machine).Error
	if err != nil {
		return nil, err
	}
	// batch fsm event
	err = fs.batchCreateEvents(machine.Id, r.Levels)
	if err != nil {
		return nil, err
	}
	_, err = fs.findEventDesc(machine.Id)
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

// =======================================================
// approval log function
// =======================================================
// first submit log
func (fs Fsm) SubmitLog(r req.FsmCreateLog) ([]EventItem, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// check whether approval is pending
	_, err := fs.getLastPendingLog(req.FsmLog{
		Category: r.Category,
		Uuid:     r.Uuid,
	})
	if err != gorm.ErrRecordNotFound {
		return nil, ErrRepeatSubmit
	}
	startEvent, err := fs.getStartEvent(r.MachineId)
	if err != nil {
		return nil, err
	}

	// first create log
	var log Log
	log.Category = uint(r.Category)
	log.Uuid = r.Uuid
	nextEvent, err := fs.getNextEvent(r.MachineId, startEvent.Level)
	if err != nil {
		return nil, err
	}
	log.ProgressId = startEvent.DstId
	log.CanApprovalRoles = nextEvent.Roles
	log.CanApprovalUsers = nextEvent.Users
	log.SubmitterRoleId = r.SubmitterRoleId
	log.SubmitterUserId = r.SubmitterUserId
	log.PrevDetail = startEvent.Dst.Name
	log.Detail = nextEvent.Name.Name
	log.CurrentEventId = startEvent.Id
	log.NextEventId = nextEvent.Id
	err = fs.session.Create(&log).Error
	if err != nil {
		return nil, err
	}

	return []EventItem{
		startEvent.Dst,
		nextEvent.Name,
	}, nil
}

// start approve log
func (fs Fsm) ApproveLog(r req.FsmApproveLog) (*resp.FsmApprovalLog, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	approved := uint(r.Approved)
	var rp resp.FsmApprovalLog
	// check current user/role permission
	oldLog, err := fs.CheckLogPermission(req.FsmPermissionLog{
		Category:       r.Category,
		Uuid:           r.Uuid,
		ApprovalRoleId: r.ApprovalRoleId,
		ApprovalUserId: r.ApprovalUserId,
		Approved:       approved,
	})
	if err != nil {
		return nil, err
	}

	// submitter cancel
	if approved == LogStatusCancelled {
		m := make(map[string]interface{}, 0)
		m["approved"] = LogStatusCancelled
		m["approval_role_id"] = r.ApprovalRoleId
		m["approval_user_id"] = r.ApprovalUserId
		m["approval_opinion"] = r.ApprovalOpinion
		m["next_event_id"] = models.Zero
		m["detail"] = MsgSubmitterCancel
		rp.Cancel = true
		err = fs.session.
			Model(&Log{}).
			Where("id = ?", oldLog.Id).
			Updates(&m).Error
		if err != nil {
			return nil, err
		}
		return &rp, nil
	}

	desc, err := fs.findEventDesc(r.MachineId)
	if err != nil {
		return nil, err
	}
	// create fsm instance from current progress
	f := fsm.NewFSM(oldLog.Progress.Name, desc, nil)

	transitions := f.AvailableTransitions()
	eventName := ""
	for _, transition := range transitions {
		match := false
		switch approved {
		case LogStatusApproved:
			if strings.HasSuffix(transition, SuffixWaiting) || strings.HasSuffix(transition, SuffixResubmit) || strings.HasSuffix(transition, SuffixConfirm) {
				match = true
			}
		case LogStatusRefused:
			if strings.HasSuffix(transition, SuffixWaiting) {
				match = true
			}
		}
		if match {
			eventName = transition
			break
		}
	}

	if eventName == "" {
		return nil, ErrStatus
	}
	nextName := getNextItemName(approved, eventName)
	f.SetState(nextName)
	event, err := fs.getEvent(r.MachineId, eventName)
	if err != nil {
		return nil, err
	}
	progressItem, err := fs.getEventItemByName(nextName)
	if err != nil {
		return nil, err
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
		var nextEvent *Event
		if approved == LogStatusApproved {
			nextEvent, err = fs.getNextEvent(r.MachineId, event.Level)
		} else {
			nextEvent, err = fs.getPrevEvent(r.MachineId, event.Level)
		}
		if err != nil {
			return nil, err
		}
		// no users/roles, maybe submitter resubmit/confirm
		noUser := false
		if len(nextEvent.Roles) == 0 && len(nextEvent.Users) == 0 {
			noUser = true
			if strings.HasSuffix(nextEvent.Name.Name, SuffixConfirm) {
				rp.WaitingConfirm = true
			} else {
				rp.WaitingResubmit = true
			}
		}
		newLog.ProgressId = progressItem.Id
		newLog.NextEventId = nextEvent.Id
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
		rp.End = true
		newLog.Approved = LogStatusApproved
		newLog.Detail = MsgEnded
	}
	err = fs.session.Create(&newLog).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, 0)
	m["approved"] = LogStatusApproved
	if approved == LogStatusRefused {
		m["approved"] = LogStatusRefused
	}
	m["approval_role_id"] = r.ApprovalRoleId
	m["approval_user_id"] = r.ApprovalUserId
	m["approval_opinion"] = r.ApprovalOpinion
	// update oldLog approved
	err = fs.session.
		Model(&Log{}).
		Where("id = ?", oldLog.Id).
		Updates(&m).Error
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// cancel log by category(it is applicable to the automatic cancellation of records to be approved when the approval configuration changes)
func (fs Fsm) CancelLog(category uint) error {
	if fs.Error != nil {
		return fs.Error
	}
	m := make(map[string]interface{}, 0)
	m["approved"] = LogStatusCancelled
	m["next_event_id"] = models.Zero
	m["detail"] = MsgConfigChanged
	return fs.session.
		Model(&Log{}).
		Where("category = ?", category).
		Where("approved = ?", LogStatusWaiting).
		Updates(&m).Error
}

// =======================================================
// query function
// =======================================================
// check verify whether the current user/role has permission to approve
func (fs Fsm) CheckLogPermission(r req.FsmPermissionLog) (*Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// check whether approval is pending
	log, err := fs.getLastPendingLog(req.FsmLog{
		Category: r.Category,
		Uuid:     r.Uuid,
	})
	if err != nil {
		return nil, ErrNoPermissionOrEnded
	}
	if r.Approved == LogStatusCancelled {
		if log.SubmitterRoleId != r.ApprovalRoleId && log.SubmitterUserId != r.ApprovalUserId {
			return nil, ErrOnlySubmitterCancel
		} else {
			if log.CurrentEvent.Level > models.Zero {
				return nil, ErrStartedCannotCancel
			}
			return log, nil
		}
	}
	roles := make([]uint, 0)
	users := make([]uint, 0)
	for _, role := range log.CanApprovalRoles {
		roles = append(roles, role.Id)
	}
	for _, user := range log.CanApprovalUsers {
		users = append(users, user.Id)
	}
	if !utils.Contains(roles, r.ApprovalRoleId) && !utils.Contains(users, r.ApprovalUserId) {
		return nil, ErrNoPermissionApprove
	}
	if r.Approved == LogStatusRefused && log.NextEvent.Refuse == models.Zero {
		return nil, ErrNoPermissionRefuse
	}
	return log, nil
}

// find machines
func (fs Fsm) FindMachine(req req.FsmMachine) ([]resp.FsmMachine, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	machines := make([]Machine, 0)
	query := fs.session
	name := strings.TrimSpace(req.Name)
	if name != "" {
		query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	submitterName := strings.TrimSpace(req.SubmitterName)
	if submitterName != "" {
		query.Where("submitter_name LIKE ?", fmt.Sprintf("%%%s%%", submitterName))
	}
	if req.SubmitterConfirm != nil {
		query.Where("submitter_confirm = ?", *req.SubmitterConfirm)
	}
	err := query.Find(&machines).Error
	list := make([]resp.FsmMachine, 0)
	utils.Struct2StructByJson(machines, &list)
	return list, err
}

// find logs
func (fs Fsm) FindLog(req req.FsmLog) ([]Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var logs []Log
	err := fs.session.
		Preload("CurrentEvent").
		Preload("CurrentEvent.Roles").
		Preload("CurrentEvent.Users").
		Preload("NextEvent").
		Preload("NextEvent.Roles").
		Preload("NextEvent.Users").
		Preload("CanApprovalRoles").
		Preload("CanApprovalUsers").
		Where("category = ?", req.Category).
		Where("uuid = ?", req.Uuid).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// find log tracks
func (fs Fsm) FindLogTrack(logs []Log) ([]resp.FsmLogTrack, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	track := make([]resp.FsmLogTrack, 0)
	if len(logs) == 0 {
		return track, nil
	}
	l := len(logs)
	for i, log := range logs {
		prevCancel := false
		prevOpinion := ""
		end := false
		cancel := log.Approved == LogStatusCancelled
		if i > 0 {
			prevCancel = logs[i-1].Approved == LogStatusCancelled
			prevOpinion = logs[i-1].ApprovalOpinion
		}
		if i == l-1 && log.NextEventId == models.Zero {
			end = true
		}
		if end || cancel {
			track = append(track, resp.FsmLogTrack{
				Name:    log.PrevDetail,
				Opinion: prevOpinion,
				Cancel:  prevCancel,
			}, resp.FsmLogTrack{
				Name:    log.Detail,
				Opinion: log.ApprovalOpinion,
				End:     end,
				Cancel:  cancel,
			})
		} else {
			track = append(track, resp.FsmLogTrack{
				Name:    log.PrevDetail,
				Opinion: prevOpinion,
				End:     end,
				Cancel:  cancel,
			})
		}
		if i == l-1 && log.Approved == LogStatusWaiting {
			track = append(track, resp.FsmLogTrack{
				Name: logs[i].Detail,
			})
		}
	}
	return track, nil
}

// get the pending approval list of a approver
func (fs Fsm) FindPendingLogByApprover(req req.FsmPendingLog) ([]Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// get user relation
	logIds1 := make([]uint, 0)
	err := fs.session.
		Model(&LogApprovalUserRelation{}).
		Where("user_id = ?", req.ApprovalUserId).
		Pluck("log_id", &logIds1).Error
	if err != nil {
		return nil, err
	}
	// get role relation
	logIds2 := make([]uint, 0)
	err = fs.session.
		Model(&LogApprovalRoleRelation{}).
		Where("role_id = ?", req.ApprovalRoleId).
		Pluck("log_id", &logIds2).Error
	if err != nil {
		return nil, err
	}
	logs := make([]Log, 0)
	query := fs.session.
		Where("approved = ?", LogStatusWaiting).
		Where("id IN (?)", append(logIds1, logIds2...))
	if uint(req.Category) > models.Zero {
		query.Where("category = ?", req.Category)
	}
	err = query.Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// =======================================================
// private function
// =======================================================
// get last pending log, err will be returned when it does not exist
// 获取最后一条待审批日志
func (fs Fsm) getLastPendingLog(req req.FsmLog) (*Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var log Log
	err := fs.session.
		Preload("CanApprovalRoles").
		Preload("CanApprovalUsers").
		Preload("Progress").
		Preload("NextEvent").
		Where("category = ?", req.Category).
		Where("uuid = ?", req.Uuid).
		Where("approved = ?", LogStatusWaiting).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (fs Fsm) getEvent(machineId uint, name string) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if event.Name.Name == name {
			return &event, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (fs Fsm) getEventItemByName(name string) (*EventItem, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var item EventItem
	err := fs.session.
		Where("name = ?", name).
		First(&item).Error
	return &item, err
}

func (fs Fsm) getStartEvent(machineId uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var event Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Where("sort = ?", models.Zero).
		First(&event).Error
	return &event, err
}

func (fs Fsm) getPrevEvent(machineId uint, level uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}

	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("machine_id = ?", machineId).
		Where("level = ?", level-1).
		Order("sort").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if strings.HasSuffix(event.Name.Name, SuffixWaiting) || strings.HasSuffix(event.Name.Name, SuffixResubmit) {
			return &event, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (fs Fsm) getNextEvent(machineId uint, level uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("machine_id = ?", machineId).
		Where("level = ?", level+1).
		Order("sort").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		if strings.HasSuffix(event.Name.Name, SuffixWaiting) || strings.HasSuffix(event.Name.Name, SuffixConfirm) {
			return &event, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (fs Fsm) getEndEvent(machineId uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var event Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Order("sort DESC").
		First(&event).Error
	return &event, err
}

func (fs Fsm) findEventDesc(machineId uint) ([]fsm.EventDesc, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	events := make([]Event, 0)
	desc := make([]fsm.EventDesc, 0)
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("machine_id = ?", machineId).
		Order("sort").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		var src []string
		for _, item := range event.Src {
			src = append(src, item.Name)
		}
		desc = append(desc, fsm.EventDesc{
			Name: event.Name.Name,
			Src:  src,
			Dst:  event.Dst.Name,
		})
	}
	err = checkEvents(desc)
	if err != nil {
		return nil, err
	}
	return desc, nil
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
func (fs Fsm) batchCreateEvents(machineId uint, r []req.FsmCreateEvent) (err error) {
	if fs.Error != nil {
		return fs.Error
	}
	if len(r) == 0 {
		return ErrEventsNil
	}
	// clear old machine
	err = fs.session.
		Where("machine_id = ?", machineId).
		Delete(&Event{}).Error
	if err != nil {
		return err
	}

	var machine Machine
	err = fs.session.
		Model(&Machine{}).
		Where("id = ?", machineId).
		First(&machine).Error
	if err != nil {
		return err
	}

	// save event names and event desc
	names := make([]string, 0)
	desc := make([]fsm.EventDesc, 0)
	// save event level for sort setup
	levels := make(map[string]uint, 0)

	// L0 waiting submit / L1 refused / L0 submitted
	l0Name := fmt.Sprintf("%s %s", machine.SubmitterName, SuffixResubmit)
	l0Srcs := []string{
		fmt.Sprintf("%s %s", r[0].Name, SuffixRefused),
	}
	l0Dst := fmt.Sprintf("%s %s", machine.SubmitterName, SuffixSubmitted)
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
		li1Name := fmt.Sprintf("%s %s", r[i].Name, SuffixWaiting)
		li1Srcs := make([]string, 0)
		if i > 0 {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", r[i-1].Name, SuffixApproved))
		} else {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", machine.SubmitterName, SuffixSubmitted))
		}
		li1Dst := fmt.Sprintf("%s %s", r[i].Name, SuffixApproved)
		if i+1 < l {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s %s", r[i+1].Name, SuffixRefused))
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
		li2Name := fmt.Sprintf("%s %s", r[i].Name, SuffixWaiting)
		li2Srcs := make([]string, 0)
		if i == 0 {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", machine.SubmitterName, SuffixSubmitted))
		} else {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", r[i-1].Name, SuffixApproved))
			if i+1 < l {
				li2Srcs = append(li2Srcs, fmt.Sprintf("%s %s", r[i+1].Name, SuffixRefused))
			}
		}
		li2Dst := fmt.Sprintf("%s %s", r[i].Name, SuffixRefused)
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
	if machine.SubmitterConfirm == models.One {
		// L0 waiting confirm / L2 approved / L0 confirmed
		l0Name := fmt.Sprintf("%s %s", machine.SubmitterName, SuffixConfirm)
		l0Srcs := []string{
			fmt.Sprintf("%s %s", r[l-1].Name, SuffixApproved),
		}
		l0Dst := fmt.Sprintf("%s %s", machine.SubmitterName, SuffixConfirmed)
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
	err = fs.session.
		Where("name IN (?)", names).
		Find(&oldItems).Error
	if err != nil {
		return err
	}
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
		err = fs.session.Create(&items).Error
		if err != nil {
			return err
		}
	}
	newItems := make([]EventItem, 0)
	err = fs.session.
		Where("name IN (?)", names).
		Find(&newItems).Error
	if err != nil {
		return err
	}
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

		// default: no edit/refuse permission
		edit := models.Zero
		refuse := models.Zero
		editFields := ""
		roles := make([]Role, 0)
		users := make([]User, 0)
		if i == 0 {
			// submitter has edit perssion
			edit = models.One
			editFields = machine.SubmitterEditFields
		} else if i < len(desc)-1 || machine.SubmitterConfirm == models.Zero {
			// machineIddle levels
			index := (i+1)/2 - 1
			edit = uint(r[index].Edit)
			editFields = r[index].EditFields
			refuse = uint(r[index].Refuse)
			// mock roles/users
			roles = fs.getRoles(r[index].Roles)
			users = fs.getUsers(r[index].Users)
		} else if i == len(desc)-1 && machine.SubmitterConfirm == models.One {
			// save submitter confirm edit fields
			edit = models.One
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
			Refuse:     refuse,
			Roles:      roles,
			Users:      users,
		})
	}
	if len(events) > 0 {
		err = fs.session.Create(&events).Error
		if err != nil {
			return err
		}
	}
	return
}

func (fs Fsm) getUsers(ids []uint) []User {
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

func (fs Fsm) getRoles(ids []uint) []Role {
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

func getNextItemName(approved uint, eventName string) string {
	name := eventName
	if strings.HasSuffix(eventName, SuffixWaiting) {
		if approved == LogStatusRefused {
			name = strings.TrimSuffix(eventName, SuffixWaiting) + SuffixRefused
		} else {
			name = strings.TrimSuffix(eventName, SuffixWaiting) + SuffixApproved
		}
	}
	if strings.HasSuffix(eventName, SuffixResubmit) {
		name = strings.TrimSuffix(eventName, SuffixResubmit) + SuffixSubmitted
	}
	if strings.HasSuffix(eventName, SuffixConfirm) {
		name = strings.TrimSuffix(eventName, SuffixConfirm) + SuffixConfirmed
	}
	return name
}

func initSession(db *gorm.DB, prefix string) *gorm.DB {
	namingStrategy := schema.NamingStrategy{
		TablePrefix:   prefix,
		SingularTable: true,
	}
	session := db.Session(&gorm.Session{})
	session.NamingStrategy = namingStrategy
	return session
}

// check that the state machine is valid(traverse each event, only one end position)
func checkEvents(desc []fsm.EventDesc) error {
	names := make([]string, 0)
	for _, item := range desc {
		names = append(names, item.Dst)
		names = append(names, item.Src...)
	}
	names = utils.RemoveRepeat(names)
	if len(names) == 0 {
		return ErrEventNameNil
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
		return ErrEventEndPointNotUnique
	}
	return nil
}