<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getVideoRotation, previewVideoRotation, submitVideoRotation } from '../api/admin'

const props = defineProps<{ video: { id: string; status: string; metadata?: Record<string, any> } }>()
const emit = defineEmits<{ completed: [] }>()
const job = ref<any>(null)
const visible = ref(false)
const angle = ref(90)
const mode = ref('auto')
const plan = ref<any>(null)
const busy = ref(false)
const error = ref('')
const requestID = ref('')
let timer: ReturnType<typeof setTimeout> | undefined
let generation = 0
let disposed = false
let completedID = ''
const active = computed(() => ['queued', 'running', 'cleanup'].includes(job.value?.status))
const blockedReason = computed(() => {
  if (props.video.status !== 'ready') return '仅已就绪视频可修正方向'
  const metadata = props.video.metadata || {}
  if (metadata.playback_compat?.dolby_vision_risk || metadata.playback_compat?.source?.dolby_vision || metadata.source_playback_path) return '杜比视界源暂不支持方向修正'
  return ''
})
const jobLabel = computed(() => ({ queued: '等待处理', running: '正在修正方向', cleanup: '画面已修正，正在清理旧文件', succeeded: '方向修正完成，旧文件已清理', failed: '方向修正失败，旧视频已保留' })[job.value?.status as string] || '')

async function poll(expected = generation) {
  const id = props.video.id
  try {
    const data = await getVideoRotation(id)
    if (disposed || expected !== generation || id !== props.video.id) return
    job.value = data.job
    error.value = ''
    if (job.value?.status === 'succeeded' && job.value.id !== completedID) {
      const wasProcessing = completedID !== '' || activeBeforeRefresh
      completedID = job.value.id
      if (wasProcessing) emit('completed')
    }
    activeBeforeRefresh = active.value
  } catch {
    if (disposed || expected !== generation) return
    error.value = '暂时无法刷新任务状态，可点击刷新重试。'
  }
  if (!disposed && expected === generation && active.value) timer = setTimeout(() => poll(expected), 3000)
}
let activeBeforeRefresh = false
watch(() => props.video.id, () => {
  generation++
  clearTimeout(timer)
  job.value = null
  plan.value = null
  visible.value = false
  error.value = ''
  completedID = ''
  activeBeforeRefresh = false
  poll()
}, { immediate: true })
watch(mode, () => { plan.value = null; requestID.value = '' })

function refreshJob() { clearTimeout(timer); poll() }
function open() {
  angle.value = 90
  mode.value = 'auto'
  plan.value = null
  requestID.value = ''
  error.value = ''
  visible.value = true
}
async function inspect() {
  busy.value = true
  error.value = ''
  const expected = generation
  try {
    const data = await previewVideoRotation(props.video.id, mode.value)
    if (disposed || expected !== generation) return
    plan.value = data
    requestID.value = crypto.randomUUID?.() || "10000000-1000-4000-8000-100000000000".replace(/[018]/g, (digit) => (Number(digit) ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> Number(digit) / 4).toString(16))
  } catch (e: any) {
    if (expected === generation) error.value = e.message || '预检失败'
  } finally { busy.value = false }
}
async function submit() {
  if (!plan.value || busy.value) return
  busy.value = true
  const expected = generation
  const videoID = props.video.id
  // Pin exactly what the administrator saw; changing the form cannot change consent.
  const payload = { request_id: requestID.value, angle: angle.value, mode: plan.value.mode, version: plan.value.version, reencode_confirmed: plan.value.mode === 'reencode' }
  try {
    await ElMessageBox.confirm(
      `${payload.mode === 'reencode' ? '本次需要重新编码，可能损失画质且处理较慢。' : '本次修改旋转标记，不重新编码；播放器需支持方向信息。'}成功后删除旧文件，不保留恢复副本。处理期间仍可播放旧视频。`,
      '确认修正方向', { type: 'warning', confirmButtonText: '确认修正', cancelButtonText: '取消' }
    )
    if (disposed || expected !== generation) return
    const data = await submitVideoRotation(videoID, payload)
    if (disposed || expected !== generation) return
    job.value = data
    activeBeforeRefresh = true
    visible.value = false
    ElMessage.success('方向修正任务已受理')
    clearTimeout(timer)
    poll(expected)
  } catch (e: any) {
    if (expected === generation && e !== 'cancel' && e !== 'close') error.value = e.message || '提交失败；可重试同一请求'
  } finally { busy.value = false }
}
onBeforeUnmount(() => { disposed = true; generation++; clearTimeout(timer) })
</script>

<template>
  <div class="video-rotation-panel">
    <div class="rotation-actions">
      <el-button plain :disabled="!!blockedReason || active" @click="open">修正方向</el-button>
      <el-button v-if="job || error" link @click="refreshJob">刷新修正状态</el-button>
      <span v-if="blockedReason">{{ blockedReason }}</span>
    </div>
    <p v-if="job" role="status">{{ jobLabel }}<span v-if="job.status === 'running'"> · {{ Math.round(job.progress || 0) }}%</span></p>
    <el-alert v-if="job?.error_message" :title="job.error_message" type="warning" :closable="false" />
    <el-alert v-if="error && !visible" :title="error" type="error" :closable="false" />
    <el-dialog v-model="visible" title="修正视频方向" width="min(520px, 94vw)" append-to-body :close-on-click-modal="!busy" :close-on-press-escape="!busy" :show-close="!busy">
      <el-form label-position="top" :disabled="busy">
        <el-form-item label="相对当前画面的旋转角度">
          <el-select v-model="angle" style="width:100%">
            <el-option label="顺时针 90°" :value="90" />
            <el-option label="180°" :value="180" />
            <el-option label="逆时针 90°" :value="270" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理方式">
          <el-select v-model="mode" style="width:100%">
            <el-option label="优先无损旋转（推荐）" value="auto" />
            <el-option label="重新编码（解决播放器不识别方向）" value="reencode" />
          </el-select>
        </el-form-item>
      </el-form>
      <p>自动截图封面随方向更新，自定义海报保留。成功后不保留旧视频，处理期间需要临时磁盘空间。</p>
      <el-alert v-if="plan" :type="plan.mode === 'reencode' ? 'warning' : 'info'" :closable="false" :title="plan.reason" />
      <p v-if="plan">预估需预留临时空间：{{ (plan.temporary_bytes / 1024 / 1024).toFixed(1) }} MB</p>
      <el-alert v-if="error" :title="error" type="error" :closable="false" />
      <template #footer>
        <el-button :disabled="busy" @click="visible = false">取消</el-button>
        <el-button :loading="busy" @click="inspect">检查处理方式</el-button>
        <el-button type="primary" :disabled="!plan || busy" @click="submit">提交修正</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.video-rotation-panel { width: 100%; }
.rotation-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
p { line-height: 1.6; color: var(--el-text-color-secondary); }
</style>
