<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import UploadProgress from '../components/UploadProgress.vue'
import { checkUpload, uploadAbort, uploadChunk, uploadComplete, uploadInit } from '../api/video'
import { getAdminActors, getAdminCollections, getAdminImageCollections, getAdminPopularVideoTags, getAdminVideoTags } from '../api/admin'
import { sha256File } from '../utils/hash'
import { createRemoteSuggestionLoader, mergeRemoteStringOptions, mergeRemoteValueOptions } from './videoUpload.remote'

const uploadFileList = ref([])
const uploadRef = ref(null)
const progress = ref(0)
const hashProgress = ref(0)
const uploading = ref(false)
const abortController = ref(null)
const sessionId = ref('')
const totalFiles = ref(0)
const currentFileName = ref('')
const uploadResults = ref([])
const cancelRequested = ref(false)
const actorOptions = ref([])
const loadingActors = ref(false)
const collectionOptions = ref([])
const loadingCollections = ref(false)
const imageCollectionOptions = ref([])
const loadingImageCollections = ref(false)
const tagOptions = ref([])
const loadingTags = ref(false)
const previousType = ref('short')
const revertingType = ref(false)

const form = reactive({
  type: 'short',
  title: '',
  description: '',
  tags: [],
  actors: [],
  collections: [],
  imageCollectionID: '',
  siteCategory: 'japanese'
})
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

const isMovieType = computed(() => form.type === 'movie')
const isShortType = computed(() => form.type === 'short')
const isAVType = computed(() => form.type === 'av')
const avSiteCategoryOptions = [
  { value: 'japanese', label: '日本' },
  { value: 'western', label: '欧美' },
  { value: 'fc2', label: 'FC2' }
]
const selectedFiles = computed(() =>
  uploadFileList.value
    .map((item) => item.raw)
    .filter((raw) => raw instanceof File)
)
const canClearSelectedFiles = computed(() => !uploading.value && selectedFiles.value.length > 0)
const completedCount = computed(() => uploadResults.value.length)
const successCount = computed(() => uploadResults.value.filter((item) => item.status === 'success').length)
const failedCount = computed(() => uploadResults.value.filter((item) => item.status === 'failed').length)
const cancelledCount = computed(() => uploadResults.value.filter((item) => item.status === 'cancelled').length)
const canClearRecords = computed(() =>
  !uploading.value &&
  (uploadResults.value.length > 0 ||
    progress.value > 0 ||
    hashProgress.value > 0 ||
    totalFiles.value > 0 ||
    currentFileName.value !== '' ||
    sessionId.value !== '')
)
const batchProgress = computed(() => {
  if (totalFiles.value <= 0) return 0
  const currentProgress = uploading.value && !cancelRequested.value ? progress.value / 100 : 0
  return Math.min(100, Math.round(((completedCount.value + currentProgress) / totalFiles.value) * 100))
})
watch(
  () => form.type,
  async (nextType, prevType) => {
    if (revertingType.value) {
      previousType.value = nextType
      revertingType.value = false
      return
    }
    if (nextType !== 'short' && form.collections.length > 0) {
      form.collections = []
    }
    if (nextType === 'av' && !form.siteCategory) {
      form.siteCategory = 'japanese'
    }
    if (nextType !== 'movie' || uploadFileList.value.length <= 1) {
      previousType.value = nextType
      return
    }
    const droppedCount = uploadFileList.value.length - 1
    try {
      await ElMessageBox.confirm(
        `电影仅支持单文件上传。确认切换为电影并丢弃其余 ${droppedCount} 个文件吗？`,
        '确认丢弃文件',
        {
          type: 'warning',
          confirmButtonText: '确认丢弃',
          cancelButtonText: '继续保留'
        }
      )
      uploadFileList.value = [uploadFileList.value[0]]
      previousType.value = nextType
      ElMessage.warning('已保留第1个文件')
    } catch (_) {
      revertingType.value = true
      form.type = prevType || previousType.value || 'short'
    }
  }
)

function getFileKey(raw) {
  return `${raw.name}__${raw.size}__${raw.lastModified}`
}

function onFileChange(_file, fileList) {
  const normalized = []
  const seen = new Set()
  for (const item of fileList) {
    const raw = item.raw
    if (!(raw instanceof File)) continue
    const key = getFileKey(raw)
    if (seen.has(key)) continue
    seen.add(key)
    normalized.push(item)
  }
  if (isMovieType.value && normalized.length > 1) {
    uploadFileList.value = [normalized[0]]
    ElMessage.warning('电影仅支持单文件上传，已保留第1个文件')
    return
  }
  uploadFileList.value = normalized
}

