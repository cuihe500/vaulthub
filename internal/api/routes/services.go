package routes

import (
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/internal/service"
)

// ServiceContainer 服务容器，统一管理所有业务服务的创建和依赖关系
// 优点：
// 1. 集中管理服务依赖图，避免在路由层重复创建
// 2. 依赖关系清晰可见，便于理解服务间的调用链
// 3. 新增服务时只需修改此文件，符合单一职责原则
type ServiceContainer struct {
	Email        *service.EmailService
	Auth         *service.AuthService
	User         *service.UserService
	Profile      *service.UserProfileService
	Encryption   *service.EncryptionService
	Recovery     *service.RecoveryService
	KeyRotation  *service.KeyRotationService
	SystemConfig *service.SystemConfigService
	Statistics   *service.StatisticsService
}

// NewServiceContainer 创建服务容器
// 按照依赖顺序构建服务实例：
// 1. 基础服务（无依赖）：Email, User, Profile, Encryption, Recovery
// 2. 依赖基础服务的服务：Auth(依赖Email), KeyRotation(依赖Encryption)
// 3. 系统服务：SystemConfig, Statistics
func NewServiceContainer(mgr *app.Manager) *ServiceContainer {
	sc := &ServiceContainer{}

	// 第一层：基础服务（无其他服务依赖）
	sc.Email = service.NewEmailService(mgr.DB, mgr.Redis, mgr.ConfigManager)
	sc.User = service.NewUserService(mgr.DB)
	sc.Profile = service.NewUserProfileService(mgr.DB)
	sc.Encryption = service.NewEncryptionService(mgr.DB)
	sc.Recovery = service.NewRecoveryService(mgr.DB)

	// 第二层：依赖其他服务的服务
	sc.Auth = service.NewAuthService(mgr.DB, mgr.JWT, mgr.Redis, sc.Email)
	sc.KeyRotation = service.NewKeyRotationService(mgr.DB, sc.Encryption, mgr.ConfigManager)

	// 第三层：系统服务
	sc.SystemConfig = service.NewSystemConfigService(mgr.DB, mgr.ConfigManager)
	sc.Statistics = service.NewStatisticsService(mgr.DB)

	return sc
}
