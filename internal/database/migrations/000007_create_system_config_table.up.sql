-- 创建系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值',
    description VARCHAR(255) COMMENT '配置说明',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_system_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 初始化admin_initialized标志为false
INSERT INTO system_config (config_key, config_value, description)
VALUES ('admin_initialized', 'false', '超级管理员是否已初始化');

-- 初始化密钥轮换配置
INSERT INTO system_config (config_key, config_value, description)
VALUES
    ('key_rotation_batch_size', '100', '密钥轮换每批处理数量'),
    ('key_rotation_batch_sleep_ms', '100', '密钥轮换批次间休眠时间(毫秒)');

-- 默认限流配置：每秒5次
INSERT INTO system_config (config_key, config_value, description)
VALUES (
    'rate_limit.default',
    '{"requests": 5, "period": "second"}',
    '默认限流配置：单用户/IP每秒5次'
);

-- 接口级限流配置
INSERT INTO system_config (config_key, config_value, description)
VALUES (
    'rate_limit.endpoints',
    '{"/api/v1/auth/login": {"requests": 10, "period": "minute"}, "/api/v1/auth/register": {"requests": 5, "period": "hour"}}',
    '接口级限流配置：登录每分钟10次，注册每小时5次'
);
