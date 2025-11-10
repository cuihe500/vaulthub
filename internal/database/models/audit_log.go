package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActionType 操作类型
type ActionType string

const (
	ActionCreate ActionType = "CREATE"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionAccess ActionType = "ACCESS"
	ActionLogin  ActionType = "LOGIN"
	ActionLogout ActionType = "LOGOUT"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceVault  ResourceType = "vault"
	ResourceSecret ResourceType = "secret"
	ResourceUser   ResourceType = "user"
	ResourceConfig ResourceType = "config"
)

// AuditStatus 审计状态
type AuditStatus string

const (
	AuditSuccess AuditStatus = "success"
	AuditFailed  AuditStatus = "failed"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UUID      string    `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time `gorm:"type:datetime;not null" json:"created_at"`

	// 操作主体
	UserUUID string `gorm:"type:char(36);not null;index:idx_user_uuid" json:"user_uuid"`
	Username string `gorm:"type:varchar(64);not null;index:idx_username" json:"username"`

	// 操作内容
	ActionType   ActionType   `gorm:"type:varchar(32);not null;index:idx_action_type" json:"action_type"`
	ResourceType ResourceType `gorm:"type:varchar(32);not null;index:idx_resource_type" json:"resource_type"`
	ResourceUUID *string      `gorm:"type:char(36);index:idx_resource_uuid" json:"resource_uuid,omitempty"`
	ResourceName *string      `gorm:"type:varchar(255)" json:"resource_name,omitempty"`

	// 操作结果
	Status       AuditStatus `gorm:"type:varchar(16);not null;index:idx_status" json:"status"`
	ErrorCode    *int        `gorm:"type:int" json:"error_code,omitempty"`
	ErrorMessage *string     `gorm:"type:text" json:"error_message,omitempty"`

	// 上下文信息
	IPAddress *string `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent *string `gorm:"type:varchar(512)" json:"user_agent,omitempty"`
	RequestID *string `gorm:"type:varchar(64)" json:"request_id,omitempty"`

	// 额外数据
	Details interface{} `gorm:"type:json" json:"details,omitempty"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate GORM钩子：创建前自动生成UUID
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.UUID == "" {
		a.UUID = uuid.New().String()
	}
	return nil
}
