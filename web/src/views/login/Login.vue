<template>
  <div class="login-container">
    <!-- 背景装饰元素 -->
    <div class="bg-decoration bg-decoration-1"></div>
    <div class="bg-decoration bg-decoration-2"></div>
    <div class="bg-decoration bg-decoration-3"></div>

    <el-card class="login-card">
      <!-- Logo 和品牌标识 -->
      <div class="login-header">
        <div class="logo-wrapper">
          <div class="logo">V</div>
        </div>
        <h2 class="login-title">VaultHub</h2>
        <p class="login-subtitle">安全的密钥管理系统</p>
      </div>

      <!-- 登录表单 -->
      <el-form :model="loginForm" :rules="rules" ref="loginFormRef" class="login-form">
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            :prefix-icon="User"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <!-- 辅助选项 -->
        <div class="login-options">
          <el-checkbox v-model="rememberMe">记住密码</el-checkbox>
          <a href="#" class="forgot-password">忘记密码？</a>
        </div>

        <el-form-item class="login-button-item">
          <el-button
            type="primary"
            @click="handleLogin"
            :loading="loading"
            size="large"
            class="login-button"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from 'vuex'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api/auth'

export default {
  name: 'Login',
  setup() {
    const router = useRouter()
    const store = useStore()
    const loginFormRef = ref(null)
    const loading = ref(false)
    const rememberMe = ref(false)

    const loginForm = reactive({
      username: '',
      password: ''
    })

    const rules = {
      username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
      password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
    }

    const handleLogin = async () => {
      try {
        const valid = await loginFormRef.value.validate()
        if (!valid) return

        loading.value = true
        const data = await login(loginForm)

        // 存储token
        store.dispatch('login', data.token)

        ElMessage.success('登录成功')
        router.push('/')
      } catch (error) {
        console.error('登录失败:', error)
      } finally {
        loading.value = false
      }
    }

    return {
      loginForm,
      rules,
      loginFormRef,
      loading,
      rememberMe,
      handleLogin,
      User,
      Lock
    }
  }
}
</script>

<style scoped>
/* 登录容器 - 全屏居中，渐变背景 */
.login-container {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  overflow: hidden;
  padding: var(--spacing-lg);
}

/* 背景装饰元素 - 几何图形 */
.bg-decoration {
  position: absolute;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.1);
  animation: float 20s ease-in-out infinite;
}

.bg-decoration-1 {
  width: 300px;
  height: 300px;
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.bg-decoration-2 {
  width: 200px;
  height: 200px;
  bottom: -50px;
  right: 10%;
  animation-delay: 5s;
}

.bg-decoration-3 {
  width: 150px;
  height: 150px;
  top: 20%;
  right: -50px;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
  }
}

/* 登录卡片 - 响应式宽度，磨砂效果，进入动画 */
.login-card {
  position: relative;
  width: 100%;
  max-width: 440px;
  padding: var(--spacing-2xl) var(--spacing-xl);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  animation: slideInUp var(--duration-slow) var(--easing);
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Logo 和品牌标识区域 */
.login-header {
  text-align: center;
  margin-bottom: var(--spacing-2xl);
}

.logo-wrapper {
  display: inline-flex;
  justify-content: center;
  align-items: center;
  width: 80px;
  height: 80px;
  margin-bottom: var(--spacing-md);
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-md);
}

.logo {
  font-size: 40px;
  font-weight: var(--font-weight-bold);
  color: var(--color-white);
  user-select: none;
}

.login-title {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  line-height: var(--line-height-tight);
}

.login-subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: var(--line-height-normal);
}

/* 登录表单 */
.login-form {
  margin-top: var(--spacing-xl);
}

/* 辅助选项 - 记住密码和忘记密码 */
.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.forgot-password {
  font-size: var(--font-size-sm);
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--duration-fast) var(--easing);
}

.forgot-password:hover {
  color: var(--color-primary-dark);
}

/* 登录按钮 */
.login-button-item {
  margin-bottom: 0;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-md);
  transition: all var(--duration-fast) var(--easing);
}

.login-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.login-button:active {
  transform: translateY(0);
}

/* 响应式设计 */
@media (max-width: 640px) {
  .login-container {
    padding: var(--spacing-md);
  }

  .login-card {
    padding: var(--spacing-xl) var(--spacing-lg);
  }

  .logo-wrapper {
    width: 64px;
    height: 64px;
  }

  .logo {
    font-size: 32px;
  }

  .login-title {
    font-size: var(--font-size-lg);
  }

  .bg-decoration {
    opacity: 0.5;
  }
}
</style>
