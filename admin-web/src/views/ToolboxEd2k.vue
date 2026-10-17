<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Back } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { cloudDownloadStatuses, createCloudDownloadBatch } from '../api/admin'
import {
  dedupeEd2kLinksByInfoHash,
  extractApiErrorMessage,
  indexCloudBatchItems,
  parseEd2kLinks,
  summarizeEd2kBatchSelection
} from './toolbox.helpers'

// Keep this aligned with models.CloudDownloadBatchLimit on the backend.
const BATCH_LIMIT = 50
const CLOUD_STATUS_META = {
  pending: { label: '排队中', type: 'info' },
  submitting: { label: '提交中', type: 'info' },
  queued: { label: '115下载中', type: 'warning' },
  ready: { label: '待下载', type: 'warning' },
  downloading: { label: '下载中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' },
  blocked: { label: '需处理', type: 'danger' }
}
const BATCH_STATUS_META = {
  submitted: { label: '已提交', type: 'success' },
  retried: { label: '已重试', type: 'warning' },
  skipped_completed: { label: '已下载', type: 'success' },
  skipped_active: { label: '进行中', type: 'info' },
  duplicate: { label: '重复', type: 'info' },
  invalid: { label: '无效', type: 'danger' },
  failed: { label: '失败', type: 'danger' }
}
const BLOCKED_REASON_LABELS = {
  login_required: '需要登录 115',
  quota_exhausted: '配额不足',
  browser_down: '浏览器未运行',
  needs_attention: '需要人工确认'
}

const router = useRouter()
const ed2kInput = ref('')
const ed2kClickedLinks = ref(new Set())
const selectedLinkIds = ref(new Set())
const duplicateLinkIds = ref(new Set())
const cloudStatuses = ref({})
const batchItemsByHash = ref({})
const batchSummary = ref(null)
const batchSubmitting = ref(false)

const ed2kParseResult = computed(() => parseEd2kLinks(ed2kInput.value))
const ed2kLinks = computed(() => ed2kParseResult.value.links)
const ed2kInvalidCount = computed(() => ed2kParseResult.value.invalidCount)
const ed2kHasInput = computed(() => ed2kInput.value.trim() !== '')
const selectedLinks = computed(() => ed2kLinks.value.filter((link) => selectedLinkIds.value.has(link.id)))
const selectedCount = computed(() => selectedLinks.value.length)
const batchSelectableCount = computed(() => Math.min(ed2kLinks.value.length, BATCH_LIMIT))
const allSelected = computed(() => batchSelectableCount.value > 0 && selectedCount.value === batchSelectableCount.value)
const isSelectionLimitReached = computed(() => selectedCount.value >= BATCH_LIMIT)
const selectionIndeterminate = computed(() => selectedCount.value > 0 && !allSelected.value)
watch(ed2kLinks, (links) => {
  const currentIds = new Set(links.map((link) => link.id))
  const next = new Set([...selectedLinkIds.value].filter((id) => currentIds.has(id)))
  if (next.size !== selectedLinkIds.value.size) {
    selectedLinkIds.value = next
  }
})

const batchSummaryParts = computed(() => {
  if (!batchSummary.value) {
    return []
  }
  const summary = batchSummary.value
  return [
    ['新提交', summary.submitted],
    ['重试', summary.retried],
    ['已下载跳过', summary.skipped_completed],
    ['进行中跳过', summary.skipped_active],
    ['重复', summary.duplicate],
    ['无效', summary.invalid],
    ['失败', summary.failed]
  ].filter(([, count]) => count > 0)
})

function isEd2kLinkClicked(link) {
  return ed2kClickedLinks.value.has(link.href)
}

function markEd2kLinkClicked(link) {
  ed2kClickedLinks.value = new Set([...ed2kClickedLinks.value, link.href])
}

function isSelected(link) {
  return selectedLinkIds.value.has(link.id)
}

