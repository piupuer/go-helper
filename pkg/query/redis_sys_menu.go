package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/utils"
)

func (rd Redis) FindMenu(currentRoleId, currentRoleSort uint) []ms.SysMenu {
	tree := make([]ms.SysMenu, 0)
	menus := rd.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	tree = rd.GenMenuTree(0, menus)
	return tree
}

// generate menu tree
func (rd Redis) GenMenuTree(parentId uint, roleMenus []ms.SysMenu) []ms.SysMenu {
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
	return genMenuTree(parentId, roleMenuIds, allMenu)
}

func (rd Redis) FindMenuByRoleId(currentRoleId, currentRoleSort, roleId uint) ([]ms.SysMenu, []uint, error) {
	tree := make([]ms.SysMenu, 0)
	accessIds := make([]uint, 0)
	allMenu := rd.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	roleMenus := rd.findMenuByRoleId(roleId)
	tree = rd.GenMenuTree(0, allMenu)
	for _, menu := range roleMenus {
		accessIds = append(accessIds, menu.Id)
	}
	accessIds = FindCheckedMenuId(accessIds, allMenu)
	return tree, accessIds, nil
}

// find all menus by role id(not menu tree)
func (rd Redis) findMenuByRoleId(roleId uint) []ms.SysMenu {
	// query current role menu relation
	relations := make([]ms.SysMenuRoleRelation, 0)
	menuIds := make([]uint, 0)
	rd.
		Table("sys_menu_role_relation").
		Where("role_id", "=", roleId).
		Find(&relations)
	for _, relation := range relations {
		menuIds = append(menuIds, relation.MenuId)
	}
	roleMenu := make([]ms.SysMenu, 0)
	if len(menuIds) > 0 {
		rd.
			Table("sys_menu").
			Where("id", "in", menuIds).
			Order("sort").
			Find(&roleMenu)
	}
	return roleMenu
}

// find all menus by current role(not menu tree)
func (rd Redis) findMenuByCurrentRole(currentRoleId, currentRoleSort uint) []ms.SysMenu {
	menus := make([]ms.SysMenu, 0)
	if currentRoleSort != constant.Zero {
		// find menus by current role id
		menus = rd.findMenuByRoleId(currentRoleId)
	} else {
		// super admin has all menus
		rd.
			Table("sys_menu").
			Order("sort").
			Find(&menus)
	}
	return menus
}
