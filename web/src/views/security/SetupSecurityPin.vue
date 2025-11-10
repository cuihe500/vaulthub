<template>
  <div class="setup-container">
    <div class="bg-decoration bg-decoration-1"></div>
    <div class="bg-decoration bg-decoration-2"></div>

    <el-card class="setup-card">
      <div class="setup-header">
        <div class="logo-wrapper">
          <div class="logo">V</div>
        </div>
        <h2 class="setup-title">设置安全密码</h2>
        <p class="setup-subtitle">安全密码用于保护您的加密数据，请妥善保管</p>
      </div>

      <!-- 步骤1：设置安全密码 -->
      <div v-if="step === 1" class="step-content">
        <el-form :model="form" :rules="rules" ref="formRef" class="setup-form">
          <el-form-item prop="securityPin">
            <el-input
              v-model="form.securityPin"
              type="password"
              placeholder="请输入安全密码（至少8位）"
              :prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>

          <el-form-item prop="confirmPin">
            <el-input
              v-model="form.confirmPin"
              type="password"
              placeholder="请再次输入安全密码"
              :prefix-icon="Lock"
              size="large"
              show-password
              @keyup.enter="handleSubmit"
            />
          </el-form-item>

          <!-- 密码强度指示 -->
          <div v-if="form.securityPin" class="password-strength">
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
              下一步
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤2：展示助记词 -->
      <div v-else-if="step === 2" class="step-content">
        <el-alert
          title="重要提示"
          type="warning"
          :closable="false"
          show-icon
          class="alert-warning"
        >
          <p>以下24个单词是您的恢复助记词，请务必妥善保管：</p>
          <ul>
            <li>这些单词仅显示一次，请抄写在安全的地方</li>
            <li>忘记安全密码时，可使用助记词重置</li>
            <li>任何人获得助记词都可以访问您的加密数据</li>
          </ul>
        </el-alert>

        <div class="mnemonic-grid">
          <div
            v-for="(word, index) in mnemonicWords"
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
            @click="handleConfirm"
            size="large"
          >
            我已妥善保管
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
import { createEncryptionKey } from '@/api/keys'

export default {
  name: 'SetupSecurityPin',
  setup() {
    const router = useRouter()
    const formRef = ref(null)
    const loading = ref(false)
    const step = ref(1)
    const mnemonicWords = ref([])

    const form = reactive({
      securityPin: '',
      confirmPin: ''
    })

    // 表单验证规则
    const validateConfirmPin = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('请再次输入安全密码'))
      } else if (value !== form.securityPin) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    const rules = {
      securityPin: [
        { required: true, message: '请输入安全密码', trigger: 'blur' },
        { min: 8, message: '安全密码至少8位', trigger: 'blur' }
      ],
      confirmPin: [
        { required: true, validator: validateConfirmPin, trigger: 'blur' }
      ]
    }

    // 计算密码强度
    const passwordStrength = computed(() => {
      const pin = form.securityPin
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
        const response = await createEncryptionKey({
          security_pin: form.securityPin
        })

        // 提取助记词
        const mnemonic = response.user_encryption_key?.recovery_key || response.recovery_key
        if (!mnemonic) {
          throw new Error('未能获取恢复助记词')
        }

        mnemonicWords.value = mnemonic.split(' ')
        step.value = 2
        ElMessage.success('安全密码设置成功')
      } catch (error) {
        console.error('设置安全密码失败:', error)
        ElMessage.error(error.message || '设置失败，请重试')
      } finally {
        loading.value = false
      }
    }

    // 复制助记词
    const copyMnemonic = () => {
      const text = mnemonicWords.value.join(' ')
      navigator.clipboard.writeText(text).then(() => {
        ElMessage.success('助记词已复制到剪贴板')
      }).catch(() => {
        ElMessage.error('复制失败，请手动抄写')
      })
    }

    // 确认已保管
    const handleConfirm = () => {
      ElMessage.success('设置完成，即将跳转')
      setTimeout(() => {
        router.push('/')
      }, 1000)
    }

    return {
      formRef,
      loading,
      step,
      form,
      rules,
      mnemonicWords,
      passwordStrength,
      passwordStrengthWidth,
      passwordStrengthText,
      handleSubmit,
      copyMnemonic,
      handleConfirm,
      Lock
    }
  }
}
</script>

<style scoped>
.setup-container {
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

.setup-card {
  width: 100%;
  max-width: 600px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--spacing-2xl);
  position: relative;
  z-index: 1;
}

.setup-header {
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

.setup-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-sm) 0;
}

.setup-subtitle {
  color: var(--color-text-secondary);
  margin: 0;
}

.setup-form {
  margin-top: var(--spacing-xl);
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

.alert-warning {
  margin-bottom: var(--spacing-lg);
}

.alert-warning ul {
  margin: var(--spacing-sm) 0 0 0;
  padding-left: var(--spacing-lg);
}

.alert-warning li {
  margin-bottom: var(--spacing-xs);
  color: var(--color-text-secondary);
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
