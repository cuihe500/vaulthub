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
	// 注意：EmailService 需要先创建，因为 AuthService 依赖它
	emailService := service.NewEmailService(mgr.DB, mgr.Redis, mgr.ConfigManager)
	authService := service.NewAuthService(mgr.DB, mgr.JWT, mgr.Redis, emailService)
	userService := service.NewUserService(mgr.DB)
	profileService := service.NewUserProfileService(mgr.DB)
	encryptionService := service.NewEncryptionService(mgr.DB)
	recoveryService := service.NewRecoveryService(mgr.DB)
	keyRotationService := service.NewKeyRotationService(mgr.DB, encryptionService, mgr.ConfigManager)
	systemConfigService := service.NewSystemConfigService(mgr.DB, mgr.ConfigManager)

	// 创建 handlers
	healthHandler := handlers.NewHealthHandler(mgr)
	authHandler := handlers.NewAuthHandler(authService, recoveryService)
	userHandler := handlers.NewUserHandler(userService)
	profileHandler := handlers.NewUserProfileHandler(profileService)
	secretHandler := handlers.NewSecretHandler(encryptionService)
	keyManagementHandler := handlers.NewKeyManagementHandler(encryptionService, recoveryService, keyRotationService)
	systemConfigHandler := handlers.NewSystemConfigHandler(systemConfigService)
	emailHandler := handlers.NewEmailHandler(emailService)

	// 健康检查接口（不需要认证）
	r.GET("/health", healthHandler.HealthCheck)

	// Swagger 文档接口
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 邮件路由（不需要认证，但需要限流）
		// 邮件验证码用于注册、登录、重置密码等场景，不需要token
		// 邮件发送接口需要严格的限流保护，防止滥用和垃圾邮件攻击
		email := v1.Group("/email")
		{
			email.POST("/send-code",
				middleware.RateLimitMiddleware(mgr.Redis, mgr.ConfigManager),
				emailHandler.SendCode)
			email.POST("/verify-code",
				middleware.RateLimitMiddleware(mgr.Redis, mgr.ConfigManager),
				emailHandler.VerifyCode)
		}

		// 认证路由（不需要token）
		// 注册和登录接口不需要认证，因为这是用户获取token的入口
		// 注册和登录接口需要限流保护，防止暴力攻击（配置从数据库动态读取）
		auth := v1.Group("/auth")
		{
			auth.POST("/register",
				middleware.RateLimitMiddleware(mgr.Redis, mgr.ConfigManager),
				authHandler.Register)
			auth.POST("/login",
				middleware.RateLimitMiddleware(mgr.Redis, mgr.ConfigManager),
				authHandler.Login)

			// 获取当前用户信息需要认证
			auth.GET("/me", middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis), authHandler.GetMe)
			// 登出需要认证
			auth.POST("/logout", middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis), authHandler.Logout)
			// 使用恢复密钥重置密码需要认证
			auth.POST("/reset-password", middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis), authHandler.ResetPassword)
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

		// 用户档案路由（需要认证）
		profile := v1.Group("/profile")
		profile.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 获取当前用户档案 - 用户只能操作自己的档案
			profile.GET("", profileHandler.GetProfile)

			// 创建用户档案
			profile.POST("", profileHandler.CreateProfile)

			// 更新用户档案
			profile.PUT("", profileHandler.UpdateProfile)

			// 创建或更新用户档案
			profile.PATCH("", profileHandler.CreateOrUpdateProfile)

			// 删除用户档案
			profile.DELETE("", profileHandler.DeleteProfile)
		}

		// 加密密钥管理路由（需要认证）
		// 用户只能操作自己的加密密钥
		keys := v1.Group("/keys")
		keys.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 创建用户加密密钥（首次使用加密功能时调用）
			keys.POST("/create", keyManagementHandler.CreateUserEncryptionKey)
			// 验证恢复密钥有效性
			keys.POST("/verify-recovery", keyManagementHandler.VerifyRecoveryKey)
			// 手动触发密钥轮换
			keys.POST("/rotate", keyManagementHandler.RotateDEK)
			// 查询密钥轮换进度
			keys.GET("/rotation-status", keyManagementHandler.GetRotationStatus)
		}

		// 秘密管理路由（需要认证）
		// 用户只能操作自己的秘密
		secrets := v1.Group("/secrets")
		secrets.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 获取秘密列表
			secrets.GET("", secretHandler.ListSecrets)

			// 创建秘密
			secrets.POST("", secretHandler.CreateSecret)

			// 解密秘密（获取明文）
			secrets.POST("/:uuid/decrypt", secretHandler.GetSecret)

			// 删除秘密
			secrets.DELETE("/:uuid", secretHandler.DeleteSecret)
		}

		// 管理员用户档案路由（需要认证和管理员权限）
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 获取用户档案列表 - 需要profile:read权限
			admin.GET("/profiles", middleware.RequirePermission(mgr.Enforcer, "profile", "read"), profileHandler.ListProfiles)

			// 获取指定用户档案 - 需要profile:read权限
			admin.GET("/users/:user_id/profile", middleware.RequirePermission(mgr.Enforcer, "profile", "read"), profileHandler.GetUserProfile)

			// 更新指定用户档案 - 需要profile:write权限
			admin.PUT("/users/:user_id/profile", middleware.RequirePermission(mgr.Enforcer, "profile", "write"), profileHandler.UpdateUserProfile)
		}

		// 系统配置路由（需要认证和管理员权限）
		// 系统配置的管理属于敏感操作，需要config:read和config:write权限
		configs := v1.Group("/configs")
		configs.Use(middleware.AuthMiddleware(mgr.JWT, mgr.DB, mgr.Redis))
		{
			// 获取配置列表 - 需要config:read权限
			configs.GET("", middleware.RequirePermission(mgr.Enforcer, "config", "read"), systemConfigHandler.ListConfigs)

			// 获取单个配置 - 需要config:read权限
			configs.GET("/:key", middleware.RequirePermission(mgr.Enforcer, "config", "read"), systemConfigHandler.GetConfig)

			// 更新配置 - 需要config:write权限
			configs.PUT("/:key", middleware.RequirePermission(mgr.Enforcer, "config", "write"), systemConfigHandler.UpdateConfig)

			// 批量更新配置 - 需要config:write权限
			configs.PUT("/batch", middleware.RequirePermission(mgr.Enforcer, "config", "write"), systemConfigHandler.BatchUpdateConfigs)

			// 重新加载配置 - 需要config:write权限
			configs.POST("/reload", middleware.RequirePermission(mgr.Enforcer, "config", "write"), systemConfigHandler.ReloadConfigs)
		}
	}
}
