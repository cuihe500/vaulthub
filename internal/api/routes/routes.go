package routes

import (
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/cuihe500/vaulthub/docs/swagger" // swagger docs
	"github.com/cuihe500/vaulthub/internal/api/middleware"
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup 注册所有路由
// mgr: 连接管理器，提供数据库等外部连接
func Setup(r *gin.Engine, mgr *app.Manager) {
	// 全局中间件
	r.Use(middleware.RequestID())
	// 注意：审计中间件不能在全局注册，因为它依赖AuthMiddleware设置的用户信息
	// 审计中间件需要在各个路由组的AuthMiddleware之后注册

	// 创建服务容器和处理器容器
	// 依赖关系由容器内部管理，避免在此处手动组装
	svc := NewServiceContainer(mgr)
	h := NewHandlerContainer(mgr, svc)

	// 创建中间件链构建器
	// 用于标准化中间件组合，避免重复代码
	chain := middleware.NewChainBuilder(mgr)

	// 健康检查接口（不需要认证）
	r.GET("/health", h.Health.HealthCheck)

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
			email.POST("/send-code", append(chain.RateLimit(), h.Email.SendCode)...)
			email.POST("/verify-code", append(chain.RateLimit(), h.Email.VerifyCode)...)
		}

		// 认证路由（不需要token）
		// 注册和登录接口不需要认证，因为这是用户获取token的入口
		// 注册和登录接口需要限流保护，防止暴力攻击（配置从数据库动态读取）
		auth := v1.Group("/auth")
		{
			// 公开路由（不需要认证，但需要审计和限流）
			// 注意：现在审计中间件可以处理未认证请求，会记录失败的登录尝试等重要安全事件
			publicChain := []gin.HandlerFunc{
				middleware.AuditMiddleware(mgr.AuditService),
			}
			auth.POST("/register", append(append(publicChain, chain.RateLimit()...), h.Auth.Register)...)
			auth.POST("/login", append(append(publicChain, chain.RateLimit()...), h.Auth.Login)...)
			auth.POST("/login-with-email", append(append(publicChain, chain.RateLimit()...), h.Auth.LoginWithEmail)...)

			// 密码找回路由（不需要认证，需要审计和限流）
			auth.POST("/request-password-reset", append(append(publicChain, chain.RateLimit()...), h.Auth.RequestPasswordReset)...)
			auth.GET("/verify-reset-token", append(publicChain, h.Auth.VerifyPasswordResetToken)...)
			auth.POST("/reset-password-with-token", append(append(publicChain, chain.RateLimit()...), h.Auth.ResetPasswordWithToken)...)

			// 需要认证的路由（使用认证+审计中间件链）
			auth.GET("/me", append(chain.AuthWithAudit(), h.Auth.GetMe)...)
			auth.POST("/logout", append(chain.AuthWithAudit(), h.Auth.Logout)...)
			auth.POST("/reset-password", append(chain.AuthWithAudit(), h.Auth.ResetPassword)...)
			auth.GET("/security-pin-status", append(chain.AuthWithAudit(), h.Auth.GetSecurityPINStatus)...)
		}

		// 用户管理路由（需要认证和管理员权限）
		// 这里使用Casbin进行权限验证，user资源的管理操作需要admin权限
		users := v1.Group("/users")
		{
			// 获取用户列表 - 需要user:read权限
			users.GET("", append(chain.AuthWithPermission(middleware.ResourceUser, middleware.ActionRead), h.User.ListUsers)...)

			// 获取单个用户 - 需要user:read权限
			users.GET("/:uuid", append(chain.AuthWithPermission(middleware.ResourceUser, middleware.ActionRead), h.User.GetUser)...)

			// 更新用户状态 - 需要user:write权限
			users.PUT("/:uuid/status", append(chain.AuthWithPermission(middleware.ResourceUser, middleware.ActionWrite), h.User.UpdateUserStatus)...)

			// 更新用户角色 - 需要user:write权限
			users.PUT("/:uuid/role", append(chain.AuthWithPermission(middleware.ResourceUser, middleware.ActionWrite), h.User.UpdateUserRole)...)
		}

		// 用户档案路由（需要认证）
		profile := v1.Group("/profile")
		profile.Use(chain.AuthWithAudit()...)
		{
			// 获取当前用户档案 - 用户只能操作自己的档案
			profile.GET("", h.Profile.GetProfile)

			// 创建用户档案
			profile.POST("", h.Profile.CreateProfile)

			// 更新用户档案
			profile.PUT("", h.Profile.UpdateProfile)

			// 创建或更新用户档案
			profile.PATCH("", h.Profile.CreateOrUpdateProfile)

			// 删除用户档案
			profile.DELETE("", h.Profile.DeleteProfile)
		}

		// 加密密钥管理路由（需要认证+权限验证）
		// 用户只能操作自己的加密密钥
		// 权限要求：key:read用于查询操作，key:write用于创建/轮换操作
		keys := v1.Group("/keys")
		{
			// 创建用户加密密钥（首次使用加密功能时调用）- 需要key:write权限
			keys.POST("/create", append(chain.AuthWithPermission(middleware.ResourceKey, middleware.ActionWrite), h.KeyManage.CreateUserEncryptionKey)...)

			// 验证恢复密钥有效性 - 需要key:read权限
			keys.POST("/verify-recovery", append(chain.AuthWithPermission(middleware.ResourceKey, middleware.ActionRead), h.KeyManage.VerifyRecoveryKey)...)

			// 手动触发密钥轮换 - 需要key:write权限
			// 注意：readonly角色不应该有此权限（安全关键操作）
			keys.POST("/rotate", append(chain.AuthWithPermission(middleware.ResourceKey, middleware.ActionWrite), h.KeyManage.RotateDEK)...)

			// 查询密钥轮换进度 - 需要key:read权限
			keys.GET("/rotation-status", append(chain.AuthWithPermission(middleware.ResourceKey, middleware.ActionRead), h.KeyManage.GetRotationStatus)...)
		}

		// 秘密管理路由（需要认证+权限验证+安全密码）
		// 用户只能操作自己的秘密
		// 注意：秘密管理需要用户先设置安全密码（Security PIN）
		// 权限要求：secret:read用于查询/解密，secret:write用于创建/删除
		secrets := v1.Group("/secrets")
		{
			// 获取秘密列表 - 需要secret:read权限
			secrets.GET("", append(chain.SecureAuthWithPermission(middleware.ResourceSecret, middleware.ActionRead), h.Secret.ListSecrets)...)

			// 创建秘密 - 需要secret:write权限
			// 注意：readonly角色不应该有此权限
			secrets.POST("", append(chain.SecureAuthWithPermission(middleware.ResourceSecret, middleware.ActionWrite), h.Secret.CreateSecret)...)

			// 解密秘密（获取明文）- 需要secret:read权限
			secrets.POST("/:uuid/decrypt", append(chain.SecureAuthWithPermission(middleware.ResourceSecret, middleware.ActionRead), h.Secret.GetSecret)...)

			// 删除秘密 - 需要secret:write权限
			// 注意：readonly角色不应该有此权限
			secrets.DELETE("/:uuid", append(chain.SecureAuthWithPermission(middleware.ResourceSecret, middleware.ActionWrite), h.Secret.DeleteSecret)...)
		}

		// 管理员用户档案路由（需要认证和管理员权限）
		admin := v1.Group("/admin")
		{
			// 获取用户档案列表 - 需要profile:read权限
			admin.GET("/profiles", append(chain.AuthWithPermission(middleware.ResourceProfile, middleware.ActionRead), h.Profile.ListProfiles)...)

			// 获取指定用户档案 - 需要profile:read权限
			admin.GET("/users/:user_id/profile", append(chain.AuthWithPermission(middleware.ResourceProfile, middleware.ActionRead), h.Profile.GetUserProfile)...)

			// 更新指定用户档案 - 需要profile:write权限
			admin.PUT("/users/:user_id/profile", append(chain.AuthWithPermission(middleware.ResourceProfile, middleware.ActionWrite), h.Profile.UpdateUserProfile)...)
		}

		// 系统配置路由（需要认证和管理员权限）
		// 系统配置的管理属于敏感操作，需要config:read和config:write权限
		configs := v1.Group("/configs")
		{
			// 获取配置列表 - 需要config:read权限
			configs.GET("", append(chain.AuthWithPermission(middleware.ResourceConfig, middleware.ActionRead), h.SysConfig.ListConfigs)...)

			// 获取单个配置 - 需要config:read权限
			configs.GET("/:key", append(chain.AuthWithPermission(middleware.ResourceConfig, middleware.ActionRead), h.SysConfig.GetConfig)...)

			// 更新配置 - 需要config:write权限
			configs.PUT("/:key", append(chain.AuthWithPermission(middleware.ResourceConfig, middleware.ActionWrite), h.SysConfig.UpdateConfig)...)

			// 批量更新配置 - 需要config:write权限
			configs.PUT("/batch", append(chain.AuthWithPermission(middleware.ResourceConfig, middleware.ActionWrite), h.SysConfig.BatchUpdateConfigs)...)

			// 重新加载配置 - 需要config:write权限
			configs.POST("/reload", append(chain.AuthWithPermission(middleware.ResourceConfig, middleware.ActionWrite), h.SysConfig.ReloadConfigs)...)

			// Casbin权限策略管理子路由
			// 用于运行时热更新权限策略，无需重启服务
			casbin := configs.Group("/casbin")
			{
				// 重新加载Casbin权限策略 - 需要casbin:reload权限
				casbin.POST("/reload", append(chain.AuthWithPermission(middleware.ResourceCasbin, middleware.ActionReload), h.Casbin.ReloadPolicy)...)
			}
		}

		// 审计日志路由（需要认证+作用域限制）
		// 作用域中间件自动处理权限：普通用户只能查询自己的日志，管理员可以查询所有用户的日志
		audit := v1.Group("/audit")
		audit.Use(chain.AuthWithAuditAndScope()...)
		{
			// 查询审计日志
			audit.GET("/logs", h.Audit.QueryAuditLogs)
			// 导出密钥类型统计
			audit.GET("/logs/export", h.Audit.ExportStatistics)
			// 导出操作统计
			audit.GET("/operations/export", h.Audit.ExportOperationStatistics)
		}

		// 统计数据路由（需要认证+作用域限制）
		// 作用域中间件自动处理权限：普通用户只能查询自己的统计，管理员可以查询所有用户的统计
		statistics := v1.Group("/statistics")
		statistics.Use(chain.AuthWithAuditAndScope()...)
		{
			// 获取用户统计数据（历史统计）
			statistics.GET("/user", h.Statistics.GetUserStatistics)

			// 获取当前统计（实时统计）
			statistics.GET("/current", h.Statistics.GetCurrentStatistics)
		}
	}

	// 静态文件服务（前端资源）
	// 从环境变量VAULTHUB_STATIC_DIR读取前端静态文件路径
	// 支持Vue Router的history模式，所有未匹配的路由fallback到index.html
	staticDir := os.Getenv("VAULTHUB_STATIC_DIR")
	if staticDir == "" {
		staticDir = "./web/dist" // 默认路径
	}

	// 检查静态文件目录是否存在
	if _, err := os.Stat(staticDir); err == nil {
		logger.Info("启用前端静态文件服务", logger.String("path", staticDir))

		// 静态资源路由（优先匹配）
		r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
		r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))

		// SPA fallback：所有未匹配的路由返回index.html
		r.NoRoute(func(c *gin.Context) {
			// 检查是否是API请求（避免API请求返回HTML）
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "接口不存在",
				})
				return
			}
			c.File(filepath.Join(staticDir, "index.html"))
		})
	} else {
		logger.Warn("前端静态文件目录不存在，跳过静态文件服务",
			logger.String("path", staticDir),
			logger.Err(err))
	}
}
