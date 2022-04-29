package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/tracing"
	"strings"
)

func (my MySql) FindOperationLog(r *req.OperationLog) []ms.SysOperationLog {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindOperationLog"))
	defer span.End()
	list := make([]ms.SysOperationLog, 0)
	q := my.Tx.
		Model(&ms.SysOperationLog{}).
		Order("created_at DESC")
	method := strings.TrimSpace(r.Method)
	if method != "" {
		q.Where("method LIKE ?", fmt.Sprintf("%%%s%%", method))
	}
	path := strings.TrimSpace(r.Path)
	if path != "" {
		q.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	ip := strings.TrimSpace(r.Ip)
	if ip != "" {
		q.Where("ip LIKE ?", fmt.Sprintf("%%%s%%", ip))
	}
	status := strings.TrimSpace(r.Status)
	if status != "" {
		q.Where("status LIKE ?", fmt.Sprintf("%%%s%%", status))
	}
	r.LimitPrimary = constant.QueryPrimaryKey
	my.FindWithPage(q, &r.Page, &list)
	return list
}
