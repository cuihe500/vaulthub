<template>
  <div class="reset-container">
    <div class="bg-decoration bg-decoration-1"></div>
    <div class="bg-decoration bg-decoration-2"></div>

    <el-card class="reset-card">
      <div class="reset-header">
        <div class="logo-wrapper">
          <div class="logo">V</div>
        </div>
        <h2 class="reset-title">重置安全密码</h2>
        <p class="reset-subtitle">使用恢复助记词重置您的安全密码</p>
      </div>

      <!-- 步骤1：输入助记词和新密码 -->
      <div v-if="step === 1" class="step-content">
        <el-form :model="form" :rules="rules" ref="formRef" class="reset-form">
          <el-form-item prop="recoveryMnemonic" label="恢复助记词">
            <el-input
              v-model="form.recoveryMnemonic"
              type="textarea"
              :rows="4"
              placeholder="请输入您的24个单词恢复助记词，用空格分隔"
            />
            <div class="form-hint">请输入注册时获得的24个单词，用空格分隔</div>
          </el-form-item>

          <el-form-item prop="newSecurityPin" label="新安全密码">
            <el-input
              v-model="form.newSecurityPin"
              type="password"
              placeholder="请输入新的安全密码（至少8位）"
              :prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>

          <el-form-item prop="confirmPin" label="确认新密码">
            <el-input
              v-model="form.confirmPin"
              type="password"
              placeholder="请再次输入新的安全密码"
              :prefix-icon="Lock"
              size="large"
              show-password
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <!-- 密码强度指示 -->
          <div v-if="form.newSecurityPin" class="password-strength">
            <div class="strength-label">密码强度：</div>
            <div class="strength-bar">
              <div
                class="strength-fill"
                :class="`strength-${passwordStrength}`"
                :style="{ width: passwordStrengthWidth }"
              ></div>
            </div>
            <div class="strength-text">{{ passwordStrengthText }}</div>
          </div>

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

          <div class="back-link">
            <router-link to="/login">返回登录</router-link>
          </div>
        </el-form>
      </div>

      <!-- 步骤2：展示新助记词 -->
      <div v-else-if="step === 2" class="step-content">
        <el-alert
          title="重置成功"
          type="success"
          :closable="false"
          show-icon
          class="alert-success"
        >
          <p>您的安全密码已重置成功！</p>
          <p>旧的恢复助记词已失效，请妥善保管以下新的恢复助记词：</p>
        </el-alert>

        <div class="mnemonic-grid">
          <div
            v-for="(word, index) in newMnemonicWords"
            :key="index"
            class="mnemonic-item"
          >
            <span class="mnemonic-index">{{ index + 1 }}</span>
            <span class="mnemonic-word">{{ word }}</span>
          </div>
        </div>

        <div class="action-buttons">
          <el-button @click="copyMnemonic" size="large">
            复制助记词
          </el-button>
          <el-button
            type="primary"
            @click="handleComplete"
            size="large"
          >
            完成
          </el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'
import { resetSecurityPIN } from '@/api/keys'

