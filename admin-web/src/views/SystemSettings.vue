<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Delete, RefreshRight, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { getSystemLogs, systemCleanup, deleteLatestOrphanFileScan, getLatestOrphanFileScan, startOrphanFileScan } from '../api/admin'
import {
  buildOrphanScanPromptKey,
  formatOrphanScanBytes,
  formatOrphanScanTime,
  getOrphanScanStatusLabel,
  getOrphanScanStatusType,
  parseEd2kLinks,
  shouldPromptDeleteOrphanScan
} from './systemSettings.helpers'

const EMPTY_ORPHAN_SCAN = {
  id: 0,
  status: 'idle',
  total_files: 0,
  referenced_files: 0,
  orphan_files: 0,
  deleted_files: 0,
  error: '',
  started_at: null,
  finished_at: null,
  created_at: null,
  updated_at: null,
  items: []
}

const logs = ref([])
const loading = ref(false)
const cleanupLoading = ref(false)
const orphanScan = ref({ ...EMPTY_ORPHAN_SCAN })
const orphanScanLoading = ref(false)
const orphanScanStartLoading = ref(false)
const orphanScanDeleteLoading = ref(false)
const orphanScanPromptedKey = ref('')
const ed2kInput = ref('')
const ed2kClickedLinks = ref(new Set())
let orphanScanTimer = null
let orphanScanLoadSeq = 0

const hasLogs = computed(() => logs.value.length > 0)
const ed2kParseResult = computed(() => parseEd2kLinks(ed2kInput.value))
const ed2kLinks = computed(() => ed2kParseResult.value.links)
const ed2kInvalidCount = computed(() => ed2kParseResult.value.invalidCount)
const ed2kHasInput = computed(() => ed2kInput.value.trim() !== '')
const orphanScanData = computed(() => orphanScan.value || EMPTY_ORPHAN_SCAN)
const orphanScanStatus = computed(() => orphanScanData.value.status || 'idle')
const orphanScanIsActive = computed(() => ['pending', 'running'].includes(orphanScanStatus.value))
const orphanScanVisibleItems = computed(() => {
  if (!['completed', 'deleted'].includes(orphanScanStatus.value)) {
    return []
  }
  return Array.isArray(orphanScanData.value.items) ? orphanScanData.value.items : []
})
const orphanScanHasItems = computed(() => orphanScanVisibleItems.value.length > 0)
const canStartOrphanScan = computed(() => !orphanScanIsActive.value && !orphanScanStartLoading.value)
const canDeleteOrphanScan = computed(() => orphanScanStatus.value === 'completed' && Number(orphanScanData.value.orphan_files || 0) > 0)
const scanResultText = computed(() => {
  if (orphanScanStatus.value === 'completed') {
    const total = Number(orphanScanData.value.orphan_files || 0)
    if (total > 0) {
      return `已发现 ${total} 个孤儿文件，系统会自动询问是否全量删除。`
    }
    return '扫描完成，未发现可删除文件。'
  }
  if (orphanScanStatus.value === 'deleted') {
    return `本次扫描结果已全量删除，共删除 ${Number(orphanScanData.value.deleted_files || 0)} 个文件。`
  }
  if (orphanScanStatus.value === 'failed') {
    return orphanScanData.value.error || '扫描失败'
  }
  if (orphanScanIsActive.value) {
    return '扫描正在进行，结果会自动刷新。'
  }
  return '点击开始扫描后，系统会检查当前存储目录中的未引用文件。'
})

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') {
    return responseMsg.trim()
  }
  const message = error?.message
  if (typeof message === 'string' && message.trim() !== '') {
    return message.trim()
  }
  return fallback
}

function normalizeOrphanScan(data) {
  const next = { ...EMPTY_ORPHAN_SCAN, ...(data || {}) }
  next.total_files = Number(next.total_files || 0)
  next.referenced_files = Number(next.referenced_files || 0)
  next.orphan_files = Number(next.orphan_files || 0)
  next.deleted_files = Number(next.deleted_files || 0)
  next.items = Array.isArray(next.items) ? next.items : []
  return next
}

function stopOrphanScanPolling() {
  if (orphanScanTimer) {
    clearInterval(orphanScanTimer)
    orphanScanTimer = null
  }
}

function syncOrphanScanPolling() {
  if (orphanScanIsActive.value) {
    if (!orphanScanTimer) {
      orphanScanTimer = setInterval(() => {
        loadLatestOrphanScan({ silent: true })
      }, 3000)
    }
    return
  }
  stopOrphanScanPolling()
}

