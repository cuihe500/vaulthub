package handlers

import (
	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// KeyManagementHandler 密钥管理处理器
type KeyManagementHandler struct {
	encryptionService  *service.EncryptionService
	recoveryService    *service.RecoveryService
	keyRotationService *service.KeyRotationService
}

// NewKeyManagementHandler 创建密钥管理处理器实例
func NewKeyManagementHandler(encryptionService *service.EncryptionService, recoveryService *service.RecoveryService, keyRotationService *service.KeyRotationService) *KeyManagementHandler {
	return &KeyManagementHandler{
		encryptionService:  encryptionService,
		recoveryService:    recoveryService,
		keyRotationService: keyRotationService,
	}
}

// CreateUserEncryptionKey 创建用户加密密钥
// @Summary 创建用户加密密钥
// @Description 在用户注册或首次使用加密功能时创建加密密钥，返回24个单词的恢复助记词（仅显示一次）
// @Tags 密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateUserEncryptionKeyRequest true "创建密钥请求"
// @Success 200 {object} response.Response{data=service.CreateUserEncryptionKeyResponse}
// @Router /api/v1/keys/create [post]
func (h *KeyManagementHandler) CreateUserEncryptionKey(c *gin.Context) {
	// 获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	var req service.CreateUserEncryptionKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建密钥请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 设置用户UUID
	req.UserUUID = user.UUID

	// 调用service
	resp, err := h.encryptionService.CreateUserEncryptionKey(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("创建用户加密密钥失败", logger.Err(err))
			response.InternalError(c, "创建用户加密密钥失败")
		}
		return
	}

	response.Success(c, resp)
}

// VerifyRecoveryKey 验证恢复密钥有效性
// @Summary 验证恢复密钥有效性
// @Description 验证用户输入的恢复助记词是否正确，用于在实际重置密码前进行确认
// @Tags 密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.VerifyRecoveryKeyRequest true "验证恢复密钥请求"
// @Success 200 {object} response.Response{data=service.VerifyRecoveryKeyResponse}
// @Router /api/v1/keys/verify-recovery [post]
func (h *KeyManagementHandler) VerifyRecoveryKey(c *gin.Context) {
	// 获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	var req service.VerifyRecoveryKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("验证恢复密钥请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 设置用户UUID
	req.UserUUID = user.UUID

	// 调用service
	resp, err := h.recoveryService.VerifyRecoveryKey(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("验证恢复密钥失败", logger.Err(err))
			response.InternalError(c, "验证恢复密钥失败")
		}
		return
	}

	response.Success(c, resp)
}

// RotateDEK 手动触发密钥轮换
// @Summary 手动触发密钥轮换
// @Description 生成新的数据加密密钥(DEK)并在后台渐进式迁移所有加密数据。注意：每30天最多轮换一次
// @Tags 密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.RotateDEKRequest true "密钥轮换请求"
// @Success 200 {object} response.Response{data=service.RotateDEKResponse}
// @Router /api/v1/keys/rotate [post]
func (h *KeyManagementHandler) RotateDEK(c *gin.Context) {
	// 获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	var req service.RotateDEKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("密钥轮换请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 设置用户UUID
	req.UserUUID = user.UUID

	// 调用service
	resp, err := h.keyRotationService.RotateDEK(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("密钥轮换失败", logger.Err(err))
			response.InternalError(c, "密钥轮换失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetRotationStatus 查询密钥轮换进度
// @Summary 查询密钥轮换进度
// @Description 获取当前用户的密钥轮换状态和数据迁移进度
// @Tags 密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=service.MigrationTask}
// @Router /api/v1/keys/rotation-status [get]
func (h *KeyManagementHandler) GetRotationStatus(c *gin.Context) {
	// 获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	// 调用service
	status, err := h.keyRotationService.GetRotationStatus(user.UUID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("查询轮换状态失败", logger.Err(err))
			response.InternalError(c, "查询轮换状态失败")
		}
		return
	}

	response.Success(c, status)
}
