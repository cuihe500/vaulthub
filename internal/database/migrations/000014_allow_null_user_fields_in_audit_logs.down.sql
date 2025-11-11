-- 回滚：恢复user_uuid和username为NOT NULL
-- 注意：这个回滚可能失败，如果表中存在NULL值的记录

-- 警告：先删除NULL值的记录，否则ALTER会失败
DELETE FROM `audit_logs` WHERE `user_uuid` IS NULL OR `username` IS NULL;

-- 恢复user_uuid为NOT NULL
ALTER TABLE `audit_logs`
  MODIFY COLUMN `user_uuid` CHAR(36) NOT NULL COMMENT '操作用户UUID';

-- 恢复username为NOT NULL
ALTER TABLE `audit_logs`
  MODIFY COLUMN `username` VARCHAR(64) NOT NULL COMMENT '操作用户名（冗余，便于查询）';
