<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Check, CircleCheck, Close, Delete, EditPen, RefreshRight, UploadFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import BulkActionBar from '../components/base/BulkActionBar.vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import {
  createAdminCollection,
  createAdminImageCollection,
  deleteAdminArchiveImportBatch,
  getAdminArchiveImportBatches,
  getAdminArchiveImportBatchDetail,
  getAdminCollections,
  getAdminImageCollections,
  getAdminPopularVideoTags,
  getAdminVideoTags,
  processAdminArchiveImportFile,
  retryAdminArchiveImportExtract,
  updateAdminArchiveImportFile,
  uploadAdminArchiveImport
} from '../api/admin'
import { createRemoteSuggestionLoader, mergeRemoteStringOptions, mergeRemoteValueOptions } from './videoUpload.remote'

const PROCESSABLE_ARCHIVE_FILE_STATUSES = new Set(['pending', 'failed', 'processing'])

const router = useRouter()
const loading = ref(false)
const batchDetailLoading = ref(false)
const uploadLoading = ref(false)
const processingBatch = ref(false)
const deletingBatchID = ref('')
const fileSaving = ref(false)
const batchEditSaving = ref(false)
const batchDrawerVisible = ref(false)
const uploadDialogVisible = ref(false)
const batchEditDialogVisible = ref(false)
const quickCollectionDialogVisible = ref(false)
const quickCollectionSaving = ref(false)
const uploadRef = ref(null)
const uploadFiles = ref([])
const batchList = ref([])
const selectedBatch = ref(null)
const selectedBatchFiles = ref([])
const selectedFileIDs = ref([])
const selectedFileIndexAnchor = ref(-1)
const selectedFile = ref(null)
const selectedFileSnapshot = ref('')
const retryExtractPassword = ref('')
const fileSortMode = ref('original')
const viewportWidth = ref(readViewportWidth())
const tagOptions = ref([])
const loadingTags = ref(false)
const collectionOptions = ref([])
const loadingCollections = ref(false)
const imageCollectionOptions = ref([])
const loadingImageCollections = ref(false)
const batchEditSnapshot = ref('')

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

const batchEditForm = reactive({
  description_enabled: false,
  description: '',
  tags_enabled: false,
  tags: [],
  video_type_enabled: false,
  video_type: 'short',
  video_collection_enabled: false,
  video_collection_ids: [],
  video_image_collection_enabled: false,
  video_image_collection_id: '',
  image_collection_enabled: false,
  image_collection_ids: []
})

const batches = computed(() => batchList.value || [])
const quickCollectionDialogTitle = computed(() => (quickCollectionForm.kind === 'image' ? '新建图片合集' : '新建视频合集'))
const displayedBatchFiles = computed(() => sortArchiveFiles(selectedBatchFiles.value))
const selectedBatchFileCount = computed(() => selectedFileIDs.value.length)
const selectedBatchFilesForActions = computed(() => {
  const selectedSet = new Set(selectedFileIDs.value)
  return displayedBatchFiles.value.filter((file) => selectedSet.has(String(file.id || '')))
})
const selectedMediaKinds = computed(() =>
  Array.from(new Set(selectedBatchFilesForActions.value.map((file) => archiveFileProcessKind(file)).filter(Boolean)))
)
const selectedMediaKind = computed(() => (selectedMediaKinds.value.length === 1 ? selectedMediaKinds.value[0] : ''))
const hasMixedSelectedMediaKinds = computed(() => selectedMediaKinds.value.length > 1)
const selectedContainsUnsupportedKinds = computed(() =>
  selectedBatchFilesForActions.value.some((file) => archiveFileProcessKind(file) === '')
)
const selectedProcessableFiles = computed(() => selectedBatchFilesForActions.value.filter((file) => canProcessArchiveFile(file)))
const selectedUnprocessableCount = computed(() => selectedBatchFilesForActions.value.length - selectedProcessableFiles.value.length)
const shouldShowSingleFileEditor = computed(() => selectedFileIDs.value.length === 1)
const selectedFileDirty = computed(() => !!selectedFile.value && serializeSelectedFileState() !== selectedFileSnapshot.value)
const canProcessSelectedFiles = computed(() =>
  selectedBatchFilesForActions.value.length > 0
  && !processingBatch.value
  && !hasMixedSelectedMediaKinds.value
  && !selectedContainsUnsupportedKinds.value
  && selectedProcessableFiles.value.length > 0
)
const canOpenBatchEdit = computed(() =>
  selectedBatchFilesForActions.value.length > 1
  && !batchEditSaving.value
  && !hasMixedSelectedMediaKinds.value
  && !selectedContainsUnsupportedKinds.value
  && (selectedMediaKind.value === 'video' || selectedMediaKind.value === 'image')
)
const batchEditTargetLabel = computed(() => (selectedMediaKind.value === 'image' ? '已选图片' : '已选视频'))
const batchEditDialogTitle = computed(() => (selectedMediaKind.value === 'image' ? '批量编辑图片文件' : '批量编辑视频文件'))
const batchEditDirty = computed(() => serializeBatchEditState() !== batchEditSnapshot.value)
const processSelectionActionLabel = computed(() => (selectedFileIDs.value.length === 1 ? '处理当前文件' : '处理所选'))
const currentBatchVideoCount = computed(() => displayedBatchFiles.value.filter((file) => archiveFileProcessKind(file) === 'video').length)
const currentBatchImageCount = computed(() => displayedBatchFiles.value.filter((file) => archiveFileProcessKind(file) === 'image').length)
const batchDrawerSize = computed(() => (viewportWidth.value < 1080 ? '100%' : '1100px'))
const selectedVideoImageCollectionID = computed({
  get() {
    if (!selectedFile.value || selectedFile.value.media_kind !== 'video') return ''
    return String(selectedFile.value.image_collection_ids?.[0] || '')
  },
  set(value) {
    if (!selectedFile.value || selectedFile.value.media_kind !== 'video') return
    selectedFile.value.image_collection_ids = value ? [value] : []
  }
})
const batchEditVideoImageCollectionID = computed({
  get() {
    return String(batchEditForm.video_image_collection_id || '')
  },
  set(value) {
    batchEditForm.video_image_collection_id = String(value || '')
  }
})
const overviewCards = computed(() => {
  const needingAction = batches.value.filter((batch) => batchNeedsAction(batch)).length
  const needPassword = batches.value.filter((batch) => batch.status === 'needs_password').length
  const processing = batches.value.filter((batch) => batch.status === 'processing').length
  return [
    { label: '批次总数', value: batches.value.length, hint: '按上传时间倒序浏览' },
    { label: '待继续处理', value: needingAction, hint: '仍有待处理、失败或待密码批次' },
    { label: '待密码', value: needPassword, hint: '补密码后可继续解包' },
    { label: '处理中', value: processing, hint: '后台仍在解包或入库' }
  ]
})
const selectionAlert = computed(() => {
  if (selectedBatchFilesForActions.value.length === 0) return { type: '', title: '' }
  if (hasMixedSelectedMediaKinds.value) {
    return { type: 'warning', title: '当前同时选中了视频和图片，请先按媒体类型拆开处理。' }
  }
  if (selectedContainsUnsupportedKinds.value) {
    return { type: 'warning', title: '当前选择包含非视频/图片项，无法进入批量编辑或批量处理。' }
  }
  if (selectedUnprocessableCount.value > 0) {
    return { type: 'info', title: `当前有 ${selectedUnprocessableCount.value} 项不是可处理状态，批量处理时会自动跳过。` }
  }
  return { type: '', title: '' }
})
const bulkActions = computed(() => [
  {
    label: '批量编辑',
    icon: EditPen,
    type: 'primary',
    disabled: !canOpenBatchEdit.value,
    onClick: openBatchEditDialog
  },
  {
    label: processSelectionActionLabel.value,
    icon: CircleCheck,
    type: 'primary',
    loading: processingBatch.value,
    disabled: !canProcessSelectedFiles.value,
    onClick: processSelectedArchiveFiles
  },
  {
    label: '清空选择',
    icon: Close,
    type: 'default',
    disabled: selectedFileIDs.value.length === 0 || processingBatch.value,
    onClick: clearArchiveFileSelection
  }
])

