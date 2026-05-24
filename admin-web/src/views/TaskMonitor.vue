<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import StatCard from '../components/base/StatCard.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { getAdminTasks } from '../api/admin'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const query = reactive({ page: 1, page_size: 20, status: '' })
let timer = null

const queuedCount = computed(() => list.value.filter((item) => item.status === 'pending').length)
const runningCount = computed(() => list.value.filter((item) => item.status === 'running').length)
const failedCount = computed(() => list.value.filter((item) => item.status === 'failed').length)
const hasTasks = computed(() => list.value.length > 0)

async function load() {
  loading.value = true
  try {
    const params = {
      page: query.page,
      page_size: query.page_size
    }
    if (query.status) {
      params.status = query.status
    }
    const data = await getAdminTasks(params)
    list.value = data.items || []
    total.value = data.total_count || 0
  } catch (error) {
    ElMessage.error(error?.message || '加载任务失败')
  } finally {
    loading.value = false
  }
}

function toNumber(value) {
  if (value === null || value === undefined || value === '') {
    return NaN
  }
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
  const totalSeconds = Math.floor(seconds)
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60
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

function setStatus(status) {
  query.status = status
  query.page = 1
  load()
}

function statusLabel(status) {
  if (!status) return '全部'
  if (status === 'pending') return '排队'
  if (status === 'running') return '处理中'
  if (status === 'failed') return '失败'
  return status
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
    <div class="page-shell task-monitor-page">
      <PageHeader title="任务监控" subtitle="每 5 秒自动刷新转码任务状态">
        <template #actions>
          <el-button :loading="loading" @click="load">立即刷新</el-button>
        </template>
      </PageHeader>

      <Toolbar>
        <template #filters>
          <el-button-group class="status-group">
            <el-button :type="query.status === '' ? 'primary' : ''" @click="setStatus('')">全部</el-button>
            <el-button :type="query.status === 'pending' ? 'primary' : ''" @click="setStatus('pending')">排队</el-button>
            <el-button :type="query.status === 'running' ? 'primary' : ''" @click="setStatus('running')">处理中</el-button>
            <el-button :type="query.status === 'failed' ? 'primary' : ''" @click="setStatus('failed')">失败</el-button>
          </el-button-group>
        </template>
        <template #actions>
          <el-tag effect="plain">自动刷新：5 秒</el-tag>
        </template>
      </Toolbar>

      <section class="metric-grid">
        <StatCard label="队列" :value="queuedCount" />
        <StatCard label="处理中" :value="runningCount" />
        <StatCard label="失败" :value="failedCount" />
      </section>

      <SectionCard>
        <template #title>任务列表</template>
        <template #description>每 5 秒自动刷新，也可手动立即刷新</template>
        <EmptyState
          v-if="!hasTasks"
          title="暂无任务"
          description="当前筛选条件下没有任务"
        >
          <template #action>
            <el-button :loading="loading" @click="load">立即刷新</el-button>
          </template>
        </EmptyState>
        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="list" border>
              <el-table-column prop="id" label="任务ID" width="90" />
              <el-table-column prop="video_id" label="视频ID" min-width="220" />
              <el-table-column prop="status" label="状态" width="120">
                <template #default="{ row }">{{ statusLabel(row.status) }}</template>
              </el-table-column>
              <el-table-column label="进度" min-width="220">
                <template #default="{ row }">
                  <el-progress :stroke-width="14" :percentage="resolveProgress(row)" :status="progressStatus(row)" />
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
          <div class="toolbar-row toolbar-row--end">
            <AdminTablePagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="total"
              @current-change="load"
            />
          </div>
        </template>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.task-monitor-page {
  display: grid;
  gap: var(--space-4);
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-4);
}

.status-group {
  display: inline-flex;
}

@media (max-width: 48rem) {
  .metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
