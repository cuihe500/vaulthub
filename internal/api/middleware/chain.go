package middleware

import (
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/gin-gonic/gin"
)

// ChainBuilder 中间件链构建器
// 用于消除路由层的中间件组装重复代码，提供标准化的中间件组合
// 优点：
// 1. 中间件组合标准化，减少遗漏和错误
// 2. 集中管理中间件链，易于维护和调整
// 3. 避免在路由文件中重复书写相同的中间件组合
type ChainBuilder struct {
	mgr *app.Manager
}

// NewChainBuilder 创建中间件链构建器
func NewChainBuilder(mgr *app.Manager) *ChainBuilder {
	return &ChainBuilder{mgr: mgr}
}

// AuthWithAudit 返回认证+审计中间件链（最常用组合）
// 使用场景：需要用户认证且需要记录操作的接口
// 中间件顺序：Auth -> Audit（Audit依赖Auth设置的用户信息）
func (b *ChainBuilder) AuthWithAudit() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
		AuditMiddleware(b.mgr.AuditService),
	}
}

// AuthWithPermission 返回认证+审计+权限验证中间件链
// 使用场景：需要特定权限的管理接口
// 参数：
//   - resource: 资源名称（如"user", "config", "profile"）
//   - action: 操作类型（如"read", "write"）
//
// 中间件顺序：Auth -> Audit -> Permission
func (b *ChainBuilder) AuthWithPermission(resource, action string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
		AuditMiddleware(b.mgr.AuditService),
		RequirePermission(b.mgr.Enforcer, resource, action),
	}
}

// SecureAuth 返回认证+审计+安全密码检查中间件链
// 使用场景：涉及敏感操作的接口（如秘密管理）
// 要求：用户必须设置安全密码（Security PIN）
// 中间件顺序：Auth -> Audit -> SecurityPINCheck
func (b *ChainBuilder) SecureAuth() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
		AuditMiddleware(b.mgr.AuditService),
		SecurityPINCheckMiddleware(b.mgr.DB),
	}
}

// SecureAuthWithPermission 返回认证+审计+权限验证+安全密码检查中间件链
// 使用场景：需要特定权限且涉及敏感操作的接口（如秘密管理）
// 要求：用户必须设置安全密码（Security PIN）并拥有对应权限
// 参数：
//   - resource: 资源名称（如"secret"）
//   - action: 操作类型（如"read", "write"）
//
// 中间件顺序：Auth -> Audit -> Permission -> SecurityPINCheck
// 注意：Permission在SecurityPIN之前，先验证权限再检查PIN，避免无权限用户触发PIN检查
func (b *ChainBuilder) SecureAuthWithPermission(resource, action string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
		AuditMiddleware(b.mgr.AuditService),
		RequirePermission(b.mgr.Enforcer, resource, action),
		SecurityPINCheckMiddleware(b.mgr.DB),
	}
}

// RateLimit 返回限流中间件（无认证）
// 使用场景：公开接口需要限流保护（如注册、登录、发送验证码）
func (b *ChainBuilder) RateLimit() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		RateLimitMiddleware(b.mgr.Redis, b.mgr.ConfigManager),
	}
}

// Auth 返回单独的认证中间件（无审计）
// 使用场景：某些特殊接口只需认证不需审计（使用时需注释说明原因）
// 注意：大部分接口应使用 AuthWithAudit，除非有明确理由不记录审计
func (b *ChainBuilder) Auth() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
	}
}

// AuthWithAuditAndScope 返回认证+审计+作用域限制中间件链
// 使用场景：需要基于用户角色限制数据访问范围的接口（如审计日志、统计数据）
// 中间件顺序：Auth -> Audit -> Scope
// 作用域规则（通过Casbin的scope:global权限控制）：
//   - 有scope:global权限的角色（如admin）：可以访问所有数据
//   - 无此权限的角色（如user, readonly）：只能访问自己的数据
func (b *ChainBuilder) AuthWithAuditAndScope() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		AuthMiddleware(b.mgr.JWT, b.mgr.DB, b.mgr.Redis),
		AuditMiddleware(b.mgr.AuditService),
		ScopeMiddleware(b.mgr.Enforcer),
	}
}
