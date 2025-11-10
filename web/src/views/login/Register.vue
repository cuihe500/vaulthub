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
        <p class="login-subtitle">创建您的账号</p>
      </div>

      <!-- 注册表单 -->
      <el-form :model="registerForm" :rules="rules" ref="registerFormRef" class="login-form">
        <el-form-item prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="请输入用户名（3-32字符）"
            :prefix-icon="User"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="请输入密码（至少8位）"
            :prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>

        <el-form-item prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            :prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>

        <el-form-item prop="email">
          <el-input
            v-model="registerForm.email"
            placeholder="请输入邮箱（必填）"
            :prefix-icon="Message"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="code">
          <div class="verification-code-wrapper">
            <el-input
              v-model="registerForm.code"
              placeholder="请输入验证码"
              :prefix-icon="Lock"
              size="large"
              class="code-input"
            />
            <el-button
              type="primary"
              size="large"
              @click="handleSendCode"
              :disabled="sendCodeDisabled || countdown > 0"
              :loading="sendingCode"
              class="send-code-button"
            >
              {{ countdown > 0 ? `${countdown}s后重试` : '发送验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item prop="nickname">
          <el-input
            v-model="registerForm.nickname"
            placeholder="请输入昵称（可选，默认使用用户名）"
            :prefix-icon="UserFilled"
            size="large"
          />
        </el-form-item>

        <el-form-item class="login-button-item">
          <el-button
            type="primary"
            @click="handleRegister"
            :loading="loading"
            size="large"
            class="login-button"
          >
            注册
          </el-button>
        </el-form-item>

        <!-- 返回登录 -->
        <div class="back-to-login">
          <span>已有账号？</span>
          <router-link to="/login" class="login-link">立即登录</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Message, UserFilled } from '@element-plus/icons-vue'
import { register, sendEmailCode } from '@/api/auth'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const registerFormRef = ref(null)
    const loading = ref(false)
    const sendingCode = ref(false)
    const countdown = ref(0)
    let countdownTimer = null

    const registerForm = reactive({
      username: '',
      password: '',
      confirmPassword: '',
      email: '',
      code: '',
      nickname: ''
    })

    // 计算发送验证码按钮是否禁用
    const sendCodeDisabled = computed(() => {
      // 邮箱为空或格式不正确时禁用
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return !registerForm.email || !emailRegex.test(registerForm.email)
    })

    // 自定义验证器：确认密码
    const validateConfirmPassword = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请再次输入密码'))
      } else if (value !== registerForm.password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    const rules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
        { min: 3, max: 32, message: '用户名长度为3-32个字符', trigger: 'blur' }
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
        { min: 8, message: '密码长度至少为8位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, validator: validateConfirmPassword, trigger: 'blur' }
      ],
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
      ],
      code: [
        { required: true, message: '请输入验证码', trigger: 'blur' },
        { min: 6, max: 6, message: '验证码为6位数字', trigger: 'blur' }
      ],
      nickname: [
        { min: 1, max: 50, message: '昵称长度为1-50个字符', trigger: 'blur' }
      ]
    }

    // 发送验证码
    const handleSendCode = async () => {
      // 先验证邮箱格式
      try {
        await registerFormRef.value.validateField('email')
      } catch (error) {
        return
      }

      try {
        sendingCode.value = true

        await sendEmailCode({
          email: registerForm.email,
          purpose: 'register'
        })

        ElMessage.success('验证码已发送，请查收邮件')

        // 开始倒计时（60秒）
        countdown.value = 60
        countdownTimer = setInterval(() => {
          countdown.value--
          if (countdown.value <= 0) {
            clearInterval(countdownTimer)
            countdownTimer = null
          }
        }, 1000)
      } catch (error) {
        console.error('发送验证码失败:', error)
      } finally {
        sendingCode.value = false
      }
    }

    const handleRegister = async () => {
      try {
        const valid = await registerFormRef.value.validate()
        if (!valid) return

        loading.value = true

        // 构造请求数据，移除 confirmPassword
        const requestData = {
          username: registerForm.username,
          password: registerForm.password,
          email: registerForm.email,
          code: registerForm.code
        }

        // 只在有值时添加可选字段
        if (registerForm.nickname) {
          requestData.nickname = registerForm.nickname
        }

        await register(requestData)

        ElMessage.success('注册成功，请登录')
        router.push('/login')
      } catch (error) {
        console.error('注册失败:', error)
      } finally {
        loading.value = false
      }
    }

    return {
      registerForm,
      rules,
      registerFormRef,
      loading,
      sendingCode,
      countdown,
      sendCodeDisabled,
      handleSendCode,
      handleRegister,
      User,
      Lock,
      Message,
      UserFilled
    }
  }
}
</script>

<style scoped>
/* 复用 Login.vue 的样式系统 */

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

/* 验证码输入框容器 */
.verification-code-wrapper {
  display: flex;
  gap: var(--spacing-sm);
}

.code-input {
  flex: 1;
}

.send-code-button {
  flex-shrink: 0;
  min-width: 120px;
  font-weight: var(--font-weight-medium);
}

/* 登录按钮 */
.login-button-item {
  margin-bottom: var(--spacing-md);
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

/* 返回登录 */
.back-to-login {
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.login-link {
  margin-left: var(--spacing-xs);
  color: var(--color-primary);
  text-decoration: none;
  font-weight: var(--font-weight-medium);
  transition: color var(--duration-fast) var(--easing);
}

.login-link:hover {
  color: var(--color-primary-dark);
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
