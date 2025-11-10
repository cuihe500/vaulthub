package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// 审计上下文key
const (
	AuditActionTypeKey   = "audit_action_type"
	AuditResourceTypeKey = "audit_resource_type"
	AuditResourceUUIDKey = "audit_resource_uuid"
	AuditResourceNameKey = "audit_resource_name"
	AuditDetailsKey      = "audit_details"
)

// responseWriter 包装gin.ResponseWriter以捕获响应状态码
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware 审计中间件
// 自动记录所有经过认证的请求，业务handler可通过context设置额外的审计信息
func AuditMiddleware(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只审计已认证的请求
		userUUID, exists := GetCurrentUserUUID(c)

		// 调试日志：记录中间件是否执行
		logger.Debug("审计中间件执行",
			logger.String("path", c.FullPath()),
			logger.String("method", c.Request.Method),
			logger.Bool("authenticated", exists),
			logger.String("user_uuid", userUUID))

		if !exists {
			logger.Debug("跳过审计：用户未认证", logger.String("path", c.FullPath()))
			c.Next()
			return
		}

		username, _ := c.Get(UserContextKey)
		var usernameStr string
		if user, ok := username.(*models.User); ok {
			usernameStr = user.Username
		}

		// 获取请求ID（用于追踪）
		requestID := c.GetString(response.RequestIDKey)

		// 记录请求开始时间
		startTime := time.Now()

		// 包装ResponseWriter以捕获状态码
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 请求处理完成，提取审计信息
		actionType := c.GetString(AuditActionTypeKey)
		resourceType := c.GetString(AuditResourceTypeKey)

		// 如果handler没有设置审计信息，根据HTTP方法推断
		if actionType == "" {
			actionType = inferActionType(c.Request.Method)
		}

		// 如果没有设置资源类型，从路径推断（可选）
		if resourceType == "" {
			resourceType = inferResourceType(c.FullPath())
		}

		// 提取其他审计信息
		resourceUUID := stringPtrOrNil(c.GetString(AuditResourceUUIDKey))
		resourceName := stringPtrOrNil(c.GetString(AuditResourceNameKey))
		details := c.GetString(AuditDetailsKey)

		// 判断操作状态
		status := models.AuditSuccess
		var errorCode *int
		var errorMessage *string

		if writer.statusCode >= 400 {
			status = models.AuditFailed
			code := writer.statusCode
			errorCode = &code

			// 尝试从响应体解析错误信息
			if writer.body.Len() > 0 {
				msg := writer.body.String()
				if len(msg) > 500 {
					msg = msg[:500] + "..."
				}
				errorMessage = &msg
			}
		}

		// 构造审计日志
		auditLog := &models.AuditLog{
			UserUUID:     userUUID,
			Username:     usernameStr,
			ActionType:   models.ActionType(actionType),
			ResourceType: models.ResourceType(resourceType),
			ResourceUUID: resourceUUID,
			ResourceName: resourceName,
			Status:       status,
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
			IPAddress:    stringPtrOrNil(c.ClientIP()),
			UserAgent:    stringPtrOrNil(c.Request.UserAgent()),
			RequestID:    stringPtrOrNil(requestID),
			CreatedAt:    startTime.UTC(),
		}

		// 添加详细信息（如果handler设置了）
		if details != "" {
			auditLog.Details = details
		}

		// 记录审计日志信息用于调试
		logger.Debug("写入审计日志",
			logger.String("user_uuid", auditLog.UserUUID),
			logger.String("username", auditLog.Username),
			logger.String("action_type", string(auditLog.ActionType)),
			logger.String("resource_type", string(auditLog.ResourceType)),
			logger.String("status", string(auditLog.Status)),
			logger.String("method", c.Request.Method),
			logger.String("path", c.FullPath()))

		// 异步写入审计日志
		auditService.LogAsync(auditLog)
	}
}

// SetAuditAction 设置审计操作类型
func SetAuditAction(c *gin.Context, actionType models.ActionType) {
	c.Set(AuditActionTypeKey, string(actionType))
}

// SetAuditResource 设置审计资源信息
func SetAuditResource(c *gin.Context, resourceType models.ResourceType, resourceUUID, resourceName string) {
	c.Set(AuditResourceTypeKey, string(resourceType))
	if resourceUUID != "" {
		c.Set(AuditResourceUUIDKey, resourceUUID)
	}
	if resourceName != "" {
		c.Set(AuditResourceNameKey, resourceName)
	}
}

// SetAuditDetails 设置审计详细信息
func SetAuditDetails(c *gin.Context, details interface{}) {
	c.Set(AuditDetailsKey, details)
}

// inferActionType 根据HTTP方法推断操作类型
func inferActionType(method string) string {
	switch method {
	case "POST":
		return string(models.ActionCreate)
	case "PUT", "PATCH":
		return string(models.ActionUpdate)
	case "DELETE":
		return string(models.ActionDelete)
	case "GET":
		return string(models.ActionAccess)
	default:
		return string(models.ActionAccess)
	}
}

// inferResourceType 从路径推断资源类型
func inferResourceType(path string) string {
	// 简单的路径匹配推断
	// 可以根据实际路由规则扩展
	if len(path) > 0 {
		// 例如：/api/v1/secrets -> secret
		// 这里简化处理，实际应该更精确
		return "unknown"
	}
	return "unknown"
}

// stringPtrOrNil 如果字符串为空返回nil，否则返回指针
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ReadRequestBody 读取并恢复请求体（用于记录请求内容）
func ReadRequestBody(c *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	// 恢复请求体，以便后续handler可以读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
