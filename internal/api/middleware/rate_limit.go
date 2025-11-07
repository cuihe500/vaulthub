package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	redisClient "github.com/cuihe500/vaulthub/pkg/redis"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// RateLimitConfig 限流配置（从数据库读取）
type RateLimitConfig struct {
	Requests int    `json:"requests"` // 时间窗口内最大请求数
	Period   string `json:"period"`   // 时间周期：second, minute, hour
}

// ToRedisRate 转换为redis_rate.Limit
func (cfg RateLimitConfig) ToRedisRate() redis_rate.Limit {
	switch cfg.Period {
	case "second":
		return redis_rate.PerSecond(cfg.Requests)
	case "minute":
		return redis_rate.PerMinute(cfg.Requests)
	case "hour":
		return redis_rate.PerHour(cfg.Requests)
	default:
		// 默认每秒
		return redis_rate.PerSecond(cfg.Requests)
	}
}

// getEndpointRateLimit 获取接口限流配置（优先接口级，回退默认）
func getEndpointRateLimit(configMgr *config.ConfigManager, path string) RateLimitConfig {
	// 1. 尝试读取接口级配置
	endpointsJSON := configMgr.GetWithDefault("rate_limit.endpoints", "")
	if endpointsJSON != "" {
		var endpoints map[string]RateLimitConfig
		if err := json.Unmarshal([]byte(endpointsJSON), &endpoints); err == nil {
			if cfg, exists := endpoints[path]; exists {
				return cfg
			}
		}
	}

	// 2. 回退到默认配置
	defaultJSON := configMgr.GetWithDefault("rate_limit.default", "")
	if defaultJSON != "" {
		var defaultCfg RateLimitConfig
		if err := json.Unmarshal([]byte(defaultJSON), &defaultCfg); err == nil {
			return defaultCfg
		}
	}

	// 3. 硬编码兜底（理论上不应该到这里）
	return RateLimitConfig{
		Requests: 5,
		Period:   "second",
	}
}

// RateLimitMiddleware 限流中间件（基于redis_rate实现，使用GCRA算法，配置从数据库动态读取）
func RateLimitMiddleware(redis *redisClient.Client, configMgr *config.ConfigManager) gin.HandlerFunc {
	// 创建限流器
	limiter := redis_rate.NewLimiter(redis.GetUniversalClient())

	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// 从ConfigManager读取限流配置
		cfg := getEndpointRateLimit(configMgr, path)

		// 生成限流key：ip:path
		key := fmt.Sprintf("rate_limit:%s:%s", clientIP, path)

		// 执行限流检查
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		res, err := limiter.Allow(ctx, key, cfg.ToRedisRate())
		if err != nil {
			logger.Error("限流检查失败",
				logger.String("ip", clientIP),
				logger.String("path", path),
				logger.Err(err))
			// 限流检查失败不阻断请求，继续执行
			c.Next()
			return
		}

		// 检查是否允许请求
		if res.Allowed == 0 {
			logger.Warn("请求被限流",
				logger.String("ip", clientIP),
				logger.String("path", path),
				logger.Int("remaining", res.Remaining),
				logger.Int("limit", res.Limit.Burst))
			response.AppError(c, errors.New(errors.CodeTooManyRequests, ""))
			c.Abort()
			return
		}

		c.Next()
	}
}
