<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import CloudImportFrame from './CloudImportFrame.vue'
import {
  createAdminCollection,
  createAdminImageCollection,
  getAdminActors,
  getAdminCollections,
  getAdminImageCollections,
  getAdminPopularVideoTags,
  getAdminVideoTags,
  importCloudDownload
} from '../api/admin'
import {
  buildArchiveImportPayload,
  buildVideoImportPayload,
  formatImportFileSize,
  normalizeImportUUIDs,
  resolveCloudImportDefaultTitle,
  validateArchiveImportForm,
  validateVideoImportForm
} from '../views/cloudDownloadImport.helpers'
import { extractApiErrorMessage } from '../views/toolbox.helpers'
import {
  createRemoteSuggestionLoader,
  filterRemoteOptionsByValues,
  mergeRemoteStringOptions,
  mergeRemoteValueOptions
} from '../views/videoUpload.remote'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  inline: { type: Boolean, default: false },
  task: { type: Object, default: null },
  preview: { type: Object, default: null }
})

const emit = defineEmits(['update:modelValue', 'submitted', 'closed', 'dirty', 'submitting', 'submit-error'])

const videoTypeOptions = [
  { value: 'short', label: '短视频' },
  { value: 'movie', label: '电影' },
  { value: 'episode', label: '剧集分集' },
  { value: 'av', label: 'AV' }
]
const avSiteCategoryOptions = [
  { value: 'japanese', label: '日本' },
  { value: 'western', label: '欧美' },
  { value: 'fc2', label: 'FC2' }
]

let initializing = false
const submitting = ref(false)
const tagOptions = ref([])
const loadingTags = ref(false)
const actorOptions = ref([])
const loadingActors = ref(false)
const collectionOptions = ref([])
const loadingCollections = ref(false)
const imageCollectionOptions = ref([])
const loadingImageCollections = ref(false)
const quickCollectionDialogVisible = ref(false)
const quickCollectionSaving = ref(false)
watch([submitting, quickCollectionSaving], ([importing, creating]) => emit('submitting', importing || creating), { flush: 'sync' })
const quickCollectionForm = reactive({
  kind: 'video',
  name: '',
  description: ''
})

const videoForm = reactive({
  type: 'short',
  site_category: 'japanese',
  title: '',
  description: '',
  tags: [],
  actors: [],
  collections: [],
  image_collection_id: ''
})

const archiveForm = reactive({
  title: '',
  default_description: '',
  default_tags: [],
  default_video_collection_ids: [],
  default_image_collection_ids: [],
  has_password: false,
  password: ''
})

watch(archiveForm, () => { if (!initializing && props.inline) emit('dirty', true) }, { deep: true, flush: 'sync' })
onMounted(() => { if (props.inline && props.modelValue) initializeForms() })

const kind = computed(() => (props.preview?.kind === 'archive' ? 'archive' : 'video'))
const isAVType = computed(() => videoForm.type === 'av')
const isShortType = computed(() => videoForm.type === 'short')
const dialogTitle = computed(() => (kind.value === 'archive' ? '压缩包入媒体库' : '视频入媒体库'))
const quickCollectionDialogTitle = computed(() =>
  quickCollectionForm.kind === 'image' ? '新建图片合集' : '新建视频合集'
)
const fileName = computed(() => String(props.preview?.name || props.task?.declared_name || '').trim() || '--')
const fileSize = computed(() => formatImportFileSize(props.preview?.size))
const fileMime = computed(() => String(props.preview?.mime || '').trim())

watch(
  () => videoForm.type,
  (nextType) => {
    if (nextType !== 'short' && videoForm.collections.length > 0) {
      videoForm.collections = []
    }
    if (nextType === 'av' && !videoForm.site_category) {
      videoForm.site_category = 'japanese'
    }
  }
)

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      initializeForms()
    }
  }
)

function selectedTagValues() {
  return [...videoForm.tags, ...archiveForm.default_tags]
}

