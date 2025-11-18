<template>
  <div class="admin-user-management">
    <div class="header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="handleCreate">新建用户</el-button>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部" clearable style="width: 120px">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
            <el-option label="只读用户" value="readonly" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="活跃" :value="1" />
            <el-option label="已禁用" :value="2" />
            <el-option label="已锁定" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadUsers">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 用户列表表格 -->
    <el-card class="table-card">
      <el-table :data="users" v-loading="loading" border stripe>
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="uuid" label="UUID" min-width="200" show-overflow-tooltip />
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)">
              {{ getRoleText(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="180">
          <template #default="{ row }">
            {{ formatDate(row.last_login_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="warning" size="small" @click="handleResetPassword(row)">重置密码</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户弹窗 -->
    <el-dialog
      :title="dialogTitle"
      v-model="showUserDialog"
      width="500px"
      @close="resetForm"
    >
      <el-form :model="userForm" :rules="userRules" ref="userFormRef" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" :disabled="isEdit" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="userForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
            <el-option label="只读用户" value="readonly" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="isEdit">
          <el-select v-model="userForm.status" placeholder="请选择状态">
            <el-option label="活跃" :value="1" />
            <el-option label="已禁用" :value="2" />
            <el-option label="已锁定" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="手机" prop="phone">
          <el-input v-model="userForm.phone" placeholder="请输入手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUserDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码弹窗 -->
    <el-dialog
      title="重置密码"
      v-model="showPasswordDialog"
      width="400px"
    >
      <el-form :model="passwordForm" :rules="passwordRules" ref="passwordFormRef" label-width="100px">
        <el-form-item label="新密码" prop="password">
          <el-input v-model="passwordForm.password" type="password" placeholder="请输入新密码" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" placeholder="请再次输入密码" show-password />
        </el-form-item>
      </el-form>
      <div class="password-actions">
        <el-button size="small" @click="generatePassword">生成随机密码</el-button>
      </div>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="handlePasswordSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getUserList, createUser, deleteUser, updateUserStatus, updateUserRole, updateUserInfo, resetUserPassword } from '@/api/user'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  name: 'AdminUserManagement',
  data() {
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.passwordForm.password) {
        callback(new Error('两次输入的密码不一致'))
      } else {
        callback()
      }
    }

    return {
      loading: false,
      submitting: false,
      users: [],
      searchForm: {
        role: '',
        status: null
      },
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0
      },
      showUserDialog: false,
      showPasswordDialog: false,
      isEdit: false,
      currentUser: null,
      userForm: {
        username: '',
        password: '',
        role: 'user',
        status: 1,
        nickname: '',
        email: '',
        phone: ''
      },
      userRules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' },
          { min: 3, max: 32, message: '用户名长度为3-32个字符', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 8, message: '密码长度不能少于8位', trigger: 'blur' }
        ],
        role: [
          { required: true, message: '请选择角色', trigger: 'change' }
        ],
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
        ]
      },
      passwordForm: {
        password: '',
        confirmPassword: ''
      },
      passwordRules: {
        password: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 8, message: '密码长度不能少于8位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请再次输入密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ]
      }
    }
  },

  computed: {
    dialogTitle() {
      return this.isEdit ? '编辑用户' : '新建用户'
    }
  },

  mounted() {
    this.loadUsers()
  },

  methods: {
    async loadUsers() {
      this.loading = true
      try {
        const params = {
          page: this.pagination.page,
          page_size: this.pagination.pageSize,
          ...this.searchForm
        }
        const result = await getUserList(params)
        this.users = result.users || []
        this.pagination.total = result.total || 0
      } catch (error) {
        console.error('加载用户列表失败:', error)
        ElMessage.error('加载用户列表失败')
      } finally {
        this.loading = false
      }
    },

    handleCreate() {
      this.isEdit = false
      this.currentUser = null
      this.showUserDialog = true
    },

    handleEdit(row) {
      this.isEdit = true
      this.currentUser = row
      this.userForm = {
        username: row.username,
        role: row.role,
        status: row.status,
        nickname: '',
        email: '',
        phone: ''
      }
      this.showUserDialog = true
    },

    handleReset() {
      this.searchForm = {
        role: '',
        status: null
      }
      this.pagination.page = 1
      this.loadUsers()
    },

    async handleSubmit() {
      const formRef = this.$refs.userFormRef
      if (!formRef) return

      formRef.validate(async (valid) => {
        if (!valid) return

        this.submitting = true
        try {
          if (this.isEdit) {
            // 编辑模式：分别更新角色、状态和基本信息
            const uuid = this.currentUser.uuid
            await Promise.all([
              updateUserRole(uuid, this.userForm.role),
              updateUserStatus(uuid, this.userForm.status),
              updateUserInfo(uuid, {
                nickname: this.userForm.nickname || undefined,
                email: this.userForm.email || undefined,
                phone: this.userForm.phone || undefined
              })
            ])
            ElMessage.success('用户信息更新成功')
          } else {
            // 创建模式
            await createUser(this.userForm)
            ElMessage.success('用户创建成功')
          }
          this.showUserDialog = false
          await this.loadUsers()
        } catch (error) {
          console.error('操作失败:', error)
        } finally {
          this.submitting = false
        }
      })
    },

    handleResetPassword(row) {
      this.currentUser = row
      this.showPasswordDialog = true
    },

    handlePasswordSubmit() {
      const formRef = this.$refs.passwordFormRef
      if (!formRef) return

      formRef.validate(async (valid) => {
        if (!valid) return

        this.submitting = true
        try {
          await resetUserPassword(this.currentUser.uuid, this.passwordForm.password)
          ElMessage.success('密码重置成功')
          this.showPasswordDialog = false
          this.passwordForm = { password: '', confirmPassword: '' }
        } catch (error) {
          console.error('重置密码失败:', error)
        } finally {
          this.submitting = false
        }
      })
    },

    generatePassword() {
      const chars = 'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678@#$%&*'
      let password = ''
      for (let i = 0; i < 12; i++) {
        password += chars.charAt(Math.floor(Math.random() * chars.length))
      }
      this.passwordForm.password = password
      this.passwordForm.confirmPassword = password
      ElMessage.success('已生成随机密码: ' + password)
    },

    handleDelete(row) {
      ElMessageBox.confirm(
        `确定要删除用户 "${row.username}" 吗？此操作不可恢复。`,
        '删除确认',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(async () => {
        try {
          await deleteUser(row.uuid)
          ElMessage.success('用户已删除')
          await this.loadUsers()
        } catch (error) {
          console.error('删除用户失败:', error)
        }
      }).catch(() => {
        // 用户取消
      })
    },

    resetForm() {
      this.userForm = {
        username: '',
        password: '',
        role: 'user',
        status: 1,
        nickname: '',
        email: '',
        phone: ''
      }
      this.$refs.userFormRef?.resetFields()
    },

    getRoleType(role) {
      const map = { admin: 'danger', user: 'success', readonly: 'info' }
      return map[role] || 'info'
    },

    getRoleText(role) {
      const map = { admin: '管理员', user: '普通用户', readonly: '只读用户' }
      return map[role] || role
    },

    getStatusType(status) {
      const map = { 1: 'success', 2: 'warning', 3: 'danger' }
      return map[status] || 'info'
    },

    getStatusText(status) {
      const map = { 1: '活跃', 2: '已禁用', 3: '已锁定' }
      return map[status] || '未知'
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
    }
  }
}
</script>

<style scoped>
.admin-user-management {
  padding: var(--spacing-lg);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.header h2 {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0;
}

.search-card {
  margin-bottom: var(--spacing-md);
}

.table-card {
  margin-bottom: var(--spacing-md);
}

.pagination {
  margin-top: var(--spacing-lg);
  display: flex;
  justify-content: flex-end;
}

.password-actions {
  margin-top: var(--spacing-md);
  text-align: right;
}
</style>
