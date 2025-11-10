package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/email"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	redisClient "github.com/cuihe500/vaulthub/pkg/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// VerificationPurpose 验证码用途
type VerificationPurpose string

const (
	PurposeRegister      VerificationPurpose = "register"       // 注册
	PurposeLogin         VerificationPurpose = "login"          // 登录
	PurposeResetPassword VerificationPurpose = "reset_password" // 重置密码
	PurposeChangeEmail   VerificationPurpose = "change_email"   // 修改邮箱
)

// String 转换为中文描述
func (p VerificationPurpose) String() string {
	switch p {
	case PurposeRegister:
		return "注册"
	case PurposeLogin:
		return "登录"
	case PurposeResetPassword:
		return "重置密码"
	case PurposeChangeEmail:
		return "修改邮箱"
	default:
		return "验证"
	}
}

// EmailService 邮件服务
type EmailService struct {
	db            *gorm.DB
	redis         *redisClient.Client
	configManager *config.ConfigManager
}

// NewEmailService 创建邮件服务实例
func NewEmailService(db *gorm.DB, redis *redisClient.Client, configManager *config.ConfigManager) *EmailService {
	return &EmailService{
		db:            db,
		redis:         redis,
		configManager: configManager,
	}
}

// getEmailConfig 从ConfigManager获取邮件配置
func (s *EmailService) getEmailConfig() (*email.Config, error) {
	// 读取配置，如果不存在则使用默认值
	host := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPHost, models.ConfigValueEmailSMTPHostDefault)
	portStr := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPPort, models.ConfigValueEmailSMTPPortDefault)
	username := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPUsername, "")
	password := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPPassword, "")
	from := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPFrom, "")
	fromName := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPFromName, models.ConfigValueEmailSMTPFromNameDefault)
	useTLSStr := s.configManager.GetWithDefault(models.ConfigKeyEmailSMTPUseTLS, models.ConfigValueEmailSMTPUseTLSDefault)

	// 转换端口
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Error("邮件SMTP端口配置无效",
			logger.String("port", portStr),
			logger.Err(err))
		return nil, errors.WithMessage(errors.CodeConfigError, "SMTP端口配置无效", err)
	}

	// 转换TLS配置
	useTLS := useTLSStr == models.ConfigValueTrue

	// 验证必需配置
	if username == "" || password == "" || from == "" {
		logger.Error("邮件配置不完整，缺少必需的SMTP凭证")
		return nil, errors.New(errors.CodeConfigError, "邮件配置不完整")
	}

	return &email.Config{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		FromName: fromName,
		UseTLS:   useTLS,
	}, nil
}

// generateCode 生成6位随机数字验证码
func (s *EmailService) generateCode() (string, error) {
	// 生成000000-999999之间的随机数
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		logger.Error("生成验证码失败", logger.Err(err))
		return "", errors.Wrap(errors.CodeInternalError, err)
	}
	// 格式化为6位数字，不足补0
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// getCodeKey 获取验证码在Redis中的键名
func (s *EmailService) getCodeKey(purpose VerificationPurpose, emailAddr string) string {
	return fmt.Sprintf("email_code:%s:%s", purpose, emailAddr)
}

// getLimitKey 获取频率限制在Redis中的键名
func (s *EmailService) getLimitKey(emailAddr string) string {
	return fmt.Sprintf("email_limit:%s", emailAddr)
}

