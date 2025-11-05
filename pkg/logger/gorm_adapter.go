package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// GormLogger GORM日志适配器，实现gorm/logger.Interface接口
type GormLogger struct {
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// NewGormLogger 创建GORM日志适配器
func NewGormLogger() *GormLogger {
	return &GormLogger{
		LogLevel:                  gormlogger.Info,
		SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
		IgnoreRecordNotFoundError: true,                   // 忽略记录不存在错误
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 记录info级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		Info(fmt.Sprintf("GORM: "+msg, data...))
	}
}

// Warn 记录warn级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		Warn(fmt.Sprintf("GORM: "+msg, data...))
	}
}

// Error 记录error级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		Error(fmt.Sprintf("GORM: "+msg, data...))
	}
}

// Trace 记录SQL执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构造日志字段
	fields := []Field{
		Duration("elapsed", elapsed),
		Int64("rows", rows),
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// SQL执行错误
		fields = append(fields, String("sql", sql))
		fields = append(fields, Err(err))
		Error("GORM SQL错误", fields...)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		// 慢查询
		fields = append(fields, String("sql", sql))
		fields = append(fields, Duration("threshold", l.SlowThreshold))
		Warn("GORM 慢查询", fields...)

	case l.LogLevel >= gormlogger.Info:
		// 正常SQL执行
		fields = append(fields, String("sql", sql))
		Debug("GORM SQL", fields...)
	}
}
