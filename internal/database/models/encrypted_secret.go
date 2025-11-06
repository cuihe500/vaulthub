package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SecretType 秘密类型枚举
type SecretType string

const (
	SecretTypeAPIKey        SecretType = "api_key"        // API密钥
	SecretTypeDBCredential  SecretType = "db_credential"  // 数据库凭证
	SecretTypeCertificate   SecretType = "certificate"    // 证书
	SecretTypeSSHKey        SecretType = "ssh_key"        // SSH密钥
	SecretTypeToken         SecretType = "token"          // 令牌
	SecretTypePassword      SecretType = "password"       // 密码
	SecretTypeOther         SecretType = "other"          // 其他
)

// SecretMetadata 秘密元数据（存储为JSON）
type SecretMetadata struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // 过期时间
	Tags      []string   `json:"tags,omitempty"`       // 标签
	Extra     map[string]interface{} `json:"extra,omitempty"` // 额外信息
}

// Scan 实现sql.Scanner接口，用于从数据库读取
func (m *SecretMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现driver.Valuer接口，用于写入数据库
func (m SecretMetadata) Value() (driver.Value, error) {
	if m.ExpiresAt == nil && len(m.Tags) == 0 && len(m.Extra) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// EncryptedSecret 加密秘密模型
// 存储用户加密后的敏感数据
type EncryptedSecret struct {
	BaseModel
	UserUUID   string     `gorm:"type:char(36);not null;index" json:"user_uuid"`
	SecretUUID string     `gorm:"type:char(36);uniqueIndex;not null" json:"secret_uuid"`

	// 业务信息
	SecretName  string     `gorm:"type:varchar(255);not null" json:"secret_name"`
	SecretType  SecretType `gorm:"type:varchar(32);not null;index" json:"secret_type"`
	Description string     `gorm:"type:text" json:"description,omitempty"`

	// 加密数据（不对外暴露原始加密数据）
	EncryptedData []byte `gorm:"type:blob;not null" json:"-"`
	DEKVersion    int    `gorm:"type:int;not null" json:"dek_version"`
	Nonce         []byte `gorm:"type:binary(12);not null" json:"-"`
	AuthTag       []byte `gorm:"type:binary(16);not null" json:"-"`

	// 元数据
	Metadata *SecretMetadata `gorm:"type:json" json:"metadata,omitempty"`

	// 审计
	LastAccessedAt *time.Time `gorm:"type:datetime" json:"last_accessed_at,omitempty"`
	AccessCount    int64      `gorm:"default:0" json:"access_count"`
}

// TableName 指定表名
func (EncryptedSecret) TableName() string {
	return "encrypted_secrets"
}

// IsExpired 判断秘密是否过期
func (s *EncryptedSecret) IsExpired() bool {
	if s.Metadata == nil || s.Metadata.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.Metadata.ExpiresAt)
}

// SafeEncryptedSecret 用于返回给前端的安全信息（不包含加密数据）
type SafeEncryptedSecret struct {
	ID             uint            `json:"id"`
	UserUUID       string          `json:"user_uuid"`
	SecretUUID     string          `json:"secret_uuid"`
	SecretName     string          `json:"secret_name"`
	SecretType     SecretType      `json:"secret_type"`
	Description    string          `json:"description,omitempty"`
	DEKVersion     int             `json:"dek_version"`
	Metadata       *SecretMetadata `json:"metadata,omitempty"`
	LastAccessedAt *time.Time      `json:"last_accessed_at,omitempty"`
	AccessCount    int64           `json:"access_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ToSafe 转换为安全信息
func (s *EncryptedSecret) ToSafe() *SafeEncryptedSecret {
	return &SafeEncryptedSecret{
		ID:             s.ID,
		UserUUID:       s.UserUUID,
		SecretUUID:     s.SecretUUID,
		SecretName:     s.SecretName,
		SecretType:     s.SecretType,
		Description:    s.Description,
		DEKVersion:     s.DEKVersion,
		Metadata:       s.Metadata,
		LastAccessedAt: s.LastAccessedAt,
		AccessCount:    s.AccessCount,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

// DecryptedSecret 解密后的秘密（包含明文数据，仅用于API响应）
type DecryptedSecret struct {
	SafeEncryptedSecret
	PlainData string `json:"plain_data"` // 解密后的明文数据
}
