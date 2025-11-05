package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultExpiration 默认过期时间（24小时）
	DefaultExpiration = 24 * time.Hour
	// Issuer JWT签发者
	Issuer = "vaulthub"
)

// Claims JWT自定义声明
type Claims struct {
	UserUUID string `json:"user_uuid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Manager JWT管理器
type Manager struct {
	secret     []byte
	expiration time.Duration
}

// NewManager 创建JWT管理器
func NewManager(secret string, expiration time.Duration) *Manager {
	if expiration == 0 {
		expiration = DefaultExpiration
	}
	return &Manager{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// GenerateToken 生成JWT token
func (m *Manager) GenerateToken(userUUID, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserUUID: userUUID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken 解析JWT token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// ValidateToken 验证token是否有效
func (m *Manager) ValidateToken(tokenString string) bool {
	_, err := m.ParseToken(tokenString)
	return err == nil
}

// GetExpiration 获取token过期时间
func (m *Manager) GetExpiration() time.Duration {
	return m.expiration
}
