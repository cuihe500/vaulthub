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
	Short: "数据库迁移操作",
	Long:  `使用 golang-migrate 运行数据库迁移。支持 up、down、version、force 和 steps 操作。`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "应用所有待执行的迁移",
	Run:   runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最后一次迁移",
	Run:   runMigrateDown,
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示当前迁移版本",
	Run:   runMigrateVersion,
}

var migrateStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "应用 N 次迁移（正数升级，负数降级）",
	Run:   runMigrateSteps,
}

var migrateForceCmd = &cobra.Command{
	Use:   "force",
	Short: "强制设置迁移版本（谨慎使用）",
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
	migrateCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "配置文件路径")

	// steps 命令参数
	migrateStepsCmd.Flags().IntVarP(&migrateSteps, "number", "n", 1, "迁移步数")

	// force 命令参数
	migrateForceCmd.Flags().IntVarP(&forceVersion, "version", "v", 0, "要强制设置的版本号")
	_ = migrateForceCmd.MarkFlagRequired("version") // 标记为必需参数

	// 添加到根命令
	rootCmd.AddCommand(migrateCmd)
}

// runMigrateUp 执行所有未应用的迁移
func runMigrateUp(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer func() {
		_ = logger.Sync() // 日志同步失败不影响程序退出
	}()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	if err := migrator.Up(); err != nil {
		logger.Fatal("数据库升级失败", logger.Err(err))
	}
}

// runMigrateDown 回滚最后一次迁移
func runMigrateDown(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer func() {
		_ = logger.Sync() // 日志同步失败不影响程序退出
	}()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	if err := migrator.Down(); err != nil {
		logger.Fatal("回滚失败", logger.Err(err))
	}
}

// runMigrateVersion 显示当前迁移版本
func runMigrateVersion(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer func() {
		_ = logger.Sync() // 日志同步失败不影响程序退出
	}()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	version, dirty, err := migrator.Version()
	if err != nil {
		logger.Fatal("获取版本失败", logger.Err(err))
	}

	status := "正常"
	if dirty {
		status = "异常"
	}

	fmt.Printf("当前版本: %d (%s)\n", version, status)
	logger.Info("版本信息", logger.Uint("version", version), logger.Bool("dirty", dirty))
}

// runMigrateSteps 执行指定步数的迁移
func runMigrateSteps(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer func() {
		_ = logger.Sync() // 日志同步失败不影响程序退出
	}()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	if err := migrator.Steps(migrateSteps); err != nil {
		logger.Fatal("数据库操作失败", logger.Err(err))
	}
}

// runMigrateForce 强制设置迁移版本
func runMigrateForce(cmd *cobra.Command, args []string) {
	cfg := loadConfig()
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer func() {
		_ = logger.Sync() // 日志同步失败不影响程序退出
	}()

	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		logger.Fatal("创建迁移器失败", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("关闭迁移器失败", logger.Err(err))
		}
	}()

	logger.Warn("警告：强制设置版本可能导致数据不一致，请谨慎使用", logger.Int("version", forceVersion))

	if err := migrator.Force(forceVersion); err != nil {
		logger.Fatal("强制设置版本失败", logger.Err(err))
	}

	logger.Info("版本设置成功", logger.Int("version", forceVersion))
}
