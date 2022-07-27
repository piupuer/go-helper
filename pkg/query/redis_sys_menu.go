package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
)

func (rd Redis) FindMenu(currentRoleId, currentRoleSort uint) (tree []ms.SysMenu) {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "FindMenu"))
	defer span.End()
	menus := rd.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	tree = rd.GenMenuTree(0, menus)
	return
}

// GenMenuTree generate menu tree
func (rd Redis) GenMenuTree(parentId uint, roleMenus []ms.SysMenu) (tree []ms.SysMenu) {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "GenMenuTree"))
	defer span.End()
	roleMenuIds := make([]uint, 0)
	allMenu := make([]ms.SysMenu, 0)
	rd.
		Table("sys_menu").
		Find(&allMenu)
	// add parent menu
	_, newRoleMenus := addParentMenu(roleMenus, allMenu)
	for _, menu := range newRoleMenus {
		if !utils.ContainsUint(roleMenuIds, menu.Id) {
			roleMenuIds = append(roleMenuIds, menu.Id)
		}
	}
	tree = genMenuTree(parentId, roleMenuIds, allMenu)
	return
}

func (rd Redis) FindMenuByRoleId(currentRoleId, currentRoleSort, roleId uint) (tree []ms.SysMenu, accessIds []uint) {
	_, span := tracer.Start(rd.Ctx, tracing.Name(tracing.Cache, "FindMenuByRoleId"))
	defer span.End()
	allMenu := rd.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	roleMenus := rd.findMenuByRoleId(roleId, constant.Zero)
	tree = rd.GenMenuTree(0, allMenu)
	for _, menu := range roleMenus {
		accessIds = append(accessIds, menu.Id)
	}
	accessIds = FindCheckedMenuId(accessIds, allMenu)
	return
}

// find all menus by role id(not menu tree)
func (rd Redis) findMenuByRoleId(roleId, roleSort uint) (rp []ms.SysMenu) {
	// q current role menu relation
	relations := make([]ms.SysMenuRoleRelation, 0)
	menuIds := make([]uint, 0)
	rd.
		Table("sys_menu_role_relation").
		Where("role_id", "=", roleId).
		Find(&relations)
	for _, relation := range relations {
		menuIds = append(menuIds, relation.MenuId)
	}
	rp = make([]ms.SysMenu, 0)
	if len(menuIds) > 0 {
		q := rd.
			Table("sys_menu").
			Where("id", "in", menuIds)
		if roleSort != constant.Zero {
			// normal user check menu status
			q.Where("status", "=", constant.One)
		}
		q.Order("sort").
			Find(&rp)
	}
	return
}

// find all menus by current role(not menu tree)
func (rd Redis) findMenuByCurrentRole(currentRoleId, currentRoleSort uint) (menus []ms.SysMenu) {
	menus = make([]ms.SysMenu, 0)
	if currentRoleSort != constant.Zero {
		// find menus by current role id
		menus = rd.findMenuByRoleId(currentRoleId, currentRoleSort)
	} else {
		// super admin has all menus
		rd.
			Table("sys_menu").
			Order("sort").
			Find(&menus)
	}
	return
}
