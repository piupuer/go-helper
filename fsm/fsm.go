package fsm

import (
	"fmt"
	"github.com/looplab/fsm"
	"github.com/piupuer/go-helper/fsm/request"
	"github.com/piupuer/go-helper/fsm/response"
	"github.com/piupuer/go-helper/models"
	"github.com/piupuer/go-helper/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

type Fsm struct {
	session *gorm.DB
	ops     Options
	Error   error
}

// mysql DDL不支持回滚操作, 因此这里单独migrate
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

// 创建状态机
func (fs Fsm) CreateMachine(req request.CreateMachineReq) (*Machine, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var machine Machine
	utils.Struct2StructByJson(req, &machine)
	// 将传入的事件转为json存储(主要是方便查询)
	machine.EventsJson = utils.Struct2Json(req.Levels)
	// 创建数据
	err := fs.session.Create(&machine).Error
	if err != nil {
		return nil, err
	}
	// 绑定事件
	err = fs.batchCreateEvents(machine.Id, req.Levels)
	if err != nil {
		return nil, err
	}
	// 获取全部事件(内部自动校验合法性)
	_, err = fs.findEventDesc(machine.Id)
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

// =======================================================
// 审批相关
// =======================================================
// 首次提交日志记录
func (fs Fsm) SubmitLog(req request.CreateLogReq) ([]EventItem, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// 判断是否未结束
	_, err := fs.getLastPendingLog(request.LogReq{
		Category: req.Category,
		Uuid:     req.Uuid,
	})
	if err != gorm.ErrRecordNotFound {
		return nil, ErrRepeatSubmit
	}
	// 获取开始事件
	startEvent, err := fs.getStartEvent(req.MId)
	if err != nil {
		return nil, err
	}

	// 绑定下一级审批人
	var log Log
	log.Category = req.Category
	log.Uuid = req.Uuid
	// 获取下一事件
	nextEvent, err := fs.getNextEvent(req.MId, startEvent.Level)
	if err != nil {
		return nil, err
	}
	log.ProgressId = startEvent.DstId
	log.CanApprovalRoles = nextEvent.Roles
	log.CanApprovalUsers = nextEvent.Users
	log.SubmitterRoleId = req.SubmitterRoleId
	log.SubmitterUserId = req.SubmitterUserId
	// 记录操作日志
	log.PrevDetail = startEvent.Dst.Name
	log.Detail = nextEvent.Name.Name
	log.CurrentEventId = startEvent.Id
	log.NextEventId = nextEvent.Id
	err = fs.session.Create(&log).Error
	if err != nil {
		return nil, err
	}

	// 返回开始事件的结束位置以及下一事件的开始位置
	return []EventItem{
		startEvent.Dst,
		nextEvent.Name,
	}, nil
}

// 审批某条日志记录
func (fs Fsm) ApproveLog(req request.ApproveLogReq) (*response.ApprovalLogResp, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var resp response.ApprovalLogResp
	// 验证权限
	oldLog, err := fs.CheckLogPermission(request.PermissionLogReq{
		Category:       req.Category,
		Uuid:           req.Uuid,
		ApprovalRoleId: req.ApprovalRoleId,
		ApprovalUserId: req.ApprovalUserId,
		Approved:       req.Approved,
	})
	if err != nil {
		return nil, err
	}

	if req.Approved == LogStatusCancelled {
		oldLog.Approved = LogStatusCancelled
		oldLog.ApprovalRoleId = req.ApprovalRoleId
		oldLog.ApprovalUserId = req.ApprovalUserId
		oldLog.ApprovalOpinion = req.ApprovalOpinion
		oldLog.Detail = MsgSubmitterCancel
		resp.Cancel = true
		// 更新为已取消
		err = fs.session.
			Model(&Log{}).
			Where("id = ?", oldLog.Id).
			Updates(&oldLog).Error
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	// 获取全部事件
	desc, err := fs.findEventDesc(req.MId)
	if err != nil {
		return nil, err
	}
	// 从当前进度开始, 创建状态机实例
	f := fsm.NewFSM(oldLog.Progress.Name, desc, nil)

	transitions := f.AvailableTransitions()
	eventName := ""
	for _, transition := range transitions {
		match := false
		switch req.Approved {
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
	nextName := getNextItemName(req.Approved, eventName)
	if f.Can(nextName) {
		fmt.Println("ok")
	}
	f.SetState(nextName)
	event, err := fs.getEvent(req.MId, eventName)
	if err != nil {
		return nil, err
	}
	progressItem, err := fs.getEventItemByName(nextName)
	if err != nil {
		return nil, err
	}
	var newLog Log
	newLog.Category = req.Category
	newLog.Uuid = req.Uuid
	newLog.SubmitterRoleId = oldLog.SubmitterRoleId
	newLog.SubmitterUserId = oldLog.SubmitterUserId
	newLog.PrevDetail = nextName
	newLog.CurrentEventId = event.Id
	if len(f.AvailableTransitions()) != 0 {
		// 绑定下一级审批人
		// 获取下一事件
		var nextEvent *Event
		if req.Approved == LogStatusApproved {
			nextEvent, err = fs.getNextEvent(req.MId, event.Level)
		} else {
			nextEvent, err = fs.getPrevEvent(req.MId, event.Level)
		}
		if err != nil {
			return nil, err
		}
		// 没有审批人, 可能是需要提交人确认/或提交人重新申请
		noUser := false
		if len(nextEvent.Roles) == 0 && len(nextEvent.Users) == 0 {
			noUser = true
			if strings.HasSuffix(nextEvent.Name.Name, SuffixConfirm) {
				resp.WaitingConfirm = true
			} else {
				resp.WaitingResubmit = true
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
		resp.End = true
		newLog.Approved = LogStatusApproved
		newLog.Detail = MsgEnded
	}
	err = fs.session.Create(&newLog).Error
	if err != nil {
		return nil, err
	}
	if req.Approved == LogStatusRefused {
		oldLog.Approved = LogStatusRefused
	} else {
		oldLog.Approved = LogStatusApproved
	}
	oldLog.ApprovalRoleId = req.ApprovalRoleId
	oldLog.ApprovalUserId = req.ApprovalUserId
	oldLog.ApprovalOpinion = req.ApprovalOpinion
	// 更新为已审批
	err = fs.session.
		Model(&Log{}).
		Where("id = ?", oldLog.Id).
		Updates(&oldLog).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// 取消某个类别下的审批日志(适用于审批配置发生变化, 待审批的记录自动取消)
func (fs Fsm) CancelLogs(category uint) error {
	if fs.Error != nil {
		return fs.Error
	}
	return fs.session.
		Where("category = ?", category).
		Where("approved = ?", LogStatusWaiting).
		Updates(Log{
			Approved: LogStatusCancelled,
			Detail:   MsgConfigChanged,
		}).Error
}

// =======================================================
// 查询相关
// =======================================================
// 校验是否有权限
func (fs Fsm) CheckLogPermission(req request.PermissionLogReq) (*Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// 判断是否未结束
	log, err := fs.getLastPendingLog(request.LogReq{
		Category: req.Category,
		Uuid:     req.Uuid,
	})
	if err != nil {
		return nil, ErrNoPermissionOrEnded
	}
	if req.Approved == LogStatusCancelled {
		if log.SubmitterRoleId != req.ApprovalRoleId && log.SubmitterUserId != req.ApprovalUserId {
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
	if !utils.Contains(roles, req.ApprovalRoleId) && !utils.Contains(users, req.ApprovalUserId) {
		return nil, ErrNoPermissionApprove
	}
	if req.Approved == LogStatusRefused && log.NextEvent.Refuse == models.Zero {
		return nil, ErrNoPermissionRefuse
	}
	return log, nil
}

// 获取全部审批日志
func (fs Fsm) FindLogs(req request.LogReq) ([]Log, error) {
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

// 获取某个用户的待审批列表
func (fs Fsm) FindPendingLogsByApprover(req request.PendingLogReq) ([]Log, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// 查询关联的日志编号
	var logIds1 []uint
	err := fs.session.
		Model(&LogApprovalUserRelation{}).
		Where("user_id = ?", req.ApprovalUserId).
		Pluck("log_id", &logIds1).Error
	if err != nil {
		return nil, err
	}
	var logIds2 []uint
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
	if req.Category > models.Zero {
		query = query.Where("category = ?", req.Category)
	}
	err = query.Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// =======================================================
// 私有方法
// =======================================================
// 获取最后一条待审批日志
func (fs Fsm) getLastPendingLog(req request.LogReq) (*Log, error) {
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

// 根据事件名称获取事件
func (fs Fsm) getEvent(mId uint, name string) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Dst").
		Where("m_id = ?", mId).
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

// 根据事件item名称获取item
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

// 获取状态机事件开始位置
func (fs Fsm) getStartEvent(mId uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var event Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("m_id = ?", mId).
		Where("sort = ?", models.Zero).
		First(&event).Error
	return &event, err
}

// 获取状态机事件前一位置
func (fs Fsm) getPrevEvent(mId uint, level uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}

	// 包含通过二字的第一条
	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("m_id = ?", mId).
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

// 获取状态机事件下一位置
func (fs Fsm) getNextEvent(mId uint, level uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	// 包含通过二字的第一条
	var events []Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Preload("Roles").
		Preload("Users").
		Where("m_id = ?", mId).
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

// 获取状态机事件结束位置
func (fs Fsm) getEndEvent(mId uint) (*Event, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	var event Event
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("m_id = ?", mId).
		Order("sort DESC").
		First(&event).Error
	return &event, err
}

// 获取状态机事件实例
func (fs Fsm) findEventDesc(mId uint) ([]fsm.EventDesc, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	events := make([]Event, 0)
	desc := make([]fsm.EventDesc, 0)
	err := fs.session.
		Preload("Name").
		Preload("Src").
		Preload("Dst").
		Where("m_id = ?", mId).
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

// 创建状态机事件
// 逻辑大致如下(层级可扩展为N级):
// 前端输入
// 1.第一级审批(记为标识符L1)
// 2.第二级审批(记为标识符L2)
// 系统需要生成(按Event顺序)
// 提交人记为L0(名称和可编辑字段在此确认SysFsm.SubmitterName/SubmitterEditFields)
// L0申请/L1已拒绝/L0已提交
// L1通过/L0已提交,L2已拒绝/L1已通过
// L1拒绝/L0已提交/L1已拒绝
// L2通过/L1已通过/L2已通过
// L2拒绝/L1已通过/L2已拒绝
// L2通过后需判断SysFsm.SubmitConfirm
// SysFsm.SubmitConfirm=0, 流程结束
// SysFsm.SubmitConfirm=1, 需加最后一个流程: L0确认/L2已通过/L0已确认
func (fs Fsm) batchCreateEvents(mId uint, req []request.CreateEventReq) (err error) {
	if fs.Error != nil {
		return fs.Error
	}
	if len(req) == 0 {
		return ErrEventsNil
	}
	// 清除旧数据
	err = fs.session.
		Where("m_id = ?", mId).
		Delete(&Event{}).Error
	if err != nil {
		return err
	}

	var machine Machine
	err = fs.session.
		Model(&Machine{}).
		Where("id = ?", mId).
		First(&machine).Error
	if err != nil {
		return err
	}

	// 记录所有事件名以及构建为符合要求的事件
	names := make([]string, 0)
	desc := make([]fsm.EventDesc, 0)
	// 记录各个事件的层级
	levels := make(map[string]uint, 0)

	// L0申请/L1已拒绝/L0已提交
	l0Name := fmt.Sprintf("%s%s", machine.SubmitterName, SuffixResubmit)
	l0Srcs := []string{
		fmt.Sprintf("%s%s", req[0].Name, SuffixRefused),
	}
	l0Dst := fmt.Sprintf("%s%s", machine.SubmitterName, SuffixSubmitted)
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

	l := len(req)
	for i := 0; i < l; i++ {
		// 通过
		// L1通过/L0已提交,L2已拒绝/L1已通过
		// L2通过/L1已通过/L2已通过
		li1Name := fmt.Sprintf("%s%s", req[i].Name, SuffixWaiting)
		li1Srcs := make([]string, 0)
		if i > 0 {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s%s", req[i-1].Name, SuffixApproved))
		} else {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s%s", machine.SubmitterName, SuffixSubmitted))
		}
		li1Dst := fmt.Sprintf("%s%s", req[i].Name, SuffixApproved)
		if i+1 < l {
			li1Srcs = append(li1Srcs, fmt.Sprintf("%s%s", req[i+1].Name, SuffixRefused))
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

		// 拒绝
		// L1拒绝/L0已提交,L2已拒绝/L1已拒绝
		// L2拒绝/L1已通过/L2已拒绝
		li2Name := fmt.Sprintf("%s%s", req[i].Name, SuffixWaiting)
		li2Srcs := make([]string, 0)
		if i == 0 {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s%s", machine.SubmitterName, SuffixSubmitted))
		} else {
			li2Srcs = append(li2Srcs, fmt.Sprintf("%s%s", req[i-1].Name, SuffixApproved))
			if i+1 < l {
				// 这里的作用是, L1误通过后, L2拒绝, L1再拒绝
				li2Srcs = append(li2Srcs, fmt.Sprintf("%s%s", req[i+1].Name, SuffixRefused))
			}
		}
		li2Dst := fmt.Sprintf("%s%s", req[i].Name, SuffixRefused)
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
		// L0确认/L2已通过/L0已确认
		l0Name := fmt.Sprintf("%s%s", machine.SubmitterName, SuffixConfirm)
		l0Srcs := []string{
			fmt.Sprintf("%s%s", req[l-1].Name, SuffixApproved),
		}
		l0Dst := fmt.Sprintf("%s%s", machine.SubmitterName, SuffixConfirmed)
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

	// 去重
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

		// 默认没有编辑/拒绝权限
		edit := models.Zero
		refuse := models.Zero
		editFields := ""
		roles := make([]Role, 0)
		users := make([]User, 0)
		if i == 0 {
			// 提交人默认有编辑权限
			edit = models.One
			editFields = machine.SubmitterEditFields
		} else if i < len(desc)-1 || machine.SubmitterConfirm == models.Zero {
			// 中间层级
			index := (i+1)/2 - 1
			edit = uint(req[index].Edit)
			editFields = req[index].EditFields
			refuse = uint(req[index].Refuse)
			// 获取全部用户
			roles = fs.getRoles(req[index].Roles)
			users = fs.getUsers(req[index].Users)
		} else if i == len(desc)-1 && machine.SubmitterConfirm == models.One {
			// 提交人确认时编辑字段
			edit = models.One
			editFields = machine.SubmitterConfirmEditFields
		}

		events = append(events, Event{
			MId:        mId,
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

// 获取下一状态名称
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

// 校验状态机是否OK(校验标准: 每个事件都执行过, 只有一个结束位置)
func checkEvents(desc []fsm.EventDesc) error {
	// 记录所有事件名称
	names := make([]string, 0)
	for _, item := range desc {
		names = append(names, item.Dst)
		names = append(names, item.Src...)
	}
	// 去重
	names = utils.RemoveRepeat(names)
	if len(names) == 0 {
		return ErrEventNameNil
	}

	// 创建状态机实例(初始位置可任意设置, 这里为了方便从0开始)
	f := fsm.NewFSM(names[0], desc, nil)
	endCount := 0
	// 遍历设置事件, 看下结束位置有多少个
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