function readViewportWidth() {
  if (typeof window === 'undefined') return 1440
  return window.innerWidth
}

function updateViewportWidth() {
  viewportWidth.value = readViewportWidth()
}

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') return responseMsg.trim()
  const message = error?.message
  if (typeof message === 'string' && message.trim() !== '') return message.trim()
  return fallback
}

function formatDateTime(value, fallback = '--') {
  return formatAdminDateTime(value, fallback)
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

function archiveFileProcessKind(file) {
  const kind = String(file?.media_kind || '').trim()
  if (kind === 'video' || kind === 'image') return kind
  return ''
}

function canProcessArchiveFile(file) {
  return archiveFileProcessKind(file) !== '' && PROCESSABLE_ARCHIVE_FILE_STATUSES.has(String(file?.status || '').trim())
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

function createSelectedFileDraft(file) {
  return normalizeArchiveFileState({
    ...file,
    tags: [...(file?.tags || [])],
    video_collection_ids: [...(file?.video_collection_ids || [])],
    image_collection_ids: [...(file?.image_collection_ids || [])]
  })
}

function formatBatchProgress(batch) {
  const processed = Number(batch?.processed_entries || 0)
  const total = Number(batch?.processable_entries || 0)
  if (total <= 0) return '暂无可处理项'
  return `已处理 ${processed}/${total}`
}

function batchNeedsAction(batch) {
  return batch?.status === 'needs_password'
    || batch?.status === 'failed'
    || batch?.status === 'partial'
    || Number(batch?.failed_entries || 0) > 0
    || Number(batch?.processed_entries || 0) < Number(batch?.processable_entries || 0)
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
  return new Set(selectedFileIDs.value.map((value) => String(value || '').trim()).filter(Boolean))
}

function findArchiveFileIndex(fileID) {
  return displayedBatchFiles.value.findIndex((item) => String(item?.id || '') === String(fileID || ''))
}

function findArchiveFileByID(fileID) {
  return selectedBatchFiles.value.find((item) => String(item?.id || '') === String(fileID || '')) || null
}

function serializeSelectedFileState() {
  if (!selectedFile.value) return ''
  return JSON.stringify({
    id: String(selectedFile.value.id || ''),
    title: String(selectedFile.value.title || ''),
    description: String(selectedFile.value.description || ''),
    tags: [...normalizeTagSelection(selectedFile.value.tags)].sort(),
    video_type: String(selectedFile.value.video_type || ''),
    video_collection_ids: [...normalizeUUIDSelection(selectedFile.value.video_collection_ids)].sort(),
    image_collection_ids: [...normalizeUUIDSelection(selectedFile.value.image_collection_ids)].sort()
  })
}

function captureSelectedFileSnapshot() {
  selectedFileSnapshot.value = serializeSelectedFileState()
}

function clearSelectedFileDraft() {
  selectedFile.value = null
  selectedFileSnapshot.value = ''
}

function syncSelectedFileFromCurrentRow() {
  if (selectedFileIDs.value.length !== 1) {
    clearSelectedFileDraft()
    return
  }
  const current = findArchiveFileByID(selectedFileIDs.value[0])
  if (!current) {
    clearSelectedFileDraft()
    return
  }
  selectedFile.value = createSelectedFileDraft(current)
  captureSelectedFileSnapshot()
}

function syncArchiveFileSelection(nextIDs, anchorIndex = selectedFileIndexAnchor.value) {
  selectedFileIDs.value = Array.from(
    new Set((nextIDs || []).map((value) => String(value || '').trim()).filter(Boolean))
  )
  selectedFileIndexAnchor.value = anchorIndex
  if (selectedFileIDs.value.length !== 1) {
    clearSelectedFileDraft()
    return
  }
  syncSelectedFileFromCurrentRow()
}

async function confirmDiscardUnsavedFileChanges(contextTitle = '确认切换') {
  if (!selectedFileDirty.value) return true
  try {
    await ElMessageBox.confirm('未保存的修改将会丢失，确认继续吗？', contextTitle, {
      type: 'warning',
      confirmButtonText: '确认丢弃',
      cancelButtonText: '继续编辑'
    })
    return true
  } catch (_) {
    return false
  }
}

async function requestArchiveFileSelection(nextIDs, anchorIndex = selectedFileIndexAnchor.value) {
  const normalized = Array.from(new Set((nextIDs || []).map((value) => String(value || '').trim()).filter(Boolean)))
  const current = selectedFileIDs.value.map((value) => String(value || '').trim())
  if (normalized.length === current.length && normalized.every((value, index) => value === current[index])) {
    return true
  }
  if (!(await confirmDiscardUnsavedFileChanges('确认切换文件'))) {
    return false
  }
  syncArchiveFileSelection(normalized, anchorIndex)
  return true
}

async function onArchiveFileRowClick(row) {
  if (!row?.id) return
  await requestArchiveFileSelection([row.id], findArchiveFileIndex(row.id))
}

async function onArchiveFileSelectToggle(row, event) {
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
    const shouldSelectRange = !isSelected
    for (let index = start; index <= end; index += 1) {
      const targetID = String(displayedBatchFiles.value[index]?.id || '').trim()
      if (!targetID) continue
      if (shouldSelectRange) {
        selectedSet.add(targetID)
      } else {
        selectedSet.delete(targetID)
      }
    }
    await requestArchiveFileSelection(
      displayedBatchFiles.value.filter((item) => selectedSet.has(String(item?.id || '').trim())).map((item) => item.id),
      selectedFileIndexAnchor.value
    )
    return
  }

  if (isSelected) {
    selectedSet.delete(rowID)
  } else {
    selectedSet.add(rowID)
  }
  await requestArchiveFileSelection(
    displayedBatchFiles.value.filter((item) => selectedSet.has(String(item?.id || '').trim())).map((item) => item.id),
    currentIndex
  )
}

async function clearArchiveFileSelection() {
  await requestArchiveFileSelection([], -1)
}

async function selectArchiveFilesByKind(kind) {
  const nextIDs = displayedBatchFiles.value
    .filter((file) => archiveFileProcessKind(file) === kind)
    .map((file) => file.id)
  await requestArchiveFileSelection(nextIDs, nextIDs.length > 0 ? findArchiveFileIndex(nextIDs[0]) : -1)
}

function applyBatchDetailData(data, options = {}) {
  const files = (data?.files || []).map((item) => normalizeArchiveFileState(item))
  selectedBatch.value = data
  selectedBatchFiles.value = files
  retryExtractPassword.value = ''

  let nextSelection = []
  if (Array.isArray(options.preferredSelectionIDs) && options.preferredSelectionIDs.length > 0) {
    const allowed = new Set(files.map((item) => String(item.id || '')))
    nextSelection = options.preferredSelectionIDs.map((value) => String(value || '')).filter((value) => allowed.has(value))
  } else if (options.preserveSelection === true) {
    const allowed = new Set(files.map((item) => String(item.id || '')))
    nextSelection = selectedFileIDs.value.filter((value) => allowed.has(String(value || '')))
  }

  syncArchiveFileSelection(nextSelection, nextSelection.length === 1 ? findArchiveFileIndex(nextSelection[0]) : -1)
  batchDrawerVisible.value = true
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
      const matched = batchList.value.find((item) => String(item.id || '') === String(selectedBatch.value?.id || ''))
      if (matched) {
        selectedBatch.value = {
          ...selectedBatch.value,
          ...matched
        }
      }
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载压缩包批次失败'))
  } finally {
    loading.value = false
  }
}

