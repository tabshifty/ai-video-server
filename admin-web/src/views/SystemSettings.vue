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
    <el-card>
      <el-button type="warning" @click="runCleanup">清理临时文件</el-button>
      <el-button @click="loadLogs" :loading="loading" style="margin-left:8px">刷新日志</el-button>
      <el-scrollbar height="520px" style="margin-top: 12px; background:#0b1021; color:#e2e8f0; padding:12px">
        <pre style="white-space: pre-wrap; margin:0">{{ logs.join('\n') }}</pre>
      </el-scrollbar>
    </el-card>
  </Layout>
</template>
