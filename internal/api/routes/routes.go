package routes

import (
	_ "github.com/cuihe500/vaulthub/docs/swagger" // swagger docs
	"github.com/cuihe500/vaulthub/internal/api/handlers"
	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 注册所有路由
// mgr: 连接管理器，提供数据库等外部连接
func Setup(r *gin.Engine, mgr *app.Manager) {
	// 全局中间件
	r.Use(middleware.RequestID())

	// 创建 handlers
	healthHandler := handlers.NewHealthHandler(mgr)

	// 健康检查接口（不需要认证）
	r.GET("/health", healthHandler.HealthCheck)

	// Swagger 文档接口
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// TODO: 添加业务路由
		_ = v1
	}
}
