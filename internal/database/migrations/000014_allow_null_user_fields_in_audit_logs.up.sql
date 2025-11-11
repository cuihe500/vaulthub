-- 修改审计日志表，允许user_uuid和username为空
-- 这样可以审计未认证的请求（失败的登录尝试、未授权访问等）

-- 修改user_uuid为可空
ALTER TABLE `audit_logs`
  MODIFY COLUMN `user_uuid` CHAR(36) DEFAULT NULL COMMENT '操作用户UUID（未认证时为空）';

-- 修改username为可空
ALTER TABLE `audit_logs`
  MODIFY COLUMN `username` VARCHAR(64) DEFAULT NULL COMMENT '操作用户名（未认证时为空）';

-- 索引保持不变，因为NULL值也可以被索引