async function loadBatchDetail(batchID, options = {}) {
  if (!batchID) return false
  if (!options.skipConfirm && !(await confirmDiscardUnsavedFileChanges('确认切换批次'))) {
    return false
  }
  batchDetailLoading.value = true
  try {
    const data = await getAdminArchiveImportBatchDetail(batchID)
    applyBatchDetailData(data, options)
    return true
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载批次详情失败'))
    return false
  } finally {
    batchDetailLoading.value = false
  }
}

async function openBatchDetail(batchID) {
  if (!batchID) return
  if (batchDrawerVisible.value && String(selectedBatch.value?.id || '') === String(batchID || '')) {
    return
  }
  await loadBatchDetail(batchID, { preserveSelection: false })
}

async function refreshBatchDetail(options = {}) {
  if (!selectedBatch.value?.id) return false
  return loadBatchDetail(selectedBatch.value.id, {
    preserveSelection: true,
    skipConfirm: options.skipConfirm === true
  })
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
    return
  }
  if (target === 'file-video-image' && selectedFile.value) {
    selectedFile.value.image_collection_ids = normalizeUUIDSelection([collectionID]).slice(0, 1)
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

function openUploadDialog() {
  uploadDialogVisible.value = true
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
    clearUploadForm()
    uploadDialogVisible.value = false
    await loadBatches()
    const opened = await loadBatchDetail(batch.id, {
      preserveSelection: false
    })
    if (!opened) {
      ElMessage.info('新批次已创建，可稍后在列表中继续处理')
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '上传压缩包失败'))
  } finally {
    uploadLoading.value = false
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
    if (String(selectedBatch.value?.id || '') === String(batch.id || '')) {
      selectedBatch.value = null
      selectedBatchFiles.value = []
      clearSelectedFileDraft()
      selectedFileIDs.value = []
      selectedFileIndexAnchor.value = -1
      batchDrawerVisible.value = false
    }
    await loadBatches()
    ElMessage.success('批次已删除')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除批次失败'))
  } finally {
    deletingBatchID.value = ''
  }
}

async function saveSelectedFile() {
  if (!selectedFile.value?.id || selectedFileIDs.value.length !== 1) return
  fileSaving.value = true
  try {
    const payload = {
      title: selectedFile.value.title || '',
      description: selectedFile.value.description || '',
      tags: normalizeTagSelection(selectedFile.value.tags),
      video_type: selectedFile.value.video_type || 'short',
      video_collection_ids: normalizeUUIDSelection(selectedFile.value.video_collection_ids),
      image_collection_ids: selectedFile.value.media_kind === 'video'
        ? normalizeUUIDSelection([selectedVideoImageCollectionID.value])
        : normalizeUUIDSelection(selectedFile.value.image_collection_ids)
    }
    await updateAdminArchiveImportFile(selectedFile.value.id, payload)
    await refreshBatchDetail({ skipConfirm: true })
    ElMessage.success('已保存文件信息')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存文件失败'))
  } finally {
    fileSaving.value = false
  }
}