function selectedVideoCollectionValues() {
  return [...videoForm.collections, ...archiveForm.default_video_collection_ids]
}

function selectedImageCollectionValues() {
  const values = [...archiveForm.default_image_collection_ids]
  if (videoForm.image_collection_id) values.push(videoForm.image_collection_id)
  return values
}

async function searchTags(keyword = '') {
  if (!keyword) {
    const data = await getAdminPopularVideoTags({ limit: 5 })
    return (data.items || [])
      .map((item) => String(item.tag || '').trim().toLowerCase())
      .filter((tag) => tag !== '')
  }
  const data = await getAdminVideoTags({ q: keyword, limit: 12 })
  return (data.items || [])
    .map((item) => String(item.tag || '').trim().toLowerCase())
    .filter((tag) => tag !== '')
}

async function searchActors(keyword = '') {
  loadingActors.value = true
  try {
    const data = await getAdminActors({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 20
    })
    const incoming = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
    actorOptions.value = mergeRemoteValueOptions(
      filterRemoteOptionsByValues(actorOptions.value, videoForm.actors, (item) => item.value),
      incoming
    )
  } finally {
    loadingActors.value = false
  }
}

async function searchCollections(keyword = '') {
  const data = await getAdminCollections({
    q: keyword,
    active: 1,
    page: 1,
    page_size: 20
  })
  return (data.items || []).map((item) => ({
    value: item.id,
    label: item.name
  }))
}

async function searchImageCollections(keyword = '') {
  const data = await getAdminImageCollections({
    q: keyword,
    active: 1,
    page: 1,
    page_size: 20
  })
  return (data.items || []).map((item) => ({
    value: item.id,
    label: item.name
  }))
}

const loadTagSuggestions = createRemoteSuggestionLoader({
  fetcher: searchTags,
  getOptions: () => filterRemoteOptionsByValues(tagOptions.value, selectedTagValues()),
  setOptions: (next) => {
    tagOptions.value = next
  },
  setLoading: (next) => {
    loadingTags.value = next
  },
  mergeOptions: mergeRemoteStringOptions
})

const loadCollectionSuggestions = createRemoteSuggestionLoader({
  fetcher: searchCollections,
  getOptions: () => filterRemoteOptionsByValues(
    collectionOptions.value,
    selectedVideoCollectionValues(),
    (item) => item.value
  ),
  setOptions: (next) => {
    collectionOptions.value = next
  },
  setLoading: (next) => {
    loadingCollections.value = next
  },
  mergeOptions: mergeRemoteValueOptions
})

const loadImageCollectionSuggestions = createRemoteSuggestionLoader({
  fetcher: searchImageCollections,
  getOptions: () => filterRemoteOptionsByValues(
    imageCollectionOptions.value,
    selectedImageCollectionValues(),
    (item) => item.value
  ),
  setOptions: (next) => {
    imageCollectionOptions.value = next
  },
  setLoading: (next) => {
    loadingImageCollections.value = next
  },
  mergeOptions: mergeRemoteValueOptions
})

function openQuickCollectionDialog(collectionKind) {
  if (submitting.value || quickCollectionSaving.value) return
  quickCollectionForm.kind = collectionKind
  quickCollectionForm.name = ''
  quickCollectionForm.description = ''
  quickCollectionDialogVisible.value = true
}

function openCreateVideoCollection() {
  openQuickCollectionDialog('video')
}

function openCreateImageCollection() {
  openQuickCollectionDialog('image')
}

function appendCreatedCollection(collectionKind, created) {
  const option = {
    value: created.id,
    label: created.name || quickCollectionForm.name.trim()
  }
  if (collectionKind === 'image') {
    imageCollectionOptions.value = mergeRemoteValueOptions(imageCollectionOptions.value, [option])
    archiveForm.default_image_collection_ids = normalizeImportUUIDs([
      ...archiveForm.default_image_collection_ids,
      created.id
    ])
    return
  }
  collectionOptions.value = mergeRemoteValueOptions(collectionOptions.value, [option])
  archiveForm.default_video_collection_ids = normalizeImportUUIDs([
    ...archiveForm.default_video_collection_ids,
    created.id
  ])
}

