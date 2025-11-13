import request from './request'

/**
 * 发送验证码
 * @param {Object} data - 请求参数
 * @param {string} data.email - 邮箱地址
 * @param {string} data.purpose - 验证码用途: register, login, reset_password, change_email
 * @returns {Promise}
 */
export const sendVerificationCode = (data) => {
  return request.post('/v1/email/send-code', data)
}

/**
 * 验证验证码
 * @param {Object} data - 请求参数
 * @param {string} data.email - 邮箱地址
 * @param {string} data.purpose - 验证码用途
 * @param {string} data.code - 验证码
 * @returns {Promise}
 */
export const verifyCode = (data) => {
  return request.post('/v1/email/verify-code', data)
}
