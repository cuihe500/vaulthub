package database

import (
	"fmt"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect 创建并返回数据库连接
// 不使用全局变量，由Manager管理连接的生命周期
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN())
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	// 使用项目统一日志接口
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.NewGormLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	logger.Info("数据库连接成功")
	return db, nil
}
