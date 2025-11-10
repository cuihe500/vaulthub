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
