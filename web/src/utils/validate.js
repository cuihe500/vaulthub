/**
 * 验证邮箱格式
 */
export const isEmail = (email) => {
  const reg = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/
  return reg.test(email)
}

/**
 * 验证手机号格式（中国大陆）
 */
export const isPhone = (phone) => {
  const reg = /^1[3-9]\d{9}$/
  return reg.test(phone)
}

/**
 * 验证密码强度
 * 至少8位，包含大小写字母、数字和特殊字符
 */
export const isStrongPassword = (password) => {
  const reg = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/
  return reg.test(password)
}

/**
 * 验证用户名
 * 4-20位字母、数字、下划线
 */
export const isUsername = (username) => {
  const reg = /^[a-zA-Z0-9_]{4,20}$/
  return reg.test(username)
}

/**
 * 验证URL格式
 */
export const isURL = (url) => {
  const reg = /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/
  return reg.test(url)
}

/**
 * 验证IP地址
 */
export const isIP = (ip) => {
  const reg = /^(\d{1,3}\.){3}\d{1,3}$/
  return reg.test(ip)
}
