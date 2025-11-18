-- 添加管理员用户管理菜单项
INSERT INTO menus (uuid, path, name, icon, title, roles, sort_order, is_visible) VALUES
(UUID(), '/admin/users', 'AdminUserManagement', 'UserCog', '用户管理', JSON_ARRAY('admin'), 4, TRUE);
