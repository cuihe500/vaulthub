package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/crypto"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/jwt"
	"github.com/cuihe500/vaulthub/pkg/logger"
	redisClient "github.com/cuihe500/vaulthub/pkg/redis"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db         *gorm.DB
	jwtManager *jwt.Manager
	redis      *redisClient.Client
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, jwtManager *jwt.Manager, redis *redisClient.Client) *AuthService {
	return &AuthService{
		db:         db,
		jwtManager: jwtManager,
		redis:      redis,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role,omitempty"` // 可选，默认为user
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User *models.SafeUser `json:"user"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// 验证密码强度
	if !crypto.ValidatePasswordStrength(req.Password) {
		return nil, errors.New(errors.CodeWeakPassword, "")
	}

	// 检查用户名是否已存在
	var count int64
	if err := s.db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		logger.Error("检查用户名失败", logger.String("username", req.Username), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}
	if count > 0 {
		return nil, errors.New(errors.CodeUsernameExists, "")
	}

	// 加密密码
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		logger.Error("密码加密失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeCryptoError, err)
	}

	// 设置默认角色
	role := req.Role
	if role == "" {
		role = "user"
	}

	// 创建用户
	user := &models.User{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		Status:       models.UserStatusActive,
		Role:         role,
	}

	if err := s.db.Create(user).Error; err != nil {
		logger.Error("创建用户失败", logger.String("username", req.Username), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("用户注册成功", logger.String("uuid", user.UUID), logger.String("username", user.Username))

	return &RegisterResponse{
		User: user.ToSafeUser(),
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string           `json:"token"`
	User  *models.SafeUser `json:"user"`
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 查找用户
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeInvalidCredentials, "")
		}
		logger.Error("查询用户失败", logger.String("username", req.Username), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 验证密码
	if !crypto.VerifyPassword(req.Password, user.PasswordHash) {
		logger.Warn("密码验证失败", logger.String("username", req.Username))
		return nil, errors.New(errors.CodeInvalidCredentials, "")
	}

	// 检查用户状态
	if user.IsDisabled() {
		return nil, errors.New(errors.CodeAccountDisabled, "")
	}
	if user.IsLocked() {
		return nil, errors.New(errors.CodeAccountLocked, "")
	}
	if !user.IsActive() {
		return nil, errors.New(errors.CodeAccountNotActivated, "")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.db.Model(&user).Update("last_login_at", now).Error; err != nil {
		logger.Error("更新最后登录时间失败", logger.String("uuid", user.UUID), logger.Err(err))
		// 不返回错误，继续登录流程
	}

	// 生成JWT token
	token, err := s.jwtManager.GenerateToken(user.UUID, user.Username, user.Role)
	if err != nil {
		logger.Error("生成JWT token失败", logger.String("uuid", user.UUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeInternalError, err)
	}

	// 将token存入Redis，使用token过期时间作为TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokenKey := makeTokenKey(token)
	expiration := s.jwtManager.GetExpiration()
	if err := s.redis.Set(ctx, tokenKey, user.UUID, expiration); err != nil {
		logger.Error("存储token到Redis失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		// Redis存储失败不影响登录流程，仅记录错误
	}

	logger.Info("用户登录成功", logger.String("uuid", user.UUID), logger.String("username", user.Username))

	return &LoginResponse{
		Token: token,
		User:  user.ToSafeUser(),
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	// 从Redis删除token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokenKey := makeTokenKey(token)
	if err := s.redis.Del(ctx, tokenKey); err != nil {
		logger.Error("从Redis删除token失败", logger.Err(err))
		return errors.Wrap(errors.CodeCacheError, err)
	}

	logger.Info("用户登出成功")
	return nil
}

// makeTokenKey 生成token在Redis中的key
func makeTokenKey(token string) string {
	return fmt.Sprintf("token:%s", token)
}
