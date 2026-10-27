<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import CloudDownloadImportDialog from '../components/CloudDownloadImportDialog.vue'
import { getCloudDownloadImportTask, previewCloudDownloadImport } from '../api/admin'
import { cloudArchiveImportPath, cloudArchiveImportState, importDestination, isStaleActiveImport } from './cloudDownloadImport.helpers'
import { extractApiErrorMessage } from './toolbox.helpers'

const route = useRoute(), router = useRouter(), taskID = String(route.params.id)
const task = ref<any>(null), preview = ref<any>(null)
const loading = ref(true), refreshing = ref(false), waiting = ref(false), submitting = ref(false), dirty = ref(false)
const errorText = ref(''), state = ref(''), accepted = ref(false)
const retryAvailable = computed(() => waiting.value && isStaleActiveImport(task.value))
let timer: ReturnType<typeof setTimeout> | undefined
let disposed = false, navigatingToBatch = false, recheckAfterSubmit = false

async function refresh() {
  if (disposed || refreshing.value || submitting.value) return
  clearTimeout(timer); refreshing.value = true; errorText.value = ''
  try {
    if (!cloudArchiveImportPath(taskID)) throw new Error('云下载任务标识无效')
    const result = await getCloudDownloadImportTask(taskID)
    if (disposed) return
    task.value = result; state.value = cloudArchiveImportState(result)
    if (state.value === 'workspace') {
      navigatingToBatch = true
      try { await router.replace(importDestination(result)) } finally { navigatingToBatch = false }
      return
    }
    waiting.value = state.value === 'waiting'
    accepted.value = waiting.value
    if (waiting.value) dirty.value = false
    if (state.value === 'unsupported') throw new Error('该任务已按视频导入，请返回云下载列表查看')
    if (state.value === 'missing-batch') throw new Error('导入已结束，但未找到关联批次，请到压缩包批次列表查看')
    if (!waiting.value && !preview.value) {
      const file = await previewCloudDownloadImport(taskID)
      if (disposed) return
      if (file.kind !== 'archive') throw new Error('该文件不是可导入的压缩包')
      preview.value = file
    }
  } catch (error) {
    if (!disposed) errorText.value = extractApiErrorMessage(error, '读取压缩包导入状态失败')
  } finally {
    refreshing.value = false; loading.value = false
    if (!disposed && waiting.value) timer = setTimeout(() => { void refresh() }, 2000)
  }
}
function handleSubmitted(info: any) {
  accepted.value = true; dirty.value = false; waiting.value = true
  if (info && task.value) task.value.import = info
}
function handleSubmitting(value: boolean) {
  submitting.value = value
  if (!value && (waiting.value || recheckAfterSubmit)) { recheckAfterSubmit = false; void refresh() }
}
function handleSubmitError() { recheckAfterSubmit = true }
function handleClose(visible: boolean) {
  if (!visible && !accepted.value && !waiting.value) void router.push('/cloud-downloads')
}
async function prepareRetry() {
  if (refreshing.value || submitting.value) return
  clearTimeout(timer); refreshing.value = true
  try {
    const file = await previewCloudDownloadImport(taskID)
    if (disposed) return
    if (file.kind !== 'archive') throw new Error('该文件不是可导入的压缩包')
    preview.value = file; waiting.value = false; accepted.value = false; errorText.value = ''
  } catch (error) { errorText.value = extractApiErrorMessage(error, '无法重新填写导入配置') }
  finally {
    refreshing.value = false
    if (!disposed && waiting.value) timer = setTimeout(() => { void refresh() }, 2000)
  }
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value || submitting.value) { event.preventDefault(); event.returnValue = '' }
}
onBeforeRouteLeave(async () => {
  if (navigatingToBatch) return true
  if (submitting.value) { ElMessage.info('正在提交导入，请稍后离开'); return false }
  if (!dirty.value) return true
  try {
    await ElMessageBox.confirm('导入配置尚未提交，离开将丢失本地编辑。', '离开导入配置', { type: 'warning', confirmButtonText: '离开', cancelButtonText: '继续编辑' })
    return true
  } catch { return false }
})
onMounted(() => { void refresh(); window.addEventListener('beforeunload', beforeUnload) })
onUnmounted(() => { disposed = true; clearTimeout(timer); window.removeEventListener('beforeunload', beforeUnload) })
</script>

<template>
  <main class="cloud-archive-import" data-density="form">
    <header class="cloud-archive-header">
      <div><small>115 云下载 · 压缩包导入</small><h1>{{ preview?.name || task?.declared_name || '压缩包入库' }}</h1></div>
      <div class="cloud-archive-actions"><el-button :disabled="submitting" @click="router.push('/cloud-downloads')">返回云下载</el-button><el-button :loading="refreshing" :disabled="submitting || dirty" @click="refresh">刷新状态</el-button></div>
    </header>
    <el-alert v-if="errorText" :title="errorText" type="error" :closable="false" show-icon />
    <el-alert v-if="task?.import?.status === 'failed' && !waiting" :title="task.import.error_message || '上次导入失败，可调整配置后重新提交'" type="warning" :closable="false" show-icon />
    <section v-if="loading || waiting" v-loading="loading || refreshing" class="cloud-archive-status" aria-live="polite">
      <h2>{{ waiting ? '正在创建压缩包批次' : '正在读取压缩包' }}</h2>
      <p>批次就绪后自动进入全屏工作台，继续整理目录、分组与入库配置。离开页面不会中断已提交的后台任务。</p>
      <el-button v-if="retryAvailable" :disabled="refreshing" @click="prepareRetry">重新填写并提交</el-button>
    </section>
    <section v-else-if="preview && !['unsupported', 'missing-batch'].includes(state)" class="cloud-archive-form">
      <p class="cloud-archive-lead">以下内容作为主分组默认配置，解压后可在工作台继续调整。</p>
      <CloudDownloadImportDialog
        inline :model-value="true" :task="task" :preview="preview"
        @dirty="dirty = $event" @submitting="handleSubmitting" @submitted="handleSubmitted"
        @submit-error="handleSubmitError" @update:model-value="handleClose"
      />
    </section>
    <el-button v-if="state === 'missing-batch'" @click="router.push('/toolbox/archive-import')">查看压缩包批次列表</el-button>
  </main>
</template>

<style scoped>
.cloud-archive-import{min-height:100dvh;padding:24px;box-sizing:border-box;background:var(--bg-canvas);color:var(--text-primary);display:flex;flex-direction:column;gap:18px;}
.cloud-archive-header{display:flex;justify-content:space-between;align-items:center;gap:16px;}
.cloud-archive-header h1{font-size:22px;margin:6px 0 0;overflow-wrap:anywhere;}
.cloud-archive-header small,.cloud-archive-lead,.cloud-archive-status p{color:var(--text-muted);line-height:1.6;}
.cloud-archive-actions{display:flex;gap:8px;flex-shrink:0;}
.cloud-archive-form{width:100%;max-width:1000px;box-sizing:border-box;align-self:center;padding:24px;background:var(--bg-surface);border:1px solid var(--line-soft);border-radius:8px;}
.cloud-archive-lead{margin:0 0 20px;font-size:13px;}
.cloud-archive-status{padding:60px 20px;text-align:center;}
.cloud-archive-status h2{font-size:18px;}
@media(max-width:700px){.cloud-archive-import{padding:12px;}.cloud-archive-header{align-items:flex-start;flex-direction:column;}.cloud-archive-form{padding:16px;}.cloud-archive-actions{flex-wrap:wrap;}}
</style>