function resetBatchEditForm() {
  batchEditForm.description_enabled = false
  batchEditForm.description = ''
  batchEditForm.tags_enabled = false
  batchEditForm.tags = []
  batchEditForm.video_type_enabled = false
  batchEditForm.video_type = 'short'
  batchEditForm.video_collection_enabled = false
  batchEditForm.video_collection_ids = []
  batchEditForm.video_image_collection_enabled = false
  batchEditForm.video_image_collection_id = ''
  batchEditForm.image_collection_enabled = false
  batchEditForm.image_collection_ids = []
}

function serializeBatchEditState() {
  return JSON.stringify({
    description_enabled: batchEditForm.description_enabled,
    description: String(batchEditForm.description || ''),
    tags_enabled: batchEditForm.tags_enabled,
    tags: [...normalizeTagSelection(batchEditForm.tags)].sort(),
    video_type_enabled: batchEditForm.video_type_enabled,
    video_type: String(batchEditForm.video_type || ''),
    video_collection_enabled: batchEditForm.video_collection_enabled,
    video_collection_ids: [...normalizeUUIDSelection(batchEditForm.video_collection_ids)].sort(),
    video_image_collection_enabled: batchEditForm.video_image_collection_enabled,
    video_image_collection_id: String(batchEditForm.video_image_collection_id || ''),
    image_collection_enabled: batchEditForm.image_collection_enabled,
    image_collection_ids: [...normalizeUUIDSelection(batchEditForm.image_collection_ids)].sort()
  })
}

function captureBatchEditSnapshot() {
  batchEditSnapshot.value = serializeBatchEditState()
}

async function confirmBatchEditClose() {
  if (!batchEditDirty.value) return true
  try {
    await ElMessageBox.confirm('未保存的批量修改将会丢失，确认关闭吗？', '确认关闭', {
      type: 'warning',
      confirmButtonText: '确认丢弃',
      cancelButtonText: '继续编辑'
    })
    return true
  } catch (_) {
    return false
  }
}

function handleBatchEditClosed() {
  resetBatchEditForm()
  captureBatchEditSnapshot()
}

async function requestBatchEditClose() {
  if (await confirmBatchEditClose()) {
    batchEditDialogVisible.value = false
  }
}

function handleBatchEditBeforeClose(done) {
  confirmBatchEditClose().then((confirmed) => {
    if (confirmed) {
      done()
    }
  })
}

async function openBatchEditDialog() {
  if (!canOpenBatchEdit.value) return
  resetBatchEditForm()
  captureBatchEditSnapshot()
  batchEditDialogVisible.value = true
  if (selectedMediaKind.value === 'video') {
    loadTagSuggestions('')
    loadCollectionSuggestions('')
    loadImageCollectionSuggestions('')
  } else if (selectedMediaKind.value === 'image') {
    loadImageCollectionSuggestions('')
  }
}

function buildBatchEditPayloadForFile(file) {
  const current = normalizeArchiveFileState(file)
  const payload = {
    title: current.title || '',
    description: batchEditForm.description_enabled ? batchEditForm.description : (current.description || ''),
    tags: archiveFileProcessKind(file) === 'video'
      ? (batchEditForm.tags_enabled ? normalizeTagSelection(batchEditForm.tags) : normalizeTagSelection(current.tags))
      : [],
    video_type: archiveFileProcessKind(file) === 'video'
      ? (batchEditForm.video_type_enabled ? batchEditForm.video_type || 'short' : current.video_type || 'short')
      : current.video_type || 'short',
    video_collection_ids: archiveFileProcessKind(file) === 'video'
      ? (batchEditForm.video_collection_enabled
        ? normalizeUUIDSelection(batchEditForm.video_collection_ids)
        : normalizeUUIDSelection(current.video_collection_ids))
      : normalizeUUIDSelection(current.video_collection_ids),
    image_collection_ids: archiveFileProcessKind(file) === 'video'
      ? (batchEditForm.video_image_collection_enabled
        ? normalizeUUIDSelection([batchEditForm.video_image_collection_id])
        : normalizeUUIDSelection(current.image_collection_ids).slice(0, 1))
      : (archiveFileProcessKind(file) === 'image'
          ? (batchEditForm.image_collection_enabled
            ? normalizeUUIDSelection(batchEditForm.image_collection_ids)
            : normalizeUUIDSelection(current.image_collection_ids))
          : normalizeUUIDSelection(current.image_collection_ids))
  }
  return payload
}

async function saveBatchEdit() {
  const targets = [...selectedBatchFilesForActions.value]
  if (targets.length <= 1 || batchEditSaving.value || !canOpenBatchEdit.value) return

  let changedFieldCount = 0
  if (batchEditForm.description_enabled) changedFieldCount += 1
  if (selectedMediaKind.value === 'video' && batchEditForm.tags_enabled) changedFieldCount += 1
  if (selectedMediaKind.value === 'video' && batchEditForm.video_type_enabled) changedFieldCount += 1
  if (selectedMediaKind.value === 'video' && batchEditForm.video_collection_enabled) changedFieldCount += 1
  if (selectedMediaKind.value === 'video' && batchEditForm.video_image_collection_enabled) changedFieldCount += 1
  if (selectedMediaKind.value === 'image' && batchEditForm.image_collection_enabled) changedFieldCount += 1

  if (changedFieldCount === 0) {
    ElMessage.warning('请至少勾选一个要批量修改的字段')
    return
  }

  batchEditSaving.value = true
  try {
    let successCount = 0
    let failureCount = 0
    for (const file of targets) {
      try {
        await updateAdminArchiveImportFile(file.id, buildBatchEditPayloadForFile(file))
        successCount += 1
      } catch (_) {
        failureCount += 1
      }
    }
    await refreshBatchDetail({ skipConfirm: true })
    batchEditDialogVisible.value = false
    if (failureCount > 0) {
      ElMessage.warning(`批量修改完成：成功 ${successCount} 项，失败 ${failureCount} 项`)
    } else {
      ElMessage.success(`已批量更新 ${successCount} 项`)
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '批量修改失败'))
  } finally {
    batchEditSaving.value = false
  }
}

