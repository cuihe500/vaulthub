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
