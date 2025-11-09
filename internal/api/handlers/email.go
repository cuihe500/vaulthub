package handlers

import (
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/gin-gonic/gin"
)

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService *service.EmailService
}

// NewEmailHandler 创建邮件处理器实例
func NewEmailHandler(emailService *service.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Email   string                      `json:"email" binding:"required,email"`
	Purpose service.VerificationPurpose `json:"purpose" binding:"required,oneof=register login reset_password change_email"`
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	Message string `json:"message"`
}

// SendCode 发送验证码
// @Summary 发送邮箱验证码
// @Description 发送邮箱验证码用于注册、登录、重置密码等操作
// @Tags 邮件
// @Accept json
// @Produce json
// @Param request body SendCodeRequest true "发送验证码请求"
// @Success 200 {object} response.Response{data=SendCodeResponse}
// @Router /api/v1/email/send-code [post]
func (h *EmailHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("发送验证码请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 发送验证码
	if err := h.emailService.SendVerificationCode(c.Request.Context(), req.Email, req.Purpose); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("发送验证码失败", logger.Err(err))
			response.InternalError(c, "发送验证码失败")
		}
		return
	}

	response.Success(c, SendCodeResponse{
		Message: "验证码已发送到您的邮箱，请注意查收",
	})
}

// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	Email   string                      `json:"email" binding:"required,email"`
	Code    string                      `json:"code" binding:"required,len=6,numeric"`
	Purpose service.VerificationPurpose `json:"purpose" binding:"required,oneof=register login reset_password change_email"`
}

// VerifyCodeResponse 验证验证码响应
type VerifyCodeResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// VerifyCode 验证验证码
// @Summary 验证邮箱验证码
// @Description 验证邮箱验证码是否正确（验证成功后验证码自动失效）
// @Tags 邮件
// @Accept json
// @Produce json
// @Param request body VerifyCodeRequest true "验证验证码请求"
// @Success 200 {object} response.Response{data=VerifyCodeResponse}
// @Router /api/v1/email/verify-code [post]
func (h *EmailHandler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("验证码验证请求参数无效", logger.Err(err))
		response.ValidationError(c, validator.TranslateError(err))
		return
	}

	// 验证验证码
	if err := h.emailService.VerifyCode(c.Request.Context(), req.Email, req.Purpose, req.Code); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.AppError(c, appErr)
		} else {
			logger.Error("验证验证码失败", logger.Err(err))
			response.InternalError(c, "验证失败")
		}
		return
	}

	response.Success(c, VerifyCodeResponse{
		Valid:   true,
		Message: "验证码验证成功",
	})
}