function clearSelectedFiles() {
  if (uploading.value) return
  uploadRef.value?.clearFiles()
  uploadFileList.value = []
}

function splitActorSelection(values) {
  const actorIDs = []
  const actorNames = []
  const seenID = new Set()
  const seenName = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value) continue
    if (uuidPattern.test(value)) {
      const key = value.toLowerCase()
      if (seenID.has(key)) continue
      seenID.add(key)
      actorIDs.push(value)
      continue
    }
    const key = value.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(value.replace(/\s+/g, ' '))
  }
  return { actorIDs, actorNames }
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
    actorOptions.value = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
  } finally {
    loadingActors.value = false
  }
}

function normalizeCollectionSelection(values) {
  const out = []
  const seen = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value) continue
    const key = value.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
  }
  return out
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

async function searchTags(keyword = '') {
  if (!keyword) {
    const data = await getAdminPopularVideoTags({ limit: 5 })
    const items = data.items || []
    return items
      .map((item) => String(item.tag || '').trim().toLowerCase())
      .filter((tag) => tag !== '')
  }

  const data = await getAdminVideoTags({ q: keyword, limit: 12 })
  return (data.items || [])
    .map((item) => String(item.tag || '').trim().toLowerCase())
    .filter((tag) => tag !== '')
}

function resolveImageCollectionID(raw) {
  const value = String(raw || '').trim()
  if (!value || !uuidPattern.test(value)) return ''
  return value
}

function buildNormalizedTags() {
  return Array.from(
    new Map(
      (form.tags || [])
        .map((tag) => String(tag).trim())
        .filter((tag) => tag !== '')
        .map((tag) => [tag.toLowerCase(), tag.toLowerCase()])
    ).values()
  )
}

function pushResult(file, status, message, videoId = '') {
  uploadResults.value.push({
    name: file.name,
    status,
    message,
    videoId
  })
}

function parseErrorMessage(err) {
  if (cancelRequested.value) return '已取消'
  if (err?.code === 'ERR_CANCELED' || err?.name === 'CanceledError') return '已取消'
  const text = String(err?.message || '').trim()
  return text || '上传失败'
}

async function uploadOneFile(targetFile, sharedPayload) {
  progress.value = 0
  hashProgress.value = 0
  currentFileName.value = targetFile.name
  abortController.value = new AbortController()
  let activeSessionID = ''

  try {
    const hash = await sha256File(targetFile, (p) => (hashProgress.value = p))
    if (cancelRequested.value) {
      return { status: 'cancelled', message: '已取消', videoId: '' }
    }

    const check = await checkUpload({ hash, file_size: targetFile.size })
    if (check.exists) {
      return { status: 'success', message: '秒传命中', videoId: check.video_id || '' }
    }

    const chunkSize = 4 * 1024 * 1024
    const totalChunks = Math.ceil(targetFile.size / chunkSize)
    const initResp = await uploadInit({
      filename: targetFile.name,
      file_size: targetFile.size,
      chunk_size: chunkSize,
      total_chunks: totalChunks,
      hash,
      type: sharedPayload.type,
      title: sharedPayload.title,
      description: sharedPayload.description,
      tags: sharedPayload.tags,
      actor_ids: sharedPayload.actorIDs,
      actor_names: sharedPayload.actorNames,
      image_collection_id: sharedPayload.imageCollectionID,
      collection_ids: sharedPayload.collectionIDs,
      site_category: sharedPayload.siteCategory
    })
    activeSessionID = initResp.upload_session_id
    sessionId.value = activeSessionID

    for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex += 1) {
      if (cancelRequested.value) {
        return { status: 'cancelled', message: '已取消', videoId: '' }
      }
      const start = chunkIndex * chunkSize
      const end = Math.min(start + chunkSize, targetFile.size)
      const chunk = targetFile.slice(start, end)
      await uploadChunk(activeSessionID, chunkIndex, chunk, abortController.value.signal)
      progress.value = Math.round(((chunkIndex + 1) / totalChunks) * 100)
    }

    const completed = await uploadComplete(activeSessionID)
    activeSessionID = ''
    sessionId.value = ''
    return { status: 'success', message: '上传成功', videoId: completed.video_id || '' }
  } catch (err) {
    if (cancelRequested.value || err?.code === 'ERR_CANCELED' || err?.name === 'CanceledError') {
      return { status: 'cancelled', message: '已取消', videoId: '' }
    }
    return { status: 'failed', message: parseErrorMessage(err), videoId: '' }
  } finally {
    const cleanupSessionID = activeSessionID || sessionId.value
    if (cleanupSessionID) {
      try {
        await uploadAbort(cleanupSessionID)
      } catch (_) {
        // 忽略会话清理失败，避免覆盖主流程错误
      }
    }
    sessionId.value = ''
    abortController.value = null
  }
}

