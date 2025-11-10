package middleware

import (
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecurityPINCheckMiddleware 检查用户是否已设置安全密码
// 用于保护需要安全密码才能访问的接口（如加密数据相关操作）
// 注意：此中间件必须在 AuthMiddleware 之后使用，因为需要从上下文获取用户信息
func SecurityPINCheckMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取当前用户信息（由 AuthMiddleware 设置）
		userUUID, exists := c.Get(UserUUIDContextKey)
		if !exists {
			logger.Error("SecurityPINCheckMiddleware: 无法从上下文获取用户UUID")
			response.Error(c, errors.CodeUnauthorized, "未授权访问")
			c.Abort()
			return
		}

		// 查询用户的加密密钥配置
		var userKey models.UserEncryptionKey
		err := db.Where("user_uuid = ?", userUUID).First(&userKey).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 用户尚未创建加密密钥，需要先设置安全密码
				logger.Info("用户未创建加密密钥",
					logger.String("user_uuid", userUUID.(string)))
				response.Error(c, errors.CodeSecurityPINNotSet, "请先设置安全密码")
				c.Abort()
				return
			}
			// 数据库查询错误
			logger.Error("查询用户加密密钥失败",
				logger.String("user_uuid", userUUID.(string)),
				logger.Err(err))
			response.Error(c, errors.CodeDatabaseError, "系统错误")
			c.Abort()
			return
		}

		// 检查是否已设置安全密码
		if !userKey.HasSecurityPIN() {
			// 理论上不应该出现此情况（创建密钥时必须设置安全密码）
			// 但作为防御性编程，仍然检查
			logger.Warn("用户加密密钥存在但安全密码未设置",
				logger.String("user_uuid", userUUID.(string)))
			response.Error(c, errors.CodeSecurityPINNotSet, "请先设置安全密码")
			c.Abort()
			return
		}

		// 安全密码已设置，允许继续访问
		c.Next()
	}
}
