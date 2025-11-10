import request from './request'

/**
 * 查询审计日志
 */
export const queryAuditLogs = (params) => {
  return request.get('/v1/audit/logs', { params })
}

/**
 * 导出密钥类型统计
 */
export const exportStatistics = (params) => {
  return request.get('/v1/audit/logs/export', { params })
}

/**
 * 导出操作统计（按时间范围）
 */
export const exportOperationStatistics = (params) => {
  return request.get('/v1/audit/operations/export', { params })
}
