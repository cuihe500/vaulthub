import request from './request'

/**
 * 获取所有系统配置
 */
export const getConfigs = () => {
  return request.get('/v1/configs')
}

/**
 * 获取单个配置
 */
export const getConfig = (key) => {
  return request.get(`/v1/configs/${key}`)
}

/**
 * 更新单个配置
 */
export const updateConfig = (key, data) => {
  return request.put(`/v1/configs/${key}`, data)
}

/**
 * 批量更新配置
 */
export const batchUpdateConfigs = (data) => {
  return request.put('/v1/configs/batch', data)
}

/**
 * 重新加载配置
 */
export const reloadConfigs = () => {
  return request.post('/v1/configs/reload')
}
