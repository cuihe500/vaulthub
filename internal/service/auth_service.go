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

// LoginWithEmailRequest 邮箱验证码登录请求
type LoginWithEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6,numeric"`
}

// LoginWithEmail 邮箱验证码登录
func (s *AuthService) LoginWithEmail(req *LoginWithEmailRequest) (*LoginResponse, error) {
	ctx := context.Background()

	// 1. 验证验证码
	if err := s.emailService.VerifyCode(ctx, req.Email, PurposeLogin, req.Code); err != nil {
		logger.Warn("邮箱验证码登录验证失败",
			logger.String("email", req.Email),
			logger.Err(err))
		return nil, err
	}

	// 2. 通过邮箱查找用户档案
	var profile models.UserProfile
	if err := s.db.Where("email = ?", req.Email).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("邮箱对应的用户不存在",
				logger.String("email", req.Email))
			return nil, errors.New(errors.CodeResourceNotFound, "用户不存在")
		}
		logger.Error("查询用户档案失败",
			logger.String("email", req.Email),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 3. 查找关联的用户
	var user models.User
	if err := s.db.Where("id = ?", profile.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("用户档案关联的用户不存在",
				logger.Uint("user_id", uint(profile.UserID)))
			return nil, errors.New(errors.CodeResourceNotFound, "用户不存在")
		}
		logger.Error("查询用户失败",
			logger.Uint("user_id", uint(profile.UserID)),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 4. 检查用户状态
	if user.IsDisabled() {
		return nil, errors.New(errors.CodeAccountDisabled, "")
	}
	if user.IsLocked() {
		return nil, errors.New(errors.CodeAccountLocked, "")
	}
	if !user.IsActive() {
		return nil, errors.New(errors.CodeAccountNotActivated, "")
	}

	// 5. 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.db.Model(&user).Update("last_login_at", now).Error; err != nil {
		logger.Error("更新最后登录时间失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		// 不返回错误，继续登录流程
	}

	// 6. 标记邮箱已验证（如果还未验证）
	if !profile.EmailVerified {
		if err := s.emailService.MarkEmailVerified(ctx, req.Email); err != nil {
			logger.Warn("标记邮箱验证状态失败",
				logger.String("email", req.Email),
				logger.Err(err))
			// 不影响登录流程
		}
	}

	// 7. 生成JWT token
	token, err := s.jwtManager.GenerateToken(user.UUID, user.Username, user.Role)
	if err != nil {
		logger.Error("生成JWT token失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeInternalError, err)
	}

	// 8. Token互踢机制：一用户一Token
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	expiration := s.jwtManager.GetExpiration()
	userTokenKey := makeUserTokenKey(user.UUID)

	// 查询并删除旧Token
	if oldToken, err := s.redis.Get(ctx2, userTokenKey); err == nil && oldToken != "" {
		oldTokenKey := makeTokenKey(oldToken)
		if err := s.redis.Del(ctx2, oldTokenKey); err != nil {
			logger.Warn("删除旧Token失败",
				logger.String("uuid", user.UUID),
				logger.Err(err))
		}
	}

	// 写入新Token（双向索引）
	tokenKey := makeTokenKey(token)
	if err := s.redis.Set(ctx2, tokenKey, user.UUID, expiration); err != nil {
		logger.Error("存储token到Redis失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		// Redis存储失败不影响登录流程，仅记录错误
	}
	if err := s.redis.Set(ctx2, userTokenKey, token, expiration); err != nil {
		logger.Error("存储user_token到Redis失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
	}

	logger.Info("邮箱验证码登录成功",
		logger.String("uuid", user.UUID),
		logger.String("username", user.Username),
		logger.String("email", req.Email))

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

// RequestPasswordResetRequest 请求密码重置
type RequestPasswordResetRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Domain string `json:"domain" binding:"required"` // 前端当前访问的域名（含协议，如 https://example.com）
}

// RequestPasswordResetResponse 请求密码重置响应
type RequestPasswordResetResponse struct {
	Message string `json:"message"`
}

// RequestPasswordReset 请求密码重置（生成token并发送邮件）
// 注意：为防止用户枚举攻击，无论邮箱是否存在都返回成功
func (s *AuthService) RequestPasswordReset(req *RequestPasswordResetRequest, baseURL string) (*RequestPasswordResetResponse, error) {
	ctx := context.Background()

	// 检查频率限制（5分钟内同一邮箱只能发送一次）
	limitKey := fmt.Sprintf("password_reset_limit:%s", req.Email)
	exists, err := s.redis.Exists(ctx, limitKey)
	if err != nil {
		logger.Error("检查密码重置频率限制失败",
			logger.String("email", req.Email),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeCacheError, err)
	}
	if exists > 0 {
		logger.Warn("密码重置请求过于频繁",
			logger.String("email", req.Email))
		return nil, errors.New(errors.CodeTooManyRequests, "请求过于频繁，请5分钟后再试")
	}

	// 查找用户档案
	var profile models.UserProfile
	if err := s.db.Where("email = ?", req.Email).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 邮箱不存在，但仍然返回成功（防止用户枚举）
			logger.Info("密码重置请求的邮箱不存在",
				logger.String("email", req.Email))
			// 设置频率限制，防止滥用
			s.redis.Set(ctx, limitKey, "1", 5*time.Minute)
			return &RequestPasswordResetResponse{
				Message: "如果该邮箱已注册，您将收到密码重置邮件",
			}, nil
		}
		logger.Error("查询用户档案失败",
			logger.String("email", req.Email),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 查找关联的用户
	var user models.User
	if err := s.db.Where("id = ?", profile.UserID).First(&user).Error; err != nil {
		logger.Error("查询用户失败",
			logger.Uint("user_id", uint(profile.UserID)),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 检查用户状态
	if user.IsDisabled() {
		logger.Warn("尝试为已禁用的用户重置密码",
			logger.String("uuid", user.UUID))
		// 返回通用消息，不泄露账户状态
		s.redis.Set(ctx, limitKey, "1", 5*time.Minute)
		return &RequestPasswordResetResponse{
			Message: "如果该邮箱已注册，您将收到密码重置邮件",
		}, nil
	}

	// 生成重置token（使用UUID）
	resetToken := uuid.New().String()
	tokenHash, err := crypto.HashPassword(resetToken)
	if err != nil {
		logger.Error("生成token哈希失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeCryptoError, err)
	}

	// 计算过期时间（30分钟）
	expiresAt := time.Now().UTC().Add(30 * time.Minute)

	// 创建密码重置token记录
	resetTokenRecord := &models.PasswordResetToken{
		UUID:      uuid.New().String(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(resetTokenRecord).Error; err != nil {
		logger.Error("创建密码重置token记录失败",
			logger.Uint("user_id", uint(user.ID)),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 构建重置链接
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, resetToken)

	// 发送邮件
	if err := s.emailService.SendPasswordResetLink(req.Email, resetURL, 30); err != nil {
		logger.Error("发送密码重置邮件失败",
			logger.String("email", req.Email),
			logger.Err(err))
		return nil, err
	}

	// 设置频率限制（5分钟）
	if err := s.redis.Set(ctx, limitKey, "1", 5*time.Minute); err != nil {
		logger.Error("设置密码重置频率限制失败",
			logger.String("email", req.Email),
			logger.Err(err))
		// 不影响主流程
	}

	logger.Info("密码重置邮件发送成功",
		logger.String("uuid", user.UUID),
		logger.String("email", req.Email))

	return &RequestPasswordResetResponse{
		Message: "如果该邮箱已注册，您将收到密码重置邮件",
	}, nil
}

// VerifyResetTokenRequest 验证重置token请求
type VerifyResetTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// VerifyResetTokenResponse 验证重置token响应
type VerifyResetTokenResponse struct {
	Valid bool `json:"valid"`
}

// VerifyResetToken 验证重置token是否有效
func (s *AuthService) VerifyResetToken(token string) error {
	// 查找所有未使用的token记录（需要遍历因为token是哈希存储的）
	var tokens []models.PasswordResetToken
	if err := s.db.Where("used_at IS NULL AND deleted_at IS NULL").Find(&tokens).Error; err != nil {
		logger.Error("查询密码重置token失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 验证token
	for _, t := range tokens {
		if crypto.VerifyPassword(token, t.TokenHash) {
			// 检查是否过期
			if t.IsExpired() {
				logger.Warn("密码重置token已过期",
					logger.String("token_uuid", t.UUID))
				return errors.New(errors.CodeTokenExpired, "重置链接已过期")
			}

			// Token有效
			logger.Info("密码重置token验证成功",
				logger.String("token_uuid", t.UUID),
				logger.Uint("user_id", uint(t.UserID)))
			return nil
		}
	}

	// Token不存在或已使用
	logger.Warn("密码重置token无效")
	return errors.New(errors.CodeInvalidToken, "无效的重置链接")
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPassword 重置密码（使用token）
func (s *AuthService) ResetPassword(req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	// 验证新密码强度
	if !crypto.ValidatePasswordStrength(req.NewPassword) {
		return nil, errors.New(errors.CodeWeakPassword, "")
	}

	// 查找所有未使用的token记录
	var tokens []models.PasswordResetToken
	if err := s.db.Where("used_at IS NULL AND deleted_at IS NULL").Find(&tokens).Error; err != nil {
		logger.Error("查询密码重置token失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 验证token并找到对应的记录
	var matchedToken *models.PasswordResetToken
	for i := range tokens {
		if crypto.VerifyPassword(req.Token, tokens[i].TokenHash) {
			matchedToken = &tokens[i]
			break
		}
	}

	if matchedToken == nil {
		logger.Warn("密码重置token无效或已使用")
		return nil, errors.New(errors.CodeInvalidToken, "无效的重置链接")
	}

	// 检查是否过期
	if matchedToken.IsExpired() {
		logger.Warn("密码重置token已过期",
			logger.String("token_uuid", matchedToken.UUID))
		return nil, errors.New(errors.CodeTokenExpired, "重置链接已过期")
	}

	// 查找用户
	var user models.User
	if err := s.db.Where("id = ?", matchedToken.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("密码重置token对应的用户不存在",
				logger.Uint("user_id", uint(matchedToken.UserID)))
			return nil, errors.New(errors.CodeResourceNotFound, "用户不存在")
		}
		logger.Error("查询用户失败",
			logger.Uint("user_id", uint(matchedToken.UserID)),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 检查用户状态
	if user.IsDisabled() {
		logger.Warn("尝试为已禁用的用户重置密码",
			logger.String("uuid", user.UUID))
		return nil, errors.New(errors.CodeAccountDisabled, "")
	}

	// 加密新密码
	newPasswordHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("密码加密失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeCryptoError, err)
	}

	// 开启事务：更新密码和标记token已使用
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户密码
	if err := tx.Model(&user).Update("password_hash", newPasswordHash).Error; err != nil {
		tx.Rollback()
		logger.Error("更新用户密码失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 标记token已使用
	now := time.Now().UTC()
	if err := tx.Model(matchedToken).Update("used_at", now).Error; err != nil {
		tx.Rollback()
		logger.Error("标记token已使用失败",
			logger.String("token_uuid", matchedToken.UUID),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交事务失败",
			logger.String("uuid", user.UUID),
			logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("密码重置成功",
		logger.String("uuid", user.UUID),
		logger.String("username", user.Username))

	return &ResetPasswordResponse{
		Message: "密码重置成功",
	}, nil
}
