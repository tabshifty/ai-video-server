<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  imageDataUrl: {
    type: String,
    default: ''
  },
  imageName: {
    type: String,
    default: ''
  },
  initialMaskDataUrl: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'save'])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const previewCanvasRef = ref(null)
const loading = ref(false)
const tool = ref('brush')
const brushSize = ref(72)
const isDrawing = ref(false)

let baseImage = null
let maskCanvas = null
let overlayCanvas = null
let lastPoint = null

watch(
  () => [props.modelValue, props.imageDataUrl, props.initialMaskDataUrl],
  ([isVisible]) => {
    if (isVisible) {
      void initializeMaskEditor()
    }
  }
)

onBeforeUnmount(() => {
  baseImage = null
  maskCanvas = null
  overlayCanvas = null
  lastPoint = null
})

function ensureMaskCanvas(width, height) {
  if (!maskCanvas) {
    maskCanvas = document.createElement('canvas')
  }
  maskCanvas.width = width
  maskCanvas.height = height
  const ctx = maskCanvas.getContext('2d')
  if (!ctx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  return ctx
}

function ensureOverlayCanvas(width, height) {
  if (!overlayCanvas) {
    overlayCanvas = document.createElement('canvas')
  }
  overlayCanvas.width = width
  overlayCanvas.height = height
  const ctx = overlayCanvas.getContext('2d')
  if (!ctx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  return ctx
}

function fillWhiteMask() {
  if (!maskCanvas) return
  const ctx = maskCanvas.getContext('2d')
  if (!ctx) return
  ctx.globalCompositeOperation = 'source-over'
  ctx.clearRect(0, 0, maskCanvas.width, maskCanvas.height)
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, maskCanvas.width, maskCanvas.height)
}

function loadDataUrlImage(dataUrl) {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片加载失败'))
    image.src = String(dataUrl || '')
  })
}

async function initializeMaskEditor() {
  if (!props.imageDataUrl) return
  loading.value = true
  try {
    baseImage = await loadDataUrlImage(props.imageDataUrl)
    ensureMaskCanvas(baseImage.naturalWidth, baseImage.naturalHeight)
    ensureOverlayCanvas(baseImage.naturalWidth, baseImage.naturalHeight)
    fillWhiteMask()
    if (props.initialMaskDataUrl) {
      const existingMask = await loadDataUrlImage(props.initialMaskDataUrl)
      if (
        existingMask.naturalWidth === baseImage.naturalWidth &&
        existingMask.naturalHeight === baseImage.naturalHeight
      ) {
        const ctx = maskCanvas.getContext('2d')
        if (ctx) {
          ctx.clearRect(0, 0, maskCanvas.width, maskCanvas.height)
          ctx.drawImage(existingMask, 0, 0, maskCanvas.width, maskCanvas.height)
        }
      }
    }
    drawPreview()
  } finally {
    loading.value = false
  }
}

function drawPreview() {
  const canvas = previewCanvasRef.value
  if (!canvas || !baseImage || !maskCanvas) return
  canvas.width = baseImage.naturalWidth
  canvas.height = baseImage.naturalHeight
  const ctx = canvas.getContext('2d')
  const overlayCtx = ensureOverlayCanvas(canvas.width, canvas.height)
  if (!ctx) return

  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.drawImage(baseImage, 0, 0, canvas.width, canvas.height)

  overlayCtx.clearRect(0, 0, overlayCanvas.width, overlayCanvas.height)
  overlayCtx.fillStyle = '#3b82f6'
  overlayCtx.fillRect(0, 0, overlayCanvas.width, overlayCanvas.height)
  overlayCtx.globalCompositeOperation = 'destination-out'
  overlayCtx.drawImage(maskCanvas, 0, 0)
  overlayCtx.globalCompositeOperation = 'source-over'

  ctx.globalAlpha = 0.58
  ctx.drawImage(overlayCanvas, 0, 0)
  ctx.globalAlpha = 1
}

function toCanvasPoint(event) {
  const canvas = previewCanvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  return {
    x: ((event.clientX - rect.left) / rect.width) * canvas.width,
    y: ((event.clientY - rect.top) / rect.height) * canvas.height
  }
}

function paintDot(ctx, point) {
  ctx.beginPath()
  ctx.arc(point.x, point.y, Math.max(1, brushSize.value / 2), 0, Math.PI * 2)
  ctx.fill()
}