// SendVerificationCode 发送验证码
func (s *EmailService) SendVerificationCode(ctx context.Context, emailAddr string, purpose VerificationPurpose) error {
	// 1. 检查频率限制
	limitKey := s.getLimitKey(emailAddr)
	exists, err := s.redis.Exists(ctx, limitKey)
	if err != nil {
		logger.Error("检查发送频率限制失败",
			logger.String("email", emailAddr),
			logger.Err(err))
		return errors.Wrap(errors.CodeCacheError, err)
	}
	if exists > 0 {
		logger.Warn("发送验证码过于频繁",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)))
		return errors.New(errors.CodeTooManyRequests, "发送过于频繁，请稍后再试")
	}

	// 2. 生成验证码
	code, err := s.generateCode()
	if err != nil {
		return err
	}

	// 3. 获取邮件配置
	emailConfig, err := s.getEmailConfig()
	if err != nil {
		return err
	}

	// 4. 发送邮件
	sender := email.NewSender(emailConfig)
	if err := sender.SendVerificationCode(emailAddr, code, purpose.String()); err != nil {
		return err
	}

	// 5. 存储验证码到Redis（带过期时间）
	codeKey := s.getCodeKey(purpose, emailAddr)
	expiryStr := s.configManager.GetWithDefault(models.ConfigKeyEmailCodeExpiry, models.ConfigValueEmailCodeExpiryDefault)
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		logger.Error("验证码有效期配置无效",
			logger.String("expiry", expiryStr),
			logger.Err(err))
		expiry = 300 // 默认5分钟
	}

	if err := s.redis.Set(ctx, codeKey, code, time.Duration(expiry)*time.Second); err != nil {
		logger.Error("存储验证码到Redis失败",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)),
			logger.Err(err))
		return errors.Wrap(errors.CodeCacheError, err)
	}

	// 6. 设置频率限制
	rateLimitStr := s.configManager.GetWithDefault(models.ConfigKeyEmailRateLimit, models.ConfigValueEmailRateLimitDefault)
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		logger.Error("发送频率限制配置无效",
			logger.String("rate_limit", rateLimitStr),
			logger.Err(err))
		rateLimit = 60 // 默认60秒
	}

	if err := s.redis.Set(ctx, limitKey, "1", time.Duration(rateLimit)*time.Second); err != nil {
		logger.Error("设置发送频率限制失败",
			logger.String("email", emailAddr),
			logger.Err(err))
		// 不影响主流程，继续执行
	}

	logger.Info("验证码发送成功",
		logger.String("email", emailAddr),
		logger.String("purpose", string(purpose)),
		logger.Int("expiry_seconds", expiry))

	return nil
}

// VerifyCode 验证验证码
// 验证成功后会删除验证码（一次性使用）
func (s *EmailService) VerifyCode(ctx context.Context, emailAddr string, purpose VerificationPurpose, code string) error {
	codeKey := s.getCodeKey(purpose, emailAddr)

	// 1. 从Redis获取验证码
	storedCode, err := s.redis.Get(ctx, codeKey)
	if err == redis.Nil {
		logger.Warn("验证码不存在或已过期",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)))
		return errors.New(errors.CodeVerificationCodeExpired, "验证码已过期")
	}
	if err != nil {
		logger.Error("获取验证码失败",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)),
			logger.Err(err))
		return errors.Wrap(errors.CodeCacheError, err)
	}

	// 2. 验证验证码
	if storedCode != code {
		logger.Warn("验证码错误",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)))
		return errors.New(errors.CodeInvalidVerificationCode, "验证码错误")
	}

	// 3. 删除验证码（一次性使用）
	if err := s.redis.Del(ctx, codeKey); err != nil {
		logger.Error("删除验证码失败",
			logger.String("email", emailAddr),
			logger.String("purpose", string(purpose)),
			logger.Err(err))
		// 不影响验证结果
	}

	logger.Info("验证码验证成功",
		logger.String("email", emailAddr),
		logger.String("purpose", string(purpose)))

	return nil
}

// MarkEmailVerified 标记邮箱已验证
func (s *EmailService) MarkEmailVerified(ctx context.Context, emailAddr string) error {
	// 查找用户档案
	var profile models.UserProfile
	if err := s.db.Where("email = ?", emailAddr).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("邮箱对应的用户档案不存在",
				logger.String("email", emailAddr))
			return errors.New(errors.CodeResourceNotFound, "用户不存在")
		}
		logger.Error("查询用户档案失败",
			logger.String("email", emailAddr),
			logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 更新验证状态
	if err := s.db.Model(&profile).Update("email_verified", true).Error; err != nil {
		logger.Error("更新邮箱验证状态失败",
			logger.String("email", emailAddr),
			logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("邮箱验证状态更新成功",
		logger.String("email", emailAddr),
		logger.Uint("user_id", uint(profile.UserID)))

	return nil
}

// SendPasswordResetLink 发送密码重置链接
func (s *EmailService) SendPasswordResetLink(emailAddr string, resetURL string, expiryMinutes int) error {
	// 获取邮件配置
	emailConfig, err := s.getEmailConfig()
	if err != nil {
		return err
	}

	// 创建邮件发送器并发送
	sender := email.NewSender(emailConfig)
	return sender.SendPasswordResetLink(emailAddr, resetURL, expiryMinutes)
}
