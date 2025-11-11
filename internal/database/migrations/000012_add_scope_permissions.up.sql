-- 添加数据作用域权限规则到Casbin
-- 背景：统一权限控制，消除硬编码的角色判断（原 scope.go 中的 if role != "admin"）
-- 目标：所有权限判断都通过Casbin完成，实现"一处定义，处处生效"

-- 定义全局数据访问权限（scope:global）
-- 只有admin角色可以访问全局数据，其他角色只能访问自己的数据
-- 用于审计日志查询、统计数据查询等需要控制数据范围的场景
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', 'scope', 'global');

-- 注意：
-- 1. user和readonly角色默认没有scope:global权限，因此只能访问自己的数据
-- 2. 如果未来需要添加审计员角色（auditor），只需在此添加：
--    INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'auditor', 'scope', 'global');
-- 3. 中间件层会检查scope:global权限，无权限时自动限制数据范围为当前用户
