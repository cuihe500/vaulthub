-- 添加密钥轮换相关字段到user_encryption_keys表
ALTER TABLE user_encryption_keys
ADD COLUMN encrypted_dek_old VARBINARY(512) NULL COMMENT '旧DEK（密钥轮换期间暂存，迁移完成后删除）',
ADD COLUMN rotation_status VARCHAR(20) NOT NULL DEFAULT 'none' COMMENT '轮换状态：none-无轮换, in_progress-轮换中, completed-已完成',
ADD COLUMN rotation_started_at DATETIME NULL COMMENT '轮换开始时间';

-- 添加索引以便查询轮换状态
CREATE INDEX idx_user_encryption_keys_rotation_status ON user_encryption_keys(rotation_status);
