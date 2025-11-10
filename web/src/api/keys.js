import request from './request'

/**
 * 创建用户加密密钥（首次设置安全密码）
 * 返回24词恢复助记词，用户必须妥善保管
 */
export const createEncryptionKey = (data) => {
  return request.post('/v1/keys/create', data)
}

/**
 * 获取安全密码设置状态
 * 返回 { has_security_pin: boolean }
 */
export const getSecurityPINStatus = () => {
  return request.get('/v1/auth/security-pin-status')
}

/**
 * 使用恢复助记词重置安全密码
 * 返回新的恢复助记词，旧助记词失效
 */
export const resetSecurityPIN = (data) => {
  return request.post('/v1/auth/reset-password', data)
}

/**
 * 验证恢复助记词是否正确
 */
export const verifyRecoveryKey = (data) => {
  return request.post('/v1/keys/verify-recovery', data)
}
