package routes

import (
	_ "github.com/cuihe500/vaulthub/docs/swagger" // swagger docs
	"github.com/cuihe500/vaulthub/internal/api/handlers"
	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 注册所有路由
// mgr: 连接管理器，提供数据库等外部连接
func Setup(r *gin.Engine, mgr *app.Manager) {
	// 全局中间件
	r.Use(middleware.RequestID())

	// 创建 services
	authService := service.NewAuthService(mgr.DB, mgr.JWT, mgr.Redis)
	userService := service.NewUserService(mgr.DB)

	// 创建 handlers
	healthHandler := handlers.NewHealthHandler(mgr)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// 健康检查接口（不需要认证）
	r.GET("/health", healthHandler.HealthCheck)

	// Swagger 文档接口
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证路由（不需要token）
		// 注册和登录接口不需要认证，因为这是用户获取token的入口
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)

			// 获取当前用户信息需要认证
			auth.GET("/me", middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis), authHandler.GetMe)
			// 登出需要认证
			auth.POST("/logout", middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis), authHandler.Logout)
		}

		// 用户管理路由（需要认证和管理员权限）
		// 这里使用Casbin进行权限验证，user资源的管理操作需要admin权限
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 获取用户列表 - 需要user:read权限
			users.GET("", middleware.RequirePermission(mgr.Enforcer, "user", "read"), userHandler.ListUsers)

			// 获取单个用户 - 需要user:read权限
			users.GET("/:uuid", middleware.RequirePermission(mgr.Enforcer, "user", "read"), userHandler.GetUser)

			// 更新用户状态 - 需要user:write权限
			users.PUT("/:uuid/status", middleware.RequirePermission(mgr.Enforcer, "user", "write"), userHandler.UpdateUserStatus)

			// 更新用户角色 - 需要user:write权限
			users.PUT("/:uuid/role", middleware.RequirePermission(mgr.Enforcer, "user", "write"), userHandler.UpdateUserRole)
		}
	}
}
