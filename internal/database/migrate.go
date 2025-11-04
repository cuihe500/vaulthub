package database

import (
	"database/sql"
	"fmt"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	// MigrationPath 迁移文件目录
	MigrationPath = "file://internal/database/migrations"
)

// Migrator 迁移管理器
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator 创建迁移管理器
func NewMigrator(cfg config.DatabaseConfig) (*Migrator, error) {
	// 打开数据库连接
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 创建 MySQL driver
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithDatabaseInstance(MigrationPath, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{m: m}, nil
}

// Up 执行所有未应用的迁移
func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	version, dirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get version: %w", err)
	}

	logger.Info("迁移完成", logger.Uint("version", version), logger.Bool("dirty", dirty))
	return nil
}

// Down 回滚最后一次迁移
func (mg *Migrator) Down() error {
	if err := mg.m.Down(); err != nil {
		return fmt.Errorf("migration down failed: %w", err)
	}

	version, dirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get version: %w", err)
	}

	logger.Info("回滚完成", logger.Uint("version", version), logger.Bool("dirty", dirty))
	return nil
}

// Steps 执行指定步数的迁移（正数向上，负数向下）
func (mg *Migrator) Steps(n int) error {
	if err := mg.m.Steps(n); err != nil {
		return fmt.Errorf("migration steps failed: %w", err)
	}

	version, dirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get version: %w", err)
	}

	logger.Info("迁移步骤完成", logger.Int("steps", n), logger.Uint("version", version), logger.Bool("dirty", dirty))
	return nil
}

// Version 获取当前数据库版本
func (mg *Migrator) Version() (uint, bool, error) {
	version, dirty, err := mg.m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}
	return version, dirty, nil
}

// Force 强制设置版本（仅在 dirty 状态下使用）
func (mg *Migrator) Force(version int) error {
	if err := mg.m.Force(version); err != nil {
		return fmt.Errorf("force version failed: %w", err)
	}
	logger.Info("强制设置版本完成", logger.Int("version", version))
	return nil
}

// Close 关闭迁移器
func (mg *Migrator) Close() error {
	srcErr, dbErr := mg.m.Close()
	if srcErr != nil {
		return fmt.Errorf("failed to close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}