function drawSegment(from, to) {
  if (!maskCanvas) return
  const ctx = maskCanvas.getContext('2d')
  if (!ctx) return
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.lineWidth = brushSize.value
  if (tool.value === 'brush') {
    ctx.globalCompositeOperation = 'destination-out'
    ctx.strokeStyle = 'rgba(0, 0, 0, 1)'
    ctx.fillStyle = 'rgba(0, 0, 0, 1)'
  } else {
    ctx.globalCompositeOperation = 'source-over'
    ctx.strokeStyle = '#ffffff'
    ctx.fillStyle = '#ffffff'
  }
  if (!from || !to) return
  ctx.beginPath()
  ctx.moveTo(from.x, from.y)
  ctx.lineTo(to.x, to.y)
  ctx.stroke()
  paintDot(ctx, to)
}

function handlePointerDown(event) {
  if (loading.value) return
  const point = toCanvasPoint(event)
  lastPoint = point
  isDrawing.value = true
  event.target?.setPointerCapture?.(event.pointerId)
  drawSegment(point, point)
  drawPreview()
}

function handlePointerMove(event) {
  if (!isDrawing.value || !lastPoint) return
  const point = toCanvasPoint(event)
  drawSegment(lastPoint, point)
  lastPoint = point
  drawPreview()
}

function stopDrawing(event) {
  if (event?.target?.releasePointerCapture) {
    try {
      event.target.releasePointerCapture(event.pointerId)
    } catch (_) {
      // Ignore release errors from browsers that already cleared capture.
    }
  }
  isDrawing.value = false
  lastPoint = null
}

function clearMask() {
  fillWhiteMask()
  drawPreview()
}

function saveMask() {
  if (!maskCanvas) return
  emit('save', maskCanvas.toDataURL('image/png'))
}
</script>

<template>
  <el-dialog
    v-model="visible"
    class="mask-editor"
    title="编辑局部蒙版"
    width="1080px"
    top="4vh"
    append-to-body
    destroy-on-close
    :close-on-click-modal="false"
  >
    <div class="mask-editor__layout">
      <div class="mask-editor__toolbar">
        <div class="mask-editor__toolbar-group">
          <strong>{{ imageName || '目标参考图' }}</strong>
          <span>蓝色区域表示会被重绘；橡皮用于恢复保留区域。</span>
        </div>
        <div class="mask-editor__toolbar-group">
          <el-button-group>
            <el-button :type="tool === 'brush' ? 'primary' : 'default'" @click="tool = 'brush'">涂抹区域</el-button>
            <el-button :type="tool === 'eraser' ? 'primary' : 'default'" @click="tool = 'eraser'">恢复保留</el-button>
          </el-button-group>
        </div>
        <div class="mask-editor__toolbar-group mask-editor__toolbar-group--slider">
          <span>笔刷 {{ brushSize }} px</span>
          <el-slider v-model="brushSize" :min="16" :max="180" />
        </div>
        <el-button @click="clearMask">清空蒙版</el-button>
      </div>

      <div class="mask-editor__stage" v-loading="loading">
        <canvas
          ref="previewCanvasRef"
          class="mask-editor__canvas"
          @pointerdown.prevent="handlePointerDown"
          @pointermove.prevent="handlePointerMove"
          @pointerup.prevent="stopDrawing"
          @pointerleave.prevent="stopDrawing"
          @pointercancel.prevent="stopDrawing"
        />
      </div>
    </div>

    <template #footer>
      <div class="mask-editor__footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="saveMask">保存蒙版</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.mask-editor :deep(.el-dialog) {
  max-width: min(96vw, 1080px);
}

.mask-editor__layout {
  display: grid;
  gap: var(--space-4);
}

.mask-editor__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.mask-editor__toolbar-group {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.mask-editor__toolbar-group strong,
.mask-editor__toolbar-group span {
  line-height: var(--leading-small);
}

.mask-editor__toolbar-group span {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.mask-editor__toolbar-group--slider {
  min-width: min(18rem, 100%);
  flex: 1 1 18rem;
}

.mask-editor__stage {
  display: grid;
  place-items: center;
  min-height: 20rem;
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background:
    linear-gradient(45deg, rgba(15, 23, 42, 0.04) 25%, transparent 25%, transparent 75%, rgba(15, 23, 42, 0.04) 75%),
    linear-gradient(45deg, rgba(15, 23, 42, 0.04) 25%, transparent 25%, transparent 75%, rgba(15, 23, 42, 0.04) 75%);
  background-position: 0 0, 12px 12px;
  background-size: 24px 24px;
}

.mask-editor__canvas {
  display: block;
  max-width: 100%;
  max-height: min(72vh, 48rem);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  touch-action: none;
  cursor: crosshair;
}

.mask-editor__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 48rem) {
  .mask-editor__toolbar {
    align-items: stretch;
  }

  .mask-editor__footer {
    width: 100%;
  }

  .mask-editor__footer .el-button {
    flex: 1;
  }
}
</style>
