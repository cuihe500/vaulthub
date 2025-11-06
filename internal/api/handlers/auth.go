package handlers

import (
	"strings"

	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService     *service.AuthService
	recoveryService *service.RecoveryService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService, recoveryService *service.RecoveryService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		recoveryService: recoveryService,
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
	if err := h.recoveryService.ResetPasswordWithRecovery(&req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("重置密码失败", logger.Err(err))
			response.InternalError(c, "重置密码失败")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "密码重置成功，请使用新密码登录",
	})
}
