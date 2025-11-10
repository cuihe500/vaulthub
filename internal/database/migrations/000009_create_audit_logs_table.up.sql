-- 创建审计日志表
CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `uuid` CHAR(36) NOT NULL UNIQUE COMMENT '审计日志UUID',

    -- 操作主体
    `user_uuid` CHAR(36) NOT NULL COMMENT '操作用户UUID',
    `username` VARCHAR(64) NOT NULL COMMENT '操作用户名（冗余，便于查询）',

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
