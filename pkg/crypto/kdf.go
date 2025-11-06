package crypto

import (
	"golang.org/x/crypto/argon2"
)

// Argon2id参数配置（平衡安全性和性能）
const (
	// Argon2Time Argon2id迭代次数
	Argon2Time = 3
	// Argon2Memory Argon2id内存消耗（64MB）
	Argon2Memory = 64 * 1024
	// Argon2Threads Argon2id并行度
	Argon2Threads = 4
	// Argon2KeyLength 输出密钥长度（32字节=256位）
	Argon2KeyLength = 32
)

// DeriveKEK 从用户密码派生KEK（密钥加密密钥）
// 使用Argon2id算法，提供高强度的密钥派生
// 参数:
//   - password: 用户密码
//   - salt: 随机盐值（必须是32字节）
// 返回:
//   - []byte: 派生的32字节密钥
//   - error: 错误信息
func DeriveKEK(password string, salt []byte) ([]byte, error) {
	// 使用Argon2id派生密钥
	// Argon2id结合了Argon2d（抗时间-内存权衡攻击）和Argon2i（抗侧信道攻击）的优点
	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,      // 时间成本（迭代次数）
		Argon2Memory,    // 内存成本（KB）
		Argon2Threads,   // 并行度
		Argon2KeyLength, // 输出长度
	)

	return key, nil
}
