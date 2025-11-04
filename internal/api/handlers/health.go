package handlers

import (
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// HealthData health check 响应数据
type HealthData struct {
	Status   string `json:"status" example:"healthy"`     // 服务状态: healthy, unhealthy
	Database string `json:"database" example:"connected"` // 数据库状态: connected, disconnected
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Code      int        `json:"code" example:"0"`                                         // 响应码，0表示成功
	Message   string     `json:"message" example:"success"`                                // 响应消息
	Data      HealthData `json:"data"`                                                     // 健康检查数据
	RequestID string     `json:"requestId" example:"b68c086f-db16-43f6-992e-bc391afbf24a"` // 请求ID
	Timestamp int64      `json:"timestamp" example:"1762269490888"`                        // 时间戳（毫秒）
}

// HealthHandler 健康检查处理器
type HealthHandler struct {
	mgr *app.Manager
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(mgr *app.Manager) *HealthHandler {
	return &HealthHandler{mgr: mgr}
}

// HealthCheck 健康检查接口
// @Summary 健康检查
// @Description 检查服务及其依赖（数据库等）的运行状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "服务健康"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	healthData := HealthData{
		Status:   "healthy",
		Database: "disconnected",
	}

	// 检查数据库连接
	if h.mgr.DB != nil {
		sqlDB, err := h.mgr.DB.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				healthData.Database = "connected"
			}
		}
	}

	// 如果数据库未连接，服务状态为 unhealthy
	if healthData.Database != "connected" {
		healthData.Status = "unhealthy"
	}

	response.Success(c, healthData)
}