async function loadLogs() {
  loading.value = true
  try {
    const data = await getSystemLogs({ lines: 300 })
    logs.value = data.lines || []
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载系统日志失败'))
  } finally {
    loading.value = false
  }
}

async function runCleanup() {
  cleanupLoading.value = true
  try {
    await systemCleanup({ older_than_hours: 24 })
    ElMessage.success('清理任务已执行')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '执行清理任务失败'))
  } finally {
    cleanupLoading.value = false
  }
}

async function deleteOrphanScan() {
  orphanScanDeleteLoading.value = true
  try {
    const data = await deleteLatestOrphanFileScan()
    orphanScan.value = normalizeOrphanScan(data)
    syncOrphanScanPolling()
    ElMessage.success(`已删除 ${Number(orphanScan.value.deleted_files || 0)} 个孤儿文件`)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除孤儿文件失败'))
  } finally {
    orphanScanDeleteLoading.value = false
  }
}

async function confirmDeleteOrphanScan({ auto = false } = {}) {
  const current = orphanScanData.value
  if (!canDeleteOrphanScan.value) {
    if (!auto) {
      ElMessage.info('当前没有可删除的孤儿文件')
    }
    return
  }

  const promptKey = buildOrphanScanPromptKey(current)
  if (auto && !shouldPromptDeleteOrphanScan(current, orphanScanPromptedKey.value)) {
    return
  }
  orphanScanPromptedKey.value = promptKey

  try {
    await ElMessageBox.confirm(
      `本次扫描发现 ${Number(current.orphan_files || 0)} 个孤儿文件，是否立即全量删除？删除后不会清空空目录。`,
      '确认删除孤儿文件',
      {
        confirmButtonText: '全量删除',
        cancelButtonText: '保留结果',
        type: 'warning',
        closeOnClickModal: false
      }
    )
    await deleteOrphanScan()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(extractErrorMessage(error, '删除确认失败'))
    }
  }
}

async function loadLatestOrphanScan({ silent = false } = {}) {
  if (orphanScanLoading.value) {
    return
  }
  const seq = ++orphanScanLoadSeq
  orphanScanLoading.value = true
  try {
    const data = await getLatestOrphanFileScan()
    if (seq !== orphanScanLoadSeq) {
      return
    }
    orphanScan.value = normalizeOrphanScan(data)
    syncOrphanScanPolling()
    await confirmDeleteOrphanScan({ auto: true })
  } catch (error) {
    if (!silent && seq === orphanScanLoadSeq) {
      ElMessage.error(extractErrorMessage(error, '加载孤儿文件扫描状态失败'))
    }
  } finally {
    if (seq === orphanScanLoadSeq) {
      orphanScanLoading.value = false
    }
  }
}

async function startOrphanScan() {
  orphanScanStartLoading.value = true
  try {
    await startOrphanFileScan()
    ElMessage.success('已开始孤儿文件扫描')
    await loadLatestOrphanScan()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '开始孤儿文件扫描失败'))
  } finally {
    orphanScanStartLoading.value = false
  }
}

function formatOrphanScanItemPath(row) {
  return row.relative_path || row.file_path || '--'
}

function isEd2kLinkClicked(link) {
  return ed2kClickedLinks.value.has(link.href)
}

function markEd2kLinkClicked(link) {
  ed2kClickedLinks.value = new Set([...ed2kClickedLinks.value, link.href])
}

function clearEd2kInput() {
  ed2kInput.value = ''
  ed2kClickedLinks.value = new Set()
}

onMounted(async () => {
  await loadLatestOrphanScan({ silent: true })
})

onUnmounted(() => {
  stopOrphanScanPolling()
})
</script>

