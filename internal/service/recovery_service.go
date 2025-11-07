package service

import (
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/crypto"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// RecoveryService 密钥恢复服务
type RecoveryService struct {
	db *gorm.DB
}

// NewRecoveryService 创建密钥恢复服务实例
func NewRecoveryService(db *gorm.DB) *RecoveryService {
	return &RecoveryService{
		db: db,
	}
}

// VerifyRecoveryKeyRequest 验证恢复密钥请求
// 用于验证用户输入的恢复助记词是否正确，不执行实际的密码重置
type VerifyRecoveryKeyRequest struct {
	UserUUID        string `json:"-"` // 不从请求体解析，由handler从上下文设置
	RecoveryMnemonic string `json:"recovery_mnemonic" binding:"required"`
}

// VerifyRecoveryKeyResponse 验证恢复密钥响应
type VerifyRecoveryKeyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// VerifyRecoveryKey 验证恢复密钥的有效性
// 只验证助记词格式和哈希是否匹配，不执行密码重置操作
func (s *RecoveryService) VerifyRecoveryKey(req *VerifyRecoveryKeyRequest) (*VerifyRecoveryKeyResponse, error) {
	// 1. 验证助记词格式
	if !crypto.IsMnemonicValid(req.RecoveryMnemonic) {
		logger.Warn("助记词格式无效", logger.String("user_uuid", req.UserUUID))
		return &VerifyRecoveryKeyResponse{
			Valid:   false,
			Message: "助记词格式无效，请检查是否输入错误",
		}, nil
	}

	// 2. 获取用户密钥
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("用户加密密钥不存在", logger.String("user_uuid", req.UserUUID))
			return nil, errors.New(errors.CodeResourceNotFound, "用户加密密钥不存在")
		}
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 3. 从助记词派生恢复密钥
	recoveryKey, err := crypto.DeriveKeyFromMnemonic(req.RecoveryMnemonic)
	if err != nil {
		logger.Error("从助记词派生恢复密钥失败", logger.Err(err))
		return nil, errors.WithMessage(errors.CodeKeyDerivationError, "恢复密钥派生失败", err)
	}
	defer crypto.ClearBytes(recoveryKey)

	// 4. 验证恢复密钥哈希
	recoveryKeyHash := crypto.HashRecoveryKey(recoveryKey)
	if recoveryKeyHash != userKey.RecoveryKeyHash {
		logger.Warn("恢复密钥哈希不匹配", logger.String("user_uuid", req.UserUUID))
		return &VerifyRecoveryKeyResponse{
			Valid:   false,
			Message: "恢复密钥错误，请检查助记词是否正确",
		}, nil
	}

	logger.Info("恢复密钥验证成功", logger.String("user_uuid", req.UserUUID))
	return &VerifyRecoveryKeyResponse{
		Valid:   true,
		Message: "恢复密钥验证成功",
	}, nil
}

// ResetPasswordWithRecoveryRequest 使用恢复密钥重置安全密码请求
type ResetPasswordWithRecoveryRequest struct {
	UserUUID         string `json:"-"`                                  // 不从请求体解析，由handler从上下文设置
	RecoveryMnemonic string `json:"recovery_mnemonic" binding:"required"` // 恢复助记词
	NewSecurityPIN   string `json:"new_security_pin" binding:"required,min=8"` // 新的安全密码（独立于登录密码）
}

