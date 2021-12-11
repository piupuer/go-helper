package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"strings"
)

// find status!=ms.SysMessageLogStatusDeleted messages
func (rd Redis) FindUnDeleteMessage(r *req.Message) []resp.Message {
	currentUserAllLogs := make([]ms.SysMessageLog, 0)
	rd.
		Table("sys_message_log").
		Preload("Message").
		Where("to_user_id", "=", r.ToUserId).
		// un delete
		Where("status", "!=", ms.SysMessageLogStatusDeleted).
		Find(&currentUserAllLogs)

	messageLogs := make([]ms.SysMessageLog, 0)
	// all log json
	q := rd.
		FromString(utils.Struct2Json(currentUserAllLogs)).
		Order("created_at DESC")
	title := strings.TrimSpace(r.Title)
	if title != "" {
		q.Where("message.title", "contains", title)
	}
	content := strings.TrimSpace(r.Content)
	if content != "" {
		q.Where("message.content", "contains", content)
	}
	if r.Type != nil {
		q.Where("type", "=", *r.Type)
	}
	if r.Status != nil {
		q.Where("status", "=", *r.Status)
	}
	rd.FindWithPage(q, &r.Page, &messageLogs)
	// convert to Message
	list := make([]resp.Message, 0)
	for _, log := range messageLogs {
		res := resp.Message{
			Base: resp.Base{
				Id:        log.Id,
				CreatedAt: log.CreatedAt,
				UpdatedAt: log.UpdatedAt,
			},
			Status:     log.Status,
			ToUserId:   log.ToUserId,
			Type:       log.Message.Type,
			Title:      log.Message.Title,
			Content:    log.Message.Content,
			FromUserId: log.Message.FromUserId,
		}
		list = append(list, res)
	}

	return list
}

// un read total count
func (rd Redis) GetUnReadMessageCount(userId uint) (int64, error) {
	var total int64
	err := rd.
		Table("sys_message_log").
		Where("to_user_id", "=", userId).
		Where("status", "=", ms.SysMessageLogStatusUnRead).
		Count(&total).Error
	return total, err
}
