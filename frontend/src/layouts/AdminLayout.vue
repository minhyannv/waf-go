<template>
  <div class="admin-layout">
    <!-- 左侧菜单 -->
    <div class="sidebar" :style="{ width: collapsed ? '64px' : '250px' }">
      <div class="logo">
        <h3 v-if="!collapsed">WAF 管理系统</h3>
        <h3 v-else>W</h3>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        
        <el-menu-item index="/rules">
          <el-icon><Lock /></el-icon>
          <span>规则管理</span>
        </el-menu-item>
        
        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <span>攻击日志</span>
        </el-menu-item>
        
        <el-menu-item index="/policies">
          <el-icon><Files /></el-icon>
          <span>策略管理</span>
        </el-menu-item>
        
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </div>

    <!-- 右侧内容区域 -->
    <div class="main-container">
      <!-- 顶部导航 -->
      <div class="header">
        <div class="header-left">
          <el-button
            type="text"
            size="large"
            @click="toggleCollapse"
          >
            <el-icon><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
          </el-button>
        </div>
        
        <div class="header-right">
          <el-dropdown @command="handleUserCommand">
            <span class="user-dropdown">
              <el-avatar :size="32" icon="UserFilled" />
              <span class="username">{{ userInfo.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 主体内容 -->
      <div class="main-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Lock, Document, Files, Setting, Fold, Expand, ArrowDown } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)
const userInfo = ref({
  username: '管理员',
  role: 'admin'
})

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/rules')) return '/rules'
  if (path.startsWith('/logs')) return '/logs'
  if (path.startsWith('/policies')) return '/policies'
  if (path.startsWith('/settings')) return '/settings'
  return '/dashboard'
})

const toggleCollapse = () => {
  collapsed.value = !collapsed.value
}

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'profile':
      // TODO: 跳转到个人中心
      break
    case 'logout':
      handleLogout()
      break
  }
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    ElMessage.success('退出登录成功')
    router.push('/login')
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  // 从 localStorage 获取用户信息
  const user = localStorage.getItem('user')
  if (user) {
    userInfo.value = JSON.parse(user)
  }
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.sidebar {
  background-color: #304156;
  transition: width 0.3s;
  flex-shrink: 0;
  height: 100vh;
  overflow-y: auto;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b2f3a;
  color: white;
  border-bottom: 1px solid #1d1f23;
}

.logo h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.header {
  height: 60px;
  background: white;
  border-bottom: 1px solid #e6e8eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 5px;
  transition: background-color 0.3s;
}

.user-dropdown:hover {
  background-color: #f5f5f5;
}

.username {
  margin: 0 10px;
  font-size: 14px;
}

.main-content {
  flex: 1;
  background: #f5f5f5;
  padding: 20px;
  overflow-y: auto;
  height: calc(100vh - 60px);
}
</style> 