package app

import (
	"fmt"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database"
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/crypto"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Initialize 应用初始化
// 在应用启动前执行，确保数据库版本一致、超级管理员已创建
// 执行顺序：
// 1. 检查并执行数据库迁移（如果版本不一致）
// 2. 检查超级管理员初始化标志
// 3. 如果未初始化且配置了admin账号，则创建超级管理员并设置标志
func Initialize(cfg *config.Config) error {
	logger.Info("开始应用初始化检查")

	// 1. 执行数据库迁移
	if err := ensureDatabaseMigration(cfg); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 2. 初始化超级管理员（如果需要）
	if err := ensureSuperAdmin(cfg); err != nil {
		return fmt.Errorf("超级管理员初始化失败: %w", err)
	}

	logger.Info("应用初始化检查完成")
	return nil
}

// ensureDatabaseMigration 确保数据库迁移已执行
// 检查数据库版本，如果不一致则执行迁移
func ensureDatabaseMigration(cfg *config.Config) error {
	logger.Info("检查数据库迁移状态")

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %w", err)
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	// 执行迁移，如果已是最新版本则跳过
	if err := migrator.Up(); err != nil {
		return fmt.Errorf("执行迁移失败: %w", err)
	}

	logger.Info("数据库迁移检查完成")
	return nil
}

// ensureSuperAdmin 确保超级管理员已创建
// 检查admin_initialized标志：
// - 如果标志为true，跳过（说明之前已经执行过初始化）
// - 如果标志为false，检查是否配置了admin账号，如果配置了则创建并设置标志为true
func ensureSuperAdmin(cfg *config.Config) error {
	logger.Info("检查超级管理员初始化状态")

	// 创建临时数据库连接用于检查和创建管理员
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Error("关闭数据库连接失败", logger.Err(err))
		}
	}()

	// 检查admin_initialized标志
	var sysConfig models.SystemConfig
	err = db.Where("config_key = ?", models.ConfigKeyAdminInitialized).First(&sysConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("未找到admin_initialized配置项，可能是数据库迁移未完成")
			return errors.New(errors.CodeDatabaseError, "系统配置表未正确初始化")
		}
		logger.Error("查询admin_initialized标志失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 如果标志已设置为true，跳过初始化
	if sysConfig.ConfigValue == models.ConfigValueTrue {
		logger.Info("超级管理员已在之前初始化过，跳过检查")
		return nil
	}

	// 标志为false，执行超级管理员初始化
	logger.Info("首次启动，开始初始化超级管理员")

	// 检查是否配置了admin账号
	if cfg.Security.AdminUsername == "" || cfg.Security.AdminPassword == "" {
		logger.Warn("未配置超级管理员账号，跳过创建")
		// 即使未配置也设置标志为true，避免每次启动都检查
		return setAdminInitializedFlag(db, true)
	}

	// 检查admin用户是否已存在
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", cfg.Security.AdminUsername).Count(&count).Error; err != nil {
		logger.Error("检查admin用户失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	if count > 0 {
		logger.Info("超级管理员账号已存在，设置初始化标志",
			logger.String("username", cfg.Security.AdminUsername))
		return setAdminInitializedFlag(db, true)
	}

	// 创建超级管理员
	if err := createSuperAdmin(db, cfg); err != nil {
		return err
	}

	// 设置初始化标志为true
	return setAdminInitializedFlag(db, true)
}

// createSuperAdmin 创建超级管理员用户
func createSuperAdmin(db *gorm.DB, cfg *config.Config) error {
	// 验证密码强度
	if !crypto.ValidatePasswordStrength(cfg.Security.AdminPassword) {
		logger.Error("超级管理员密码强度不足")
		return errors.New(errors.CodeWeakPassword, "超级管理员密码必须至少8个字符,包含大小写字母、数字和特殊字符")
	}

	// 加密密码
	passwordHash, err := crypto.HashPassword(cfg.Security.AdminPassword)
	if err != nil {
		logger.Error("密码加密失败", logger.Err(err))
		return errors.Wrap(errors.CodeCryptoError, err)
	}

	// 创建admin用户
	admin := &models.User{
		UUID:         uuid.New().String(),
		Username:     cfg.Security.AdminUsername,
		PasswordHash: passwordHash,
		Status:       models.UserStatusActive,
		Role:         "admin", // 超级管理员角色
	}

	if err := db.Create(admin).Error; err != nil {
		logger.Error("创建超级管理员失败",
			logger.String("username", cfg.Security.AdminUsername),
			logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("超级管理员账号创建成功",
		logger.String("uuid", admin.UUID),
		logger.String("username", admin.Username),
		logger.String("role", admin.Role))

	return nil
}

// setAdminInitializedFlag 设置admin_initialized标志
func setAdminInitializedFlag(db *gorm.DB, initialized bool) error {
	value := models.ConfigValueFalse
	if initialized {
		value = models.ConfigValueTrue
	}

	result := db.Model(&models.SystemConfig{}).
		Where("config_key = ?", models.ConfigKeyAdminInitialized).
		Update("config_value", value)

	if result.Error != nil {
		logger.Error("更新admin_initialized标志失败", logger.Err(result.Error))
		return errors.Wrap(errors.CodeDatabaseError, result.Error)
	}

	logger.Info("超级管理员初始化标志已设置",
		logger.String("value", value))
	return nil
}
