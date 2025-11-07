package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Security SecurityConfig `mapstructure:"security"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"` // mysql, postgres
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// DSN 返回用于正常数据库连接的DSN，不包含multiStatements参数以提高安全性
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// MigrationDSN 返回用于数据库迁移的DSN，包含multiStatements=true以支持迁移文件中的多条SQL语句
func (d DatabaseConfig) MigrationDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

type RedisConfig struct {
	Mode     string `mapstructure:"mode"`     // 部署模式: standalone(单机), sentinel(哨兵), cluster(集群)
	Host     string `mapstructure:"host"`     // 单机模式使用
	Port     int    `mapstructure:"port"`     // 单机模式使用
	Password string `mapstructure:"password"` // Redis密码，为空则无密码
	DB       int    `mapstructure:"db"`       // Redis数据库编号，集群模式下无效

	// 哨兵模式配置
	MasterName       string   `mapstructure:"master_name"`       // 哨兵模式主节点名称
	Sentinels        []string `mapstructure:"sentinels"`         // 哨兵节点地址列表，格式: ["host1:port1", "host2:port2"]
	SentinelPassword string   `mapstructure:"sentinel_password"` // 哨兵节点密码，为空则无密码

	// 集群模式配置
	Addrs []string `mapstructure:"addrs"` // 集群节点地址列表，格式: ["host1:port1", "host2:port2"]
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type SecurityConfig struct {
	JWTSecret       string `mapstructure:"jwt_secret"`
	JWTExpiration   int    `mapstructure:"jwt_expiration"` // JWT过期时间（小时）
	EncryptionKey   string `mapstructure:"encryption_key"`
	CasbinModelPath string `mapstructure:"casbin_model_path"` // Casbin模型文件路径
	AdminUsername   string `mapstructure:"admin_username"`    // 超级管理员用户名（首次启动时创建）
	AdminPassword   string `mapstructure:"admin_password"`    // 超级管理员密码（首次启动时创建）
}

type LoggerConfig struct {
	Level            string   `mapstructure:"level"`              // 日志级别: debug, info, warn, error, fatal
	Encoding         string   `mapstructure:"encoding"`           // 编码格式: json, console
	OutputPaths      []string `mapstructure:"output_paths"`       // 输出路径
	ErrorOutputPaths []string `mapstructure:"error_output_paths"` // 错误输出路径
}

func Load() *Config {
	return load("")
}

// LoadFromPath 从指定路径加载配置
func LoadFromPath(path string) *Config {
	return load(path)
}

func load(configPath string) *Config {
	// 设置默认值
	setDefaults()

	// 配置环境变量绑定
	setupEnvBinding()

	// 配置文件读取
	if configPath != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(configPath)
	} else {
		// 使用默认配置文件搜索
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// 读取配置文件（可选，环境变量可以提供所有配置）
	// 注意：此阶段使用标准库fmt而非项目logger，避免循环依赖
	if err := viper.ReadInConfig(); err != nil {
		if configPath != "" {
			// 指定了配置文件但读取失败，这是错误
			fmt.Fprintf(os.Stderr, "读取配置文件失败: path=%s, error=%v\n", configPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "配置文件未找到，使用默认值和环境变量: %v\n", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "解析配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("加载配置文件并解析成功: %s\n", viper.ConfigFileUsed())

	return &cfg
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("redis.mode", "standalone") // 默认单机模式，保持向后兼容
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("security.jwt_expiration", 24) // 默认24小时
	viper.SetDefault("security.casbin_model_path", "./configs/rbac_model.conf")
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "console")
	viper.SetDefault("logger.output_paths", []string{"stdout"})
	viper.SetDefault("logger.error_output_paths", []string{"stderr"})
}

func setupEnvBinding() {
	// 环境变量使用下划线分隔符和大写
	// 示例: server.host -> SERVER_HOST
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 显式绑定特定环境变量以提高清晰度
	bindings := [][2]string{
		{"server.host", "SERVER_HOST"},
		{"server.port", "SERVER_PORT"},
		{"server.mode", "SERVER_MODE"},
		{"database.driver", "DATABASE_DRIVER"},
		{"database.host", "DATABASE_HOST"},
		{"database.port", "DATABASE_PORT"},
		{"database.name", "DATABASE_NAME"},
		{"database.user", "DATABASE_USER"},
		{"database.password", "DATABASE_PASSWORD"},
		{"redis.mode", "REDIS_MODE"},
		{"redis.host", "REDIS_HOST"},
		{"redis.port", "REDIS_PORT"},
		{"redis.password", "REDIS_PASSWORD"},
		{"redis.db", "REDIS_DB"},
		{"redis.master_name", "REDIS_MASTER_NAME"},
		{"redis.sentinels", "REDIS_SENTINELS"},
		{"redis.sentinel_password", "REDIS_SENTINEL_PASSWORD"},
		{"redis.addrs", "REDIS_ADDRS"},
		{"security.jwt_secret", "SECURITY_JWT_SECRET"},
		{"security.jwt_expiration", "SECURITY_JWT_EXPIRATION"},
		{"security.encryption_key", "SECURITY_ENCRYPTION_KEY"},
		{"security.casbin_model_path", "SECURITY_CASBIN_MODEL_PATH"},
		{"security.admin_username", "SECURITY_ADMIN_USERNAME"},
		{"security.admin_password", "SECURITY_ADMIN_PASSWORD"},
		{"logger.level", "LOGGER_LEVEL"},
		{"logger.encoding", "LOGGER_ENCODING"},
	}

	// 注意：此阶段使用标准库fmt而非项目logger，避免循环依赖
	for _, binding := range bindings {
		if err := viper.BindEnv(binding[0], binding[1]); err != nil {
			fmt.Fprintf(os.Stderr, "环境变量绑定失败: key=%s, env=%s, error=%v\n", binding[0], binding[1], err)
			os.Exit(1)
		}
	}
}
