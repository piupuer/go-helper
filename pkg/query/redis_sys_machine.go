package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (rd Redis) FindMachine(req *req.Machine) []ms.SysMachine {
	list := make([]ms.SysMachine, 0)
	q := rd.
		Table("sys_machine").
		Order("created_at DESC")
	host := strings.TrimSpace(req.Host)
	if host != "" {
		q.Where("host", "contains", host)
	}
	loginName := strings.TrimSpace(req.LoginName)
	if loginName != "" {
		q.Where("login_name", "contains", loginName)
	}
	if req.Status != nil {
		q.Where("status", "=", *req.Status)
	}
	rd.FindWithPage(q, &req.Page, &list)
	return list
}
