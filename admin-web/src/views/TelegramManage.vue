<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import {
  CircleCheck,
  Connection,
  Key,
  Link,
  Refresh,
  RefreshRight,
  VideoPause,
  VideoPlay,
  WarningFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import MetricStrip from '../components/base/MetricStrip.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import {
  cancelAdminTelegramAuthorization,
  confirmAdminTelegramSource,
  getAdminTelegramAudits,
  getAdminTelegramAuthorization,
  getAdminTelegramSourceProgress,
  getAdminTelegramSources,
  getAdminTelegramStatus,
  issueAdminTelegramConfirmation,
  pauseAdminTelegramSource,
  previewAdminTelegramSource,
  recoverAdminTelegramSource,
  resumeAdminTelegramSource,
  startAdminTelegramPhoneAuthorization,
  startAdminTelegramQRAuthorization,
  startAdminTelegramSourceBackfill,
  submitAdminTelegramAuthorizationCode,
  submitAdminTelegramAuthorizationPassword
} from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'
import {
  getTelegramPollingInterval,
  isPrivateTelegramInviteReference,
  isTelegramAuthorizationActive,
  shouldPollTelegramPage
} from './telegramManage.helpers'

const status = ref(null)
const sources = ref([])
const sourceProgress = ref({})
const audits = ref([])
const auditTotal = ref(0)
const authorization = ref(null)
const authorizationStartedHere = ref(false)
const pageLoaded = ref(false)
const refreshing = ref(false)
const loadError = ref('')
const sourceBusy = ref(false)
const sourceActionID = ref('')
const authorizationBusy = ref(false)
const confirmationBusy = ref(false)
const sourceInput = ref('')
const sourcePreview = ref(null)
const sourceError = ref('')
const authorizationForm = reactive({ code: '', password: '' })
const auditQuery = reactive({ page: 1, page_size: 10 })
const confirmation = reactive({
  visible: false,
  action: '',
  title: '',
  detail: '',
  acknowledged: false,
  password: '',
  execute: null
})

let pageActive = true
let refreshInFlight = false
let refreshSequence = 0
let pollTimer = null

const activeAuthorization = computed(() => authorization.value || status.value?.authorization || null)
const authorizationActive = computed(() => isTelegramAuthorizationActive(activeAuthorization.value))
const account = computed(() => status.value?.account || {})
const heartbeat = computed(() => status.value?.heartbeat || {})
const accountBound = computed(() => Boolean(
  account.value?.telegram_user_id || account.value?.authorized_at
))
const sourceErrorCount = computed(() => sources.value.filter((item) => item.sync_status === 'error').length)
const initialLoading = computed(() => refreshing.value && !pageLoaded.value)
const canSubmitAuthorization = computed(() => {
  const ownerID = String(activeAuthorization.value?.owner_id || '').trim()
  return authorizationStartedHere.value || (ownerID !== '' && ownerID === currentAdminID())
})
const summaryMetrics = computed(() => [
  {
    key: 'account',
    label: '采集账号',
    value: accountStatusText(status.value?.account_status || account.value?.status),
    scope: accountDisplayName.value,
    tone: accountStatusTone(status.value?.account_status || account.value?.status)
  },
  {
    key: 'ingestor',
    label: '采集器',
    value: ingestorStatusText(status.value?.ingestor_status || heartbeat.value?.status),
    scope: heartbeat.value?.last_seen_at ? `心跳 ${formatDateTime(heartbeat.value.last_seen_at)}` : '暂无心跳',
    tone: ingestorStatusTone(status.value?.ingestor_status || heartbeat.value?.status)
  },
  {
    key: 'sources',
    label: '来源',
    value: sources.value.length,
    scope: '已配置',
    tone: 'info'
  },
  {
    key: 'source-errors',
    label: '来源异常',
    value: sourceErrorCount.value,
    scope: sourceErrorCount.value > 0 ? '需要处理' : '无异常',
    tone: sourceErrorCount.value > 0 ? 'danger' : 'success'
  }
])
const accountDisplayName = computed(() => {
  const name = [account.value?.first_name, account.value?.last_name]
    .map((item) => String(item || '').trim())
    .filter(Boolean)
    .join(' ')
  return name || account.value?.username || account.value?.phone_masked || '未绑定'
})

function extractErrorMessage(error, fallback) {
  const responseMessage = error?.response?.data?.msg
  if (typeof responseMessage === 'string' && responseMessage.trim()) return responseMessage.trim()
  if (typeof error?.message === 'string' && error.message.trim()) return error.message.trim()
  return fallback
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
}

function accountStatusText(value) {
  const labels = {
    unconfigured: '未配置',
    authorizing: '授权中',
    authorized: '已授权',
    reauthorizing: '重新授权中',
    error: '异常'
  }
  return labels[value] || value || '未知'
}

