import request from './request'

/**
 * 获取密钥列表
 */
export const getSecretList = (params) => {
  return request.get('/v1/secrets', { params })
}

/**
 * 创建密钥
 */
export const createSecret = (data) => {
  return request.post('/v1/secrets', data)
}

/**
 * 删除密钥
 */
export const deleteSecret = (uuid) => {
  return request.delete(`/v1/secrets/${uuid}`)
}

/**
 * 解密密钥
 */
export const decryptSecret = (uuid, data) => {
  return request.post(`/v1/secrets/${uuid}/decrypt`, data)
}
