package main

import (
	"fmt"

	"github.com/cuihe500/vaulthub/internal/database"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	migrateSteps int
	forceVersion int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration operations",
	Long:  `Run database migrations using golang-migrate. Supports up, down, version, force and steps.`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Run:   runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Run:   runMigrateDown,
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Run:   runMigrateVersion,
}

var migrateStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Apply N migrations (positive for up, negative for down)",
	Run:   runMigrateSteps,
}

var migrateForceCmd = &cobra.Command{
	Use:   "force",
	Short: "Force set migration version (use with caution)",
	Run:   runMigrateForce,
}

func init() {
	// 添加迁移子命令
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
	migrateCmd.AddCommand(migrateStepsCmd)
	migrateCmd.AddCommand(migrateForceCmd)

	// 配置文件路径（所有子命令共用）
	migrateCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")

	// steps 命令参数
	migrateStepsCmd.Flags().IntVarP(&migrateSteps, "number", "n", 1, "number of steps to migrate")

	// force 命令参数
	migrateForceCmd.Flags().IntVarP(&forceVersion, "version", "v", 0, "version to force set")
	migrateForceCmd.MarkFlagRequired("version")

	// 添加到根命令
	rootCmd.AddCommand(migrateCmd)
}

// runMigrateUp 执行所有未应用的迁移
func runMigrateUp(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		logger.Fatal("迁移失败", logger.Err(err))
	}

	logger.Info("迁移成功完成")
}

// runMigrateDown 回滚最后一次迁移
func runMigrateDown(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil {
		logger.Fatal("回滚失败", logger.Err(err))
	}

	logger.Info("回滚成功完成")
}

// runMigrateVersion 显示当前迁移版本
func runMigrateVersion(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil {
		logger.Fatal("获取版本失败", logger.Err(err))
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	fmt.Printf("Current version: %d (%s)\n", version, status)
	logger.Info("版本信息", logger.Uint("version", version), logger.Bool("dirty", dirty))
}

// runMigrateSteps 执行指定步数的迁移
func runMigrateSteps(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer migrator.Close()

	if err := migrator.Steps(migrateSteps); err != nil {
		logger.Fatal("执行迁移步骤失败", logger.Err(err))
	}

	logger.Info("迁移步骤成功完成", logger.Int("steps", migrateSteps))
}

// runMigrateForce 强制设置迁移版本
func runMigrateForce(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer migrator.Close()

	logger.Warn("警告：强制设置版本可能导致数据不一致，请谨慎使用", logger.Int("version", forceVersion))

	if err := migrator.Force(forceVersion); err != nil {
		logger.Fatal("强制设置版本失败", logger.Err(err))
	}

	logger.Info("版本设置成功", logger.Int("version", forceVersion))
}