function accountStatusTone(value) {
  if (value === 'authorized') return 'success'
  if (value === 'authorizing' || value === 'reauthorizing') return 'warning'
  if (value === 'error') return 'danger'
  if (value === 'unconfigured') return 'info'
  return 'neutral'
}

function ingestorStatusText(value) {
  const labels = {
    running: '运行中',
    authorizing: '授权维护中',
    draining: '排空中',
    error: '异常',
    stopped: '未运行'
  }
  return labels[value] || value || '未知'
}

function ingestorStatusTone(value) {
  if (value === 'running') return 'success'
  if (value === 'authorizing' || value === 'draining') return 'warning'
  if (value === 'error') return 'danger'
  if (value === 'stopped') return 'info'
  return 'neutral'
}

function authorizationStatusText(value) {
  const labels = {
    pending: '准备中',
    awaiting_code: '等待验证码',
    awaiting_password: '等待二次验证',
    scanning: '等待扫码',
    succeeded: '已完成',
    failed: '失败',
    cancelled: '已取消',
    expired: '已过期'
  }
  return labels[value] || value || '未知'
}

function authorizationStatusTone(value) {
  if (value === 'succeeded') return 'success'
  if (value === 'awaiting_code' || value === 'awaiting_password' || value === 'scanning' || value === 'pending') return 'warning'
  if (value === 'failed' || value === 'expired') return 'danger'
  return 'neutral'
}

function sourceStatusText(value) {
  const labels = {
    pending: '等待同步',
    backfilling: '历史回填中',
    live: '实时采集',
    paused: '已暂停',
    error: '异常'
  }
  return labels[value] || value || '未知'
}

function sourceStatusTone(value) {
  if (value === 'live') return 'success'
  if (value === 'pending' || value === 'backfilling') return 'warning'
  if (value === 'error') return 'danger'
  if (value === 'paused') return 'info'
  return 'neutral'
}

function auditResultText(value) {
  if (value === 'succeeded') return '成功'
  if (value === 'failed') return '失败'
  return value || '--'
}

function auditResultTone(value) {
  if (value === 'succeeded') return 'success'
  if (value === 'failed') return 'danger'
  return 'neutral'
}

function sourceTitle(source) {
  return String(source?.title || '').trim() || source?.username || String(source?.chat_id || '--')
}

function sourceProgressSummary(source) {
  const counts = sourceProgress.value[source.id]?.counts || {}
  const imported = toCount(counts.imported) + toCount(counts.duplicate)
  const waiting = toCount(counts.discovered) + toCount(counts.queued) + toCount(counts.downloading) + toCount(counts.importing)
  const failed = toCount(counts.failed)
  return `已完成 ${imported} · 处理中 ${waiting} · 失败 ${failed}`
}

function toCount(value) {
  const count = Number(value || 0)
  return Number.isFinite(count) && count > 0 ? count : 0
}

function previewStatusText(preview) {
  if (preview?.requires_approval) return '需审批加入'
  if (preview?.requires_join) return '确认后加入'
  return '可添加'
}

function previewStatusTone(preview) {
  if (preview?.requires_approval || preview?.requires_join) return 'warning'
  return 'success'
}

function auditSummary(value) {
  if (!value) return '--'
  let summary = value
  if (typeof summary === 'string') {
    try {
      summary = JSON.parse(summary)
    } catch (_) {
      return summary.trim() || '--'
    }
  }
  if (!summary || typeof summary !== 'object' || Array.isArray(summary)) return '--'
  const labels = {
    chat_id: '群组 ID',
    chat_type: '来源类型',
    requires_join: '需加入',
    requires_approval: '需审批',
    enabled: '启用',
    sync_status: '同步状态'
  }
  const parts = Object.entries(summary)
    .filter(([, item]) => ['string', 'number', 'boolean'].includes(typeof item))
    .map(([key, item]) => `${labels[key] || key}：${String(item)}`)
  return parts.join(' · ') || '--'
}

function currentAdminID() {
  if (typeof window === 'undefined') return ''
  try {
    const token = window.localStorage.getItem('admin_token') || ''
    const payload = token.split('.')[1]
    if (!payload) return ''
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded = JSON.parse(window.atob(normalized))
    return String(decoded?.uid || '').trim()
  } catch (_) {
    return ''
  }
}

function updateAuthorization(nextAuthorization, startedHere = false) {
  if (!nextAuthorization?.id) return
  authorization.value = nextAuthorization
  if (startedHere) authorizationStartedHere.value = true
  schedulePolling()
}

function resetAuthorizationInputs() {
  authorizationForm.code = ''
  authorizationForm.password = ''
}

function clearPollTimer() {
  if (pollTimer !== null) {
    window.clearTimeout(pollTimer)
    pollTimer = null
  }
}

