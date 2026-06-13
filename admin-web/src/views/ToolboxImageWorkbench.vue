<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Back, Download, FolderOpened, Picture, Plus, Refresh, Search, Upload, WarningFilled } from '@element-plus/icons-vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import ImageWorkbenchMaskEditor from './ImageWorkbenchMaskEditor.vue'
import {
  generateAdminImage,
  getAdminImageCollections,
  getAdminImageDetail,
  getAdminImageGenerationStatus,
  getAdminImageViewBlob,
  getAdminImages,
  updateAdminImage,
  uploadAdminImages
} from '../api/admin'
import {
  IMAGE_WORKBENCH_LIMITS,
  buildImageGenerationPayload,
  buildReferenceImageSnapshots,
  createWorkbenchMaskPreview,
  createImageWorkbenchTask,
  dataUrlToFile,
  extensionForImageMime,
  formatWorkbenchFileSize,
  hydrateReferenceImageFromSnapshot,
  normalizeImageWorkbenchParams,
  readWorkbenchPreferences,
  validateWorkbenchMask,
  validateReferenceImageFiles,
  writeWorkbenchPreferences
} from './imageWorkbench.helpers'
import {
  getAllWorkbenchTasks,
  getWorkbenchImage,
  getWorkbenchImagePreview,
  putWorkbenchTask,
  storeWorkbenchImage
} from './imageWorkbench.db'

const LIBRARY_REFERENCE_PREVIEW_PARAMS = {
  w: 360,
  h: 270,
  fit: 'cover',
  q: 78
}
const LIBRARY_REFERENCE_ALLOWED_MIMES = 'image/png,image/jpeg,image/webp'

const router = useRouter()
const fileInputRef = ref(null)
const prompt = ref('')
const params = reactive(readWorkbenchPreferences())
const references = ref([])
const maskDraft = ref(null)
const maskEditorVisible = ref(false)
const maskEditorTargetId = ref('')
const status = ref({
  loading: false,
  enabled: false,
  model: '',
  base_url_configured: false,
  api_key_configured: false,
  timeout_seconds: 180
})
const generating = ref(false)
const importingIds = ref(new Set())
const tasks = ref([])
const selectedTask = ref(null)
const currentResults = ref([])
const currentTaskReferences = ref([])
const currentTaskReferenceError = ref('')
const dropActive = ref(false)
const libraryPickerVisible = ref(false)
const libraryLoading = ref(false)
const libraryAdding = ref(false)
const libraryItems = ref([])
const libraryTotal = ref(0)
const librarySelectedItems = ref([])
const libraryPreviewUrls = ref({})
const libraryPreviewErrors = ref({})
const libraryRequestSeq = ref(0)
const libraryCollectionOptions = ref([])
const libraryCollectionsLoading = ref(false)
const libraryQuery = reactive({
  page: 1,
  page_size: 12,
  q: '',
  collection_id: ''
})
const selectedTaskError = computed(() => {
  if (selectedTask.value?.status !== 'error') return ''
  return selectedTask.value.error?.trim() || '未返回错误原因'
})
const canContinueSelectedTask = computed(() => Boolean(selectedTask.value && !currentTaskReferenceError.value))
const maskEditorTarget = computed(() => references.value.find((item) => item.id === maskEditorTargetId.value) || null)
const selectedMaskTarget = computed(() => references.value.find((item) => item.id === maskDraft.value?.targetImageId) || null)
const hasMask = computed(() => Boolean(maskDraft.value?.maskDataUrl && selectedMaskTarget.value))
const activeInputMode = computed(() => {
  if (hasMask.value) return '局部蒙版编辑'
  return hasReferences.value ? '参考图编辑' : '文本生图'
})
const maskEditorInitialMaskDataUrl = computed(() => {
  if (maskDraft.value?.targetImageId !== maskEditorTargetId.value) return ''
  return maskDraft.value?.maskDataUrl || ''
})
const maskSummary = computed(() => {
  if (!hasMask.value || !selectedMaskTarget.value) return ''
  return `当前正在编辑「${selectedMaskTarget.value.name || '目标参考图'}」的局部区域`
})

const canGenerate = computed(() => status.value.enabled && !generating.value && prompt.value.trim() !== '')
const hasReferences = computed(() => references.value.length > 0)
const remainingReferenceSlots = computed(() => Math.max(0, IMAGE_WORKBENCH_LIMITS.maxReferenceImages - references.value.length))
const statusType = computed(() => {
  if (status.value.loading) return 'info'
  return status.value.enabled ? 'success' : 'warning'
})
const statusLabel = computed(() => {
  if (status.value.loading) return '检查配置中'
  if (status.value.enabled) return `已配置 · ${status.value.model || '图像模型'}`
  return '后端未配置图像生成服务'
})

watch(
  () => ({ ...params }),
  (value) => writeWorkbenchPreferences(value),
  { deep: true }
)

watch(
  () => params.output_format,
  (format) => {
    if (format !== 'png' && (params.output_compression === null || params.output_compression === undefined)) {
      params.output_compression = 82
    }
    if (format === 'png') {
      params.output_compression = null
    }
  }
)

onMounted(() => {
  void loadStatus()
  void loadHistory()
  window.addEventListener('paste', handlePaste)
})

onBeforeUnmount(() => {
  window.removeEventListener('paste', handlePaste)
  clearLibraryPreviewUrls()
})

function returnToToolbox() {
  router.push('/toolbox')
}

function openMediaLibrary() {
  const href = router.resolve('/images').href
  window.open(href, '_blank', 'noopener,noreferrer')
}

async function openLibraryPicker() {
  if (remainingReferenceSlots.value <= 0) {
    ElMessage.warning(`参考图最多 ${IMAGE_WORKBENCH_LIMITS.maxReferenceImages} 张`)
    return
  }
  libraryPickerVisible.value = true
  await Promise.all([
    loadLibraryImages(),
    searchLibraryCollections('')
  ])
}

function buildLibraryImageParams() {
  const params = {
    page: libraryQuery.page,
    page_size: libraryQuery.page_size,
    q: libraryQuery.q,
    status: 'ready',
    active: 1,
    stored_mime: LIBRARY_REFERENCE_ALLOWED_MIMES
  }
  if (libraryQuery.collection_id) {
    params.collection_id = libraryQuery.collection_id
  }
  return params
}

