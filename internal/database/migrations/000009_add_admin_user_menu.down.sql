-- 删除管理员用户管理菜单项
DELETE FROM menus WHERE path = '/admin/users' AND name = 'AdminUserManagement';
