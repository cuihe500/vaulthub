package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局logger实例
var std *Logger

func init() {
	// 默认初始化为开发模式
	std = newDevelopmentLogger()
}

// Logger 日志接口封装
type Logger struct {
	zap *zap.Logger
}

// Init 初始化全局logger
func Init(cfg Config) error {
	logger, err := newLogger(cfg)
	if err != nil {
		return err
	}
	std = logger
	return nil
}

// newLogger 创建logger实例
func newLogger(cfg Config) (*Logger, error) {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 选择encoder
	var encoder zapcore.Encoder
	if cfg.Encoding == "json" {
		// JSON模式：机器友好，无颜色
		encoderCfg := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		// Console模式：人类友好，彩色输出
		encoderCfg := zapcore.EncoderConfig{
			TimeKey:       "时间",
			LevelKey:      "级别",
			NameKey:       "日志器",
			CallerKey:     "位置",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "消息",
			StacktraceKey: "堆栈",
			LineEnding:    zapcore.DefaultLineEnding,
			// 彩色大写级别，Console专用
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			// 友好的时间格式：2006-01-02 15:04:05
			EncodeTime: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			// 人类可读的时长格式：1.5s 而不是 1.500000
			EncodeDuration: zapcore.StringDurationEncoder,
			// 短路径调用位置
			EncodeCaller: zapcore.ShortCallerEncoder,
		}
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// 构建输出writer
	writeSyncer := zapcore.AddSync(os.Stdout)
	if len(cfg.OutputPaths) > 0 {
		// TODO: 支持文件输出
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{zap: zapLogger}, nil
}

// newDevelopmentLogger 创建开发模式logger
func newDevelopmentLogger() *Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger, _ := cfg.Build()
	return &Logger{zap: zapLogger.WithOptions(zap.AddCallerSkip(1))}
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug 输出调试日志
func Debug(msg string, fields ...Field) {
	std.Debug(msg, fields...)
}

// Info 输出信息日志
func Info(msg string, fields ...Field) {
	std.Info(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...Field) {
	std.Warn(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...Field) {
	std.Error(msg, fields...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(msg string, fields ...Field) {
	std.Fatal(msg, fields...)
}

// Debug 输出调试日志
func (l *Logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

// Info 输出信息日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

// Warn 输出警告日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

// Error 输出错误日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	return std.zap.Sync()
}
