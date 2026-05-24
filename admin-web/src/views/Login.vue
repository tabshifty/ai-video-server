<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/base/PageHeader.vue'
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
  <main class="login-page">
    <section class="login-card" aria-label="管理员登录">
      <div class="login-brand">
        <div class="login-brand__mark">VS</div>
        <div class="login-brand__copy">
          <span>Video Server</span>
          <strong>管理后台</strong>
        </div>
      </div>

      <PageHeader title="管理员登录" subtitle="请使用具有后台权限的账号继续" />

      <el-form class="login-form" label-position="top" @submit.prevent="submit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="admin" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-button type="primary" :loading="loading" class="login-btn" @click="submit">登录</el-button>
      </el-form>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: grid;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--blue-50) 72%, transparent), transparent 48%),
    var(--bg-canvas);
}

.login-card {
  width: min(100%, calc(var(--space-12) * 9));
  display: grid;
  gap: var(--space-6);
  padding: var(--space-6);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  background: var(--bg-surface);
  box-shadow: var(--shadow-lg);
}

.login-brand {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-3);
}

.login-brand__mark {
  display: grid;
  width: calc(var(--space-8) + var(--space-2));
  height: calc(var(--space-8) + var(--space-2));
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  color: var(--primary);
  background: var(--primary-soft);
  font-weight: 600;
}

.login-brand__copy {
  display: grid;
  min-width: 0;
}

.login-brand__copy span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.login-brand__copy strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
}

.login-form :deep(.el-form-item) {
  margin-bottom: var(--space-4);
}

.login-form :deep(.el-form-item__label) {
  padding-bottom: var(--space-2);
  color: var(--text-primary);
  font-weight: 600;
}

.login-form :deep(.el-input__wrapper) {
  min-height: calc(var(--space-12) - var(--space-1));
}

.login-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.login-btn {
  width: 100%;
  min-height: calc(var(--space-12) - var(--space-1));
  margin-top: var(--space-1);
  font-weight: 600;
}

@media (max-width: 32.5rem) {
  .login-page {
    padding: var(--space-4);
  }

  .login-card {
    padding: var(--space-5);
  }
}
</style>
