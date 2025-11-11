package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ScopeMiddleware 数据作用域中间件（基于Casbin）
// 用于统一控制用户的数据访问范围，避免在handler内部重复判断权限
// 核心逻辑：
//   - 检查用户是否有 scope:global 权限（通过Casbin）
//   - 有权限：可以访问所有数据，不设置scope_user_uuid
//   - 无权限：只能访问自己的数据，在context中设置scope_user_uuid强制限制
//
// 使用场景：audit日志查询、statistics统计查询等需要基于用户角色控制数据范围的接口
//
// 权限配置（在数据库casbin_rule表中）：
//   - admin角色有 scope:global 权限，可查询全局数据
//   - 其他角色无此权限，只能查询自己的数据
func ScopeMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户信息（由AuthMiddleware设置）
		userUUID, exists := GetCurrentUserUUID(c)
		if !exists {
			// 如果没有用户信息，说明AuthMiddleware未执行或失败
			// 这里不处理，让后续的handler处理未授权错误
			c.Next()
			return
		}

		role, exists := GetCurrentUserRole(c)
		if !exists {
			logger.Warn("作用域检查失败：无法获取用户角色",
				logger.String("user_uuid", userUUID))
			// 安全起见，无法获取角色时强制限制为当前用户
			c.Set("scope_user_uuid", userUUID)
			c.Next()
			return
		}

		// 通过Casbin检查是否有全局数据访问权限
		hasGlobalScope, err := enforcer.Enforce(role, "scope", "global")
		if err != nil {
			logger.Error("作用域权限检查失败",
				logger.String("role", role),
				logger.Err(err))
			// 出错时安全起见，限制为当前用户
			c.Set("scope_user_uuid", userUUID)
			c.Next()
			return
		}

		if !hasGlobalScope {
			// 无全局访问权限：强制限制数据范围为当前用户
			c.Set("scope_user_uuid", userUUID)
			logger.Debug("用户数据访问受限，仅限本人数据",
				logger.String("role", role),
				logger.String("user_uuid", userUUID))
		} else {
			// 有全局访问权限：不设置限制，handler可根据请求参数决定查询范围
			logger.Debug("用户拥有全局数据访问权限",
				logger.String("role", role),
				logger.String("user_uuid", userUUID))
		}

		c.Next()
	}
}

// GetScopeUserUUID 获取作用域限制的用户UUID
// 返回值：
//   - userUUID: 作用域限制的用户UUID
//   - restricted: true表示有作用域限制（普通用户），false表示无限制（管理员）
//
// 使用示例：
//
//	if scopeUUID, restricted := middleware.GetScopeUserUUID(c); restricted {
//	    // 普通用户，强制使用scopeUUID查询
//	    req.UserUUID = scopeUUID
//	} else if req.UserUUID == "" {
//	    // 管理员未指定用户UUID，查询全局数据
//	}
func GetScopeUserUUID(c *gin.Context) (userUUID string, restricted bool) {
	if uuid, exists := c.Get("scope_user_uuid"); exists {
		return uuid.(string), true
	}
	return "", false
}