function pageDocumentVisible() {
  return typeof document === 'undefined' || !document.hidden
}

function schedulePolling() {
  clearPollTimer()
  if (!shouldPollTelegramPage({ pageActive, documentVisible: pageDocumentVisible() })) return
  const delay = getTelegramPollingInterval(activeAuthorization.value)
  pollTimer = window.setTimeout(() => {
    pollTimer = null
    void loadDashboard({ silent: true })
  }, delay)
}

async function loadDashboard({ silent = false } = {}) {
  if (refreshInFlight || !pageActive) return
  clearPollTimer()
  refreshInFlight = true
  refreshing.value = true
  const sequence = ++refreshSequence
  const authorizationID = activeAuthorization.value?.id
  const authorizationRequest = authorizationID && isTelegramAuthorizationActive(activeAuthorization.value)
    ? getAdminTelegramAuthorization(authorizationID)
    : Promise.resolve(null)
  try {
    const [statusResult, sourcesResult, auditsResult, authorizationResult] = await Promise.allSettled([
      getAdminTelegramStatus(),
      getAdminTelegramSources(),
      getAdminTelegramAudits({ page: auditQuery.page, page_size: auditQuery.page_size }),
      authorizationRequest
    ])
    if (sequence !== refreshSequence || !pageActive) return

    const errors = []
    if (statusResult.status === 'fulfilled') {
      status.value = statusResult.value || null
      if (!authorization.value && statusResult.value?.authorization?.id) {
        updateAuthorization(statusResult.value.authorization)
      }
    } else {
      errors.push(extractErrorMessage(statusResult.reason, '读取 Telegram 运行状态失败'))
    }

    if (sourcesResult.status === 'fulfilled') {
      const items = Array.isArray(sourcesResult.value?.items) ? sourcesResult.value.items : []
      sources.value = items
      const progressResults = await Promise.allSettled(items.map((item) => getAdminTelegramSourceProgress(item.id)))
      if (sequence !== refreshSequence || !pageActive) return
      const nextProgress = {}
      progressResults.forEach((result, index) => {
        if (result.status === 'fulfilled' && items[index]?.id) {
          nextProgress[items[index].id] = result.value || {}
        }
      })
      sourceProgress.value = nextProgress
    } else {
      errors.push(extractErrorMessage(sourcesResult.reason, '读取 Telegram 来源失败'))
    }

    if (auditsResult.status === 'fulfilled') {
      audits.value = Array.isArray(auditsResult.value?.items) ? auditsResult.value.items : []
      auditTotal.value = Number(auditsResult.value?.total || 0)
    } else {
      errors.push(extractErrorMessage(auditsResult.reason, '读取 Telegram 审计失败'))
    }

    if (authorizationResult.status === 'fulfilled' && authorizationResult.value?.id) {
      updateAuthorization(authorizationResult.value)
    } else if (authorizationResult.status === 'rejected') {
      errors.push(extractErrorMessage(authorizationResult.reason, '读取 Telegram 授权状态失败'))
    }
    loadError.value = errors.join('；')
  } catch (error) {
    if (sequence === refreshSequence && pageActive) {
      loadError.value = extractErrorMessage(error, silent ? '刷新 Telegram 管理状态失败' : '加载 Telegram 管理状态失败')
    }
  } finally {
    if (sequence === refreshSequence) {
      refreshInFlight = false
      refreshing.value = false
      pageLoaded.value = true
      schedulePolling()
    }
  }
}

function openHighRiskConfirmation({ action, title, detail, execute }) {
  confirmation.action = action
  confirmation.title = title
  confirmation.detail = detail
  confirmation.acknowledged = false
  confirmation.password = ''
  confirmation.execute = execute
  confirmation.visible = true
}

function resetHighRiskConfirmation() {
  confirmation.visible = false
  confirmation.action = ''
  confirmation.title = ''
  confirmation.detail = ''
  confirmation.acknowledged = false
  confirmation.password = ''
  confirmation.execute = null
}

async function submitHighRiskConfirmation() {
  if (!confirmation.acknowledged || !confirmation.execute) {
    ElMessage.warning('请确认继续该操作')
    return
  }
  const action = confirmation.action
  const execute = confirmation.execute
  const password = confirmation.password
  confirmationBusy.value = true
  confirmation.password = ''
  try {
    const ticket = await issueAdminTelegramConfirmation({
      action,
      current_password: password,
      confirmed: true
    })
    confirmation.visible = false
    confirmation.execute = null
    await execute(ticket?.ticket || '')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, 'Telegram 高风险确认失败'))
  } finally {
    confirmationBusy.value = false
    resetHighRiskConfirmation()
  }
}

