package logger

// Config 日志配置
type Config struct {
	Level            string   // 日志级别: debug, info, warn, error, fatal
	Encoding         string   // 编码格式: json, console
	OutputPaths      []string // 输出路径，支持stdout和文件路径
	ErrorOutputPaths []string // 错误输出路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Level:            "info",
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}
