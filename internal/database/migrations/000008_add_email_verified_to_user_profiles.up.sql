-- 添加邮箱验证状态字段到用户基本信息表
ALTER TABLE user_profiles
ADD COLUMN email_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证' AFTER email;
