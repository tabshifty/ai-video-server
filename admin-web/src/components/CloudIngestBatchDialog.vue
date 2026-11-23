<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  cloudDownloadStatuses,
  createAdminCollection,
  createCloudIngestBatch,
  getAdminActors,
  getAdminCollections,
  getAdminImageCollections,
  getAdminVideoTags
} from '../api/admin'
import {
  CLOUD_INGEST_SITE_CATEGORIES,
  CLOUD_INGEST_TITLE_MAX_RUNES,
  CLOUD_INGEST_TITLE_MODES,
  CLOUD_INGEST_VIDEO_TYPES,
  buildEd2kIngestLinks,
  buildEd2kIngestPayload,
  ed2kIngestKey,
  validateEd2kIngestForm
} from '../views/ed2kIngest.helpers'
import { extractApiErrorMessage } from '../views/toolbox.helpers'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  plan: { type: Object, default: () => ({ videos: [], notVideos: [], unknown: [] }) },
  duplicateCount: { type: Number, default: 0 },
  invalidCount: { type: Number, default: 0 }
})

const emit = defineEmits(['update:modelValue', 'submitted'])

const PREVIEW_PAGE_SIZE = 10

const form = reactive({
  titleMode: 'filename',
  uniformTitle: '',
  videoType: 'short',
  siteCategory: 'japanese',
  description: '',
  tags: [],
  collectionIDs: [],
  actorNames: [],
  imageCollectionID: ''
})

const tagOptions = ref([])
const actorOptions = ref([])
const collectionOptions = ref([])
const imageCollectionOptions = ref([])
const forcedKeys = ref(new Set())
const statuses = ref({})
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const previewPage = ref(1)

const isShortType = computed(() => form.videoType === 'short')
const isAVType = computed(() => form.videoType === 'av')
const countUnknown = computed(() => props.plan.unknown.length)
const countExcluded = computed(() => props.plan.notVideos.length)
const countVideos = computed(() => props.plan.videos.length)
const countSelected = computed(() => countVideos.value + countExcluded.value + countUnknown.value)

const items = computed(() => buildEd2kIngestLinks(props.plan, { forcedKeys: forcedKeys.value }))
const totalPages = computed(() => Math.max(1, Math.ceil(items.value.length / PREVIEW_PAGE_SIZE)))
const pagedItems = computed(() => {
  const start = (previewPage.value - 1) * PREVIEW_PAGE_SIZE
  return items.value.slice(start, start + PREVIEW_PAGE_SIZE)
})

function displayTitle(item) {
  if (!item) {
    return ''
  }
  return form.titleMode === 'uniform'
    ? String(form.uniformTitle || '').trim()
    : String(item.title || '')
}

function statusLabel(item) {
  const status = statuses.value[item.infoHash]
  if (!status) {
    return item.forced ? '强制放行' : '新提交'
  }
  switch (status.status) {
    case 'completed':
      return '已下载·待入库'
    case 'failed':
    case 'blocked':
      return '将重试'
    default:
      return '下载中·待入库'
  }
}

