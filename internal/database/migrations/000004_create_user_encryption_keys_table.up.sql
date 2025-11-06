-- 创建用户加密密钥表
CREATE TABLE IF NOT EXISTS user_encryption_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_uuid CHAR(36) NOT NULL UNIQUE COMMENT '用户UUID（关联users表）',

    -- KEK 派生参数
    kek_salt BINARY(32) NOT NULL COMMENT '用于从密码派生KEK的盐值',
    kek_algorithm VARCHAR(32) NOT NULL DEFAULT 'argon2id' COMMENT 'KEK派生算法',

    -- DEK 存储（被KEK加密）
    encrypted_dek VARBINARY(512) NOT NULL COMMENT '加密后的DEK（包含密文+nonce+tag）',
    dek_version INT NOT NULL DEFAULT 1 COMMENT 'DEK版本号（用于密钥轮换）',
    dek_algorithm VARCHAR(32) NOT NULL DEFAULT 'AES-256-GCM' COMMENT 'DEK加密算法',

    -- 恢复密钥
    recovery_key_hash CHAR(64) NOT NULL COMMENT '恢复密钥的SHA256哈希（用于验证）',
    encrypted_dek_recovery VARBINARY(512) NOT NULL COMMENT '用恢复密钥加密的DEK（备份）',

    -- 元数据
    last_rotation_at DATETIME NULL COMMENT '最后一次密钥轮换时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME NULL COMMENT '删除时间',

    INDEX idx_user_encryption_keys_user_uuid (user_uuid),
    INDEX idx_user_encryption_keys_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户加密密钥表';
