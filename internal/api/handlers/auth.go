package handlers

import (
	"strings"

	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService     *service.AuthService
	recoveryService *service.RecoveryService
	db              *gorm.DB
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService, recoveryService *service.RecoveryService, db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		recoveryService: recoveryService,
		db:              db,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号,所有新用户默认为普通用户角色,只有管理员才能提权
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=service.RegisterResponse}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("注册请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("注册失败", logger.Err(err))
			response.InternalError(c, "注册失败")
		}
		return
	}

	response.Success(c, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("登录请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("登录失败", logger.Err(err))
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.Success(c, resp)
}

// LoginWithEmail 邮箱验证码登录
// @Summary 邮箱验证码登录
// @Description 使用邮箱和验证码登录，无需密码。验证码通过 /api/v1/email/send-code 接口获取
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.LoginWithEmailRequest true "邮箱验证码登录请求"
// @Success 200 {object} response.Response{data=service.LoginResponse}
// @Router /api/v1/auth/login-with-email [post]
func (h *AuthHandler) LoginWithEmail(c *gin.Context) {
	var req service.LoginWithEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("邮箱验证码登录请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.authService.LoginWithEmail(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("邮箱验证码登录失败", logger.Err(err))
			response.InternalError(c, "登录失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetMe 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=github_com_cuihe500_vaulthub_internal_database_models.SafeUser}
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	response.Success(c, user.ToSafeUser())
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使当前token失效
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "缺少授权头")
		return
	}

	// 解析Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Unauthorized(c, "授权头格式无效")
		return
	}

	token := parts[1]

	// 调用service层登出
	if err := h.authService.Logout(token); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("登出失败", logger.Err(err))
			response.InternalError(c, "登出失败")
		}
		return
	}

	response.Success(c, nil)
}

// ResetPassword 使用恢复密钥重置密码
// @Summary 使用恢复密钥重置密码
// @Description 用户忘记密码时，可使用注册时获得的24个单词恢复助记词重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.ResetPasswordWithRecoveryRequest true "重置密码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	// 获取当前用户
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		response.Unauthorized(c, "上下文中未找到用户信息")
		return
	}

	var req service.ResetPasswordWithRecoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("重置密码请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 设置用户UUID
	req.UserUUID = user.UUID

	// 调用recovery service
	resp, err := h.recoveryService.ResetPasswordWithRecovery(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("重置密码失败", logger.Err(err))
			response.InternalError(c, "重置密码失败")
		}
		return
	}

	response.Success(c, resp)
}

// RequestPasswordReset 请求密码重置
// @Summary 请求密码重置
// @Description 通过邮箱申请密码重置，系统将发送重置链接到邮箱。无需登录即可访问
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.RequestPasswordResetRequest true "请求密码重置"
// @Success 200 {object} response.Response{data=service.RequestPasswordResetResponse}
// @Router /api/v1/auth/request-password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req service.RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("请求密码重置参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 使用前端传入的 domain 参数构建重置链接
	resp, err := h.authService.RequestPasswordReset(&req, req.Domain)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("请求密码重置失败", logger.Err(err))
			response.InternalError(c, "请求失败")
		}
		return
	}

	response.Success(c, resp)
}

// VerifyPasswordResetToken 验证密码重置token
// @Summary 验证密码重置token
// @Description 验证密码重置token是否有效（未过期且未使用）。无需登录即可访问
// @Tags 认证
// @Accept json
// @Produce json
// @Param token query string true "重置token"
// @Success 200 {object} response.Response{data=service.VerifyResetTokenResponse}
// @Router /api/v1/auth/verify-reset-token [get]
func (h *AuthHandler) VerifyPasswordResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.ValidationError(c, "缺少token参数")
		return
	}

	err := h.authService.VerifyResetToken(token)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("验证token失败", logger.Err(err))
			response.InternalError(c, "验证失败")
		}
		return
	}

	response.Success(c, service.VerifyResetTokenResponse{Valid: true})
}

// ResetPasswordWithToken 使用token重置密码
// @Summary 使用token重置密码
// @Description 使用邮件中的重置token设置新密码。无需登录即可访问
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} response.Response{data=service.ResetPasswordResponse}
// @Router /api/v1/auth/reset-password-with-token [post]
func (h *AuthHandler) ResetPasswordWithToken(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("重置密码请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	resp, err := h.authService.ResetPassword(&req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("重置密码失败", logger.Err(err))
			response.InternalError(c, "重置密码失败")
		}
		return
	}

	response.Success(c, resp)
}

// GetSecurityPINStatus 获取安全密码设置状态
// @Summary 获取安全密码设置状态
// @Description 检查当前用户是否已设置安全密码，用于前端判断是否需要引导用户设置
// @Tags 认证
// @Security BearerAuth
// @Success 200 {object} response.Response "返回 has_security_pin 字段"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/security-pin-status [get]
func (h *AuthHandler) GetSecurityPINStatus(c *gin.Context) {
	// 从上下文获取当前用户UUID（由 AuthMiddleware 设置）
	userUUID, exists := c.Get(middleware.UserUUIDContextKey)
	if !exists {
		logger.Error("无法从上下文获取用户UUID")
		response.Error(c, errors.CodeUnauthorized, "未授权访问")
		return
	}

	// 查询用户的加密密钥配置
	var userKey models.UserEncryptionKey
	err := h.db.Where("user_uuid = ?", userUUID).First(&userKey).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户未创建加密密钥，即未设置安全密码
			response.Success(c, gin.H{
				"has_security_pin": false,
			})
			return
		}
		// 数据库查询错误
		logger.Error("查询用户加密密钥失败",
			logger.String("user_uuid", userUUID.(string)),
			logger.Err(err))
		response.Error(c, errors.CodeDatabaseError, "查询失败")
		return
	}

	// 检查是否已设置安全密码
	response.Success(c, gin.H{
		"has_security_pin": userKey.HasSecurityPIN(),
	})
}