// ResetPasswordWithRecovery 使用恢复密钥重置安全密码
// 重要：安全密码重置后，所有已加密的数据不需要重新加密，因为DEK没有变化，只是KEK变了
func (s *RecoveryService) ResetPasswordWithRecovery(req *ResetPasswordWithRecoveryRequest) error {
	// 1. 验证助记词格式
	if !crypto.IsMnemonicValid(req.RecoveryMnemonic) {
		logger.Warn("助记词格式无效", logger.String("user_uuid", req.UserUUID))
		return errors.New(errors.CodeInvalidFormat, "助记词格式无效")
	}

	// 2. 获取用户密钥
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("用户加密密钥不存在", logger.String("user_uuid", req.UserUUID))
			return errors.New(errors.CodeResourceNotFound, "用户加密密钥不存在")
		}
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 3. 从助记词派生恢复密钥
	recoveryKey, err := crypto.DeriveKeyFromMnemonic(req.RecoveryMnemonic)
	if err != nil {
		logger.Error("从助记词派生恢复密钥失败", logger.Err(err))
		return errors.WithMessage(errors.CodeKeyDerivationError, "恢复密钥派生失败", err)
	}
	defer crypto.ClearBytes(recoveryKey)

	// 4. 验证恢复密钥正确性
	recoveryKeyHash := crypto.HashRecoveryKey(recoveryKey)
	if recoveryKeyHash != userKey.RecoveryKeyHash {
		logger.Warn("恢复密钥错误", logger.String("user_uuid", req.UserUUID))
		return errors.New(errors.CodeInvalidCredentials, "恢复密钥错误")
	}

	// 5. 用恢复密钥解密DEK
	dek, err := s.decryptDEKWithRecovery(userKey.EncryptedDEKRecovery, recoveryKey)
	if err != nil {
		logger.Error("用恢复密钥解密DEK失败", logger.Err(err), logger.String("user_uuid", req.UserUUID))
		return errors.WithMessage(errors.CodeDecryptionFailed, "恢复密钥解密DEK失败", err)
	}
	defer crypto.ClearBytes(dek)

	// 6. 生成新安全密码的哈希
	newSecurityPINHash, err := crypto.HashPassword(req.NewSecurityPIN)
	if err != nil {
		logger.Error("生成新安全密码哈希失败", logger.Err(err))
		return errors.Wrap(errors.CodeCryptoError, err)
	}

	// 7. 从新安全密码派生新KEK
	newKEKSalt, err := crypto.GenerateRandomBytes(crypto.SaltSize)
	if err != nil {
		logger.Error("生成新KEK盐值失败", logger.Err(err))
		return err
	}

	newKEK, err := crypto.DeriveKEK(req.NewSecurityPIN, newKEKSalt)
	if err != nil {
		logger.Error("派生新KEK失败", logger.Err(err))
		return errors.WithMessage(errors.CodeKeyDerivationError, "新密钥派生失败", err)
	}
	defer crypto.ClearBytes(newKEK)

	// 7. 用新KEK重新加密DEK
	newEncryptedDEK, newNonce, newAuthTag, err := crypto.EncryptAESGCM(dek, newKEK)
	if err != nil {
		logger.Error("用新KEK加密DEK失败", logger.Err(err))
		return err
	}

	// 组装新的加密后的DEK：[密文][nonce(12)][tag(16)]
	newEncryptedDEKBlob := make([]byte, 0, len(newEncryptedDEK)+len(newNonce)+len(newAuthTag))
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, newEncryptedDEK...)
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, newNonce...)
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, newAuthTag...)

	// 9. 更新数据库（在事务中）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新密钥派生参数和安全密码哈希
		if err := tx.Model(&userKey).Updates(map[string]interface{}{
			"kek_salt":          newKEKSalt,
			"encrypted_dek":     newEncryptedDEKBlob,
			"security_pin_hash": newSecurityPINHash, // 更新安全密码哈希
		}).Error; err != nil {
			return err
		}

		// 注意：所有encrypted_secrets表中的数据不需要重新加密
		// 因为DEK没有变化，只是KEK变了

		return nil
	})

	if err != nil {
		logger.Error("更新用户加密密钥失败", logger.Err(err), logger.String("user_uuid", req.UserUUID))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("使用恢复密钥重置密码成功", logger.String("user_uuid", req.UserUUID))
	return nil
}

// decryptDEKWithRecovery 用恢复密钥解密DEK的辅助函数
// blob格式: [密文][nonce(12)][tag(16)]
func (s *RecoveryService) decryptDEKWithRecovery(blob, recoveryKey []byte) ([]byte, error) {
	if len(blob) < 28 {
		return nil, errors.New(errors.CodeCryptoError, "无效的加密密钥数据")
	}

	encryptedDEK := blob[:len(blob)-28]
	nonce := blob[len(blob)-28 : len(blob)-16]
	authTag := blob[len(blob)-16:]

	return crypto.DecryptAESGCM(encryptedDEK, recoveryKey, nonce, authTag)
}
