package config

import (
	"fmt"
	"strings"

	"github.com/cuihe500/vaulthub/pkg/logger"
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
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"` // Redis密码，为空则无密码
	DB       int    `mapstructure:"db"`       // Redis数据库编号
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type SecurityConfig struct {
	JWTSecret        string `mapstructure:"jwt_secret"`
	JWTExpiration    int    `mapstructure:"jwt_expiration"` // JWT过期时间（小时）
	EncryptionKey    string `mapstructure:"encryption_key"`
	CasbinModelPath  string `mapstructure:"casbin_model_path"` // Casbin模型文件路径
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
	// Set defaults
	setDefaults()

	// Configure environment variable binding
	setupEnvBinding()

	// Configure file reading
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

	// Read config file (optional, environment variables can provide all config)
	if err := viper.ReadInConfig(); err != nil {
		if configPath != "" {
			// 指定了配置文件但读取失败，这是错误
			logger.Fatal("读取配置文件失败", logger.String("path", configPath), logger.Err(err))
		}
		logger.Warn("配置文件未找到，使用默认值和环境变量", logger.Err(err))
	} else {
		logger.Info("加载配置文件", logger.String("path", viper.ConfigFileUsed()))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Fatal("解析配置失败", logger.Err(err))
	}

	return &cfg
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
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
	// Environment variables use underscore separator and uppercase
	// Example: server.host -> SERVER_HOST
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables explicitly for clarity
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
		{"redis.host", "REDIS_HOST"},
		{"redis.port", "REDIS_PORT"},
		{"redis.password", "REDIS_PASSWORD"},
		{"redis.db", "REDIS_DB"},
		{"security.jwt_secret", "SECURITY_JWT_SECRET"},
		{"security.jwt_expiration", "SECURITY_JWT_EXPIRATION"},
		{"security.encryption_key", "SECURITY_ENCRYPTION_KEY"},
		{"security.casbin_model_path", "SECURITY_CASBIN_MODEL_PATH"},
		{"logger.level", "LOGGER_LEVEL"},
		{"logger.encoding", "LOGGER_ENCODING"},
	}

	for _, binding := range bindings {
		if err := viper.BindEnv(binding[0], binding[1]); err != nil {
			logger.Fatal("环境变量绑定失败",
				logger.String("key", binding[0]),
				logger.String("env", binding[1]),
				logger.Err(err))
		}
	}
}
