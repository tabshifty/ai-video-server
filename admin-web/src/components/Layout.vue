<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  DataAnalysis,
  Film,
  UploadFilled,
  MagicStick,
  User,
  List,
  Setting,
  SwitchButton
} from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const active = computed(() => route.path)
const pageTitle = computed(() => {
  const map = {
    '/dashboard': '系统仪表盘',
    '/videos': '视频管理',
    '/upload': '上传中心',
    '/scrape': '刮削管理',
    '/users': '用户管理',
    '/tasks': '任务监控',
    '/settings': '系统设置'
  }
  return map[route.path] || 'Admin Console'
})

async function onLogout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="admin-shell">
    <el-aside width="240px" class="shell-aside">
      <div class="brand-block">
        <div class="brand-badge">AV</div>
        <div>
          <div class="brand-title">Admin Console</div>
          <div class="brand-subtitle">Home Video Server</div>
        </div>
      </div>

      <el-menu :default-active="active" router class="nav-menu">
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/videos">
          <el-icon><Film /></el-icon>
          <span>视频管理</span>
        </el-menu-item>
        <el-menu-item index="/upload">
          <el-icon><UploadFilled /></el-icon>
          <span>上传视频</span>
        </el-menu-item>
        <el-menu-item index="/scrape">
          <el-icon><MagicStick /></el-icon>
          <span>刮削管理</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><List /></el-icon>
          <span>任务监控</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="shell-header">
        <div>
          <div class="shell-header-title">{{ pageTitle }}</div>
          <div class="shell-header-subtitle">管理员工作台</div>
        </div>
        <el-button plain type="danger" :icon="SwitchButton" @click="onLogout">退出登录</el-button>
      </el-header>
      <el-main class="shell-main"><slot /></el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
}

.shell-aside {
  background: linear-gradient(180deg, #881337 0%, #be123c 45%, #1e3a8a 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.15);
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 16px 14px;
  color: #fff;
}

.brand-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-size: 14px;
  font-family: 'Fira Code', monospace;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.22);
  border: 1px solid rgba(255, 255, 255, 0.42);
}

.brand-title {
  font-weight: 700;
  letter-spacing: 0.2px;
}

.brand-subtitle {
  margin-top: 2px;
  font-size: 12px;
  opacity: 0.88;
}

.nav-menu {
  background: transparent;
  border-right: none;
}

.nav-menu.el-menu .el-menu-item {
  margin: 4px 10px;
  border-radius: 10px;
  color: rgba(255, 255, 255, 0.8);
}

.nav-menu.el-menu .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.nav-menu.el-menu .el-menu-item.is-active {
  color: #fff;
  background: rgba(255, 255, 255, 0.24);
}

.shell-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(136, 19, 55, 0.12);
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(8px);
  position: sticky;
  top: 0;
  z-index: 20;
}

.shell-header-title {
  font-size: 18px;
  font-weight: 600;
  color: #7f1d1d;
}

.shell-header-subtitle {
  margin-top: 2px;
  font-size: 12px;
  color: #6b7280;
}

.shell-main {
  padding: 18px;
}

@media (max-width: 992px) {
  .shell-aside {
    width: 200px !important;
  }
}
</style>