function toggleForce(item) {
  const key = ed2kIngestKey(item)
  const next = new Set(forcedKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  forcedKeys.value = next
}

function forceAll() {
  forcedKeys.value = new Set(props.plan.unknown.map((item) => ed2kIngestKey(item)))
}

function clearForce() {
  forcedKeys.value = new Set()
}

function resetForm() {
  form.titleMode = 'filename'
  form.uniformTitle = ''
  form.videoType = 'short'
  form.siteCategory = 'japanese'
  form.description = ''
  form.tags = []
  form.collectionIDs = []
  form.actorNames = []
  form.imageCollectionID = ''
  forcedKeys.value = new Set()
  errorMessage.value = ''
  previewPage.value = 1
}

async function loadOptions() {
  const [tags, actors, collections, imageCollections] = await Promise.all([
    getAdminVideoTags({ page: 1, page_size: 100 }).catch(() => ({ items: [] })),
    getAdminActors({ page: 1, page_size: 100 }).catch(() => ({ items: [] })),
    getAdminCollections({ active: 1, page: 1, page_size: 100 }).catch(() => ({ items: [] })),
    getAdminImageCollections({ active: 1, page: 1, page_size: 100 }).catch(() => ({ items: [] }))
  ])
  tagOptions.value = (tags.items || []).map((item) => ({ value: item.name || item.value, label: item.name || item.value }))
  actorOptions.value = (actors.items || []).map((item) => ({ value: item.name, label: item.name }))
  collectionOptions.value = (collections.items || []).map((item) => ({ value: item.id, label: item.name }))
  imageCollectionOptions.value = (imageCollections.items || []).map((item) => ({
    value: item.id,
    label: item.name
  }))
}

async function loadStatuses() {
  const hashes = [...props.plan.videos, ...props.plan.unknown]
    .map((item) => item.infoHash)
    .filter(Boolean)
  if (hashes.length === 0) {
    statuses.value = {}
    return
  }
  try {
    const data = await cloudDownloadStatuses(hashes)
    statuses.value = data.statuses || {}
  } catch {
    // Status lookup is advisory: the submit endpoint re-checks conflicts.
    statuses.value = {}
  }
}

async function onOpen() {
  resetForm()
  loading.value = true
  try {
    await Promise.all([loadOptions(), loadStatuses()])
  } finally {
    loading.value = false
  }
}

function closeDialog() {
  emit('update:modelValue', false)
}

function handleBeforeClose(done) {
  if (submitting.value) {
    return
  }
  done()
}

watch(
  () => form.videoType,
  (type) => {
    if (type !== 'short' && form.collectionIDs.length > 0) {
      form.collectionIDs = []
    }
    if (type === 'av' && !form.siteCategory) {
      form.siteCategory = 'japanese'
    }
  }
)

watch(
  () => items.value.length,
  () => {
    if (previewPage.value > totalPages.value) {
      previewPage.value = totalPages.value
    }
  }
)

async function createCollection() {
  const name = window.prompt('新建视频合集名称')
  if (!name || !name.trim()) {
    return
  }
  try {
    const created = await createAdminCollection({
      name: name.trim(),
      description: '',
      cover_url: '',
      sort_order: 0,
      active: true
    })
    const id = created?.id || created?.data?.id
    if (id) {
      collectionOptions.value = [...collectionOptions.value, { value: id, label: name.trim() }]
      form.collectionIDs = [...form.collectionIDs, id]
      ElMessage.success('合集已创建')
    }
  } catch (error) {
    ElMessage.error(extractApiErrorMessage(error, '创建合集失败'))
  }
}

async function submit() {
  if (items.value.length === 0) {
    ElMessage.warning('没有可提交的链接')
    return
  }
  const invalid = validateEd2kIngestForm(form)
  if (invalid) {
    ElMessage.warning(invalid)
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const payload = buildEd2kIngestPayload({ links: items.value, form })
    const data = await createCloudIngestBatch(payload)
    const summary = data?.summary || {}
    const forced = summary.forced || 0
    ElMessage.success(
      forced > 0
        ? `已创建入库批，共 ${summary.total || items.value.length} 条（其中强制放行 ${forced} 条，下载后仍可能入库失败）`
        : `已创建入库批，共 ${summary.total || items.value.length} 条`
    )
    emit('submitted', { batchId: data?.batch?.id || '', result: data })
    emit('update:modelValue', false)
  } catch (error) {
    errorMessage.value = extractApiErrorMessage(error, '批量入库失败')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :width="'min(94vw, 900px)'"
    data-density="form"
    title="批量入库到媒体库"
    :close-on-click-modal="!submitting"
    :before-close="handleBeforeClose"
    @open="onOpen"
    @closed="resetForm"
  >
    <div class="ingest-dialog">
      <div class="ingest-stats" aria-label="本批统计">
        <span>共选 {{ countSelected }}</span>
        <span>视频 {{ countVideos }}</span>
        <span>排除 {{ countExcluded }}</span>
        <span>无法识别 {{ countUnknown }}</span>
        <span v-if="duplicateCount">重复 {{ duplicateCount }}</span>
        <span v-if="invalidCount">无效 {{ invalidCount }}</span>
        <span v-if="items.length">将提交 {{ items.length }}</span>
        <span v-if="loading">正在核对云端…</span>
      </div>

      <el-collapse v-if="countExcluded" class="ingest-excluded">
        <el-collapse-item :title="`已排除 ${countExcluded} 条（不是单文件视频，不会提交）`" name="excluded">
          <ul class="ingest-list">
            <li v-for="item in plan.notVideos" :key="ed2kIngestKey(item)">
              <span class="ingest-list__name">{{ item.label }}</span>
              <span class="ingest-list__reason">{{ item.reason }}</span>
            </li>
          </ul>
        </el-collapse-item>
      </el-collapse>

      <section v-if="countUnknown" class="ingest-unknown">
        <header class="ingest-unknown__header">
          <strong>无法识别 {{ countUnknown }} 条</strong>
          <span class="ingest-unknown__hint">后缀未知，可能下载完成后仍无法入库</span>
          <div class="ingest-unknown__actions">
            <el-button size="small" @click="forceAll">全部强制按视频</el-button>
            <el-button size="small" :disabled="forcedKeys.size === 0" @click="clearForce">
              全部取消
            </el-button>
          </div>
        </header>
        <div class="ingest-unknown__body">
          <el-checkbox
            v-for="item in plan.unknown"
            :key="ed2kIngestKey(item)"
            :model-value="forcedKeys.has(ed2kIngestKey(item))"
            :aria-label="`强制按视频：${item.label}`"
            @change="toggleForce(item)"
          >
            <span class="ingest-list__name">{{ item.label }}</span>
            <span class="ingest-list__reason">{{ item.reason }}</span>
          </el-checkbox>
        </div>
      </section>

      <el-form label-width="96px" class="ingest-form">
        <el-form-item label="视频类型">
          <el-select v-model="form.videoType" aria-label="视频类型">
            <el-option
              v-for="option in CLOUD_INGEST_VIDEO_TYPES"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isAVType" label="AV 地区">
          <el-radio-group v-model="form.siteCategory">
            <el-radio-button
              v-for="option in CLOUD_INGEST_SITE_CATEGORIES"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标题模式">
          <el-radio-group v-model="form.titleMode">
            <el-radio
              v-for="option in CLOUD_INGEST_TITLE_MODES"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.titleMode === 'uniform'" label="统一标题">
          <el-input
            v-model="form.uniformTitle"
            :maxlength="CLOUD_INGEST_TITLE_MAX_RUNES"
            show-word-limit
            placeholder="整批共用的标题"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
        <el-form-item label="视频标签">
          <el-select
            v-model="form.tags"
            multiple
            filterable
            allow-create
            default-first-option
            aria-label="视频标签"
            placeholder="可选"
          >
            <el-option
              v-for="option in tagOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isShortType" label="所属合集">
          <div class="ingest-collection">
            <el-select
              v-model="form.collectionIDs"
              multiple
              filterable
              aria-label="所属视频合集"
              placeholder="仅短视频可用"
            >
              <el-option
                v-for="option in collectionOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-button text type="primary" @click="createCollection">新建视频合集</el-button>
          </div>
        </el-form-item>
        <el-form-item label="关联演员">
          <el-select
            v-model="form.actorNames"
            multiple
            filterable
            allow-create
            default-first-option
            aria-label="关联演员"
            placeholder="可选"
          >
            <el-option
              v-for="option in actorOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="图片图集">
          <el-select
            v-model="form.imageCollectionID"
            filterable
            clearable
            aria-label="图片图集"
            placeholder="可选，只能选择已有图集"
          >
            <el-option
              v-for="option in imageCollectionOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <section class="ingest-preview">
        <header class="ingest-preview__header">
          <strong>将提交 {{ items.length }} 条</strong>
          <span>每条最终标题（提交后仍可在视频详情里修改）</span>
        </header>
        <div class="table-wrap">
          <el-table :data="pagedItems" size="small" empty-text="没有可提交的链接">
            <el-table-column label="声明文件名" min-width="160" show-overflow-tooltip>
              <template #default="{ row }">{{ row.label }}</template>
            </el-table-column>
            <el-table-column label="最终标题" min-width="180">
              <template #default="{ row }">
                <span class="ingest-preview__title">{{ displayTitle(row) || '（空，将回退文件名）' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="类型" width="90">
              <template #default="{ row }">
                <el-tag v-if="row.forced" size="small" type="warning" effect="plain">强制</el-tag>
                <el-tag v-else size="small" effect="plain">视频</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="核对" width="120">
              <template #default="{ row }">{{ statusLabel(row) }}</template>
            </el-table-column>
          </el-table>
        </div>
        <div v-if="totalPages > 1" class="ingest-preview__pager">
          <el-button size="small" :disabled="previewPage <= 1" @click="previewPage -= 1">上一页</el-button>
          <span>{{ previewPage }} / {{ totalPages }}</span>
          <el-button size="small" :disabled="previewPage >= totalPages" @click="previewPage += 1">
            下一页
          </el-button>
        </div>
      </section>

      <el-alert v-if="errorMessage" :title="errorMessage" type="error" :closable="false" show-icon />
    </div>

    <template #footer>
      <el-button :disabled="submitting" @click="closeDialog">取消</el-button>
      <el-button type="primary" :loading="submitting" :disabled="loading || items.length === 0" @click="submit">
        提交入库
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.ingest-dialog {
  display: grid;
  gap: 12px;
}

.ingest-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.ingest-list {
  margin: 0;
  padding-left: 16px;
}

.ingest-list li {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 2px 0;
}

.ingest-list__name {
  word-break: break-all;
}

.ingest-list__reason {
  color: var(--text-secondary);
}

.ingest-unknown {
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 8px 12px;
}

.ingest-unknown__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.ingest-unknown__hint {
  color: var(--warning-text, #b8860b);
  font-size: 12px;
}

.ingest-unknown__actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.ingest-unknown__body {
  max-height: 40vh;
  overflow-y: auto;
  overscroll-behavior: contain;
  display: grid;
  gap: 4px;
  margin-top: 8px;
}

.ingest-collection {
  display: flex;
  gap: 8px;
  width: 100%;
  align-items: center;
}

.ingest-preview__header {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: baseline;
  margin-bottom: 6px;
}

.ingest-preview__header span {
  color: var(--text-secondary);
  font-size: 12px;
}

.ingest-preview__title {
  overflow-wrap: anywhere;
}

.ingest-preview__pager {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

@media (max-width: 63.9375rem) {
  .ingest-unknown__body :deep(.el-checkbox),
  .ingest-unknown__actions :deep(.el-button),
  .ingest-preview__pager :deep(.el-button) {
    min-height: 44px;
  }
}
</style>
