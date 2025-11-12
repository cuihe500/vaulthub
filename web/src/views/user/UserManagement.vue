<template>
  <div class="user-management">
    <!-- 仪表盘部分 - 图表统计 -->
    <div class="dashboard-section">
      <h2 class="section-title">数据统计</h2>

      <div class="charts-grid">
        <!-- 密钥类型分布环形图 -->
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">密钥类型分布</span>
              <span class="total-count">总计: {{ totalSecrets }}</span>
            </div>
          </template>
          <div
            ref="keyTypeChartRef"
            class="chart-container"
            v-loading="chartLoading"
          ></div>
        </el-card>

        <!-- 最近24小时操作统计饼状图 -->
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span class="card-title">最近24小时操作统计</span>
              <span class="total-count">总计: {{ todayOperations }}</span>
            </div>
          </template>
          <div
            ref="operationChartRef"
            class="chart-container"
            v-loading="chartLoading"
          ></div>
        </el-card>
      </div>
    </div>

    <!-- 用户信息部分 -->
    <div class="info-section">
      <h2 class="section-title">基本信息</h2>
      <el-card>
        <!-- 用户档案不存在时显示提示 -->
        <div v-if="!profileExists" class="profile-not-exist">
          <el-empty description="您还没有创建用户档案">
            <el-button type="primary" @click="showCreateProfile = true">
              创建用户档案
            </el-button>
          </el-empty>
        </div>

        <!-- 用户档案存在时显示详细信息 -->
        <template v-else>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="用户名">
              {{ userInfo.username || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="昵称">
              {{ userProfile.nickname || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="邮箱">
              {{ userProfile.email || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="手机">
              {{ userProfile.phone || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="角色">
              <el-tag :type="getRoleType(userInfo.role)">
                {{ getRoleText(userInfo.role) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="加入时间">
              {{ formatDate(userInfo.created_at) }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="edit-actions">
            <el-button type="primary" @click="showEditProfile = true">
              编辑资料
            </el-button>
          </div>
        </template>
      </el-card>
    </div>

    <!-- 安全管理部分 -->
    <div class="security-section">
      <h2 class="section-title">安全管理</h2>
      <el-card>
        <div class="security-item">
          <div class="security-info">
            <h3>登录密码</h3>
            <p>用于登录系统的密码</p>
          </div>
          <el-button @click="showChangePassword = true">修改密码</el-button>
        </div>

        <el-divider />

        <div class="security-item">
          <div class="security-info">
            <h3>安全密钥</h3>
            <p>用于加密解密数据的密钥，修改后需要使用恢复密钥</p>
          </div>
          <el-button @click="showResetSecurityPin = true">重置安全密钥</el-button>
        </div>
      </el-card>
    </div>

    <!-- 创建用户档案弹窗 -->
    <el-dialog
      title="创建用户档案"
      v-model="showCreateProfile"
      width="500px"
    >
      <el-form :model="profileForm" :rules="profileRules" ref="createProfileFormRef" label-width="80px">
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="profileForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="profileForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机" prop="phone">
          <el-input v-model="profileForm.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateProfile = false">取消</el-button>
        <el-button type="primary" @click="handleCreateProfile" :loading="submitting">
          创建
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑资料弹窗 -->
    <el-dialog
      title="编辑资料"
      v-model="showEditProfile"
      width="500px"
    >
      <el-form :model="profileForm" :rules="profileRules" ref="profileFormRef" label-width="80px">
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="profileForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="profileForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机" prop="phone">
          <el-input v-model="profileForm.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditProfile = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateProfile" :loading="submitting">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改密码弹窗 -->
    <el-dialog
      title="修改密码"
      v-model="showChangePassword"
      width="500px"
    >
      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input
            v-model="passwordForm.oldPassword"
            type="password"
            placeholder="请输入旧密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showChangePassword = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置安全密钥弹窗 -->
    <el-dialog
      title="重置安全密钥"
      v-model="showResetSecurityPin"
      width="600px"
    >
      <el-alert
        type="warning"
        :closable="false"
        style="margin-bottom: 20px"
      >
        <template #title>
          <strong>警告</strong>：重置安全密钥需要提供恢复密钥（24个单词）
        </template>
      </el-alert>
      <el-form :model="securityPinForm" :rules="securityPinRules" ref="securityPinFormRef" label-width="120px">
        <el-form-item label="恢复密钥" prop="recoveryMnemonic">
          <el-input
            v-model="securityPinForm.recoveryMnemonic"
            type="textarea"
            :rows="4"
            placeholder="请输入恢复密钥（24个单词，空格分隔）"
          />
        </el-form-item>
        <el-form-item label="新安全密钥" prop="newSecurityPin">
          <el-input
            v-model="securityPinForm.newSecurityPin"
            type="password"
            placeholder="请输入新的安全密钥"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密钥" prop="confirmSecurityPin">
          <el-input
            v-model="securityPinForm.confirmSecurityPin"
            type="password"
            placeholder="请再次输入新的安全密钥"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetSecurityPin = false">取消</el-button>
        <el-button type="primary" @click="handleResetSecurityPin" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getCurrentUser } from '@/api/auth'
import { getCurrentProfile, createProfile, updateProfile } from '@/api/profile'
import { getCurrentStatistics } from '@/api/statistics'
import { exportOperationStatistics } from '@/api/audit'
import { toRFC3339 } from '@/utils/date'
import { ElMessage } from 'element-plus'

// ECharts按需异步加载,只引入需要的图表类型(PieChart)和组件,大幅减少bundle大小
let echartsCore = null
const loadECharts = async () => {
  if (echartsCore) return echartsCore

  // 只引入核心和必需的组件,而不是整个echarts库
  const [{ init, use }, { PieChart }, { TitleComponent, TooltipComponent, LegendComponent }] =
    await Promise.all([
      import('echarts/core'),
      import('echarts/charts'),
      import('echarts/components')
    ])

  // 注册需要的组件
  use([PieChart, TitleComponent, TooltipComponent, LegendComponent])

  echartsCore = { init }
  return echartsCore
}

export default {
  name: 'UserManagement',
  data() {
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.passwordForm.newPassword) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    const validateConfirmSecurityPin = (rule, value, callback) => {
      if (value !== this.securityPinForm.newSecurityPin) {
        callback(new Error('两次输入的安全密钥不一致'))
      } else {
        callback()
      }
    }

    return {
      loading: false,
      submitting: false,
      chartLoading: false,

      // 图表实例
      keyTypeChart: null,
      operationChart: null,

      // 统计数据
      statistics: {},
      totalSecrets: 0,
      todayOperations: 0,

      userInfo: {},
      userProfile: {},
      profileExists: false,

      showCreateProfile: false,
      showEditProfile: false,
      showChangePassword: false,
      showResetSecurityPin: false,

      profileForm: {
        nickname: '',
        email: '',
        phone: ''
      },

      profileRules: {
        nickname: [
          { required: true, message: '请输入昵称', trigger: 'blur' }
        ],
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
        ],
        phone: [
          { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
        ]
      },

      passwordForm: {
        oldPassword: '',
        newPassword: '',
        confirmPassword: ''
      },

      passwordRules: {
        oldPassword: [
          { required: true, message: '请输入旧密码', trigger: 'blur' }
        ],
        newPassword: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 8, message: '密码长度不能少于8位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请再次输入新密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ]
      },

      securityPinForm: {
        recoveryMnemonic: '',
        newSecurityPin: '',
        confirmSecurityPin: ''
      },

      securityPinRules: {
        recoveryMnemonic: [
          { required: true, message: '请输入恢复密钥', trigger: 'blur' }
        ],
        newSecurityPin: [
          { required: true, message: '请输入新的安全密钥', trigger: 'blur' },
          { min: 8, message: '安全密钥长度不能少于8位', trigger: 'blur' }
        ],
        confirmSecurityPin: [
          { required: true, message: '请再次输入新的安全密钥', trigger: 'blur' },
          { validator: validateConfirmSecurityPin, trigger: 'blur' }
        ]
      }
    }
  },

  mounted() {
    this.loadData()
    this.initCharts()
    window.addEventListener('resize', this.handleResize)
  },

  beforeUnmount() {
    this.destroyCharts()
    window.removeEventListener('resize', this.handleResize)
  },

  methods: {
    async loadData() {
      this.loading = true
      try {
        await Promise.all([
          this.loadUserInfo(),
          this.loadUserProfile(),
          this.loadStatistics()
        ])
      } finally {
        this.loading = false
      }
    },

    async loadUserInfo() {
      try {
        this.userInfo = await getCurrentUser()
      } catch (error) {
        console.error('加载用户信息失败:', error)
      }
    },

    async loadUserProfile() {
      try {
        this.userProfile = await getCurrentProfile()
        this.profileExists = true
      } catch (error) {
        console.error('加载用户档案失败:', error)
        this.userProfile = {}
        this.profileExists = false
      }
    },

    async loadStatistics() {
      try {
        this.statistics = await getCurrentStatistics()
        await this.loadChartData()
      } catch (error) {
        console.error('加载统计数据失败:', error)
        this.statistics = {}
      }
    },

    // 初始化图表
    initCharts() {
      this.$nextTick(() => {
        this.initKeyTypeChart()
        this.initOperationChart()
      })
    },

    // 初始化密钥类型环形图
    async initKeyTypeChart() {
      if (!this.$refs.keyTypeChartRef) return

      const { init } = await loadECharts()
      this.keyTypeChart = init(this.$refs.keyTypeChartRef)

      const option = {
        tooltip: {
          trigger: 'item',
          formatter: '{b}: {c} ({d}%)'
        },
        legend: {
          orient: 'vertical',
          right: 'right',
          top: 'center',
          textStyle: {
            color: '#1f2937'
          }
        },
        series: [
          {
            name: '密钥类型',
            type: 'pie',
            radius: ['45%', '70%'],
            center: ['40%', '50%'],
            avoidLabelOverlap: true,
            itemStyle: {
              borderRadius: 8,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: false
            },
            emphasis: {
              label: {
                show: true,
                fontSize: 16,
                fontWeight: 'bold'
              },
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.3)'
              }
            },
            data: []
          }
        ],
        color: ['#667eea', '#764ba2', '#f59e0b', '#10b981', '#3b82f6', '#6b7280']
      }

      this.keyTypeChart.setOption(option)
    },

    // 初始化今日操作饼状图
    async initOperationChart() {
      if (!this.$refs.operationChartRef) return

      const { init } = await loadECharts()
      this.operationChart = init(this.$refs.operationChartRef)

      const option = {
        tooltip: {
          trigger: 'item',
          formatter: '{b}: {c} ({d}%)'
        },
        legend: {
          orient: 'vertical',
          right: 'right',
          top: 'center',
          textStyle: {
            color: '#1f2937'
          }
        },
        series: [
          {
            name: '操作类型',
            type: 'pie',
            radius: ['45%', '70%'],
            center: ['40%', '50%'],
            avoidLabelOverlap: true,
            itemStyle: {
              borderRadius: 8,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: false
            },
            emphasis: {
              label: {
                show: true,
                fontSize: 16,
                fontWeight: 'bold'
              },
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.3)'
              }
            },
            data: []
          }
        ],
        color: ['#10b981', '#3b82f6', '#f59e0b', '#ef4444']
      }

      this.operationChart.setOption(option)
    },

    // 加载图表数据
    async loadChartData() {
      this.chartLoading = true
      try {
        await Promise.all([
          this.loadKeyTypeData(),
          this.loadOperationData()
        ])
      } finally {
        this.chartLoading = false
      }
    },

    // 加载密钥类型数据
    async loadKeyTypeData() {
      try {
        // 从getCurrentStatistics获取的密钥统计数据
        const KEY_TYPE_LABELS = {
          api_key_count: 'API密钥',
          ssh_key_count: 'SSH密钥',
          private_key_count: '私钥/凭证',
          certificate_count: '证书',
          password_count: '密码',
          other_count: '其他'
        }

        const chartData = []
        let total = 0

        // 从statistics对象中读取各类型密钥数量
        Object.keys(KEY_TYPE_LABELS).forEach((key) => {
          const value = this.statistics[key] || 0
          if (value > 0) {
            chartData.push({
              name: KEY_TYPE_LABELS[key],
              value: value
            })
            total += value
          }
        })

        this.totalSecrets = total

        if (this.keyTypeChart) {
          this.keyTypeChart.setOption({
            series: [
              {
                data: chartData
              }
            ]
          })
        }
      } catch (error) {
        console.error('加载密钥类型数据失败:', error)
      }
    },

    // 加载今日操作数据
    async loadOperationData() {
      try {
        // 计算最近24小时的时间范围（当前时刻-24小时 到 当前时刻）
        const now = new Date()
        const twentyFourHoursAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        const endTime = toRFC3339(now)
        const startTime = toRFC3339(twentyFourHoursAgo)

        // 调用操作统计导出接口
        const operationStats = await exportOperationStatistics({
          start_time: startTime,
          end_time: endTime
        })

        // 操作类型标签映射
        const OPERATION_TYPE_LABELS = {
          CREATE: '创建',
          UPDATE: '更新',
          DELETE: '删除',
          ACCESS: '访问',
          LOGIN: '登录',
          LOGOUT: '登出'
        }

        const chartData = []
        let total = 0

        // 从by_action对象中读取操作数据
        if (operationStats && operationStats.by_action) {
          Object.keys(OPERATION_TYPE_LABELS).forEach((key) => {
            const value = operationStats.by_action[key] || 0
            if (value > 0) {
              chartData.push({
                name: OPERATION_TYPE_LABELS[key],
                value: value
              })
              total += value
            }
          })
        }

        this.todayOperations = total

        if (this.operationChart) {
          this.operationChart.setOption({
            series: [
              {
                data: chartData
              }
            ]
          })
        }
      } catch (error) {
        console.error('加载今日操作数据失败:', error)
      }
    },

    // 销毁图表
    destroyCharts() {
      if (this.keyTypeChart) {
        this.keyTypeChart.dispose()
        this.keyTypeChart = null
      }
      if (this.operationChart) {
        this.operationChart.dispose()
        this.operationChart = null
      }
    },

    // 处理窗口大小变化
    handleResize() {
      this.keyTypeChart?.resize()
      this.operationChart?.resize()
    },

    getRoleType(role) {
      const roleMap = {
        admin: 'danger',
        user: 'success',
        readonly: 'info'
      }
      return roleMap[role] || 'info'
    },

    getRoleText(role) {
      const roleMap = {
        admin: '管理员',
        user: '普通用户',
        readonly: '只读用户'
      }
      return roleMap[role] || role
    },

    formatDate(dateString) {
      if (!dateString) return '-'
      const date = new Date(dateString)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      })
    },

    handleCreateProfile() {
      this.$refs.createProfileFormRef.validate(async (valid) => {
        if (!valid) return

        this.submitting = true
        try {
          await createProfile(this.profileForm)
          ElMessage.success('用户档案创建成功')
          this.showCreateProfile = false
          await this.loadUserProfile()
        } catch (error) {
          console.error('创建用户档案失败:', error)
        } finally {
          this.submitting = false
        }
      })
    },

    handleUpdateProfile() {
      this.$refs.profileFormRef.validate(async (valid) => {
        if (!valid) return

        this.submitting = true
        try {
          await updateProfile(this.profileForm)
          ElMessage.success('资料更新成功')
          this.showEditProfile = false
          await this.loadUserProfile()
        } catch (error) {
          console.error('更新资料失败:', error)
        } finally {
          this.submitting = false
        }
      })
    },

    handleChangePassword() {
      this.$refs.passwordFormRef.validate(async (valid) => {
        if (!valid) return

        ElMessage.warning('修改登录密码功能暂未实现，请联系管理员')
        this.showChangePassword = false
      })
    },

    handleResetSecurityPin() {
      this.$refs.securityPinFormRef.validate(async (valid) => {
        if (!valid) return

        ElMessage.warning('重置安全密钥功能暂未实现')
        this.showResetSecurityPin = false
      })
    }
  },

  watch: {
    showEditProfile(val) {
      if (val) {
        this.profileForm = {
          nickname: this.userProfile.nickname || '',
          email: this.userProfile.email || '',
          phone: this.userProfile.phone || ''
        }
      }
    }
  }
}
</script>

<style scoped>
.user-management {
  padding: var(--spacing-lg);
  max-width: 1200px;
  margin: 0 auto;
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-md);
}

.dashboard-section {
  margin-bottom: var(--spacing-xl);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-lg);
}

@media (max-width: 1024px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
}

.chart-card {
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--duration-fast) var(--easing);
}

.chart-card:hover {
  box-shadow: var(--shadow-md);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.total-count {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  font-weight: var(--font-weight-medium);
}

.chart-container {
  width: 100%;
  height: 200px;
  min-height: 200px;
}

.info-section,
.security-section {
  margin-bottom: var(--spacing-xl);
}

.edit-actions {
  margin-top: var(--spacing-lg);
  text-align: right;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) 0;
}

.security-info h3 {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-xs) 0;
}

.security-info p {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.profile-not-exist {
  padding: var(--spacing-xl) 0;
}
</style>
