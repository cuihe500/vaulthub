-- 创建Casbin规则表并初始化完整权限策略
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

-- ========== Admin角色权限（通配符）==========
-- admin角色拥有所有权限（通配符策略）
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', '*', '*');

-- ========== User角色权限 ==========
-- vault和key资源的读写权限
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'user', 'vault', 'read'),
    ('p', 'user', 'vault', 'write'),
    ('p', 'user', 'key', 'read'),
    ('p', 'user', 'key', 'write');

-- secret资源的读写权限
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'user', 'secret', 'read'),
    ('p', 'user', 'secret', 'write');

-- ========== Readonly角色权限 ==========
-- vault和key资源的只读权限
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'readonly', 'vault', 'read'),
    ('p', 'readonly', 'key', 'read');

-- secret资源的只读权限
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'readonly', 'secret', 'read');

-- ========== Admin角色显式权限（提高可维护性）==========
-- 虽然admin已通过通配符拥有所有权限，但显式声明有助于：
-- 1. 提高可读性：清晰展示系统的完整权限矩阵
-- 2. 便于审计：直观看到哪些资源需要权限控制
-- 3. 简化维护：新增角色时可参考现有策略

-- user资源权限（用户管理）
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'user', 'read'),
    ('p', 'admin', 'user', 'write');

-- profile资源权限（用户档案管理）
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'profile', 'read'),
    ('p', 'admin', 'profile', 'write');

-- config资源权限（系统配置管理）
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'config', 'read'),
    ('p', 'admin', 'config', 'write');

-- casbin资源权限（权限策略热更新）
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'casbin', 'reload');

-- scope资源权限（数据作用域控制）
-- 背景：统一权限控制，消除硬编码的角色判断
-- 目标：所有权限判断都通过Casbin完成，实现"一处定义，处处生效"
-- 只有admin角色可以访问全局数据（scope:global），其他角色只能访问自己的数据
INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'scope', 'global');

-- ========== 权限策略说明 ==========
-- 1. admin角色通过通配符('*', '*')拥有所有权限，显式声明是为了提高可读性和可维护性
--
-- 2. user角色权限边界：
--    - 可以：管理自己的vault/key/secret/profile
--    - 不可以：查看或管理其他用户、系统配置、用户列表
--
-- 3. readonly角色权限边界：
--    - 可以：查看自己的vault/key/secret（只读）
--    - 不可以：创建、修改、删除任何数据
--
-- 4. scope:global权限：
--    - admin可以访问全局数据（所有用户的审计日志、统计数据等）
--    - user和readonly只能访问自己的数据
--    - 中间件层会检查scope:global权限，无权限时自动限制数据范围为当前用户
--
-- 5. 如果未来需要添加新角色（如auditor审计员），参考模板：
--    INSERT IGNORE INTO casbin_rule (ptype, v0, v1, v2) VALUES
--        ('p', 'auditor', 'scope', 'global'),
--        ('p', 'auditor', 'audit', 'read');
