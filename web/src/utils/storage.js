const TOKEN_KEY = 'vaulthub_token'

/**
 * 获取Token
 */
export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * 设置Token
 */
export const setToken = (token) => {
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * 移除Token
 */
export const removeToken = () => {
  localStorage.removeItem(TOKEN_KEY)
}

/**
 * 获取本地存储
 */
export const getStorage = (key) => {
  const value = localStorage.getItem(key)
  try {
    return JSON.parse(value)
  } catch {
    return value
  }
}

/**
 * 设置本地存储
 */
export const setStorage = (key, value) => {
  localStorage.setItem(key, JSON.stringify(value))
}

/**
 * 移除本地存储
 */
export const removeStorage = (key) => {
  localStorage.removeItem(key)
}

/**
 * 清空本地存储
 */
export const clearStorage = () => {
  localStorage.clear()
}
