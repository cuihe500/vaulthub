/**
 * 日期时间工具函数
 */

/**
 * 将Date对象转换为RFC3339格式字符串（不含毫秒）
 * 后端期望的时间格式: 2006-01-02T15:04:05Z07:00
 * @param {Date} date - 日期对象
 * @returns {string} RFC3339格式字符串，例如: 2025-11-09T16:00:00Z
 */
export const toRFC3339 = (date) => {
  if (!(date instanceof Date)) {
    throw new Error('参数必须是Date对象')
  }

  // 获取ISO字符串并移除毫秒部分
  // 原始格式: 2025-11-09T16:00:00.123Z
  // 目标格式: 2025-11-09T16:00:00Z
  const isoString = date.toISOString()
  return isoString.replace(/\.\d{3}Z$/, 'Z')
}

/**
 * 获取今天开始时间（本地时区00:00:00对应的UTC时间）
 * 例如：本地时区2025-11-10 00:00:00 (CST) -> UTC 2025-11-09 16:00:00Z
 * @returns {string} RFC3339格式字符串
 */
export const getTodayStart = () => {
  const now = new Date()
  // Date构造函数使用本地时区，这里是CST（UTC+8）
  const localToday = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0)
  // toISOString()自动转换为UTC时间
  return toRFC3339(localToday)
}

/**
 * 获取今天结束时间（本地时区明天00:00:00对应的UTC时间）
 * 例如：本地时区2025-11-11 00:00:00 (CST) -> UTC 2025-11-10 16:00:00Z
 * @returns {string} RFC3339格式字符串
 */
export const getTodayEnd = () => {
  const now = new Date()
  // 本地时区的明天0点
  const localTomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1, 0, 0, 0, 0)
  return toRFC3339(localTomorrow)
}

/**
 * 获取指定日期范围（UTC时区）
 * @param {Date} startDate - 开始日期
 * @param {Date} endDate - 结束日期
 * @returns {{ start: string, end: string }} RFC3339格式的时间范围
 */
export const getDateRange = (startDate, endDate) => {
  const start = new Date(startDate)
  start.setUTCHours(0, 0, 0, 0)

  const end = new Date(endDate)
  end.setUTCHours(23, 59, 59, 0)

  return {
    start: toRFC3339(start),
    end: toRFC3339(end)
  }
}
