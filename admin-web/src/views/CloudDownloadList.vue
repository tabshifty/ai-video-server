<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import Toolbar from '../components/base/Toolbar.vue'
import CloudDownloadImportDialog from '../components/CloudDownloadImportDialog.vue'
import {
  deleteCloudDownloadCloudFile,
  getCloudBrowserState,
  getCloudLoginQR,
  listCloudDownloads,
  previewCloudDownloadImport,
  retryCloudDownload
} from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'
import {
  hasActiveImportTasks,
  importDestination,
  resolveCloudImportAction
} from './cloudDownloadImport.helpers'
import {
  canDeleteCloudDownloadFile,
  cloudDownloadFileName,
  cloudFileDeletionActionLabel,
  cloudFileDeletionMeta,
  hasActiveCloudFileDeletions
} from './cloudDownloadFileDeletion.helpers'
import { extractApiErrorMessage } from './toolbox.helpers'

const POLL_INTERVAL_MS = 5000
const ACTIVE_STATUSES = new Set(['pending', 'submitting', 'queued', 'ready', 'downloading'])
const router = useRouter()
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
const previewingImportId = ref(0)
const importDialogVisible = ref(false)
const importTask = ref(null)
const importPreview = ref(null)
const deletingCloudFileId = ref(0)
const deleteCloudFileDialogVisible = ref(false)
const deleteCloudFileTask = ref(null)
const deleteCloudFileError = ref('')
const blockedCloudFileDeletionIds = ref([])
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
const hasActiveImport = computed(() => hasActiveImportTasks(list.value))
const hasActiveFileDeletion = computed(() => hasActiveCloudFileDeletions(list.value))
const shouldPoll = computed(() => hasActiveTasks.value || hasActiveImport.value || hasActiveFileDeletion.value)
const hasStatusFilter = computed(() => query.status !== '')
const refreshIndicatorLabel = computed(() => {
  if (backgroundRefreshing.value) return '正在刷新'
  if (hasActiveFileDeletion.value) return '云盘清理中，每 5 秒自动刷新'
  if (hasActiveImport.value) return '导入中，每 5 秒自动刷新'
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
  if (shouldPoll.value) {
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

async function startImport(row) {
  if (!row?.id || previewingImportId.value) return
  previewingImportId.value = row.id
  try {
    const preview = await previewCloudDownloadImport(row.id)
    if (!preview || (preview.kind !== 'video' && preview.kind !== 'archive')) {
      ElMessage.error('该文件类型不支持入媒体库')
      return
    }
    const existingKind = row.import?.kind
    if (existingKind && existingKind !== preview.kind) {
      ElMessage.error('文件类型与已有导入记录不一致，无法重试')
      return
    }
    importTask.value = row
    importPreview.value = preview
    importDialogVisible.value = true
  } catch (error) {
    ElMessage.error(extractApiErrorMessage(error, '无法入媒体库'))
  } finally {
    previewingImportId.value = 0
  }
}

function openImportResult(row) {
  const destination = importDestination(row)
  if (destination) {
    router.push(destination)
  }
}

async function handleImportSubmitted(importInfo) {
  if (importTask.value && importInfo && typeof importInfo === 'object') {
    importTask.value.import = importInfo
  }
  syncPolling()
  await load()
}

function resetImportDialog() {
  importTask.value = null
  importPreview.value = null
}

function canDeleteCloudFileRow(row) {
  if (!row?.id || blockedCloudFileDeletionIds.value.includes(String(row.id))) return false
  return canDeleteCloudDownloadFile(row)
}

function openDeleteCloudFileDialog(row) {
  if (deletingCloudFileId.value || !canDeleteCloudFileRow(row)) return
  deleteCloudFileTask.value = row
  deleteCloudFileError.value = ''
  deleteCloudFileDialogVisible.value = true
}

function handleCloudFileDeletionCommand(command, row) {
  if (command === 'delete-cloud-file') {
    openDeleteCloudFileDialog(row)
  }
}

function applyDeletedTask(updated) {
  if (!updated || typeof updated !== 'object') return
  const index = list.value.findIndex((task) => task.id === updated.id)
  if (index >= 0) {
    list.value[index] = { ...list.value[index], ...updated }
  }
}

async function confirmDeleteCloudFile() {
  const task = deleteCloudFileTask.value
  if (!task?.id || deletingCloudFileId.value) return
  deletingCloudFileId.value = task.id
  deleteCloudFileError.value = ''
  try {
    const data = await deleteCloudDownloadCloudFile(task.id)
    applyDeletedTask(data?.task)
    if (deleteCloudFileTask.value?.id === task.id) {
      deleteCloudFileDialogVisible.value = false
    }
    ElMessage.success('已移入 115 回收站')
    await load()
  } catch (error) {
    let message = extractErrorMessage(error, '移入 115 回收站失败')
    if (error?.code === 'ECONNABORTED' || error?.code === 'ETIMEDOUT' || /timeout/i.test(message)) {
      message = '请求超时，可能仍在处理；正在重新确认结果，请勿重复操作'
    }
    if (deleteCloudFileTask.value?.id === task.id) {
      deleteCloudFileError.value = message
    }
    ElMessage.error(message)
    const errorCode = Number(error?.code ?? error?.response?.data?.code)
    if (errorCode === 410 || message.includes('本地文件不存在')) {
      blockedCloudFileDeletionIds.value = [...new Set([...blockedCloudFileDeletionIds.value, String(task.id)])]
    }
    if (message.includes('登录')) {
      await loadBrowserState()
      openQrDialog()
    } else if (message.includes('浏览器')) {
      await loadBrowserState()
    }
    await load()
  } finally {
    deletingCloudFileId.value = 0
  }
}

function resetDeleteCloudFileDialog() {
  deleteCloudFileTask.value = null
  deleteCloudFileError.value = ''
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
              <el-table-column label="来源" min-width="200" show-overflow-tooltip>
                <template #default="{ row }">
                  <el-tag v-if="row.source === 'toolbox'" size="small" effect="plain">工具箱</el-tag>
                  <el-tag v-else-if="row.source === 'manual'" size="small" effect="plain" type="info">手动</el-tag>
                  <span v-else>{{ row.post_title || '原帖已删除' }}</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" min-width="140">
                <template #default="{ row }">
                  <div class="cloud-status-cell">
                    <el-tag size="small" effect="plain" :type="statusMeta(row.status).type">
                      {{ statusMeta(row.status).label }}
                    </el-tag>
                    <el-tag v-if="row.blocked_reason" size="small" effect="plain" type="danger">
                      {{ blockedReasonLabel(row.blocked_reason) }}
                    </el-tag>
                    <el-tooltip
                      v-if="cloudFileDeletionMeta(row)?.detail"
                      :content="cloudFileDeletionMeta(row).detail"
                      placement="top"
                    >
                      <el-tag size="small" effect="plain" :type="cloudFileDeletionMeta(row).type">
                        {{ cloudFileDeletionMeta(row).label }}
                      </el-tag>
                    </el-tooltip>
                    <el-tag
                      v-else-if="cloudFileDeletionMeta(row)"
                      size="small"
                      effect="plain"
                      :type="cloudFileDeletionMeta(row).type"
                    >
                      {{ cloudFileDeletionMeta(row).label }}
                    </el-tag>
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
              <el-table-column label="操作" width="160">
                <template #default="{ row }">
                  <div class="cloud-actions">
                    <template v-if="resolveCloudImportAction(row) === 'import'">
                      <el-button
                        size="small"
                        type="primary"
                        plain
                        :loading="previewingImportId === row.id"
                        @click="startImport(row)"
                      >
                        入媒体库
                      </el-button>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'missing'">
                      <span class="cloud-empty-action">文件不存在</span>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'importing'">
                      <el-button size="small" disabled>导入中</el-button>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'imported'">
                      <el-button link type="success" @click="openImportResult(row)">已导入</el-button>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'existing'">
                      <el-button link @click="openImportResult(row)">已存在</el-button>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'retry-import'">
                      <el-tooltip
                        :content="row.import?.status === 'failed'
                          ? (row.import?.error_message || '导入失败，点击重试')
                          : '导入任务可能已中断，点击重新提交'"
                        placement="top"
                      >
                        <el-button
                          size="small"
                          type="danger"
                          plain
                          :loading="previewingImportId === row.id"
                          @click="startImport(row)"
                        >
                          {{ row.import?.status === 'failed' ? '重试导入' : '重新导入' }}
                        </el-button>
                      </el-tooltip>
                    </template>
                    <template v-else-if="resolveCloudImportAction(row) === 'retry-download'">
                      <el-button
                        size="small"
                        :loading="retryingId === row.id"
                        @click="retryTask(row)"
                      >
                        重试
                      </el-button>
                    </template>
                    <span v-else class="cloud-empty-action">--</span>
                    <el-dropdown
                      v-if="canDeleteCloudFileRow(row)"
                      trigger="click"
                      placement="bottom-end"
                      popper-class="cloud-file-actions-popper"
                      @command="handleCloudFileDeletionCommand($event, row)"
                    >
                      <el-button
                        :icon="MoreFilled"
                        circle
                        size="small"
                        aria-label="更多云下载操作"
                        title="更多云下载操作"
                      />
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="delete-cloud-file">
                            {{ cloudFileDeletionActionLabel(row) }}
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                    <el-button
                      v-else-if="row.file_deletion?.status === 'deleting'"
                      size="small"
                      disabled
                    >
                      移入中
                    </el-button>
                  </div>
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

    <CloudDownloadImportDialog
      v-model="importDialogVisible"
      :task="importTask"
      :preview="importPreview"
      @submitted="handleImportSubmitted"
      @closed="resetImportDialog"
    />

    <el-dialog
      v-model="deleteCloudFileDialogVisible"
      title="移入 115 回收站"
      width="min(32rem, calc(100vw - 2rem))"
      destroy-on-close
      :close-on-click-modal="!deletingCloudFileId"
      :close-on-press-escape="!deletingCloudFileId"
      :show-close="!deletingCloudFileId"
      @closed="resetDeleteCloudFileDialog"
    >
      <div class="cloud-file-delete">
        <p class="cloud-file-delete__lead">
          确认将以下 115 云盘源文件移入回收站？
        </p>
        <dl class="cloud-file-delete__file">
          <div>
            <dt>文件名</dt>
            <dd>{{ cloudDownloadFileName(deleteCloudFileTask) }}</dd>
          </div>
          <div>
            <dt>大小</dt>
            <dd>{{ formatBytes(deleteCloudFileTask?.declared_size) }}</dd>
          </div>
        </dl>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          title="仅影响 115 云盘源文件，不会删除本地文件或导入记录；可在 115 回收站恢复。移入后不能从 115 重新下载。"
        />
        <el-alert
          v-if="deleteCloudFileError"
          type="error"
          :closable="false"
          :title="deleteCloudFileError"
        />
      </div>
      <template #footer>
        <el-button :disabled="Boolean(deletingCloudFileId)" @click="deleteCloudFileDialogVisible = false">
          取消
        </el-button>
        <el-button
          type="danger"
          :loading="deletingCloudFileId === deleteCloudFileTask?.id"
          :disabled="Boolean(deletingCloudFileId)"
          @click="confirmDeleteCloudFile"
        >
          移入回收站
        </el-button>
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
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.cloud-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.cloud-actions :deep(.el-button) {
  margin-left: 0;
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

.cloud-file-delete {
  display: grid;
  gap: var(--space-3);
}

.cloud-file-delete__lead {
  margin: 0;
  color: var(--text-secondary);
  line-height: var(--leading-small);
}

.cloud-file-delete__file {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.cloud-file-delete__file > div {
  display: grid;
  grid-template-columns: 4rem minmax(0, 1fr);
  gap: var(--space-3);
}

.cloud-file-delete__file dt {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.cloud-file-delete__file dd {
  margin: 0;
  font-size: var(--text-small);
  line-height: var(--leading-small);
  overflow-wrap: anywhere;
}

@media (max-width: 63.9375rem) {
  .cloud-refresh {
    min-height: 44px;
  }

  .cloud-actions :deep(.el-button) {
    min-width: 44px;
    min-height: 44px;
  }

  .cloud-status-filter {
    width: 100%;
    max-width: 20rem;
  }

  :global(.cloud-file-actions-popper .el-dropdown-menu__item) {
    min-height: 44px;
  }
}
</style>
