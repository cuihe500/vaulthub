import request from './request'

/**
 * 用户登录
 */
export const login = (data) => {
  return request.post('/v1/auth/login', data)
}

/**
 * 用户注册
 */
export const register = (data) => {
  return request.post('/v1/auth/register', data)
}

/**
 * 用户登出
 */
export const logout = () => {
  return request.post('/v1/auth/logout')
}

/**
 * 获取当前用户信息
 */
export const getCurrentUser = () => {
  return request.get('/v1/auth/current')
}

/**
 * 刷新Token
 */
export const refreshToken = () => {
  return request.post('/v1/auth/refresh')
}

/**
 * 请求密码重置
 */
export const requestPasswordReset = (data) => {
  return request.post('/v1/auth/request-password-reset', data)
}

/**
 * 验证密码重置token
 */
export const verifyResetToken = (token) => {
  return request.get('/v1/auth/verify-reset-token', { params: { token } })
}

/**
 * 使用token重置密码
 */
export const resetPasswordWithToken = (data) => {
  return request.post('/v1/auth/reset-password-with-token', data)
}
