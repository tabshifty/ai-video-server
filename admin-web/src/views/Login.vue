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
      <section class="hero-panel" aria-label="系统介绍">
        <div class="hero-copy">
          <p class="hero-eyebrow">AI VIDEO ADMIN</p>
          <h1 class="hero-title">家用视频服务器</h1>
          <p class="hero-subtitle">内容、账号、图片合集与任务流转统一收口到一个稳定的管理控制台。</p>
        </div>
        <div class="hero-metrics">
          <article class="hero-metric">
            <span class="hero-metric__label">权限边界</span>
            <strong class="hero-metric__value">管理员专用</strong>
          </article>
          <article class="hero-metric">
            <span class="hero-metric__label">任务视图</span>
            <strong class="hero-metric__value">上传到刮削全链路</strong>
          </article>
          <article class="hero-metric">
            <span class="hero-metric__label">内容组织</span>
            <strong class="hero-metric__value">视频与图片统一治理</strong>
          </article>
        </div>
        <div class="hero-pills">
          <span>后台权限校验</span>
          <span>任务监控</span>
          <span>内容治理</span>
        </div>
      </section>

      <div class="login-column">
        <el-card class="soft-card login-card">
        <template #header>
          <div class="section-head">
            <div>
              <h2 class="login-title">管理员登录</h2>
              <p class="login-subtitle">请使用具有后台权限的账号登录</p>
            </div>
          </div>
        </template>
        <div class="login-form-wrap">
          <el-form class="login-form" label-position="top" @submit.prevent="submit">
            <el-form-item label="用户名">
              <el-input v-model="form.username" placeholder="admin" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" show-password />
            </el-form-item>
            <el-button type="primary" :loading="loading" @click="submit" class="login-btn">登录</el-button>
          </el-form>
        </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(20px, 4vw, 40px);
}

.login-grid {
  width: min(1120px, 100%);
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(360px, 430px);
  gap: clamp(18px, 3vw, 28px);
  align-items: stretch;
}

.hero-panel {
  position: relative;
  overflow: hidden;
  min-height: clamp(440px, 64vh, 620px);
  padding: clamp(28px, 4vw, 42px);
  display: grid;
  align-content: space-between;
  gap: 22px;
  border-radius: 28px;
  color: #f8fafc;
  background:
    radial-gradient(circle at 18% 18%, rgba(255, 255, 255, 0.14), rgba(255, 255, 255, 0) 34%),
    radial-gradient(circle at 82% 18%, rgba(96, 165, 250, 0.18), rgba(96, 165, 250, 0) 30%),
    linear-gradient(145deg, #0f172a 0%, #7f1d1d 52%, #1d4ed8 100%);
  box-shadow: 0 36px 80px rgba(15, 23, 42, 0.28);
}

.hero-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.08), rgba(15, 23, 42, 0.3));
  pointer-events: none;
}

.hero-panel > * {
  position: relative;
  z-index: 1;
}

.hero-copy {
  display: grid;
  gap: 14px;
}

.hero-eyebrow {
  margin: 0;
  color: rgba(241, 245, 249, 0.72);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.18em;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.hero-metric {
  display: grid;
  gap: 8px;
  padding: 16px;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
}

.hero-metric__label {
  font-size: 12px;
  color: rgba(241, 245, 249, 0.72);
}

.hero-metric__value {
  font-size: 15px;
  line-height: 1.5;
  color: #fff;
}

.hero-title {
  margin: 0;
  max-width: 8ch;
  font-size: clamp(34px, 5vw, 52px);
  line-height: 1.04;
  color: #fff;
  text-wrap: balance;
}

.hero-subtitle {
  max-width: 32rem;
  margin: 0;
  font-size: 16px;
  line-height: 1.75;
  color: rgba(241, 245, 249, 0.88);
}

.hero-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-pills span {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgba(255, 255, 255, 0.22);
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  padding: 7px 12px;
  font-size: 12px;
}

.login-column {
  display: flex;
  align-items: center;
}

.login-card {
  width: 100%;
  border-radius: 24px;
  border: 1px solid rgba(203, 213, 225, 0.72);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 28px 64px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(18px);
}

.login-card :deep(.el-card__header) {
  padding: 28px 28px 18px;
  border-bottom: none;
}

.login-card :deep(.el-card__body) {
  padding: 0 28px 28px;
}

.login-form-wrap {
  display: grid;
  gap: 12px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.login-form :deep(.el-form-item__label) {
  padding-bottom: 8px;
  color: #0f172a;
  font-weight: 600;
}

.login-form :deep(.el-input__wrapper) {
  min-height: 46px;
  border-radius: 12px;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.55) inset;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(100, 116, 139, 0.72) inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 3px rgba(33, 84, 226, 0.16);
}

.login-form :deep(.el-input__inner) {
  color: #0f172a;
}

.login-title {
  margin: 0;
  color: #0f172a;
  font-size: 28px;
  line-height: 1.1;
}

.login-subtitle {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 14px;
  line-height: 1.6;
}

.login-btn {
  width: 100%;
  min-height: 46px;
  margin-top: 6px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

@media (max-width: 980px) {
  .login-grid {
    grid-template-columns: 1fr;
  }

  .hero-panel {
    min-height: auto;
  }
}

@media (max-width: 720px) {
  .hero-metrics {
    grid-template-columns: 1fr;
  }

  .login-card :deep(.el-card__header) {
    padding: 24px 20px 16px;
  }

  .login-card :deep(.el-card__body) {
    padding: 0 20px 24px;
  }
}

@media (max-width: 520px) {
  .login-page {
    padding: 16px;
  }

  .hero-panel {
    padding: 24px 20px;
    border-radius: 22px;
  }

  .hero-title {
    font-size: 32px;
  }
}
</style>