function startPhoneAuthorization() {
  openHighRiskConfirmation({
    action: 'authorization_phone',
    title: '开始手机号授权',
    detail: '将暂停读取并创建短期授权会话。',
    execute: async (ticket) => {
      authorizationBusy.value = true
      try {
        const view = await startAdminTelegramPhoneAuthorization({ confirmation_ticket: ticket })
        updateAuthorization(view, true)
        resetAuthorizationInputs()
        ElMessage.success('已开始手机号授权')
      } finally {
        authorizationBusy.value = false
      }
    }
  })
}

function startQRAuthorization() {
  if (!accountBound.value) {
    ElMessage.warning('首次绑定请使用手机号授权')
    return
  }
  openHighRiskConfirmation({
    action: 'authorization_qr',
    title: '开始二维码重新授权',
    detail: '将暂停读取并创建短期授权会话。',
    execute: async (ticket) => {
      authorizationBusy.value = true
      try {
        const view = await startAdminTelegramQRAuthorization({ confirmation_ticket: ticket })
        updateAuthorization(view, true)
        resetAuthorizationInputs()
        ElMessage.success('二维码授权已开始')
      } finally {
        authorizationBusy.value = false
      }
    }
  })
}

async function submitAuthorizationCode() {
  const authorizationID = activeAuthorization.value?.id
  const code = authorizationForm.code.trim()
  if (!authorizationID || !code) {
    ElMessage.warning('请输入验证码')
    return
  }
  authorizationBusy.value = true
  authorizationForm.code = ''
  try {
    const view = await submitAdminTelegramAuthorizationCode(authorizationID, { code })
    updateAuthorization(view)
    ElMessage.success('验证码已提交')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '提交验证码失败'))
  } finally {
    authorizationBusy.value = false
  }
}

async function submitAuthorizationPassword() {
  const authorizationID = activeAuthorization.value?.id
  const password = authorizationForm.password
  if (!authorizationID || !password.trim()) {
    ElMessage.warning('请输入二次验证密码')
    return
  }
  authorizationBusy.value = true
  authorizationForm.password = ''
  try {
    const view = await submitAdminTelegramAuthorizationPassword(authorizationID, { password })
    updateAuthorization(view)
    ElMessage.success('二次验证密码已提交')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '提交二次验证密码失败'))
  } finally {
    authorizationBusy.value = false
  }
}

async function cancelAuthorization() {
  const authorizationID = activeAuthorization.value?.id
  if (!authorizationID) return
  authorizationBusy.value = true
  try {
    await cancelAdminTelegramAuthorization(authorizationID)
    authorization.value = { ...activeAuthorization.value, status: 'cancelled' }
    resetAuthorizationInputs()
    ElMessage.success('授权已取消')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '取消授权失败'))
  } finally {
    authorizationBusy.value = false
    void loadDashboard({ silent: true })
  }
}

function onSourceInputChanged() {
  sourcePreview.value = null
  sourceError.value = ''
}

async function runSourcePreview(chatRef, confirmationTicket = '') {
  sourceBusy.value = true
  sourceError.value = ''
  try {
    sourcePreview.value = await previewAdminTelegramSource({
      chat_ref: chatRef,
      confirmation_ticket: confirmationTicket
    })
  } catch (error) {
    sourcePreview.value = null
    sourceError.value = extractErrorMessage(error, '预解析 Telegram 来源失败')
  } finally {
    sourceBusy.value = false
  }
}

function previewSource() {
  const chatRef = sourceInput.value.trim()
  if (!chatRef) {
    ElMessage.warning('请输入 Telegram 群组、频道或邀请链接')
    return
  }
  if (isPrivateTelegramInviteReference(chatRef)) {
    openHighRiskConfirmation({
      action: 'source_private_preview',
      title: '预解析私密邀请',
      detail: '仅检查来源状态；不会立即创建采集来源。',
      execute: (ticket) => runSourcePreview(chatRef, ticket)
    })
    return
  }
  void runSourcePreview(chatRef)
}

async function runSourceConfirmation(previewID, chatRef, confirmationTicket = '') {
  sourceBusy.value = true
  sourceError.value = ''
  try {
    await confirmAdminTelegramSource({
      preview_id: previewID,
      chat_ref: chatRef,
      confirmation_ticket: confirmationTicket
    })
    sourceInput.value = ''
    sourcePreview.value = null
    ElMessage.success('Telegram 来源已添加')
    await loadDashboard({ silent: true })
  } catch (error) {
    sourceError.value = extractErrorMessage(error, '确认 Telegram 来源失败')
  } finally {
    sourceBusy.value = false
  }
}