async function processSelectedArchiveFiles() {
  const selectedFiles = [...selectedBatchFilesForActions.value]
  if (selectedFiles.length === 0) {
    ElMessage.warning('请先勾选要处理的文件')
    return
  }
  if (hasMixedSelectedMediaKinds.value) {
    ElMessage.warning('当前同时选中了视频和图片，请先按媒体类型拆开处理')
    return
  }
  if (selectedContainsUnsupportedKinds.value) {
    ElMessage.warning('当前选择包含非视频/图片项，无法批量处理')
    return
  }

  const processable = [...selectedProcessableFiles.value]
  if (processable.length === 0) {
    ElMessage.warning('当前选择里没有可处理的文件')
    return
  }

  processingBatch.value = true
  try {
    let processedCount = 0
    let failedCount = 0
    const remainingSelectedIDs = selectedFiles
      .filter((file) => !canProcessArchiveFile(file))
      .map((file) => String(file.id || ''))

    for (const file of processable) {
      try {
        await processAdminArchiveImportFile(file.id)
        processedCount += 1
      } catch (_) {
        failedCount += 1
        remainingSelectedIDs.push(String(file.id || ''))
      }
    }

    await refreshBatchDetail({ skipConfirm: true })
    syncArchiveFileSelection(remainingSelectedIDs, remainingSelectedIDs.length === 1 ? findArchiveFileIndex(remainingSelectedIDs[0]) : -1)

    const skippedCount = selectedFiles.length - processable.length
    if (failedCount > 0 || skippedCount > 0) {
      const parts = [`成功 ${processedCount}`]
      if (failedCount > 0) parts.push(`失败 ${failedCount}`)
      if (skippedCount > 0) parts.push(`跳过 ${skippedCount}`)
      ElMessage.warning(`批量处理完成：${parts.join('，')}`)
      return
    }
    ElMessage.success(`已处理 ${processedCount} 个文件`)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '批量处理失败'))
  } finally {
    processingBatch.value = false
  }
}

async function runRetryExtract() {
  if (!selectedBatch.value?.id) return
  const password = String(retryExtractPassword.value || '').trim()
  if (!password) {
    ElMessage.warning('请输入密码后再重试')
    return
  }
  processingBatch.value = true
  try {
    await retryAdminArchiveImportExtract(selectedBatch.value.id, { password })
    retryExtractPassword.value = ''
    ElMessage.success('已重新解包')
    await refreshBatchDetail({ skipConfirm: true })
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '重试解包失败'))
  } finally {
    processingBatch.value = false
  }
}

async function confirmBatchDrawerClose() {
  return confirmDiscardUnsavedFileChanges('确认关闭')
}

async function requestBatchDrawerClose() {
  if (await confirmBatchDrawerClose()) {
    batchDrawerVisible.value = false
  }
}

function handleBatchDrawerBeforeClose(done) {
  confirmBatchDrawerClose().then((confirmed) => {
    if (confirmed) {
      done()
    }
  })
}

function handleBatchDrawerClosed() {
  selectedFileIDs.value = []
  selectedFileIndexAnchor.value = -1
  clearSelectedFileDraft()
  retryExtractPassword.value = ''
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

async function returnToToolbox() {
  if (!(await confirmDiscardUnsavedFileChanges('确认离开'))) {
    return
  }
  router.push('/toolbox')
}

onMounted(async () => {
  window.addEventListener('resize', updateViewportWidth)
  await loadBatches()
  loadTagSuggestions('')
  loadCollectionSuggestions('')
  loadImageCollectionSuggestions('')
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewportWidth)
})
</script>

