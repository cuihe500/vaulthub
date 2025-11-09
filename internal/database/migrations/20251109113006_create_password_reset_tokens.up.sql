-- 创建密码重置token表
CREATE TABLE IF NOT EXISTS `password_reset_tokens` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uuid` CHAR(36) NOT NULL COMMENT '对外唯一标识符',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '关联用户ID',
    `token_hash` VARCHAR(255) NOT NULL COMMENT 'Token哈希值（SHA256）',
    `expires_at` DATETIME NOT NULL COMMENT '过期时间（UTC）',
    `used_at` DATETIME DEFAULT NULL COMMENT '使用时间（UTC，NULL表示未使用）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（UTC）',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间（UTC）',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（UTC，软删除）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uuid` (`uuid`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_token_hash` (`token_hash`),
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置token表';
