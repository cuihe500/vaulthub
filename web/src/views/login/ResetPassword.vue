<template>
  <div class="reset-password-container">
    <div class="bg-decoration bg-decoration-1"></div>
    <div class="bg-decoration bg-decoration-2"></div>
    <div class="bg-decoration bg-decoration-3"></div>

    <el-card class="reset-password-card">
      <div class="reset-password-header">
        <div class="logo-wrapper">
          <div class="logo">V</div>
        </div>
        <h2 class="reset-password-title">重置密码</h2>
        <p class="reset-password-subtitle" v-if="!tokenVerified">正在验证重置链接...</p>
        <p class="reset-password-subtitle" v-else-if="tokenError">{{ tokenError }}</p>
        <p class="reset-password-subtitle" v-else>请输入新密码</p>
      </div>

      <div v-if="tokenVerified && !tokenError">
        <el-form :model="form" :rules="rules" ref="formRef" class="reset-password-form">
          <el-form-item prop="newPassword">
            <el-input
              v-model="form.newPassword"
              type="password"
              placeholder="请输入新密码（至少8位）"
              :prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>

          <el-form-item prop="confirmPassword">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
              :prefix-icon="Lock"
              size="large"
              show-password
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
              重置密码
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <div class="back-to-login">
        <router-link to="/login">返回登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'
import { verifyResetToken, resetPasswordWithToken } from '@/api/auth'

export default {
  name: 'ResetPassword',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const formRef = ref(null)
    const loading = ref(false)
    const tokenVerified = ref(false)
    const tokenError = ref('')
    const token = route.query.token

    const form = reactive({
      newPassword: '',
      confirmPassword: ''
    })

    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== form.newPassword) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    const rules = {
      newPassword: [
        { required: true, message: '请输入新密码', trigger: 'blur' },
        { min: 8, message: '密码至少8位', trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, message: '请再次输入密码', trigger: 'blur' },
        { validator: validateConfirmPassword, trigger: 'blur' }
      ]
    }

    // 验证token
    const verifyToken = async () => {
      if (!token) {
        tokenError.value = '缺少重置令牌'
        return
      }

      try {
        await verifyResetToken(token)
        tokenVerified.value = true
      } catch (error) {
        tokenError.value = error.message || '重置链接无效或已过期'
      }
    }

    const handleSubmit = async () => {
      try {
        const valid = await formRef.value.validate()
        if (!valid) return

        loading.value = true
        const data = await resetPasswordWithToken({
          token: token,
          new_password: form.newPassword
        })

        ElMessage.success(data.message || '密码重置成功')

        // 跳转到登录页
        setTimeout(() => {
          router.push('/login')
        }, 1500)
      } catch (error) {
        ElMessage.error(error.message || '重置失败，请重试')
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      verifyToken()
    })

    return {
      form,
      rules,
      formRef,
      loading,
      tokenVerified,
      tokenError,
      handleSubmit,
      Lock
    }
  }
}
</script>

<style scoped>
.reset-password-container {
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

.reset-password-card {
  width: 420px;
  padding: var(--spacing-2xl);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  position: relative;
  z-index: 1;
}

.reset-password-header {
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

.reset-password-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm);
}

.reset-password-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.reset-password-form {
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
