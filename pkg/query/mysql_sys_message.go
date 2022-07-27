package query

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// FindUnDeleteMessage find status!=ms.SysMessageLogStatusDeleted messages
func (my MySql) FindUnDeleteMessage(r *req.Message) []resp.Message {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindUnDeleteMessage"))
	defer span.End()
	sysMessageLogTableName := my.Tx.NamingStrategy.TableName("sys_message_log")
	sysMessageTableName := my.Tx.NamingStrategy.TableName("sys_message")
	list := make([]resp.Message, 0)
	fields := []string{
		fmt.Sprintf("%s.id AS id", sysMessageLogTableName),
		fmt.Sprintf("%s.to_user_id AS to_user_id", sysMessageLogTableName),
		fmt.Sprintf("%s.status AS status", sysMessageLogTableName),
		fmt.Sprintf("%s.type AS type", sysMessageTableName),
		fmt.Sprintf("%s.title AS title", sysMessageTableName),
		fmt.Sprintf("%s.content AS content", sysMessageTableName),
		fmt.Sprintf("%s.created_at AS created_at", sysMessageTableName),
		fmt.Sprintf("%s.from_user_id AS from_user_id", sysMessageTableName),
	}
	q := my.Tx.
		Model(&ms.SysMessageLog{}).
		Select(fields).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.message_id = %s.id", sysMessageTableName, sysMessageLogTableName, sysMessageTableName))

	q.
		Order(fmt.Sprintf("%s.created_at DESC", sysMessageLogTableName)).
		Where(fmt.Sprintf("%s.to_user_id = ?", sysMessageLogTableName), r.ToUserId)
	title := strings.TrimSpace(r.Title)
	if title != "" {
		q.Where(fmt.Sprintf("%s.title LIKE ?", sysMessageTableName), fmt.Sprintf("%%%s%%", title))
	}
	content := strings.TrimSpace(r.Title)
	if content != "" {
		q.Where(fmt.Sprintf("%s.content LIKE ?", sysMessageTableName), fmt.Sprintf("%%%s%%", content))
	}
	if r.Type != nil {
		q.Where("type = ?", *r.Type)
	}
	if r.Status != nil {
		q.Where(fmt.Sprintf("%s.status = ?", sysMessageLogTableName), *r.Status)
	} else {
		// un delete
		q.Where(fmt.Sprintf("%s.status != ?", sysMessageLogTableName), ms.SysMessageLogStatusDeleted)
	}

	// multi tables use ScanWithPage not FindWithPage
	my.ScanWithPage(q, &r.Page, &list)
	return list
}

func (my MySql) GetUnReadMessageCount(userId uint) (total int64) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GetUnReadMessageCount"))
	defer span.End()
	my.Tx.
		Model(&ms.SysMessageLog{}).
		Where("to_user_id = ?", userId).
		Where("status = ?", ms.SysMessageLogStatusUnRead).
		Count(&total)
	return
}

func (my MySql) BatchUpdateMessageRead(messageLogIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchUpdateMessageRead"))
	defer span.End()
	my.BatchUpdateMessageStatus(messageLogIds, ms.SysMessageLogStatusRead)
}

func (my MySql) BatchUpdateMessageDeleted(messageLogIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchUpdateMessageDeleted"))
	defer span.End()
	my.BatchUpdateMessageStatus(messageLogIds, ms.SysMessageLogStatusDeleted)
}

func (my MySql) UpdateAllMessageRead(userId uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateAllMessageRead"))
	defer span.End()
	my.UpdateAllMessageStatus(userId, ms.SysMessageLogStatusRead)
}

func (my MySql) UpdateAllMessageDeleted(userId uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateAllMessageDeleted"))
	defer span.End()
	my.UpdateAllMessageStatus(userId, ms.SysMessageLogStatusDeleted)
}

func (my MySql) BatchUpdateMessageStatus(messageLogIds []uint, status uint) {
	my.Tx.
		Model(&ms.SysMessageLog{}).
		Where("status != ?", ms.SysMessageLogStatusDeleted).
		Where("id IN (?)", messageLogIds).
		Update("status", status)
}

