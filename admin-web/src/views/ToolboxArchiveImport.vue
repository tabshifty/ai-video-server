<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Check, CircleCheck, Close, Delete, FolderOpened, RefreshRight, UploadFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import BulkActionBar from '../components/base/BulkActionBar.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import {
  createAdminCollection,
  createAdminImageCollection,
  deleteAdminArchiveImportBatch,
  getAdminArchiveImportBatches,
  getAdminArchiveImportBatchDetail,
  getAdminArchiveImportFileDetail,
  getAdminCollections,
  getAdminImageCollections,
  getAdminPopularVideoTags,
  getAdminVideoTags,
  processAdminArchiveImportBatch,
  processAdminArchiveImportFile,
  retryAdminArchiveImportExtract,
  updateAdminArchiveImportFile,
  uploadAdminArchiveImport
} from '../api/admin'
import { createRemoteSuggestionLoader, mergeRemoteStringOptions, mergeRemoteValueOptions } from './videoUpload.remote'

const router = useRouter()
const loading = ref(false)
const uploadLoading = ref(false)
const processingBatch = ref(false)
const deletingBatchID = ref('')
const processingFileID = ref('')
const batchList = ref([])
const selectedBatch = ref(null)
const selectedBatchFiles = ref([])
const selectedFileIDs = ref([])
const selectedFileIndexAnchor = ref(-1)
const selectionSyncing = ref(false)
const selectedFile = ref(null)
const fileDetailLoading = ref(false)
const fileSaving = ref(false)
const uploadRef = ref(null)
const uploadFiles = ref([])
const tagOptions = ref([])
const loadingTags = ref(false)
const collectionOptions = ref([])
const loadingCollections = ref(false)
const imageCollectionOptions = ref([])
const loadingImageCollections = ref(false)
const quickCollectionDialogVisible = ref(false)
const quickCollectionSaving = ref(false)
const fileSortMode = ref('original')

const fileSortOptions = [
  { label: '按原始顺序', value: 'original' },
  { label: '按类型排序', value: 'type' }
]

const query = reactive({
  page: 1,
  page_size: 20
})

const uploadForm = reactive({
  title: '',
  default_description: '',
  default_tags: [],
  default_video_collection_ids: [],
  default_image_collection_ids: [],
  has_password: false,
  password: ''
})

const quickCollectionForm = reactive({
  kind: 'video',
  target: '',
  name: '',
  description: ''
})

const batches = computed(() => batchList.value || [])
const quickCollectionDialogTitle = computed(() => (quickCollectionForm.kind === 'image' ? '新建图片合集' : '新建视频合集'))
const displayedBatchFiles = computed(() => sortArchiveFiles(selectedBatchFiles.value))
const selectedBatchFileCount = computed(() => selectedFileIDs.value.length)
const selectedBatchFilesForProcess = computed(() => {
  const selectedSet = new Set(selectedFileIDs.value)
  return displayedBatchFiles.value.filter((file) => selectedSet.has(file.id))
})
const canBatchProcessSelectedFiles = computed(() => selectedBatchFilesForProcess.value.length > 0 && !processingBatch.value)
const bulkActions = computed(() => [
  {
    label: '处理所选',
    icon: CircleCheck,
    type: 'primary',
    loading: processingBatch.value,
    disabled: !canBatchProcessSelectedFiles.value,
    onClick: processSelectedArchiveFiles
  },
  {
    label: '全不选',
    icon: Close,
    type: 'default',
    disabled: selectedFileIDs.value.length === 0 || processingBatch.value,
    onClick: clearArchiveFileSelection
  }
])

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') return responseMsg.trim()
  const message = error?.message
  if (typeof message === 'string' && message.trim() !== '') return message.trim()
  return fallback
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function archiveMediaKindLabel(kind) {
  const map = {
    video: '视频',
    image: '图片',
    archive: '压缩包',
    directory: '目录',
    other: '不支持'
  }
  return map[kind] || kind || '未知'
}

function formatArchiveFormat(file) {
  const mime = String(file?.mime_type || '').trim()
  if (mime) {
    if (mime.startsWith('image/')) return mime.slice('image/'.length).toLowerCase()
    if (mime.startsWith('video/')) return mime.slice('video/'.length).toLowerCase()
    return mime
  }
  const relativePath = String(file?.relative_path || '').trim()
  const matched = relativePath.match(/\.([^.\/\\]+)$/)
  return matched ? matched[1].toLowerCase() : '-'
}

function formatArchiveFileType(file) {
  return `${archiveMediaKindLabel(file?.media_kind)} · ${formatArchiveFormat(file)}`
}

function formatArchiveReason(reason) {
  const value = String(reason || '').trim()
  const map = {
    directory: '目录项不入库',
    nested_archive_not_allowed: '嵌套压缩包已跳过',
    unsupported_file_type: '不支持的文件类型',
    empty_path: '路径为空'
  }
  return map[value] || value
}

function normalizeSelectionValues(values) {
  return Array.from(
    new Set(
      (Array.isArray(values) ? values : String(values || '').split(/[,，\n]/))
        .map((value) => String(value || '').trim())
        .filter(Boolean)
    )
  )
}

