import request from './request'

/**
 * 创建加密秘密
 */
export const createSecret = (data) => {
  return request.post('/v1/secrets', data)
}

/**
 * 获取秘密列表
 */
export const getSecretList = (params) => {
  return request.get('/v1/secrets', { params })
}

/**
 * 解密秘密（获取明文数据）
 */
export const decryptSecret = (secretUuid, data) => {
  return request.post(`/v1/secrets/${secretUuid}/decrypt`, data)
}

/**
 * 删除秘密
 */
export const deleteSecret = (secretUuid) => {
  return request.delete(`/v1/secrets/${secretUuid}`)
}
