<template>
  <el-container class="main-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="main-layout__aside">
      <div class="logo-container">
        <div v-if="!isCollapse" class="logo">
          <span class="logo-text">VaultHub</span>
        </div>
        <div v-else class="logo-mini">
          <span class="logo-text-mini">V</span>
        </div>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/vault">
          <el-icon><Lock /></el-icon>
          <template #title>密钥管理</template>
        </el-menu-item>
        <el-menu-item index="/user">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container class="main-layout__container">
      <!-- 顶部导航栏 -->
      <el-header class="main-layout__header">
        <div class="header-left">
          <el-button
            :icon="isCollapse ? Expand : Fold"
            text
            @click="toggleCollapse"
            class="collapse-btn"
          />
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" class="user-avatar">
                {{ userInitial }}
              </el-avatar>
              <span class="username">{{ username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  <span>退出登录</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-layout__main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, User, Expand, Fold, SwitchButton } from '@element-plus/icons-vue'
import { logout, getCurrentUser } from '@/api/auth'
import { removeToken } from '@/utils/storage'

const router = useRouter()
const route = useRoute()

// 侧边栏折叠状态
const isCollapse = ref(false)

// 用户信息
const username = ref('用户')

// 切换侧边栏折叠
const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 当前激活的菜单
const activeMenu = computed(() => route.path)

// 用户名首字母
const userInitial = computed(() => {
  return username.value.charAt(0).toUpperCase()
})

// 获取当前用户信息
const fetchUserInfo = async () => {
  try {
    const userInfo = await getCurrentUser()
    username.value = userInfo.username || '用户'
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

// 处理下拉菜单命令
const handleCommand = async (command) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm(
        '确定要退出登录吗？',
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )

      // 调用退出登录接口
      await logout()

      // 清除本地Token
      removeToken()

      // 提示并跳转登录页
      ElMessage.success('已退出登录')
      router.push('/login')
    } catch (error) {
      if (error !== 'cancel') {
        console.error('退出登录失败:', error)
      }
    }
  }
}

// 组件挂载时获取用户信息
onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.main-layout {
  width: 100%;
  height: 100vh;
}

/* ==================== 侧边栏样式 ==================== */
.main-layout__aside {
  background-color: var(--color-white);
  border-right: 1px solid var(--color-border);
  transition: width var(--duration-base) var(--easing);
  overflow: hidden;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid var(--color-border);
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.logo-mini {
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text-mini {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sidebar-menu {
  border-right: none;
  height: calc(100vh - 60px);
}

/* ==================== 顶部导航栏样式 ==================== */
.main-layout__header {
  background-color: var(--color-white);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-btn {
  font-size: var(--font-size-lg);
  color: var(--color-text-primary);
}

.collapse-btn:hover {
  color: var(--color-primary);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);
  transition: background-color var(--duration-fast) var(--easing);
}

.user-info:hover {
  background-color: var(--color-bg);
}

.user-avatar {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: var(--color-white);
  font-weight: var(--font-weight-medium);
}

.username {
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  font-weight: var(--font-weight-medium);
}

/* ==================== 主内容区样式 ==================== */
.main-layout__container {
  display: flex;
  flex-direction: column;
}

.main-layout__main {
  background-color: var(--color-bg);
  padding: var(--spacing-lg);
  overflow-y: auto;
}
</style>
