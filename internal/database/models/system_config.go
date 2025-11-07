package models

import (
	"time"
)

// SystemConfig 系统配置模型
// 用于存储系统级别的配置项，如初始化标志等
type SystemConfig struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;type:varchar(64);uniqueIndex;not null" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:text;not null" json:"config_value"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_config"
}

// 预定义的配置键
const (
	ConfigKeyAdminInitialized = "admin_initialized" // 超级管理员是否已初始化

	// 密钥轮换相关配置
	ConfigKeyKeyRotationBatchSize    = "key_rotation_batch_size"     // 密钥轮换批次大小
	ConfigKeyKeyRotationBatchSleepMS = "key_rotation_batch_sleep_ms" // 密钥轮换批次间休眠时间(毫秒)
)

// 配置值
const (
	ConfigValueTrue  = "true"
	ConfigValueFalse = "false"

	// 密钥轮换默认配置值
	ConfigValueKeyRotationBatchSizeDefault    = "100" // 默认每批处理100条
	ConfigValueKeyRotationBatchSleepMSDefault = "100" // 默认批次间休眠100ms
)
