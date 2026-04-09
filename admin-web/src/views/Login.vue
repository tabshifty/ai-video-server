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
  <div style="height:100vh;display:flex;align-items:center;justify-content:center;background:linear-gradient(135deg,#0ea5e9,#22c55e)">
    <el-card style="width: 420px">
      <h2 style="margin-bottom: 18px">管理员登录</h2>
      <el-form @submit.prevent="submit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="admin" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="submit" style="width: 100%">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>
