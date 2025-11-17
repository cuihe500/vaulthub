-- 创建菜单表
CREATE TABLE IF NOT EXISTS menus (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    uuid CHAR(36) NOT NULL UNIQUE COMMENT '菜单UUID',
    path VARCHAR(128) NOT NULL COMMENT '路由路径',
    name VARCHAR(64) NOT NULL COMMENT '路由名称',
    icon VARCHAR(64) NOT NULL COMMENT '图标名称',
    title VARCHAR(64) NOT NULL COMMENT '菜单标题',
    roles JSON NOT NULL COMMENT '可访问角色列表，空数组表示所有角色',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序顺序，数字越小越靠前',
    is_visible BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否可见',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',
    INDEX idx_menus_uuid (uuid),
    INDEX idx_menus_path (path),
    INDEX idx_menus_sort (sort_order),
    INDEX idx_menus_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单配置表';

-- 插入初始菜单数据
INSERT INTO menus (uuid, path, name, icon, title, roles, sort_order, is_visible) VALUES
(UUID(), '/vault', 'Vault', 'Lock', '密钥管理', JSON_ARRAY(), 1, TRUE),
(UUID(), '/user', 'User', 'User', '用户管理', JSON_ARRAY(), 2, TRUE),
(UUID(), '/system', 'System', 'Setting', '系统配置', JSON_ARRAY('admin'), 3, TRUE);
