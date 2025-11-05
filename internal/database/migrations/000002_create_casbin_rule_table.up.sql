-- 创建Casbin规则表
-- 特例豁免：此表为 Casbin 框架标准表结构，不包含 created_at/updated_at/deleted_at 字段
-- 原因：Casbin 适配器依赖固定的表结构，添加额外字段可能影响框架功能
CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(100) COMMENT '策略类型',
    v0 VARCHAR(100) COMMENT '主体/角色',
    v1 VARCHAR(100) COMMENT '对象/资源',
    v2 VARCHAR(100) COMMENT '操作',
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    UNIQUE KEY unique_key_casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Casbin权限规则表';

-- 初始化默认权限规则
-- admin角色拥有所有权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', '*', '*');

-- user角色的权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'user', 'vault', 'read'),
    ('p', 'user', 'vault', 'write'),
    ('p', 'user', 'key', 'read'),
    ('p', 'user', 'key', 'write');

-- readonly角色的权限
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'readonly', 'vault', 'read'),
    ('p', 'readonly', 'key', 'read');

-- user管理权限（仅admin拥有）
-- admin角色已经通过通配符获得所有权限，这里不需要额外添加
