import request from './request'

/**
 * 获取用户列表
 */
export const getUserList = (params) => {
  return request.get('/v1/users', { params })
}

/**
 * 获取用户详情
 */
export const getUserDetail = (uuid) => {
  return request.get(`/v1/users/${uuid}`)
}

/**
 * 创建用户
 */
export const createUser = (data) => {
  return request.post('/v1/users', data)
}

/**
 * 更新用户
 */
export const updateUser = (uuid, data) => {
  return request.put(`/v1/users/${uuid}`, data)
}

/**
 * 删除用户
 */
export const deleteUser = (uuid) => {
  return request.delete(`/v1/users/${uuid}`)
}

/**
 * 修改密码
 */
export const changePassword = (data) => {
  return request.post('/v1/users/change-password', data)
}

/**
 * 更新用户状态
 */
export const updateUserStatus = (uuid, status) => {
  return request.put(`/v1/users/${uuid}/status`, { status })
}

/**
 * 更新用户角色
 */
export const updateUserRole = (uuid, role) => {
  return request.put(`/v1/users/${uuid}/role`, { role })
}

/**
 * 更新用户基本信息
 */
export const updateUserInfo = (uuid, data) => {
  return request.put(`/v1/users/${uuid}/info`, data)
}

/**
 * 重置用户密码（管理员操作）
 */
export const resetUserPassword = (uuid, password) => {
  return request.post(`/v1/users/${uuid}/reset-password`, { password })
}
