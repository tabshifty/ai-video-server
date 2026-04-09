<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import { getSystemLogs, systemCleanup } from '../api/admin'

const logs = ref([])
const loading = ref(false)

async function loadLogs() {
  loading.value = true
  try {
    const data = await getSystemLogs({ lines: 300 })
    logs.value = data.lines || []
  } finally {
    loading.value = false
  }
}

async function runCleanup() {
  await systemCleanup({ older_than_hours: 24 })
  ElMessage.success('清理任务已执行')
}
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">系统设置</h1>
          <p class="page-subtitle">执行清理任务并查看近期系统日志</p>
        </div>
      </div>

    <el-card class="soft-card">
      <el-button type="warning" @click="runCleanup">清理临时文件</el-button>
      <el-button @click="loadLogs" :loading="loading" style="margin-left:8px">刷新日志</el-button>
      <el-scrollbar height="520px" class="log-box">
        <pre class="log-text">{{ logs.join('\n') }}</pre>
      </el-scrollbar>
    </el-card>
    </div>
  </Layout>
</template>

<style scoped>
.log-box {
  margin-top: 12px;
  background: #111827;
  border: 1px solid #1f2937;
  border-radius: 12px;
}

.log-text {
  margin: 0;
  padding: 12px;
  white-space: pre-wrap;
  color: #e5e7eb;
  font-size: 12px;
  font-family: 'Fira Code', monospace;
}
</style>
