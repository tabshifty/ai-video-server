<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function submit() {
  loading.value = true
  try {
    await auth.login(form)
    if (!auth.isAdmin) {
      ElMessage.error('当前账号不是管理员')
      await auth.logout()
      return
    }
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-grid">
      <div class="hero-panel">
        <h1 class="hero-title">家用视频服务器</h1>
        <p class="hero-subtitle">管理内容、用户与系统任务的统一控制台</p>
        <div class="hero-pills">
          <span>Video Admin</span>
          <span>Secure JWT</span>
          <span>Real-time Task</span>
        </div>
      </div>

      <el-card class="soft-card login-card">
        <h2 class="login-title">管理员登录</h2>
        <el-form @submit.prevent="submit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="admin" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
          <el-button type="primary" :loading="loading" @click="submit" class="login-btn">登录</el-button>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
}

.login-grid {
  width: min(980px, 100%);
  display: grid;
  grid-template-columns: 1.2fr 0.9fr;
  gap: 16px;
}

.hero-panel {
  border-radius: 18px;
  padding: 28px;
  color: #fff;
  background:
    radial-gradient(circle at 18% 18%, rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0) 45%),
    linear-gradient(140deg, #881337 0%, #be123c 45%, #2563eb 100%);
  box-shadow: 0 25px 65px rgba(31, 41, 55, 0.24);
}

.hero-title {
  margin: 0;
  font-size: 34px;
  line-height: 1.15;
}

.hero-subtitle {
  margin-top: 10px;
  font-size: 15px;
  line-height: 1.7;
  opacity: 0.92;
}

.hero-pills {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.hero-pills span {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.15);
  border-radius: 999px;
  padding: 5px 10px;
  font-size: 12px;
}

.login-card {
  padding-top: 6px;
}

.login-title {
  margin: 0 0 14px;
  color: #881337;
}

.login-btn {
  width: 100%;
  margin-top: 4px;
}

@media (max-width: 900px) {
  .login-grid {
    grid-template-columns: 1fr;
  }
}
</style>
