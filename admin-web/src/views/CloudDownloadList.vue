<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import Toolbar from '../components/base/Toolbar.vue'
import {
  getCloudBrowserState,
  getCloudLoginQR,
  listCloudDownloads,
  retryCloudDownload
} from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'

const POLL_INTERVAL_MS = 5000
const ACTIVE_STATUSES = new Set(['pending', 'submitting', 'queued', 'ready', 'downloading'])
const STATUS_META = {
  pending: { label: '排队中', type: 'info' },
  submitting: { label: '提交中', type: 'info' },
  queued: { label: '115下载中', type: 'warning' },
  ready: { label: '待下载', type: 'warning' },
  downloading: { label: '下载中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' },
  blocked: { label: '需处理', type: 'danger' }
}
const BLOCKED_REASON_LABELS = {
  login_required: '需要登录 115',
  quota_exhausted: '配额不足',
  browser_down: '浏览器未运行',
  needs_attention: '需要人工确认'
}
const statusOptions = [
  { label: '全部状态', value: '' },
  ...Object.entries(STATUS_META).map(([value, meta]) => ({ label: meta.label, value }))
]

const list = ref([])
const total = ref(0)
const loading = ref(false)
const loaded = ref(false)
const loadError = ref('')
const query = reactive({ page: 1, page_size: 20, status: '' })
const browserState = ref({})
const browserStateError = ref('')
const retryingId = ref(0)
const qrDialogVisible = ref(false)
const qrImage = ref('')
const qrLoading = ref(false)
const qrError = ref('')
const qrRemainingSeconds = ref(0)
let timer = null
let qrTimer = null
let qrDeadline = 0
let qrSeq = 0
let loadSeq = 0

const initialLoading = computed(() => loading.value && !loaded.value)
const backgroundRefreshing = computed(() => loading.value && loaded.value)
const hasActiveTasks = computed(() => list.value.some((task) => ACTIVE_STATUSES.has(task.status)))
const hasStatusFilter = computed(() => query.status !== '')
const refreshIndicatorLabel = computed(() => {
  if (backgroundRefreshing.value) return '正在刷新'
  if (hasActiveTasks.value) return '每 5 秒自动刷新'
  return '自动刷新已停止'
})

function extractErrorMessage(error, fallback) {
  const responseMessage = error?.response?.data?.msg
  if (typeof responseMessage === 'string' && responseMessage.trim()) {
    return responseMessage.trim()
  }
  if (typeof error?.message === 'string' && error.message.trim()) {
    return error.message.trim()
  }
  return fallback
}

function syncPolling() {
  if (hasActiveTasks.value) {
    if (!timer) {
      timer = setInterval(() => load({ skipIfLoading: true }), POLL_INTERVAL_MS)
    }
    return
  }
  if (timer) {
    clearInterval(timer)
    timer = null
  }
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
    const data = await listCloudDownloads(params)
    if (seq !== loadSeq) {
      return
    }
    loadError.value = ''
    list.value = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total_count || 0)
    syncPolling()
  } catch (error) {
    if (seq !== loadSeq) {
      return
    }
    loadError.value = extractErrorMessage(error, '加载云下载任务失败')
  } finally {
    if (seq === loadSeq) {
      loaded.value = true
      loading.value = false
    }
  }
}

async function loadBrowserState() {
  try {
    const data = await getCloudBrowserState()
    browserState.value = data && typeof data === 'object' ? data : {}
    browserStateError.value = ''
  } catch (error) {
    browserStateError.value = extractErrorMessage(error, '获取 115 浏览器状态失败')
  }
}

async function refreshAll() {
  await Promise.all([load(), loadBrowserState()])
}

function applyStatusFilter() {
  query.page = 1
  list.value = []
  total.value = 0
  loaded.value = false
  loadError.value = ''
  load()
}

function clearStatusFilter() {
  query.status = ''
  applyStatusFilter()
}

function setPage(page) {
  query.page = page
  list.value = []
  total.value = 0
  loaded.value = false
  loadError.value = ''
  load()
}

function statusMeta(status) {
  return STATUS_META[status] || { label: status || '未知', type: 'info' }
}

function blockedReasonLabel(reason) {
  return BLOCKED_REASON_LABELS[reason] || reason || ''
}

function formatBytes(bytes) {
  const size = Number(bytes)
  if (!Number.isFinite(size) || size <= 0) return '--'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 100 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
}

function resolveProgress(row) {
  if (row?.progress_percent === null || row?.progress_percent === undefined || row?.progress_percent === '') {
    return row?.status === 'completed' ? 100 : 0
  }
  const value = Number(row.progress_percent)
  if (!Number.isFinite(value)) {
    return row?.status === 'completed' ? 100 : 0
  }
  if (value < 0) return 0
  if (value > 100) return 100
  return Number(value.toFixed(2))
}

function progressStatus(row) {
  if (row.status === 'failed') return 'exception'
  if (row.status === 'completed') return 'success'
  return ''
}

