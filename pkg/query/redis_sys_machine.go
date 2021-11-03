package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (rd Redis) FindMachine(req *req.Machine) []ms.SysMachine {
	list := make([]ms.SysMachine, 0)
	query := rd.
		Table("sys_machine").
		Order("created_at DESC")
	host := strings.TrimSpace(req.Host)
	if host != "" {
		query = query.Where("host", "contains", host)
	}
	loginName := strings.TrimSpace(req.LoginName)
	if loginName != "" {
		query = query.Where("login_name", "contains", loginName)
	}
	if req.Status != nil {
		query = query.Where("status", "=", *req.Status)
	}
	rd.FindWithPage(query, &req.Page, &list)
	return list
}