async function loadLibraryImages() {
  const requestSeq = libraryRequestSeq.value + 1
  libraryRequestSeq.value = requestSeq
  libraryLoading.value = true
  try {
    const data = await getAdminImages(buildLibraryImageParams())
    if (libraryRequestSeq.value !== requestSeq) return
    libraryItems.value = data.items || []
    libraryTotal.value = Number(data.total_count || 0)
    syncLibrarySelectionWithCurrentReferences()
    await loadLibraryPreviewUrls(libraryItems.value, requestSeq)
  } catch (error) {
    if (libraryRequestSeq.value === requestSeq) {
      libraryItems.value = []
      libraryTotal.value = 0
      clearLibraryPreviewUrls()
      ElMessage.error(extractErrorMessage(error, '读取图库资产失败'))
    }
  } finally {
    if (libraryRequestSeq.value === requestSeq) {
      libraryLoading.value = false
    }
  }
}

async function searchLibraryCollections(keyword = '') {
  libraryCollectionsLoading.value = true
  try {
    const data = await getAdminImageCollections({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 50
    })
    const optionMap = new Map(libraryCollectionOptions.value.map((item) => [item.value, item]))
    for (const item of data.items || []) {
      if (!item?.id) continue
      optionMap.set(item.id, { value: item.id, label: item.name || item.id })
    }
    libraryCollectionOptions.value = Array.from(optionMap.values())
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '读取图片合集失败'))
  } finally {
    libraryCollectionsLoading.value = false
  }
}

function applyLibrarySearch() {
  libraryQuery.page = 1
  void loadLibraryImages()
}

function resetLibraryFilters() {
  libraryQuery.q = ''
  libraryQuery.collection_id = ''
  libraryQuery.page = 1
  void loadLibraryImages()
}

function onLibraryPickerClosed() {
  libraryRequestSeq.value += 1
  clearLibraryPreviewUrls()
  libraryItems.value = []
  libraryTotal.value = 0
  librarySelectedItems.value = []
  libraryLoading.value = false
  libraryAdding.value = false
}

function clearLibraryPreviewUrls() {
  for (const url of Object.values(libraryPreviewUrls.value)) {
    if (url) URL.revokeObjectURL(url)
  }
  libraryPreviewUrls.value = {}
  libraryPreviewErrors.value = {}
}

async function readLibraryImageBlob(blob, fallback) {
  if (!blob?.type?.includes('application/json')) {
    return blob
  }
  const text = await blob.text()
  let payload = null
  try {
    payload = JSON.parse(text)
  } catch (_) {
    payload = null
  }
  throw new Error(payload?.msg || fallback)
}

async function loadLibraryPreviewUrls(items, requestSeq) {
  clearLibraryPreviewUrls()
  if (!items.length) return
  const nextUrls = {}
  const nextErrors = {}
  await Promise.all(
    items.map(async (item) => {
      if (libraryDisplayUrl(item)) return
      try {
        const blob = await getAdminImageViewBlob(item.id, LIBRARY_REFERENCE_PREVIEW_PARAMS)
        const imageBlob = await readLibraryImageBlob(blob, '加载图库缩略图失败')
        nextUrls[item.id] = URL.createObjectURL(imageBlob)
      } catch (error) {
        nextErrors[item.id] = extractErrorMessage(error, '加载图库缩略图失败')
      }
    })
  )
  if (libraryRequestSeq.value !== requestSeq) {
    for (const url of Object.values(nextUrls)) {
      if (url) URL.revokeObjectURL(url)
    }
    return
  }
  libraryPreviewUrls.value = nextUrls
  libraryPreviewErrors.value = nextErrors
}

function libraryDisplayUrl(item) {
  return item?.view_url || item?.url || item?.thumbnail_url || libraryPreviewUrls.value[item?.id] || ''
}

function isLibraryItemSelected(item) {
  return librarySelectedItems.value.some((selected) => selected.id === item.id)
}

function isLibraryItemAlreadyReferenced(item) {
  return references.value.some((reference) => reference.sourceKind === 'library_asset' && reference.sourceImageId === item.id)
}

function syncLibrarySelectionWithCurrentReferences() {
  librarySelectedItems.value = librarySelectedItems.value.filter((item) => !isLibraryItemAlreadyReferenced(item))
}

function toggleLibrarySelection(item, checked) {
  if (!item?.id) return
  if (isLibraryItemAlreadyReferenced(item)) {
    ElMessage.info('该图库资产已在参考图中')
    librarySelectedItems.value = librarySelectedItems.value.filter((selected) => selected.id !== item.id)
    return
  }
  const next = librarySelectedItems.value.filter((selected) => selected.id !== item.id)
  if (checked) {
    if (next.length >= remainingReferenceSlots.value) {
      ElMessage.warning(`还可以选择 ${remainingReferenceSlots.value} 张图库参考图`)
      return
    }
    next.push(item)
  }
  librarySelectedItems.value = next
}

function clearLibrarySelection() {
  librarySelectedItems.value = []
}

async function addSelectedLibraryReferences() {
  const selected = librarySelectedItems.value.filter((item) => !isLibraryItemAlreadyReferenced(item))
  if (!selected.length) {
    ElMessage.warning('请先选择图库图片')
    return
  }
  const targets = selected.slice(0, remainingReferenceSlots.value)
  if (targets.length < selected.length) {
    ElMessage.warning(`参考图最多 ${IMAGE_WORKBENCH_LIMITS.maxReferenceImages} 张，已按剩余槽位加入`)
  }
  if (!targets.length) {
    ElMessage.warning(`参考图最多 ${IMAGE_WORKBENCH_LIMITS.maxReferenceImages} 张`)
    return
  }

  libraryAdding.value = true
  const descriptors = []
  const failed = []
  const frozenAt = Date.now()
  for (const item of targets) {
    try {
      const blob = await getAdminImageViewBlob(item.id)
      const imageBlob = await readLibraryImageBlob(blob, '读取图库图片失败')
      const dataUrl = await blobToDataUrl(imageBlob)
      const mime = imageBlob.type || item.stored_mime || 'image/png'
      const name = libraryReferenceFilename(item, mime)
      descriptors.push({
        file: new File([imageBlob], name, { type: mime }),
        name,
        mime,
        size: imageBlob.size || item.file_size || 0,
        dataUrl,
        sourceKind: 'library_asset',
        sourceImageId: item.id,
        sourceTitle: item.title || item.id,
        sourceStatus: item.status || '',
        sourceActive: item.active,
        sourceViewUrl: item.view_url || item.url || `/api/v1/admin/images/${encodeURIComponent(item.id)}/view`,
        sourceFrozenAt: frozenAt
      })
    } catch (error) {
      failed.push(`${item.title || item.id}：${extractErrorMessage(error, '读取失败')}`)
    }
  }

  try {
    const added = await appendReferenceItems(descriptors)
    if (added.length > 0) {
      librarySelectedItems.value = librarySelectedItems.value.filter(
        (item) => !added.some((reference) => reference.sourceImageId === item.id)
      )
      syncLibrarySelectionWithCurrentReferences()
      ElMessage.success(`已加入 ${added.length} 张图库参考图`)
    }
    if (failed.length > 0) {
      ElMessage.warning(`部分图库图片加入失败：${failed.join('；')}`)
    }
    if (added.length > 0 && failed.length === 0) {
      libraryPickerVisible.value = false
    }
  } finally {
    libraryAdding.value = false
  }
}

