import request from './request'

/**
 * 获取密钥库列表
 */
export const getVaultList = (params) => {
  return request.get('/v1/vaults', { params })
}

/**
 * 获取密钥库详情
 */
export const getVaultDetail = (uuid) => {
  return request.get(`/v1/vaults/${uuid}`)
}

/**
 * 创建密钥库
 */
export const createVault = (data) => {
  return request.post('/v1/vaults', data)
}

/**
 * 更新密钥库
 */
export const updateVault = (uuid, data) => {
  return request.put(`/v1/vaults/${uuid}`, data)
}

/**
 * 删除密钥库
 */
export const deleteVault = (uuid) => {
  return request.delete(`/v1/vaults/${uuid}`)
}

/**
 * 轮换密钥
 */
export const rotateVault = (uuid) => {
  return request.post(`/v1/vaults/${uuid}/rotate`)
}
