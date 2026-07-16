<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, MoreFilled, Search, SwitchButton, Upload } from '@element-plus/icons-vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import BulkActionBar from '../components/base/BulkActionBar.vue'
import EmptyState from '../components/base/EmptyState.vue'
import MetricStrip from '../components/base/MetricStrip.vue'
import SavedViewTabs from '../components/base/SavedViewTabs.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import Toolbar from '../components/base/Toolbar.vue'
import { useSavedViews } from '../components/base/useSavedViews'
import {
  checkAdminImageUpload,
  deleteAdminImage,
  getAdminActors,
  getAdminImageCollections,
  getAdminImageDetail,
  getAdminImageViewBlob,
  getAdminImages,
  updateAdminImage,
  uploadAdminImages
} from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'
import { sha256File } from '../utils/hash'
import {
  DEFAULT_IMAGE_ACTIVE,
  createImageBuiltInViews,
  hasImageActiveFilters,
  normalizeImageViewSnapshot
} from './imageManage.helpers'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)
const listError = ref('')
const IMAGEMANAGE_VIEW_KEY = 'admin-imagemanage-view'
const SAVED_VIEWS_KEY = 'admin-imagemanage-saved-views-v1'
const tableRef = ref(null)
const selectedImageRows = ref([])
const bulkOperating = ref(false)
const filterDrawerVisible = ref(false)
const viewportWidth = ref(readViewportWidth())
const viewMode = ref(readStoredImageView())
let loadSeq = 0

const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  status: '',
  active: DEFAULT_IMAGE_ACTIVE,
  actor_id: '',
  collection_id: ''
})
const quickSearch = ref(query.q)
const filterDraft = reactive({
  q: query.q,
  status: query.status,
  active: query.active,
  actor_id: query.actor_id,
  collection_id: query.collection_id
})
const builtInViews = createImageBuiltInViews(query.active, viewMode.value)
const {
  availableViews,
  activeViewId,
  editableSourceId,
  selectView,
  saveView,
  updateView,
  renameView,
  removeView
} = useSavedViews({
  storageKey: SAVED_VIEWS_KEY,
  builtInViews,
  normalizeSnapshot: normalizeImageViewSnapshot,
  getCurrentSnapshot: () => ({
    q: query.q,
    status: query.status,
    active: query.active,
    actor_id: query.actor_id,
    collection_id: query.collection_id,
    viewMode: viewMode.value
  }),
  applySnapshot: applyImageViewSnapshot,
  refresh: refreshSavedView
})

const uploadRef = ref(null)
const uploadFileList = ref([])
const uploading = ref(false)
const uploadDialogVisible = ref(false)
const uploadForm = reactive({
  description: '',
  actor_tokens: [],
  collection_ids: []
})
const uploadSummary = ref(null)
const uploadDrawerSnapshot = ref('')

const detailVisible = ref(false)
const detail = ref(null)
const saving = ref(false)
const detailDrawerSnapshot = ref('')

const actorOptions = ref([])
const loadingActors = ref(false)
const imageCollectionOptions = ref([])
const loadingCollections = ref(false)
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

const preview = reactive({
  loading: false,
  url: '',
  error: '',
  w: 0,
  h: 0,
  fit: 'inside',
  q: 82,
  zoom: 100
})

const readyCount = computed(() => list.value.filter((item) => item.status === 'ready').length)
const failedCount = computed(() => list.value.filter((item) => item.status === 'failed').length)
const inactiveCount = computed(() => list.value.filter((item) => !item.active).length)
const allPageSelected = computed(() => list.value.length > 0 && list.value.every((item) => isGridSelected(item)))
const somePageSelected = computed(() => !allPageSelected.value && list.value.some((item) => isGridSelected(item)))
const drawerSize = computed(() => (viewportWidth.value < 1024 ? '100%' : '560px'))
const uploadDrawerDirty = computed(() => uploadDrawerSnapshot.value !== serializeUploadDrawerState())
const detailDrawerDirty = computed(() => detailDrawerSnapshot.value !== serializeDetailDrawerState())
const hasActiveFilters = computed(() => hasImageActiveFilters(query))
const summaryMetrics = computed(() => [
  { key: 'total', label: '结果总数', value: total.value, scope: '当前条件·全部页' },
  { key: 'ready', label: '可用', value: readyCount.value, scope: '本页', tone: 'success' },
  { key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' },
  { key: 'inactive', label: '停用', value: inactiveCount.value, scope: '本页', tone: 'warning' }
])
const activeFilterChips = computed(() => {
  const chips = []
  if (query.q) chips.push({ key: 'q', label: '搜索', value: query.q })
  if (query.status) chips.push({ key: 'status', label: '状态', value: statusLabel(query.status) })
  if (query.active !== DEFAULT_IMAGE_ACTIVE) {
    chips.push({ key: 'active', label: '启用', value: query.active === '0' ? '仅停用' : '全部状态' })
  }
  if (query.actor_id) chips.push({ key: 'actor_id', label: '演员', value: optionLabel(actorOptions.value, query.actor_id) })
  if (query.collection_id) chips.push({ key: 'collection_id', label: '图片合集', value: optionLabel(imageCollectionOptions.value, query.collection_id) })
  return chips
})
const bulkActions = computed(() => [
  { label: '批量启用', icon: SwitchButton, type: 'primary', loading: bulkOperating.value, disabled: bulkOperating.value, onClick: () => bulkToggleActive(true) },
  { label: '批量停用', icon: SwitchButton, type: 'warning', loading: bulkOperating.value, disabled: bulkOperating.value, onClick: () => bulkToggleActive(false) },
  { label: '批量删除', icon: Delete, type: 'danger', loading: bulkOperating.value, disabled: bulkOperating.value, onClick: doBulkDelete }
])

function readViewportWidth() {
  if (typeof window === 'undefined') return 1440
  return window.innerWidth
}

function updateViewportWidth() {
  viewportWidth.value = readViewportWidth()
}

function readStoredImageView() {
  if (typeof window === 'undefined') return 'grid'
  try {
    return window.localStorage.getItem(IMAGEMANAGE_VIEW_KEY) === 'list' ? 'list' : 'grid'
  } catch (_) {
    return 'grid'
  }
}

function setViewMode(mode) {
  viewMode.value = mode === 'list' ? 'list' : 'grid'
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(IMAGEMANAGE_VIEW_KEY, viewMode.value)
  } catch (_) {
    // localStorage 不可用时仅保留本会话视图模式。
  }
}