function normalizeUUIDSelection(values) {
  return normalizeSelectionValues(values).filter((value) =>
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  )
}

function normalizeTagSelection(values) {
  const out = []
  const seen = new Set()
  for (const item of normalizeSelectionValues(values)) {
    const value = String(item || '').trim().replace(/\s+/g, ' ').toLowerCase()
    if (!value) continue
    if (seen.has(value)) continue
    seen.add(value)
    out.push(value)
  }
  return out
}

function normalizeArchiveFileState(data) {
  if (!data || typeof data !== 'object') return data
  return {
    ...data,
    tags: normalizeTagSelection(data.tags),
    video_collection_ids: normalizeUUIDSelection(data.video_collection_ids),
    image_collection_ids: normalizeUUIDSelection(data.image_collection_ids)
  }
}

function buildQuickCollectionPayload() {
  const payload = {
    name: quickCollectionForm.name.trim(),
    description: quickCollectionForm.description.trim(),
    cover_url: '',
    sort_order: 0,
    active: true
  }
  if (quickCollectionForm.kind === 'image') {
    payload.cover_image_id = null
  }
  return payload
}

function pushOption(optionsRef, option) {
  if (!option?.value) return
  const value = String(option.value)
  const existingIndex = optionsRef.value.findIndex((item) => String(item?.value) === value)
  if (existingIndex >= 0) {
    optionsRef.value.splice(existingIndex, 1, option)
    return
  }
  optionsRef.value.unshift(option)
}

function pushSelectedCollectionValue(target, collectionID) {
  if (!collectionID) return
  if (target === 'upload-video') {
    uploadForm.default_video_collection_ids = normalizeUUIDSelection([...uploadForm.default_video_collection_ids, collectionID])
    return
  }
  if (target === 'upload-image') {
    uploadForm.default_image_collection_ids = normalizeUUIDSelection([...uploadForm.default_image_collection_ids, collectionID])
    return
  }
  if (target === 'file-video' && selectedFile.value) {
    selectedFile.value.video_collection_ids = normalizeUUIDSelection([...(selectedFile.value.video_collection_ids || []), collectionID])
    return
  }
  if (target === 'file-image' && selectedFile.value) {
    selectedFile.value.image_collection_ids = normalizeUUIDSelection([...(selectedFile.value.image_collection_ids || []), collectionID])
  }
}

function openQuickCollectionDialog(kind, target) {
  quickCollectionForm.kind = kind
  quickCollectionForm.target = target
  quickCollectionForm.name = ''
  quickCollectionForm.description = ''
  quickCollectionDialogVisible.value = true
}

function openCreateVideoCollection(target) {
  openQuickCollectionDialog('video', target)
}

function openCreateImageCollection(target) {
  openQuickCollectionDialog('image', target)
}

async function saveQuickCollection() {
  if (!quickCollectionForm.name.trim()) {
    ElMessage.warning(`请输入${quickCollectionForm.kind === 'image' ? '图片合集' : '视频合集'}名称`)
    return
  }
  quickCollectionSaving.value = true
  const target = quickCollectionForm.target
  try {
    const payload = buildQuickCollectionPayload()
    const created = quickCollectionForm.kind === 'image'
      ? await createAdminImageCollection(payload)
      : await createAdminCollection(payload)
    const option = {
      value: created.id,
      label: created.name
    }
    if (quickCollectionForm.kind === 'image') {
      pushOption(imageCollectionOptions, option)
    } else {
      pushOption(collectionOptions, option)
    }
    pushSelectedCollectionValue(target, created.id)
    quickCollectionDialogVisible.value = false
    ElMessage.success(`${quickCollectionForm.kind === 'image' ? '图片合集' : '视频合集'}已创建`)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '创建合集失败'))
  } finally {
    quickCollectionSaving.value = false
  }
}

function onUploadChange(file, files) {
  uploadFiles.value = Array.isArray(files) ? files.slice(-1) : []
}

function onUploadRemove(file, files) {
  uploadFiles.value = Array.isArray(files) ? files : []
}

