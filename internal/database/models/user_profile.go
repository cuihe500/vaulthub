package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/cuihe500/vaulthub/pkg/errors"
	"gorm.io/gorm"
)

// UserProfile 用户基本信息模型
type UserProfile struct {
	BaseModel
	UserID        uint   `gorm:"uniqueIndex;not null;comment:关联用户ID" json:"user_id"`                        // 外键关联 users 表
	Nickname      string `gorm:"type:varchar(50);not null;comment:用户昵称" json:"nickname"`                      // 昵称
	Phone         string `gorm:"type:varchar(20);comment:手机号" json:"phone"`                                   // 手机号（可选）
	Email         string `gorm:"type:varchar(100);not null;comment:邮箱地址" json:"email"`                        // 邮箱
	EmailVerified bool   `gorm:"type:tinyint(1);not null;default:0;comment:邮箱是否已验证" json:"email_verified"` // 邮箱验证状态
	User          User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// BeforeCreate GORM 钩子：创建前验证
func (p *UserProfile) BeforeCreate(tx *gorm.DB) error {
	return p.validate()
}

// BeforeUpdate GORM 钩子：更新前验证
func (p *UserProfile) BeforeUpdate(tx *gorm.DB) error {
	return p.validate()
}

// validate 验证用户基本信息
func (p *UserProfile) validate() error {
	// 昵称不能为空
	if strings.TrimSpace(p.Nickname) == "" {
		return errors.New(errors.CodeNicknameRequired, "")
	}

	// 邮箱不能为空
	if strings.TrimSpace(p.Email) == "" {
		return errors.New(errors.CodeEmailRequired, "")
	}

	// 邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(p.Email) {
		return errors.New(errors.CodeInvalidEmail, "")
	}

	// 手机号格式验证（可选）
	if p.Phone != "" && !isValidPhone(p.Phone) {
		return errors.New(errors.CodeInvalidPhone, "")
	}

	return nil
}

// isValidPhone 验证手机号格式
func isValidPhone(phone string) bool {
	// 简单的中国手机号验证
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// ToSafeProfile 转换为安全的用户档案信息
func (p *UserProfile) ToSafeProfile() *SafeUserProfile {
	return &SafeUserProfile{
		ID:            p.ID,
		UserID:        p.UserID,
		Nickname:      p.Nickname,
		Phone:         p.Phone,
		Email:         p.Email,
		EmailVerified: p.EmailVerified,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// SafeUserProfile 安全的用户档案信息（用于API返回）
type SafeUserProfile struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	Nickname      string    `json:"nickname"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}