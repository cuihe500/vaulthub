-- 创建用户相关表

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid CHAR(36) NOT NULL UNIQUE COMMENT '用户UUID',
    username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名',
    password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt哈希',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1=active, 2=disabled, 3=locked',
    role VARCHAR(32) NOT NULL DEFAULT 'user' COMMENT '角色名称',
    last_login_at DATETIME COMMENT '最后登录时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    INDEX idx_users_uuid (uuid),
    INDEX idx_users_username (username),
    INDEX idx_users_status (status),
    INDEX idx_users_role (role),
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

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
