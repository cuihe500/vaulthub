<template>
  <div class="forgot-password-container">
    <div class="bg-decoration bg-decoration-1"></div>
    <div class="bg-decoration bg-decoration-2"></div>
    <div class="bg-decoration bg-decoration-3"></div>

    <el-card class="forgot-password-card">
      <div class="forgot-password-header">
        <div class="logo-wrapper">
          <div class="logo">V</div>
        </div>
        <h2 class="forgot-password-title">找回密码</h2>
        <p class="forgot-password-subtitle">请输入您的邮箱，我们将发送重置链接</p>
      </div>

      <el-form :model="form" :rules="rules" ref="formRef" class="forgot-password-form">
        <el-form-item prop="email">
          <el-input
            v-model="form.email"
            placeholder="请输入注册邮箱"
            :prefix-icon="Message"
            size="large"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-form-item class="submit-button-item">
          <el-button
            type="primary"
            @click="handleSubmit"
            :loading="loading"
            size="large"
            class="submit-button"
          >
            发送重置链接
          </el-button>
        </el-form-item>

        <div class="back-to-login">
          <router-link to="/login">返回登录</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Message } from '@element-plus/icons-vue'
import { requestPasswordReset } from '@/api/auth'

export default {
  name: 'ForgotPassword',
  setup() {
    const router = useRouter()
    const formRef = ref(null)
    const loading = ref(false)

    const form = reactive({
      email: ''
    })

    const rules = {
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
      ]
    }

    const handleSubmit = async () => {
      try {
        const valid = await formRef.value.validate()
        if (!valid) return

        loading.value = true
        // 获取当前访问的域名（包含协议，如 https://example.com）
        const domain = window.location.origin
        const data = await requestPasswordReset({
          email: form.email,
          domain: domain
        })

        ElMessage.success(data.message || '重置邮件已发送，请查收')

        // 延迟跳转到登录页
        setTimeout(() => {
          router.push('/login')
        }, 2000)
      } catch (error) {
        ElMessage.error(error.message || '发送失败，请稍后重试')
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      rules,
      formRef,
      loading,
      handleSubmit,
      Message
    }
  }
}
</script>

<style scoped>
.forgot-password-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  position: relative;
  overflow: hidden;
}

.bg-decoration {
  position: absolute;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

.bg-decoration-1 {
  width: 300px;
  height: 300px;
  top: -100px;
  left: -100px;
}

.bg-decoration-2 {
  width: 200px;
  height: 200px;
  bottom: -50px;
  right: -50px;
}

.bg-decoration-3 {
  width: 150px;
  height: 150px;
  top: 50%;
  left: 10%;
}

.forgot-password-card {
  width: 420px;
  padding: var(--spacing-2xl);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  position: relative;
  z-index: 1;
}

.forgot-password-header {
  text-align: center;
  margin-bottom: var(--spacing-2xl);
}

.logo-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: var(--spacing-lg);
}

.logo {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: var(--font-weight-bold);
  color: var(--color-white);
  box-shadow: var(--shadow-md);
}

.forgot-password-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm);
}

.forgot-password-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.forgot-password-form {
  margin-top: var(--spacing-xl);
}

.submit-button-item {
  margin-bottom: var(--spacing-md);
}

.submit-button {
  width: 100%;
  height: 44px;
  font-weight: var(--font-weight-medium);
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border: none;
  transition: all var(--duration-fast) var(--easing);
}

.submit-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.back-to-login {
  text-align: center;
  padding-top: var(--spacing-md);
}

.back-to-login a {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-sm);
  transition: color var(--duration-fast) var(--easing);
}

.back-to-login a:hover {
  color: var(--color-primary-dark);
  text-decoration: underline;
}
</style>
