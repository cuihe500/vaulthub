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
	// 使用MigrationDSN以支持迁移文件中的多条SQL语句
	db, err := sql.Open("mysql", cfg.MigrationDSN())
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
	// 获取升级前的版本
	oldVersion, _, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}

	// 执行升级
	if err := mg.m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			// 已经是最新版本，无需升级
			if err == migrate.ErrNilVersion || oldVersion == 0 {
				logger.Info("数据库版本已是最新，无需升级")
			} else {
				logger.Info("数据库版本已是最新，无需升级", logger.Uint("当前版本", oldVersion))
			}
			return nil
		}
		return fmt.Errorf("数据库升级执行失败: %w", err)
	}

	// 获取升级后的版本
	newVersion, newDirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取升级后版本失败: %w", err)
	}

	// 检查是否处于dirty状态
	if newDirty {
		logger.Warn("数据库升级完成但处于不一致状态",
			logger.Uint("目标版本", newVersion),
			logger.String("提示", "请检查迁移文件或使用force命令修复"))
		return nil
	}

	// 打印升级路径
	if err == migrate.ErrNilVersion || oldVersion == 0 {
		logger.Info("数据库初始化成功",
			logger.String("操作", fmt.Sprintf("创建数据库结构并升级到版本%d", newVersion)),
			logger.Uint("当前版本", newVersion))
	} else {
		logger.Info("数据库升级成功",
			logger.String("升级路径", fmt.Sprintf("版本%d -> 版本%d", oldVersion, newVersion)),
			logger.Uint("当前版本", newVersion))
	}

	return nil
}

// Down 回滚最后一次迁移
func (mg *Migrator) Down() error {
	// 获取回滚前的版本
	oldVersion, _, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}

	// 执行回滚
	if err := mg.m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("数据库已是初始状态，无需回滚")
			return nil
		}
		return fmt.Errorf("数据库回滚执行失败: %w", err)
	}

	// 获取回滚后的版本
	newVersion, newDirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取回滚后版本失败: %w", err)
	}

	// 检查是否处于dirty状态
	if newDirty {
		logger.Warn("数据库回滚完成但处于不一致状态",
			logger.Uint("目标版本", newVersion),
			logger.String("提示", "请检查迁移文件或使用force命令修复"))
		return nil
	}

	// 打印回滚路径
	if err == migrate.ErrNilVersion || newVersion == 0 {
		logger.Info("数据库回滚成功完成",
			logger.String("回滚路径", fmt.Sprintf("版本%d -> 初始状态", oldVersion)),
			logger.String("当前状态", "初始状态"))
	} else {
		logger.Info("数据库回滚成功完成",
			logger.String("回滚路径", fmt.Sprintf("版本%d -> 版本%d", oldVersion, newVersion)),
			logger.Uint("当前版本", newVersion))
	}

	return nil
}

// Steps 执行指定步数的迁移（正数向上，负数向下）
func (mg *Migrator) Steps(n int) error {
	if n == 0 {
		logger.Info("步数为0，无需执行")
		return nil
	}

	// 获取操作前的版本
	oldVersion, _, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}

	// 执行步进操作
	if err := mg.m.Steps(n); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("数据库版本已是目标状态，无需操作", logger.Int("步数", n))
			return nil
		}
		if n > 0 {
			return fmt.Errorf("数据库升级执行失败: %w", err)
		}
		return fmt.Errorf("数据库回滚执行失败: %w", err)
	}

	// 获取操作后的版本
	newVersion, newDirty, err := mg.m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取操作后版本失败: %w", err)
	}

	// 检查是否处于dirty状态
	if newDirty {
		if n > 0 {
			logger.Warn("数据库升级完成但处于不一致状态",
				logger.Uint("目标版本", newVersion),
				logger.String("提示", "请检查迁移文件或使用force命令修复"))
		} else {
			logger.Warn("数据库回滚完成但处于不一致状态",
				logger.Uint("目标版本", newVersion),
				logger.String("提示", "请检查迁移文件或使用force命令修复"))
		}
		return nil
	}

	// 打印操作结果
	if n > 0 {
		logger.Info("数据库升级成功",
			logger.Int("升级步数", n),
			logger.String("版本变化", fmt.Sprintf("版本%d -> 版本%d", oldVersion, newVersion)),
			logger.Uint("当前版本", newVersion))
	} else {
		logger.Info("数据库回滚成功",
			logger.Int("回滚步数", -n),
			logger.String("版本变化", fmt.Sprintf("版本%d -> 版本%d", oldVersion, newVersion)),
			logger.Uint("当前版本", newVersion))
	}

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
