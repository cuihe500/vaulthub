package handlers

import (
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表（需要管理员权限）。不传分页参数时全量导出（最多10000条）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码（可选，不传则全量导出）" minimum(1)
// @Param page_size query int false "每页数量（可选，不传则全量导出）" minimum(1) maximum(10000)
// @Param status query int false "用户状态" Enums(1, 2, 3)
// @Param role query string false "用户角色" Enums(admin, user, readonly)
// @Success 200 {object} response.Response{data=service.ListUsersResponse}
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req service.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("用户列表请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.userService.ListUsers(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户列表失败", logger.Err(err))
			response.InternalError(c, "获取用户列表失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetUser 获取单个用户信息
// @Summary 获取用户信息
// @Description 根据UUID获取用户详细信息（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "用户UUID"
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.SafeUser}
// @Router /api/v1/users/{uuid} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userUUID := c.Param("uuid")
	if userUUID == "" {
		response.MissingParam(c, "uuid参数必填")
		return
	}

	user, err := h.userService.GetUserByUUID(userUUID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户信息失败", logger.String("uuid", userUUID), logger.Err(err))
			response.InternalError(c, "获取用户信息失败")
		}
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新用户状态（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "用户UUID"
// @Param request body service.UpdateUserStatusRequest true "更新状态请求"
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.SafeUser}
// @Router /api/v1/users/{uuid}/status [put]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userUUID := c.Param("uuid")
	if userUUID == "" {
		response.MissingParam(c, "uuid参数必填")
		return
	}

	var req service.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新用户状态请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	user, err := h.userService.UpdateUserStatus(userUUID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("更新用户状态失败", logger.String("uuid", userUUID), logger.Err(err))
			response.InternalError(c, "更新用户状态失败")
		}
		return
	}

	response.Success(c, user)
}

// UpdateUserRole 更新用户角色
// @Summary 更新用户角色
// @Description 更新用户角色（需要管理员权限）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "用户UUID"
// @Param request body service.UpdateUserRoleRequest true "更新角色请求"
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.SafeUser}
// @Router /api/v1/users/{uuid}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	userUUID := c.Param("uuid")
	if userUUID == "" {
		response.MissingParam(c, "uuid参数必填")
		return
	}

	var req service.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新用户角色请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	user, err := h.userService.UpdateUserRole(userUUID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("更新用户角色失败", logger.String("uuid", userUUID), logger.Err(err))
			response.InternalError(c, "更新用户角色失败")
		}
		return
	}

	response.Success(c, user)
}