function toggleSelection(link) {
  if (batchSubmitting.value) {
    return
  }
  const next = new Set(selectedLinkIds.value)
  if (next.has(link.id)) {
    next.delete(link.id)
  } else {
    if (selectedCount.value >= BATCH_LIMIT) {
      ElMessage.warning(`单批最多 ${BATCH_LIMIT} 条，请先取消一条`)
      return
    }
    next.add(link.id)
  }
  selectedLinkIds.value = next
}

function toggleSelectAll(value) {
  if (batchSubmitting.value) {
    return
  }
  selectedLinkIds.value = value
    ? new Set(ed2kLinks.value.slice(0, BATCH_LIMIT).map((link) => link.id))
    : new Set()
}

function cloudStatusFor(link) {
  if (duplicateLinkIds.value.has(link.id)) {
    return { label: '重复', type: 'info', detail: '同一批内重复' }
  }
  const batchItem = batchItemsByHash.value[link.infoHash || link.href]
  if (batchItem) {
    const meta = BATCH_STATUS_META[batchItem.status] || { label: batchItem.status, type: 'info' }
    return { ...meta, detail: batchItem.message || '' }
  }
  const status = cloudStatuses.value[link.infoHash]
  if (!status) {
    return null
  }
  const meta = CLOUD_STATUS_META[status.status] || { label: status.status, type: 'info' }
  const detail = status.status === 'blocked'
    ? BLOCKED_REASON_LABELS[status.blocked_reason] || status.blocked_reason || ''
    : ''
  return { ...meta, detail }
}

function clearEd2kInput() {
  ed2kInput.value = ''
  ed2kClickedLinks.value = new Set()
  selectedLinkIds.value = new Set()
  duplicateLinkIds.value = new Set()
  cloudStatuses.value = {}
  batchItemsByHash.value = {}
  batchSummary.value = null
}

function returnToToolbox() {
  router.push('/toolbox')
}

function goToCloudDownloads() {
  router.push('/cloud-downloads')
}

