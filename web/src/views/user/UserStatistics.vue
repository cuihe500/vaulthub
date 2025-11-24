<template>
  <div class="user-statistics">
    <div class="header">
      <h2>我的统计</h2>
      <div class="controls">
        <!-- 时间范围类型切换 -->
        <el-radio-group v-model="rangeType" size="default" @change="handleRangeTypeChange">
          <el-radio-button label="preset">预设范围</el-radio-button>
          <el-radio-button label="custom">自定义</el-radio-button>
        </el-radio-group>

        <!-- 预设范围选择 -->
        <el-select
          v-if="rangeType === 'preset'"
          v-model="presetRange"
          size="default"
          style="width: 150px; margin-left: 12px"
          @change="handlePresetRangeChange"
        >
          <el-option :value="7" label="最近7天" />
          <el-option :value="30" label="最近30天" />
          <el-option :value="90" label="最近90天" />
          <el-option :value="365" label="最近一年" />
        </el-select>

        <!-- 自定义范围选择 -->
        <el-date-picker
          v-else
          v-model="customRange"
          type="daterange"
          range-separator="-"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          size="default"
          style="margin-left: 12px"
          @change="handleCustomRangeChange"
        />

        <!-- 刷新按钮 -->
        <el-button
          type="primary"
          :icon="RefreshIcon"
          size="default"
          style="margin-left: 12px"
          @click="handleRefresh"
          :loading="loading"
        >
          刷新
        </el-button>
      </div>
    </div>

    <!-- 数据说明提示 -->
    <el-alert
      title="数据说明"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: var(--spacing-lg)"
    >
      <template #default>
        本页面展示的是历史统计数据，数据按日汇总并存储，非实时数据。统计数据每日定时更新，反映您密钥管理的历史趋势。
      </template>
    </el-alert>

    <el-row :gutter="20" v-loading="loading">
      <!-- 趋势图 -->
      <el-col :span="24">
        <el-card>
          <template #header>密钥数量趋势</template>
          <trend-chart v-if="statisticsData.length > 0" :data="statisticsData" />
          <el-empty v-else description="暂无数据" />
        </el-card>
      </el-col>

      <!-- 类型饼图 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>密钥类型分布（最新）</span>
              <span v-if="latestData" class="card-subtitle">
                {{ formatDate(latestData.stat_date) }}
              </span>
            </div>
          </template>
          <type-pie-chart v-if="latestData && secretTypeDistribution.length > 0" :distribution="secretTypeDistribution" />
          <el-empty v-else description="暂无数据" />
        </el-card>
      </el-col>

      <!-- 操作活跃度热力图 -->
      <el-col :span="12">
        <el-card>
          <template #header>操作活跃度</template>
          <activity-heatmap v-if="statisticsData.length > 0" :data="statisticsData" />
          <el-empty v-else description="暂无数据" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { getUserStatistics } from '@/api/statistics'
import { Refresh as RefreshIcon } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import TrendChart from '@/components/business/TrendChart.vue'
import TypePieChart from '@/components/business/TypePieChart.vue'
import ActivityHeatmap from '@/components/business/ActivityHeatmap.vue'

export default {
  name: 'UserStatistics',
  components: { TrendChart, TypePieChart, ActivityHeatmap },

  data() {
    return {
      RefreshIcon,
      rangeType: 'preset', // preset | custom
      presetRange: 30, // 预设天数
      customRange: null, // 自定义日期范围 [startDate, endDate]
      statisticsData: [],
      loading: false,
      cache: new Map(), // 带失效时间的缓存
      cacheTTL: 5 * 60 * 1000 // 5分钟过期
    }
  },

  computed: {
    // 计算查询参数
    queryParams() {
      let startDate, endDate

      if (this.rangeType === 'preset') {
        endDate = dayjs().format('YYYY-MM-DD')
        startDate = dayjs().subtract(this.presetRange, 'day').format('YYYY-MM-DD')
      } else {
        if (!this.customRange || this.customRange.length !== 2) {
          return null
        }
        startDate = dayjs(this.customRange[0]).format('YYYY-MM-DD')
        endDate = dayjs(this.customRange[1]).format('YYYY-MM-DD')
      }

      return {
        stat_type: 'daily',
        start_date: startDate,
        end_date: endDate
      }
    },

    // 缓存键
    cacheKey() {
      if (!this.queryParams) return null
      return `${this.queryParams.start_date}_${this.queryParams.end_date}`
    },

    // 最新一天的数据（用于饼图）
    latestData() {
      return this.statisticsData.length > 0 ? this.statisticsData[0] : null
    },

    // 密钥类型分布（用于饼图）
    secretTypeDistribution() {
      if (!this.latestData) return []
      return [
        { name: 'API密钥', value: this.latestData.api_key_count },
        { name: '密码', value: this.latestData.password_count },
        { name: '证书', value: this.latestData.certificate_count },
        { name: 'SSH密钥', value: this.latestData.ssh_key_count },
        { name: '数据库凭证', value: this.latestData.database_count },
        { name: 'Token', value: this.latestData.token_count },
        { name: '其他', value: this.latestData.other_count }
      ].filter((item) => item.value > 0)
    }
  },

  mounted() {
    this.fetchStatistics()
  },

  methods: {
    // 获取统计数据（带缓存）
    async fetchStatistics(forceRefresh = false) {
      if (!this.queryParams) {
        this.$message.warning('请选择日期范围')
        return
      }

      const key = this.cacheKey
      const cached = this.cache.get(key)

      // 检查缓存是否有效
      if (!forceRefresh && cached && Date.now() - cached.timestamp < this.cacheTTL) {
        this.statisticsData = cached.data
        return
      }

      // 请求新数据
      this.loading = true
      try {
        const result = await getUserStatistics(this.queryParams)
        this.statisticsData = result || []
        this.cache.set(key, { data: result || [], timestamp: Date.now() })

        if (this.statisticsData.length === 0) {
          this.$message.info('该时间范围内暂无统计数据')
        }
      } catch (error) {
        this.$message.error('加载统计数据失败')
        this.statisticsData = []
      } finally {
        this.loading = false
      }
    },

    // 强制刷新
    handleRefresh() {
      this.fetchStatistics(true)
    },

    // 时间范围类型切换
    handleRangeTypeChange() {
      if (this.rangeType === 'custom') {
        // 初始化自定义范围为最近30天
        this.customRange = [dayjs().subtract(30, 'day').toDate(), dayjs().toDate()]
      }
      this.fetchStatistics()
    },

    // 预设范围切换
    handlePresetRangeChange() {
      this.fetchStatistics()
    },

    // 自定义范围切换
    handleCustomRangeChange() {
      this.fetchStatistics()
    },

    // 格式化日期
    formatDate(date) {
      return dayjs(date).format('YYYY-MM-DD')
    }
  }
}
</script>

<style scoped>
.user-statistics {
  padding: var(--spacing-lg);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.controls {
  display: flex;
  align-items: center;
}

.el-col {
  margin-bottom: var(--spacing-lg);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-subtitle {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
