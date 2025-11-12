package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// EntropyBits 助记词熵位数，256位对应24个单词
	EntropyBits = 256

	// RecoverySalt 恢复密钥派生时使用的盐值
	RecoverySalt = "vaulthub-recovery"

	// RecoveryIterations PBKDF2迭代次数
	RecoveryIterations = 100000

	// RecoveryKeyLength 派生密钥长度（字节）
	RecoveryKeyLength = 32
)

// GenerateBIP39Mnemonic 生成24个单词的BIP39助记词
// 使用256位熵，生成24个英文单词
// 返回助记词字符串（单词之间用空格分隔）和可能的错误
func GenerateBIP39Mnemonic() (string, error) {
	// 生成256位熵
	entropy, err := bip39.NewEntropy(EntropyBits)
	if err != nil {
		return "", fmt.Errorf("生成熵失败: %w", err)
	}

	// 转换为助记词（24个单词）
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("生成助记词失败: %w", err)
	}

	return mnemonic, nil
}

// DeriveKeyFromMnemonic 从BIP39助记词派生加密密钥
// 使用PBKDF2算法，迭代100000次，输出32字节密钥
// 参数:
//   - mnemonic: BIP39助记词（24个单词，空格分隔）
//
// 返回:
//   - 32字节密钥和可能的错误
func DeriveKeyFromMnemonic(mnemonic string) ([]byte, error) {
	// 验证助记词有效性
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("无效的助记词")
	}

	// 使用PBKDF2派生32字节密钥
	// PBKDF2(password=mnemonic, salt="vaulthub-recovery", iterations=100000, keyLen=32, hashFunc=SHA256)
	key := pbkdf2.Key(
		[]byte(mnemonic),
		[]byte(RecoverySalt),
		RecoveryIterations,
		RecoveryKeyLength,
		sha256.New,
	)

	return key, nil
}

// IsMnemonicValid 验证BIP39助记词是否有效
// 检查助记词格式、长度和校验和
// 参数:
//   - mnemonic: 待验证的助记词
//
// 返回:
//   - true表示有效，false表示无效
func IsMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// HashRecoveryKey 计算恢复密钥的SHA256哈希
// 用于在数据库中存储恢复密钥的哈希值，以便验证用户输入的恢复密钥是否正确
// 参数:
//   - recoveryKey: 从助记词派生的32字节密钥
//
// 返回:
//   - 64字符的十六进制哈希字符串
func HashRecoveryKey(recoveryKey []byte) string {
	hash := sha256.Sum256(recoveryKey)
	return fmt.Sprintf("%x", hash)
}
