package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger 返回Gin框架使用的日志中间件
// 替代gin.Logger()，使用项目统一的日志接口
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 请求结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 请求信息
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 根据状态码选择日志级别
		logFields := []Field{
			Int("status", statusCode),
			String("method", method),
			String("path", path),
			String("ip", clientIP),
			Duration("latency", latency),
		}

		if errorMessage != "" {
			logFields = append(logFields, String("error", errorMessage))
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			Error("HTTP请求", logFields...)
		case statusCode >= 400:
			Warn("HTTP请求", logFields...)
		default:
			Info("HTTP请求", logFields...)
		}
	}
}

// GinRecovery 返回Gin框架使用的恢复中间件
// 替代gin.Recovery()，使用项目统一的日志接口
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				Error("HTTP处理panic",
					String("path", c.Request.URL.Path),
					String("method", c.Request.Method),
					String("ip", c.ClientIP()),
					Any("error", err),
				)

				// 返回500错误
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": "内部服务器错误",
				})
			}
		}()

		c.Next()
	}
}

// GinWriter 实现io.Writer接口，将Gin的默认输出重定向到项目日志
// 用于捕获Gin框架自身的日志输出
type GinWriter struct{}

// Write 实现io.Writer接口
func (w *GinWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if msg != "" && msg != "\n" {
		Info(fmt.Sprintf("Gin框架: %s", msg))
	}
	return len(p), nil
}
