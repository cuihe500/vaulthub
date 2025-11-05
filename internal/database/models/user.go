package models

import (
	"time"
)

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusActive   UserStatus = 1 // 活跃
	UserStatusDisabled UserStatus = 2 // 已禁用
	UserStatusLocked   UserStatus = 3 // 已锁定
)

// User 用户模型
type User struct {
	BaseModel
	UUID         string     `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	Username     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"` // 密码哈希不返回给前端
	Status       UserStatus `gorm:"type:tinyint;not null;default:1" json:"status"`
	Role         string     `gorm:"type:varchar(32);not null;default:'user';index" json:"role"`
	LastLoginAt  *time.Time `gorm:"type:datetime" json:"last_login_at,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsActive 判断用户是否为活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsDisabled 判断用户是否被禁用
func (u *User) IsDisabled() bool {
	return u.Status == UserStatusDisabled
}

// IsLocked 判断用户是否被锁定
func (u *User) IsLocked() bool {
	return u.Status == UserStatusLocked
}

// CanOperate 判断用户是否可以操作（只有活跃用户可以操作）
func (u *User) CanOperate() bool {
	return u.IsActive()
}

// IsAdmin 判断用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// SafeUser 用于返回给前端的安全用户信息（不包含敏感字段）
type SafeUser struct {
	ID          uint       `json:"id"`
	UUID        string     `json:"uuid"`
	Username    string     `json:"username"`
	Status      UserStatus `json:"status"`
	Role        string     `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToSafeUser 转换为安全用户信息
func (u *User) ToSafeUser() *SafeUser {
	return &SafeUser{
		ID:          u.ID,
		UUID:        u.UUID,
		Username:    u.Username,
		Status:      u.Status,
		Role:        u.Role,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