function confirmSource() {
  const previewID = sourcePreview.value?.preview_id
  const chatRef = sourceInput.value.trim()
  if (!previewID || !chatRef) {
    ElMessage.warning('来源预览已失效，请重新预解析')
    return
  }
  if (isPrivateTelegramInviteReference(chatRef)) {
    openHighRiskConfirmation({
      action: 'source_private_confirm',
      title: '确认添加私密来源',
      detail: '确认后可能加入该私密群组或频道，并开始采集。',
      execute: (ticket) => runSourceConfirmation(previewID, chatRef, ticket)
    })
    return
  }
  void runSourceConfirmation(previewID, chatRef)
}

async function runSourceAction(source, action) {
  if (!source?.id || sourceActionID.value) return
  sourceActionID.value = source.id
  try {
    if (action === 'pause') await pauseAdminTelegramSource(source.id)
    if (action === 'resume') await resumeAdminTelegramSource(source.id)
    if (action === 'recover') await recoverAdminTelegramSource(source.id)
    if (action === 'backfill') await startAdminTelegramSourceBackfill(source.id)
    ElMessage.success('来源状态已更新')
    await loadDashboard({ silent: true })
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '更新 Telegram 来源失败'))
  } finally {
    sourceActionID.value = ''
  }
}

async function requestSourceAction(source, action) {
  if (action !== 'backfill') {
    await runSourceAction(source, action)
    return
  }
  try {
    await ElMessageBox.confirm(
      `将从头重新扫描“${sourceTitle(source)}”的历史消息，已有视频仍会保持去重。确认继续吗？`,
      '重新回填',
      {
        type: 'warning',
        confirmButtonText: '开始回填',
        cancelButtonText: '取消'
      }
    )
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error('操作确认失败，请重试')
    return
  }
  await runSourceAction(source, action)
}

function setAuditPage(page) {
  auditQuery.page = page
  void loadDashboard({ silent: true })
}

function handleVisibilityChange() {
  if (!pageDocumentVisible()) {
    clearPollTimer()
    return
  }
  void loadDashboard({ silent: true })
}

onMounted(() => {
  document.addEventListener('visibilitychange', handleVisibilityChange)
  void loadDashboard()
})

