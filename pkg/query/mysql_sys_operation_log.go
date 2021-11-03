package query

import (
	"fmt"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/req"
	"strings"
)

func (my MySql) FindOperationLog(req *req.OperationLog) []ms.SysOperationLog {
	list := make([]ms.SysOperationLog, 0)
	q := my.Tx.
		Model(&ms.SysOperationLog{}).
		Order("created_at DESC")
	method := strings.TrimSpace(req.Method)
	if method != "" {
		q.Where("method LIKE ?", fmt.Sprintf("%%%s%%", method))
	}
	path := strings.TrimSpace(req.Path)
	if path != "" {
		q.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	ip := strings.TrimSpace(req.Ip)
	if ip != "" {
		q.Where("ip LIKE ?", fmt.Sprintf("%%%s%%", ip))
	}
	status := strings.TrimSpace(req.Status)
	if status != "" {
		q.Where("status LIKE ?", fmt.Sprintf("%%%s%%", status))
	}
	my.FindWithPage(q, &req.Page, &list)
	return list
}