function optionLabel(options, value) {
  return options.find((item) => item.value === value)?.label || value
}

function serializeUploadDrawerState() {
  return JSON.stringify({
    files: uploadFileList.value.map((item) => item.name || item.raw?.name || ''),
    description: uploadForm.description || '',
    actor_tokens: [...(uploadForm.actor_tokens || [])].sort(),
    collection_ids: [...(uploadForm.collection_ids || [])].sort()
  })
}

function captureUploadDrawerSnapshot() {
  uploadDrawerSnapshot.value = serializeUploadDrawerState()
}

function serializeDetailDrawerState() {
  if (!detail.value) return ''
  return JSON.stringify({
    id: detail.value.id || '',
    title: detail.value.title || '',
    description: detail.value.description || '',
    active: !!detail.value.active,
    actor_tokens: [...(detail.value.actor_tokens || [])].sort(),
    collection_ids: [...(detail.value.collection_ids || [])].sort(),
    metadata_text: detail.value.metadata_text || ''
  })
}

function captureDetailDrawerSnapshot() {
  detailDrawerSnapshot.value = serializeDetailDrawerState()
}

async function confirmDrawerClose(isDirty) {
  if (!isDirty) return true
  try {
    await ElMessageBox.confirm('未保存的修改将会丢失，确认关闭吗？', '确认关闭', {
      type: 'warning',
      confirmButtonText: '确认丢弃',
      cancelButtonText: '继续编辑'
    })
    return true
  } catch (_) {
    return false
  }
}

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

function statusLabel(status) {
  const map = {
    ready: '可用',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
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
    const cleaned = value.replace(/\s+/g, ' ')
    const key = cleaned.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(cleaned)
  }
  return { actorIDs, actorNames }
}

function normalizeCollectionSelection(values) {
  const out = []
  const seen = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value || !uuidPattern.test(value)) continue
    const key = value.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
  }
  return out
}

function mergeActorOptions(actors = []) {
  const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
  for (const actor of actors) {
    if (!actor?.id) continue
    optionMap.set(actor.id, { value: actor.id, label: actor.name || actor.id })
  }
  actorOptions.value = Array.from(optionMap.values())
}

function mergeCollectionOptions(collections = []) {
  const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
  for (const collection of collections) {
    if (!collection?.id) continue
    optionMap.set(collection.id, { value: collection.id, label: collection.name || collection.id })
  }
  imageCollectionOptions.value = Array.from(optionMap.values())
}

function buildListParams() {
  const params = {
    page: query.page,
    page_size: query.page_size,
    q: query.q,
    status: query.status
  }
  if (query.active === DEFAULT_IMAGE_ACTIVE) {
    params.active = 1
  } else if (query.active === '0') {
    params.active = 0
  }
  if (query.actor_id) {
    params.actor_id = query.actor_id
  }
  if (query.collection_id) {
    params.collection_id = query.collection_id
  }
  return params
}

function resetQueryIdentity() {
  list.value = []
  total.value = 0
  clearImageSelection()
  listError.value = ''
}

function syncFilterDraftFromQuery() {
  filterDraft.q = query.q
  filterDraft.status = query.status
  filterDraft.active = query.active
  filterDraft.actor_id = query.actor_id
  filterDraft.collection_id = query.collection_id
}

function openFilterDrawer() {
  syncFilterDraftFromQuery()
  filterDrawerVisible.value = true
}

function applyImageViewSnapshot(snapshot) {
  const next = normalizeImageViewSnapshot(snapshot)
  query.q = next.q
  query.status = next.status
  query.active = next.active
  query.actor_id = next.actor_id
  query.collection_id = next.collection_id
  query.page = 1
  setViewMode(next.viewMode)
  quickSearch.value = query.q
  syncFilterDraftFromQuery()
  resetQueryIdentity()
}

async function refreshSavedView() {
  const loaded = await load()
  if (loaded === false) {
    throw new Error(listError.value || '加载图片列表失败')
  }
  return loaded
}

async function selectImageView(id) {
  try {
    await selectView(id)
  } catch (error) {
    listError.value = extractErrorMessage(error, '加载图片列表失败')
  }
}

async function removeImageView(id) {
  try {
    await removeView(id)
  } catch (error) {
    listError.value = extractErrorMessage(error, '加载图片列表失败')
  }
}

