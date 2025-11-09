-- 删除邮箱验证状态字段
ALTER TABLE `user_profiles`
DROP COLUMN `email_verified`;