async function submit() {
  if (uploading.value) return
  const queue = selectedFiles.value
  if (queue.length === 0) {
    ElMessage.warning('请选择文件')
    return
  }

  uploading.value = true
  cancelRequested.value = false
  totalFiles.value = queue.length
  currentFileName.value = ''
  progress.value = 0
  hashProgress.value = 0
  uploadResults.value = []

  const normalizedTags = buildNormalizedTags()
  const { actorIDs, actorNames } = splitActorSelection(form.actors)
  const sharedPayload = {
    type: form.type,
    title: form.title,
    description: form.description,
    tags: normalizedTags,
    actorIDs,
    actorNames,
    collectionIDs: isShortType.value ? normalizeCollectionSelection(form.collections) : [],
    imageCollectionID: resolveImageCollectionID(form.imageCollectionID),
    siteCategory: isAVType.value ? (form.siteCategory || 'japanese') : ''
  }

  for (let index = 0; index < queue.length; index += 1) {
    const targetFile = queue[index]
    if (cancelRequested.value) {
      for (let pendingIndex = index; pendingIndex < queue.length; pendingIndex += 1) {
        pushResult(queue[pendingIndex], 'cancelled', '未上传（批量任务已取消）')
      }
      break
    }

    const itemResult = await uploadOneFile(targetFile, sharedPayload)
    pushResult(targetFile, itemResult.status, itemResult.message, itemResult.videoId)

    if (itemResult.status === 'cancelled') {
      for (let pendingIndex = index + 1; pendingIndex < queue.length; pendingIndex += 1) {
        pushResult(queue[pendingIndex], 'cancelled', '未上传（批量任务已取消）')
      }
      break
    }
  }

  uploading.value = false
  currentFileName.value = ''
  progress.value = 0
  hashProgress.value = 0

  const success = successCount.value
  const failed = failedCount.value
  const cancelled = cancelledCount.value
  if (cancelRequested.value) {
    ElMessage.warning(`批量上传已取消：成功 ${success}，失败 ${failed}，取消 ${cancelled}`)
    return
  }
  if (failed > 0) {
    ElMessage.warning(`批量上传完成：成功 ${success}，失败 ${failed}`)
    return
  }
  ElMessage.success(`批量上传完成：成功 ${success}`)
}

async function cancelUpload() {
  if (!uploading.value || cancelRequested.value) return
  cancelRequested.value = true
  abortController.value?.abort()
  const currentSessionID = sessionId.value
  if (currentSessionID) {
    try {
      await uploadAbort(currentSessionID)
    } catch (_) {
      // 忽略取消时的会话清理错误
    } finally {
      sessionId.value = ''
    }
  }
  ElMessage.warning('正在取消当前上传，后续文件将停止')
}

function clearUploadRecords() {
  if (uploading.value) return
  uploadResults.value = []
  progress.value = 0
  hashProgress.value = 0
  currentFileName.value = ''
  totalFiles.value = 0
  cancelRequested.value = false
  sessionId.value = ''
}

const loadTagSuggestions = createRemoteSuggestionLoader({
  fetcher: searchTags,
  getOptions: () => form.tags,
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
  getOptions: () => collectionOptions.value.filter((item) => form.collections.includes(item.value)),
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
  getOptions: () => imageCollectionOptions.value.filter((item) => item.value === form.imageCollectionID),
  setOptions: (next) => {
    imageCollectionOptions.value = next
  },
  setLoading: (next) => {
    loadingImageCollections.value = next
  },
  mergeOptions: mergeRemoteValueOptions
})

onMounted(() => {
  searchActors().catch(() => {})
  loadCollectionSuggestions('')
  loadImageCollectionSuggestions('')
  loadTagSuggestions('')
})
</script>

