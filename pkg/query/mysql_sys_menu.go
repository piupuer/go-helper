package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/utils"
)

// get menu tree by role id
func (my MySql) GetMenuTree(roleId, roleSort uint) ([]ms.SysMenu, error) {
	tree := make([]ms.SysMenu, 0)
	// q all menus
	allMenu := make([]ms.SysMenu, 0)
	my.Tx.
		Model(&ms.SysMenu{}).
		Find(&allMenu)
	roleMenu := my.findMenuByRoleId(roleId, roleSort)
	_, newMenus := addParentMenu(roleMenu, allMenu)

	tree = my.GenMenuTree(0, newMenus)
	return tree, nil
}

func (my MySql) FindMenu(currentRoleId, currentRoleSort uint) []ms.SysMenu {
	tree := make([]ms.SysMenu, 0)
	menus := my.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	tree = my.GenMenuTree(0, menus)
	return tree
}

// generate menu tree
func (my MySql) GenMenuTree(parentId uint, roleMenus []ms.SysMenu) []ms.SysMenu {
	roleMenuIds := make([]uint, 0)
	allMenu := make([]ms.SysMenu, 0)
	my.Tx.
		Model(&ms.SysMenu{}).
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

func genMenuTree(parentId uint, roleMenuIds []uint, allMenu []ms.SysMenu) []ms.SysMenu {
	tree := make([]ms.SysMenu, 0)
	for _, menu := range allMenu {
		if !utils.ContainsUint(roleMenuIds, menu.Id) {
			continue
		}
		if menu.ParentId == parentId {
			menu.Children = genMenuTree(menu.Id, roleMenuIds, allMenu)
			tree = append(tree, menu)
		}
	}
	return tree
}

func (my MySql) FindMenuByRoleId(currentRoleId, currentRoleSort, roleId uint) ([]ms.SysMenu, []uint, error) {
	tree := make([]ms.SysMenu, 0)
	accessIds := make([]uint, 0)
	allMenu := my.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	roleMenus := my.findMenuByRoleId(roleId, constant.Zero)
	tree = my.GenMenuTree(0, allMenu)
	for _, menu := range roleMenus {
		accessIds = append(accessIds, menu.Id)
	}
	accessIds = FindCheckedMenuId(accessIds, allMenu)
	return tree, accessIds, nil
}

func FindCheckedMenuId(list []uint, allMenu []ms.SysMenu) []uint {
	checked := make([]uint, 0)
	for _, c := range list {
		children := FindChildrenId(c, allMenu)
		count := 0
		for _, child := range children {
			contains := false
			for _, v := range list {
				if v == child {
					contains = true
				}
			}
			if contains {
				count++
			}
		}
		if len(children) == count {
			// all checked
			checked = append(checked, c)
		}
	}
	return checked
}

// find children menu ids
func FindChildrenId(parentId uint, allMenu []ms.SysMenu) []uint {
	childrenIds := make([]uint, 0)
	for _, menu := range allMenu {
		if menu.ParentId == parentId {
			childrenIds = append(childrenIds, menu.Id)
		}
	}
	return childrenIds
}

func FindIncrementalMenu(r req.UpdateMenuIncrementalIds, oldMenuIds []uint, allMenu []ms.SysMenu) []uint {
	createIds := FindCheckedMenuId(r.Create, allMenu)
	deleteIds := FindCheckedMenuId(r.Delete, allMenu)
	newList := make([]uint, 0)
	for _, oldItem := range oldMenuIds {
		// not in delete
		if !utils.Contains(deleteIds, oldItem) {
			newList = append(newList, oldItem)
		}
	}
	// need create
	return append(newList, createIds...)
}

func (my MySql) CreateMenu(currentRoleId, currentRoleSort uint, r *req.CreateMenu) (err error) {
	var menu ms.SysMenu
	utils.Struct2StructByJson(r, &menu)
	err = my.Tx.Create(&menu).Error
	menuReq := req.UpdateMenuIncrementalIds{
		Create: []uint{menu.Id},
	}
	err = my.UpdateMenuByRoleId(currentRoleId, currentRoleSort, currentRoleId, menuReq)
	return
}

func (my MySql) UpdateMenuByRoleId(currentRoleId, currentRoleSort, targetRoleId uint, r req.UpdateMenuIncrementalIds) (err error) {
	allMenu := my.FindMenu(currentRoleId, currentRoleSort)
	roleMenus := my.findMenuByRoleId(targetRoleId, constant.Zero)
	menuIds := make([]uint, 0)
	for _, menu := range roleMenus {
		menuIds = append(menuIds, menu.Id)
	}
	incremental := FindIncrementalMenu(r, menuIds, allMenu)
	incrementalMenus := make([]ms.SysMenu, 0)
	my.Tx.
		Model(&ms.SysMenu{}).
		Where("id in (?)", incremental).
		Find(&incrementalMenus)
	newRelations := make([]ms.SysMenuRoleRelation, 0)
	for _, menu := range incrementalMenus {
		newRelations = append(newRelations, ms.SysMenuRoleRelation{
			MenuId: menu.Id,
			RoleId: targetRoleId,
		})
	}
	my.Tx.
		Where("role_id = ?", targetRoleId).
		Delete(&ms.SysMenuRoleRelation{})
	my.Tx.
		Model(&ms.SysMenuRoleRelation{}).
		Create(&newRelations)
	return
}

// find all menus by role id(not menu tree)
func (my MySql) findMenuByRoleId(roleId, roleSort uint) []ms.SysMenu {
	// q current role menu relation
	menuIds := make([]uint, 0)
	my.Tx.
		Model(&ms.SysMenuRoleRelation{}).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &menuIds)
	roleMenu := make([]ms.SysMenu, 0)
	if len(menuIds) > 0 {
		q := my.Tx.
			Model(&ms.SysMenu{}).
			Where("id IN (?)", menuIds)
		if roleSort != constant.Zero {
			// normal user check menu status
			q.Where("status = ?", constant.One)
		}
		q.Order("sort").
			Find(&roleMenu)
	}
	return roleMenu
}

