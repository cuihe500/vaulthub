package handlers

import (
	"context"
	"runtime"
	"time"

	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/cuihe500/vaulthub/pkg/version"
	"github.com/gin-gonic/gin"
)

// ComponentStatus 组件状态
type ComponentStatus struct {
	Name        string    `json:"name" example:"database"`                  // 组件名称
	Status      string    `json:"status" example:"healthy"`                 // 状态：healthy, degraded, unhealthy
	latency     time.Duration `json:"-"`                                    // 延迟，不序列化到JSON
	Message     string    `json:"message,omitempty" example:"数据库连接正常"` // 状态描述
	LastChecked time.Time `json:"last_checked" example:"2025-11-06T10:30:00+08:00"` // 最后检查时间
}

// HealthData health check 响应数据
type HealthData struct {
	Status    string                     `json:"status" example:"healthy"`     // 整体状态：healthy, degraded, unhealthy
	Timestamp int64                      `json:"timestamp" example:"1762269490888"` // 检查时间戳（毫秒）
	Uptime    int64                      `json:"uptime" example:"3600000"`     // 服务运行时间（毫秒）
	Version   string                     `json:"version" example:"dev"`        // 服务版本
	Components map[string]ComponentStatus `json:"components"`                   // 各组件状态详情
	System    SystemInfo                 `json:"system"`                       // 系统信息
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version" example:"go1.25.1"`     // Go版本
	Goroutines   int    `json:"goroutines" example:"10"`           // 当前goroutine数量
	MemoryUsed   uint64 `json:"memory_used" example:"5242880"`     // 内存使用量（字节）
	NumCPU       int    `json:"num_cpu" example:"8"`               // CPU核心数
	NumGoroutine int    `json:"num_goroutine" example:"10"`        // Goroutine数量
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
	mgr     *app.Manager
	startTime time.Time // 服务启动时间
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(mgr *app.Manager) *HealthHandler {
	return &HealthHandler{
		mgr:        mgr,
		startTime: time.Now(),
	}
}

// HealthCheck 健康检查接口
// @Summary 健康检查
// @Description 检查服务及其依赖（数据库、Redis、Casbin权限系统等）的运行状态。返回整体健康状态、各组件详细状态、系统资源使用情况和服务运行时间。此接口无需认证，用于监控系统健康状态。
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "健康检查响应。code=0表示服务健康，code=70001表示服务降级，code=70002表示服务不可用"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// 设置超时上下文，避免健康检查阻塞
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	now := time.Now()
	healthData := HealthData{
		Status:     "healthy",
		Timestamp:  now.UnixMilli(),
		Uptime:     now.Sub(h.startTime).Milliseconds(),
		Version:    version.Version,
		Components: make(map[string]ComponentStatus),
		System:     h.getSystemInfo(),
	}

	// 并发检查各组件状态，使用buffered channel避免goroutine阻塞
	componentChan := make(chan ComponentStatus, 3)

	// 检查数据库
	go h.checkDatabase(ctx, componentChan)

	// 检查Redis
	go h.checkRedis(ctx, componentChan)

	// 检查Casbin权限系统
	go h.checkCasbin(ctx, componentChan)

	// 收集组件检查结果
CollectLoop:
	for i := 0; i < 3; i++ {
		select {
		case status := <-componentChan:
			healthData.Components[status.Name] = status
			// 根据组件状态调整整体状态
			if status.Status == "unhealthy" {
				healthData.Status = "unhealthy"
			} else if status.Status == "degraded" && healthData.Status == "healthy" {
				healthData.Status = "degraded"
			}
		case <-ctx.Done():
			logger.Error("健康检查超时")
			healthData.Status = "degraded"
			break CollectLoop
		}
	}

	// 显式关闭channel，所有goroutine已经完成
	close(componentChan)

	// 根据整体状态设置响应码和消息
	var respCode int
	var respMessage string
	switch healthData.Status {
	case "healthy":
		respCode = errors.CodeSuccess
		respMessage = "服务健康"
	case "degraded":
		respCode = errors.CodeServiceDegraded
		respMessage = "服务降级"
	case "unhealthy":
		respCode = errors.CodeServiceDown
		respMessage = "服务不可用"
	}

	response.SuccessWithCode(c, respCode, respMessage, healthData)
}

// checkDatabase 检查数据库连接状态
func (h *HealthHandler) checkDatabase(ctx context.Context, resultChan chan<- ComponentStatus) {
	start := time.Now()
	status := ComponentStatus{
		Name:        "database",
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	defer func() {
		status.latency = time.Since(start)
		select {
		case resultChan <- status:
		case <-ctx.Done():
		}
	}()

	if h.mgr.DB == nil {
		status.Message = "数据库连接未初始化"
		return
	}

	sqlDB, err := h.mgr.DB.DB()
	if err != nil {
		status.Message = "获取数据库连接失败: " + err.Error()
		return
	}

	// 使用ping检查连接
	if err := sqlDB.PingContext(ctx); err != nil {
		status.Message = "数据库连接失败: " + err.Error()
		return
	}

	status.Status = "healthy"
	status.Message = "数据库连接正常"
}

// checkRedis 检查Redis连接状态
func (h *HealthHandler) checkRedis(ctx context.Context, resultChan chan<- ComponentStatus) {
	start := time.Now()
	status := ComponentStatus{
		Name:        "redis",
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	defer func() {
		status.latency = time.Since(start)
		select {
		case resultChan <- status:
		case <-ctx.Done():
		}
	}()

	if h.mgr.Redis == nil {
		status.Message = "Redis连接未初始化"
		return
	}

	// 使用ping检查Redis连接
	if err := h.mgr.Redis.Ping(ctx); err != nil {
		status.Message = "Redis连接失败: " + err.Error()
		return
	}

	status.Status = "healthy"
	status.Message = "Redis连接正常"
}

// checkCasbin 检查Casbin权限系统状态
func (h *HealthHandler) checkCasbin(ctx context.Context, resultChan chan<- ComponentStatus) {
	start := time.Now()
	status := ComponentStatus{
		Name:        "casbin",
		Status:      "unhealthy",
		LastChecked: time.Now(),
	}

	defer func() {
		status.latency = time.Since(start)
		select {
		case resultChan <- status:
		case <-ctx.Done():
		}
	}()

	if h.mgr.Enforcer == nil {
		status.Message = "Casbin权限系统未初始化"
		return
	}

	// 简单检查enforcer是否可用
	if _, err := h.mgr.Enforcer.GetAllRoles(); err != nil {
		status.Message = "Casbin权限系统异常: " + err.Error()
		return
	}

	status.Status = "healthy"
	status.Message = "Casbin权限系统正常"
}

// getSystemInfo 获取系统信息
func (h *HealthHandler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:     runtime.Version(),
		Goroutines:    runtime.NumGoroutine(),
		MemoryUsed:    m.Alloc,
		NumCPU:        runtime.NumCPU(),
		NumGoroutine:  runtime.NumGoroutine(),
	}
}
