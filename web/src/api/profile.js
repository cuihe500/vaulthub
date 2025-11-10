import request from './request'

/**
 * 获取当前用户档案
 */
export const getCurrentProfile = () => {
  return request.get('/v1/profile')
}

/**
 * 创建用户档案
 */
export const createProfile = (data) => {
  return request.post('/v1/profile', data)
}

/**
 * 更新用户档案（完整更新）
 */
export const updateProfile = (data) => {
  return request.put('/v1/profile', data)
}

/**
 * 部分更新用户档案
 */
export const patchProfile = (data) => {
  return request.patch('/v1/profile', data)
}

/**
 * 删除用户档案
 */
export const deleteProfile = () => {
  return request.delete('/v1/profile')
}

/**
 * 获取用户档案列表（管理员）
 */
export const getProfileList = (params) => {
  return request.get('/v1/admin/profiles', { params })
}

/**
 * 获取指定用户档案（管理员）
 */
export const getUserProfile = (userId) => {
  return request.get(`/v1/admin/users/${userId}/profile`)
}

/**
 * 更新指定用户档案（管理员）
 */
export const updateUserProfile = (userId, data) => {
  return request.put(`/v1/admin/users/${userId}/profile`, data)
}