async function saveQuickCollection() {
  if (quickCollectionSaving.value || submitting.value) return
  const name = quickCollectionForm.name.trim()
  if (!name) {
    ElMessage.warning(`请输入${quickCollectionForm.kind === 'image' ? '图片合集' : '视频合集'}名称`)
    return
  }

  quickCollectionSaving.value = true
  try {
    const payload = {
      name,
      description: quickCollectionForm.description.trim(),
      cover_url: '',
      sort_order: 0,
      active: true
    }
    if (quickCollectionForm.kind === 'image') {
      payload.cover_image_id = null
    }
    const created = quickCollectionForm.kind === 'image'
      ? await createAdminImageCollection(payload)
      : await createAdminCollection(payload)
    appendCreatedCollection(quickCollectionForm.kind, created)
    quickCollectionDialogVisible.value = false
    ElMessage.success(`${quickCollectionForm.kind === 'image' ? '图片合集' : '视频合集'}已创建并选中`)
  } catch (error) {
    ElMessage.error(extractApiErrorMessage(
      error,
      quickCollectionForm.kind === 'image' ? '创建图片合集失败' : '创建视频合集失败'
    ))
  } finally {
    quickCollectionSaving.value = false
  }
}

function resetVideoForm() {
  videoForm.type = 'short'
  videoForm.site_category = 'japanese'
  videoForm.title = ''
  videoForm.description = ''
  videoForm.tags = []
  videoForm.actors = []
  videoForm.collections = []
  videoForm.image_collection_id = ''
}

function resetArchiveForm() {
  archiveForm.title = ''
  archiveForm.default_description = ''
  archiveForm.default_tags = []
  archiveForm.default_video_collection_ids = []
  archiveForm.default_image_collection_ids = []
  archiveForm.has_password = false
  archiveForm.password = ''
}

function initializeForms() {
  initializing = true
  resetVideoForm()
  resetArchiveForm()
  const defaultTitle = resolveCloudImportDefaultTitle(props.task, props.preview)
  videoForm.title = defaultTitle
  archiveForm.title = defaultTitle
  tagOptions.value = []
  actorOptions.value = []
  collectionOptions.value = []
  imageCollectionOptions.value = []
  loadTagSuggestions('')
  loadCollectionSuggestions('')
  loadImageCollectionSuggestions('')
  if (!props.inline) searchActors().catch(() => {})
  initializing = false
  emit('dirty', false)
}

function closeDialog() {
  if (submitting.value || quickCollectionSaving.value) return
  emit('update:modelValue', false)
}

function handleVisibleChange(visible) {
  if (!visible) {
    closeDialog()
    return
  }
  emit('update:modelValue', true)
}

function handleBeforeClose(done) {
  if (submitting.value || quickCollectionSaving.value) return
  done()
}

function handleClosed() {
  submitting.value = false
  resetVideoForm()
  resetArchiveForm()
  emit('closed')
}

