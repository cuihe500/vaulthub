import request from './request'

// 获取当前用户可访问的菜单
export const getUserMenus = () => {
  return request.get('/v1/menus')
}

// 获取所有菜单（管理员用）
export const getAllMenus = () => {
  return request.get('/v1/menus/all')
}