async function retryTask(row) {
  if (!row?.id || retryingId.value) return
  retryingId.value = row.id
  try {
    await retryCloudDownload(row.id)
    ElMessage.success('已提交重试')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '重试云下载任务失败'))
  } finally {
    retryingId.value = 0
  }
}

function stopQrTimer() {
  if (qrTimer) {
    clearInterval(qrTimer)
    qrTimer = null
  }
}

function syncQrCountdown() {
  const remaining = Math.max(0, Math.ceil((qrDeadline - Date.now()) / 1000))
  qrRemainingSeconds.value = remaining
  if (remaining <= 0) {
    stopQrTimer()
  }
}

function startQrCountdown(seconds) {
  stopQrTimer()
  const total = Math.max(0, Math.floor(Number(seconds) || 0))
  qrRemainingSeconds.value = total
  if (total <= 0) return
  qrDeadline = Date.now() + total * 1000
  qrTimer = setInterval(syncQrCountdown, 1000)
}

function formatCountdown(seconds) {
  const total = Math.max(0, Math.floor(Number(seconds) || 0))
  const minutes = Math.floor(total / 60)
  const rest = total % 60
  return `${String(minutes).padStart(2, '0')}:${String(rest).padStart(2, '0')}`
}

function resetQr() {
  qrSeq += 1
  stopQrTimer()
  qrImage.value = ''
  qrError.value = ''
  qrRemainingSeconds.value = 0
  qrLoading.value = false
}

async function loadQr() {
  // 二维码只保存在当前弹窗内存中，不写入 storage，也不会跨次复用。
  const seq = ++qrSeq
  qrLoading.value = true
  qrError.value = ''
  qrImage.value = ''
  qrRemainingSeconds.value = 0
  stopQrTimer()
  try {
    const data = await getCloudLoginQR()
    if (seq !== qrSeq || !qrDialogVisible.value) {
      return
    }
    const image = typeof data?.image === 'string' ? data.image.trim() : ''
    if (!image) {
      throw new Error('未获取到登录二维码')
    }
    qrImage.value = image
    startQrCountdown(data?.expires_in_seconds)
  } catch (error) {
    if (seq !== qrSeq) {
      return
    }
    qrError.value = extractErrorMessage(error, '获取登录二维码失败')
  } finally {
    if (seq === qrSeq) {
      qrLoading.value = false
    }
  }
}

function openQrDialog() {
  qrDialogVisible.value = true
  resetQr()
  loadQr()
}

