package service

import (
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"gorm.io/gorm"
)

// MenuService 菜单服务
type MenuService struct {
	db *gorm.DB
}

// NewMenuService 创建菜单服务实例
func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{
		db: db,
	}
}

// GetUserMenus 获取用户可访问的菜单列表
// userRole: 用户角色
// 返回用户有权限访问的菜单列表，按sort_order排序
func (s *MenuService) GetUserMenus(userRole string) ([]*models.MenuItem, error) {
	var menus []models.Menu

	// 查询所有可见菜单，按排序顺序
	if err := s.db.Where("is_visible = ?", true).
		Order("sort_order ASC").
		Find(&menus).Error; err != nil {
		return nil, errors.WithMessage(errors.CodeInternalError, "查询菜单失败", err)
	}

	// 过滤出用户有权限访问的菜单
	var result []*models.MenuItem
	for _, menu := range menus {
		if menu.HasRole(userRole) {
			result = append(result, menu.ToMenuItem())
		}
	}

	return result, nil
}

// GetAllMenus 获取所有菜单（管理员用）
func (s *MenuService) GetAllMenus() ([]*models.MenuItem, error) {
	var menus []models.Menu

	if err := s.db.Order("sort_order ASC").Find(&menus).Error; err != nil {
		return nil, errors.WithMessage(errors.CodeInternalError, "查询菜单失败", err)
	}

	var result []*models.MenuItem
	for _, menu := range menus {
		result = append(result, menu.ToMenuItem())
	}

	return result, nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *models.Menu) error {
	if err := s.db.Create(menu).Error; err != nil {
		return errors.WithMessage(errors.CodeInternalError, "创建菜单失败", err)
	}
	return nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(uuid string, updates map[string]interface{}) error {
	if err := s.db.Model(&models.Menu{}).
		Where("uuid = ?", uuid).
		Updates(updates).Error; err != nil {
		return errors.WithMessage(errors.CodeInternalError, "更新菜单失败", err)
	}
	return nil
}

// DeleteMenu 删除菜单（软删除）
func (s *MenuService) DeleteMenu(uuid string) error {
	if err := s.db.Where("uuid = ?", uuid).Delete(&models.Menu{}).Error; err != nil {
		return errors.WithMessage(errors.CodeInternalError, "删除菜单失败", err)
	}
	return nil
}
