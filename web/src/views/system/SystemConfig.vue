<template>
  <div class="system-config">
    <!-- 顶部操作栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="page-title">系统配置</h2>
      </div>
      <div class="toolbar-right">
        <el-button type="primary" :icon="Refresh" @click="handleReload">
          刷新配置
        </el-button>
      </div>
    </div>

    <!-- 配置列表表格 -->
    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="configs"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="config_key" label="配置键" width="300" />
        <el-table-column prop="config_value" label="配置值" show-overflow-tooltip />
        <el-table-column prop="description" label="说明" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      title="编辑配置"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="editForm"
        label-width="100px"
      >
        <el-form-item label="配置键">
          <el-input v-model="editForm.config_key" disabled />
        </el-form-item>
        <el-form-item label="配置值">
          <el-input
            v-model="editForm.config_value"
            type="textarea"
            :rows="4"
          />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="editForm.description" disabled />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getConfigs, updateConfig, reloadConfigs } from '@/api/system'

const loading = ref(false)
const configs = ref([])
const dialogVisible = ref(false)
const editForm = ref({
  config_key: '',
  config_value: '',
  description: ''
})

const fetchConfigs = async () => {
  loading.value = true
  try {
    const result = await getConfigs()
    // 后端返回 {configs: [], total: number}，提取 configs 数组
    configs.value = result.configs
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败')
  } finally {
    loading.value = false
  }
}

const handleEdit = (row) => {
  editForm.value = { ...row }
  dialogVisible.value = true
}

const handleSave = async () => {
  try {
    await updateConfig(editForm.value.config_key, {
      config_value: editForm.value.config_value
    })
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchConfigs()
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存失败')
  }
}

const handleReload = async () => {
  try {
    await reloadConfigs()
    ElMessage.success('配置已重新加载')
    fetchConfigs()
  } catch (error) {
    console.error('重载配置失败:', error)
    ElMessage.error('重载失败')
  }
}

onMounted(() => {
  fetchConfigs()
})
</script>

<style scoped>
.system-config {
  width: 100%;
  height: 100%;
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

.toolbar-right {
  display: flex;
  gap: var(--spacing-sm);
}

/* 表格卡片 */
.table-card {
  border-radius: var(--radius-md);
}
</style>
