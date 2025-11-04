package response

import (
	"time"

	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/gin-gonic/gin"
)

const (
	RequestIDKey = "X-Request-ID"
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}

func getTimestamp() int64 {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).UnixMilli()
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:      errors.CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: getTimestamp(),
	})
}

func Error(c *gin.Context, code int, message string) {
	if message == "" {
		message = errors.GetMessage(code)
	}
	c.JSON(200, Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
		Timestamp: getTimestamp(),
	})
}

func AppError(c *gin.Context, err *errors.AppError) {
	c.JSON(200, Response{
		Code:      err.Code,
		Message:   err.Message,
		RequestID: getRequestID(c),
		Timestamp: getTimestamp(),
	})
}

func ValidationError(c *gin.Context, message string) {
	Error(c, errors.CodeValidationFailed, message)
}

func InvalidParam(c *gin.Context, message string) {
	Error(c, errors.CodeInvalidParam, message)
}

func MissingParam(c *gin.Context, message string) {
	Error(c, errors.CodeMissingParam, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, errors.CodeUnauthorized, message)
}

func InvalidToken(c *gin.Context, message string) {
	Error(c, errors.CodeInvalidToken, message)
}

func TokenExpired(c *gin.Context, message string) {
	Error(c, errors.CodeTokenExpired, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, errors.CodeForbidden, message)
}

func InsufficientPermission(c *gin.Context, message string) {
	Error(c, errors.CodeInsufficientPermission, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, errors.CodeResourceNotFound, message)
}

func AlreadyExists(c *gin.Context, message string) {
	Error(c, errors.CodeResourceAlreadyExists, message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, errors.CodeInternalError, message)
}

func DatabaseError(c *gin.Context, message string) {
	Error(c, errors.CodeDatabaseError, message)
}