async function load() {
  const seq = ++loadSeq
  loading.value = true
  try {
    const data = await getAdminImages(buildListParams())
    if (seq !== loadSeq) {
      return null
    }
    list.value = data.items || []
    total.value = data.total_count || 0
    listError.value = ''
    clearImageSelection()
    return true
  } catch (error) {
    if (seq !== loadSeq) {
      return null
    }
    listError.value = extractErrorMessage(error, '加载图片列表失败')
    return false
  } finally {
    if (seq === loadSeq) {
      loading.value = false
    }
  }
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
    const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
    for (const item of data.items || []) {
      optionMap.set(item.id, { value: item.id, label: item.name })
    }
    actorOptions.value = Array.from(optionMap.values())
  } finally {
    loadingActors.value = false
  }
}

async function searchImageCollections(keyword = '') {
  loadingCollections.value = true
  try {
    const data = await getAdminImageCollections({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 50
    })
    const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
    for (const item of data.items || []) {
      optionMap.set(item.id, { value: item.id, label: item.name })
    }
    imageCollectionOptions.value = Array.from(optionMap.values())
  } finally {
    loadingCollections.value = false
  }
}

function resetUploadForm() {
  uploadForm.description = ''
  uploadForm.actor_tokens = []
  uploadForm.collection_ids = []
}

function openUploadDialog() {
  uploadDialogVisible.value = true
  uploadSummary.value = null
  captureUploadDrawerSnapshot()
}

function onUploadDialogClosed() {
  if (uploading.value) return
  uploadRef.value?.clearFiles()
  uploadFileList.value = []
  resetUploadForm()
  uploadSummary.value = null
  captureUploadDrawerSnapshot()
}

async function requestUploadDrawerClose() {
  if (uploading.value) return
  if (await confirmDrawerClose(uploadDrawerDirty.value)) {
    captureUploadDrawerSnapshot()
    uploadDialogVisible.value = false
  }
}

function handleUploadDrawerBeforeClose(done) {
  if (uploading.value) return
  confirmDrawerClose(uploadDrawerDirty.value).then((confirmed) => {
    if (confirmed) {
      captureUploadDrawerSnapshot()
      done()
    }
  })
}

function handleUploadExceed() {
  ElMessage.warning('单次最多选择 100 张图片')
}

async function submitUpload() {
  if (uploadFileList.value.length === 0) {
    ElMessage.warning('请先选择图片文件')
    return
  }
  const rawFiles = uploadFileList.value
    .map((item) => item?.raw)
    .filter((item) => item instanceof File)
  if (rawFiles.length === 0) {
    ElMessage.warning('未读取到可上传的图片文件')
    return
  }

  const { actorIDs, actorNames } = splitActorSelection(uploadForm.actor_tokens)
  const collectionIDs = normalizeCollectionSelection(uploadForm.collection_ids)

  uploading.value = true
  try {
    const instantItems = []
    const pendingFiles = []
    for (const file of rawFiles) {
      let hit = false
      try {
        const hash = await sha256File(file)
        const checked = await checkAdminImageUpload({ hash, file_size: file.size })
        if (checked?.exists) {
          instantItems.push({
            filename: file.name,
            success: true,
            image_id: checked.image_id || '',
            already_exists: true,
            status: 'ready',
            message: '秒传命中'
          })
          hit = true
        }
      } catch (_) {
        // 预检失败时回退到正常上传，避免阻断主流程。
      }
      if (!hit) {
        pendingFiles.push(file)
      }
    }

    const buildFormData = () => {
      const formData = new FormData()
      for (const file of pendingFiles) {
        formData.append('files', file)
      }
      if (uploadForm.description?.trim()) {
        formData.append('description', uploadForm.description.trim())
      }
      if (actorIDs.length > 0) {
        formData.append('actor_ids', JSON.stringify(actorIDs))
      }
      if (actorNames.length > 0) {
        formData.append('actor_names', JSON.stringify(actorNames))
      }
      if (collectionIDs.length > 0) {
        formData.append('collection_ids', JSON.stringify(collectionIDs))
      }
      return formData
    }

    let uploadedItems = []
    let uploadSuccessCount = 0
    let uploadFailedCount = 0
    if (pendingFiles.length > 0) {
      try {
        const uploaded = await uploadAdminImages(buildFormData())
        uploadedItems = uploaded.items || []
        uploadSuccessCount = Number(uploaded.success_count || 0)
        uploadFailedCount = Number(uploaded.failed_count || 0)
      } catch (error) {
        const message = extractErrorMessage(error, '上传失败')
        uploadedItems = pendingFiles.map((file) => ({
          filename: file.name,
          success: false,
          error: message
        }))
        uploadSuccessCount = 0
        uploadFailedCount = pendingFiles.length
      }
    }

    const result = {
      items: [...instantItems, ...uploadedItems],
      total_count: rawFiles.length,
      success_count: instantItems.length + uploadSuccessCount,
      failed_count: uploadFailedCount
    }
    uploadSummary.value = result
    const success = Number(result.success_count || 0)
    const failed = Number(result.failed_count || 0)
    ElMessage.success(`上传完成：成功 ${success}，失败 ${failed}`)
    uploadRef.value?.clearFiles()
    uploadFileList.value = []
    captureUploadDrawerSnapshot()
    uploading.value = false
    try {
      await load()
    } catch (error) {
      ElMessage.warning(extractErrorMessage(error, '上传完成，但刷新图片列表失败'))
    }
    return
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '批量上传图片失败'))
  } finally {
    uploading.value = false
  }
}

