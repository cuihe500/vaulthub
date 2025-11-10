<template>
  <div class="vault-list">
    <!-- 顶部操作栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="page-title">密钥管理</h2>
      </div>
      <div class="toolbar-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新建密钥
        </el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-select
        v-model="filterType"
        placeholder="全部类型"
        clearable
        @change="handleFilterChange"
        class="filter-select"
      >
        <el-option label="全部类型" value="" />
        <el-option label="API密钥" value="api_key" />
        <el-option label="数据库凭证" value="db_credential" />
        <el-option label="证书" value="certificate" />
        <el-option label="SSH密钥" value="ssh_key" />
        <el-option label="令牌" value="token" />
        <el-option label="密码" value="password" />
        <el-option label="其他" value="other" />
      </el-select>
    </div>

    <!-- 密钥列表表格 -->
    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="secretList"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="secret_name" label="密钥名称" min-width="180">
          <template #default="{ row }">
            <div class="secret-name">
              <el-icon class="name-icon"><Key /></el-icon>
              <span>{{ row.secret_name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="secret_type" label="类型" width="140">
          <template #default="{ row }">
            <el-tag :type="getSecretTypeTag(row.secret_type)">
              {{ getSecretTypeName(row.secret_type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column
          prop="description"
          label="描述"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span class="description">
              {{ row.description || '-' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column prop="access_count" label="访问次数" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.access_count || 0 }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="last_accessed_at" label="最后访问" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_accessed_at) }}
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              :icon="View"
              @click="handleDecrypt(row)"
            >
              查看
            </el-button>
            <el-button
              type="danger"
              link
              :icon="Delete"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>

        <template #empty>
          <el-empty description="暂无密钥数据" />
        </template>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 解密密钥对话框 -->
    <el-dialog
      v-model="decryptDialogVisible"
      title="查看密钥"
      width="600px"
      @close="handleDecryptDialogClose"
    >
      <el-form :model="decryptForm" label-width="100px">
        <el-form-item label="密钥名称">
          <el-input v-model="currentSecret.secret_name" disabled />
        </el-form-item>
        <el-form-item label="安全PIN码" required>
          <el-input
            v-model="decryptForm.security_pin"
            type="password"
            placeholder="请输入您的安全PIN码以解密密钥"
            show-password
          />
        </el-form-item>
        <el-form-item v-if="decryptedData" label="密钥内容">
          <el-input
            v-model="decryptedData"
            type="textarea"
            :rows="6"
            readonly
            class="decrypted-content"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="decryptDialogVisible = false">取消</el-button>
        <el-button
          v-if="!decryptedData"
          type="primary"
          :loading="decrypting"
          @click="handleConfirmDecrypt"
        >
          解密
        </el-button>
        <el-button
          v-else
          type="primary"
          :icon="CopyDocument"
          @click="handleCopySecret"
        >
          复制
        </el-button>
      </template>
    </el-dialog>

    <!-- 新建密钥对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="新建密钥"
      width="600px"
      @close="handleCreateDialogClose"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createFormRules"
        label-width="100px"
      >
        <el-form-item label="密钥名称" prop="secret_name">
          <el-input v-model="createForm.secret_name" placeholder="请输入密钥名称" />
        </el-form-item>
        <el-form-item label="密钥类型" prop="secret_type">
          <el-select v-model="createForm.secret_type" placeholder="请选择密钥类型">
            <el-option label="API密钥" value="api_key" />
            <el-option label="数据库凭证" value="db_credential" />
            <el-option label="证书" value="certificate" />
            <el-option label="SSH密钥" value="ssh_key" />
            <el-option label="令牌" value="token" />
            <el-option label="密码" value="password" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="密钥内容" prop="plain_data">
          <el-input
            v-model="createForm.plain_data"
            type="textarea"
            :rows="4"
            placeholder="请输入密钥内容"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入密钥描述（可选）"
          />
        </el-form-item>
        <el-form-item label="安全PIN码" prop="security_pin">
          <el-input
            v-model="createForm.security_pin"
            type="password"
            placeholder="请输入您的安全PIN码以加密密钥"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="creating"
          @click="handleConfirmCreate"
        >
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Key, View, Delete, CopyDocument } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getSecretList, createSecret, deleteSecret, decryptSecret } from '@/api/vault'

// 加载状态
const loading = ref(false)
const decrypting = ref(false)
const creating = ref(false)

// 密钥列表
const secretList = ref([])

// 筛选条件
const filterType = ref('')

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 解密对话框
const decryptDialogVisible = ref(false)
const currentSecret = ref({})
const decryptForm = reactive({
  security_pin: ''
})
const decryptedData = ref('')

// 新建对话框
const createDialogVisible = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  secret_name: '',
  secret_type: '',
  plain_data: '',
  description: '',
  security_pin: ''
})