onUnmounted(() => {
  pageActive = false
  refreshSequence += 1
  clearPollTimer()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  resetAuthorizationInputs()
  resetHighRiskConfirmation()
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-tooltip content="刷新 Telegram 管理状态" placement="bottom">
        <el-button
          class="telegram-refresh"
          circle
          :icon="Refresh"
          :loading="refreshing"
          aria-label="刷新 Telegram 管理状态"
          title="刷新 Telegram 管理状态"
          @click="loadDashboard()"
        />
      </el-tooltip>
    </template>

    <div class="page-shell telegram-page" data-density="monitor">
      <MetricStrip :items="summaryMetrics" aria-label="Telegram 采集摘要" />

      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default>
          <el-button link type="primary" @click="loadDashboard()">重试</el-button>
        </template>
      </el-alert>

      <el-skeleton v-if="initialLoading" :rows="12" animated />

      <template v-else>
        <SectionCard>
          <template #title>采集账号</template>
          <template #actions>
            <StatusIndicator
              :label="accountStatusText(status?.account_status || account.status)"
              :tone="accountStatusTone(status?.account_status || account.status)"
            />
          </template>

          <div class="telegram-status-grid">
            <div class="telegram-detail">
              <span>账号</span>
              <strong>{{ accountDisplayName }}</strong>
            </div>
            <div class="telegram-detail">
              <span>手机号</span>
              <strong>{{ account.phone_masked || '--' }}</strong>
            </div>
            <div class="telegram-detail">
              <span>采集器</span>
              <StatusIndicator
                :label="ingestorStatusText(status?.ingestor_status || heartbeat.status)"
                :tone="ingestorStatusTone(status?.ingestor_status || heartbeat.status)"
              />
            </div>
            <div class="telegram-detail">
              <span>最近心跳</span>
              <strong class="tabular-num">{{ formatDateTime(heartbeat.last_seen_at) }}</strong>
            </div>
          </div>

          <el-alert
            v-if="status?.error || account.last_error || heartbeat.error_summary"
            class="telegram-runtime-error"
            type="warning"
            :closable="false"
            :title="status?.error || account.last_error || heartbeat.error_summary"
          />

          <div class="authorization-workspace">
            <div class="authorization-workspace__actions">
              <el-button
                type="primary"
                :icon="Key"
                :disabled="authorizationActive"
                :loading="authorizationBusy"
                @click="startPhoneAuthorization"
              >
                手机号授权
              </el-button>
              <el-tooltip :disabled="accountBound" content="首次绑定请使用手机号授权" placement="top">
                <span>
                  <el-button
                    :icon="Connection"
                    :disabled="authorizationActive || !accountBound"
                    :loading="authorizationBusy"
                    @click="startQRAuthorization"
                  >
                    二维码重新授权
                  </el-button>
                </span>
              </el-tooltip>
            </div>

            <div v-if="activeAuthorization" class="authorization-state">
              <div class="authorization-state__head">
                <StatusIndicator
                  :label="authorizationStatusText(activeAuthorization.status)"
                  :tone="authorizationStatusTone(activeAuthorization.status)"
                />
                <span class="authorization-expiry tabular-num">有效至 {{ formatDateTime(activeAuthorization.expires_at) }}</span>
              </div>

              <p v-if="activeAuthorization.error" class="authorization-error">{{ activeAuthorization.error }}</p>

              <div v-if="activeAuthorization.status === 'scanning'" class="authorization-qr">
                <div class="authorization-qr__image">
                  <img
                    v-if="activeAuthorization.qr_image_data_url"
                    :src="activeAuthorization.qr_image_data_url"
                    alt="Telegram 授权二维码"
                    width="220"
                    height="220"
                  >
                  <el-skeleton v-else :rows="6" animated />
                </div>
                <span class="tabular-num">二维码有效至 {{ formatDateTime(activeAuthorization.qr_image_expires_at) }}</span>
              </div>

              <form
                v-if="activeAuthorization.status === 'awaiting_code' && canSubmitAuthorization"
                class="authorization-form"
                @submit.prevent="submitAuthorizationCode"
              >
                <el-input
                  v-model="authorizationForm.code"
                  inputmode="numeric"
                  autocomplete="one-time-code"
                  maxlength="16"
                  placeholder="Telegram 验证码"
                  aria-label="Telegram 验证码"
                />
                <el-button type="primary" :loading="authorizationBusy" native-type="submit">提交验证码</el-button>
              </form>

              <form
                v-if="activeAuthorization.status === 'awaiting_password' && canSubmitAuthorization"
                class="authorization-form"
                @submit.prevent="submitAuthorizationPassword"
              >
                <el-input
                  v-model="authorizationForm.password"
                  type="password"
                  show-password
                  autocomplete="current-password"
                  placeholder="Telegram 二次验证密码"
                  aria-label="Telegram 二次验证密码"
                />
                <el-button type="primary" :loading="authorizationBusy" native-type="submit">提交密码</el-button>
              </form>

              <div v-if="authorizationActive" class="authorization-state__footer">
                <span v-if="!canSubmitAuthorization" class="authorization-owner-note">仅发起该授权的管理员可以提交验证信息。</span>
                <el-button
                  v-if="canSubmitAuthorization"
                  text
                  type="danger"
                  :loading="authorizationBusy"
                  @click="cancelAuthorization"
                >
                  取消授权
                </el-button>
              </div>
            </div>
          </div>
        </SectionCard>

        <SectionCard>
          <template #title>添加来源</template>
          <template #actions>
            <StatusIndicator
              v-if="sourcePreview"
              :label="previewStatusText(sourcePreview)"
              :tone="previewStatusTone(sourcePreview)"
            />
          </template>

          <form class="source-form" @submit.prevent="previewSource">
            <el-input
              v-model="sourceInput"
              clearable
              maxlength="500"
              placeholder="@频道、Telegram 链接、邀请链接或 chat ID"
              aria-label="Telegram 来源引用"
              :prefix-icon="Link"
              @input="onSourceInputChanged"
              @clear="onSourceInputChanged"
            />
            <el-button type="primary" :icon="RefreshRight" :loading="sourceBusy" native-type="submit">预解析</el-button>
          </form>

          <el-alert v-if="sourceError" class="source-error" type="error" :closable="false" :title="sourceError" />

          <div v-if="sourcePreview" class="source-preview">
            <div class="source-preview__details">
              <div>
                <span>标题</span>
                <strong>{{ sourcePreview.title || '--' }}</strong>
              </div>
              <div>
                <span>类型</span>
                <strong>{{ sourcePreview.chat_type || '--' }}</strong>
              </div>
              <div>
                <span>规范 ID</span>
                <strong class="tabular-num">{{ sourcePreview.chat_id || '确认后获取' }}</strong>
              </div>
              <div>
                <span>预览有效至</span>
                <strong class="tabular-num">{{ formatDateTime(sourcePreview.expires_at) }}</strong>
              </div>
            </div>
            <div class="source-preview__actions">
              <StatusIndicator
                v-if="sourcePreview.requires_join || sourcePreview.requires_approval"
                :label="previewStatusText(sourcePreview)"
                tone="warning"
                :icon="WarningFilled"
              />
              <el-button type="primary" :loading="sourceBusy" @click="confirmSource">
                {{ sourcePreview.requires_join ? '确认加入并添加' : '确认添加' }}
              </el-button>
            </div>
          </div>
        </SectionCard>

        <SectionCard dense>
          <template #title>采集来源</template>
          <template #actions>
            <el-tag effect="plain">共 {{ sources.length }} 个</el-tag>
          </template>

          <EmptyState
            v-if="sources.length === 0"
            :icon="Connection"
            title="暂无 Telegram 来源"
          />

          <div v-else class="table-wrap">
            <el-table class="telegram-source-table" :data="sources" row-key="id" border>
              <el-table-column label="来源" min-width="230">
                <template #default="{ row }">
                  <div class="source-name-cell">
                    <strong>{{ sourceTitle(row) }}</strong>
                    <span>{{ row.username ? `@${row.username}` : row.chat_id }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="状态" min-width="132">
                <template #default="{ row }">
                  <StatusIndicator :label="sourceStatusText(row.sync_status)" :tone="sourceStatusTone(row.sync_status)" />
                  <span v-if="row.last_error" class="source-error-text">{{ row.last_error }}</span>
                </template>
              </el-table-column>
              <el-table-column label="处理进度" min-width="210">
                <template #default="{ row }">
                  <span class="source-progress-summary tabular-num">{{ sourceProgressSummary(row) }}</span>
                  <span v-if="row.backfill_completed_at" class="source-progress-meta">回填完成 {{ formatDateTime(row.backfill_completed_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="最近更新" min-width="166">
                <template #default="{ row }">
                  <time class="tabular-num" :datetime="row.updated_at">{{ formatDateTime(row.updated_at) }}</time>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="176" fixed="right">
                <template #default="{ row }">
                  <div class="source-actions">
                    <el-tooltip v-if="row.enabled" content="暂停来源" placement="top">
                      <el-button
                        text
                        :icon="VideoPause"
                        :loading="sourceActionID === row.id"
                        aria-label="暂停来源"
                        title="暂停来源"
                        @click="requestSourceAction(row, 'pause')"
                      />
                    </el-tooltip>
                    <el-tooltip v-else content="恢复来源" placement="top">
                      <el-button
                        text
                        :icon="VideoPlay"
                        :loading="sourceActionID === row.id"
                        aria-label="恢复来源"
                        title="恢复来源"
                        @click="requestSourceAction(row, 'resume')"
                      />
                    </el-tooltip>
                    <el-tooltip v-if="row.sync_status === 'error'" content="恢复失败来源" placement="top">
                      <el-button
                        text
                        type="warning"
                        :icon="WarningFilled"
                        :loading="sourceActionID === row.id"
                        aria-label="恢复失败来源"
                        title="恢复失败来源"
                        @click="requestSourceAction(row, 'recover')"
                      />
                    </el-tooltip>
                    <el-tooltip content="重新回填历史消息" placement="top">
                      <el-button
                        text
                        :icon="RefreshRight"
                        :loading="sourceActionID === row.id"
                        aria-label="重新回填历史消息"
                        title="重新回填历史消息"
                        @click="requestSourceAction(row, 'backfill')"
                      />
                    </el-tooltip>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </SectionCard>

        <SectionCard dense>
          <template #title>管理审计</template>
          <template #actions>
            <el-tag effect="plain">最近操作</el-tag>
          </template>

          <EmptyState
            v-if="audits.length === 0"
            :icon="CircleCheck"
            title="暂无 Telegram 管理审计"
          />

          <template v-else>
            <div class="table-wrap">
              <el-table class="telegram-audit-table" :data="audits" row-key="id" border>
                <el-table-column label="时间" width="168">
                  <template #default="{ row }">
                    <time class="tabular-num" :datetime="row.created_at">{{ formatDateTime(row.created_at) }}</time>
                  </template>
                </el-table-column>
                <el-table-column label="操作" min-width="200">
                  <template #default="{ row }"><code>{{ row.action || '--' }}</code></template>
                </el-table-column>
                <el-table-column label="结果" width="100">
                  <template #default="{ row }">
                    <StatusIndicator :label="auditResultText(row.result)" :tone="auditResultTone(row.result)" />
                  </template>
                </el-table-column>
                <el-table-column label="目标" min-width="180">
                  <template #default="{ row }">
                    <span>{{ row.target_type || '--' }} · {{ row.target_id || '--' }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="摘要" min-width="260">
                  <template #default="{ row }"><span class="audit-summary">{{ auditSummary(row.summary) }}</span></template>
                </el-table-column>
              </el-table>
            </div>
            <div class="audit-pagination">
              <AdminTablePagination
                :current-page="auditQuery.page"
                :page-size="auditQuery.page_size"
                layout="total, prev, pager, next"
                :total="auditTotal"
                :disabled="refreshing"
                @current-change="setAuditPage"
              />
            </div>
          </template>
        </SectionCard>
      </template>
    </div>

    <el-dialog
      v-model="confirmation.visible"
      :title="confirmation.title"
      width="min(30rem, calc(100vw - 2rem))"
      :close-on-click-modal="!confirmationBusy"
      :close-on-press-escape="!confirmationBusy"
      :show-close="!confirmationBusy"
      @closed="resetHighRiskConfirmation"
    >
      <p class="confirmation-detail">{{ confirmation.detail }}</p>
      <el-form label-position="top" @submit.prevent="submitHighRiskConfirmation">
        <el-form-item label="当前管理员密码">
          <el-input
            v-model="confirmation.password"
            type="password"
            show-password
            autocomplete="current-password"
            :disabled="confirmationBusy"
          />
        </el-form-item>
        <el-checkbox v-model="confirmation.acknowledged" :disabled="confirmationBusy">我已确认继续该操作</el-checkbox>
      </el-form>
      <template #footer>
        <el-button :disabled="confirmationBusy" @click="resetHighRiskConfirmation">取消</el-button>
        <el-button type="primary" :loading="confirmationBusy" @click="submitHighRiskConfirmation">确认继续</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.telegram-page {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-4);
}

.telegram-status-grid,
.source-preview__details {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.telegram-detail,
.source-preview__details > div {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.telegram-detail > span,
.source-preview__details span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.telegram-detail > strong,
.source-preview__details strong {
  min-width: 0;
  color: var(--text-primary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  overflow-wrap: anywhere;
}

.telegram-runtime-error,
.source-error {
  margin-top: var(--space-3);
}

.authorization-workspace {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line-soft);
}

.authorization-workspace__actions,
.authorization-state__footer,
.source-form,
.source-preview__actions,
.source-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.authorization-state {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-3) 0 0;
  border-top: 1px solid var(--line-soft);
}

.authorization-state__head {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.authorization-expiry,
.authorization-qr > span,
.source-progress-meta,
.source-name-cell span,
.source-error-text,
.authorization-owner-note {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.authorization-error,
.source-error-text {
  margin: 0;
  color: var(--danger-600);
  overflow-wrap: anywhere;
}

.authorization-qr {
  display: grid;
  width: fit-content;
  max-width: 100%;
  gap: var(--space-2);
}

.authorization-qr__image {
  display: grid;
  width: 220px;
  height: 220px;
  max-width: 100%;
  place-items: center;
  border: 1px solid var(--line-soft);
  background: var(--bg-surface-muted);
}

.authorization-qr__image img {
  display: block;
  width: 220px;
  height: 220px;
  max-width: 100%;
  object-fit: contain;
}

.authorization-form {
  display: flex;
  width: min(34rem, 100%);
  align-items: center;
  gap: var(--space-2);
}

.authorization-form :deep(.el-input) {
  min-width: 0;
  flex: 1 1 16rem;
}

.authorization-state__footer {
  justify-content: space-between;
}

.source-form :deep(.el-input) {
  min-width: 0;
  flex: 1 1 24rem;
}

.source-preview {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line-soft);
}

.source-preview__actions {
  justify-content: space-between;
}

.source-name-cell {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.source-name-cell strong,
.source-progress-summary,
.audit-summary {
  min-width: 0;
  overflow-wrap: anywhere;
}

.source-progress-summary,
.source-progress-meta {
  display: block;
}

.source-actions {
  min-height: 32px;
  flex-wrap: nowrap;
}

.telegram-source-table :deep(.el-table__cell),
.telegram-audit-table :deep(.el-table__cell) {
  vertical-align: top;
}

.audit-pagination {
  padding-top: var(--space-3);
}

.confirmation-detail {
  margin: 0 0 var(--space-4);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

@media (max-width: 63.9375rem) {
  .telegram-refresh {
    width: 44px;
    height: 44px;
    min-width: 44px;
    min-height: 44px;
  }

  .telegram-status-grid,
  .source-preview__details {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .authorization-workspace__actions > .el-button,
  .authorization-workspace__actions > span > .el-button,
  .source-form > .el-button,
  .source-preview__actions > .el-button {
    min-height: 44px;
  }

  .source-actions :deep(.el-button) {
    min-width: 44px;
    min-height: 44px;
  }

  .source-form,
  .authorization-form {
    align-items: stretch;
  }
}

@media (max-width: 36rem) {
  .telegram-status-grid,
  .source-preview__details {
    grid-template-columns: minmax(0, 1fr);
  }

  .authorization-workspace__actions,
  .source-form,
  .authorization-form,
  .source-preview__actions {
    align-items: stretch;
    flex-direction: column;
  }

  .authorization-workspace__actions > .el-button,
  .authorization-workspace__actions > span,
  .authorization-workspace__actions > span > .el-button,
  .source-form > .el-button,
  .authorization-form > .el-button,
  .source-preview__actions > .el-button {
    width: 100%;
  }

  .source-form :deep(.el-input),
  .authorization-form :deep(.el-input) {
    width: 100%;
    flex: 0 1 auto;
  }
}
</style>
