<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Delete, RefreshRight, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { deleteLatestOrphanFileScan, getLatestOrphanFileScan, startOrphanFileScan } from '../api/admin'
import {
  buildOrphanScanPromptKey,
  formatOrphanScanBytes,
  formatOrphanScanTime,
  getOrphanScanStatusLabel,
  getOrphanScanStatusType,
  shouldPromptDeleteOrphanScan
} from './toolboxOrphanFiles.helpers'

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

const router = useRouter()
const orphanScan = ref({ ...EMPTY_ORPHAN_SCAN })
const orphanScanLoading = ref(false)
const orphanScanStartLoading = ref(false)
const orphanScanDeleteLoading = ref(false)
const orphanScanPromptedKey = ref('')
let orphanScanTimer = null
let orphanScanLoadSeq = 0

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

function returnToToolbox() {
  router.push('/toolbox')
}

onMounted(async () => {
  await loadLatestOrphanScan({ silent: true })
})

onUnmounted(() => {
  stopOrphanScanPolling()
})
</script>

<template>
  <main class="tool-workspace orphan-tool">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader title="孤儿文件扫描" subtitle="盘点存储目录中的未引用持久文件，并按扫描结果统一处理。" />

      <SectionCard>
        <template #title>扫描状态</template>
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
    </div>
  </main>
</template>

<style scoped>
.tool-workspace {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.tool-workspace__inner {
  display: grid;
  width: min(100%, 80rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
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

@media (max-width: 64rem) {
  .scan-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .scan-metrics {
    grid-template-columns: 1fr;
  }
}
</style>
