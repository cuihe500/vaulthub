-- 回滚权限策略补充
-- 删除本次迁移添加的所有权限规则

-- 删除secret资源权限
DELETE FROM casbin_rule WHERE ptype = 'p' AND v1 = 'secret';

-- 删除user资源显式权限（仅删除admin的显式声明，通配符策略保留）
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin' AND v1 = 'user';

-- 删除profile资源显式权限
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin' AND v1 = 'profile';

-- 删除config资源显式权限
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin' AND v1 = 'config';

-- 删除casbin资源权限
DELETE FROM casbin_rule WHERE ptype = 'p' AND v1 = 'casbin';

-- 注意：回滚后admin仍通过通配符拥有所有权限，但user/readonly角色将失去secret权限
