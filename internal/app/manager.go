package app

import (
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/jwt"
	"github.com/cuihe500/vaulthub/pkg/logger"
	redisClient "github.com/cuihe500/vaulthub/pkg/redis"
	"gorm.io/gorm"
)

// Manager 管理应用的所有外部连接
// 所有连接在应用启动时初始化一次，之后复用
type Manager struct {
	DB            *gorm.DB
	Enforcer      *casbin.Enforcer      // Casbin权限enforcer
	JWT           *jwt.Manager          // JWT管理器
	Redis         *redisClient.Client   // Redis客户端
	ConfigManager *config.ConfigManager // 系统配置管理器
	AuditService  *service.AuditService // 审计服务
	// Cache *cache.Client // 未来添加其他连接
}

// Initialize 初始化所有连接
// 按顺序初始化各个连接，任何失败都会返回错误
func (m *Manager) Initialize(cfg *config.Config) error {
	// 初始化数据库连接
	if err := m.initDatabase(cfg.Database); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	// 初始化JWT管理器
	if err := m.initJWT(cfg.Security); err != nil {
		return fmt.Errorf("初始化JWT管理器失败: %w", err)
	}

	// 初始化Casbin enforcer
	if err := m.initCasbin(cfg.Security); err != nil {
		return fmt.Errorf("初始化Casbin失败: %w", err)
	}

	// 初始化Redis连接
	if err := m.initRedis(cfg.Redis); err != nil {
		return fmt.Errorf("初始化Redis连接失败: %w", err)
	}

	// 初始化配置管理器
	if err := m.initConfigManager(); err != nil {
		return fmt.Errorf("初始化配置管理器失败: %w", err)
	}

	// 初始化审计服务
	if err := m.initAuditService(cfg.Audit); err != nil {
		return fmt.Errorf("初始化审计服务失败: %w", err)
	}

	// 未来在这里添加其他连接的初始化

	return nil
}

// Close 关闭所有连接
// 在应用关闭时调用，确保所有资源被正确释放
func (m *Manager) Close() error {
	// 关闭审计服务（必须在关闭数据库之前）
	if m.AuditService != nil {
		m.AuditService.Stop()
	}

	// 关闭数据库连接
	if m.DB != nil {
		sqlDB, err := m.DB.DB()
		if err != nil {
			logger.Error("获取数据库连接失败", logger.Err(err))
		} else {
			if err := sqlDB.Close(); err != nil {
				logger.Error("关闭数据库连接失败", logger.Err(err))
			} else {
				logger.Info("数据库连接已关闭")
			}
		}
	}

	// 关闭Redis连接
	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			logger.Error("关闭Redis连接失败", logger.Err(err))
		} else {
			logger.Info("Redis连接已关闭")
		}
	}

	// 未来在这里添加其他连接的关闭逻辑

	return nil
}

// initDatabase 初始化数据库连接
func (m *Manager) initDatabase(cfg config.DatabaseConfig) error {
	db, err := database.Connect(cfg)
	if err != nil {
		return err
	}
	m.DB = db
	return nil
}

// initJWT 初始化JWT管理器
func (m *Manager) initJWT(cfg config.SecurityConfig) error {
	expiration := time.Duration(cfg.JWTExpiration) * time.Hour
	m.JWT = jwt.NewManager(cfg.JWTSecret, expiration)
	logger.Info("JWT管理器初始化成功", logger.Int("expiration_hours", cfg.JWTExpiration))
	return nil
}

// initCasbin 初始化Casbin enforcer
func (m *Manager) initCasbin(cfg config.SecurityConfig) error {
	// 使用GORM adapter，自动创建casbin_rule表
	adapter, err := gormadapter.NewAdapterByDB(m.DB)
	if err != nil {
		return fmt.Errorf("创建Casbin adapter失败: %w", err)
	}

	// 加载模型配置
	enforcer, err := casbin.NewEnforcer(cfg.CasbinModelPath, adapter)
	if err != nil {
		return fmt.Errorf("创建Casbin enforcer失败: %w", err)
	}

	// 从数据库加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("加载Casbin策略失败: %w", err)
	}

	m.Enforcer = enforcer
	logger.Info("Casbin初始化成功", logger.String("model_path", cfg.CasbinModelPath))
	return nil
}

// initRedis 初始化Redis连接
func (m *Manager) initRedis(cfg config.RedisConfig) error {
	client, err := redisClient.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("创建Redis客户端失败: %w", err)
	}
	m.Redis = client
	return nil
}

// initConfigManager 初始化配置管理器
func (m *Manager) initConfigManager() error {
	configManager, err := config.NewConfigManager(m.DB)
	if err != nil {
		return fmt.Errorf("创建配置管理器失败: %w", err)
	}
	m.ConfigManager = configManager
	return nil
}

// initAuditService 初始化审计服务
func (m *Manager) initAuditService(auditCfg config.AuditConfig) error {
	// 从配置中读取审计服务参数
	// 缓冲区满时新审计日志会被丢弃（不阻塞业务）
	auditService := service.NewAuditService(m.DB, auditCfg.BufferSize, auditCfg.WorkerCount)
	auditService.Start()
	m.AuditService = auditService
	logger.Info("审计服务初始化成功",
		logger.Int("buffer_size", auditCfg.BufferSize),
		logger.Int("worker_count", auditCfg.WorkerCount))
	return nil
}

// 未来添加其他连接的初始化方法
