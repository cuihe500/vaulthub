package crypto

import (
	"crypto/rand"
	"io"

	"github.com/cuihe500/vaulthub/pkg/errors"
)

const (
	// SaltSize 盐值大小（32字节）
	SaltSize = 32
)

// GenerateRandomBytes 生成加密安全的随机字节
// 使用 crypto/rand 包，适用于生成密钥、盐值、Nonce等敏感随机数
// 参数:
//   - size: 需要生成的字节数
// 返回:
//   - []byte: 随机字节
//   - error: 错误信息
func GenerateRandomBytes(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New(errors.CodeInvalidParam, "随机字节大小必须大于0")
	}

	bytes := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, errors.WithMessage(errors.CodeCryptoError, "生成随机字节失败", err)
	}

	return bytes, nil
}

// ClearBytes 安全清零字节切片
// 用于清除敏感数据（如密钥、密码）在内存中的痕迹
// 防止内存泄露或被dump时暴露敏感信息
// 参数:
//   - data: 需要清零的字节切片
func ClearBytes(data []byte) {
	// len()对nil切片返回0，无需单独检查nil
	if len(data) == 0 {
		return
	}

	// 将所有字节设置为0
	// Go的优化器可能会删除"无用"的赋值，但crypto/rand包建议的清零方式
	for i := range data {
		data[i] = 0
	}
}