function batchStatusLabel(status) {
  const map = {
    uploaded: '已上传',
    ready: '待处理',
    processing: '处理中',
    partial: '部分完成',
    completed: '已完成',
    needs_password: '待密码',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function batchStatusType(status) {
  if (status === 'completed') return 'success'
  if (status === 'partial' || status === 'needs_password') return 'warning'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'info'
  return 'primary'
}

function fileStatusLabel(status) {
  const map = {
    pending: '待处理',
    processing: '处理中',
    ready: '已入库',
    existing: '已存在',
    skipped: '已跳过',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function fileStatusType(status) {
  if (status === 'ready' || status === 'existing') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'warning'
  if (status === 'skipped') return 'info'
  return 'primary'
}

function sortArchiveFiles(files) {
  const list = Array.isArray(files) ? files.map((file, index) => ({ file, index })) : []
  if (fileSortMode.value !== 'type') {
    return list.map((item) => item.file)
  }

  const kindOrder = {
    video: 0,
    image: 1,
    archive: 2,
    directory: 3,
    other: 4
  }

  return list
    .sort((left, right) => {
      const leftKind = String(left.file?.media_kind || 'other')
      const rightKind = String(right.file?.media_kind || 'other')
      const kindDiff = (kindOrder[leftKind] ?? kindOrder.other) - (kindOrder[rightKind] ?? kindOrder.other)
      if (kindDiff !== 0) return kindDiff
      return left.index - right.index
    })
    .map((item) => item.file)
}

function selectedArchiveFileSet() {
  return new Set(selectedFileIDs.value)
}

function findArchiveFileIndex(fileID) {
  return displayedBatchFiles.value.findIndex((item) => String(item?.id || '') === String(fileID || ''))
}

function syncArchiveFileSelection(nextIDs, anchorIndex = selectedFileIndexAnchor.value) {
  selectionSyncing.value = true
  selectedFileIDs.value = Array.from(
    new Set((nextIDs || []).map((value) => String(value || '').trim()).filter(Boolean))
  )
  selectedFileIndexAnchor.value = anchorIndex
  if (selectedFile.value && !selectedFileIDs.value.includes(selectedFile.value.id)) {
    selectedFile.value = null
  }
  queueMicrotask(() => {
    selectionSyncing.value = false
  })
}

function applyArchiveFileSelection(rows, anchorIndex = selectedFileIndexAnchor.value) {
  const nextIDs = Array.isArray(rows) ? rows.map((item) => String(item?.id || '').trim()).filter(Boolean) : []
  syncArchiveFileSelection(nextIDs, anchorIndex)
}

function onArchiveFileSelectionChange(rows) {
  const nextIDs = Array.isArray(rows) ? rows.map((item) => String(item?.id || '').trim()).filter(Boolean) : []
  if (selectionSyncing.value) {
    selectedFileIDs.value = nextIDs
    return
  }
  selectedFileIDs.value = nextIDs
  if (selectedFileIDs.value.length === 0) {
    selectedFileIndexAnchor.value = -1
  }
}

function onArchiveFileSelectToggle(row, event) {
  if (!row?.id) return
  const currentIndex = findArchiveFileIndex(row.id)
  if (currentIndex < 0) return

  const rowID = String(row.id)
  const selectedSet = selectedArchiveFileSet()
  const isSelected = selectedSet.has(rowID)

  const shouldRangeSelect = !!event?.shiftKey && selectedFileIndexAnchor.value >= 0

  if (shouldRangeSelect) {
    const start = Math.min(selectedFileIndexAnchor.value, currentIndex)
    const end = Math.max(selectedFileIndexAnchor.value, currentIndex)
    for (let index = start; index <= end; index += 1) {
      const targetID = String(displayedBatchFiles.value[index]?.id || '').trim()
      if (!targetID) continue
      if (isSelected) {
        selectedSet.add(targetID)
      } else {
        selectedSet.delete(targetID)
      }
    }
    applyArchiveFileSelection(
      displayedBatchFiles.value.filter((item) => selectedSet.has(String(item?.id || '').trim())),
      selectedFileIndexAnchor.value
    )
    return
  }

  if (isSelected) {
    selectedSet.delete(rowID)
  } else {
    selectedSet.add(rowID)
  }
  applyArchiveFileSelection(
    displayedBatchFiles.value.filter((item) => selectedSet.has(String(item?.id || '').trim())),
    currentIndex
  )
}

function clearArchiveFileSelection() {
  syncArchiveFileSelection([], -1)
}

async function loadBatches() {
  loading.value = true
  try {
    const data = await getAdminArchiveImportBatches({
      page: query.page,
      page_size: query.page_size
    })
    batchList.value = data.items || []
    if (selectedBatch.value) {
      const matched = batchList.value.find((item) => item.id === selectedBatch.value.id)
      if (matched) {
        selectedBatch.value = matched
      }
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载压缩包批次失败'))
  } finally {
    loading.value = false
  }
}

async function loadBatchDetail(batchID) {
  if (!batchID) return
  loading.value = true
  try {
    const data = await getAdminArchiveImportBatchDetail(batchID)
    selectedBatch.value = data
    selectedBatchFiles.value = (data.files || []).map((item) => normalizeArchiveFileState(item))
    selectedFile.value = null
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载批次详情失败'))
  } finally {
    loading.value = false
  }
}

async function uploadArchive() {
  const input = uploadFiles.value[0]?.raw
  if (!(input instanceof File)) {
    ElMessage.warning('请选择一个压缩包文件')
    return
  }

  const formData = new FormData()
  formData.append('file', input)
  if (uploadForm.title.trim()) formData.append('title', uploadForm.title.trim())
  if (uploadForm.default_description.trim()) formData.append('default_description', uploadForm.default_description.trim())
  const defaultTags = normalizeTagSelection(uploadForm.default_tags)
  if (defaultTags.length > 0) {
    formData.append('default_tags', JSON.stringify(defaultTags))
  }
  const defaultVideoCollectionIDs = normalizeUUIDSelection(uploadForm.default_video_collection_ids)
  if (defaultVideoCollectionIDs.length > 0) {
    formData.append('default_video_collection_ids', JSON.stringify(defaultVideoCollectionIDs))
  }
  const defaultImageCollectionIDs = normalizeUUIDSelection(uploadForm.default_image_collection_ids)
  if (defaultImageCollectionIDs.length > 0) {
    formData.append('default_image_collection_ids', JSON.stringify(defaultImageCollectionIDs))
  }
  formData.append('has_password', uploadForm.has_password ? '1' : '0')
  if (uploadForm.has_password && uploadForm.password.trim()) {
    formData.append('password', uploadForm.password.trim())
  }

  uploadLoading.value = true
  try {
    const batch = await uploadAdminArchiveImport(formData)
    ElMessage.success('压缩包已上传')
    await loadBatches()
    await loadBatchDetail(batch.id)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '上传压缩包失败'))
  } finally {
    uploadLoading.value = false
  }
}

async function refreshBatchDetail() {
  if (!selectedBatch.value?.id) return
  await loadBatchDetail(selectedBatch.value.id)
}

async function runBatchProcess() {
  if (!selectedBatch.value?.id) return
  processingBatch.value = true
  try {
    const data = await processAdminArchiveImportBatch(selectedBatch.value.id)
    selectedBatchFiles.value = (data.items || []).map((item) => normalizeArchiveFileState(item))
    await refreshBatchDetail()
    ElMessage.success('已处理当前批次的待处理文件')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '处理批次失败'))
  } finally {
    processingBatch.value = false
  }
}

async function runRetryExtract() {
  if (!selectedBatch.value?.id) return
  const password = String(uploadForm.password || '').trim()
  if (!password) {
    ElMessage.warning('请输入密码后再重试')
    return
  }
  processingBatch.value = true
  try {
    await retryAdminArchiveImportExtract(selectedBatch.value.id, { password })
    ElMessage.success('已重新解包')
    await refreshBatchDetail()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '重试解包失败'))
  } finally {
    processingBatch.value = false
  }
}

