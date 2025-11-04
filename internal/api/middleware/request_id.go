package middleware

import (
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(response.RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(response.RequestIDKey, requestID)
		c.Header(response.RequestIDKey, requestID)

		c.Next()
	}
}