async function loadStatus() {
  status.value = { ...status.value, loading: true }
  try {
    const data = await getAdminImageGenerationStatus()
    status.value = { ...status.value, ...data, loading: false }
  } catch (error) {
    status.value = { ...status.value, enabled: false, loading: false }
    ElMessage.error(extractErrorMessage(error, '读取图像生成配置失败'))
  }
}

async function loadHistory() {
  try {
    tasks.value = await hydrateTasks(await getAllWorkbenchTasks())
    if (!selectedTask.value && tasks.value.length) {
      await selectTask(tasks.value[0])
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '读取本地历史失败'))
  }
}

async function hydrateTasks(items) {
  return Promise.all(
    (items || []).map(async (task) => ({
      ...task,
      thumbnail: task.outputImageIds?.[0] ? await getWorkbenchImagePreview(task.outputImageIds[0]) : ''
    }))
  )
}

async function selectTask(task) {
  selectedTask.value = task
  currentResults.value = []
  currentTaskReferences.value = []
  currentTaskReferenceError.value = ''
  const [results, taskReferences] = await Promise.all([
    Promise.all(
      (task.outputImageIds || []).map(async (id, index) => {
        const image = await getWorkbenchImage(id)
        return {
          id,
          dataUrl: image?.dataUrl || '',
          mime: image?.mime || task.results?.[index]?.mime || 'image/png',
          width: image?.width || task.results?.[index]?.width || 0,
          height: image?.height || task.results?.[index]?.height || 0,
          revisedPrompt: task.results?.[index]?.revised_prompt || task.results?.[index]?.revisedPrompt || ''
        }
      })
    ),
    hydrateTaskReferenceSnapshots(task)
  ])
  if (selectedTask.value?.id !== task.id) return
  currentResults.value = results
  currentTaskReferences.value = taskReferences
  currentTaskReferenceError.value = taskReferences.some((item) => item.missing)
    ? '部分历史参考图的本地冻结文件已丢失，不能直接继续改稿'
    : ''
}

async function hydrateTaskReferenceSnapshots(task) {
  const snapshots = task?.referenceSnapshots?.length
    ? [...task.referenceSnapshots].sort((a, b) => Number(a.slot_index || 0) - Number(b.slot_index || 0))
    : (task?.referenceImageIds || []).map((id, index) => ({ image_id: id, slot_index: index, source_kind: 'browser_input' }))

  return Promise.all(
    snapshots.map(async (snapshot) => {
      const image = snapshot.image_id ? await getWorkbenchImage(snapshot.image_id) : null
      if (!image?.dataUrl) {
        return {
          ...hydrateReferenceImageFromSnapshot(snapshot, { id: snapshot.image_id || '' }),
          missing: true
        }
      }
      return {
        ...hydrateReferenceImageFromSnapshot(snapshot, image),
        missing: false
      }
    })
  )
}

function triggerFilePicker() {
  fileInputRef.value?.click()
}

async function handleFileInput(event) {
  await addReferenceFiles(event.target.files)
  event.target.value = ''
}

async function handleDrop(event) {
  dropActive.value = false
  await addReferenceFiles(event.dataTransfer?.files)
}

async function handlePaste(event) {
  const files = Array.from(event.clipboardData?.files || []).filter((file) => file.type?.startsWith('image/'))
  if (files.length) {
    await addReferenceFiles(files)
  }
}

async function addReferenceFiles(files) {
  const descriptors = Array.from(files || []).map((file) => ({
    file,
    name: file.name,
    mime: file.type,
    size: file.size,
    sourceKind: 'browser_input'
  }))
  await appendReferenceItems(descriptors)
}

async function appendReferenceItems(items) {
  const incoming = Array.from(items || [])
  if (!incoming.length) return []
  const existingForValidation = references.value.map((item) => ({
    type: item.mime,
    size: Number(item.size || 0)
  }))
  const incomingForValidation = incoming.map((item) => item.file || {
    type: item.mime,
    size: Number(item.size || 0)
  })
  const merged = [...existingForValidation, ...incomingForValidation]
  const validation = validateReferenceImageFiles(merged)
  if (!validation.ok) {
    ElMessage.warning(validation.message)
    return []
  }
  const nextItems = []
  for (const incomingItem of incoming) {
    const file = incomingItem.file
    const dataUrl = incomingItem.dataUrl || (file ? await readFileAsDataUrl(file) : '')
    const id = await storeWorkbenchImage(dataUrl, 'reference', incomingItem.name || file?.name || 'reference-image')
    nextItems.push({
      id,
      file,
      name: incomingItem.name || file?.name || 'reference-image',
      mime: incomingItem.mime || file?.type || 'image/png',
      size: incomingItem.size ?? file?.size ?? 0,
      dataUrl,
      sourceKind: incomingItem.sourceKind || 'browser_input',
      sourceTaskId: incomingItem.sourceTaskId || '',
      sourceResultId: incomingItem.sourceResultId || '',
      sourceImageId: incomingItem.sourceImageId || '',
      sourceTitle: incomingItem.sourceTitle || '',
      sourceStatus: incomingItem.sourceStatus || '',
      sourceActive: incomingItem.sourceActive,
      sourceViewUrl: incomingItem.sourceViewUrl || '',
      sourceFrozenAt: incomingItem.sourceFrozenAt || null
    })
  }
  references.value = [...references.value, ...nextItems]
  return nextItems
}

