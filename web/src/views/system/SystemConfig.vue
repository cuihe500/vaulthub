<template>
  <div class="system-config">
    <el-card class="config-card">
      <template #header>
        <div class="card-header">
          <span>系统配置</span>
          <el-button type="primary" @click="handleReload">
            <el-icon><Refresh /></el-icon>
            刷新配置
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="configs"
        border
        stripe
      >
        <el-table-column prop="config_key" label="配置键" width="300" />
        <el-table-column prop="config_value" label="配置值" show-overflow-tooltip />
        <el-table-column prop="description" label="说明" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
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
  padding: var(--spacing-lg);
}

.config-card {
  max-width: 1400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
