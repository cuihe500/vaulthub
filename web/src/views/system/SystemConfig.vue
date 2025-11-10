<template>
  <div class="system-config-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2 class="page-title">系统配置</h2>
      <p class="page-subtitle">管理系统运行参数，配置更新后立即生效</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button
        type="primary"
        @click="handleReload"
        :loading="reloading"
      >
        重新加载配置
      </el-button>
    </div>

    <!-- 配置列表 -->
    <el-card class="config-card">
      <el-table
        :data="configs"
        v-loading="loading"
        stripe
        class="config-table"
      >
        <el-table-column
          prop="config_key"
          label="配置键"
          min-width="200"
        >
          <template #default="{ row }">
            <span class="config-key">{{ row.config_key }}</span>
          </template>
        </el-table-column>

        <el-table-column
          prop="config_value"
          label="配置值"
          min-width="200"
        >
          <template #default="{ row }">
            <el-input
              v-if="editingKey === row.config_key"
              v-model="editingValue"
              placeholder="请输入配置值"
              size="small"
            />
            <span v-else class="config-value">{{ row.config_value }}</span>
          </template>
        </el-table-column>

        <el-table-column
          prop="description"
          label="说明"
          min-width="300"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span class="config-description">{{ row.description || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column
          prop="updated_at"
          label="更新时间"
          min-width="180"
        >
          <template #default="{ row }">
            <span class="config-time">{{ formatTime(row.updated_at) }}</span>
          </template>
        </el-table-column>

        <el-table-column
          label="操作"
          width="150"
          fixed="right"
        >
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                v-if="editingKey !== row.config_key"
                type="primary"
                size="small"
                link
                @click="handleEdit(row)"
              >
                编辑
              </el-button>
              <template v-else>
                <el-button
                  type="success"
                  size="small"
                  link
                  @click="handleSave(row)"
                  :loading="saving"
                >
                  保存
                </el-button>
                <el-button
                  type="info"
                  size="small"
                  link
                  @click="handleCancel"
                >
                  取消
                </el-button>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getConfigs, updateConfig, reloadConfigs } from '@/api/config'

export default {
  name: 'SystemConfig',
  setup() {
    const configs = ref([])
    const loading = ref(false)
    const saving = ref(false)
    const reloading = ref(false)
    const editingKey = ref(null)
    const editingValue = ref('')
    const originalValue = ref('')

    // 加载配置列表
    const loadConfigs = async () => {
      try {
        loading.value = true
        const data = await getConfigs()
        configs.value = data.configs || []
      } catch (error) {
        console.error('加载配置失败:', error)
        ElMessage.error('加载配置失败')
      } finally {
        loading.value = false
      }
    }

    // 编辑配置
    const handleEdit = (row) => {
      editingKey.value = row.config_key
      editingValue.value = row.config_value
      originalValue.value = row.config_value
    }

    // 取消编辑
    const handleCancel = () => {
      editingKey.value = null
      editingValue.value = ''
      originalValue.value = ''
    }

    // 保存配置
    const handleSave = async (row) => {
      // 值未改变
      if (editingValue.value === originalValue.value) {
        handleCancel()
        return
      }

      // 值为空
      if (!editingValue.value.trim()) {
        ElMessage.warning('配置值不能为空')
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认要修改配置项 "${row.config_key}" 吗？修改后将立即生效。`,
          '确认修改',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        saving.value = true
        await updateConfig(row.config_key, {
          config_value: editingValue.value.trim()
        })

        ElMessage.success('配置更新成功')
        handleCancel()
        await loadConfigs()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('更新配置失败:', error)
          ElMessage.error('更新配置失败')
        }
      } finally {
        saving.value = false
      }
    }

    // 重新加载配置
    const handleReload = async () => {
      try {
        await ElMessageBox.confirm(
          '确认要重新加载配置吗？这将从数据库重新读取所有配置项。',
          '确认重新加载',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        reloading.value = true
        await reloadConfigs()
        ElMessage.success('配置重新加载成功')
        await loadConfigs()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('重新加载配置失败:', error)
          ElMessage.error('重新加载配置失败')
        }
      } finally {
        reloading.value = false
      }
    }

    // 格式化时间
    const formatTime = (time) => {
      if (!time) return '-'
      const date = new Date(time)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      const seconds = String(date.getSeconds()).padStart(2, '0')
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    }

    onMounted(() => {
      loadConfigs()
    })

    return {
      configs,
      loading,
      saving,
      reloading,
      editingKey,
      editingValue,
      handleEdit,
      handleCancel,
      handleSave,
      handleReload,
      formatTime
    }
  }
}
</script>

<style scoped>
/* 容器 */
.system-config-container {
  padding: var(--spacing-lg);
  min-height: 100vh;
  background-color: var(--color-bg);
}

/* 页面标题区域 */
.page-header {
  margin-bottom: var(--spacing-xl);
}

.page-title {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  line-height: var(--line-height-tight);
}

.page-subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: var(--line-height-normal);
}

/* 操作栏 */
.action-bar {
  margin-bottom: var(--spacing-md);
  display: flex;
  justify-content: flex-end;
}

/* 配置卡片 */
.config-card {
  box-shadow: var(--shadow-sm);
  border-radius: var(--radius-md);
}

/* 配置表格 */
.config-table {
  width: 100%;
}

/* 配置键 */
.config-key {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  color: var(--color-primary);
  font-weight: var(--font-weight-medium);
}

/* 配置值 */
.config-value {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
}

/* 配置说明 */
.config-description {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* 配置时间 */
.config-time {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: var(--spacing-xs);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .system-config-container {
    padding: var(--spacing-md);
  }

  .page-title {
    font-size: var(--font-size-lg);
  }
}
</style>
