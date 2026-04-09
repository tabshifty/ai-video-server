<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const active = computed(() => route.path)

async function onLogout() {
  await auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container style="min-height: 100vh">
    <el-aside width="220px" style="background: #0f172a; color: #fff">
      <div style="font-size: 18px; font-weight: 700; padding: 18px">Admin Console</div>
      <el-menu :default-active="active" router background-color="#0f172a" text-color="#cbd5e1" active-text-color="#fff">
        <el-menu-item index="/dashboard">仪表盘</el-menu-item>
        <el-menu-item index="/videos">视频管理</el-menu-item>
        <el-menu-item index="/upload">上传视频</el-menu-item>
        <el-menu-item index="/scrape">刮削管理</el-menu-item>
        <el-menu-item index="/users">用户管理</el-menu-item>
        <el-menu-item index="/tasks">任务监控</el-menu-item>
        <el-menu-item index="/settings">系统设置</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="display:flex;justify-content:flex-end;align-items:center;border-bottom:1px solid #e5e7eb">
        <el-button type="danger" plain @click="onLogout">退出登录</el-button>
      </el-header>
      <el-main style="background: #f8fafc"><slot /></el-main>
    </el-container>
  </el-container>
</template>
