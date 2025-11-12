package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatType 统计类型
type StatType string

const (
	StatDaily   StatType = "daily"
	StatWeekly  StatType = "weekly"
	StatMonthly StatType = "monthly"
)

// UserStatistics 用户统计模型
type UserStatistics struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UUID      string         `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	CreatedAt time.Time      `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 统计维度
	UserUUID string    `gorm:"type:char(36);not null;uniqueIndex:idx_user_date_type" json:"user_uuid"`
	StatDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_date_type;index:idx_stat_date" json:"stat_date"`
	StatType StatType  `gorm:"type:varchar(16);not null;uniqueIndex:idx_user_date_type;index:idx_stat_type" json:"stat_type"`

	// 密钥数量统计
	TotalSecrets     int `gorm:"not null;default:0" json:"total_secrets"`
	APIKeyCount      int `gorm:"not null;default:0" json:"api_key_count"`
	PasswordCount    int `gorm:"not null;default:0" json:"password_count"`
	CertificateCount int `gorm:"not null;default:0" json:"certificate_count"`
	SSHKeyCount      int `gorm:"not null;default:0" json:"ssh_key_count"`
	PrivateKeyCount  int `gorm:"not null;default:0" json:"private_key_count"`
	OtherCount       int `gorm:"not null;default:0" json:"other_count"`

	// 操作次数统计
	CreateCount     int `gorm:"not null;default:0" json:"create_count"`
	UpdateCount     int `gorm:"not null;default:0" json:"update_count"`
	DeleteCount     int `gorm:"not null;default:0" json:"delete_count"`
	AccessCount     int `gorm:"not null;default:0" json:"access_count"`
	TotalOperations int `gorm:"not null;default:0" json:"total_operations"`

	// 登录统计
	LoginCount       int `gorm:"not null;default:0" json:"login_count"`
	FailedLoginCount int `gorm:"not null;default:0" json:"failed_login_count"`
}

// TableName 指定表名
func (UserStatistics) TableName() string {
	return "user_statistics"
}

// BeforeCreate GORM钩子：创建前自动生成UUID
func (u *UserStatistics) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}
