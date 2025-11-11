-- 回滚数据作用域权限规则
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin' AND v1 = 'scope' AND v2 = 'global';
