package models

import (
	"time"
)

// PasswordResetToken 密码重置token模型
type PasswordResetToken struct {
	BaseModel
	UUID      string     `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`       // 对外唯一标识符
	UserID    uint       `gorm:"not null;index" json:"user_id"`                        // 关联用户ID
	TokenHash string     `gorm:"type:varchar(255);not null;index" json:"-"`            // Token哈希值（不返回给前端）
	ExpiresAt time.Time  `gorm:"type:datetime;not null;index" json:"expires_at"`       // 过期时间（UTC）
	UsedAt    *time.Time `gorm:"type:datetime" json:"used_at,omitempty"`               // 使用时间（NULL表示未使用）
	User      User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired 检查token是否已过期
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsUsed 检查token是否已使用
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// CanUse 检查token是否可用（未过期且未使用）
func (t *PasswordResetToken) CanUse() bool {
	return !t.IsExpired() && !t.IsUsed()
}
