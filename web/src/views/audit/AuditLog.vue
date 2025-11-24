<template>
  <div class="audit-log">
    <!-- 顶部操作栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="page-title">审计日志</h2>
      </div>
    </div>

    <!-- 搜索过滤条件 -->
    <div class="filter-bar">
      <el-form :inline="true" :model="queryParams" class="filter-form">
        <el-form-item label="操作类型" class="filter-item action-type">
          <el-select
            v-model="queryParams.action_type"
            placeholder="全部"
            clearable
            class="filter-select"
          >
            <el-option label="创建" value="CREATE" />
            <el-option label="更新" value="UPDATE" />
            <el-option label="删除" value="DELETE" />
            <el-option label="访问" value="ACCESS" />
            <el-option label="登录" value="LOGIN" />
            <el-option label="退出" value="LOGOUT" />
          </el-select>
        </el-form-item>

        <el-form-item label="资源路径" class="filter-item resource-path">
          <el-input
            v-model="queryParams.resource_type"
            placeholder="如: /api/v1/secrets"
            clearable
            class="filter-input"
          />
        </el-form-item>

        <el-form-item label="状态" class="filter-item status">
          <el-select
            v-model="queryParams.status"
            placeholder="全部"
            clearable
            class="filter-select"
          >
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>

        <el-form-item label="时间范围" class="filter-item date-range">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            :shortcuts="dateShortcuts"
            class="filter-range"
          />
        </el-form-item>

        <el-form-item class="filter-item filter-actions">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 审计日志列表表格 -->
    <el-card class="table-card" shadow="never">
      <div class="table-wrapper">
        <el-table
          v-loading="loading"
          :data="auditLogs"
          stripe
          style="width: 100%"
        >
          <el-table-column prop="username" label="用户" width="120" />
          <el-table-column prop="action_type" label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action_type)">
              {{ getActionText(row.action_type) }}
            </el-tag>
          </template>
        </el-table-column>
          <el-table-column
            prop="resource_type"
            label="资源路径"
            min-width="220"
            show-overflow-tooltip
          />
          <el-table-column
            prop="resource_name"
            label="资源名称"
            min-width="160"
            show-overflow-tooltip
          />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="140" />
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link @click="handleViewDetail(row)">
                详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="showDetail"
      title="审计日志详情"
      width="700px"
      destroy-on-close
    >
      <el-descriptions :column="1" border v-if="currentLog">
        <el-descriptions-item label="日志ID">
          {{ currentLog.uuid }}
        </el-descriptions-item>
        <el-descriptions-item label="用户">
          {{ currentLog.username }}
        </el-descriptions-item>
        <el-descriptions-item label="用户UUID">
          {{ currentLog.user_uuid }}
        </el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getActionType(currentLog.action_type)">
            {{ getActionText(currentLog.action_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源路径">
          {{ currentLog.resource_type }}
        </el-descriptions-item>
        <el-descriptions-item
          label="资源UUID"
          v-if="currentLog.resource_uuid"
        >
          {{ currentLog.resource_uuid }}
        </el-descriptions-item>
        <el-descriptions-item
          label="资源名称"
          v-if="currentLog.resource_name"
        >
          {{ currentLog.resource_name }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag
            :type="currentLog.status === 'success' ? 'success' : 'danger'"
          >
            {{ currentLog.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item
          label="错误代码"
          v-if="currentLog.error_code"
        >
          {{ currentLog.error_code }}
        </el-descriptions-item>
        <el-descriptions-item
          label="错误信息"
          v-if="currentLog.error_message"
        >
          {{ currentLog.error_message }}
        </el-descriptions-item>
        <el-descriptions-item label="IP地址" v-if="currentLog.ip_address">
          {{ currentLog.ip_address }}
        </el-descriptions-item>
        <el-descriptions-item
          label="User Agent"
          v-if="currentLog.user_agent"
        >
          {{ currentLog.user_agent }}
        </el-descriptions-item>
        <el-descriptions-item label="请求ID" v-if="currentLog.request_id">
          {{ currentLog.request_id }}
        </el-descriptions-item>
        <el-descriptions-item label="操作时间">
          {{ formatDate(currentLog.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="详细信息" v-if="currentLog.details">
          <pre class="details-json">{{ formatJSON(currentLog.details) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { queryAuditLogs } from '@/api/audit'
import dayjs from 'dayjs'

// 查询参数
const queryParams = reactive({
  action_type: '',
  resource_type: '',
  status: '',
  start_time: '',
  end_time: '',
  page: 1,
  page_size: 20
})

// 日期范围
const dateRange = ref([])

// 日期快捷选项
const dateShortcuts = [
  {
    text: '最近1小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000)
      return [start, end]
    }
  },
  {
    text: '最近24小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24)
      return [start, end]
    }
  },
  {
    text: '最近7天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 7)
      return [start, end]
    }
  },
  {
    text: '最近30天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24 * 30)
      return [start, end]
    }
  }
]

// 审计日志列表
const auditLogs = ref([])
const total = ref(0)
const loading = ref(false)

// 详情对话框
const showDetail = ref(false)
const currentLog = ref(null)

// 获取审计日志列表
const fetchAuditLogs = async () => {
  loading.value = true
  try {
    // 处理时间范围
    if (dateRange.value && dateRange.value.length === 2) {
      queryParams.start_time = dateRange.value[0]
      queryParams.end_time = dateRange.value[1]
    } else {
      queryParams.start_time = ''
      queryParams.end_time = ''
    }

    const result = await queryAuditLogs(queryParams)

    // 确保正确提取数据结构
    if (result && result.logs) {
      auditLogs.value = result.logs
      total.value = result.total || 0
    } else {
      auditLogs.value = []
      total.value = 0
    }
  } catch (error) {
    console.error('获取审计日志失败:', error)
    ElMessage.error('获取审计日志失败')
    auditLogs.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  queryParams.page = 1
  fetchAuditLogs()
}

// 重置
const handleReset = () => {
  queryParams.action_type = ''
  queryParams.resource_type = ''
  queryParams.status = ''
  queryParams.page = 1
  queryParams.page_size = 20
  dateRange.value = []
  fetchAuditLogs()
}

// 分页变化
const handlePageChange = (page) => {
  queryParams.page = page
  fetchAuditLogs()
}

// 每页数量变化
const handleSizeChange = (size) => {
  queryParams.page_size = size
  queryParams.page = 1
  fetchAuditLogs()
}

// 查看详情
const handleViewDetail = (row) => {
  currentLog.value = row
  showDetail.value = true
}

// 获取操作类型标签
const getActionType = (action) => {
  const typeMap = {
    CREATE: 'success',
    UPDATE: 'warning',
    DELETE: 'danger',
    ACCESS: 'info',
    LOGIN: 'success',
    LOGOUT: 'info'
  }
  return typeMap[action] || 'info'
}

// 获取操作文本
const getActionText = (action) => {
  const textMap = {
    CREATE: '创建',
    UPDATE: '更新',
    DELETE: '删除',
    ACCESS: '访问',
    LOGIN: '登录',
    LOGOUT: '退出'
  }
  return textMap[action] || action
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// 格式化JSON
const formatJSON = (data) => {
  if (!data) return ''
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

// 组件挂载时加载数据
onMounted(() => {
  fetchAuditLogs()
})
</script>

<style scoped>
.audit-log {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* 顶部操作栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.page-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0;
}

/* 筛选栏 */
.filter-bar {
  margin-bottom: var(--spacing-md);
  overflow-x: auto;
}

.filter-form {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--spacing-sm);
  align-items: flex-end;
}

.filter-item {
  margin-bottom: 0;
}

.filter-select,
.filter-input,
.filter-range {
  width: 100%;
}

.filter-actions {
  display: flex;
  gap: var(--spacing-sm);
  justify-content: flex-end;
}

/* 各筛选项在一行中的宽度比例设置 */
.action-type,
.status {
  flex: 0 0 140px;
  min-width: 120px;
}

.resource-path {
  flex: 1 1 260px;
  min-width: 180px;
}

.date-range {
  flex: 0 0 320px;
  min-width: 240px;
}

/* 表格卡片 */
.table-card {
  border-radius: var(--radius-md);
  flex: 1;
  min-height: 0;
}

.table-card :deep(.el-card__body) {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrapper {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

/* 分页 */
.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

.details-json {
  background-color: var(--color-bg);
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  overflow-x: auto;
  max-height: 300px;
  margin: 0;
}

/* 小屏下表单和表格的自适应布局 */
@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
}
</style>