async function showDetail(row) {
  const data = await getAdminImageDetail(row.id)
  data.actor_tokens = (data.actors || []).map((item) => item.id)
  data.collection_ids = (data.collections || []).map((item) => item.id)
  data.metadata_text = JSON.stringify(data.metadata || {}, null, 2)
  detail.value = data
  preview.w = 0
  preview.h = 0
  preview.fit = 'inside'
  preview.q = 82
  preview.zoom = 100
  mergeActorOptions(data.actors || [])
  mergeCollectionOptions(data.collections || [])
  detailVisible.value = true
  captureDetailDrawerSnapshot()
  await Promise.all([searchActors(''), searchImageCollections('')])
  await loadPreview()
}

async function openDetailFromRouteQuery() {
  const imageID = String(route.query.image_id || '').trim()
  if (!imageID) return
  try {
    await showDetail({ id: imageID })
  } catch (error) {
    ElMessage.warning(extractErrorMessage(error, '来源图片当前不可用，仅保留历史来源摘要'))
  } finally {
    const nextQuery = { ...route.query }
    delete nextQuery.image_id
    router.replace({ path: route.path, query: nextQuery }).catch(() => {})
  }
}

function parseMetadata(text) {
  const raw = String(text || '').trim()
  if (!raw) return {}
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('metadata 必须是 JSON 对象')
  }
  return parsed
}

async function saveDetail() {
  if (!detail.value?.id) return
  if (!detail.value.title || !detail.value.title.trim()) {
    ElMessage.warning('图片标题不能为空')
    return
  }
  let metadata
  try {
    metadata = parseMetadata(detail.value.metadata_text)
  } catch (error) {
    ElMessage.error(error.message || 'metadata JSON 格式错误')
    return
  }

  const { actorIDs, actorNames } = splitActorSelection(detail.value.actor_tokens)
  const payload = {
    title: detail.value.title.trim(),
    description: detail.value.description || '',
    active: !!detail.value.active,
    metadata,
    actor_ids: actorIDs,
    actor_names: actorNames,
    collection_ids: normalizeCollectionSelection(detail.value.collection_ids)
  }

  saving.value = true
  try {
    await updateAdminImage(detail.value.id, payload)
    ElMessage.success('图片信息已更新')
    captureDetailDrawerSnapshot()
    detailVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存图片信息失败'))
  } finally {
    saving.value = false
  }
}

async function requestDetailDrawerClose() {
  if (await confirmDrawerClose(detailDrawerDirty.value)) {
    captureDetailDrawerSnapshot()
    detailVisible.value = false
  }
}

function handleDetailDrawerBeforeClose(done) {
  confirmDrawerClose(detailDrawerDirty.value).then((confirmed) => {
    if (confirmed) {
      captureDetailDrawerSnapshot()
      done()
    }
  })
}

async function toggleActive(row) {
  await updateAdminImage(row.id, { active: !row.active })
  ElMessage.success(!row.active ? '图片已启用' : '图片已停用')
  await load()
}

function onImageSelectionChange(rows) {
  selectedImageRows.value = Array.isArray(rows) ? rows : []
}

function toggleGridSelection(row, checked) {
  const next = selectedImageRows.value.filter((item) => item.id !== row.id)
  if (checked) {
    next.push(row)
  }
  onImageSelectionChange(next)
}

function isGridSelected(row) {
  return selectedImageRows.value.some((item) => item.id === row.id)
}

function toggleCurrentPageSelection(checked) {
  const currentPageIDs = new Set(list.value.map((item) => item.id))
  const next = selectedImageRows.value.filter((item) => !currentPageIDs.has(item.id))
  if (checked) {
    next.push(...list.value)
  }
  onImageSelectionChange(next)
}

function clearImageSelection() {
  selectedImageRows.value = []
  tableRef.value?.clearSelection?.()
}