function clearMaskDraftIfTargetRemoved(id) {
  if (maskDraft.value?.targetImageId === id) {
    maskDraft.value = null
  }
  if (maskEditorTargetId.value === id) {
    maskEditorTargetId.value = ''
    maskEditorVisible.value = false
  }
}

function removeReference(id) {
  clearMaskDraftIfTargetRemoved(id)
  references.value = references.value.filter((item) => item.id !== id)
}

function clearReferences() {
  references.value = []
  maskDraft.value = null
  maskEditorTargetId.value = ''
  maskEditorVisible.value = false
}

async function continueSelectedTaskDraft() {
  if (!selectedTask.value) return
  if (currentTaskReferenceError.value) {
    ElMessage.warning(currentTaskReferenceError.value)
    return
  }
  prompt.value = selectedTask.value.prompt || ''
  Object.assign(params, normalizeImageWorkbenchParams(selectedTask.value.params || {}))
  references.value = currentTaskReferences.value.map((item) => ({
    ...item,
    file: null
  }))
  maskDraft.value = null
  maskEditorTargetId.value = ''
  maskEditorVisible.value = false
  ElMessage.success('已从历史恢复输入快照')
}

function isMaskTarget(id) {
  return maskDraft.value?.targetImageId === id
}

function getReferencePreview(item) {
  if (isMaskTarget(item.id) && maskDraft.value?.previewDataUrl) {
    return maskDraft.value.previewDataUrl
  }
  return item.dataUrl
}

function openMaskEditor(id) {
  const target = references.value.find((item) => item.id === id)
  if (!target) return
  maskEditorTargetId.value = id
  maskEditorVisible.value = true
}

async function handleMaskSaved(maskDataUrl) {
  if (!maskEditorTarget.value?.dataUrl) {
    ElMessage.error('遮罩目标参考图已不存在，请重新选择')
    return
  }
  try {
    const previewDataUrl = await createWorkbenchMaskPreview(maskEditorTarget.value.dataUrl, maskDataUrl)
    maskDraft.value = {
      targetImageId: maskEditorTarget.value.id,
      maskDataUrl,
      previewDataUrl
    }
    maskEditorVisible.value = false
    ElMessage.success('已保存局部蒙版')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存蒙版失败'))
  }
}

function removeMask() {
  if (!maskDraft.value) return
  maskDraft.value = null
  ElMessage.success('已移除蒙版')
}

function buildReferenceSnapshots() {
  return buildReferenceImageSnapshots(references.value)
}

function buildMaskTaskSnapshot(maskImageId) {
  if (!maskImageId || !maskDraft.value?.targetImageId) return null
  const targetIndex = references.value.findIndex((item) => item.id === maskDraft.value?.targetImageId)
  if (targetIndex < 0) return null
  return {
    image_id: maskImageId,
    mime: 'image/png',
    target_reference_index: targetIndex
  }
}

async function normalizeMaskDraft(allowFullMask = false) {
  if (!maskDraft.value?.targetImageId || !maskDraft.value?.maskDataUrl) {
    return null
  }
  const target = references.value.find((item) => item.id === maskDraft.value.targetImageId)
  if (!target?.dataUrl) {
    maskDraft.value = null
    throw new Error('遮罩目标参考图已不存在，请重新选择')
  }
  const coverage = await validateWorkbenchMask(maskDraft.value.maskDataUrl, target.dataUrl)
  if (coverage === 'empty') {
    throw new Error('请先涂抹需要编辑的区域')
  }
  if (coverage === 'full' && !allowFullMask) {
    await ElMessageBox.confirm(
      '当前遮罩覆盖了整张目标参考图，提交后可能会重绘全部内容。是否继续？',
      '确认编辑整张图片？',
      {
        confirmButtonText: '继续提交',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  }
  return maskDraft.value
}

async function generateImage(options = {}) {
  if (!canGenerate.value) return
  let activeMask = null
  try {
    activeMask = await normalizeMaskDraft(Boolean(options.allowFullMask))
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    if (error?.message === 'cancel' || error?.message === 'close') return
    ElMessage.error(extractErrorMessage(error, '校验蒙版失败'))
    return
  }
  generating.value = true
  const startedAt = Date.now()
  try {
    const payload = await buildImageGenerationPayload(prompt.value, params, references.value, activeMask)
    let storedMaskImageId = ''
    if (activeMask?.maskDataUrl) {
      storedMaskImageId = await storeWorkbenchImage(activeMask.maskDataUrl, 'mask', 'workbench-mask.png')
    }
    const referenceSnapshots = buildReferenceSnapshots()
    const taskMask = buildMaskTaskSnapshot(storedMaskImageId)
    const response = await generateAdminImage(payload)
    const resultItems = response?.items || []
    const outputImageIds = []
    const storedResults = []
    for (const item of resultItems) {
      const id = await storeWorkbenchImage(item.data_url, 'generated', 'generated-image')
      outputImageIds.push(id)
      storedResults.push({
        image_id: id,
        mime: item.mime,
        width: item.width || 0,
        height: item.height || 0,
        revised_prompt: item.revised_prompt || ''
      })
    }
    const task = createImageWorkbenchTask({
      prompt: prompt.value,
      params,
      referenceImageIds: references.value.map((item) => item.id),
      referenceSnapshots,
      outputImageIds,
      results: storedResults,
      mask: taskMask,
      status: 'done',
      startedAt,
      finishedAt: Date.now()
    })
    await putWorkbenchTask(task)
    await loadHistory()
    const fresh = tasks.value.find((item) => item.id === task.id) || task
    await selectTask(fresh)
    ElMessage.success(`已生成 ${outputImageIds.length} 张图片`)
  } catch (error) {
    const failedTask = createImageWorkbenchTask({
      prompt: prompt.value,
      params,
      referenceImageIds: references.value.map((item) => item.id),
      referenceSnapshots: buildReferenceSnapshots(),
      outputImageIds: [],
      results: [],
      mask: activeMask?.maskDataUrl ? buildMaskTaskSnapshot(await storeWorkbenchImage(activeMask.maskDataUrl, 'mask', 'workbench-mask.png')) : null,
      status: 'error',
      error: extractErrorMessage(error, '图像生成失败'),
      startedAt,
      finishedAt: Date.now()
    })
    await putWorkbenchTask(failedTask)
    await loadHistory()
    const fresh = tasks.value.find((item) => item.id === failedTask.id) || failedTask
    await selectTask(fresh)
    ElMessage.error(failedTask.error)
  } finally {
    generating.value = false
  }
}

function downloadResult(result, index) {
  const link = document.createElement('a')
  const ext = extensionForImageMime(result.mime)
  link.href = result.dataUrl
  link.download = `image-workbench-${index + 1}.${ext}`
  document.body.appendChild(link)
  link.click()
  link.remove()
}

async function reuseAsReference(result) {
  const file = dataUrlToFile(result.dataUrl, 'generated-reference.png', result.mime)
  const added = await appendReferenceItems([{
    file,
    name: file.name,
    mime: file.type,
    size: file.size,
    dataUrl: result.dataUrl,
    sourceKind: 'previous_result',
    sourceTaskId: selectedTask.value?.id || '',
    sourceResultId: result.id || ''
  }])
  if (added.length > 0) {
    ElMessage.success('已复用为参考图')
  }
}

async function importToMediaLibrary(result, index) {
  if (!result?.dataUrl) return
  importingIds.value = new Set([...importingIds.value, result.id])
  try {
    const file = dataUrlToFile(result.dataUrl, `image-workbench-${Date.now()}-${index + 1}.${extensionForImageMime(result.mime)}`, result.mime)
    const formData = new FormData()
    formData.append('files', file)
    if (selectedTask.value?.prompt) {
      formData.append('description', selectedTask.value.prompt)
    }
    const uploadResult = await uploadAdminImages(formData)
    const imageID = uploadResult?.items?.find((item) => item.success)?.image_id
    if (imageID) {
      const detail = await getAdminImageDetail(imageID)
      const existingMetadata = detail?.metadata && typeof detail.metadata === 'object' && !Array.isArray(detail.metadata) ? detail.metadata : {}
      await updateAdminImage(imageID, {
        metadata: {
          ...existingMetadata,
          source: 'admin_image_workbench',
          prompt: selectedTask.value?.prompt || '',
          model: status.value.model || '',
          size: selectedTask.value?.params?.size || '',
          quality: selectedTask.value?.params?.quality || '',
          output_format: selectedTask.value?.params?.output_format || '',
          generated_at: new Date(selectedTask.value?.finishedAt || Date.now()).toISOString(),
          local_task_id: selectedTask.value?.id || ''
        }
      })
    }
    ElMessage.success(imageID ? '已导入媒体库' : '上传完成，请在图片管理中确认结果')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '导入媒体库失败'))
  } finally {
    const next = new Set(importingIds.value)
    next.delete(result.id)
    importingIds.value = next
  }
}

