package handlers

import (
	"strconv"

	"github.com/cuihe500/vaulthub/internal/api/middleware"
	_ "github.com/cuihe500/vaulthub/internal/database/models" // 仅用于Swagger文档类型引用
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// UserProfileHandler 用户档案处理器
type UserProfileHandler struct {
	profileService *service.UserProfileService
}

// NewUserProfileHandler 创建用户档案处理器实例
func NewUserProfileHandler(profileService *service.UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{
		profileService: profileService,
	}
}

// GetProfile 获取当前用户档案信息
// @Summary 获取用户档案信息
// @Description 获取当前登录用户的档案信息
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/profile [get]
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	// 从context获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	profile, err := h.profileService.GetProfile(user.ID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户档案失败", logger.Uint("user_id", user.ID), logger.Err(err))
			response.InternalError(c, "获取用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}

// CreateProfile 创建用户档案
// @Summary 创建用户档案
// @Description 为当前用户创建档案信息
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body service.CreateProfileRequest true "用户档案信息"
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/profile [post]
func (h *UserProfileHandler) CreateProfile(c *gin.Context) {
	// 从context获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req service.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建用户档案请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	profile, err := h.profileService.CreateProfile(user.ID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("创建用户档案失败", logger.Uint("user_id", user.ID), logger.Err(err))
			response.InternalError(c, "创建用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}

// UpdateProfile 更新用户档案
// @Summary 更新用户档案
// @Description 更新当前用户的档案信息
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body service.UpdateProfileRequest true "用户档案信息"
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/profile [put]
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	// 从context获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新用户档案请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	profile, err := h.profileService.UpdateProfile(user.ID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("更新用户档案失败", logger.Uint("user_id", user.ID), logger.Err(err))
			response.InternalError(c, "更新用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}

// CreateOrUpdateProfile 创建或更新用户档案
// @Summary 创建或更新用户档案
// @Description 创建或更新当前用户的档案信息（如果不存在则创建，存在则更新）
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body service.CreateProfileRequest true "用户档案信息"
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/profile [patch]
func (h *UserProfileHandler) CreateOrUpdateProfile(c *gin.Context) {
	// 从context获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	var req service.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建或更新用户档案请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	profile, err := h.profileService.CreateOrUpdateProfile(user.ID, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("创建或更新用户档案失败", logger.Uint("user_id", user.ID), logger.Err(err))
			response.InternalError(c, "创建或更新用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}

// DeleteProfile 删除用户档案
// @Summary 删除用户档案
// @Description 删除当前用户的档案信息
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/profile [delete]
func (h *UserProfileHandler) DeleteProfile(c *gin.Context) {
	// 从context获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	err := h.profileService.DeleteProfile(user.ID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("删除用户档案失败", logger.Uint("user_id", user.ID), logger.Err(err))
			response.InternalError(c, "删除用户档案失败")
		}
		return
	}

	response.Success(c, nil)
}

// ListProfiles 获取用户档案列表（仅管理员）
// @Summary 获取用户档案列表
// @Description 获取用户档案列表（需要管理员权限）。不传分页参数时全量导出（最多10000条）
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码（可选，不传则全量导出）" minimum(1)
// @Param page_size query int false "每页数量（可选，不传则全量导出）" minimum(1) maximum(10000)
// @Param nickname query string false "昵称筛选"
// @Param email query string false "邮箱筛选"
// @Success 200 {object} response.Response{data=service.ListProfilesResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/profiles [get]
func (h *UserProfileHandler) ListProfiles(c *gin.Context) {
	var req service.ListProfilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("用户档案列表请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.profileService.ListProfiles(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户档案列表失败", logger.Err(err))
			response.InternalError(c, "获取用户档案列表失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetUserProfile 获取指定用户档案（仅管理员）
// @Summary 获取指定用户档案
// @Description 根据用户ID获取指定用户的档案信息（需要管理员权限）
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/users/{user_id}/profile [get]
func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ValidationError(c, "无效的用户ID")
		return
	}

	profile, err := h.profileService.GetProfile(uint(userID))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取用户档案失败", logger.Uint("user_id", uint(userID)), logger.Err(err))
			response.InternalError(c, "获取用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}

// UpdateUserProfile 更新指定用户档案（仅管理员）
// @Summary 更新指定用户档案
// @Description 根据用户ID更新指定用户的档案信息（需要管理员权限）
// @Tags 用户档案
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param profile body service.UpdateProfileRequest true "用户档案信息"
// @Success 200 {object} response.Response{data=models.SafeUserProfile}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/users/{user_id}/profile [put]
func (h *UserProfileHandler) UpdateUserProfile(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.ValidationError(c, "无效的用户ID")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新用户档案请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	profile, err := h.profileService.UpdateProfile(uint(userID), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("更新用户档案失败", logger.Uint("user_id", uint(userID)), logger.Err(err))
			response.InternalError(c, "更新用户档案失败")
		}
		return
	}

	response.Success(c, profile)
}