const createFormRules = {
  secret_name: [
    { required: true, message: '请输入密钥名称', trigger: 'blur' }
  ],
  secret_type: [
    { required: true, message: '请选择密钥类型', trigger: 'change' }
  ],
  plain_data: [
    { required: true, message: '请输入密钥内容', trigger: 'blur' }
  ],
  security_pin: [
    { required: true, message: '请输入安全PIN码', trigger: 'blur' }
  ]
}

// 获取密钥列表
const fetchSecretList = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filterType.value) {
      params.secret_type = filterType.value
    }

    const response = await getSecretList(params)
    secretList.value = response.secrets || []
    pagination.total = response.total || 0
  } catch (error) {
    console.error('获取密钥列表失败:', error)
    ElMessage.error('获取密钥列表失败')
  } finally {
    loading.value = false
  }
}

// 筛选变化
const handleFilterChange = () => {
  pagination.page = 1
  fetchSecretList()
}

// 分页变化
const handlePageChange = () => {
  fetchSecretList()
}

const handleSizeChange = () => {
  pagination.page = 1
  fetchSecretList()
}

// 新建密钥
const handleCreate = () => {
  createDialogVisible.value = true
}

const handleConfirmCreate = async () => {
  try {
    await createFormRef.value.validate()
    creating.value = true

    await createSecret(createForm)
    ElMessage.success('密钥创建成功')
    createDialogVisible.value = false
    fetchSecretList()
  } catch (error) {
    if (error !== false) {
      console.error('创建密钥失败:', error)
      ElMessage.error('创建密钥失败')
    }
  } finally {
    creating.value = false
  }
}

const handleCreateDialogClose = () => {
  createFormRef.value?.resetFields()
  Object.assign(createForm, {
    secret_name: '',
    secret_type: '',
    plain_data: '',
    description: '',
    security_pin: ''
  })
}

// 解密密钥
const handleDecrypt = (row) => {
  currentSecret.value = row
  decryptDialogVisible.value = true
}

const handleConfirmDecrypt = async () => {
  if (!decryptForm.security_pin) {
    ElMessage.warning('请输入安全PIN码')
    return
  }

  try {
    decrypting.value = true
    const response = await decryptSecret(currentSecret.value.secret_uuid, {
      security_pin: decryptForm.security_pin
    })
    decryptedData.value = response.plain_data
    ElMessage.success('解密成功')
  } catch (error) {
    console.error('解密失败:', error)
    ElMessage.error('解密失败，请检查PIN码是否正确')
  } finally {
    decrypting.value = false
  }
}

const handleDecryptDialogClose = () => {
  decryptForm.security_pin = ''
  decryptedData.value = ''
  currentSecret.value = {}
}

// 复制密钥
const handleCopySecret = async () => {
  try {
    await navigator.clipboard.writeText(decryptedData.value)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 删除密钥
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除密钥"${row.secret_name}"吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteSecret(row.secret_uuid)
    ElMessage.success('删除成功')
    fetchSecretList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 获取密钥类型名称
const getSecretTypeName = (type) => {
  const typeMap = {
    api_key: 'API密钥',
    db_credential: '数据库凭证',
    certificate: '证书',
    ssh_key: 'SSH密钥',
    token: '令牌',
    password: '密码',
    other: '其他'
  }
  return typeMap[type] || type
}

// 获取密钥类型标签颜色
const getSecretTypeTag = (type) => {
  const tagMap = {
    api_key: 'primary',
    db_credential: 'success',
    certificate: 'warning',
    ssh_key: 'info',
    token: 'danger',
    password: '',
    other: 'info'
  }
  return tagMap[type] || 'info'
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 组件挂载时获取列表
onMounted(() => {
  fetchSecretList()
})
</script>

<style scoped>
.vault-list {
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

/* 筛选栏 */
.filter-bar {
  margin-bottom: var(--spacing-md);
  display: flex;
  gap: var(--spacing-sm);
}

.filter-select {
  width: 200px;
}

/* 表格卡片 */
.table-card {
  border-radius: var(--radius-md);
}

.secret-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.name-icon {
  color: var(--color-primary);
  font-size: var(--font-size-lg);
}

.description {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

/* 分页 */
.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

/* 解密内容样式 */
.decrypted-content :deep(textarea) {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
}
</style>
