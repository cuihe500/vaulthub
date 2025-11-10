import request from './request'

/**
 * 获取用户当前统计（实时统计）
 */
export const getCurrentStatistics = (userUuid) => {
  const params = userUuid ? { user_uuid: userUuid } : {}
  return request.get('/v1/statistics/current', { params })
}

/**
 * 获取用户历史统计数据
 */
export const getUserStatistics = (params) => {
  return request.get('/v1/statistics/user', { params })
}
