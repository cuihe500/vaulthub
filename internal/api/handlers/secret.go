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

// SecretHandler 秘密处理器
type SecretHandler struct {
	encryptionService *service.EncryptionService
}

// NewSecretHandler 创建秘密处理器实例
func NewSecretHandler(encryptionService *service.EncryptionService) *SecretHandler {
	return &SecretHandler{
		encryptionService: encryptionService,
	}
}

// CreateEncryptionKey 创建用户加密密钥
// @Summary 创建用户加密密钥
// @Description 为当前用户创建加密密钥（首次使用加密功能时调用）
// @Description 警告：返回的恢复密钥仅显示一次，请务必妥善保管
// @Tags 秘密管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateUserEncryptionKeyRequest true "创建加密密钥请求"
// @Success 200 {object} response.Response{data=service.CreateUserEncryptionKeyResponse}
// @Router /api/v1/encryption/keys [post]
func (h *SecretHandler) CreateEncryptionKey(c *gin.Context) {
	// 获取当前用户UUID
	userUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		logger.Error("无法获取当前用户UUID")
		response.Unauthorized(c, "未授权")
		return
	}

	var req service.CreateUserEncryptionKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建加密密钥请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 使用当前用户的UUID（防止用户伪造其他用户的UUID）
	req.UserUUID = userUUID

	resp, err := h.encryptionService.CreateUserEncryptionKey(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("创建加密密钥失败", logger.Err(err))
			response.InternalError(c, "创建加密密钥失败")
		}
		return
	}

	response.Success(c, resp)
}

// CreateSecret 加密并存储秘密
// @Summary 创建加密秘密
// @Description 加密并存储敏感数据（需要输入密码）
// @Tags 秘密管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.EncryptAndStoreSecretRequest true "创建秘密请求"
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.SafeEncryptedSecret}
// @Router /api/v1/secrets [post]
func (h *SecretHandler) CreateSecret(c *gin.Context) {
	// 获取当前用户UUID
	userUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		logger.Error("无法获取当前用户UUID")
		response.Unauthorized(c, "未授权")
		return
	}

	var req service.EncryptAndStoreSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建秘密请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 使用当前用户的UUID（防止用户伪造其他用户的UUID）
	req.UserUUID = userUUID

	resp, err := h.encryptionService.EncryptAndStoreSecret(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("创建秘密失败", logger.Err(err))
			response.InternalError(c, "创建秘密失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetSecret 解密秘密
// @Summary 解密秘密
// @Description 解密并获取秘密的明文数据（需要输入密码）
// @Tags 秘密管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "秘密UUID"
// @Param request body service.DecryptSecretRequest true "解密请求"
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.DecryptedSecret}
// @Router /api/v1/secrets/{uuid}/decrypt [post]
func (h *SecretHandler) GetSecret(c *gin.Context) {
	// 获取当前用户UUID
	userUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		logger.Error("无法获取当前用户UUID")
		response.Unauthorized(c, "未授权")
		return
	}

	secretUUID := c.Param("uuid")
	if secretUUID == "" {
		response.MissingParam(c, "uuid参数必填")
		return
	}

	var req service.DecryptSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("解密秘密请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 使用当前用户的UUID和URL中的secretUUID（防止用户伪造）
	req.UserUUID = userUUID
	req.SecretUUID = secretUUID

	resp, err := h.encryptionService.DecryptSecret(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("解密秘密失败", logger.Err(err))
			response.InternalError(c, "解密秘密失败")
		}
		return
	}

	response.Success(c, resp)
}

// ListSecrets 获取秘密列表
// @Summary 获取秘密列表
// @Description 获取当前用户的秘密列表（不包含加密数据）。不传分页参数时全量导出（最多10000条）
// @Tags 秘密管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param secret_type query string false "秘密类型" Enums(api_key, db_credential, certificate, ssh_key, token, password, other)
// @Param page query int false "页码（可选，不传则全量导出）" minimum(1)
// @Param page_size query int false "每页数量（可选，不传则全量导出）" minimum(1) maximum(10000)
// @Success 200 {object} response.Response{data=service.ListUserSecretsResponse}
// @Router /api/v1/secrets [get]
func (h *SecretHandler) ListSecrets(c *gin.Context) {
	// 获取当前用户UUID
	userUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		logger.Error("无法获取当前用户UUID")
		response.Unauthorized(c, "未授权")
		return
	}

	var req service.ListUserSecretsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("获取秘密列表请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 使用当前用户的UUID
	req.UserUUID = userUUID

	resp, err := h.encryptionService.ListUserSecrets(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("获取秘密列表失败", logger.Err(err))
			response.InternalError(c, "获取秘密列表失败")
		}
		return
	}

	response.Success(c, resp)
}

// DeleteSecret 删除秘密
// @Summary 删除秘密
// @Description 删除指定的秘密（软删除）
// @Tags 秘密管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "秘密UUID"
// @Success 200 {object} response.Response
// @Router /api/v1/secrets/{uuid} [delete]
func (h *SecretHandler) DeleteSecret(c *gin.Context) {
	// 获取当前用户UUID
	userUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		logger.Error("无法获取当前用户UUID")
		response.Unauthorized(c, "未授权")
		return
	}

	secretUUID := c.Param("uuid")
	if secretUUID == "" {
		response.MissingParam(c, "uuid参数必填")
		return
	}

	err := h.encryptionService.DeleteSecret(userUUID, secretUUID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("删除秘密失败", logger.Err(err))
			response.InternalError(c, "删除秘密失败")
		}
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}