<template>
  <Layout>
    <div class="page-shell upload-page">
      <PageHeader title="上传中心" />

      <SectionCard>
        <template #title>文件与基础信息</template>
        <el-form label-width="100px">
          <el-form-item label="视频文件">
            <el-upload
              ref="uploadRef"
              v-model:file-list="uploadFileList"
              drag
              :auto-upload="false"
              :on-change="onFileChange"
              :multiple="!isMovieType"
              :limit="isMovieType ? 1 : 999"
              class="upload-drop"
            >
              <el-icon><UploadFilled /></el-icon>
              <div>拖拽文件到此，或点击选择文件</div>
            </el-upload>
            <div class="upload-tip">
              <span v-if="isMovieType">电影仅支持单文件上传</span>
              <span v-else>当前类型支持批量上传，可一次选择多个文件</span>
              <span>已选择 {{ selectedFiles.length }} 个文件</span>
            </div>
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="form.type" style="width: 220px">
              <el-option label="短视频" value="short" />
              <el-option label="电影" value="movie" />
              <el-option label="剧集分集" value="episode" />
              <el-option label="AV" value="av" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="isAVType" label="AV 地区分类">
            <el-radio-group v-model="form.siteCategory">
              <el-radio-button v-for="item in avSiteCategoryOptions" :key="item.value" :label="item.value">
                {{ item.label }}
              </el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
          <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
        </el-form>
      </SectionCard>

      <SectionCard>
        <template #title>关联信息</template>
        <el-form label-width="100px">
          <el-form-item label="视频标签">
            <el-select
              v-model="form.tags"
              multiple
              filterable
              remote
              reserve-keyword
              :remote-method="loadTagSuggestions"
              allow-create
              default-first-option
              clearable
              placeholder="可选择或输入标签"
              style="width: 100%"
              :loading="loadingTags"
            >
              <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
            </el-select>
          </el-form-item>
          <el-form-item label="图片图集">
            <el-select
              v-model="form.imageCollectionID"
              filterable
              remote
              reserve-keyword
              clearable
              :remote-method="loadImageCollectionSuggestions"
              :loading="loadingImageCollections"
              placeholder="可选，仅可关联一个图片图集"
              style="width: 100%"
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
              v-model="form.collections"
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
              style="width: 100%"
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
              v-model="form.actors"
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
              style="width: 100%"
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
      </SectionCard>

      <SectionCard>
        <template #title>上传控制</template>
        <div class="upload-actions">
          <el-button type="primary" :loading="uploading" :disabled="selectedFiles.length === 0" @click="submit">开始上传</el-button>
          <el-button v-if="uploading" type="danger" @click="cancelUpload">取消上传</el-button>
          <el-button :disabled="!canClearSelectedFiles" @click="clearSelectedFiles">清空已选文件</el-button>
          <el-button :disabled="!canClearRecords" @click="clearUploadRecords">清空上传记录</el-button>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>进度区</template>
        <UploadProgress
          :percentage="progress"
          :status-text="`当前文件：${currentFileName || '-'} ｜ 哈希计算 ${hashProgress}%`"
        />
        <div class="batch-progress">
          <div class="batch-summary">
            批次进度：{{ completedCount }}/{{ totalFiles || selectedFiles.length }}，成功 {{ successCount }}，失败 {{ failedCount }}，取消 {{ cancelledCount }}
          </div>
          <el-progress :percentage="batchProgress" />
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>结果区</template>
        <div v-if="uploadResults.length" class="table-wrap upload-result">
          <el-table :data="uploadResults" size="small" border>
            <el-table-column prop="name" label="文件名" min-width="280" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">
                  {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : '已取消' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="信息" min-width="180" show-overflow-tooltip />
            <el-table-column prop="videoId" label="视频ID" min-width="260" show-overflow-tooltip />
          </el-table>
        </div>
        <div v-else class="upload-result-empty">暂无上传记录，开始上传后将在这里展示结果。</div>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.upload-page {
  padding-bottom: var(--space-1);
}

.upload-drop :deep(.el-upload-dragger) {
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.upload-tip {
  margin-top: var(--space-3);
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  width: 100%;
  color: var(--text-muted);
  font-size: var(--text-small);
}

.upload-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.batch-progress {
  display: grid;
  gap: var(--space-2);
}

.batch-summary {
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.upload-result-empty {
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--line-strong);
  color: var(--text-muted);
  font-size: var(--text-small);
  background: var(--bg-surface-muted);
}

@media (max-width: 768px) {
  .upload-tip {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
  }
}
</style>