async function submitImport() {
  if (submitting.value || quickCollectionSaving.value) return
  const taskId = props.task?.id
  if (!taskId) {
    ElMessage.error('缺少云下载任务')
    return
  }

  let payload
  if (kind.value === 'archive') {
    const validation = validateArchiveImportForm(archiveForm)
    if (validation) {
      ElMessage.warning(validation)
      return
    }
    payload = buildArchiveImportPayload(archiveForm)
  } else {
    const validation = validateVideoImportForm(videoForm)
    if (validation) {
      ElMessage.warning(validation)
      return
    }
    payload = buildVideoImportPayload(videoForm)
  }

  submitting.value = true
  try {
    const data = await importCloudDownload(taskId, payload)
    ElMessage.success('已提交导入任务')
    emit('submitted', data?.import || null)
    emit('update:modelValue', false)
  } catch (error) {
    ElMessage.error(extractApiErrorMessage(error, '提交导入任务失败'))
    emit('submit-error', error)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <CloudImportFrame
    :inline="inline"
    :model-value="modelValue"
    :title="dialogTitle"
    width="min(94vw, 760px)"
    destroy-on-close
    data-density="form"
    :before-close="handleBeforeClose"
    @update:model-value="handleVisibleChange"
    @closed="handleClosed"
  >
    <div class="cloud-import-dialog" :inert="submitting || quickCollectionSaving">
      <p class="cloud-import-dialog__lead">文件已下载到本地，提交后由后台导入媒体库，不会重新上传文件。</p>
      <div class="cloud-import-file">
        <span class="cloud-import-file__name">{{ fileName }}</span>
        <span class="cloud-import-file__meta">
          <span>{{ fileSize }}</span>
          <span v-if="fileMime">格式：{{ fileMime }}</span>
        </span>
      </div>

      <el-form v-if="kind === 'video'" label-width="104px" class="cloud-import-form">
        <el-form-item label="类型">
          <el-select v-model="videoForm.type" class="cloud-import-control">
            <el-option
              v-for="item in videoTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isAVType" label="AV 地区分类">
          <el-radio-group v-model="videoForm.site_category">
            <el-radio-button
              v-for="item in avSiteCategoryOptions"
              :key="item.value"
              :value="item.value"
            >
              {{ item.label }}
            </el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="videoForm.title" placeholder="默认取云下载文件名" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="videoForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="视频标签">
          <el-select
            v-model="videoForm.tags"
            class="cloud-import-control"
            multiple
            filterable
            remote
            reserve-keyword
            allow-create
            default-first-option
            clearable
            :remote-method="loadTagSuggestions"
            :loading="loadingTags"
            placeholder="可选择或输入标签"
          >
            <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片图集">
          <el-select
            v-model="videoForm.image_collection_id"
            class="cloud-import-control"
            filterable
            remote
            reserve-keyword
            clearable
            :remote-method="loadImageCollectionSuggestions"
            :loading="loadingImageCollections"
            placeholder="可选，仅可关联一个图片图集"
          >
            <el-option
              v-for="collection in imageCollectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isShortType" label="所属合集">
          <el-select
            v-model="videoForm.collections"
            class="cloud-import-control"
            multiple
            filterable
            remote
            reserve-keyword
            clearable
            collapse-tags
            collapse-tags-tooltip
            :remote-method="loadCollectionSuggestions"
            :loading="loadingCollections"
            placeholder="可选，可多选"
          >
            <el-option
              v-for="collection in collectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关联演员">
          <el-select
            v-model="videoForm.actors"
            class="cloud-import-control"
            multiple
            filterable
            remote
            reserve-keyword
            allow-create
            default-first-option
            clearable
            :remote-method="searchActors"
            :loading="loadingActors"
            placeholder="可搜索演员，也可直接输入新演员姓名"
          >
            <el-option
              v-for="actor in actorOptions"
              :key="actor.value"
              :label="actor.label"
              :value="actor.value"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <el-form v-else label-width="116px" class="cloud-import-form">
        <el-form-item label="视频默认标题">
          <el-input v-model="archiveForm.title" placeholder="默认取云下载文件名" />
        </el-form-item>
        <el-form-item label="默认说明">
          <el-input v-model="archiveForm.default_description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="默认标签">
          <el-select
            v-model="archiveForm.default_tags"
            class="cloud-import-control"
            multiple
            filterable
            remote
            reserve-keyword
            allow-create
            default-first-option
            clearable
            :remote-method="loadTagSuggestions"
            :loading="loadingTags"
            placeholder="可选择或输入标签"
          >
            <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="默认视频合集">
          <div class="cloud-import-collection-row">
            <el-select
              v-model="archiveForm.default_video_collection_ids"
              class="cloud-import-control cloud-import-collection-select"
              multiple
              filterable
              remote
              reserve-keyword
              clearable
              default-first-option
              collapse-tags
              collapse-tags-tooltip
              :remote-method="loadCollectionSuggestions"
              :loading="loadingCollections"
              placeholder="可选，可多选"
            >
              <el-option
                v-for="collection in collectionOptions"
                :key="collection.value"
                :label="collection.label"
                :value="collection.value"
              />
            </el-select>
            <el-button class="cloud-import-collection-button" @click="openCreateVideoCollection">
              新建视频合集
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="默认图片合集">
          <div class="cloud-import-collection-row">
            <el-select
              v-model="archiveForm.default_image_collection_ids"
              class="cloud-import-control cloud-import-collection-select"
              multiple
              filterable
              remote
              reserve-keyword
              clearable
              default-first-option
              collapse-tags
              collapse-tags-tooltip
              :remote-method="loadImageCollectionSuggestions"
              :loading="loadingImageCollections"
              placeholder="可选，可多选"
            >
              <el-option
                v-for="collection in imageCollectionOptions"
                :key="collection.value"
                :label="collection.label"
                :value="collection.value"
              />
            </el-select>
            <el-button class="cloud-import-collection-button" @click="openCreateImageCollection">
              新建图片合集
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="有密码">
          <el-switch v-model="archiveForm.has_password" />
        </el-form-item>
        <el-form-item v-if="archiveForm.has_password" label="压缩包密码">
          <el-input
            v-model="archiveForm.password"
            type="password"
            show-password
            placeholder="由管理员填写密码"
          />
        </el-form-item>
      </el-form>
    </div>

    <CloudImportFrame
      :inline="inline"
      v-model="quickCollectionDialogVisible"
      :title="quickCollectionDialogTitle"
      width="min(94vw, 520px)"
      append-to-body
      destroy-on-close
      data-density="form"
    >
      <el-form label-width="104px" class="cloud-import-quick-collection-form" :disabled="quickCollectionSaving || submitting">
        <el-form-item :label="quickCollectionForm.kind === 'image' ? '图片合集名称' : '视频合集名称'">
          <el-input
            v-model="quickCollectionForm.name"
            placeholder="请输入合集名称"
            @keyup.enter="saveQuickCollection"
          />
        </el-form-item>
        <el-form-item label="合集简介">
          <el-input
            v-model="quickCollectionForm.description"
            type="textarea"
            :rows="3"
            placeholder="可选，简介用于后台识别"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :disabled="quickCollectionSaving || submitting" @click="quickCollectionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="quickCollectionSaving" @click="saveQuickCollection">
          创建并选中
        </el-button>
      </template>
    </CloudImportFrame>

    <template #footer>
      <el-button :disabled="submitting || quickCollectionSaving" @click="closeDialog">取消</el-button>
      <el-button type="primary" :loading="submitting" :disabled="submitting || quickCollectionSaving" @click="submitImport">
        提交导入
      </el-button>
    </template>
  </CloudImportFrame>
</template>

<style scoped>
.cloud-import-dialog {
  display: grid;
  gap: var(--space-3);
}

.cloud-import-dialog__lead {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.cloud-import-file {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-2) var(--space-4);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.cloud-import-file__name {
  min-width: 0;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.cloud-import-file__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  color: var(--text-muted);
}

.cloud-import-control {
  width: 100%;
}

.cloud-import-collection-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  min-width: 0;
}

.cloud-import-collection-select {
  flex: 1;
  min-width: 0;
}

.cloud-import-collection-button {
  flex: 0 0 auto;
}

@media (max-width: 48rem) {
  .cloud-import-form :deep(.el-form-item) {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
  }

  .cloud-import-form :deep(.el-form-item__label) {
    width: auto !important;
    justify-content: flex-start;
  }

  .cloud-import-form :deep(.el-form-item__content) {
    min-width: 0;
    margin-left: 0 !important;
  }

  .cloud-import-collection-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
