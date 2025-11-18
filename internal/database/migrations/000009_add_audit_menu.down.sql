-- 删除审计日志菜单项（软删除）
UPDATE menus
SET deleted_at = CURRENT_TIMESTAMP
WHERE path = '/audit' AND deleted_at IS NULL;
