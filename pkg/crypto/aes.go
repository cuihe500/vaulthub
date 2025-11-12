package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/cuihe500/vaulthub/pkg/errors"
)

const (
	// AESKeySize AES-256密钥长度（32字节=256位）
	AESKeySize = 32
	// GCMNonceSize GCM模式Nonce长度（12字节=96位）
	GCMNonceSize = 12
	// GCMTagSize GCM模式认证标签长度（16字节=128位）
	GCMTagSize = 16
)

// EncryptAESGCM 使用AES-256-GCM加密数据
// AES-GCM是一种认证加密（AEAD）模式，同时提供机密性和完整性保护
// 参数:
//   - plaintext: 待加密的明文数据
//   - key: 加密密钥（必须是32字节）
//
// 返回:
//   - ciphertext: 密文数据
//   - nonce: 随机生成的Nonce（12字节，每次加密必须唯一）
//   - authTag: 认证标签（16字节，用于验证数据完整性）
//   - error: 错误信息
func EncryptAESGCM(plaintext, key []byte) (ciphertext, nonce, authTag []byte, err error) {
	// 验证密钥长度
	if len(key) != AESKeySize {
		return nil, nil, nil, errors.New(errors.CodeInvalidParam, "密钥长度必须是32字节")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, errors.WithMessage(errors.CodeEncryptionFailed, "创建AES cipher失败", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, errors.WithMessage(errors.CodeEncryptionFailed, "创建GCM模式失败", err)
	}

	// 生成随机Nonce（12字节）
	// Nonce必须在每次加密时唯一，使用加密安全的随机数生成器
	nonce = make([]byte, GCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, errors.WithMessage(errors.CodeEncryptionFailed, "生成随机Nonce失败", err)
	}

	// 加密数据并附加认证标签
	// GCM.Seal会在密文末尾附加16字节的认证标签
	sealed := gcm.Seal(nil, nonce, plaintext, nil)

	// 分离密文和认证标签
	// sealed格式: [密文][认证标签(16字节)]
	if len(sealed) < GCMTagSize {
		return nil, nil, nil, errors.New(errors.CodeEncryptionFailed, "加密结果长度异常")
	}

	ciphertext = sealed[:len(sealed)-GCMTagSize]
	authTag = sealed[len(sealed)-GCMTagSize:]

	return ciphertext, nonce, authTag, nil
}

// DecryptAESGCM 使用AES-256-GCM解密数据
// 参数:
//   - ciphertext: 密文数据
//   - key: 解密密钥（必须是32字节）
//   - nonce: 加密时使用的Nonce（12字节）
//   - authTag: 认证标签（16字节）
//
// 返回:
//   - []byte: 解密后的明文数据
//   - error: 错误信息（如果认证失败，说明数据被篡改）
func DecryptAESGCM(ciphertext, key, nonce, authTag []byte) ([]byte, error) {
	// 验证参数长度
	if len(key) != AESKeySize {
		return nil, errors.New(errors.CodeInvalidParam, "密钥长度必须是32字节")
	}
	if len(nonce) != GCMNonceSize {
		return nil, errors.New(errors.CodeInvalidParam, "Nonce长度必须是12字节")
	}
	if len(authTag) != GCMTagSize {
		return nil, errors.New(errors.CodeInvalidParam, "认证标签长度必须是16字节")
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithMessage(errors.CodeDecryptionFailed, "创建AES cipher失败", err)
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithMessage(errors.CodeDecryptionFailed, "创建GCM模式失败", err)
	}

	// 重新组合密文和认证标签
	// GCM.Open期望的格式: [密文][认证标签(16字节)]
	// 注意：必须创建新的slice并预分配容量，避免底层数组共享导致的数据错误
	sealed := make([]byte, 0, len(ciphertext)+len(authTag))
	sealed = append(sealed, ciphertext...)
	sealed = append(sealed, authTag...)

	// 解密并验证数据完整性
	// 如果认证标签验证失败，说明数据被篡改，会返回错误
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, errors.WithMessage(errors.CodeDecryptionFailed, "解密失败或数据被篡改", err)
	}

	return plaintext, nil
}