<template>
  <main class="tool-workspace archive-import-tool">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader title="压缩包导入" subtitle="先继续处理已有批次；需要新建批次时，再从上传弹窗进入。">
        <template #actions>
          <el-button :icon="RefreshRight" :loading="loading" @click="loadBatches">刷新批次</el-button>
          <el-button type="primary" :icon="UploadFilled" @click="openUploadDialog">上传压缩包</el-button>
        </template>
      </PageHeader>

      <section class="archive-overview-grid">
        <article v-for="item in overviewCards" :key="item.label" class="archive-overview-card">
          <span>{{ item.label }}</span>
          <strong class="tabular-num">{{ item.value }}</strong>
          <p>{{ item.hint }}</p>
        </article>
      </section>

      <SectionCard class="archive-batch-panel">
        <template #title>批次列表</template>
        <template #description>从这里继续处理已有批次。点击卡片后会打开批次详情抽屉；上传新包不再挤占主界面。</template>

        <EmptyState v-if="!batches.length" title="暂无压缩包批次" description="上传一个压缩包后，批次会出现在这里。">
          <template #action>
            <el-button type="primary" :icon="UploadFilled" @click="openUploadDialog">上传压缩包</el-button>
          </template>
        </EmptyState>

        <div v-else class="archive-batch-grid">
          <button
            v-for="batch in batches"
            :key="batch.id"
            type="button"
            class="archive-batch-card"
            :class="{ 'is-active': selectedBatch?.id === batch.id && batchDrawerVisible }"
            @click="openBatchDetail(batch.id)"
          >
            <div class="archive-batch-card__head">
              <div class="archive-batch-card__copy">
                <strong>{{ batch.title || batch.original_filename }}</strong>
                <span>{{ batch.original_filename }}</span>
              </div>
              <el-tag effect="plain" size="small" :type="batchStatusType(batch.status)">{{ batchStatusLabel(batch.status) }}</el-tag>
            </div>

            <div class="archive-batch-card__stats">
              <div>
                <span>处理进度</span>
                <strong class="tabular-num">{{ formatBatchProgress(batch) }}</strong>
              </div>
              <div>
                <span>上传时间</span>
                <strong>{{ formatDateTime(batch.created_at) }}</strong>
              </div>
            </div>

            <div class="archive-batch-card__footer">
              <span>{{ batchNeedsAction(batch) ? '仍有待继续处理项' : '可直接查看批次记录' }}</span>
              <div class="archive-batch-card__actions">
                <el-button size="small" text @click.stop="openBatchDetail(batch.id)">继续处理</el-button>
                <el-button
                  v-if="canDeleteArchiveBatch(batch)"
                  size="small"
                  type="danger"
                  text
                  :icon="Delete"
                  :loading="deletingBatchID === batch.id"
                  @click.stop="removeArchiveBatch(batch)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </button>
        </div>
      </SectionCard>

      <el-drawer
        v-model="batchDrawerVisible"
        title="批次详情"
        direction="rtl"
        :size="batchDrawerSize"
        destroy-on-close
        :before-close="handleBatchDrawerBeforeClose"
        @closed="handleBatchDrawerClosed"
      >
        <div v-if="selectedBatch" class="archive-drawer">
          <section class="archive-drawer__hero">
            <div class="archive-drawer__hero-copy">
              <span class="archive-drawer__eyebrow">当前批次</span>
              <h2>{{ selectedBatch.title || selectedBatch.original_filename }}</h2>
              <p>{{ selectedBatch.original_filename }} · {{ selectedBatch.archive_format }}</p>
            </div>
            <div class="archive-drawer__hero-actions">
              <el-button :loading="batchDetailLoading" @click="refreshBatchDetail">刷新详情</el-button>
              <el-button v-if="selectedBatch.status === 'needs_password'" :loading="processingBatch" @click="runRetryExtract">重试解包</el-button>
            </div>
          </section>

          <section class="archive-batch-summary">
            <div><span>状态</span><strong>{{ batchStatusLabel(selectedBatch.status) }}</strong></div>
            <div><span>总项</span><strong class="tabular-num">{{ selectedBatch.total_entries }}</strong></div>
            <div><span>可处理</span><strong class="tabular-num">{{ selectedBatch.processable_entries }}</strong></div>
            <div><span>已处理</span><strong class="tabular-num">{{ selectedBatch.processed_entries }}</strong></div>
            <div><span>失败</span><strong class="tabular-num">{{ selectedBatch.failed_entries }}</strong></div>
            <div><span>视频项</span><strong class="tabular-num">{{ currentBatchVideoCount }}</strong></div>
            <div><span>图片项</span><strong class="tabular-num">{{ currentBatchImageCount }}</strong></div>
            <div><span>上传时间</span><strong>{{ formatDateTime(selectedBatch.created_at) }}</strong></div>
          </section>

          <el-alert
            v-if="selectedBatch.last_error"
            type="warning"
            :closable="false"
            :title="selectedBatch.last_error"
            class="archive-inline-alert"
          />

          <SectionCard v-if="selectedBatch.status === 'needs_password'" dense class="archive-password-card">
            <template #title>补密码后继续解包</template>
            <template #description>当前批次在等待管理员重新输入密码。输入正确密码后，详情会刷新到最新状态。</template>
            <div class="archive-password-card__form">
              <el-input v-model="retryExtractPassword" type="password" show-password placeholder="请输入压缩包密码" />
              <el-button type="primary" :loading="processingBatch" @click="runRetryExtract">重试解包</el-button>
            </div>
          </SectionCard>

          <SectionCard class="archive-file-panel">
            <template #title>文件清单</template>
            <template #description>优先看路径、状态和类型，再决定是否按类型拆开勾选、批量编辑或批量处理。</template>
            <template #actions>
              <el-segmented v-model="fileSortMode" class="archive-file-sort" :options="fileSortOptions" />
            </template>

            <div class="archive-file-toolbar">
              <div class="archive-file-toolbar__selection">
                <strong>已选 {{ selectedBatchFileCount }} 项</strong>
                <span>点击文件行会切换为单选精修；左侧勾选位支持多选和 Shift 连选。</span>
              </div>
              <div class="archive-file-toolbar__actions">
                <el-button size="small" @click="selectArchiveFilesByKind('video')">全选视频</el-button>
                <el-button size="small" @click="selectArchiveFilesByKind('image')">全选图片</el-button>
                <el-button size="small" :disabled="selectedBatchFileCount === 0" @click="clearArchiveFileSelection">清空选择</el-button>
              </div>
            </div>

            <el-alert
              v-if="selectionAlert.title"
              :type="selectionAlert.type"
              :closable="false"
              :title="selectionAlert.title"
              class="archive-inline-alert"
            />

            <div class="archive-file-list">
              <button
                v-for="file in displayedBatchFiles"
                :key="file.id"
                type="button"
                class="archive-file-item"
                :class="{ 'is-active': selectedFileIDs.length === 1 && selectedFileIDs[0] === String(file.id), 'is-selected': selectedFileIDs.includes(String(file.id)), 'has-reason': file.reason }"
                @click="onArchiveFileRowClick(file)"
              >
                <span
                  class="archive-file-item__selection"
                  :class="{ 'is-selected': selectedFileIDs.includes(String(file.id)) }"
                  role="button"
                  tabindex="0"
                  :aria-label="selectedFileIDs.includes(String(file.id)) ? '取消选择文件' : '选择文件'"
                  @click.stop="onArchiveFileSelectToggle(file, $event)"
                  @keydown.enter.stop.prevent="onArchiveFileSelectToggle(file, $event)"
                  @keydown.space.stop.prevent="onArchiveFileSelectToggle(file, $event)"
                >
                  <el-icon v-if="selectedFileIDs.includes(String(file.id))"><Check /></el-icon>
                </span>

                <div class="archive-file-item__copy">
                  <strong>{{ file.relative_path }}</strong>
                  <span class="archive-file-item__meta">{{ formatArchiveFileType(file) }}</span>
                  <span class="archive-file-item__submeta">{{ formatFileSize(file.file_size) }}</span>
                  <span v-if="file.reason" class="archive-file-item__reason">{{ formatArchiveReason(file.reason) }}</span>
                </div>

                <span class="archive-file-item__status">
                  <el-tag size="small" effect="plain" :type="fileStatusType(file.status)">{{ fileStatusLabel(file.status) }}</el-tag>
                </span>
              </button>
            </div>
          </SectionCard>

          <SectionCard v-if="shouldShowSingleFileEditor" dense class="archive-file-editor">
            <template #title>单文件精修</template>
            <template #description>这里只处理例外项修正；真正的处理动作统一留在下方主动作区。</template>

            <div v-if="!selectedFile" class="archive-file-editor__empty">
              <span>正在准备当前文件信息…</span>
            </div>

            <el-form v-else label-width="96px" class="archive-file-editor__form">
              <el-form-item label="相对路径"><el-input :model-value="selectedFile.relative_path" disabled /></el-form-item>
              <el-form-item label="文件类型"><el-input :model-value="formatArchiveFileType(selectedFile)" disabled /></el-form-item>
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
                  <el-button class="collection-picker__button" @click="openCreateVideoCollection('file-video')">新建视频合集</el-button>
                </div>
              </el-form-item>
              <el-form-item label="图片合集" v-if="selectedFile.media_kind === 'video'">
                <div class="collection-picker">
                  <el-select
                    v-model="selectedVideoImageCollectionID"
                    class="collection-picker__select"
                    filterable
                    remote
                    reserve-keyword
                    clearable
                    default-first-option
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
                  <el-button class="collection-picker__button" @click="openCreateImageCollection('file-video-image')">新建图片合集</el-button>
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
                  <el-button class="collection-picker__button" @click="openCreateImageCollection('file-image')">新建图片合集</el-button>
                </div>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="fileSaving" @click="saveSelectedFile">保存文件</el-button>
              </el-form-item>
            </el-form>
          </SectionCard>

          <SectionCard v-else dense class="archive-selection-card">
            <template #title>当前操作模式</template>
            <template #description>
              <template v-if="selectedFileIDs.length === 0">先从文件清单里勾选一组同类型文件，再走批量编辑或批量处理。</template>
              <template v-else>当前是多选模式，单文件精修已隐藏，主路径请使用下方批量动作。</template>
            </template>
          </SectionCard>

          <BulkActionBar :count="selectedFileIDs.length" :actions="bulkActions" />
        </div>
      </el-drawer>

      <el-dialog
        v-model="uploadDialogVisible"
        title="上传压缩包"
        width="min(94vw, 760px)"
        destroy-on-close
      >
        <div class="archive-upload-dialog">
          <p class="archive-upload-dialog__lead">支持密码包；上传成功后会自动打开新批次详情。压缩包内不允许嵌套压缩包。</p>

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
            <el-icon class="archive-upload-dialog__icon"><UploadFilled /></el-icon>
            <div class="archive-upload-dialog__hint">拖拽或点击选择压缩包文件</div>
            <div class="archive-upload-dialog__subhint">仅支持 zip、rar、7z</div>
          </el-upload>

          <el-form label-width="116px" class="archive-upload-form">
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
          </el-form>
        </div>

        <template #footer>
          <el-button @click="uploadDialogVisible = false">取消</el-button>
          <el-button :disabled="uploadLoading" @click="clearUploadForm">清空表单</el-button>
          <el-button type="primary" :loading="uploadLoading" @click="uploadArchive">上传压缩包</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="batchEditDialogVisible"
        :title="batchEditDialogTitle"
        width="min(94vw, 720px)"
        destroy-on-close
        :before-close="handleBatchEditBeforeClose"
        @closed="handleBatchEditClosed"
      >
        <el-form label-width="112px" class="archive-batch-edit-form">
          <SectionCard dense>
            <template #title>应用范围</template>
            <el-form-item :label="batchEditTargetLabel">
              <el-tag type="info">共 {{ selectedBatchFilesForActions.length }} 项</el-tag>
            </el-form-item>
          </SectionCard>

          <SectionCard dense>
            <template #title>批量字段</template>
            <el-form-item>
              <el-checkbox v-model="batchEditForm.description_enabled">说明</el-checkbox>
              <el-input
                v-model="batchEditForm.description"
                :disabled="!batchEditForm.description_enabled"
                type="textarea"
                :rows="3"
                placeholder="统一覆盖为同一份说明"
              />
            </el-form-item>

            <template v-if="selectedMediaKind === 'video'">
              <el-form-item>
                <el-checkbox v-model="batchEditForm.tags_enabled">标签</el-checkbox>
                <el-select
                  v-model="batchEditForm.tags"
                  :disabled="!batchEditForm.tags_enabled"
                  multiple
                  filterable
                  remote
                  reserve-keyword
                  allow-create
                  default-first-option
                  clearable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="统一覆盖为同一组标签"
                  style="width: 100%"
                  :remote-method="loadTagSuggestions"
                  :loading="loadingTags"
                >
                  <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-checkbox v-model="batchEditForm.video_type_enabled">视频类型</el-checkbox>
                <el-select v-model="batchEditForm.video_type" :disabled="!batchEditForm.video_type_enabled" style="width: 100%">
                  <el-option label="短视频" value="short" />
                  <el-option label="电影" value="movie" />
                  <el-option label="剧集分集" value="episode" />
                  <el-option label="AV" value="av" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-checkbox v-model="batchEditForm.video_collection_enabled">视频合集</el-checkbox>
                <el-select
                  v-model="batchEditForm.video_collection_ids"
                  :disabled="!batchEditForm.video_collection_enabled"
                  multiple
                  filterable
                  remote
                  reserve-keyword
                  clearable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="统一覆盖为同一组视频合集"
                  style="width: 100%"
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
              </el-form-item>
              <el-form-item>
                <el-checkbox v-model="batchEditForm.video_image_collection_enabled">视频关联的图片合集</el-checkbox>
                <el-select
                  v-model="batchEditVideoImageCollectionID"
                  :disabled="!batchEditForm.video_image_collection_enabled"
                  filterable
                  remote
                  reserve-keyword
                  clearable
                  placeholder="统一覆盖为同一个图片图集"
                  style="width: 100%"
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
              </el-form-item>
            </template>

            <template v-else-if="selectedMediaKind === 'image'">
              <el-form-item>
                <el-checkbox v-model="batchEditForm.image_collection_enabled">图片入库后加入的合集</el-checkbox>
                <el-select
                  v-model="batchEditForm.image_collection_ids"
                  :disabled="!batchEditForm.image_collection_enabled"
                  multiple
                  filterable
                  remote
                  reserve-keyword
                  clearable
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="统一覆盖为同一组图片合集"
                  style="width: 100%"
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
              </el-form-item>
            </template>
          </SectionCard>
        </el-form>

        <template #footer>
          <el-button @click="requestBatchEditClose">取消</el-button>
          <el-button type="primary" :loading="batchEditSaving" @click="saveBatchEdit">保存批量修改</el-button>
        </template>
      </el-dialog>

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
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--primary) 10%, transparent) 0, transparent 32%),
    linear-gradient(180deg, var(--bg-canvas) 0%, color-mix(in srgb, var(--bg-canvas) 92%, var(--bg-surface)) 100%);
}

