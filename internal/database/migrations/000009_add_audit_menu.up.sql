-- 添加审计日志菜单项
-- 幂等性：先检查是否已存在该路径的菜单，不存在才插入
INSERT INTO menus (uuid, path, name, icon, title, roles, sort_order, is_visible)
SELECT UUID(), '/audit', 'AuditLog', 'Document', '审计日志', JSON_ARRAY(), 4, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM menus WHERE path = '/audit' AND deleted_at IS NULL
);
