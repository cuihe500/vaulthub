-- 创建加密和密钥管理相关表

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

    -- 安全密码（Security PIN）
    security_pin_hash VARCHAR(255) COMMENT '安全密码的bcrypt哈希（用于验证加密密码，独立于认证密码）',

    -- 恢复密钥
    recovery_key_hash CHAR(64) NOT NULL COMMENT '恢复密钥的SHA256哈希（用于验证）',
    encrypted_dek_recovery VARBINARY(512) NOT NULL COMMENT '用恢复密钥加密的DEK（备份）',

    -- 密钥轮换字段
    encrypted_dek_old VARBINARY(512) NULL COMMENT '旧DEK（密钥轮换期间暂存，迁移完成后删除）',
    rotation_status VARCHAR(20) NOT NULL DEFAULT 'none' COMMENT '轮换状态：none-无轮换, in_progress-轮换中, completed-已完成',
    rotation_started_at DATETIME NULL COMMENT '轮换开始时间',

    -- 元数据
    last_rotation_at DATETIME NULL COMMENT '最后一次密钥轮换时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME NULL COMMENT '删除时间',

    INDEX idx_user_encryption_keys_user_uuid (user_uuid),
    INDEX idx_user_encryption_keys_rotation_status (rotation_status),
    INDEX idx_user_encryption_keys_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户加密密钥表';

-- 创建加密秘密表
CREATE TABLE IF NOT EXISTS encrypted_secrets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- 所属用户和唯一标识
    user_uuid CHAR(36) NOT NULL COMMENT '所属用户UUID',
    secret_uuid CHAR(36) NOT NULL UNIQUE COMMENT '秘密的唯一标识（对外暴露）',

    -- 业务信息
    secret_name VARCHAR(255) NOT NULL COMMENT '秘密名称',
    secret_type VARCHAR(32) NOT NULL COMMENT '秘密类型：api_key, db_credential, certificate, ssh_key',
    description TEXT COMMENT '描述信息',

    -- 加密数据
    encrypted_data BLOB NOT NULL COMMENT '加密后的数据',
    dek_version INT NOT NULL COMMENT '使用的DEK版本',
    nonce BINARY(12) NOT NULL COMMENT 'AES-GCM的Nonce（12字节）',
    auth_tag BINARY(16) NOT NULL COMMENT 'AES-GCM的认证标签（16字节）',

    -- 元数据（JSON格式，可扩展）
    metadata JSON COMMENT '额外元数据：过期时间、标签等',

    -- 审计
    last_accessed_at DATETIME NULL COMMENT '最后访问时间',
    access_count BIGINT DEFAULT 0 COMMENT '访问次数',

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME NULL COMMENT '删除时间',

    INDEX idx_encrypted_secrets_user_uuid (user_uuid),
    INDEX idx_encrypted_secrets_secret_uuid (secret_uuid),
    INDEX idx_encrypted_secrets_secret_type (secret_type),
    INDEX idx_encrypted_secrets_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加密秘密表';