<template>
  <Layout>
    <div class="page-shell settings-page">
      <PageHeader title="系统设置" subtitle="执行临时文件清理、孤儿文件扫描和日志查看" />

      <SectionCard>
        <template #title>ED2K 链接生成器</template>
        <template #description>把多行 ED2K 文本转换为可点击链接，点击后在本页标记状态。</template>
        <template #actions>
          <el-button :disabled="!ed2kHasInput" @click="clearEd2kInput">清空</el-button>
        </template>

        <div class="ed2k-tool">
          <el-input
            v-model="ed2kInput"
            type="textarea"
            :rows="6"
            resize="vertical"
            placeholder="每行一个 ed2k:// 链接"
            class="ed2k-tool__input"
          />
          <div class="ed2k-tool__summary">
            <span>有效链接：{{ ed2kLinks.length }}</span>
            <span v-if="ed2kInvalidCount > 0">已忽略 {{ ed2kInvalidCount }} 行非 ED2K 文本</span>
          </div>

          <EmptyState
            v-if="ed2kHasInput && ed2kLinks.length === 0"
            title="未识别到 ED2K 链接"
            description="请输入以 ed2k:// 开头的链接"
          />

          <div v-else-if="ed2kLinks.length > 0" class="ed2k-link-list" aria-label="生成的 ED2K 链接">
            <a
              v-for="link in ed2kLinks"
              :key="link.id"
              class="ed2k-link"
              :class="{ 'is-clicked': isEd2kLinkClicked(link) }"
              :href="link.href"
              @click="markEd2kLinkClicked(link)"
            >
              <span class="ed2k-link__line tabular-num">{{ link.lineNumber }}</span>
              <span class="ed2k-link__text">{{ link.label }}</span>
              <el-tag v-if="isEd2kLinkClicked(link)" size="small" type="info" effect="plain">已点击</el-tag>
              <el-tag v-else size="small" type="success" effect="plain">未点击</el-tag>
            </a>
          </div>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>临时文件清理</template>
        <template #description>清理超过 24 小时的临时文件，释放磁盘占用。</template>
        <template #actions>
          <el-button type="warning" :loading="cleanupLoading" @click="runCleanup">执行清理</el-button>
        </template>
        <p class="section-note">清理仅作用于上传暂存目录，不会删除业务库中的媒体资源。</p>
      </SectionCard>

      <SectionCard>
        <template #title>孤儿文件扫描</template>
        <template #description>盘点 `STORAGE_ROOT` 下已知业务子树中的未引用持久文件。扫描完成后会自动询问是否全量删除。</template>
        <template #actions>
          <el-button :loading="orphanScanLoading" @click="loadLatestOrphanScan()">
            <el-icon><RefreshRight /></el-icon>
            <span>刷新状态</span>
          </el-button>
          <el-button type="primary" :loading="orphanScanStartLoading" :disabled="!canStartOrphanScan" @click="startOrphanScan">
            <el-icon><Search /></el-icon>
            <span>开始扫描</span>
          </el-button>
          <el-button v-if="canDeleteOrphanScan" type="danger" :loading="orphanScanDeleteLoading" @click="confirmDeleteOrphanScan()">
            <el-icon><Delete /></el-icon>
            <span>全量删除</span>
          </el-button>
        </template>

        <div class="scan-summary">
          <div class="scan-summary__state">
            <span class="scan-summary__label">状态</span>
            <el-tag effect="plain" :type="getOrphanScanStatusType(orphanScanStatus)">{{ getOrphanScanStatusLabel(orphanScanStatus) }}</el-tag>
          </div>
          <div class="scan-summary__meta">
            <span>开始：{{ formatOrphanScanTime(orphanScanData.started_at) }}</span>
            <span>完成：{{ formatOrphanScanTime(orphanScanData.finished_at) }}</span>
            <span>更新：{{ formatOrphanScanTime(orphanScanData.updated_at) }}</span>
          </div>
          <div class="scan-metrics">
            <div class="scan-metric">
              <span class="scan-metric__label">总文件</span>
              <strong class="scan-metric__value tabular-num">{{ orphanScanData.total_files }}</strong>
            </div>
            <div class="scan-metric">
              <span class="scan-metric__label">已引用</span>
              <strong class="scan-metric__value tabular-num">{{ orphanScanData.referenced_files }}</strong>
            </div>
            <div class="scan-metric">
              <span class="scan-metric__label">孤儿文件</span>
              <strong class="scan-metric__value tabular-num">{{ orphanScanData.orphan_files }}</strong>
            </div>
            <div class="scan-metric">
              <span class="scan-metric__label">已删除</span>
              <strong class="scan-metric__value tabular-num">{{ orphanScanData.deleted_files }}</strong>
            </div>
          </div>
          <el-alert
            v-if="orphanScanStatus === 'failed'"
            :title="scanResultText"
            type="error"
            show-icon
            :closable="false"
          />
          <el-alert
            v-else-if="orphanScanStatus === 'pending' || orphanScanStatus === 'running'"
            :title="scanResultText"
            type="info"
            show-icon
            :closable="false"
          />
          <el-alert
            v-else-if="orphanScanStatus === 'completed' && Number(orphanScanData.orphan_files || 0) > 0"
            :title="scanResultText"
            type="warning"
            show-icon
            :closable="false"
          />
          <el-alert
            v-else-if="orphanScanStatus === 'completed' || orphanScanStatus === 'deleted'"
            :title="scanResultText"
            type="success"
            show-icon
            :closable="false"
          />
          <el-alert
            v-else
            :title="scanResultText"
            type="info"
            show-icon
            :closable="false"
          />
        </div>

        <EmptyState
          v-if="!orphanScanHasItems && !orphanScanIsActive"
          :title="orphanScanStatus === 'failed' ? '扫描失败' : orphanScanStatus === 'completed' ? '未发现可删除文件' : orphanScanStatus === 'deleted' ? '本次结果已删除' : '尚未扫描'"
          :description="orphanScanStatus === 'failed' ? '请先修正错误后重新扫描' : orphanScanStatus === 'completed' ? '完成后会在此处显示最新的孤儿文件清单' : orphanScanStatus === 'deleted' ? '本次扫描结果已保留在列表中，便于核对' : '点击开始扫描后，系统会盘点未引用文件'"
        >
          <template #action>
            <el-button :loading="orphanScanLoading" @click="loadLatestOrphanScan()">刷新状态</el-button>
          </template>
        </EmptyState>

        <div v-else class="scan-table-wrap">
          <el-table v-loading="orphanScanLoading && !orphanScanIsActive" :data="orphanScanVisibleItems" border>
            <el-table-column prop="relative_path" label="相对路径" min-width="260" show-overflow-tooltip>
              <template #default="{ row }">
                {{ formatOrphanScanItemPath(row) }}
              </template>
            </el-table-column>
            <el-table-column prop="file_path" label="绝对路径" min-width="320" show-overflow-tooltip />
            <el-table-column prop="size_bytes" label="大小" width="120" align="right">
              <template #default="{ row }">
                {{ formatOrphanScanBytes(row.size_bytes) }}
              </template>
            </el-table-column>
            <el-table-column prop="mod_time" label="更新时间" width="180">
              <template #default="{ row }">
                {{ formatOrphanScanTime(row.mod_time) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>系统日志</template>
        <template #description>按时间顺序展示最近日志输出</template>
        <template #actions>
          <el-button :loading="loading" @click="loadLogs">
            <el-icon><RefreshRight /></el-icon>
            <span>刷新日志</span>
          </el-button>
        </template>
        <EmptyState
          v-if="!hasLogs"
          title="暂无日志"
          description="点击刷新日志按钮拉取最近日志"
        >
          <template #action>
            <el-button :loading="loading" @click="loadLogs">刷新</el-button>
          </template>
        </EmptyState>
        <el-scrollbar v-else max-height="60vh" class="log-box">
          <pre class="log-text">{{ logs.join('\n') }}</pre>
        </el-scrollbar>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.settings-page {
  display: grid;
  gap: var(--space-6);
}

.section-note {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-tool {
  display: grid;
  gap: var(--space-3);
}

.ed2k-tool__input {
  max-width: 56rem;
}

.ed2k-tool__summary {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link-list {
  display: grid;
  gap: var(--space-2);
}

.ed2k-link {
  display: grid;
  grid-template-columns: 3rem minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  background: var(--bg-surface-muted);
  text-decoration: none;
}

.ed2k-link:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-surface);
}

.ed2k-link.is-clicked {
  color: var(--text-muted);
  background: var(--bg-surface);
}

.ed2k-link__line {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: var(--font-mono);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.scan-summary {
  display: grid;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.scan-summary__state {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.scan-summary__label {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.scan-summary__meta {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.scan-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.scan-metric {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.scan-metric__label {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.scan-metric__value {
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.scan-table-wrap {
  min-width: 0;
}

.log-box {
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--slate-950);
}

.log-text {
  margin: 0;
  padding: var(--space-3);
  white-space: pre-wrap;
  color: var(--slate-100);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  font-family: var(--font-mono);
}

@media (max-width: 64rem) {
  .scan-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 48rem) {
  .scan-metrics {
    grid-template-columns: 1fr;
  }

  .ed2k-link {
    grid-template-columns: 2.5rem minmax(0, 1fr);
  }

  .ed2k-link .el-tag {
    grid-column: 2;
    justify-self: start;
  }
}
</style>