// find all menus by current role(not menu tree)
func (my MySql) findMenuByCurrentRole(currentRoleId, currentRoleSort uint) []ms.SysMenu {
	menus := make([]ms.SysMenu, 0)
	if currentRoleSort != constant.Zero {
		// find menus by current role id
		menus = my.findMenuByRoleId(currentRoleId, currentRoleSort)
	} else {
		// super admin has all menus
		my.Tx.
			Order("sort").
			Find(&menus)
	}
	return menus
}

func addParentMenu(menus, all []ms.SysMenu) ([]uint, []ms.SysMenu) {
	parentIds := make([]uint, 0)
	menuIds := make([]uint, 0)
	for _, menu := range menus {
		if menu.ParentId > 0 {
			parentIds = append(parentIds, menu.ParentId)
			// find parent menu
			parentMenuIds := findParentMenuId(menu.ParentId, all)
			if len(parentMenuIds) > 0 {
				parentIds = append(parentIds, parentMenuIds...)
			}
		}
		menuIds = append(menuIds, menu.Id)
	}
	// merge parent menu
	if len(parentIds) > 0 {
		menuIds = append(menuIds, parentIds...)
	}
	newMenuIds := make([]uint, 0)
	newMenus := make([]ms.SysMenu, 0)
	for _, menu := range all {
		for _, id := range menuIds {
			if id == menu.Id && !utils.ContainsUint(newMenuIds, id) {
				newMenus = append(newMenus, menu)
				newMenuIds = append(newMenuIds, id)
			}
		}
	}
	return newMenuIds, newMenus
}

// find parent menu ids
func findParentMenuId(menuId uint, all []ms.SysMenu) []uint {
	var currentMenu ms.SysMenu
	parentIds := make([]uint, 0)
	for _, menu := range all {
		if menuId == menu.Id {
			currentMenu = menu
			break
		}
	}
	if currentMenu.ParentId == 0 {
		return parentIds
	}
	parentIds = append(parentIds, currentMenu.ParentId)
	newParentIds := findParentMenuId(currentMenu.ParentId, all)
	if len(newParentIds) > 0 {
		parentIds = append(parentIds, newParentIds...)
	}
	return parentIds
}