async function bulkToggleActive(active) {
  const targets = [...selectedImageRows.value]
  if (targets.length === 0 || bulkOperating.value) return
  try {
    await ElMessageBox.confirm(`确认将已选 ${targets.length} 张图片批量${active ? '启用' : '停用'}？`, '批量改状态', {
      type: 'warning'
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    throw error
  }
  bulkOperating.value = true
  try {
    const results = await Promise.allSettled(targets.map((item) => updateAdminImage(item.id, { active })))
    const ok = results.filter((item) => item.status === 'fulfilled').length
    const failed = results.length - ok
    if (failed > 0) {
      ElMessage.warning(`批量${active ? '启用' : '停用'}完成：成功 ${ok}，失败 ${failed}`)
    } else {
      ElMessage.success(`已批量${active ? '启用' : '停用'} ${ok} 张图片`)
    }
    clearImageSelection()
    await load()
  } finally {
    bulkOperating.value = false
  }
}

async function doBulkDelete() {
  const targets = [...selectedImageRows.value]
  if (targets.length === 0 || bulkOperating.value) return
  try {
    await ElMessageBox.confirm(`确认删除已选 ${targets.length} 张图片？`, '批量删除图片', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    throw error
  }
  bulkOperating.value = true
  try {
    const results = await Promise.allSettled(targets.map((item) => deleteAdminImage(item.id)))
    const ok = results.filter((item) => item.status === 'fulfilled').length
    const failed = results.length - ok
    if (failed > 0) {
      ElMessage.warning(`批量删除完成：成功 ${ok}，失败 ${failed}`)
    } else {
      ElMessage.success(`已删除 ${ok} 张图片`)
    }
    clearImageSelection()
    await load()
  } finally {
    bulkOperating.value = false
  }
}

function applyQuickSearch() {
  query.q = quickSearch.value
  query.page = 1
  syncFilterDraftFromQuery()
  resetQueryIdentity()
  load()
}

function applyFilterDrawer() {
  query.q = filterDraft.q
  query.status = filterDraft.status
  query.active = filterDraft.active
  query.actor_id = filterDraft.actor_id
  query.collection_id = filterDraft.collection_id
  query.page = 1
  quickSearch.value = query.q
  filterDrawerVisible.value = false
  resetQueryIdentity()
  load()
}

function removeFilter(key) {
  if (key === 'active') {
    query.active = DEFAULT_IMAGE_ACTIVE
  } else {
    query[key] = ''
  }
  query.page = 1
  quickSearch.value = query.q
  syncFilterDraftFromQuery()
  resetQueryIdentity()
  load()
}

function resetFilters() {
  query.q = ''
  query.status = ''
  query.active = DEFAULT_IMAGE_ACTIVE
  query.actor_id = ''
  query.collection_id = ''
  query.page = 1
  quickSearch.value = query.q
  syncFilterDraftFromQuery()
  resetQueryIdentity()
  load()
}

function setPage(page) {
  query.page = page
  resetQueryIdentity()
  load()
}

async function doDelete(row) {
  try {
    await ElMessageBox.confirm(`确认删除图片「${row.title || row.id}」？`, '删除图片', {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    ElMessage.error('删除确认失败，请重试')
    return
  }

  try {
    await deleteAdminImage(row.id)
    ElMessage.success('图片已删除')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除图片失败'))
  }
}

function handleImageRowAction(command, row) {
  if (command === 'toggle') {
    toggleActive(row).catch((error) => {
      if (error === 'cancel' || error === 'close') return
      ElMessage.error(extractErrorMessage(error, '更新图片状态失败'))
    })
    return true
  }
  if (command === 'delete') {
    doDelete(row).catch((error) => {
      if (error === 'cancel' || error === 'close') return
      ElMessage.error(extractErrorMessage(error, '删除图片失败'))
    })
    return true
  }
  return false
}

function currentPreviewParams() {
  const params = {}
  if (Number(preview.w) > 0) {
    params.w = Number(preview.w)
  }
  if (Number(preview.h) > 0) {
    params.h = Number(preview.h)
  }
  if (preview.fit) {
    params.fit = preview.fit
  }
  if (Number(preview.q) > 0) {
    params.q = Number(preview.q)
  }
  return params
}

async function loadPreview() {
  if (!detail.value?.id) return
  if (preview.url) {
    URL.revokeObjectURL(preview.url)
    preview.url = ''
  }
  preview.error = ''
  preview.loading = true
  try {
    const blob = await getAdminImageViewBlob(detail.value.id, currentPreviewParams())
    if (blob?.type?.includes('application/json')) {
      const text = await blob.text()
      let payload
      try {
        payload = JSON.parse(text)
      } catch (_) {
        payload = null
      }
      throw new Error(payload?.msg || '加载图片预览失败')
    }
    preview.url = URL.createObjectURL(blob)
  } catch (error) {
    preview.error = extractErrorMessage(error, '加载图片预览失败')
  } finally {
    preview.loading = false
  }
}

function clearPreview() {
  if (preview.url) {
    URL.revokeObjectURL(preview.url)
    preview.url = ''
  }
  preview.error = ''
}

function onDetailClosed() {
  clearPreview()
  detail.value = null
  captureDetailDrawerSnapshot()
}

onMounted(async () => {
  window.addEventListener('resize', updateViewportWidth)
  await Promise.all([load(), searchActors(''), searchImageCollections('')])
  await openDetailFromRouteQuery()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateViewportWidth)
  clearPreview()
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <div class="image-header-actions">
        <el-button :icon="Upload" type="primary" @click="openUploadDialog">上传图片</el-button>
      </div>
    </template>

    <div class="page-shell image-page" data-density="compact">
      <SavedViewTabs
        :items="availableViews"
        :active-id="activeViewId"
        :editable-source-id="editableSourceId"
        @select="selectImageView"
        @save="saveView"
        @update="updateView"
        @rename="renameView"
        @remove="removeImageView"
      />

      <MetricStrip :items="summaryMetrics" aria-label="图片摘要" />

      <Toolbar dense>
        <template #filters>
          <el-input
            v-model="quickSearch"
            class="quick-search"
            placeholder="搜索图片"
            clearable
            :prefix-icon="Search"
            @keyup.enter="applyQuickSearch"
            @clear="applyQuickSearch"
          />
          <el-tag v-for="chip in activeFilterChips" :key="chip.key" closable @close="removeFilter(chip.key)">
            {{ chip.label }}：{{ chip.value }}
          </el-tag>
          <el-button plain @click="openFilterDrawer">更多筛选</el-button>
        </template>
        <template #actions>
          <el-segmented
            class="image-view-switch"
            aria-label="图片视图模式"
            :model-value="viewMode"
            :options="[{ label: '网格', value: 'grid' }, { label: '列表', value: 'list' }]"
            @update:model-value="setViewMode"
          />
        </template>
      </Toolbar>

      <el-alert v-if="listError" type="error" :closable="false" :title="listError">
        <template #default><el-button link type="primary" @click="load">重试</el-button></template>
      </el-alert>

      <el-skeleton v-if="loading && list.length === 0" :rows="12" animated />

      <SectionCard v-else-if="!listError || list.length > 0" dense>
        <template #title>图片资产</template>
        <template #actions>
          <el-button v-if="selectedImageRows.length > 0" @click="clearImageSelection">取消选择</el-button>
        </template>

        <EmptyState
          v-if="list.length === 0"
          :title="hasActiveFilters ? '当前筛选无结果' : '暂无图片'"
          :description="hasActiveFilters ? '清除筛选后查看全部图片' : '上传图片后会显示在这里'"
        >
          <template #action>
            <el-button v-if="hasActiveFilters" @click="resetFilters">清除筛选</el-button>
            <el-button v-else type="primary" :icon="Upload" @click="openUploadDialog">上传图片</el-button>
          </template>
        </EmptyState>

        <div v-else-if="viewMode === 'grid'" class="image-grid">
          <article
            v-for="item in list"
            :key="item.id"
            class="image-grid-card"
            :class="{ 'is-selected': isGridSelected(item) }"
          >
            <el-checkbox
              class="image-grid-card__select"
              :aria-label="`选择图片：${item.title || item.id}`"
              :model-value="isGridSelected(item)"
              @update:model-value="(checked) => toggleGridSelection(item, checked)"
            />
            <div class="image-grid-card__preview">
              <img
                v-if="item.view_url || item.url || item.thumbnail_url"
                :src="item.view_url || item.url || item.thumbnail_url"
                :alt="item.title || '图片预览'"
              />
              <span v-else>{{ item.title || '图片' }}</span>
            </div>
            <div class="image-grid-card__body">
              <strong>{{ item.title || item.id }}</strong>
              <span>{{ item.width || 0 }} × {{ item.height || 0 }} · {{ formatFileSize(item.file_size) }}</span>
            </div>
            <div class="image-grid-card__meta">
              <StatusIndicator :label="statusLabel(item.status)" :tone="item.status === 'failed' ? 'danger' : 'success'" />
              <StatusIndicator :label="item.active ? '启用' : '停用'" :tone="item.active ? 'success' : 'warning'" />
            </div>
            <div class="image-grid-card__actions image-row-actions">
              <el-button link type="primary" @click="showDetail(item)">详情</el-button>
              <el-dropdown
                trigger="click"
                popper-class="image-row-actions-popper"
                @command="(command) => handleImageRowAction(command, item)"
              >
                <el-tooltip content="图片操作" placement="top">
                  <el-button :icon="MoreFilled" circle aria-label="图片操作" />
                </el-tooltip>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="toggle">{{ item.active ? '停用' : '启用' }}</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </article>
        </div>

        <div v-else class="table-wrap has-media-rows">
          <el-table ref="tableRef" :data="list" row-key="id" border>
            <el-table-column width="44">
              <template #header>
                <el-checkbox
                  class="image-table-selection"
                  aria-label="选择本页全部图片"
                  :model-value="allPageSelected"
                  :indeterminate="somePageSelected"
                  @update:model-value="toggleCurrentPageSelection"
                />
              </template>
              <template #default="{ row }">
                <el-checkbox
                  class="image-table-selection"
                  :aria-label="`选择图片：${row.title || row.id}`"
                  :model-value="isGridSelected(row)"
                  @update:model-value="(checked) => toggleGridSelection(row, checked)"
                />
              </template>
            </el-table-column>
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="104">
              <template #default="{ row }">
                <StatusIndicator :label="statusLabel(row.status)" :tone="row.status === 'failed' ? 'danger' : 'success'" />
              </template>
            </el-table-column>
            <el-table-column label="启用" width="88">
              <template #default="{ row }">{{ row.active ? '是' : '否' }}</template>
            </el-table-column>
            <el-table-column prop="stored_mime" label="格式" width="104" />
            <el-table-column label="尺寸" width="120">
              <template #default="{ row }">{{ row.width || 0 }} × {{ row.height || 0 }}</template>
            </el-table-column>
            <el-table-column label="文件大小" width="120">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="上传时间" width="168">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="108" fixed="right">
              <template #default="{ row }">
                <div class="image-row-actions">
                  <el-button link type="primary" @click="showDetail(row)">详情</el-button>
                  <el-dropdown
                    trigger="click"
                    popper-class="image-row-actions-popper"
                    @command="(command) => handleImageRowAction(command, row)"
                  >
                    <el-tooltip content="图片操作" placement="top">
                      <el-button :icon="MoreFilled" circle aria-label="图片操作" />
                    </el-tooltip>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="toggle">{{ row.active ? '停用' : '启用' }}</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="toolbar-row toolbar-row--end">
          <AdminTablePagination
            :current-page="query.page"
            :page-size="query.page_size"
            layout="total, prev, pager, next"
            :total="total"
            @current-change="setPage"
          />
        </div>
      </SectionCard>

      <BulkActionBar :count="selectedImageRows.length" :actions="bulkActions" />
    </div>

    <el-drawer
      v-model="uploadDialogVisible"
      title="新增图片"
      direction="rtl"
      :size="drawerSize"
      :close-on-click-modal="!uploading"
      :close-on-press-escape="!uploading"
      :before-close="handleUploadDrawerBeforeClose"
      @closed="onUploadDialogClosed"
    >
      <div class="drawer-body">
        <SectionCard dense>
          <template #title>批量选图</template>
          <el-upload
            ref="uploadRef"
            v-model:file-list="uploadFileList"
            drag
            multiple
            :auto-upload="false"
            :limit="100"
            accept="image/*"
            @exceed="handleUploadExceed"
          >
            <div class="el-upload__text">将图片拖到这里，或<em>点击选择文件</em></div>
            <template #tip>
              <div class="el-upload__tip">
                支持批量上传；非 GIF 文件会自动转换并保存为 WebP，源文件会在转换后删除。
              </div>
            </template>
          </el-upload>
        </SectionCard>

        <SectionCard dense collapsible>
          <template #title>默认关联</template>
          <el-form label-width="88px" class="upload-form">
            <el-form-item label="默认演员">
              <el-select
                v-model="uploadForm.actor_tokens"
                multiple
                filterable
                remote
                reserve-keyword
                allow-create
                default-first-option
                clearable
                :remote-method="searchActors"
                :loading="loadingActors"
                placeholder="可搜索演员，也可输入新演员姓名"
                style="width: 100%"
              >
                <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="默认合集">
              <el-select
                v-model="uploadForm.collection_ids"
                multiple
                filterable
                remote
                reserve-keyword
                clearable
                collapse-tags
                collapse-tags-tooltip
                :remote-method="searchImageCollections"
                :loading="loadingCollections"
                placeholder="可选，可多选"
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
            <el-form-item label="默认备注">
              <el-input
                v-model="uploadForm.description"
                type="textarea"
                :rows="2"
                placeholder="可选，批量上传时会作为所有图片的描述"
              />
            </el-form-item>
          </el-form>
        </SectionCard>
      </div>

      <div v-if="uploadSummary" class="upload-summary">
        <el-alert
          type="success"
          :closable="false"
          :title="`上传结果：总计 ${uploadSummary.total_count || 0}，成功 ${uploadSummary.success_count || 0}，失败 ${uploadSummary.failed_count || 0}`"
        />
        <el-table :data="uploadSummary.items || []" size="small" border>
          <el-table-column prop="filename" label="文件名" min-width="220" show-overflow-tooltip />
          <el-table-column label="结果" width="120">
            <template #default="{ row }">
              <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="说明" min-width="140" show-overflow-tooltip />
          <el-table-column prop="image_id" label="图片ID" min-width="220" />
          <el-table-column prop="error" label="失败原因" min-width="200" show-overflow-tooltip />
        </el-table>
      </div>

      <template #footer>
        <el-button :disabled="uploading" @click="requestUploadDrawerClose">取消</el-button>
        <el-button :disabled="uploading" @click="resetUploadForm">清空默认参数</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">开始上传</el-button>
      </template>
    </el-drawer>

    <el-drawer
      v-model="detailVisible"
      title="图片详情"
      direction="rtl"
      :size="drawerSize"
      :before-close="handleDetailDrawerBeforeClose"
      @closed="onDetailClosed"
    >
      <div v-if="detail" class="drawer-body">
        <SectionCard dense>
          <template #title>基础信息</template>
          <el-form label-width="88px">
            <el-form-item label="标题">
              <el-input v-model="detail.title" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="detail.description" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="启用状态">
              <el-switch v-model="detail.active" active-text="启用" inactive-text="停用" />
            </el-form-item>
            <el-form-item label="关联演员">
              <el-select
                v-model="detail.actor_tokens"
                multiple
                filterable
                remote
                reserve-keyword
                allow-create
                default-first-option
                clearable
                :remote-method="searchActors"
                :loading="loadingActors"
                placeholder="可搜索演员，也可输入新演员姓名"
                style="width: 100%"
              >
                <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="关联合集">
              <el-select
                v-model="detail.collection_ids"
                multiple
                filterable
                remote
                reserve-keyword
                clearable
                collapse-tags
                collapse-tags-tooltip
                :remote-method="searchImageCollections"
                :loading="loadingCollections"
                placeholder="可选，可多选"
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
            <el-form-item label="Metadata">
              <el-input v-model="detail.metadata_text" type="textarea" :rows="8" />
            </el-form-item>
          </el-form>
        </SectionCard>

        <SectionCard dense collapsible>
          <template #title>图片查看</template>
          <div class="preview-controls">
            <el-input-number v-model="preview.w" :min="0" :step="50" controls-position="right" placeholder="宽" />
            <el-input-number v-model="preview.h" :min="0" :step="50" controls-position="right" placeholder="高" />
            <el-select v-model="preview.fit" style="width: 120px">
              <el-option label="inside" value="inside" />
              <el-option label="cover" value="cover" />
              <el-option label="contain" value="contain" />
            </el-select>
            <el-input-number v-model="preview.q" :min="1" :max="100" :step="1" controls-position="right" />
            <el-button :loading="preview.loading" type="primary" plain @click="loadPreview">应用缩放接口</el-button>
          </div>
          <div class="zoom-row">
            <span>前端缩放</span>
            <el-slider v-model="preview.zoom" :min="20" :max="300" :step="5" style="flex: 1" />
            <span class="zoom-label">{{ preview.zoom }}%</span>
          </div>
          <el-alert v-if="preview.error" type="error" :title="preview.error" :closable="false" />
          <div class="preview-canvas" v-loading="preview.loading">
            <img
              v-if="preview.url"
              :src="preview.url"
              :alt="`图片预览：${detail.title || detail.id}`"
              class="preview-image"
              :style="{ transform: `scale(${preview.zoom / 100})` }"
            />
            <div v-else class="preview-placeholder">暂无图片预览</div>
          </div>
          <div class="preview-meta">
            <span>原始尺寸：{{ detail.width || 0 }} x {{ detail.height || 0 }}</span>
            <span>存储格式：{{ detail.stored_mime || '-' }}</span>
            <span>文件大小：{{ formatFileSize(detail.file_size) }}</span>
          </div>
        </SectionCard>
      </div>

      <template #footer>
        <el-button @click="requestDetailDrawerClose">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveDetail">保存</el-button>
      </template>
    </el-drawer>

    <el-drawer
      v-model="filterDrawerVisible"
      title="更多筛选"
      direction="rtl"
      :size="drawerSize"
      @closed="syncFilterDraftFromQuery"
    >
      <el-form label-width="88px">
        <el-form-item label="搜索">
          <el-input v-model="filterDraft.q" placeholder="按标题或描述搜索" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterDraft.status" placeholder="状态筛选" clearable style="width: 100%">
            <el-option label="可用" value="ready" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-select v-model="filterDraft.active" placeholder="启用状态" clearable style="width: 100%">
            <el-option label="全部状态" value="" />
            <el-option label="仅启用" value="1" />
            <el-option label="仅停用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="演员">
          <el-select
            v-model="filterDraft.actor_id"
            filterable
            remote
            clearable
            reserve-keyword
            placeholder="按演员筛选"
            style="width: 100%"
            :remote-method="searchActors"
            :loading="loadingActors"
          >
            <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片合集">
          <el-select
            v-model="filterDraft.collection_id"
            filterable
            remote
            clearable
            reserve-keyword
            placeholder="按图片合集筛选"
            style="width: 100%"
            :remote-method="searchImageCollections"
            :loading="loadingCollections"
          >
            <el-option
              v-for="collection in imageCollectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetFilters">重置</el-button>
        <el-button type="primary" @click="applyFilterDrawer">应用筛选</el-button>
      </template>
    </el-drawer>
  </Layout>
</template>

<style scoped>
.image-page {
  gap: var(--space-4);
  padding-bottom: var(--space-1);
}

.image-header-actions,
.image-row-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-1);
}

.image-header-actions {
  flex-wrap: wrap;
  gap: var(--space-2);
}

.image-row-actions {
  justify-content: flex-end;
}

.image-row-actions :deep(.el-button:focus-visible) {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

.quick-search {
  width: 240px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(184px, 1fr));
  gap: 12px;
}

.image-grid-card {
  position: relative;
  display: grid;
  gap: var(--space-2);
  padding: 8px;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  transition: border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.image-grid-card.is-selected,
.image-grid-card:hover {
  border-color: var(--primary);
}

.image-grid-card__select {
  position: absolute;
  top: var(--space-2);
  left: var(--space-2);
  z-index: 2;
}

.image-grid-card__select :deep(.el-checkbox__input.is-focus .el-checkbox__inner) {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

.image-grid-card__preview {
  display: grid;
  aspect-ratio: 4 / 3;
  place-items: center;
  overflow: hidden;
  border-radius: var(--radius-md);
  background-color: var(--bg-surface-muted);
  background-image:
    linear-gradient(45deg, var(--line-soft) 25%, transparent 25%),
    linear-gradient(-45deg, var(--line-soft) 25%, transparent 25%),
    linear-gradient(45deg, transparent 75%, var(--line-soft) 75%),
    linear-gradient(-45deg, transparent 75%, var(--line-soft) 75%);
  background-position: 0 0, 0 8px, 8px -8px, -8px 0;
  background-size: 16px 16px;
  color: var(--text-muted);
  font-size: var(--text-small);
}

.image-grid-card__preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.image-grid-card__body {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.image-grid-card__body strong {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-body);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-grid-card__body span,
.preview-placeholder,
.preview-meta {
  color: var(--text-muted);
}

.image-grid-card__meta,
.image-grid-card__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.image-grid-card__actions {
  align-items: center;
  justify-content: space-between;
  opacity: 1;
}

.has-media-rows :deep(.el-table) {
  min-width: 960px;
}

.drawer-body,
.upload-form {
  width: 100%;
}

.drawer-body {
  display: grid;
  gap: var(--space-3);
}

.upload-summary {
  margin-top: var(--space-3);
  display: grid;
  gap: var(--space-3);
}

.preview-controls {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  align-items: center;
  margin-bottom: var(--space-3);
}

.zoom-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-3);
}

.zoom-label {
  width: 50px;
  text-align: right;
  color: var(--primary);
  font-variant-numeric: tabular-nums;
}

.preview-canvas {
  min-height: 340px;
  max-height: 440px;
  border: 1px dashed var(--line-strong);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
  overflow: auto;
  display: grid;
  place-items: center;
  padding: var(--space-4);
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  transform-origin: center center;
  transition: transform 160ms ease;
}

.preview-meta {
  margin-top: var(--space-3);
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  font-size: var(--text-caption);
}

@media (max-width: 63.9375rem) {
  .image-header-actions :deep(.el-button) {
    min-height: 44px;
  }

  .image-view-switch :deep(.el-segmented__item) {
    min-width: 44px;
    min-height: 44px;
  }

  .image-row-actions :deep(.el-button) {
    min-width: 44px;
    min-height: 44px;
  }

  :global(.image-row-actions-popper .el-dropdown-menu__item) {
    min-height: 44px;
  }

  .quick-search {
    width: 100%;
  }
}
</style>
