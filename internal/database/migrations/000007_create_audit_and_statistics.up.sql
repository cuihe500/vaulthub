-- 创建审计日志表
CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` CHAR(36) NOT NULL UNIQUE COMMENT '审计日志UUID',

    -- 操作主体（未认证时为NULL，用于审计失败的登录尝试等）
    `user_uuid` CHAR(36) DEFAULT NULL COMMENT '操作用户UUID（未认证时为NULL）',
    `username` VARCHAR(64) DEFAULT NULL COMMENT '操作用户名（未认证时为NULL）',

    -- 操作内容
    `action_type` VARCHAR(32) NOT NULL COMMENT '操作类型：CREATE/UPDATE/DELETE/ACCESS/LOGIN/LOGOUT',
    `resource_type` VARCHAR(32) NOT NULL COMMENT '资源类型：vault/secret/user/config',
    `resource_uuid` CHAR(36) DEFAULT NULL COMMENT '资源UUID',
    `resource_name` VARCHAR(255) DEFAULT NULL COMMENT '资源名称（冗余，便于展示）',

    -- 操作结果
    `status` VARCHAR(16) NOT NULL COMMENT '操作状态：success/failed',
    `error_code` INT DEFAULT NULL COMMENT '错误码（失败时记录）',
    `error_message` TEXT DEFAULT NULL COMMENT '错误信息（失败时记录）',

    -- 上下文信息
    `ip_address` VARCHAR(45) DEFAULT NULL COMMENT '客户端IP地址',
    `user_agent` VARCHAR(512) DEFAULT NULL COMMENT '客户端User-Agent',
    `request_id` VARCHAR(64) DEFAULT NULL COMMENT '请求ID（用于追踪）',

    -- 额外数据
    `details` JSON DEFAULT NULL COMMENT '操作详细信息（JSON格式）',

    -- 时间戳（UTC）
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（UTC）',

    -- 索引
    INDEX `idx_user_uuid` (`user_uuid`),
    INDEX `idx_username` (`username`),
    INDEX `idx_action_type` (`action_type`),
    INDEX `idx_resource_type` (`resource_type`),
    INDEX `idx_resource_uuid` (`resource_uuid`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_user_created` (`user_uuid`, `created_at`) COMMENT '用户时间复合索引，优化用户查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- 创建用户统计表
CREATE TABLE IF NOT EXISTS `user_statistics` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` CHAR(36) NOT NULL UNIQUE COMMENT '统计记录UUID',

    -- 统计维度
    `user_uuid` CHAR(36) NOT NULL COMMENT '用户UUID',
    `stat_date` DATE NOT NULL COMMENT '统计日期（UTC）',
    `stat_type` VARCHAR(16) NOT NULL COMMENT '统计类型：daily/weekly/monthly',

    -- 密钥数量统计
    `total_secrets` INT NOT NULL DEFAULT 0 COMMENT '总密钥数',
    `api_key_count` INT NOT NULL DEFAULT 0 COMMENT 'API密钥数量',
    `password_count` INT NOT NULL DEFAULT 0 COMMENT '密码类型数量',
    `certificate_count` INT NOT NULL DEFAULT 0 COMMENT '证书类型数量',
    `ssh_key_count` INT NOT NULL DEFAULT 0 COMMENT 'SSH密钥数量',
    `private_key_count` INT NOT NULL DEFAULT 0 COMMENT '私钥数量',
    `other_count` INT NOT NULL DEFAULT 0 COMMENT '其他类型数量',

    -- 操作次数统计
    `create_count` INT NOT NULL DEFAULT 0 COMMENT '创建操作次数',
    `update_count` INT NOT NULL DEFAULT 0 COMMENT '更新操作次数',
    `delete_count` INT NOT NULL DEFAULT 0 COMMENT '删除操作次数',
    `access_count` INT NOT NULL DEFAULT 0 COMMENT '访问操作次数',
    `total_operations` INT NOT NULL DEFAULT 0 COMMENT '总操作次数',

    -- 登录统计
    `login_count` INT NOT NULL DEFAULT 0 COMMENT '登录次数',
    `failed_login_count` INT NOT NULL DEFAULT 0 COMMENT '失败登录次数',

    -- 时间戳（UTC）
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（UTC）',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间（UTC）',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（UTC）',

    -- 索引
    UNIQUE INDEX `idx_user_date_type` (`user_uuid`, `stat_date`, `stat_type`) COMMENT '用户日期类型唯一索引',
    INDEX `idx_stat_date` (`stat_date`),
    INDEX `idx_stat_type` (`stat_type`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户统计表';

-- 创建密码重置token表
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    uuid CHAR(36) NOT NULL COMMENT '对外唯一标识符',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '关联用户ID',
    token_hash VARCHAR(255) NOT NULL COMMENT 'Token哈希值（SHA256）',
    expires_at DATETIME NOT NULL COMMENT '过期时间（UTC）',
    used_at DATETIME DEFAULT NULL COMMENT '使用时间（UTC，NULL表示未使用）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（UTC）',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间（UTC）',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间（UTC，软删除）',
    PRIMARY KEY (id),
    UNIQUE KEY idx_password_reset_tokens_uuid (uuid),
    KEY idx_password_reset_tokens_user_id (user_id),
    KEY idx_password_reset_tokens_token_hash (token_hash),
    KEY idx_password_reset_tokens_expires_at (expires_at),
    KEY idx_password_reset_tokens_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置token表';
