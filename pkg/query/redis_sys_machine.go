package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/tracing"
	"strings"
)

func (rd Redis) FindMachine(r *req.Machine) []ms.SysMachine {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "FindMachine"))
	defer span.End()
	list := make([]ms.SysMachine, 0)
	q := rd.
		Table("sys_machine").
		Order("created_at DESC")
	host := strings.TrimSpace(r.Host)
	if host != "" {
		q.Where("host", "contains", host)
	}
	loginName := strings.TrimSpace(r.LoginName)
	if loginName != "" {
		q.Where("login_name", "contains", loginName)
	}
	if r.Status != nil {
		q.Where("status", "=", *r.Status)
	}
	rd.FindWithPage(q, &r.Page, &list)
	return list
}
