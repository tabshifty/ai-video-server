<script setup>
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import Layout from '../components/Layout.vue'
import { getAdminTasks } from '../api/admin'

const list = ref([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20 })
let timer = null

async function load() {
  const data = await getAdminTasks(query)
  list.value = data.items || []
  total.value = data.total_count || 0
}

function toNumber(value) {
  const n = Number(value)
  return Number.isFinite(n) ? n : NaN
}

function clampPercent(value) {
  if (!Number.isFinite(value)) return 0
  if (value < 0) return 0
  if (value > 100) return 100
  return Number(value.toFixed(2))
}

function resolveProgress(row) {
  const progress = toNumber(row.progress_percent)
  if (Number.isFinite(progress)) {
    return clampPercent(progress)
  }
  const source = toNumber(row.source_duration_seconds)
  const processed = toNumber(row.processed_seconds)
  if (source > 0 && Number.isFinite(processed) && processed >= 0) {
    return clampPercent((processed / source) * 100)
  }
  if (row.status === 'success') {
    return 100
  }
  return 0
}

function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return '--'
  const total = Math.floor(seconds)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) {
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function formatRemaining(row) {
  if (row.status !== 'running') return '--'
  return formatDuration(toNumber(row.remaining_seconds))
}

function formatElapsed(row) {
  if (!row.started_at) return '--'
  const start = Date.parse(row.started_at)
  if (!Number.isFinite(start)) return '--'
  let end = Date.now()
  if (row.finished_at) {
    const finished = Date.parse(row.finished_at)
    if (Number.isFinite(finished)) {
      end = finished
    }
  }
  if (end < start) return '--'
  return formatDuration((end - start) / 1000)
}

function progressStatus(row) {
  if (row.status === 'failed') return 'exception'
  if (row.status === 'success') return 'success'
  return ''
}

onMounted(async () => {
  await load()
  timer = setInterval(load, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">任务监控</h1>
          <p class="page-subtitle">每 5 秒自动刷新转码任务状态</p>
        </div>
      </div>

    <el-card class="soft-card">
      <div class="table-wrap">
      <el-table :data="list" border>
        <el-table-column prop="id" label="任务ID" width="90" />
        <el-table-column prop="video_id" label="视频ID" min-width="220" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column label="进度" min-width="220">
          <template #default="{ row }">
            <el-progress
              :stroke-width="14"
              :percentage="resolveProgress(row)"
              :status="progressStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="剩余时间" width="120">
          <template #default="{ row }">
            {{ formatRemaining(row) }}
          </template>
        </el-table-column>
        <el-table-column label="已耗时" width="120">
          <template #default="{ row }">
            {{ formatElapsed(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="retry_count" label="重试" width="80" />
        <el-table-column prop="error" label="错误信息" min-width="280" />
        <el-table-column prop="started_at" label="开始时间" width="180" />
        <el-table-column prop="progress_updated_at" label="进度更新时间" width="180" />
      </el-table>
      </div>
      <div class="action-row">
        <el-pagination v-model:current-page="query.page" :total="total" @current-change="load" />
      </div>
    </el-card>
    </div>
  </Layout>
</template>
