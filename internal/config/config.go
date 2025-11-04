package config

import (
	"fmt"
	"strings"

	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Host string
	Port int
	Mode string // debug, release, test
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Driver   string // mysql, postgres
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func (d DatabaseConfig) DSN() string {
	// MySQL DSN format
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

type SecurityConfig struct {
	JWTSecret     string
	EncryptionKey string
}

type LoggerConfig struct {
	Level            string   // 日志级别: debug, info, warn, error, fatal
	Encoding         string   // 编码格式: json, console
	OutputPaths      []string // 输出路径
	ErrorOutputPaths []string // 错误输出路径
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
		{"security.jwt_secret", "SECURITY_JWT_SECRET"},
		{"security.encryption_key", "SECURITY_ENCRYPTION_KEY"},
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
