package handlers

import (
	"time"

	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// StatisticsHandler 统计数据处理器
type StatisticsHandler struct {
	statisticsService *service.StatisticsService
}

// NewStatisticsHandler 创建统计数据处理器
func NewStatisticsHandler(statisticsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

// GetUserStatistics 获取用户统计数据（历史统计）
// @Summary 获取用户统计数据
// @Description 获取用户的历史统计数据，支持时间范围和统计类型过滤。普通用户只能查询自己的统计，管理员可以查询所有用户的统计
// @Tags 统计
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_uuid query string false "用户UUID（管理员可指定，普通用户自动使用当前用户）"
// @Param stat_type query string false "统计类型：daily/weekly/monthly"
// @Param start_date query string false "开始日期（格式：2024-01-01）"
// @Param end_date query string false "结束日期（格式：2024-01-31）"
// @Success 200 {array} github_com_cuihe500_vaulthub_internal_database_models.UserStatistics "查询成功"
// @Failure 400 {string} string "参数错误"
// @Failure 401 {string} string "未授权"
// @Failure 500 {string} string "服务器错误"
// @Router /api/v1/statistics/user [get]
func (h *StatisticsHandler) GetUserStatistics(c *gin.Context) {
	var req service.GetStatisticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("统计数据查询参数绑定失败", logger.Err(err))
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

	// 权限检查：普通用户只能查询自己的统计
	if currentRole != "admin" {
		// 普通用户，强制使用当前用户UUID
		req.UserUUID = currentUserUUID
	} else if req.UserUUID == "" {
		// 管理员未指定用户UUID，默认查询所有（这可能返回大量数据）
		// 建议管理员应该指定用户UUID
		logger.Warn("管理员查询统计数据未指定用户UUID")
	}

	// 时间范围验证
	if !req.StartDate.IsZero() && !req.EndDate.IsZero() {
		if req.EndDate.Before(req.StartDate) {
			response.InvalidParam(c, "结束日期不能早于开始日期")
			return
		}
	}

	// 查询统计数据
	stats, err := h.statisticsService.GetUserStatistics(&req)
	if err != nil {
		logger.Error("查询统计数据失败",
			logger.String("user_uuid", currentUserUUID),
			logger.Err(err))
		response.InternalError(c, "查询统计数据失败")
		return
	}

	// 转换为上海时区用于展示
	for _, stat := range stats {
		stat.StatDate = stat.StatDate.In(time.FixedZone("CST", 8*3600))
		stat.CreatedAt = stat.CreatedAt.In(time.FixedZone("CST", 8*3600))
		stat.UpdatedAt = stat.UpdatedAt.In(time.FixedZone("CST", 8*3600))
	}

	response.Success(c, stats)
}

// GetCurrentStatistics 获取用户当前统计（实时统计）
// @Summary 获取用户当前统计
// @Description 获取用户的实时统计数据（密钥数量、今日操作数等）。普通用户只能查询自己的统计，管理员可以查询指定用户的统计
// @Tags 统计
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_uuid query string false "用户UUID（管理员可指定，普通用户自动使用当前用户）"
// @Success 200 {object} service.CurrentStatistics "查询成功"
// @Failure 401 {string} string "未授权"
// @Failure 500 {string} string "服务器错误"
// @Router /api/v1/statistics/current [get]
func (h *StatisticsHandler) GetCurrentStatistics(c *gin.Context) {
	// 获取查询参数
	queryUserUUID := c.Query("user_uuid")

	// 获取当前用户信息
	currentUserUUID, exists := middleware.GetCurrentUserUUID(c)
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	currentRole, _ := middleware.GetCurrentUserRole(c)

	// 确定要查询的用户UUID
	targetUserUUID := currentUserUUID
	if currentRole == "admin" && queryUserUUID != "" {
		// 管理员可以查询指定用户
		targetUserUUID = queryUserUUID
	}

	// 查询当前统计
	stats, err := h.statisticsService.GetCurrentStatistics(targetUserUUID)
	if err != nil {
		logger.Error("查询当前统计失败",
			logger.String("target_user_uuid", targetUserUUID),
			logger.Err(err))
		response.InternalError(c, "查询当前统计失败")
		return
	}

	response.Success(c, stats)
}
