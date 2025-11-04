package app

import (
	"fmt"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// Manager 管理应用的所有外部连接
// 所有连接在应用启动时初始化一次，之后复用
type Manager struct {
	DB *gorm.DB
	// Redis *redis.Client // 未来添加 Redis
	// Cache *cache.Client // 未来添加其他连接
}

// Initialize 初始化所有连接
// 按顺序初始化各个连接，任何失败都会返回错误
func (m *Manager) Initialize(cfg *config.Config) error {
	// 初始化数据库连接
	if err := m.initDatabase(cfg.Database); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	// 未来在这里添加其他连接的初始化
	// if err := m.initRedis(cfg.Redis); err != nil {
	//     return fmt.Errorf("初始化Redis连接失败: %w", err)
	// }

	return nil
}

// Close 关闭所有连接
// 在应用关闭时调用，确保所有资源被正确释放
func (m *Manager) Close() error {
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

	// 未来在这里添加其他连接的关闭逻辑
	// if m.Redis != nil {
	//     if err := m.Redis.Close(); err != nil {
	//         logger.Error("关闭Redis连接失败", logger.Err(err))
	//     }
	// }

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

// 未来添加其他连接的初始化方法
// func (m *Manager) initRedis(cfg config.RedisConfig) error {
//     // Redis 初始化逻辑
//     return nil
// }