.tool-workspace__inner {
  display: grid;
  width: min(100%, 84rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.archive-overview-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.archive-overview-card {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, var(--bg-surface) 0%, var(--bg-surface-muted) 100%);
  box-shadow: var(--shadow-xs);
}

.archive-overview-card span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.archive-overview-card strong {
  color: var(--text-primary);
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
  font-weight: 600;
}

.archive-overview-card p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-batch-panel {
  min-width: 0;
}

.archive-batch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(19rem, 1fr));
  gap: var(--space-3);
}

.archive-batch-card {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
  text-align: left;
  color: var(--text-primary);
  cursor: pointer;
  transition: border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.archive-batch-card:hover,
.archive-batch-card:focus-visible,
.archive-batch-card.is-active {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.archive-batch-card:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 2px;
}

.archive-batch-card__head,
.archive-batch-card__footer {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.archive-batch-card__copy {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.archive-batch-card__copy strong,
.archive-batch-card__copy span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-batch-card__copy strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.archive-batch-card__copy span,
.archive-batch-card__footer span {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-batch-card__stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.archive-batch-card__stats div {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.archive-batch-card__stats span {
  color: var(--text-muted);
  font-size: var(--text-caption);
}

.archive-batch-card__stats strong {
  min-width: 0;
  color: var(--text-primary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-batch-card__actions {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-1);
}

.archive-drawer {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
}

.archive-drawer__hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  background: linear-gradient(135deg, color-mix(in srgb, var(--primary) 8%, var(--bg-surface)) 0%, var(--bg-surface) 70%);
}

.archive-drawer__hero-copy {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.archive-drawer__hero-copy h2,
.archive-drawer__hero-copy p {
  margin: 0;
}

.archive-drawer__hero-copy h2 {
  color: var(--text-primary);
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
  font-weight: 600;
}

.archive-drawer__hero-copy p {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-drawer__eyebrow {
  color: var(--primary);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  font-weight: 600;
}

.archive-drawer__hero-actions {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.archive-batch-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: var(--space-3);
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

.archive-batch-summary strong {
  min-width: 0;
  color: var(--text-primary);
}

.archive-inline-alert {
  margin: 0;
}

.archive-password-card__form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-2);
  align-items: start;
}

.archive-file-panel,
.archive-file-editor,
.archive-selection-card {
  min-width: 0;
}

.archive-file-sort {
  min-width: 13rem;
}

.archive-file-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.archive-file-toolbar__selection {
  display: grid;
  gap: var(--space-1);
}

.archive-file-toolbar__selection strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.archive-file-toolbar__selection span {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-file-toolbar__actions {
  display: inline-flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.archive-file-list {
  display: grid;
  gap: var(--space-2);
}

.archive-file-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
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
    box-shadow var(--motion-duration-base) var(--motion-easing-standard);
}

.archive-file-item:hover,
.archive-file-item:focus-visible,
.archive-file-item.is-selected,
.archive-file-item.is-active {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.archive-file-item:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 2px;
}

.archive-file-item__selection {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-sm);
  background: var(--bg-surface);
  color: var(--primary);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--bg-surface) 80%, transparent);
  transition: border-color var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard),
    color var(--motion-duration-fast) var(--motion-easing-standard);
}

.archive-file-item__selection:hover,
.archive-file-item__selection:focus-visible,
.archive-file-item__selection.is-selected {
  border-color: var(--primary);
  background: color-mix(in srgb, var(--primary) 10%, var(--bg-surface));
}

.archive-file-item__selection:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

.archive-file-item__copy {
  display: grid;
  min-width: 0;
  gap: 0.125rem;
}

.archive-file-item__copy strong,
.archive-file-item__copy span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-file-item__copy strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.archive-file-item__meta {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-file-item__submeta,
.archive-file-item__reason {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.archive-file-item__status {
  justify-self: end;
}

.archive-file-editor__empty {
  display: flex;
  align-items: center;
  min-height: 5rem;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.archive-file-editor__form {
  min-width: 0;
}

.archive-upload-dialog {
  display: grid;
  gap: var(--space-4);
}

.archive-upload-dialog__lead {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.archive-upload-dialog :deep(.el-upload) {
  width: 100%;
}

.archive-upload-dialog :deep(.el-upload-dragger) {
  width: 100%;
  min-height: 13rem;
  border: 1px dashed var(--line-strong);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, var(--bg-surface) 0%, var(--bg-surface-muted) 100%);
}

.archive-upload-dialog__icon {
  margin-top: var(--space-2);
  font-size: 2.25rem;
  color: var(--primary);
}

.archive-upload-dialog__hint {
  margin-top: var(--space-3);
  color: var(--text-primary);
  font-size: var(--text-body);
  font-weight: 600;
}

.archive-upload-dialog__subhint {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.archive-upload-form,
.archive-batch-edit-form {
  min-width: 0;
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

.tabular-num {
  font-variant-numeric: tabular-nums;
}

@media (max-width: 80rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .archive-overview-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 64rem) {
  .archive-file-toolbar,
  .archive-drawer__hero,
  .archive-password-card__form {
    grid-template-columns: 1fr;
    flex-direction: column;
  }

  .archive-file-toolbar__actions,
  .archive-drawer__hero-actions {
    justify-content: flex-start;
  }

  .archive-password-card__form {
    display: grid;
  }
}

@media (max-width: 48rem) {
  .archive-overview-grid,
  .archive-batch-grid,
  .archive-batch-card__stats,
  .archive-batch-summary {
    grid-template-columns: 1fr;
  }

  .archive-file-item {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .archive-file-item__status {
    grid-column: 2;
    justify-self: start;
  }

  .collection-picker {
    grid-template-columns: 1fr;
  }
}
</style>