func (my MySql) UpdateAllMessageStatus(userId uint, status uint) {
	var log ms.SysMessageLog
	log.ToUserId = userId
	my.Tx.
		Model(&log).
		Where("status != ?", ms.SysMessageLogStatusDeleted).
		Where(&log).
		Update("status", status)
}

func (my MySql) SyncMessageByUserIds(users []ms.User) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "SyncMessageByUserIds"))
	defer span.End()
	for _, user := range users {
		messages := make([]ms.SysMessage, 0)
		my.Tx.
			// > user register time
			Where("created_at > ?", user.CreatedAt).
			// expire < now
			Where("expired_at > ?", time.Now()).
			// one2many requires consistent roles, system is not required
			Where("(type = ? AND role_id = ?) OR type = ?", ms.SysMessageTypeOneToMany, user.RoleId, ms.SysMessageTypeSystem).
			Find(&messages)
		messageIds := make([]uint, 0)
		for _, message := range messages {
			messageIds = append(messageIds, message.Id)
		}
		// check whether is synced
		logs := make([]ms.SysMessageLog, 0)
		my.Tx.
			Where("to_user_id = ?", user.Id).
			Where("message_id IN (?)", messageIds).
			Find(&logs)
		// old messages
		oldMessageIds := make([]uint, 0)
		for _, log := range logs {
			if !utils.ContainsUint(oldMessageIds, log.MessageId) {
				oldMessageIds = append(oldMessageIds, log.MessageId)
			}
		}
		for _, messageId := range messageIds {
			if !utils.ContainsUint(oldMessageIds, messageId) {
				// need create
				my.Tx.Create(&ms.SysMessageLog{
					ToUserId:  user.Id,
					MessageId: messageId,
				})
			}
		}
	}
	return
}

func (my MySql) CreateMessage(r *req.PushMessage) (err error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateMessage"))
	defer span.End()
	if r.Type != nil {
		message := ms.SysMessage{
			FromUserId: r.FromUserId,
			Title:      r.Title,
			Content:    r.Content,
			Type:       uint(*r.Type),
		}
		switch uint(*r.Type) {
		case ms.SysMessageTypeOneToOne:
			if len(r.ToUserIds) == 0 {
				err = errors.Errorf("to user is empty")
				return
			}
			my.BatchCreateOneToOneMessage(message, r.ToUserIds)
			return
		case ms.SysMessageTypeOneToMany:
			if len(r.ToRoleIds) == 0 {
				err = errors.Errorf("to role is empty")
				return
			}
			my.BatchCreateOneToManyMessage(message, r.ToRoleIds)
			return
		case ms.SysMessageTypeSystem:
			my.CreateSystemMessage(message)
			return
		}
	}
	err = errors.Errorf("message type is illegal")
	return
}

// BatchCreateOneToOneMessage one2one message
func (my MySql) BatchCreateOneToOneMessage(message ms.SysMessage, toIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchCreateOneToOneMessage"))
	defer span.End()
	message.Type = ms.SysMessageTypeOneToOne

	// default expire
	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.DateTime{
			Carbon: carbon.Now().AddDays(30),
		}
	}
	my.Tx.Create(&message)
	// save ToUsers
	for _, id := range toIds {
		var log ms.SysMessageLog
		log.MessageId = message.Id
		log.ToUserId = id
		my.Tx.Create(&log)
	}

	return
}

// BatchCreateOneToManyMessage one2many message
func (my MySql) BatchCreateOneToManyMessage(message ms.SysMessage, toRoleIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchCreateOneToManyMessage"))
	defer span.End()
	message.Type = ms.SysMessageTypeOneToMany

	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.DateTime{
			Carbon: carbon.Now().AddDays(30),
		}
	}

	// save ToRoles
	for _, id := range toRoleIds {
		message.Id = 0
		message.RoleId = id
		my.Tx.Create(&message)
	}
}

// CreateSystemMessage one2all message
func (my MySql) CreateSystemMessage(message ms.SysMessage) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateSystemMessage"))
	defer span.End()
	message.Type = ms.SysMessageTypeSystem

	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.DateTime{
			Carbon: carbon.Now().AddDays(30),
		}
	}

	my.Tx.Create(&message)
}