async function submitSelected() {
  if (batchSubmitting.value) {
    return
  }
  if (selectedLinks.value.length === 0) {
    ElMessage.warning('请先勾选要提交的链接')
    return
  }
  const { links: uniqueLinks, duplicates } = dedupeEd2kLinksByInfoHash(selectedLinks.value)
  if (uniqueLinks.length > BATCH_LIMIT) {
    ElMessage.error(`单次最多提交 ${BATCH_LIMIT} 条链接，请减少勾选`)
    return
  }

  const preliminary = summarizeEd2kBatchSelection(uniqueLinks, duplicates, {})
  if (preliminary.submittableLinks.length === 0) {
    ElMessage.warning('所选链接缺少有效的 info_hash，无法提交')
    return
  }

  batchSubmitting.value = true
  try {
    let statuses = {}
    if (preliminary.hashes.length > 0) {
      const payload = await cloudDownloadStatuses(preliminary.hashes)
      statuses = payload?.statuses || {}
    }
    const selection = summarizeEd2kBatchSelection(uniqueLinks, duplicates, statuses)

    const summaryText = [
      `共选 ${selectedLinks.value.length} 条，去重后 ${uniqueLinks.length} 条`,
      `无效 ${selection.invalidLinks.length} 条（不提交）`,
      `重复 ${selection.duplicateCount} 条（已合并）`,
      `将提交 ${selection.submittableLinks.length} 条`,
      `新提交 ${selection.fresh} 条`,
      `重试 ${selection.retry} 条`,
      `已下载跳过 ${selection.completed} 条`,
      `进行中跳过 ${selection.active} 条`,
      ed2kInvalidCount.value > 0 ? `另有 ${ed2kInvalidCount.value} 行非 ED2K 文本已忽略` : ''
    ].filter(Boolean).join('；')
    try {
      await ElMessageBox.confirm(
        `${summaryText}。确认发送到 115 云下载？`,
        '发送选中到115',
        { confirmButtonText: '提交', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    const result = await createCloudDownloadBatch(selection.submittableLinks.map((link) => link.href))
    const byHash = indexCloudBatchItems(result?.items)
    for (const link of selection.invalidLinks) {
      byHash[link.href] = { link: link.href, status: 'invalid', message: 'info_hash 无效' }
    }
    batchItemsByHash.value = byHash
    duplicateLinkIds.value = new Set(duplicates.map((link) => link.id))
    const summary = result?.summary || {}
    batchSummary.value = {
      submitted: summary.submitted || 0,
      retried: summary.retried || 0,
      skipped_completed: summary.skipped_completed || 0,
      skipped_active: summary.skipped_active || 0,
      duplicate: (summary.duplicate || 0) + selection.duplicateCount,
      invalid: (summary.invalid || 0) + selection.invalidLinks.length,
      failed: summary.failed || 0
    }
    selectedLinkIds.value = new Set(
      uniqueLinks
        .filter((link) => ['failed', 'invalid'].includes(byHash[link.infoHash || link.href]?.status))
        .map((link) => link.id)
    )
    if (batchSummary.value.failed > 0) {
      ElMessage.warning('部分链接发送失败，请查看结果')
    } else if (batchSummary.value.invalid > 0) {
      ElMessage.warning('部分链接无效，请查看结果')
    } else {
      ElMessage.success('批量提交完成')
    }

    if (selection.hashes.length > 0) {
      try {
        const refreshed = await cloudDownloadStatuses(selection.hashes)
        cloudStatuses.value = refreshed?.statuses || {}
      } catch {
        // The batch result is already visible; a status refresh failure must
        // not be reported as a submit failure.
      }
    }
  } catch (error) {
    ElMessage.error(extractApiErrorMessage(error, '批量提交失败'))
  } finally {
    batchSubmitting.value = false
  }
}
</script>

<template>
  <main class="tool-workspace" data-density="form">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader
        title="ED2K 链接生成器"
        subtitle="把多行 ED2K 文本转换为可点击链接，并可显式批量发送到 115 云下载。"
      />

      <SectionCard>
        <template #title>链接文本</template>
        <template #description>每行输入一个以 ed2k:// 开头的链接。</template>
        <template #actions>
          <div class="ed2k-tool__actions">
            <el-button :disabled="!ed2kHasInput || batchSubmitting" @click="clearEd2kInput">清空</el-button>
            <el-button
              type="primary"
              :loading="batchSubmitting"
              :disabled="selectedCount === 0 || batchSubmitting"
              @click="submitSelected"
            >
              发送选中到115
            </el-button>
          </div>
        </template>

        <div class="ed2k-tool">
          <el-input
            v-model="ed2kInput"
            type="textarea"
            :rows="8"
            resize="vertical"
            placeholder="每行一个 ed2k:// 链接"
            aria-label="ED2K 链接文本"
            :disabled="batchSubmitting"
            class="ed2k-tool__input"
          />
          <div class="ed2k-tool__summary">
            <span>有效链接：{{ ed2kLinks.length }}</span>
            <span v-if="ed2kInvalidCount > 0">已忽略 {{ ed2kInvalidCount }} 行非 ED2K 文本</span>
            <span v-if="ed2kLinks.length > 0">已选 {{ selectedCount }} / {{ ed2kLinks.length }}</span>
          </div>

          <EmptyState
            v-if="ed2kHasInput && ed2kLinks.length === 0"
            title="未识别到 ED2K 链接"
            description="请输入以 ed2k:// 开头的链接"
          />

          <div v-else-if="ed2kLinks.length > 0" class="ed2k-link-list" aria-label="生成的 ED2K 链接">
            <div class="ed2k-batch-bar">
              <el-checkbox
                :model-value="allSelected"
                :indeterminate="selectionIndeterminate"
                :disabled="batchSubmitting"
                @change="toggleSelectAll"
              >
                {{ ed2kLinks.length > BATCH_LIMIT ? `全选前 ${BATCH_LIMIT} 条` : '全选' }}
              </el-checkbox>
              <span class="ed2k-batch-hint">
                单次最多 {{ BATCH_LIMIT }} 条
                <template v-if="ed2kLinks.length > BATCH_LIMIT">
                  ；全选只选前 {{ BATCH_LIMIT }} 条，取消一条后可改选其他链接
                </template>
              </span>
            </div>

            <div
              v-for="link in ed2kLinks"
              :key="link.id"
              class="ed2k-link-row"
              :class="{ 'ed2k-link-row--selected': isSelected(link) }"
            >
              <el-checkbox
                :model-value="isSelected(link)"
                :aria-label="`选择第 ${link.lineNumber} 行 ${link.label}`"
                :disabled="batchSubmitting || (!isSelected(link) && isSelectionLimitReached)"
                @change="toggleSelection(link)"
              />
              <span class="ed2k-link__line tabular-num">{{ link.lineNumber }}</span>
              <a
                class="ed2k-link__anchor"
                :href="link.href"
                :title="link.href"
                @click="markEd2kLinkClicked(link)"
              >
                <span class="ed2k-link__text">{{ link.label }}</span>
              </a>
              <span class="ed2k-link__status">
                <el-tag v-if="isEd2kLinkClicked(link)" size="small" type="info" effect="plain">已点击</el-tag>
                <el-tag
                  v-if="cloudStatusFor(link)"
                  size="small"
                  :type="cloudStatusFor(link).type"
                  effect="plain"
                  :title="cloudStatusFor(link).detail"
                >
                  {{ cloudStatusFor(link).label }}
                </el-tag>
              </span>
            </div>

            <div v-if="batchSummaryParts.length > 0" class="ed2k-batch-result" role="status" aria-live="polite">
              <span
                v-for="([label, count]) in batchSummaryParts"
                :key="label"
                class="ed2k-batch-result__item"
              >
                {{ label }} {{ count }}
              </span>
              <el-button link type="primary" @click="goToCloudDownloads">查看任务</el-button>
            </div>
          </div>
        </div>
      </SectionCard>
    </div>
  </main>
</template>

<style scoped>
.tool-workspace {
  min-height: 100vh;
  min-height: 100dvh;
  min-width: 0;
  overflow-x: clip;
  background: var(--bg-canvas);
}

.tool-workspace__inner {
  display: grid;
  width: min(100%, 72rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.ed2k-tool {
  display: grid;
  gap: var(--space-3);
}

.ed2k-tool__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
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

.ed2k-batch-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2) var(--space-4);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.ed2k-batch-hint {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link-row {
  display: grid;
  grid-template-columns: auto 3rem minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.ed2k-link-row:hover,
.ed2k-link-row--selected {
  border-color: var(--primary);
  background: var(--bg-surface);
}

.ed2k-link__line {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link__anchor {
  display: flex;
  min-width: 0;
  color: var(--text-primary);
  text-decoration: none;
}

.ed2k-link__anchor:hover {
  color: var(--primary);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.ed2k-link__anchor:focus-visible {
  border-radius: var(--radius-sm);
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
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

.ed2k-link__status {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-1);
}

.ed2k-batch-result {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2) var(--space-3);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  background: var(--bg-surface-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .ed2k-link-row {
    grid-template-columns: auto 2.5rem minmax(0, 1fr);
    align-items: center;
    min-height: 44px;
  }

  .ed2k-link__anchor {
    min-height: 44px;
    align-items: center;
  }

  .ed2k-link__status {
    grid-column: 3;
    justify-content: flex-start;
  }

  .ed2k-tool__actions {
    justify-content: flex-start;
  }

  .ed2k-tool__actions :deep(.el-button),
  .ed2k-batch-result :deep(.el-button),
  .ed2k-batch-bar :deep(.el-checkbox),
  .ed2k-link-row :deep(.el-checkbox) {
    min-height: 44px;
  }
}
</style>
