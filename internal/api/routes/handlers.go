package routes

import (
	"github.com/cuihe500/vaulthub/internal/api/handlers"
	"github.com/cuihe500/vaulthub/internal/app"
)

// HandlerContainer 处理器容器，统一管理所有HTTP处理器的创建
// 优点：
// 1. 将处理器创建逻辑从路由层剥离，职责更清晰
// 2. 依赖注入统一管理，避免在路由文件中散落大量构造代码
// 3. 新增处理器时只需修改此文件，降低维护成本
type HandlerContainer struct {
	Health     *handlers.HealthHandler
	Auth       *handlers.AuthHandler
	User       *handlers.UserHandler
	Profile    *handlers.UserProfileHandler
	Secret     *handlers.SecretHandler
	KeyManage  *handlers.KeyManagementHandler
	SysConfig  *handlers.SystemConfigHandler
	Email      *handlers.EmailHandler
	Audit      *handlers.AuditHandler
	Statistics *handlers.StatisticsHandler
	Casbin     *handlers.CasbinHandler
}

// NewHandlerContainer 创建处理器容器
// 参数：
//   - mgr: 全局管理器，提供基础设施依赖（如AuditService、Enforcer）
//   - svc: 服务容器，提供业务服务依赖
//
// 注意：处理器的创建顺序可以任意，因为它们之间没有依赖关系
func NewHandlerContainer(mgr *app.Manager, svc *ServiceContainer) *HandlerContainer {
	return &HandlerContainer{
		Health:     handlers.NewHealthHandler(mgr),
		Auth:       handlers.NewAuthHandler(svc.Auth, svc.Recovery, mgr.DB),
		User:       handlers.NewUserHandler(svc.User),
		Profile:    handlers.NewUserProfileHandler(svc.Profile),
		Secret:     handlers.NewSecretHandler(svc.Encryption),
		KeyManage:  handlers.NewKeyManagementHandler(svc.Encryption, svc.Recovery, svc.KeyRotation),
		SysConfig:  handlers.NewSystemConfigHandler(svc.SystemConfig),
		Email:      handlers.NewEmailHandler(svc.Email),
		Audit:      handlers.NewAuditHandler(mgr.AuditService),
		Statistics: handlers.NewStatisticsHandler(svc.Statistics),
		Casbin:     handlers.NewCasbinHandler(mgr.Enforcer),
	}
}