function canDeleteArchiveBatch(batch) {
  return !!batch?.id && batch.status !== 'processing'
}

async function removeArchiveBatch(batch) {
  if (!batch?.id || !canDeleteArchiveBatch(batch)) return
  try {
    await ElMessageBox.confirm(
      `删除后会清空该批次的压缩包记录、文件清单和解包目录，但不会删除已经入库的视频或图片。确认删除“${batch.title || batch.original_filename}”吗？`,
      '删除批次',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        distinguishCancelAndClose: true
      }
    )
  } catch {
    return
  }

  deletingBatchID.value = batch.id
  try {
    await deleteAdminArchiveImportBatch(batch.id)
    if (selectedBatch.value?.id === batch.id) {
      selectedBatch.value = null
      selectedBatchFiles.value = []
      selectedFile.value = null
      syncArchiveFileSelection([], -1)
    }
    await loadBatches()
    ElMessage.success('批次已删除')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除批次失败'))
  } finally {
    deletingBatchID.value = ''
  }
}

async function selectFile(row) {
  if (!row?.id) return
  fileDetailLoading.value = true
  processingFileID.value = row.id
  try {
    const data = await getAdminArchiveImportFileDetail(row.id)
    selectedFile.value = normalizeArchiveFileState(data)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载文件详情失败'))
  } finally {
    fileDetailLoading.value = false
    processingFileID.value = ''
  }
}

async function saveSelectedFile() {
  if (!selectedFile.value?.id) return
  fileSaving.value = true
  try {
    const payload = {
      title: selectedFile.value.title || '',
      description: selectedFile.value.description || '',
      tags: normalizeTagSelection(selectedFile.value.tags),
      video_type: selectedFile.value.video_type || 'short',
      video_collection_ids: normalizeUUIDSelection(selectedFile.value.video_collection_ids),
      image_collection_ids: normalizeUUIDSelection(selectedFile.value.image_collection_ids)
    }
    const data = await updateAdminArchiveImportFile(selectedFile.value.id, payload)
    selectedFile.value = normalizeArchiveFileState(data)
    await refreshBatchDetail()
    ElMessage.success('已保存文件信息')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存文件失败'))
  } finally {
    fileSaving.value = false
  }
}

async function processSelectedFile() {
  if (!selectedFile.value?.id) return
  processingFileID.value = selectedFile.value.id
  try {
    const data = await processAdminArchiveImportFile(selectedFile.value.id)
    selectedFile.value = normalizeArchiveFileState(data)
    await refreshBatchDetail()
    ElMessage.success('已处理文件')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '处理文件失败'))
  } finally {
    processingFileID.value = ''
  }
}

async function processSelectedArchiveFiles() {
  const files = selectedBatchFilesForProcess.value
  if (files.length === 0) {
    ElMessage.warning('请先勾选要处理的文件')
    return
  }
  processingBatch.value = true
  try {
    let processedCount = 0
    let failedCount = 0
    const remainingSelectedIDs = []
    for (const file of files) {
      try {
        await processAdminArchiveImportFile(file.id)
        processedCount += 1
      } catch (error) {
        failedCount += 1
        remainingSelectedIDs.push(file.id)
        if (file.id === selectedFile.value?.id) {
          ElMessage.error(extractErrorMessage(error, '处理文件失败'))
        }
      }
    }
    await refreshBatchDetail()
    syncArchiveFileSelection(remainingSelectedIDs, selectedFileIndexAnchor.value)
    if (processedCount > 0 && failedCount > 0) {
      ElMessage.warning(`已处理 ${processedCount} 个文件，${failedCount} 个失败`)
      return
    }
    if (processedCount > 0) {
      ElMessage.success(`已处理 ${processedCount} 个文件`)
      return
    }
    ElMessage.error('批量处理失败')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '批量处理失败'))
  } finally {
    processingBatch.value = false
  }
}

