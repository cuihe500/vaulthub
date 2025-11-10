import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getToken } from '@/utils/storage'
import { getSecurityPINStatus } from '@/api/keys'
import { getCurrentUser } from '@/api/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/login/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/views/login/ForgotPassword.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('@/views/login/ResetPassword.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/setup-security-pin',
    name: 'SetupSecurityPin',
    component: () => import('@/views/security/SetupSecurityPin.vue'),
    meta: { requiresAuth: true, skipSecurityPinCheck: true }
  },
  {
    path: '/reset-security-pin',
    name: 'ResetSecurityPin',
    component: () => import('@/views/security/ResetSecurityPin.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    redirect: '/vault'
  },
  {
    path: '/vault',
    name: 'Vault',
    component: () => import('@/views/vault/VaultList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/user',
    name: 'User',
    component: () => import('@/views/user/UserManagement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/system/config',
    name: 'SystemConfig',
    component: () => import('@/views/system/SystemConfig.vue'),
    meta: { requiresAuth: true, requiresRole: 'admin' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局前置守卫：验证登录状态、角色权限和安全密码设置
router.beforeEach(async (to, from, next) => {
  const token = getToken()

  // 1. 检查登录状态
  if (to.meta.requiresAuth && !token) {
    // 需要登录但未登录，跳转登录页
    next('/login')
    return
  }

  if ((to.path === '/login' || to.path === '/register') && token) {
    // 已登录但访问登录/注册页，跳转首页
    next('/')
    return
  }

  // 2. 检查角色权限（仅对需要特定角色的路由）
  if (to.meta.requiresRole && token) {
    try {
      const userInfo = await getCurrentUser()
      const userRole = userInfo.role

      // 检查用户角色是否匹配
      if (userRole !== to.meta.requiresRole) {
        ElMessage.error('您没有权限访问此页面')
        next('/')
        return
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      ElMessage.error('获取用户信息失败')
      next('/login')
      return
    }
  }

  // 3. 检查安全密码设置状态（仅对需要认证且未跳过检查的路由）
  if (to.meta.requiresAuth && !to.meta.skipSecurityPinCheck && token) {
    try {
      const status = await getSecurityPINStatus()
      if (!status.has_security_pin) {
        // 未设置安全密码，强制跳转到设置页面
        if (to.path !== '/setup-security-pin') {
          next('/setup-security-pin')
          return
        }
      }
    } catch (error) {
      // API 调用失败，允许继续（避免因网络问题阻塞）
      console.error('检查安全密码状态失败:', error)
    }
  }

  next()
})

export default router
