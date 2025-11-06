package models

import "time"

// UserEncryptionKey 用户加密密钥模型
// 存储用户的加密密钥配置和加密后的DEK
type UserEncryptionKey struct {
	BaseModel
	UserUUID string `gorm:"type:char(36);uniqueIndex;not null" json:"user_uuid"`

	// KEK 派生参数
	KEKSalt      []byte `gorm:"type:binary(32);not null" json:"-"` // 盐值不对外暴露
	KEKAlgorithm string `gorm:"type:varchar(32);not null;default:'argon2id'" json:"kek_algorithm"`

	// DEK 存储（被KEK加密）
	EncryptedDEK []byte `gorm:"type:varbinary(512);not null" json:"-"` // 加密密钥不对外暴露
	DEKVersion   int    `gorm:"type:int;not null;default:1" json:"dek_version"`
	DEKAlgorithm string `gorm:"type:varchar(32);not null;default:'AES-256-GCM'" json:"dek_algorithm"`

	// 恢复密钥
	RecoveryKeyHash       string `gorm:"type:char(64);not null" json:"-"`          // 恢复密钥哈希不对外暴露
	EncryptedDEKRecovery  []byte `gorm:"type:varbinary(512);not null" json:"-"`    // 恢复密钥加密的DEK不对外暴露
	LastRotationAt        *time.Time `gorm:"type:datetime" json:"last_rotation_at"` // 最后一次密钥轮换时间
}

// TableName 指定表名
func (UserEncryptionKey) TableName() string {
	return "user_encryption_keys"
}

// SafeUserEncryptionKey 用于返回给前端的安全信息（不包含敏感字段）
type SafeUserEncryptionKey struct {
	ID             uint       `json:"id"`
	UserUUID       string     `json:"user_uuid"`
	KEKAlgorithm   string     `json:"kek_algorithm"`
	DEKVersion     int        `json:"dek_version"`
	DEKAlgorithm   string     `json:"dek_algorithm"`
	LastRotationAt *time.Time `json:"last_rotation_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ToSafe 转换为安全信息
func (k *UserEncryptionKey) ToSafe() *SafeUserEncryptionKey {
	return &SafeUserEncryptionKey{
		ID:             k.ID,
		UserUUID:       k.UserUUID,
		KEKAlgorithm:   k.KEKAlgorithm,
		DEKVersion:     k.DEKVersion,
		DEKAlgorithm:   k.DEKAlgorithm,
		LastRotationAt: k.LastRotationAt,
		CreatedAt:      k.CreatedAt,
		UpdatedAt:      k.UpdatedAt,
	}
}