function clearUploadForm() {
  uploadForm.title = ''
  uploadForm.default_description = ''
  uploadForm.default_tags = []
  uploadForm.default_video_collection_ids = []
  uploadForm.default_image_collection_ids = []
  uploadForm.has_password = false
  uploadForm.password = ''
  uploadFiles.value = []
  uploadRef.value?.clearFiles?.()
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
  getOptions: () => tagOptions.value,
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
  getOptions: () => collectionOptions.value,
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
  getOptions: () => imageCollectionOptions.value,
  setOptions: (next) => {
    imageCollectionOptions.value = next
  },
  setLoading: (next) => {
    loadingImageCollections.value = next
  },
  mergeOptions: mergeRemoteValueOptions
})

function returnToToolbox() {
  router.push('/toolbox')
}

onMounted(async () => {
  await loadBatches()
  loadTagSuggestions('')
  loadCollectionSuggestions('')
  loadImageCollectionSuggestions('')
})
</script>

<template>
  <main class="tool-workspace archive-import-tool">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader title="压缩包导入" subtitle="管理员上传 zip、rar、7z 压缩包后解包、审核并导入视频与图片。" />

      <div class="archive-workspace">
        <div class="archive-workspace__hero">
          <SectionCard class="archive-hero-card">
            <template #title>上传入口</template>
            <template #description>支持密码包，上传时如果勾选“有密码”，请由管理员填写密码。压缩包内不允许嵌套压缩包。</template>
            <template #actions>
              <el-button :icon="FolderOpened" @click="returnToToolbox">返回工具箱</el-button>
              <el-button :icon="RefreshRight" :loading="loading" @click="loadBatches">刷新批次</el-button>
            </template>

            <div class="archive-upload-grid">
              <div class="archive-upload-grid__file">
                <el-upload
                  ref="uploadRef"
                  v-model:file-list="uploadFiles"
                  drag
                  :auto-upload="false"
                  :limit="1"
                  :show-file-list="true"
                  accept=".zip,.rar,.7z"
                  :on-change="onUploadChange"
                  :on-remove="onUploadRemove"
                >
                  <el-icon class="archive-upload-grid__icon"><UploadFilled /></el-icon>
                  <div class="archive-upload-grid__hint">拖拽或点击选择压缩包文件</div>
                  <div class="archive-upload-grid__subhint">仅支持 zip、rar、7z</div>
                </el-upload>
              </div>

              <el-form label-width="120px" class="archive-upload-form">
                <el-form-item label="批次标题"><el-input v-model="uploadForm.title" placeholder="可不填，默认取文件名" /></el-form-item>
                <el-form-item label="默认说明"><el-input v-model="uploadForm.default_description" type="textarea" :rows="2" /></el-form-item>
                <el-form-item label="默认标签">
                  <el-select
                    v-model="uploadForm.default_tags"
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
                    style="width: 100%"
                  >
                    <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
                  </el-select>
                </el-form-item>
                <el-form-item label="默认视频合集">
                  <div class="collection-picker">
                    <el-select
                      v-model="uploadForm.default_video_collection_ids"
                      class="collection-picker__select"
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
                    <el-button class="collection-picker__button" @click="openCreateVideoCollection('upload-video')">新建视频合集</el-button>
                  </div>
                </el-form-item>
                <el-form-item label="默认图片合集">
                  <div class="collection-picker">
                    <el-select
                      v-model="uploadForm.default_image_collection_ids"
                      class="collection-picker__select"
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
                    <el-button class="collection-picker__button" @click="openCreateImageCollection('upload-image')">新建图片合集</el-button>
                  </div>
                </el-form-item>
                <el-form-item label="有密码">
                  <el-switch v-model="uploadForm.has_password" />
                </el-form-item>
                <el-form-item v-if="uploadForm.has_password" label="压缩包密码">
                  <el-input v-model="uploadForm.password" type="password" show-password placeholder="由管理员填写密码" />
                </el-form-item>
                <el-form-item>
                  <div class="archive-actions">
                    <el-button type="primary" :loading="uploadLoading" @click="uploadArchive">上传压缩包</el-button>
                    <el-button :disabled="uploadLoading" @click="clearUploadForm">清空表单</el-button>
                  </div>
                </el-form-item>
              </el-form>
            </div>
          </SectionCard>

          <SectionCard class="archive-stats-card" dense>
            <template #title>当前批次概览</template>
            <template #description>从最近上传到当前选中批次的处理进度，便于先扫一眼再进入细节。</template>

            <div class="archive-stats-grid">
              <div class="archive-stat">
                <span>批次总数</span>
                <strong class="tabular-num">{{ batches.length }}</strong>
              </div>
              <div class="archive-stat">
                <span>当前已选</span>
                <strong>{{ selectedBatch ? (selectedBatch.title || selectedBatch.original_filename || '已选中') : '未选择' }}</strong>
              </div>
              <div class="archive-stat">
                <span>待处理</span>
                <strong class="tabular-num">{{ selectedBatch ? selectedBatch.processable_entries : 0 }}</strong>
              </div>
              <div class="archive-stat">
                <span>已完成</span>
                <strong class="tabular-num">{{ selectedBatch ? selectedBatch.processed_entries : 0 }}</strong>
              </div>
            </div>
          </SectionCard>
        </div>

        <div class="archive-workspace__grid">
          <SectionCard class="archive-batch-panel">
            <template #title>批次列表</template>
            <template #description>选择一个批次查看文件清单，并按文件级别处理视频、图片和跳过项。</template>

            <EmptyState v-if="!batches.length" title="暂无压缩包批次" description="上传一个压缩包后，批次会出现在这里。" />

            <div v-else class="archive-batch-list">
              <button
                v-for="batch in batches"
                :key="batch.id"
                type="button"
                class="archive-batch-item"
                :class="{ 'is-active': selectedBatch?.id === batch.id }"
                @click="loadBatchDetail(batch.id)"
              >
                <strong>{{ batch.title || batch.original_filename }}</strong>
                <span>{{ batch.original_filename }}</span>
                <span>
                  <el-tag effect="plain" size="small" :type="batchStatusType(batch.status)">{{ batchStatusLabel(batch.status) }}</el-tag>
                </span>
                <div class="archive-batch-item__meta">
                  <span class="tabular-num">{{ batch.processed_entries }}/{{ batch.processable_entries }}</span>
                  <el-button
                    v-if="canDeleteArchiveBatch(batch)"
                    class="archive-batch-item__delete"
                    type="danger"
                    text
                    size="small"
                    :icon="Delete"
                    :loading="deletingBatchID === batch.id"
                    @click.stop="removeArchiveBatch(batch)"
                  >
                    删除
                  </el-button>
                </div>
              </button>
            </div>
          </SectionCard>

          <SectionCard v-if="selectedBatch" class="archive-detail-panel">
            <template #title>批次详情</template>
            <template #description>查看当前批次文件清单，逐条修正标题、标签、合集和视频类型后再处理。</template>
            <template #actions>
              <el-segmented v-model="fileSortMode" class="archive-file-sort" :options="fileSortOptions" />
              <el-button :loading="processingBatch" @click="refreshBatchDetail">刷新详情</el-button>
              <el-button type="primary" :loading="processingBatch" @click="runBatchProcess">处理待处理文件</el-button>
              <el-button v-if="selectedBatch.status === 'needs_password'" :loading="processingBatch" @click="runRetryExtract">重试解包</el-button>
            </template>

            <div class="archive-batch-summary">
              <div><span>标题</span><strong>{{ selectedBatch.title }}</strong></div>
              <div><span>格式</span><strong>{{ selectedBatch.archive_format }}</strong></div>
              <div><span>状态</span><strong>{{ batchStatusLabel(selectedBatch.status) }}</strong></div>
              <div><span>总项</span><strong class="tabular-num">{{ selectedBatch.total_entries }}</strong></div>
              <div><span>可处理</span><strong class="tabular-num">{{ selectedBatch.processable_entries }}</strong></div>
              <div><span>已处理</span><strong class="tabular-num">{{ selectedBatch.processed_entries }}</strong></div>
              <div><span>跳过</span><strong class="tabular-num">{{ selectedBatch.skipped_entries }}</strong></div>
              <div><span>失败</span><strong class="tabular-num">{{ selectedBatch.failed_entries }}</strong></div>
            </div>
            <p v-if="selectedBatch.last_error" class="archive-batch-error">{{ selectedBatch.last_error }}</p>

            <div class="archive-file-layout">
              <div class="archive-file-list">
                <div class="archive-file-list__header">
                  <el-checkbox
                    :model-value="selectedBatchFileCount > 0 && selectedBatchFileCount === displayedBatchFiles.length"
                    :indeterminate="selectedBatchFileCount > 0 && selectedBatchFileCount < displayedBatchFiles.length"
                    @change="(checked) => checked ? applyArchiveFileSelection(displayedBatchFiles) : clearArchiveFileSelection()"
                  >
                    已选 {{ selectedBatchFileCount }} 项
                  </el-checkbox>
                  <span>点击文件行查看详情，点左侧选择位勾选；按住 Shift 可连续选择。</span>
                </div>
                <button
                  v-for="file in displayedBatchFiles"
                  :key="file.id"
                  type="button"
                  class="archive-file-item"
                  :class="{ 'is-active': selectedFile?.id === file.id, 'is-selected': selectedFileIDs.includes(file.id), 'has-reason': file.reason }"
                  @click="selectFile(file)"
                >
                  <span
                    class="archive-file-item__selection"
                    role="button"
                    tabindex="0"
                    :aria-label="selectedFileIDs.includes(file.id) ? '取消选择文件' : '选择文件'"
                    @click.stop="onArchiveFileSelectToggle(file, $event)"
                    @keydown.enter.stop.prevent="onArchiveFileSelectToggle(file, $event)"
                    @keydown.space.stop.prevent="onArchiveFileSelectToggle(file, $event)"
                  >
                    <el-icon v-if="selectedFileIDs.includes(file.id)"><Check /></el-icon>
                  </span>
                  <strong>{{ file.relative_path }}</strong>
                  <span>{{ formatArchiveFileType(file) }} · {{ formatFileSize(file.file_size) }}</span>
                  <span v-if="file.reason" class="archive-file-item__reason">{{ formatArchiveReason(file.reason) }}</span>
                  <span><el-tag size="small" effect="plain" :type="fileStatusType(file.status)">{{ fileStatusLabel(file.status) }}</el-tag></span>
                </button>
              </div>

              <SectionCard dense>
                <template #title>文件详情</template>
                <template #description>选中一个文件后可以修改标题、说明、标签和合集，再单独处理。</template>

                <EmptyState
                  v-if="!selectedFile"
                  title="请选择一个文件"
                  description="左侧列表中选择文件后，这里会显示可编辑内容。"
                />

                <div v-else class="archive-file-detail">
                  <el-form label-width="104px">
                    <el-form-item label="相对路径"><el-input :model-value="selectedFile.relative_path" disabled /></el-form-item>
                    <el-form-item label="类型"><el-input :model-value="formatArchiveFileType(selectedFile)" disabled /></el-form-item>
                    <el-form-item v-if="selectedFile.reason" label="跳过原因">
                      <el-input :model-value="formatArchiveReason(selectedFile.reason)" disabled />
                    </el-form-item>
                    <el-form-item label="视频类型" v-if="selectedFile.media_kind === 'video'">
                      <el-select v-model="selectedFile.video_type" style="width: 100%">
                        <el-option label="短视频" value="short" />
                        <el-option label="电影" value="movie" />
                        <el-option label="剧集分集" value="episode" />
                        <el-option label="AV" value="av" />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="标题"><el-input v-model="selectedFile.title" /></el-form-item>
                    <el-form-item label="说明"><el-input v-model="selectedFile.description" type="textarea" :rows="3" /></el-form-item>
                    <el-form-item label="标签" v-if="selectedFile.media_kind === 'video'">
                      <el-select
                        v-model="selectedFile.tags"
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
                        style="width: 100%"
                      >
                        <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="视频合集" v-if="selectedFile.media_kind === 'video'">
                      <div class="collection-picker">
                        <el-select
                          v-model="selectedFile.video_collection_ids"
                          class="collection-picker__select"
                          multiple
                          filterable
                          remote
                          reserve-keyword
                          default-first-option
                          clearable
                          collapse-tags
                          collapse-tags-tooltip
                          placeholder="可选，可多选"
                          :remote-method="loadCollectionSuggestions"
                          :loading="loadingCollections"
                        >
                          <el-option
                            v-for="collection in collectionOptions"
                            :key="collection.value"
                            :label="collection.label"
                            :value="collection.value"
                          />
                        </el-select>
                        <el-button class="collection-picker__button" @click="openCreateVideoCollection('file-video')">新建视频合集</el-button>
                      </div>
                    </el-form-item>
                    <el-form-item label="图片合集" v-if="selectedFile.media_kind === 'image'">
                      <div class="collection-picker">
                        <el-select
                          v-model="selectedFile.image_collection_ids"
                          class="collection-picker__select"
                          multiple
                          filterable
                          remote
                          reserve-keyword
                          default-first-option
                          clearable
                          collapse-tags
                          collapse-tags-tooltip
                          placeholder="可选，可多选"
                          :remote-method="loadImageCollectionSuggestions"
                          :loading="loadingImageCollections"
                        >
                          <el-option
                            v-for="collection in imageCollectionOptions"
                            :key="collection.value"
                            :label="collection.label"
                            :value="collection.value"
                          />
                        </el-select>
                        <el-button class="collection-picker__button" @click="openCreateImageCollection('file-image')">新建图片合集</el-button>
                      </div>
                    </el-form-item>
                    <el-form-item>
                      <div class="archive-actions">
                        <el-button type="primary" :loading="fileSaving" @click="saveSelectedFile">保存文件</el-button>
                        <el-button :loading="fileDetailLoading || processingFileID === selectedFile.id" @click="processSelectedFile">处理文件</el-button>
                      </div>
                    </el-form-item>
                  </el-form>
                </div>
              </SectionCard>
            </div>

            <BulkActionBar :count="selectedFileIDs.length" :actions="bulkActions" />
          </SectionCard>
        </div>
      </div>

      <el-dialog
        v-model="quickCollectionDialogVisible"
        :title="quickCollectionDialogTitle"
        width="min(94vw, 560px)"
        destroy-on-close
      >
        <el-form label-width="104px" class="quick-collection-form">
          <el-form-item :label="quickCollectionForm.kind === 'image' ? '图片合集名称' : '视频合集名称'">
            <el-input v-model="quickCollectionForm.name" placeholder="请输入合集名称" @keyup.enter="saveQuickCollection" />
          </el-form-item>
          <el-form-item label="合集简介">
            <el-input v-model="quickCollectionForm.description" type="textarea" :rows="3" placeholder="可选，简介用于后台识别" />
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="quickCollectionDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="quickCollectionSaving" @click="saveQuickCollection">创建并选中</el-button>
        </template>
      </el-dialog>
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

