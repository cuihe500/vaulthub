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
  return request.get('/v1/auth/me')
}

/**
 * 刷新Token
 */
export const refreshToken = () => {
  return request.post('/v1/auth/refresh')
}
