-- 创建用户基本信息表
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '关联用户ID',
    nickname VARCHAR(50) NOT NULL COMMENT '用户昵称',
    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号',
    email VARCHAR(100) NOT NULL COMMENT '邮箱地址',
    email_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证',
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at DATETIME(3) DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (id),
    UNIQUE KEY idx_user_profiles_user_id (user_id),
    KEY idx_user_profiles_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基本信息表';
