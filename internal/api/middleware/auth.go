package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/jwt"
	"github.com/cuihe500/vaulthub/pkg/logger"
	redisClient "github.com/cuihe500/vaulthub/pkg/redis"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// UserContextKey 用户信息在context中的key
	UserContextKey = "user"
	// UserUUIDContextKey 用户UUID在context中的key
	UserUUIDContextKey = "user_uuid"
	// RoleContextKey 用户角色在context中的key
	RoleContextKey = "role"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtManager *jwt.Manager, db *gorm.DB, redis *redisClient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少授权头")
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "授权头格式无效")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			logger.Warn("JWT token验证失败", logger.Err(err))
			response.InvalidToken(c, "token无效或已过期")
			c.Abort()
			return
		}

		// 优先从Redis验证token
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		tokenKey := makeTokenKey(tokenString)
		userUUID, err := redis.Get(ctx, tokenKey)
		if err != nil {
			if err == goredis.Nil {
				// Redis中不存在token，可能已登出或过期
				logger.Warn("Token在Redis中不存在，可能已失效",
					logger.String("uuid", claims.UserUUID))
				response.InvalidToken(c, "token已失效")
				c.Abort()
				return
			}
			// Redis查询失败，降级到数据库验证
			logger.Warn("从Redis查询token失败，降级到数据库验证",
				logger.String("uuid", claims.UserUUID),
				logger.Err(err))
		} else {
			// Redis中存在token，验证UUID是否匹配
			if userUUID != claims.UserUUID {
				logger.Error("Token与用户UUID不匹配",
					logger.String("expected", claims.UserUUID),
					logger.String("actual", userUUID))
				response.InvalidToken(c, "token无效")
				c.Abort()
				return
			}
		}

		// 从数据库获取用户信息（确保用户仍然存在且状态正常）
		var user models.User
		if err := db.Where("uuid = ?", claims.UserUUID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Unauthorized(c, "用户不存在")
			} else {
				logger.Error("查询用户失败", logger.String("uuid", claims.UserUUID), logger.Err(err))
				response.InternalError(c, "查询用户失败")
			}
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.CanOperate() {
			var message string
			if user.IsDisabled() {
				message = errors.GetMessage(errors.CodeAccountDisabled)
			} else if user.IsLocked() {
				message = errors.GetMessage(errors.CodeAccountLocked)
			} else {
				message = errors.GetMessage(errors.CodeAccountNotActivated)
			}
			response.Unauthorized(c, message)
			c.Abort()
			return
		}

		// 将用户信息存入context
		c.Set(UserContextKey, &user)
		c.Set(UserUUIDContextKey, user.UUID)
		c.Set(RoleContextKey, user.Role)

		c.Next()
	}
}

// GetCurrentUser 从context获取当前用户
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	return user.(*models.User), true
}

// GetCurrentUserUUID 从context获取当前用户UUID
func GetCurrentUserUUID(c *gin.Context) (string, bool) {
	uuid, exists := c.Get(UserUUIDContextKey)
	if !exists {
		return "", false
	}
	return uuid.(string), true
}

// GetCurrentUserRole 从context获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleContextKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// makeTokenKey 生成token在Redis中的key
func makeTokenKey(token string) string {
	return fmt.Sprintf("token:%s", token)
}
