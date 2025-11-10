import request from './request'

/**
 * 创建用户加密密钥（首次使用）
 */
export const createEncryptionKey = (data) => {
  return request.post('/v1/keys/create', data)
}

/**
 * 验证恢复密钥有效性
 */
export const verifyRecoveryKey = (data) => {
  return request.post('/v1/keys/verify-recovery', data)
}

/**
 * 手动触发密钥轮换
 */
export const rotateKey = (data) => {
  return request.post('/v1/keys/rotate', data)
}

/**
 * 查询密钥轮换进度
 */
export const getRotationStatus = () => {
  return request.get('/v1/keys/rotation-status')
}
