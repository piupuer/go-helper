package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"strings"
)

// find status!=ms.SysMessageLogStatusDeleted messages
func (rd Redis) FindUnDeleteMessage(req *req.Message) []resp.Message {
	currentUserAllLogs := make([]ms.SysMessageLog, 0)
	rd.
		Table("sys_message_log").
		Preload("Message").
		Where("to_user_id", "=", req.ToUserId).
		// un delete
		Where("status", "!=", ms.SysMessageLogStatusDeleted).
		Find(&currentUserAllLogs)

	messageLogs := make([]ms.SysMessageLog, 0)
	// all log json
	query := rd.
		FromString(utils.Struct2Json(currentUserAllLogs)).
		Order("created_at DESC")
	title := strings.TrimSpace(req.Title)
	if title != "" {
		query = query.Where("message.title", "contains", title)
	}
	content := strings.TrimSpace(req.Content)
	if content != "" {
		query = query.Where("message.content", "contains", content)
	}
	if req.Type != nil {
		query = query.Where("type", "=", *req.Type)
	}
	if req.Status != nil {
		query = query.Where("status", "=", *req.Status)
	}
	rd.FindWithPage(query, &req.Page, &messageLogs)
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
