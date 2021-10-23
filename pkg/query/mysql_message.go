package query

import (
	"fmt"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"strings"
	"time"
)

// find status!=ms.SysMessageLogStatusDeleted messages
func (my MySql) FindUnDeleteMessage(req *req.MessageReq) []resp.MessageResp {
	sysMessageLogTableName := my.Tx.NamingStrategy.TableName("sys_message_log")
	sysMessageTableName := my.Tx.NamingStrategy.TableName("sys_message")
	list := make([]resp.MessageResp, 0)
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
	query := my.Tx.
		Model(&ms.SysMessageLog{}).
		Select(fields).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.message_id = %s.id", sysMessageTableName, sysMessageLogTableName, sysMessageTableName))

	query = query.
		Order(fmt.Sprintf("%s.created_at DESC", sysMessageLogTableName)).
		Where(fmt.Sprintf("%s.to_user_id = ?", sysMessageLogTableName), req.ToUserId)
	title := strings.TrimSpace(req.Title)
	if title != "" {
		query = query.Where(fmt.Sprintf("%s.title LIKE ?", sysMessageTableName), fmt.Sprintf("%%%s%%", title))
	}
	content := strings.TrimSpace(req.Title)
	if content != "" {
		query = query.Where(fmt.Sprintf("%s.content LIKE ?", sysMessageTableName), fmt.Sprintf("%%%s%%", content))
	}
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Status != nil {
		query = query.Where(fmt.Sprintf("%s.status = ?", sysMessageLogTableName), *req.Status)
	} else {
		// un delete
		query = query.Where(fmt.Sprintf("%s.status != ?", sysMessageLogTableName), ms.SysMessageLogStatusDeleted)
	}

	// multi tables use ScanWithPage not FindWithPage
	my.ScanWithPage(query, &req.Page, &list)
	return list
}

func (my MySql) GetUnReadMessageCount(userId uint) (int64, error) {
	var total int64
	err := my.Tx.
		Model(&ms.SysMessageLog{}).
		Where("to_user_id = ?", userId).
		Where("status = ?", ms.SysMessageLogStatusUnRead).
		Count(&total).Error
	return total, err
}

func (my MySql) BatchUpdateMessageRead(messageLogIds []uint) error {
	return my.BatchUpdateMessageStatus(messageLogIds, ms.SysMessageLogStatusRead)
}

func (my MySql) BatchUpdateMessageDeleted(messageLogIds []uint) error {
	return my.BatchUpdateMessageStatus(messageLogIds, ms.SysMessageLogStatusDeleted)
}

func (my MySql) UpdateAllMessageRead(userId uint) error {
	return my.UpdateAllMessageStatus(userId, ms.SysMessageLogStatusRead)
}

func (my MySql) UpdateAllMessageDeleted(userId uint) error {
	return my.UpdateAllMessageStatus(userId, ms.SysMessageLogStatusDeleted)
}

func (my MySql) BatchUpdateMessageStatus(messageLogIds []uint, status uint) error {
	return my.Tx.
		Model(&ms.SysMessageLog{}).
		Where("status != ?", ms.SysMessageLogStatusDeleted).
		Where("id IN (?)", messageLogIds).
		Update("status", status).Error
}

func (my MySql) UpdateAllMessageStatus(userId uint, status uint) error {
	var log ms.SysMessageLog
	log.ToUserId = userId
	return my.Tx.
		Model(&log).
		Where("status != ?", ms.SysMessageLogStatusDeleted).
		Where(&log).
		Update("status", status).Error
}

func (my MySql) SyncMessageByUserIds(userIds []uint) error {
	users := make([]ms.SysMessageUser, 0)
	my.Tx.
		Where("id IN (?)", userIds).
		Find(&users)
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
	return nil
}

func (my MySql) CreateMessage(req *req.PushMessageReq) error {
	if req.Type != nil {
		message := ms.SysMessage{
			FromUserId: req.FromUserId,
			Title:      req.Title,
			Content:    req.Content,
			Type:       uint(*req.Type),
		}
		switch uint(*req.Type) {
		case ms.SysMessageTypeOneToOne:
			if len(req.ToUserIds) == 0 {
				return fmt.Errorf("to user is empty")
			}
			return my.BatchCreateOneToOneMessage(message, req.ToUserIds)
		case ms.SysMessageTypeOneToMany:
			if len(req.ToRoleIds) == 0 {
				return fmt.Errorf("to role is empty")
			}
			return my.BatchCreateOneToManyMessage(message, req.ToRoleIds)
		case ms.SysMessageTypeSystem:
			return my.CreateSystemMessage(message)
		}
	}
	return fmt.Errorf("message type is illegal")
}

// one2one message
func (my MySql) BatchCreateOneToOneMessage(message ms.SysMessage, toIds []uint) error {
	message.Type = ms.SysMessageTypeOneToOne

	// default expire
	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.ToDateTimeString{
			Carbon: carbon.Now().AddDays(30),
		}
	}

	err := my.Tx.Create(&message).Error
	if err != nil {
		return err
	}
	// save ToUsers
	for _, id := range toIds {
		var log ms.SysMessageLog
		log.MessageId = message.Id
		log.ToUserId = id
		err = my.Tx.Create(&log).Error
		if err != nil {
			return err
		}
	}

	return err
}

// one2many message
func (my MySql) BatchCreateOneToManyMessage(message ms.SysMessage, toRoleIds []uint) error {
	message.Type = ms.SysMessageTypeOneToMany

	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.ToDateTimeString{
			Carbon: carbon.Now().AddDays(30),
		}
	}

	// save ToRoles
	for _, id := range toRoleIds {
		message.Id = 0
		message.RoleId = id
		err := my.Tx.Create(&message).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// one2all message
func (my MySql) CreateSystemMessage(message ms.SysMessage) error {
	message.Type = ms.SysMessageTypeSystem

	if message.ExpiredAt == nil {
		message.ExpiredAt = &carbon.ToDateTimeString{
			Carbon: carbon.Now().AddDays(30),
		}
	}

	return my.Tx.Create(&message).Error
}
