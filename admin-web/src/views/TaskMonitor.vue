<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import Toolbar from '../components/base/Toolbar.vue'
import MetricStrip from '../components/base/MetricStrip.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import { getAdminTasks } from '../api/admin'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const loaded = ref(false)
const loadError = ref('')
const query = reactive({ page: 1, page_size: 20, status: '' })
let timer = null
let loadSeq = 0

const queuedCount = computed(() => list.value.filter((item) => item.status === 'pending').length)
const runningCount = computed(() => list.value.filter((item) => item.status === 'running').length)
const failedCount = computed(() => list.value.filter((item) => item.status === 'failed').length)
const initialLoading = computed(() => loading.value && !loaded.value)
const backgroundRefreshing = computed(() => loading.value && loaded.value)
const hasStatusFilter = computed(() => query.status !== '')
const statusOptions = [
  { label: '全部', value: '' },
  { label: '排队', value: 'pending' },
  { label: '处理中', value: 'running' },
  { label: '已完成', value: 'success' },
  { label: '失败', value: 'failed' }
]
const summaryMetrics = computed(() => [
  { key: 'total', label: '任务总量', value: total.value, scope: query.status ? '当前筛选·全部页' : '全局' },
  { key: 'queued', label: '排队', value: queuedCount.value, scope: '本页', tone: 'info' },
  { key: 'running', label: '处理中', value: runningCount.value, scope: '本页', tone: 'warning' },
  { key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' }
])

function taskStatusTone(status) {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'warning'
  if (status === 'pending') return 'info'
  return 'neutral'
}

async function load(options = {}) {
  const { skipIfLoading = false } = options
  if (skipIfLoading && loading.value) {
    return
  }
  const seq = ++loadSeq
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
    if (seq !== loadSeq) {
      return
    }
    loadError.value = ''
    list.value = data.items || []
    total.value = data.total_count || 0
  } catch (error) {
    if (seq !== loadSeq) {
      return
    }
    loadError.value = error?.message || '加载任务失败'
  } finally {
    if (seq === loadSeq) {
      loaded.value = true
      loading.value = false
    }
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

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
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

function resetQueryIdentity() {
  list.value = []
  total.value = 0
  loaded.value = false
  loadError.value = ''
}

function setStatus(status) {
  query.status = status
  query.page = 1
  resetQueryIdentity()
  load()
}

function setPage(page) {
  query.page = page
  resetQueryIdentity()
  load()
}

function statusLabel(status) {
  if (!status) return '全部'
  if (status === 'pending') return '排队'
  if (status === 'running') return '处理中'
  if (status === 'success') return '已完成'
  if (status === 'failed') return '失败'
  return status
}

function taskTitle(row) {
  const title = String(row.video_title || '').trim()
  return title || '未命名视频'
}

onMounted(async () => {
  await load()
  timer = setInterval(() => load({ skipIfLoading: true }), 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-button class="task-refresh" :loading="loading" @click="load">立即刷新</el-button>
    </template>

    <div class="page-shell task-monitor-page" data-density="monitor">
      <Toolbar dense>
        <template #filters>
          <el-segmented
            class="status-filter"
            aria-label="任务状态筛选"
            :model-value="query.status"
            :options="statusOptions"
            @update:model-value="setStatus"
          />
        </template>
        <template #actions>
          <StatusIndicator
            :label="backgroundRefreshing ? '正在刷新' : '每 5 秒自动刷新'"
            :tone="backgroundRefreshing ? 'info' : 'neutral'"
          />
        </template>
      </Toolbar>

      <MetricStrip :items="summaryMetrics" aria-label="任务摘要" />

      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default>
          <el-button link type="primary" @click="load">重试</el-button>
        </template>
      </el-alert>

      <el-skeleton v-if="initialLoading" :rows="12" animated />

      <SectionCard v-else-if="!loadError || list.length > 0" dense>
        <template #title>任务列表</template>
        <EmptyState
          v-if="list.length === 0"
          :title="hasStatusFilter ? '当前筛选无结果' : '暂无任务'"
          :description="hasStatusFilter ? '清除状态筛选后查看全部任务' : '任务创建后会显示在这里'"
        >
          <template v-if="hasStatusFilter" #action>
            <el-button @click="setStatus('')">清除筛选</el-button>
          </template>
        </EmptyState>
        <template v-else>
          <div class="table-wrap">
            <el-table :data="list" border>
              <el-table-column prop="video_title" label="任务" min-width="260">
                <template #default="{ row }">
                  <div class="task-cell">
                    <strong>{{ taskTitle(row) }}</strong>
                    <span>任务 ID：{{ row.id }} · 视频 ID：{{ row.video_id || '--' }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="112">
                <template #default="{ row }">
                  <StatusIndicator :label="statusLabel(row.status)" :tone="taskStatusTone(row.status)" />
                </template>
              </el-table-column>
              <el-table-column label="进度" min-width="190">
                <template #default="{ row }">
                  <el-progress :stroke-width="6" :percentage="resolveProgress(row)" :status="progressStatus(row)" />
                </template>
              </el-table-column>
              <el-table-column label="剩余时间" width="112">
                <template #default="{ row }">
                  {{ formatRemaining(row) }}
                </template>
              </el-table-column>
              <el-table-column label="已耗时" width="112">
                <template #default="{ row }">
                  {{ formatElapsed(row) }}
                </template>
              </el-table-column>
              <el-table-column prop="retry_count" label="重试" width="72" />
              <el-table-column prop="error" label="错误" min-width="220">
                <template #default="{ row }">
                  <el-tooltip :content="row.error || '无错误'" placement="top">
                    <span class="task-error" tabindex="0" :aria-label="row.error || '无错误'">{{ row.error || '--' }}</span>
                  </el-tooltip>
                </template>
              </el-table-column>
              <el-table-column label="开始时间" width="168">
                <template #default="{ row }">{{ formatDateTime(row.started_at) }}</template>
              </el-table-column>
              <el-table-column label="进度更新时间" width="168">
                <template #default="{ row }">{{ formatDateTime(row.progress_updated_at) }}</template>
              </el-table-column>
            </el-table>
          </div>
          <div class="toolbar-row toolbar-row--end">
            <AdminTablePagination
              :current-page="query.page"
              :page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="total"
              @current-change="setPage"
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

.task-cell {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.task-cell strong,
.task-cell span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-cell strong {
  color: var(--text-primary);
}

.task-cell span {
  color: var(--text-secondary);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.task-error {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-error:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

@media (max-width: 63.9375rem) {
  .task-refresh {
    min-height: 44px;
  }

  .status-filter {
    max-width: 100%;
    overflow-x: auto;
    overscroll-behavior-x: contain;
  }

  .status-filter :deep(.el-segmented__group) {
    width: max-content;
    min-width: max-content;
  }

  :deep(.el-segmented__item) {
    min-width: 72px;
    min-height: 44px;
    flex: 0 0 auto;
  }

  .status-filter :deep(.el-segmented__item-label) {
    overflow: visible;
    text-overflow: clip;
    white-space: nowrap;
  }
}
</style>