export default {
  name: 'ResetSecurityPin',
  setup() {
    const router = useRouter()
    const formRef = ref(null)
    const loading = ref(false)
    const step = ref(1)
    const newMnemonicWords = ref([])

    const form = reactive({
      recoveryMnemonic: '',
      newSecurityPin: '',
      confirmPin: ''
    })

    // 验证助记词
    const validateMnemonic = (rule, value, callback) => {
      if (!value) {
        callback(new Error('请输入恢复助记词'))
        return
      }
      const words = value.trim().split(/\s+/)
      if (words.length !== 24) {
        callback(new Error('恢复助记词应该是24个单词'))
      } else {
        callback()
      }
    }

    // 验证确认密码
    const validateConfirmPin = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请再次输入新密码'))
      } else if (value !== form.newSecurityPin) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    const rules = {
      recoveryMnemonic: [
        { required: true, validator: validateMnemonic, trigger: 'blur' }
      ],
      newSecurityPin: [
        { required: true, message: '请输入新的安全密码', trigger: 'blur' },
        { min: 8, message: '安全密码至少8位', trigger: 'blur' }
      ],
      confirmPin: [
        { required: true, validator: validateConfirmPin, trigger: 'blur' }
      ]
    }

    // 计算密码强度
    const passwordStrength = computed(() => {
      const pin = form.newSecurityPin
      if (!pin) return 'weak'

      let strength = 0
      if (pin.length >= 8) strength++
      if (pin.length >= 12) strength++
      if (/[a-z]/.test(pin) && /[A-Z]/.test(pin)) strength++
      if (/\d/.test(pin)) strength++
      if (/[^a-zA-Z0-9]/.test(pin)) strength++

      if (strength >= 4) return 'strong'
      if (strength >= 2) return 'medium'
      return 'weak'
    })

    const passwordStrengthWidth = computed(() => {
      const strength = passwordStrength.value
      if (strength === 'strong') return '100%'
      if (strength === 'medium') return '66%'
      return '33%'
    })

    const passwordStrengthText = computed(() => {
      const strength = passwordStrength.value
      if (strength === 'strong') return '强'
      if (strength === 'medium') return '中'
      return '弱'
    })

    // 提交表单
    const handleSubmit = async () => {
      try {
        const valid = await formRef.value.validate()
        if (!valid) return

        loading.value = true
        const response = await resetSecurityPIN({
          recovery_mnemonic: form.recoveryMnemonic.trim(),
          new_security_pin: form.newSecurityPin
        })

        // 提取新助记词
        const mnemonic = response.new_recovery_mnemonic
        if (!mnemonic) {
          throw new Error('未能获取新的恢复助记词')
        }

        newMnemonicWords.value = mnemonic.split(' ')
        step.value = 2
        ElMessage.success('安全密码重置成功')
      } catch (error) {
        console.error('重置安全密码失败:', error)
        ElMessage.error(error.message || '重置失败，请检查助记词是否正确')
      } finally {
        loading.value = false
      }
    }

    // 复制助记词
    const copyMnemonic = () => {
      const text = newMnemonicWords.value.join(' ')
      navigator.clipboard.writeText(text).then(() => {
        ElMessage.success('助记词已复制到剪贴板')
      }).catch(() => {
        ElMessage.error('复制失败，请手动抄写')
      })
    }

    // 完成
    const handleComplete = () => {
      ElMessage.success('重置完成，即将跳转到登录页')
      setTimeout(() => {
        router.push('/login')
      }, 1000)
    }

    return {
      formRef,
      loading,
      step,
      form,
      rules,
      newMnemonicWords,
      passwordStrength,
      passwordStrengthWidth,
      passwordStrengthText,
      handleSubmit,
      copyMnemonic,
      handleComplete,
      Lock
    }
  }
}
</script>

<style scoped>
.reset-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  padding: var(--spacing-lg);
  position: relative;
  overflow: hidden;
}

.bg-decoration {
  position: absolute;
  border-radius: var(--radius-full);
  opacity: 0.1;
  background: var(--color-white);
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
  right: 100px;
}

.reset-card {
  width: 100%;
  max-width: 600px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--spacing-2xl);
  position: relative;
  z-index: 1;
}

.reset-header {
  text-align: center;
  margin-bottom: var(--spacing-2xl);
}

.logo-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: var(--spacing-md);
}

.logo {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-white);
}

.reset-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm) 0;
}

.reset-subtitle {
  color: var(--color-text-secondary);
  margin: 0;
}

.reset-form {
  margin-top: var(--spacing-xl);
}

.form-hint {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin-top: var(--spacing-xs);
}

.password-strength {
  margin-top: calc(-1 * var(--spacing-sm));
  margin-bottom: var(--spacing-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-sm);
}

.strength-label {
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.strength-bar {
  flex: 1;
  height: 4px;
  background: var(--color-bg);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: all var(--duration-base) var(--easing);
}

.strength-fill.strength-weak {
  background: var(--color-error);
}

.strength-fill.strength-medium {
  background: var(--color-warning);
}

.strength-fill.strength-strong {
  background: var(--color-success);
}

.strength-text {
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.submit-button-item {
  margin-top: var(--spacing-xl);
}

.submit-button {
  width: 100%;
}

.back-link {
  text-align: center;
  margin-top: var(--spacing-md);
}

.back-link a {
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--duration-fast) var(--easing);
}

.back-link a:hover {
  color: var(--color-primary-dark);
}

.alert-success {
  margin-bottom: var(--spacing-lg);
}

.alert-success p {
  margin: var(--spacing-xs) 0;
}

.mnemonic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
  padding: var(--spacing-lg);
  background: var(--color-bg);
  border-radius: var(--radius-md);
}

.mnemonic-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--color-white);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
}

.mnemonic-index {
  font-size: var(--font-size-xs);
  color: var(--color-text-disabled);
  min-width: 20px;
}

.mnemonic-word {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  font-family: var(--font-family-mono);
  color: var(--color-text-primary);
}

.action-buttons {
  display: flex;
  gap: var(--spacing-md);
  justify-content: center;
}

.action-buttons .el-button {
  flex: 1;
  max-width: 200px;
}
</style>