.archive-workspace {
  display: grid;
  gap: var(--space-5);
}

.archive-workspace__hero {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) minmax(22rem, 0.7fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-hero-card {
  min-width: 0;
}

.archive-stats-card {
  min-width: 0;
}

.archive-stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.archive-stat {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.archive-stat span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.archive-stat strong {
  min-width: 0;
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.archive-workspace__grid {
  display: grid;
  grid-template-columns: minmax(20rem, 26rem) minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-batch-panel,
.archive-detail-panel {
  min-width: 0;
}

.archive-upload-grid {
  display: grid;
  grid-template-columns: minmax(18rem, 20rem) minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-upload-grid__file :deep(.el-upload) {
  width: 100%;
}

.archive-upload-grid__file :deep(.el-upload-dragger) {
  width: 100%;
  min-height: 14rem;
  border: 1px dashed var(--line-strong);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, var(--bg-surface) 0%, var(--bg-surface-muted) 100%);
  box-shadow: var(--shadow-xs);
}

.archive-upload-grid__icon {
  margin-top: var(--space-2);
  font-size: 2.25rem;
  color: var(--primary);
}

.archive-upload-grid__hint {
  margin-top: var(--space-3);
  color: var(--text-primary);
  font-size: var(--text-body);
  font-weight: 600;
}

.archive-upload-grid__subhint {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.archive-upload-form {
  min-width: 0;
  padding-top: var(--space-1);
}

.archive-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.archive-file-sort {
  min-width: 13rem;
}

.archive-file-list__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-1);
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.collection-picker {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-2);
  width: 100%;
}

.collection-picker__select {
  min-width: 0;
  width: 100%;
}

.collection-picker__button {
  white-space: nowrap;
}

.archive-batch-list {
  display: grid;
  gap: var(--space-2);
}

.archive-batch-item,
.archive-file-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  gap: var(--space-3);
  align-items: center;
  width: 100%;
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
  color: var(--text-primary);
  text-align: left;
  cursor: pointer;
  transition: border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.archive-batch-item.is-active,
.archive-file-item.is-active,
.archive-file-item.is-selected {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.archive-batch-item:hover,
.archive-batch-item:focus-visible,
.archive-file-item:hover,
.archive-file-item:focus-visible {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.archive-batch-item:focus-visible,
.archive-file-item:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 2px;
}

.archive-batch-item strong,
.archive-file-item strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-batch-item__meta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
  min-width: 0;
}

.archive-batch-item__delete {
  flex: 0 0 auto;
}

.archive-file-item__selection {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  color: var(--primary);
}

.archive-batch-item span,
.archive-file-item span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.archive-file-item.has-reason {
  grid-template-columns: auto minmax(0, 1fr) auto minmax(0, 9rem) auto;
}

.archive-file-item__reason {
  color: var(--text-muted);
}

.archive-batch-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.archive-batch-summary div {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.archive-batch-summary span {
  color: var(--text-muted);
  font-size: var(--text-caption);
}

.archive-batch-summary strong,
.archive-file-detail strong {
  min-width: 0;
}

.archive-batch-error {
  margin: 0 0 var(--space-3);
  color: var(--danger);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  padding: var(--space-3) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--danger) 20%, var(--line-soft));
  border-radius: var(--radius-md);
  background: var(--danger-50);
}

.archive-file-layout {
  display: grid;
  grid-template-columns: minmax(18rem, 24rem) minmax(0, 1fr);
  gap: var(--space-4);
}

.archive-file-list {
  display: grid;
  gap: var(--space-2);
  align-content: start;
}

.archive-file-detail {
  min-width: 0;
  padding-top: var(--space-1);
}

@media (max-width: 64rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .archive-workspace__hero,
  .archive-workspace__grid,
  .archive-upload-grid,
  .archive-file-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 80rem) {
  .archive-workspace__hero,
  .archive-workspace__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 48rem) {
  .archive-workspace__hero,
  .archive-workspace__grid {
    gap: var(--space-3);
  }

  .archive-stats-grid {
    grid-template-columns: 1fr;
  }

  .archive-batch-item,
  .archive-file-item {
    grid-template-columns: auto minmax(0, 1fr) auto;
    row-gap: var(--space-2);
  }

  .collection-picker {
    grid-template-columns: 1fr;
  }

  .archive-batch-item span:nth-child(2),
  .archive-file-item span:nth-child(2) {
    grid-column: 1;
  }

  .archive-batch-item span:nth-child(3),
  .archive-file-item span:nth-child(3) {
    justify-self: end;
  }

  .archive-file-item__selection {
    grid-column: 1;
  }

  .archive-file-item strong {
    grid-column: 2;
  }

  .archive-batch-item__meta,
  .archive-file-item span:last-child {
    grid-column: 3;
    justify-self: end;
  }

  .archive-file-item.has-reason .archive-file-item__reason {
    grid-column: 2 / -1;
    justify-self: start;
  }

  .archive-upload-grid__file :deep(.el-upload-dragger) {
    min-height: 12rem;
  }
}
</style>
