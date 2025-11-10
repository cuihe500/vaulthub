package handlers

import (
	"time"

	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// QueryAuditLogs 查询审计日志
// @Summary 查询审计日志
// @Description 查询审计日志，支持多条件过滤和分页。普通用户只能查询自己的日志，管理员可以查询所有用户的日志
// @Tags 审计
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_uuid query string false "用户UUID（管理员可指定，普通用户自动使用当前用户）"
// @Param action_type query string false "操作类型：CREATE/UPDATE/DELETE/ACCESS/LOGIN/LOGOUT"
// @Param resource_type query string false "资源类型：vault/secret/user/config"
// @Param status query string false "操作状态：success/failed"
// @Param start_time query string false "开始时间（RFC3339格式，如2024-01-01T00:00:00Z）"
// @Param end_time query string false "结束时间（RFC3339格式）"
// @Param page query int true "页码" minimum(1)
// @Param page_size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} QueryAuditLogsResponse "查询成功"
// @Failure 400 {string} string "参数错误"
// @Failure 401 {string} string "未授权"
// @Failure 500 {string} string "服务器错误"
// @Router /api/v1/audit/logs [get]
func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	var req service.QueryAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("审计日志查询参数绑定失败", logger.Err(err))
		response.InvalidParam(c, "参数错误")
		return
	}

	// 获取当前用户信息
	currentUserUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	currentRole, _ := middleware.GetCurrentUserRole(c)

	// 权限检查：普通用户只能查询自己的日志
	if currentRole != "admin" {
		// 普通用户，强制使用当前用户UUID
		req.UserUUID = currentUserUUID
	} else if req.UserUUID == "" {
		// 管理员未指定用户UUID，默认查询所有
		// 不做限制
	}

	// 时间范围验证
	if !req.StartTime.IsZero() && !req.EndTime.IsZero() {
		if req.EndTime.Before(req.StartTime) {
			response.InvalidParam(c, "结束时间不能早于开始时间")
			return
		}
	}

	// 查询审计日志
	logs, total, err := h.auditService.QueryLogs(&req)
	if err != nil {
		logger.Error("查询审计日志失败",
			logger.String("user_uuid", currentUserUUID),
			logger.Err(err))
		response.InternalError(c, "查询审计日志失败")
		return
	}

	// 转换为上海时区用于展示
	for _, log := range logs {
		log.CreatedAt = log.CreatedAt.In(time.FixedZone("CST", 8*3600))
	}

	resp := QueryAuditLogsResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Logs:     logs,
	}

	response.Success(c, resp)
}

// QueryAuditLogsResponse 查询审计日志响应
type QueryAuditLogsResponse struct {
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Logs     []*service.AuditLogDTO `json:"logs"`
}
