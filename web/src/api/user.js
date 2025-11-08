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
