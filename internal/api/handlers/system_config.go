package handlers

import (
	_ "github.com/cuihe500/vaulthub/internal/database/models" // 仅用于Swagger文档类型引用
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// SystemConfigHandler 系统配置处理器
type SystemConfigHandler struct {
	configService *service.SystemConfigService
}

// NewSystemConfigHandler 创建系统配置处理器实例
func NewSystemConfigHandler(configService *service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{
		configService: configService,
	}
}

// ListConfigs 获取所有系统配置
// @Summary 获取系统配置列表
// @Description 获取所有系统配置项（管理员权限）
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.ListConfigsResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs [get]
func (h *SystemConfigHandler) ListConfigs(c *gin.Context) {
	// 需要管理员权限
	// 权限验证在路由层通过Casbin中间件完成

	result, err := h.configService.ListConfigs()
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			logger.Warn("获取配置列表失败", logger.String("code", string(appErr.Code)))
			response.AppError(c, appErr)
			return
		}
		logger.Error("获取配置列表失败", logger.Err(err))
		response.InternalError(c, "获取配置列表失败")
		return
	}

	response.Success(c, result)
}

// GetConfig 获取单个系统配置
// @Summary 获取单个系统配置
// @Description 根据配置键获取配置详情（管理员权限）
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Success 200 {object} response.Response{data=service.ConfigItem}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs/{key} [get]
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	// 需要管理员权限
	// 权限验证在路由层通过Casbin中间件完成

	key := c.Param("key")
	if key == "" {
		response.InvalidParam(c, "配置键不能为空")
		return
	}

	result, err := h.configService.GetConfig(key)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			logger.Warn("获取配置失败",
				logger.String("key", key),
				logger.String("code", string(appErr.Code)))
			response.AppError(c, appErr)
			return
		}
		logger.Error("获取配置失败", logger.String("key", key), logger.Err(err))
		response.InternalError(c, "获取配置失败")
		return
	}

	response.Success(c, result)
}

// UpdateConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新指定配置项的值（管理员权限）。配置更新后会立即生效并触发相关观察者回调
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "配置键"
// @Param request body service.UpdateConfigRequest true "更新配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs/{key} [put]
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	// 需要管理员权限
	// 权限验证在路由层通过Casbin中间件完成

	key := c.Param("key")
	if key == "" {
		response.InvalidParam(c, "配置键不能为空")
		return
	}

	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新配置请求参数错误",
			logger.String("key", key),
			logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	if err := h.configService.UpdateConfig(key, &req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			logger.Warn("更新配置失败",
				logger.String("key", key),
				logger.String("code", string(appErr.Code)))
			response.AppError(c, appErr)
			return
		}
		logger.Error("更新配置失败", logger.String("key", key), logger.Err(err))
		response.InternalError(c, "更新配置失败")
		return
	}

	response.Success(c, gin.H{"message": "配置更新成功"})
}

// BatchUpdateConfigs 批量更新系统配置
// @Summary 批量更新系统配置
// @Description 批量更新多个配置项（管理员权限）。所有配置在同一事务中更新，全部成功或全部失败
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.BatchUpdateConfigRequest true "批量更新配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs/batch [put]
func (h *SystemConfigHandler) BatchUpdateConfigs(c *gin.Context) {
	// 需要管理员权限
	// 权限验证在路由层通过Casbin中间件完成

	var req service.BatchUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("批量更新配置请求参数错误", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	if err := h.configService.BatchUpdateConfigs(&req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			logger.Warn("批量更新配置失败", logger.String("code", string(appErr.Code)))
			response.AppError(c, appErr)
			return
		}
		logger.Error("批量更新配置失败", logger.Err(err))
		response.InternalError(c, "批量更新配置失败")
		return
	}

	response.Success(c, gin.H{"message": "批量更新配置成功"})
}

// ReloadConfigs 重新加载配置
// @Summary 重新加载配置
// @Description 从数据库重新加载所有配置到内存（管理员权限）
// @Tags 系统配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs/reload [post]
func (h *SystemConfigHandler) ReloadConfigs(c *gin.Context) {
	// 需要管理员权限
	// 权限验证在路由层通过Casbin中间件完成

	if err := h.configService.ReloadConfigs(); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			logger.Warn("重新加载配置失败", logger.String("code", string(appErr.Code)))
			response.AppError(c, appErr)
			return
		}
		logger.Error("重新加载配置失败", logger.Err(err))
		response.InternalError(c, "重新加载配置失败")
		return
	}

	response.Success(c, gin.H{"message": "配置重新加载成功"})
}
