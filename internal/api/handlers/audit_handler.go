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
// @Description 查询审计日志，支持多条件过滤和分页。普通用户只能查询自己的日志，管理员可以查询所有用户的日志。不传分页参数时全量导出（最多10000条）
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
// @Param page query int false "页码（可选，不传则全量导出）" minimum(1)
// @Param page_size query int false "每页数量（可选，不传则全量导出）" minimum(1) maximum(10000)
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

	// 作用域控制：由ScopeMiddleware统一处理
	// 普通用户：强制只能查询自己的日志
	// 管理员：可以查询指定用户或所有日志
	if scopeUUID, restricted := middleware.GetScopeUserUUID(c); restricted {
		// 作用域受限（普通用户），强制使用受限的用户UUID
		req.UserUUID = scopeUUID
	}
	// 无作用域限制（管理员）且未指定用户UUID，查询所有日志（不做任何限制）

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
			logger.String("user_uuid", req.UserUUID),
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

// ExportStatistics 导出密钥类型统计
// @Summary 导出密钥类型统计
// @Description 导出各类型加密数据的统计总量。普通用户只能查询自己的统计，管理员可以查询指定用户或全局统计
// @Tags 审计
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_uuid query string false "用户UUID（管理员可指定，普通用户自动使用当前用户，管理员不指定则查询全局）"
// @Success 200 {object} service.SecretStatisticsExport "导出成功"
// @Failure 401 {string} string "未授权"
// @Failure 500 {string} string "服务器错误"
// @Router /api/v1/audit/logs/export [get]
func (h *AuditHandler) ExportStatistics(c *gin.Context) {
	// 获取查询参数
	queryUserUUID := c.Query("user_uuid")

	// 作用域控制：由ScopeMiddleware统一处理
	// 普通用户：强制只能查询自己的统计
	// 管理员：可以查询指定用户或全局统计
	targetUserUUID := ""
	if scopeUUID, restricted := middleware.GetScopeUserUUID(c); restricted {
		// 作用域受限（普通用户），强制使用受限的用户UUID
		targetUserUUID = scopeUUID
	} else if queryUserUUID != "" {
		// 无作用域限制（管理员）且指定了用户UUID
		targetUserUUID = queryUserUUID
	}
	// 管理员未指定用户UUID，targetUserUUID为空字符串，表示查询全局统计

	// 导出统计数据
	stats, err := h.auditService.ExportStatistics(targetUserUUID)
	if err != nil {
		logger.Error("导出密钥统计失败",
			logger.String("target_user_uuid", targetUserUUID),
			logger.Err(err))
		response.InternalError(c, "导出统计数据失败")
		return
	}

	response.Success(c, stats)
}

// ExportOperationStatistics 导出操作统计
// @Summary 导出操作统计
// @Description 导出指定时间范围内的操作统计数据（按操作类型分组）。普通用户只能查询自己的统计，管理员可以查询指定用户或全局统计
// @Tags 审计
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_uuid query string false "用户UUID（管理员可指定，普通用户自动使用当前用户，管理员不指定则查询全局）"
// @Param start_time query string false "开始时间（RFC3339格式，如2024-01-01T00:00:00Z）"
// @Param end_time query string false "结束时间（RFC3339格式）"
// @Success 200 {object} service.OperationStatisticsExport "导出成功"
// @Failure 400 {string} string "参数错误"
// @Failure 401 {string} string "未授权"
// @Failure 500 {string} string "服务器错误"
// @Router /api/v1/audit/operations/export [get]
func (h *AuditHandler) ExportOperationStatistics(c *gin.Context) {
	// 获取查询参数
	queryUserUUID := c.Query("user_uuid")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 作用域控制：由ScopeMiddleware统一处理
	// 普通用户：强制只能查询自己的统计
	// 管理员：可以查询指定用户或全局统计
	targetUserUUID := ""
	if scopeUUID, restricted := middleware.GetScopeUserUUID(c); restricted {
		// 作用域受限（普通用户），强制使用受限的用户UUID
		targetUserUUID = scopeUUID
	} else if queryUserUUID != "" {
		// 无作用域限制（管理员）且指定了用户UUID
		targetUserUUID = queryUserUUID
	}
	// 管理员未指定用户UUID，targetUserUUID为空字符串，表示查询全局统计

	// 解析时间参数
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			logger.Warn("解析开始时间失败",
				logger.String("start_time", startTimeStr),
				logger.Err(err))
			response.InvalidParam(c, "开始时间格式错误，请使用RFC3339格式")
			return
		}
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			logger.Warn("解析结束时间失败",
				logger.String("end_time", endTimeStr),
				logger.Err(err))
			response.InvalidParam(c, "结束时间格式错误，请使用RFC3339格式")
			return
		}
	}

	// 时间范围验证
	if !startTime.IsZero() && !endTime.IsZero() && endTime.Before(startTime) {
		response.InvalidParam(c, "结束时间不能早于开始时间")
		return
	}

	// 导出操作统计
	stats, err := h.auditService.ExportOperationStatistics(targetUserUUID, startTime, endTime)
	if err != nil {
		logger.Error("导出操作统计失败",
			logger.String("target_user_uuid", targetUserUUID),
			logger.Err(err))
		response.InternalError(c, "导出操作统计失败")
		return
	}

	response.Success(c, stats)
}

// QueryAuditLogsResponse 查询审计日志响应
type QueryAuditLogsResponse struct {
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	Logs     []*service.AuditLogDTO `json:"logs"`
}
