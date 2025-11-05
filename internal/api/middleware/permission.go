package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware Casbin权限检查中间件
// 需要在AuthMiddleware之后使用
func PermissionMiddleware(enforcer *casbin.Enforcer, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		role, exists := GetCurrentUserRole(c)
		if !exists {
			logger.Error("权限检查失败：无法获取用户角色")
			response.Unauthorized(c, "未授权")
			c.Abort()
			return
		}

		// 检查权限
		allowed, err := enforcer.Enforce(role, resource, action)
		if err != nil {
			logger.Error("权限检查失败",
				logger.String("role", role),
				logger.String("resource", resource),
				logger.String("action", action),
				logger.Err(err))
			response.InternalError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !allowed {
			logger.Warn("权限不足",
				logger.String("role", role),
				logger.String("resource", resource),
				logger.String("action", action))
			response.InsufficientPermission(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 创建需要特定权限的中间件
// 这是PermissionMiddleware的便捷包装函数
func RequirePermission(enforcer *casbin.Enforcer, resource, action string) gin.HandlerFunc {
	return PermissionMiddleware(enforcer, resource, action)
}
