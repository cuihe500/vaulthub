-- 回滚密钥轮换相关字段
DROP INDEX idx_user_encryption_keys_rotation_status ON user_encryption_keys;

ALTER TABLE user_encryption_keys
DROP COLUMN encrypted_dek_old,
DROP COLUMN rotation_status,
DROP COLUMN rotation_started_at;
