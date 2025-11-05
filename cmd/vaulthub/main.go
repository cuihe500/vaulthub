package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuihe500/vaulthub/internal/api/routes"
	"github.com/cuihe500/vaulthub/internal/app"
	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/validator"
	"github.com/cuihe500/vaulthub/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// @title VaultHub API
// @version 1.0
// @description VaultHub 是一个密钥管理系统，旨在安全地存储、管理和轮换加密密钥、API 密钥及其他敏感凭证。
// @description
// @description ## 特性
// @description - 安全的密钥存储和管理
// @description - 密钥轮换支持
// @description - 细粒度的访问控制
// @description - 审计日志
// @description
// @description ## 认证
// @description 大部分接口需要JWT Bearer Token认证。
// @description 1. 调用 /api/v1/auth/register 注册账号
// @description 2. 调用 /api/v1/auth/login 获取 token
// @description 3. 在请求头中添加 Authorization: Bearer {token}
// @description
// @description ## 响应格式
// @description 所有接口返回统一的响应格式，HTTP状态码均为200，错误码在响应体的code字段中：
// @description ```json
// @description {
// @description   "code": 0,
// @description   "message": "success",
// @description   "data": {},
// @description   "requestId": "xxx",
// @description   "timestamp": 1762269490888
// @description }
// @description ```
// @termsOfService https://github.com/cuihe500/vaulthub
// @contact.name API Support
// @contact.email admin@thankseveryone.top
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer" 后跟一个空格和 JWT token。

var (
	// 配置文件路径
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "vaulthub",
	Short: "VaultHub - 安全密钥管理系统",
	Long:  `VaultHub 是一个密钥管理系统，旨在安全地存储、管理和轮换加密密钥、API 密钥及其他敏感凭证。`,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 VaultHub API 服务器",
	Run:   runServer,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get().String())
	},
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)

	// 添加配置文件路径 flag
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("命令执行失败", logger.Err(err))
	}
}

// runServer 启动服务器
func runServer(cmd *cobra.Command, args []string) {
	// 1. 加载配置
	cfg := loadConfig()

	// 2. 初始化日志
	if err := initLogger(cfg); err != nil {
		logger.Fatal("初始化日志失败", logger.Err(err))
	}
	defer logger.Sync()

	// 3. 初始化参数校验翻译器
	if err := validator.Init(); err != nil {
		logger.Fatal("初始化参数校验翻译器失败", logger.Err(err))
	}

	// 4. 打印版本信息
	logger.Info("启动 VaultHub",
		logger.String("version", version.Version),
		logger.String("commit", version.GitCommit),
		logger.String("build_time", version.BuildTime),
	)

	// 5. 初始化 Manager（包含所有外部连接）
	mgr := &app.Manager{}
	if err := mgr.Initialize(cfg); err != nil {
		logger.Fatal("初始化连接管理器失败", logger.Err(err))
	}
	defer mgr.Close()

	// 6. 自动执行数据库迁移
	if err := runAutoMigrate(cfg); err != nil {
		logger.Fatal("数据库迁移失败", logger.Err(err))
	}

	// 7. 初始化路由
	router := initRouter(cfg, mgr)

	// 8. 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router,
	}

	// 9. 启动服务器（非阻塞）
	go func() {
		logger.Info("启动服务器",
			logger.String("host", cfg.Server.Host),
			logger.Int("port", cfg.Server.Port),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", logger.Err(err))
		}
	}()

	// 10. 优雅关闭
	gracefulShutdown(srv)
}

// loadConfig 加载配置
func loadConfig() *config.Config {
	if configPath != "" {
		return config.LoadFromPath(configPath)
	}
	return config.Load()
}

// initLogger 初始化日志系统
func initLogger(cfg *config.Config) error {
	return logger.Init(logger.Config{
		Level:            cfg.Logger.Level,
		Encoding:         cfg.Logger.Encoding,
		OutputPaths:      cfg.Logger.OutputPaths,
		ErrorOutputPaths: cfg.Logger.ErrorOutputPaths,
	})
}

// runAutoMigrate 自动执行数据库迁移
func runAutoMigrate(cfg *config.Config) error {
	migrator, err := database.NewMigrator(cfg.Database)
	if err != nil {
		return fmt.Errorf("创建迁移器失败: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		return fmt.Errorf("迁移失败: %w", err)
	}

	logger.Info("数据库操作执行完毕")
	return nil
}

// initRouter 初始化路由
func initRouter(cfg *config.Config, mgr *app.Manager) *gin.Engine {
	// 设置 Gin 运行模式
	gin.SetMode(cfg.Server.Mode)

	// 禁用Gin默认的控制台颜色输出，使用统一日志接口
	gin.DisableConsoleColor()

	// 将Gin的默认输出重定向到项目日志
	gin.DefaultWriter = &logger.GinWriter{}
	gin.DefaultErrorWriter = &logger.GinWriter{}

	// 创建路由引擎
	router := gin.New()

	// 使用项目统一日志中间件，替代gin.Logger()和gin.Recovery()
	router.Use(logger.GinLogger())
	router.Use(logger.GinRecovery())

	// 注册业务路由，传入 Manager
	routes.Setup(router, mgr)

	return router
}

// gracefulShutdown 优雅关闭服务器
func gracefulShutdown(srv *http.Server) {
	// 创建信号通道
	quit := make(chan os.Signal, 1)
	// 监听中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	logger.Info("收到关闭信号", logger.String("signal", sig.String()))

	// 创建超时上下文（30秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务器
	logger.Info("正在关闭服务器...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭失败", logger.Err(err))
		return
	}

	logger.Info("服务器已安全关闭")
}
