-- 补充完整的权限策略矩阵
-- 背景：修复权限系统缺失问题
--   1. keys/secrets接口缺少权限验证（安全漏洞）
--   2. user/profile/config资源无策略定义（功能阻塞）
--   3. 资源字符串与实际策略不一致（维护隐患）
--
-- 修复策略：
--   - 补充secret资源的完整权限定义
--   - 为user/profile/config/casbin资源添加显式策略
--   - 明确各角色的权限边界（最小权限原则）

-- ========== Secret资源权限（新增）==========
-- 背景：secrets接口原本只有SecureAuth，缺少Casbin权限验证
-- 修复：为user角色添加secret的读写权限，readonly角色只能读
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'user', 'secret', 'read'),      -- user可以查看自己的秘密列表
    ('p', 'user', 'secret', 'write'),     -- user可以创建、解密、删除自己的秘密
    ('p', 'readonly', 'secret', 'read');  -- readonly只能查看秘密列表，不能创建/删除

-- ========== User资源权限（显式声明）==========
-- 背景：user资源已在路由中使用，但casbin_rule表中无策略
-- 说明：admin通过通配符已有权限，此处为显式声明提高可维护性
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'user', 'read'),   -- admin可查看所有用户信息
    ('p', 'admin', 'user', 'write');  -- admin可修改用户状态和角色

-- 注意：普通用户和readonly角色不应该有user资源权限（防止横向越权）

-- ========== Profile资源权限（显式声明）==========
-- 背景：profile资源已在管理员路由中使用，但casbin_rule表中无策略
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'profile', 'read'),   -- admin可查看所有用户档案
    ('p', 'admin', 'profile', 'write');  -- admin可修改所有用户档案

-- 注意：普通用户通过/api/v1/profile路由管理自己的档案，不走Casbin验证
--       该路由使用AuthWithAudit，业务层自动限制为当前用户

-- ========== Config资源权限（显式声明）==========
-- 背景：config资源已在路由中使用，但casbin_rule表中无策略
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'config', 'read'),   -- admin可查看系统配置
    ('p', 'admin', 'config', 'write');  -- admin可修改系统配置

-- 注意：系统配置属于敏感操作，仅admin可访问

-- ========== Casbin资源权限（新增）==========
-- 用于Casbin策略的热更新操作（reload接口）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'casbin', 'reload');  -- admin可重新加载权限策略

-- ========== 策略说明 ==========
-- 1. admin角色通过通配符('*', '*')拥有所有权限，以上显式声明是为了：
--    - 提高可读性：清晰展示系统的完整权限矩阵
--    - 便于审计：可直观看到哪些资源需要权限控制
--    - 简化维护：新增角色时可参考现有策略
--
-- 2. user角色权限边界：
--    - 可以：管理自己的vault/key/secret/profile
--    - 不可以：查看或管理其他用户、系统配置、用户列表
--
-- 3. readonly角色权限边界：
--    - 可以：查看自己的vault/key/secret（只读）
--    - 不可以：创建、修改、删除任何数据
--
-- 4. 如果未来需要添加新角色（如auditor审计员），参考模板：
--    INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
--        ('p', 'auditor', 'scope', 'global'),   -- 可查看全局审计日志
--        ('p', 'auditor', 'audit', 'read');     -- 可查看审计数据（需配合新资源）