function isImporting(result) {
  return importingIds.value.has(result.id)
}

function readFileAsDataUrl(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error || new Error('读取图片失败'))
    reader.readAsDataURL(file)
  })
}

function blobToDataUrl(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error || new Error('读取图库图片失败'))
    reader.readAsDataURL(blob)
  })
}

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim()) return responseMsg.trim()
  if (typeof error?.message === 'string' && error.message.trim()) return error.message.trim()
  return fallback
}

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', { hour12: false })
}

function formatDimensions(item) {
  if (!item?.width || !item?.height) return item?.mime || '图片'
  return `${item.width} x ${item.height}`
}

function libraryReferenceFilename(item, mime) {
  const title = sanitizeFilename(item?.title || item?.id || 'library-reference')
  return `${title}.${extensionForImageMime(mime)}`
}

function sanitizeFilename(value) {
  const normalized = String(value || '').trim().replace(/[\\/:*?"<>|]+/g, '-').replace(/\s+/g, '-')
  return normalized || 'library-reference'
}

function sourceKindLabel(item) {
  if (item?.sourceKind === 'library_asset') return '图库资产'
  if (item?.sourceKind === 'previous_result') return '前序结果'
  return '浏览器输入'
}

function referenceSourceSummary(item) {
  if (item?.sourceKind === 'library_asset') {
    const title = item.sourceTitle || '图库资产'
    return item.sourceImageId ? `来源：${title} · ID ${item.sourceImageId}` : `来源：${title}`
  }
  if (item?.sourceKind === 'previous_result') {
    return '来源：前序生成结果'
  }
  return ''
}

function openReferenceSource(item) {
  if (item?.sourceKind !== 'library_asset') return
  if (!item.sourceImageId) {
    ElMessage.warning('该图库来源缺少图片 ID，只能保留历史摘要')
    return
  }
  const href = router.resolve({
    path: '/images',
    query: { image_id: item.sourceImageId }
  }).href
  window.open(href, '_blank', 'noopener,noreferrer')
}

function formatLibraryItemMeta(item) {
  const dimensions = item?.width && item?.height ? `${item.width} x ${item.height}` : '尺寸未知'
  return `${dimensions} · ${formatWorkbenchFileSize(item?.file_size)}`
}
</script>

<template>
  <main class="image-workbench">
    <div class="image-workbench__inner">
      <header class="workbench-topbar">
        <div class="workbench-topbar__actions">
          <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
          <el-button :icon="FolderOpened" @click="openMediaLibrary">打开媒体库</el-button>
        </div>
        <el-tag :type="statusType" effect="plain">{{ statusLabel }}</el-tag>
      </header>

      <PageHeader title="图像生成工作台" subtitle="使用提示词和参考图生成、编辑、保存图像结果。" />

      <div class="workbench-grid">
        <aside class="workbench-panel workbench-panel--input">
          <SectionCard>
            <template #title>输入</template>
            <template #description>{{ activeInputMode }}</template>

            <div class="input-stack">
              <el-input
                v-model="prompt"
                type="textarea"
                :rows="8"
                resize="vertical"
                placeholder="描述要生成或编辑的图像"
              />

              <div
                class="reference-drop"
                :class="{ 'is-active': dropActive }"
                @dragover.prevent="dropActive = true"
                @dragleave.prevent="dropActive = false"
                @drop.prevent="handleDrop"
              >
                <input ref="fileInputRef" class="reference-drop__input" type="file" accept="image/png,image/jpeg,image/webp" multiple @change="handleFileInput" />
                <div class="reference-drop__actions">
                  <el-button :icon="Upload" @click="triggerFilePicker">上传参考图</el-button>
                  <el-button :icon="FolderOpened" :disabled="remainingReferenceSlots <= 0" @click="openLibraryPicker">从图库选择</el-button>
                </div>
                <span>支持拖拽、粘贴、上传 PNG/JPEG/WebP；还可添加 {{ remainingReferenceSlots }} 张</span>
                <small>图库资产加入后会立即冻结为本地参考图快照</small>
              </div>

              <div v-if="hasMask" class="mask-summary">
                <div>
                  <strong>局部蒙版已启用</strong>
                  <p>{{ maskSummary }}</p>
                </div>
                <div class="mask-summary__actions">
                  <el-button size="small" @click="openMaskEditor(maskDraft?.targetImageId)">继续编辑</el-button>
                  <el-button size="small" text @click="removeMask">移除蒙版</el-button>
                </div>
              </div>

              <div v-if="references.length" class="reference-list">
                <article v-for="item in references" :key="item.id" class="reference-card">
                  <img :src="getReferencePreview(item)" :alt="item.name" />
                  <div class="reference-card__copy">
                    <strong>{{ item.name }}</strong>
                    <span>{{ item.mime }}</span>
                    <div class="reference-card__tags">
                      <el-tag v-if="isMaskTarget(item.id)" size="small" type="primary">蒙版目标</el-tag>
                      <el-tag size="small" effect="plain">{{ sourceKindLabel(item) }}</el-tag>
                    </div>
                    <small v-if="referenceSourceSummary(item)" class="reference-card__source">{{ referenceSourceSummary(item) }}</small>
                    <el-button v-if="item.sourceKind === 'library_asset'" size="small" link @click="openReferenceSource(item)">查看来源</el-button>
                  </div>
                  <div class="reference-card__actions">
                    <el-button size="small" @click="openMaskEditor(item.id)">
                      {{ isMaskTarget(item.id) ? '编辑蒙版' : '添加蒙版' }}
                    </el-button>
                    <el-button v-if="isMaskTarget(item.id)" size="small" text @click="removeMask">移除蒙版</el-button>
                    <el-button size="small" text @click="removeReference(item.id)">移除参考图</el-button>
                  </div>
                </article>
                <el-button size="small" @click="clearReferences">清空参考图</el-button>
              </div>
            </div>
          </SectionCard>

          <SectionCard collapsible>
            <template #title>参数</template>
            <div class="param-grid">
              <label>
                <span>尺寸</span>
                <el-select v-model="params.size">
                  <el-option label="自动" value="auto" />
                  <el-option label="1024 x 1024" value="1024x1024" />
                  <el-option label="1024 x 1536" value="1024x1536" />
                  <el-option label="1536 x 1024" value="1536x1024" />
                </el-select>
              </label>
              <label>
                <span>质量</span>
                <el-select v-model="params.quality">
                  <el-option label="自动" value="auto" />
                  <el-option label="低" value="low" />
                  <el-option label="中" value="medium" />
                  <el-option label="高" value="high" />
                </el-select>
              </label>
              <label>
                <span>格式</span>
                <el-select v-model="params.output_format">
                  <el-option label="PNG" value="png" />
                  <el-option label="JPEG" value="jpeg" />
                  <el-option label="WebP" value="webp" />
                </el-select>
              </label>
              <label>
                <span>张数</span>
                <el-input-number v-model="params.n" :min="1" :max="4" controls-position="right" />
              </label>
              <label v-if="params.output_format !== 'png'" class="param-grid__wide">
                <span>压缩质量</span>
                <el-slider v-model="params.output_compression" :min="0" :max="100" />
              </label>
            </div>
            <el-button class="generate-button" type="primary" :icon="Picture" :loading="generating" :disabled="!canGenerate" @click="generateImage">
              生成图像
            </el-button>
          </SectionCard>
        </aside>

        <section class="workbench-panel workbench-panel--preview">
          <SectionCard>
            <template #title>当前结果</template>
            <template #actions>
              <el-button :disabled="!canContinueSelectedTask" @click="continueSelectedTaskDraft">继续改稿</el-button>
              <el-button :icon="Refresh" :loading="status.loading" @click="loadStatus">刷新配置</el-button>
            </template>

            <div v-if="selectedTask" class="task-input-summary">
              <div>
                <strong>历史输入</strong>
                <span>{{ selectedTask.params?.size || 'auto' }} · {{ selectedTask.params?.quality || 'auto' }} · {{ selectedTask.params?.output_format || 'png' }}</span>
              </div>
              <el-alert
                v-if="currentTaskReferenceError"
                type="warning"
                :title="currentTaskReferenceError"
                :closable="false"
              />
              <div v-if="currentTaskReferences.length" class="task-reference-strip">
                <article
                  v-for="item in currentTaskReferences"
                  :key="item.id || item.name"
                  class="task-reference-chip"
                  :class="{ 'is-missing': item.missing }"
                >
                  <img v-if="!item.missing && item.dataUrl" :src="item.dataUrl" :alt="item.name" />
                  <span v-else class="task-reference-chip__missing">缺失</span>
                  <span class="task-reference-chip__copy">
                    <strong>{{ item.name }}</strong>
                    <small>{{ sourceKindLabel(item) }}</small>
                    <small v-if="referenceSourceSummary(item)">{{ referenceSourceSummary(item) }}</small>
                    <el-button v-if="item.sourceKind === 'library_asset'" size="small" link @click="openReferenceSource(item)">查看来源</el-button>
                  </span>
                </article>
              </div>
              <p v-else class="task-input-summary__empty">该历史没有参考图，可按原提示词继续文本生图。</p>
            </div>

            <EmptyState
              v-if="selectedTaskError"
              :icon="WarningFilled"
              title="生成失败"
              :description="selectedTaskError"
            />
            <EmptyState v-else-if="!currentResults.length" title="暂无结果" description="生成图片后会在这里预览和操作。" />
            <div v-else class="result-grid">
              <article v-for="(item, index) in currentResults" :key="item.id" class="result-card">
                <img :src="item.dataUrl" :alt="'生成结果 ' + (index + 1)" />
                <div class="result-card__meta">
                  <strong>结果 {{ index + 1 }}</strong>
                  <span>{{ formatDimensions(item) }}</span>
                </div>
                <div class="result-card__actions">
                  <el-button :icon="Download" @click="downloadResult(item, index)">下载</el-button>
                  <el-button :icon="Plus" @click="reuseAsReference(item)">复用为参考图</el-button>
                  <el-button type="primary" :loading="isImporting(item)" @click="importToMediaLibrary(item, index)">导入媒体库</el-button>
                </div>
                <p v-if="item.revisedPrompt" class="result-card__prompt">{{ item.revisedPrompt }}</p>
              </article>
            </div>
          </SectionCard>
        </section>

        <aside class="workbench-panel workbench-panel--history">
          <SectionCard>
            <template #title>本地历史</template>
            <template #description>保存在当前浏览器</template>
            <template #actions>
              <el-button :icon="Refresh" @click="loadHistory">刷新</el-button>
            </template>

            <EmptyState v-if="!tasks.length" title="没有本地历史" description="生成结果后会自动保存到浏览器本地。" />
            <div v-else class="history-list">
              <button
                v-for="task in tasks"
                :key="task.id"
                type="button"
                class="history-item"
                :class="{ 'is-active': selectedTask?.id === task.id }"
                @click="selectTask(task)"
              >
                <img v-if="task.thumbnail" :src="task.thumbnail" alt="" />
                <span v-else class="history-item__empty">无图</span>
                <span class="history-item__copy">
                  <strong>{{ task.prompt || '未命名任务' }}</strong>
                  <small>{{ formatDate(task.createdAt) }}</small>
                  <small v-if="task.status === 'error' && task.error" class="history-item__error">{{ task.error }}</small>
                </span>
                <el-tag v-if="task.status === 'error'" size="small" type="danger">失败</el-tag>
              </button>
            </div>
          </SectionCard>
        </aside>
      </div>
    </div>
    <el-drawer
      v-model="libraryPickerVisible"
      title="选择图库参考图"
      direction="rtl"
      size="min(100%, 760px)"
      :close-on-click-modal="!libraryAdding"
      :close-on-press-escape="!libraryAdding"
      @closed="onLibraryPickerClosed"
    >
      <div class="library-picker">
        <SectionCard dense>
          <template #title>图库筛选</template>
          <div class="library-picker__filters">
            <el-input
              v-model="libraryQuery.q"
              placeholder="搜索图片标题或描述"
              clearable
              :prefix-icon="Search"
              @keyup.enter="applyLibrarySearch"
              @clear="applyLibrarySearch"
            />
            <el-select
              v-model="libraryQuery.collection_id"
              filterable
              remote
              clearable
              reserve-keyword
              placeholder="按图片合集筛选"
              :remote-method="searchLibraryCollections"
              :loading="libraryCollectionsLoading"
              @change="applyLibrarySearch"
              @clear="applyLibrarySearch"
            >
              <el-option
                v-for="collection in libraryCollectionOptions"
                :key="collection.value"
                :label="collection.label"
                :value="collection.value"
              />
            </el-select>
            <el-button type="primary" :loading="libraryLoading" @click="applyLibrarySearch">搜索</el-button>
            <el-button @click="resetLibraryFilters">重置</el-button>
          </div>
        </SectionCard>

        <SectionCard v-loading="libraryLoading" dense>
          <template #title>可用图片资产</template>
          <template #description>仅展示可用且启用的图库资产</template>
          <template #actions>
            <el-button :disabled="librarySelectedItems.length === 0 || libraryAdding" @click="clearLibrarySelection">取消选择</el-button>
          </template>

          <EmptyState v-if="!libraryLoading && libraryItems.length === 0" title="暂无可选图片" description="换个关键词或图片合集再试。" />
          <div v-else class="library-grid">
            <article
              v-for="item in libraryItems"
              :key="item.id"
              class="library-card"
              :class="{ 'is-selected': isLibraryItemSelected(item), 'is-disabled': isLibraryItemAlreadyReferenced(item) }"
            >
              <el-checkbox
                class="library-card__select"
                :model-value="isLibraryItemSelected(item)"
                :disabled="isLibraryItemAlreadyReferenced(item)"
                @update:model-value="(checked) => toggleLibrarySelection(item, checked)"
              />
              <div class="library-card__preview">
                <img v-if="libraryDisplayUrl(item)" :src="libraryDisplayUrl(item)" :alt="item.title || item.id" />
                <span v-else>{{ libraryPreviewErrors[item.id] || item.title || '图片' }}</span>
              </div>
              <div class="library-card__body">
                <strong>{{ item.title || item.id }}</strong>
                <span>{{ formatLibraryItemMeta(item) }}</span>
              </div>
              <div class="library-card__meta">
                <el-tag size="small" type="success">可用</el-tag>
                <el-tag v-if="isLibraryItemAlreadyReferenced(item)" size="small" type="info">已在参考图中</el-tag>
              </div>
            </article>
          </div>

          <div class="library-picker__pager">
            <el-pagination
              v-model:current-page="libraryQuery.page"
              v-model:page-size="libraryQuery.page_size"
              layout="total, prev, pager, next"
              :total="libraryTotal"
              @current-change="loadLibraryImages"
            />
          </div>
        </SectionCard>
      </div>

      <template #footer>
        <div class="library-picker__footer">
          <span>已选 {{ librarySelectedItems.length }} 张，还可加入 {{ remainingReferenceSlots }} 张</span>
          <div>
            <el-button :disabled="libraryAdding" @click="libraryPickerVisible = false">关闭</el-button>
            <el-button
              type="primary"
              :loading="libraryAdding"
              :disabled="librarySelectedItems.length === 0 || remainingReferenceSlots <= 0"
              @click="addSelectedLibraryReferences"
            >
              添加为参考图
            </el-button>
          </div>
        </div>
      </template>
    </el-drawer>
    <ImageWorkbenchMaskEditor
      v-model="maskEditorVisible"
      :image-data-url="maskEditorTarget?.dataUrl || ''"
      :image-name="maskEditorTarget?.name || ''"
      :initial-mask-data-url="maskEditorInitialMaskDataUrl"
      @save="handleMaskSaved"
    />
  </main>
</template>

<style scoped>
.image-workbench {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.image-workbench__inner {
  display: grid;
  width: min(100%, 118rem);
  margin: 0 auto;
  padding: var(--space-5);
  gap: var(--space-5);
}

.workbench-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.workbench-topbar__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.workbench-grid {
  display: grid;
  grid-template-columns: minmax(19rem, 24rem) minmax(0, 1fr) minmax(17rem, 22rem);
  gap: var(--space-4);
  align-items: start;
}

.workbench-panel {
  display: grid;
  min-width: 0;
  gap: var(--space-4);
}

.input-stack {
  display: grid;
  gap: var(--space-3);
}

.reference-drop {
  display: grid;
  gap: var(--space-2);
  place-items: center;
  min-height: 8rem;
  padding: var(--space-4);
  border: 1px dashed var(--line-strong);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  text-align: center;
  background: var(--bg-surface-muted);
}

.reference-drop.is-active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.reference-drop__input {
  display: none;
}

.reference-drop__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--space-2);
}

.reference-drop small {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.reference-list {
  display: grid;
  gap: var(--space-2);
}

.mask-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--primary) 30%, var(--line-soft));
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--primary) 8%, var(--bg-surface));
}

