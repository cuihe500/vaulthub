package service

import (
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/crypto"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EncryptionService 加密服务
type EncryptionService struct {
	db *gorm.DB
}

// NewEncryptionService 创建加密服务实例
func NewEncryptionService(db *gorm.DB) *EncryptionService {
	return &EncryptionService{
		db: db,
	}
}

// CreateUserEncryptionKeyRequest 创建用户加密密钥请求
// 注意：UserUUID 由服务端从认证上下文中提取，不需要客户端传入
type CreateUserEncryptionKeyRequest struct {
	UserUUID    string `json:"-"`                                  // 不从请求体解析，由handler从上下文设置
	SecurityPIN string `json:"security_pin" binding:"required,min=8"` // 安全密码，用于保护加密数据（独立于登录密码）
}

// CreateUserEncryptionKeyResponse 创建用户加密密钥响应
// 注意：恢复密钥仅在创建时返回一次，用户必须妥善保管
type CreateUserEncryptionKeyResponse struct {
	UserEncryptionKey *models.SafeUserEncryptionKey `json:"user_encryption_key"`
	RecoveryKey       string                        `json:"recovery_key,omitempty"` // 恢复密钥（阶段2实现）
}

// CreateUserEncryptionKey 创建用户加密密钥
// 在用户注册或首次使用加密功能时调用
func (s *EncryptionService) CreateUserEncryptionKey(req *CreateUserEncryptionKeyRequest) (*CreateUserEncryptionKeyResponse, error) {
	// 检查用户是否已存在加密密钥
	var existing models.UserEncryptionKey
	err := s.db.Where("user_uuid = ?", req.UserUUID).First(&existing).Error
	if err == nil {
		logger.Warn("用户加密密钥已存在", logger.String("user_uuid", req.UserUUID))
		return nil, errors.New(errors.CodeResourceAlreadyExists, "用户加密密钥已存在")
	}
	if err != gorm.ErrRecordNotFound {
		logger.Error("检查用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 1. 生成随机DEK（数据加密密钥，32字节）
	dek, err := crypto.GenerateRandomBytes(crypto.AESKeySize)
	if err != nil {
		logger.Error("生成DEK失败", logger.Err(err))
		return nil, err
	}
	defer crypto.ClearBytes(dek) // 使用完毕后清零内存

	// 2. 生成安全密码的bcrypt哈希（用于快速验证）
	securityPINHash, err := crypto.HashPassword(req.SecurityPIN)
	if err != nil {
		logger.Error("生成安全密码哈希失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeCryptoError, err)
	}

	// 3. 生成随机盐值用于派生KEK
	kekSalt, err := crypto.GenerateRandomBytes(crypto.SaltSize)
	if err != nil {
		logger.Error("生成KEK盐值失败", logger.Err(err))
		return nil, err
	}

	// 4. 从安全密码派生KEK（密钥加密密钥）
	kek, err := crypto.DeriveKEK(req.SecurityPIN, kekSalt)
	if err != nil {
		logger.Error("派生KEK失败", logger.Err(err))
		return nil, errors.WithMessage(errors.CodeKeyDerivationError, "密钥派生失败", err)
	}
	defer crypto.ClearBytes(kek) // 使用完毕后清零内存

	// 4. 用KEK加密DEK
	encryptedDEK, nonce1, authTag1, err := crypto.EncryptAESGCM(dek, kek)
	if err != nil {
		logger.Error("加密DEK失败", logger.Err(err))
		return nil, err
	}

	// 组装加密后的DEK：[密文][nonce(12)][tag(16)]
	// 注意：必须创建新的slice并预分配容量，避免底层数组共享导致的数据错误
	encryptedDEKBlob := make([]byte, 0, len(encryptedDEK)+len(nonce1)+len(authTag1))
	encryptedDEKBlob = append(encryptedDEKBlob, encryptedDEK...)
	encryptedDEKBlob = append(encryptedDEKBlob, nonce1...)
	encryptedDEKBlob = append(encryptedDEKBlob, authTag1...)

	// 5. 生成BIP39助记词（24个单词）
	recoveryMnemonic, err := crypto.GenerateBIP39Mnemonic()
	if err != nil {
		logger.Error("生成BIP39助记词失败", logger.Err(err))
		return nil, err
	}

	// 从助记词派生恢复密钥
	recoveryKey, err := crypto.DeriveKeyFromMnemonic(recoveryMnemonic)
	if err != nil {
		logger.Error("从助记词派生恢复密钥失败", logger.Err(err))
		return nil, err
	}
	defer crypto.ClearBytes(recoveryKey)

	// 计算恢复密钥哈希（用于后续验证）
	recoveryKeyHash := crypto.HashRecoveryKey(recoveryKey)

	// 6. 用恢复密钥加密DEK（备份）
	encryptedDEKRecovery, nonce2, authTag2, err := crypto.EncryptAESGCM(dek, recoveryKey)
	if err != nil {
		logger.Error("用恢复密钥加密DEK失败", logger.Err(err))
		return nil, err
	}

	// 注意：必须创建新的slice并预分配容量，避免底层数组共享导致的数据错误
	encryptedDEKRecoveryBlob := make([]byte, 0, len(encryptedDEKRecovery)+len(nonce2)+len(authTag2))
	encryptedDEKRecoveryBlob = append(encryptedDEKRecoveryBlob, encryptedDEKRecovery...)
	encryptedDEKRecoveryBlob = append(encryptedDEKRecoveryBlob, nonce2...)
	encryptedDEKRecoveryBlob = append(encryptedDEKRecoveryBlob, authTag2...)

	// 7. 存储到数据库
	userKey := models.UserEncryptionKey{
		UserUUID:             req.UserUUID,
		KEKSalt:              kekSalt,
		KEKAlgorithm:         "argon2id",
		EncryptedDEK:         encryptedDEKBlob,
		DEKVersion:           1,
		DEKAlgorithm:         "AES-256-GCM",
		SecurityPINHash:      securityPINHash, // 存储安全密码哈希
		RecoveryKeyHash:      recoveryKeyHash,
		EncryptedDEKRecovery: encryptedDEKRecoveryBlob,
	}

	if err := s.db.Create(&userKey).Error; err != nil {
		logger.Error("创建用户加密密钥失败", logger.Err(err), logger.String("user_uuid", req.UserUUID))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("创建用户加密密钥成功", logger.String("user_uuid", req.UserUUID))

	// 8. 返回响应（恢复助记词仅显示一次，用户必须妥善保管）
	return &CreateUserEncryptionKeyResponse{
		UserEncryptionKey: userKey.ToSafe(),
		RecoveryKey:       recoveryMnemonic, // 返回24个单词的助记词
	}, nil
}

// EncryptAndStoreSecretRequest 加密并存储秘密请求
// 注意：UserUUID 由服务端从认证上下文中提取，不需要客户端传入
type EncryptAndStoreSecretRequest struct {
	UserUUID    string                 `json:"-"`                           // 不从请求体解析，由handler从上下文设置
	SecurityPIN string                 `json:"security_pin" binding:"required"` // 安全密码，用于解密DEK
	SecretName  string                 `json:"secret_name" binding:"required"`
	SecretType  models.SecretType      `json:"secret_type" binding:"required"`
	PlainData   string                 `json:"plain_data" binding:"required"`
	Description string                 `json:"description"`
	Metadata    *models.SecretMetadata `json:"metadata"`
}

// EncryptAndStoreSecret 加密并存储秘密
func (s *EncryptionService) EncryptAndStoreSecret(req *EncryptAndStoreSecretRequest) (*models.SafeEncryptedSecret, error) {
	// 1. 获取用户的加密密钥配置
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("用户加密密钥不存在", logger.String("user_uuid", req.UserUUID))
			return nil, errors.New(errors.CodeResourceNotFound, "用户加密密钥不存在，请先创建")
		}
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 2. 验证安全密码（快速失败，避免昂贵的Argon2计算）
	if !crypto.VerifyPassword(req.SecurityPIN, userKey.SecurityPINHash) {
		logger.Warn("安全密码验证失败", logger.String("user_uuid", req.UserUUID))
		return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
	}

	// 3. 从安全密码派生KEK
	kek, err := crypto.DeriveKEK(req.SecurityPIN, userKey.KEKSalt)
	if err != nil {
		logger.Error("派生KEK失败", logger.Err(err))
		return nil, errors.WithMessage(errors.CodeKeyDerivationError, "密钥派生失败", err)
	}
	defer crypto.ClearBytes(kek)

	// 4. 解密DEK
	dek, err := s.decryptDEK(userKey.EncryptedDEK, kek)
	if err != nil {
		logger.Warn("解密DEK失败，安全密码可能错误", logger.String("user_uuid", req.UserUUID), logger.Err(err))
		return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
	}
	defer crypto.ClearBytes(dek)

	// 4. 用DEK加密实际数据
	encryptedData, dataNonce, dataAuthTag, err := crypto.EncryptAESGCM([]byte(req.PlainData), dek)
	if err != nil {
		logger.Error("加密秘密数据失败", logger.Err(err))
		return nil, err
	}

	// 5. 生成秘密UUID
	secretUUID := uuid.New().String()

	// 6. 存储到数据库
	secret := models.EncryptedSecret{
		UserUUID:      req.UserUUID,
		SecretUUID:    secretUUID,
		SecretName:    req.SecretName,
		SecretType:    req.SecretType,
		Description:   req.Description,
		EncryptedData: encryptedData,
		DEKVersion:    userKey.DEKVersion,
		Nonce:         dataNonce,
		AuthTag:       dataAuthTag,
		Metadata:      req.Metadata,
	}

	if err := s.db.Create(&secret).Error; err != nil {
		logger.Error("存储加密秘密失败", logger.Err(err), logger.String("user_uuid", req.UserUUID))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("加密并存储秘密成功",
		logger.String("user_uuid", req.UserUUID),
		logger.String("secret_uuid", secretUUID),
		logger.String("secret_type", string(req.SecretType)))

	return secret.ToSafe(), nil
}

// DecryptSecretRequest 解密秘密请求
// 注意：UserUUID 和 SecretUUID 由服务端从认证上下文和URL路径中提取，不需要客户端传入
type DecryptSecretRequest struct {
	UserUUID    string `json:"-"`                           // 不从请求体解析，由handler从上下文设置
	SecretUUID  string `json:"-"`                           // 不从请求体解析，由handler从URL路径设置
	SecurityPIN string `json:"security_pin" binding:"required"` // 安全密码，用于解密DEK
}

// DecryptSecret 解密秘密
func (s *EncryptionService) DecryptSecret(req *DecryptSecretRequest) (*models.DecryptedSecret, error) {
	// 1. 获取加密的秘密
	var secret models.EncryptedSecret
	if err := s.db.Where("user_uuid = ? AND secret_uuid = ?", req.UserUUID, req.SecretUUID).First(&secret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("秘密不存在或无权访问", logger.String("user_uuid", req.UserUUID), logger.String("secret_uuid", req.SecretUUID))
			return nil, errors.New(errors.CodeResourceNotFound, "秘密不存在或无权访问")
		}
		logger.Error("查询秘密失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 检查是否过期
	if secret.IsExpired() {
		logger.Warn("秘密已过期", logger.String("secret_uuid", req.SecretUUID))
		return nil, errors.New(errors.CodeResourceNotFound, "秘密已过期")
	}

	// 2. 获取用户密钥
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&userKey).Error; err != nil {
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 3. 验证DEK版本匹配
	if secret.DEKVersion != userKey.DEKVersion {
		logger.Warn("DEK版本不匹配", logger.Int("secret_version", secret.DEKVersion), logger.Int("current_version", userKey.DEKVersion))
		return nil, errors.New(errors.CodeCryptoError, "密钥版本不匹配，请联系管理员")
	}

	// 4. 验证安全密码（快速失败，避免昂贵的Argon2计算）
	if !crypto.VerifyPassword(req.SecurityPIN, userKey.SecurityPINHash) {
		logger.Warn("安全密码验证失败", logger.String("user_uuid", req.UserUUID))
		return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
	}

	// 5. 从安全密码派生KEK
	kek, err := crypto.DeriveKEK(req.SecurityPIN, userKey.KEKSalt)
	if err != nil {
		logger.Error("派生KEK失败", logger.Err(err))
		return nil, errors.WithMessage(errors.CodeKeyDerivationError, "密钥派生失败", err)
	}
	defer crypto.ClearBytes(kek)

	// 6. 解密DEK
	dek, err := s.decryptDEK(userKey.EncryptedDEK, kek)
	if err != nil {
		logger.Warn("解密DEK失败，安全密码可能错误", logger.String("user_uuid", req.UserUUID), logger.Err(err))
		return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
	}
	defer crypto.ClearBytes(dek)

	// 6. 用DEK解密数据
	plainData, err := crypto.DecryptAESGCM(secret.EncryptedData, dek, secret.Nonce, secret.AuthTag)
	if err != nil {
		logger.Error("解密秘密数据失败", logger.Err(err), logger.String("secret_uuid", req.SecretUUID))
		return nil, errors.WithMessage(errors.CodeDecryptionFailed, "解密失败或数据被篡改", err)
	}

	// 7. 更新访问统计（异步，不影响主流程）
	go func() {
		if err := s.db.Model(&models.EncryptedSecret{}).
			Where("id = ?", secret.ID).
			Updates(map[string]interface{}{
				"last_accessed_at": gorm.Expr("NOW()"),
				"access_count":     gorm.Expr("access_count + ?", 1),
			}).Error; err != nil {
			logger.Error("更新秘密访问统计失败", logger.Err(err), logger.Uint("secret_id", secret.ID))
		}
	}()

	logger.Info("解密秘密成功", logger.String("user_uuid", req.UserUUID), logger.String("secret_uuid", req.SecretUUID))

	// 8. 返回解密后的数据
	return &models.DecryptedSecret{
		SafeEncryptedSecret: *secret.ToSafe(),
		PlainData:           string(plainData),
	}, nil
}

// DeleteSecret 删除秘密（软删除）
func (s *EncryptionService) DeleteSecret(userUUID, secretUUID string) error {
	result := s.db.Where("user_uuid = ? AND secret_uuid = ?", userUUID, secretUUID).Delete(&models.EncryptedSecret{})
	if result.Error != nil {
		logger.Error("删除秘密失败", logger.Err(result.Error))
		return errors.Wrap(errors.CodeDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		logger.Warn("秘密不存在或无权删除", logger.String("user_uuid", userUUID), logger.String("secret_uuid", secretUUID))
		return errors.New(errors.CodeResourceNotFound, "秘密不存在或无权删除")
	}

	logger.Info("删除秘密成功", logger.String("user_uuid", userUUID), logger.String("secret_uuid", secretUUID))
	return nil
}

// ListUserSecretsRequest 列出用户秘密请求
type ListUserSecretsRequest struct {
	UserUUID   string            `form:"-"`
	SecretType models.SecretType `form:"secret_type"`
	Page       int               `form:"page" binding:"omitempty,min=1"`
	PageSize   int               `form:"page_size" binding:"omitempty,min=1,max=10000"`
}

// ListUserSecretsResponse 列出用户秘密响应
type ListUserSecretsResponse struct {
	Secrets    []*models.SafeEncryptedSecret `json:"secrets"`
	Total      int64                         `json:"total"`
	Page       int                           `json:"page"`
	PageSize   int                           `json:"page_size"`
	TotalPages int                           `json:"total_pages"`
}

// ListUserSecrets 列出用户的秘密列表（不包含加密数据）
func (s *EncryptionService) ListUserSecrets(req *ListUserSecretsRequest) (*ListUserSecretsResponse, error) {
	// 构建查询
	query := s.db.Model(&models.EncryptedSecret{}).Where("user_uuid = ?", req.UserUUID)

	// 应用过滤条件
	if req.SecretType != "" {
		query = query.Where("secret_type = ?", req.SecretType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取秘密总数失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 分页查询
	var secrets []models.EncryptedSecret
	query = query.Order("created_at DESC")

	// 如果未传分页参数，则全量导出（添加安全上限10000）
	if req.Page <= 0 && req.PageSize <= 0 {
		query = query.Limit(10000)
	} else {
		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	if err := query.Find(&secrets).Error; err != nil {
		logger.Error("查询秘密列表失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 转换为安全信息
	safeSecrets := make([]*models.SafeEncryptedSecret, len(secrets))
	for i, secret := range secrets {
		safeSecrets[i] = secret.ToSafe()
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListUserSecretsResponse{
		Secrets:    safeSecrets,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// decryptDEK 解密DEK的辅助函数
// blob格式: [密文][nonce(12)][tag(16)]
func (s *EncryptionService) decryptDEK(blob, kek []byte) ([]byte, error) {
	if len(blob) < 28 {
		return nil, errors.New(errors.CodeCryptoError, "无效的加密密钥数据")
	}

	encryptedDEK := blob[:len(blob)-28]
	nonce := blob[len(blob)-28 : len(blob)-16]
	authTag := blob[len(blob)-16:]

	return crypto.DecryptAESGCM(encryptedDEK, kek, nonce, authTag)
}
