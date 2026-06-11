<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Back, Download, FolderOpened, Picture, Plus, Refresh, Upload, WarningFilled } from '@element-plus/icons-vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import ImageWorkbenchMaskEditor from './ImageWorkbenchMaskEditor.vue'
import {
  generateAdminImage,
  getAdminImageDetail,
  getAdminImageGenerationStatus,
  updateAdminImage,
  uploadAdminImages
} from '../api/admin'
import {
  buildImageGenerationPayload,
  createWorkbenchMaskPreview,
  createImageWorkbenchTask,
  dataUrlToFile,
  extensionForImageMime,
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
const dropActive = ref(false)
const selectedTaskError = computed(() => {
  if (selectedTask.value?.status !== 'error') return ''
  return selectedTask.value.error?.trim() || '未返回错误原因'
})
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
})

function returnToToolbox() {
  router.push('/toolbox')
}

function openMediaLibrary() {
  const href = router.resolve('/images').href
  window.open(href, '_blank', 'noopener,noreferrer')
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
  currentResults.value = await Promise.all(
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
  if (!incoming.length) return
  const merged = [...references.value.map((item) => item.file).filter(Boolean), ...incoming.map((item) => item.file).filter(Boolean)]
  const validation = validateReferenceImageFiles(merged)
  if (!validation.ok) {
    ElMessage.warning(validation.message)
    return
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
      sourceResultId: incomingItem.sourceResultId || ''
    })
  }
  references.value = [...references.value, ...nextItems]
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
  return references.value.map((item, index) => ({
    image_id: item.id,
    name: item.name,
    mime: item.mime,
    slot_index: index,
    source_kind: item.sourceKind || 'browser_input',
    source_task_id: item.sourceTaskId || '',
    source_result_id: item.sourceResultId || ''
  }))
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
  await appendReferenceItems([{
    file,
    name: file.name,
    mime: file.type,
    size: file.size,
    dataUrl: result.dataUrl,
    sourceKind: 'previous_result',
    sourceTaskId: selectedTask.value?.id || '',
    sourceResultId: result.id || ''
  }])
  ElMessage.success('已复用为参考图')
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
                <el-button :icon="Upload" @click="triggerFilePicker">添加参考图</el-button>
                <span>支持拖拽、粘贴、上传 PNG/JPEG/WebP</span>
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
                      <el-tag v-if="item.sourceKind === 'previous_result'" size="small" effect="plain">前序结果</el-tag>
                    </div>
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
              <el-button :icon="Refresh" :loading="status.loading" @click="loadStatus">刷新配置</el-button>
            </template>

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