.mask-summary strong,
.mask-summary p {
  margin: 0;
}

.mask-summary p {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.mask-summary__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.reference-card {
  display: grid;
  grid-template-columns: 4rem minmax(0, 1fr) auto;
  align-items: start;
  gap: var(--space-2);
  min-width: 0;
  padding: var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
}

.reference-card img,
.history-item img,
.result-card img {
  display: block;
  width: 100%;
  object-fit: cover;
  background: var(--bg-surface-muted);
}

.reference-card img {
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
}

.reference-card__copy {
  display: grid;
  gap: 0.1rem;
  min-width: 0;
}

.reference-card strong,
.history-item strong {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.reference-card span,
.reference-card__source,
.history-item small,
.result-card__meta span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.reference-card__tags,
.reference-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.reference-card__actions {
  justify-content: flex-end;
}

.param-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.param-grid label {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.param-grid__wide {
  grid-column: 1 / -1;
}

.generate-button {
  width: 100%;
  margin-top: var(--space-4);
}

.result-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: var(--space-4);
}

.task-input-summary {
  display: grid;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.task-input-summary > div:first-child {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.task-input-summary span,
.task-input-summary__empty {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.task-input-summary__empty {
  margin: 0;
}

.task-reference-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
  gap: var(--space-2);
}

.task-reference-chip {
  display: grid;
  grid-template-columns: 3.5rem minmax(0, 1fr);
  gap: var(--space-2);
  align-items: center;
  min-width: 0;
  padding: var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  background: var(--bg-surface);
}

.task-reference-chip.is-missing {
  border-style: dashed;
}

.task-reference-chip img,
.task-reference-chip__missing {
  width: 3.5rem;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
}

.task-reference-chip img {
  object-fit: cover;
}

.task-reference-chip__missing {
  display: grid;
  place-items: center;
  color: var(--danger);
  background: color-mix(in srgb, var(--danger) 10%, var(--bg-surface-muted));
  font-size: var(--text-caption);
}

.task-reference-chip__copy {
  display: grid;
  gap: 0.1rem;
  min-width: 0;
}

.task-reference-chip__copy strong {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-card {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
}

.result-card img {
  max-height: 34rem;
  aspect-ratio: 1;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  object-fit: contain;
}

.result-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.result-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.result-card__prompt {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.history-list {
  display: grid;
  gap: var(--space-2);
  max-height: calc(100vh - 15rem);
  overflow: auto;
}

.history-item {
  display: grid;
  grid-template-columns: 4rem minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  min-width: 0;
  padding: var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  background: var(--bg-surface);
  text-align: left;
  cursor: pointer;
}

.history-item:hover,
.history-item.is-active {
  border-color: var(--primary);
}

.history-item img,
.history-item__empty {
  width: 4rem;
  aspect-ratio: 1;
  border-radius: var(--radius-sm);
}

.history-item__empty {
  display: grid;
  place-items: center;
  color: var(--text-muted);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
}

.history-item__copy {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.history-item__error {
  overflow: hidden;
  color: var(--danger);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  word-break: break-word;
}

.library-picker {
  display: grid;
  gap: var(--space-3);
}

.library-picker__filters {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr) auto auto;
  gap: var(--space-2);
  align-items: center;
}

.library-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
  gap: var(--space-3);
}

.library-card {
  position: relative;
  display: grid;
  gap: var(--space-2);
  min-width: 0;
  padding: var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  transition:
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.library-card.is-selected,
.library-card:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.library-card.is-disabled {
  opacity: 0.62;
}

.library-card__select {
  position: absolute;
  top: var(--space-2);
  left: var(--space-2);
  z-index: 2;
}

.library-card__preview {
  display: grid;
  aspect-ratio: 4 / 3;
  place-items: center;
  overflow: hidden;
  border-radius: var(--radius-sm);
  background: var(--bg-surface-muted);
  color: var(--text-muted);
  font-size: var(--text-caption);
  text-align: center;
}

.library-card__preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.library-card__body {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.library-card__body strong {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.library-card__body span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.library-card__meta,
.library-picker__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.library-picker__pager {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-3);
}

.library-picker__footer {
  justify-content: space-between;
  width: 100%;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.library-picker__footer > div {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

@media (max-width: 72rem) {
  .workbench-grid {
    grid-template-columns: 1fr;
  }

  .history-list {
    max-height: none;
  }
}

@media (max-width: 48rem) {
  .image-workbench__inner {
    padding: var(--space-4);
  }

  .mask-summary {
    align-items: stretch;
    flex-direction: column;
  }

  .mask-summary__actions {
    justify-content: flex-start;
  }

  .param-grid {
    grid-template-columns: 1fr;
  }

  .library-picker__filters {
    grid-template-columns: 1fr;
  }

  .library-picker__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .reference-card,
  .history-item {
    grid-template-columns: 3.5rem minmax(0, 1fr);
  }

  .reference-card__actions,
  .reference-card .el-button,
  .history-item .el-tag {
    grid-column: 2;
    justify-self: start;
  }
}
</style>
