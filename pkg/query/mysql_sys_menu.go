package query

import (
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
)

// GetMenuTree get menu tree by role id
func (my MySql) GetMenuTree(roleId, roleSort uint) (tree []ms.SysMenu) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GetMenuTree"))
	defer span.End()
	tree = make([]ms.SysMenu, 0)
	// q all menus
	allMenu := make([]ms.SysMenu, 0)
	my.Tx.
		Model(&ms.SysMenu{}).
		Find(&allMenu)
	roleMenu := my.findMenuByRoleId(roleId, roleSort)
	_, newMenus := addParentMenu(roleMenu, allMenu)

	tree = my.GenMenuTree(0, newMenus)
	return
}

func (my MySql) FindMenu(currentRoleId, currentRoleSort uint) (tree []ms.SysMenu) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindMenu"))
	defer span.End()
	tree = make([]ms.SysMenu, 0)
	menus := my.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	tree = my.GenMenuTree(0, menus)
	return
}

// GenMenuTree generate menu tree
func (my MySql) GenMenuTree(parentId uint, roleMenus []ms.SysMenu) (tree []ms.SysMenu) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "GenMenuTree"))
	defer span.End()
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
	tree = genMenuTree(parentId, roleMenuIds, allMenu)
	return
}

func genMenuTree(parentId uint, roleMenuIds []uint, allMenu []ms.SysMenu) (tree []ms.SysMenu) {
	tree = make([]ms.SysMenu, 0)
	for _, menu := range allMenu {
		if !utils.ContainsUint(roleMenuIds, menu.Id) {
			continue
		}
		if menu.ParentId == parentId {
			menu.Children = genMenuTree(menu.Id, roleMenuIds, allMenu)
			tree = append(tree, menu)
		}
	}
	return
}

func (my MySql) FindMenuByRoleId(currentRoleId, currentRoleSort, roleId uint) (tree []ms.SysMenu, accessIds []uint) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindMenuByRoleId"))
	defer span.End()
	allMenu := my.findMenuByCurrentRole(currentRoleId, currentRoleSort)
	roleMenus := my.findMenuByRoleId(roleId, constant.Zero)
	tree = my.GenMenuTree(0, allMenu)
	for _, menu := range roleMenus {
		accessIds = append(accessIds, menu.Id)
	}
	accessIds = FindCheckedMenuId(accessIds, allMenu)
	return
}

func FindCheckedMenuId(list []uint, allMenu []ms.SysMenu) (checked []uint) {
	checked = make([]uint, 0)
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
	return
}

// FindChildrenId find children menu ids
func FindChildrenId(parentId uint, allMenu []ms.SysMenu) (childrenIds []uint) {
	childrenIds = make([]uint, 0)
	for _, menu := range allMenu {
		if menu.ParentId == parentId {
			childrenIds = append(childrenIds, menu.Id)
		}
	}
	return
}

func FindIncrementalMenu(r req.UpdateMenuIncrementalIds, oldMenuIds []uint, allMenu []ms.SysMenu) (rp []uint) {
	createIds := FindCheckedMenuId(r.Create, allMenu)
	deleteIds := FindCheckedMenuId(r.Delete, allMenu)
	rp = make([]uint, 0)
	for _, oldItem := range oldMenuIds {
		// not in delete
		if !utils.Contains(deleteIds, oldItem) {
			rp = append(rp, oldItem)
		}
	}
	// need create
	rp = append(rp, createIds...)
	return
}

func (my MySql) CreateMenu(currentRoleId, currentRoleSort uint, r *req.CreateMenu) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateMenu"))
	defer span.End()
	var menu ms.SysMenu
	utils.Struct2StructByJson(r, &menu)
	my.Tx.Create(&menu)
	menuReq := req.UpdateMenuIncrementalIds{
		Create: []uint{menu.Id},
	}
	my.UpdateMenuByRoleId(currentRoleId, currentRoleSort, currentRoleId, menuReq)
	return
}

func (my MySql) UpdateMenuByRoleId(currentRoleId, currentRoleSort, targetRoleId uint, r req.UpdateMenuIncrementalIds) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "UpdateMenuByRoleId"))
	defer span.End()
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
func (my MySql) findMenuByRoleId(roleId, roleSort uint) (rp []ms.SysMenu) {
	// q current role menu relation
	menuIds := make([]uint, 0)
	my.Tx.
		Model(&ms.SysMenuRoleRelation{}).
		Where("role_id = ?", roleId).
		Pluck("menu_id", &menuIds)
	rp = make([]ms.SysMenu, 0)
	if len(menuIds) > 0 {
		q := my.Tx.
			Model(&ms.SysMenu{}).
			Where("id IN (?)", menuIds)
		if roleSort != constant.Zero {
			// normal user check menu status
			q.Where("status = ?", constant.One)
		}
		q.Order("sort").
			Find(&rp)
	}
	return
}

// find all menus by current role(not menu tree)
func (my MySql) findMenuByCurrentRole(currentRoleId, currentRoleSort uint) (menus []ms.SysMenu) {
	menus = make([]ms.SysMenu, 0)
	if currentRoleSort != constant.Zero {
		// find menus by current role id
		menus = my.findMenuByRoleId(currentRoleId, currentRoleSort)
	} else {
		// super admin has all menus
		my.Tx.
			Order("sort").
			Find(&menus)
	}
	return
}

func addParentMenu(menus, all []ms.SysMenu) (newMenuIds []uint, newMenus []ms.SysMenu) {
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
	newMenuIds = make([]uint, 0)
	newMenus = make([]ms.SysMenu, 0)
	for _, menu := range all {
		for _, id := range menuIds {
			if id == menu.Id && !utils.ContainsUint(newMenuIds, id) {
				newMenus = append(newMenus, menu)
				newMenuIds = append(newMenuIds, id)
			}
		}
	}
	return
}

// find parent menu ids
func findParentMenuId(menuId uint, all []ms.SysMenu) (parentIds []uint) {
	var currentMenu ms.SysMenu
	parentIds = make([]uint, 0)
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
	return
}
