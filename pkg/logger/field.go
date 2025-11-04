package logger

import (
	"time"

	"go.uber.org/zap"
)

// Field 日志字段类型，直接使用zap.Field
type Field = zap.Field

// String 构造字符串字段
func String(key, val string) Field {
	return zap.String(key, val)
}

// Int 构造整数字段
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 构造int64字段
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Uint 构造无符号整数字段
func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

// Float64 构造浮点数字段
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool 构造布尔字段
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Time 构造时间字段
func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

// Duration 构造时间段字段
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Err 构造错误字段
func Err(err error) Field {
	return zap.Error(err)
}

// Any 构造任意类型字段，使用反射
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Strings 构造字符串数组字段
func Strings(key string, val []string) Field {
	return zap.Strings(key, val)
}

// Ints 构造整数数组字段
func Ints(key string, val []int) Field {
	return zap.Ints(key, val)
}
