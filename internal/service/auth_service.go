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
	db           *gorm.DB
	jwtManager   *jwt.Manager
	redis        *redisClient.Client
	emailService *EmailService
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, jwtManager *jwt.Manager, redis *redisClient.Client, emailService *EmailService) *AuthService {
	return &AuthService{
		db:           db,
		jwtManager:   jwtManager,
		redis:        redis,
		emailService: emailService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"omitempty,email"`              // 邮箱（可选，但如果提供则必须验证）
	Code     string `json:"code" binding:"omitempty,len=6,numeric"`       // 验证码（可选，与email配合使用）
	Nickname string `json:"nickname" binding:"omitempty,min=1,max=50"`    // 昵称（可选，默认使用username）
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User *models.SafeUser `json:"user"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	ctx := context.Background()

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

	// 如果提供了邮箱，进行邮箱相关验证
	var emailVerified bool
	if req.Email != "" {
		// 检查邮箱是否已存在
		var profileCount int64
		if err := s.db.Model(&models.UserProfile{}).Where("email = ?", req.Email).Count(&profileCount).Error; err != nil {
			logger.Error("检查邮箱失败",
				logger.String("email", req.Email),
				logger.Err(err))
			return nil, errors.Wrap(errors.CodeDatabaseError, err)
		}
		if profileCount > 0 {
			return nil, errors.New(errors.CodeEmailExists, "")
		}

		// 如果提供了验证码，进行验证
		if req.Code != "" {
			if err := s.emailService.VerifyCode(ctx, req.Email, PurposeRegister, req.Code); err != nil {
				logger.Warn("注册时验证码验证失败",
					logger.String("email", req.Email),
					logger.Err(err))
				return nil, err
			}
			emailVerified = true
			logger.Info("注册时验证码验证成功",
				logger.String("email", req.Email))
		} else {
			// 提供了邮箱但没有验证码，要求验证
			logger.Warn("注册时提供了邮箱但未提供验证码",
				logger.String("email", req.Email))
			return nil, errors.New(errors.CodeInvalidParam, "提供邮箱时必须提供验证码")
		}
	}

	// 加密密码
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		logger.Error("密码加密失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeCryptoError, err)
	}

	// 开启事务创建用户和档案
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户,强制设置角色为普通用户,只有管理员才能提权
	user := &models.User{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		Status:       models.UserStatusActive,
		Role:         "user",
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		logger.Error("创建用户失败", logger.String("username", req.Username), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 如果提供了邮箱，创建用户档案
	if req.Email != "" {
		nickname := req.Nickname
		if nickname == "" {
			nickname = req.Username // 默认使用用户名作为昵称
		}

		profile := &models.UserProfile{
			UserID:        user.ID,
			Nickname:      nickname,
			Email:         req.Email,
			EmailVerified: emailVerified,
		}

		if err := tx.Create(profile).Error; err != nil {
			tx.Rollback()
			logger.Error("创建用户档案失败",
				logger.Uint("user_id", uint(user.ID)),
				logger.String("email", req.Email),
				logger.Err(err))
			return nil, errors.Wrap(errors.CodeDatabaseError, err)
		}

		logger.Info("用户档案创建成功",
			logger.Uint("user_id", uint(user.ID)),
			logger.String("email", req.Email),
			logger.Bool("email_verified", emailVerified))
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败",
			logger.String("username", req.Username),
			logger.Err(err))
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

	// Token互踢机制：一用户一Token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	expiration := s.jwtManager.GetExpiration()
	userTokenKey := makeUserTokenKey(user.UUID)

	// 查询并删除旧Token
	if oldToken, err := s.redis.Get(ctx, userTokenKey); err == nil && oldToken != "" {
		oldTokenKey := makeTokenKey(oldToken)
		if err := s.redis.Del(ctx, oldTokenKey); err != nil {
			logger.Warn("删除旧Token失败",
				logger.String("uuid", user.UUID),
				logger.Err(err))
		}
	}

	// 写入新Token（双向索引）
	tokenKey := makeTokenKey(token)
	if err := s.redis.Set(ctx, tokenKey, user.UUID, expiration); err != nil {
		logger.Error("存储token到Redis失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		// Redis存储失败不影响登录流程，仅记录错误
	}
	if err := s.redis.Set(ctx, userTokenKey, token, expiration); err != nil {
		logger.Error("存储user_token到Redis失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
	}

	logger.Info("用户登录成功", logger.String("uuid", user.UUID), logger.String("username", user.Username))

	return &LoginResponse{
		Token: token,
		User:  user.ToSafeUser(),
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 解析Token获取userUUID
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		logger.Warn("登出时解析token失败", logger.Err(err))
		// 即使解析失败，也尝试删除token
	}

	// 删除token（主索引）
	tokenKey := makeTokenKey(token)
	if err := s.redis.Del(ctx, tokenKey); err != nil {
		logger.Error("从Redis删除token失败", logger.Err(err))
		return errors.Wrap(errors.CodeCacheError, err)
	}

	// 删除user_token（反向索引）
	if claims != nil {
		userTokenKey := makeUserTokenKey(claims.UserUUID)
		if err := s.redis.Del(ctx, userTokenKey); err != nil {
			logger.Warn("从Redis删除user_token失败",
				logger.String("uuid", claims.UserUUID),
				logger.Err(err))
		}
	}

	logger.Info("用户登出成功")
	return nil
}

// makeTokenKey 生成token在Redis中的key
func makeTokenKey(token string) string {
	return fmt.Sprintf("token:%s", token)
}

// makeUserTokenKey 生成user_token在Redis中的key（用于反向索引）
func makeUserTokenKey(userUUID string) string {
	return fmt.Sprintf("user_token:%s", userUUID)
}