onMounted(() => {
  load()
  loadBrowserState()
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
  stopQrTimer()
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-button class="cloud-refresh" :loading="loading" @click="refreshAll">立即刷新</el-button>
    </template>

    <div class="page-shell cloud-download-page" data-density="compact">
      <el-alert
        v-if="browserState.up === false"
        type="warning"
        :closable="false"
        show-icon
        title="115 浏览器未运行"
      >
        <template #default>请启动 115 浏览器后点击「立即刷新」。</template>
      </el-alert>
      <el-alert
        v-else-if="browserState.logged_in === false"
        type="warning"
        :closable="false"
        show-icon
        title="115 账号未登录"
      >
        <template #default>
          <div class="cloud-login-row">
            <span>登录后才能提交云下载任务。</span>
            <el-button size="small" type="primary" @click="openQrDialog">获取登录二维码</el-button>
          </div>
        </template>
      </el-alert>
      <el-alert v-if="browserStateError" type="error" :closable="false" :title="browserStateError">
        <template #default>
          <el-button link type="primary" @click="loadBrowserState">重试</el-button>
        </template>
      </el-alert>

      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default>
          <el-button link type="primary" @click="load">重试</el-button>
        </template>
      </el-alert>

      <el-skeleton v-if="initialLoading" :rows="10" animated />

      <SectionCard v-else-if="!loadError || list.length > 0" dense>
        <template #title>云下载任务</template>
        <template #actions>
          <el-tag effect="plain">共 {{ total }} 条</el-tag>
        </template>

        <Toolbar dense>
          <template #filters>
            <el-select
              v-model="query.status"
              class="cloud-status-filter"
              aria-label="云下载状态筛选"
              placeholder="状态筛选"
              @change="applyStatusFilter"
            >
              <el-option
                v-for="option in statusOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </template>
          <template #actions>
            <StatusIndicator
              :label="refreshIndicatorLabel"
              :tone="backgroundRefreshing ? 'info' : 'neutral'"
            />
          </template>
        </Toolbar>

        <EmptyState
          v-if="list.length === 0"
          :title="hasStatusFilter ? '当前筛选无结果' : '暂无云下载任务'"
          :description="hasStatusFilter ? '清除状态筛选后查看全部任务' : '从论坛资源页发送 ED2K 链接后会显示在这里'"
        >
          <template v-if="hasStatusFilter" #action>
            <el-button @click="clearStatusFilter">清除筛选</el-button>
          </template>
        </EmptyState>

        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="list" border>
              <el-table-column type="expand" width="44">
                <template #default="{ row }">
                  <dl class="cloud-detail">
                    <div class="cloud-detail__item">
                      <dt>错误信息</dt>
                      <dd>{{ row.error_message || '--' }}</dd>
                    </div>
                    <div class="cloud-detail__item">
                      <dt>阻塞原因</dt>
                      <dd>{{ blockedReasonLabel(row.blocked_reason) || '--' }}</dd>
                    </div>
                    <div class="cloud-detail__item">
                      <dt>本地路径</dt>
                      <dd class="cloud-detail__path">{{ row.local_path || '--' }}</dd>
                    </div>
                    <div class="cloud-detail__item">
                      <dt>pick_code</dt>
                      <dd>{{ row.pick_code || '--' }}</dd>
                    </div>
                    <div class="cloud-detail__item">
                      <dt>重试次数</dt>
                      <dd>{{ row.retry_count ?? 0 }}</dd>
                    </div>
                  </dl>
                </template>
              </el-table-column>
              <el-table-column label="文件名" min-width="220" show-overflow-tooltip>
                <template #default="{ row }">{{ row.declared_name || '--' }}</template>
              </el-table-column>
              <el-table-column label="大小" width="110">
                <template #default="{ row }">{{ formatBytes(row.declared_size) }}</template>
              </el-table-column>
              <el-table-column label="来源帖" min-width="200" show-overflow-tooltip>
                <template #default="{ row }">{{ row.post_title || '--' }}</template>
              </el-table-column>
              <el-table-column label="状态" min-width="140">
                <template #default="{ row }">
                  <div class="cloud-status-cell">
                    <el-tag size="small" effect="plain" :type="statusMeta(row.status).type">
                      {{ statusMeta(row.status).label }}
                    </el-tag>
                    <span v-if="row.blocked_reason" class="cloud-blocked-reason">
                      {{ blockedReasonLabel(row.blocked_reason) }}
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="进度" min-width="170">
                <template #default="{ row }">
                  <el-progress
                    :stroke-width="6"
                    :percentage="resolveProgress(row)"
                    :status="progressStatus(row)"
                  />
                </template>
              </el-table-column>
              <el-table-column label="创建时间" width="168">
                <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="96">
                <template #default="{ row }">
                  <el-button
                    v-if="row.status === 'failed' || row.status === 'blocked'"
                    size="small"
                    :loading="retryingId === row.id"
                    @click="retryTask(row)"
                  >
                    重试
                  </el-button>
                  <span v-else class="cloud-empty-action">--</span>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="action-row">
            <AdminTablePagination
              :current-page="query.page"
              :page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="total"
              :disabled="loading"
              @current-change="setPage"
            />
          </div>
        </template>
      </SectionCard>
    </div>

    <el-dialog
      v-model="qrDialogVisible"
      title="115 登录二维码"
      width="min(26rem, calc(100vw - 2rem))"
      @closed="resetQr"
    >
      <div class="cloud-qr">
        <el-alert v-if="qrError" type="error" :closable="false" :title="qrError" />
        <div v-if="qrImage" class="cloud-qr__image">
          <img :src="qrImage" alt="115 登录二维码" />
        </div>
        <el-skeleton v-else-if="qrLoading" :rows="4" animated />
        <p v-if="qrImage" class="cloud-qr__countdown">
          <template v-if="qrRemainingSeconds > 0">
            二维码剩余有效时间：<span class="tabular-num">{{ formatCountdown(qrRemainingSeconds) }}</span>
          </template>
          <span v-else class="cloud-qr__expired">二维码已过期，请刷新</span>
        </p>
        <p class="cloud-qr__hint">使用 115 App 扫码登录。二维码仅显示在当前弹窗中，不会保存。</p>
      </div>
      <template #footer>
        <el-button @click="qrDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="qrLoading" @click="loadQr">刷新二维码</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.cloud-download-page {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-4);
}

.cloud-status-filter {
  width: min(14rem, 60vw);
}

.cloud-login-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.cloud-status-cell {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.cloud-blocked-reason {
  color: var(--danger-600);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  overflow-wrap: anywhere;
}

.cloud-empty-action {
  color: var(--text-muted);
}

.cloud-detail {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding: var(--space-2) var(--space-4);
}

.cloud-detail__item {
  display: grid;
  grid-template-columns: 6rem minmax(0, 1fr);
  gap: var(--space-3);
  align-items: start;
}

.cloud-detail dt {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.cloud-detail dd {
  margin: 0;
  font-size: var(--text-small);
  line-height: var(--leading-small);
  overflow-wrap: anywhere;
}

.cloud-detail__path {
  font-family: var(--font-mono, monospace);
}

.cloud-qr {
  display: grid;
  gap: var(--space-3);
}

.cloud-qr__image {
  display: grid;
  width: min(18rem, 100%);
  place-items: center;
  justify-self: center;
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.cloud-qr__image img {
  display: block;
  width: 100%;
  height: auto;
}

.cloud-qr__countdown {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  justify-self: center;
}

.cloud-qr__expired {
  color: var(--danger-600);
}

.cloud-qr__hint {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  text-align: center;
  justify-self: center;
}

@media (max-width: 63.9375rem) {
  .cloud-refresh {
    min-height: 44px;
  }

  .cloud-status-filter {
    width: 100%;
    max-width: 20rem;
  }
}
</style>
