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
    <div class="page page-shell">
      <section class="section-head">
        <div>
          <h1 class="page-title">系统设置</h1>
          <p class="page-subtitle">执行清理任务并查看近期系统日志</p>
        </div>
      </section>

      <section class="page-section">
        <el-card class="soft-card content-card">
          <template #header>
            <div class="card-title">控制中心</div>
          </template>
          <div class="ops-grid">
            <div class="ops-item">
              <div class="ops-name">临时文件清理</div>
              <p>清理超过 24 小时的临时文件，释放磁盘占用。</p>
              <el-button type="warning" @click="runCleanup">执行清理</el-button>
            </div>
            <div class="ops-item">
              <div class="ops-name">日志拉取</div>
              <p>拉取系统日志最近 300 行，用于排查上传与任务异常。</p>
              <el-button :loading="loading" @click="loadLogs">刷新日志</el-button>
            </div>
          </div>
        </el-card>
      </section>

      <section class="page-section">
        <el-card class="soft-card content-card log-panel">
          <template #header>
            <div class="log-head">
              <div>
                <div class="card-title">系统日志</div>
                <p>按时间顺序展示最近日志输出</p>
              </div>
              <el-button :loading="loading" @click="loadLogs">刷新</el-button>
            </div>
          </template>
          <el-scrollbar height="520px" class="log-box">
            <pre class="log-text">{{ logs.join('\n') }}</pre>
          </el-scrollbar>
        </el-card>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.page-shell {
  gap: 16px;
}

.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
}

.page-section {
  display: grid;
  gap: 12px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #7f1d1d;
}

.ops-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.ops-item {
  border: 1px solid rgba(136, 19, 55, 0.12);
  border-radius: 12px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.78);
  display: grid;
  gap: 10px;
}

.ops-name {
  font-size: 14px;
  font-weight: 600;
  color: #9f1239;
}

.ops-item p {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: #6b7280;
}

.log-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.log-head p {
  margin: 4px 0 0;
  font-size: 12px;
  color: #6b7280;
}

.log-box {
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

@media (max-width: 900px) {
  .ops-grid {
    grid-template-columns: 1fr;
  }
}
</style>
