package handlers

import (
	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	menuService *service.MenuService
}

// NewMenuHandler 创建菜单处理器实例
func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

// GetUserMenus 获取当前用户可访问的菜单列表
// @Summary 获取用户菜单
// @Description 根据当前用户角色返回可访问的菜单列表
// @Tags 菜单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]github_com_cuihe500_vaulthub_internal_database_models.MenuItem}
// @Router /api/v1/menus [get]
func (h *MenuHandler) GetUserMenus(c *gin.Context) {
	// 从context获取用户角色
	role, exists := middleware.GetCurrentUserRole(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户角色")
		return
	}

	menus, err := h.menuService.GetUserMenus(role)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户菜单失败", logger.Err(err))
			response.InternalError(c, "获取菜单失败")
		}
		return
	}

	response.Success(c, menus)
}

// GetAllMenus 获取所有菜单（管理员用）
// @Summary 获取所有菜单
// @Description 获取所有菜单配置（管理员权限）
// @Tags 菜单
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]github_com_cuihe500_vaulthub_internal_database_models.MenuItem}
// @Router /api/v1/menus/all [get]
func (h *MenuHandler) GetAllMenus(c *gin.Context) {
	menus, err := h.menuService.GetAllMenus()
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取所有菜单失败", logger.Err(err))
			response.InternalError(c, "获取菜单失败")
		}
		return
	}

	response.Success(c, menus)
}
