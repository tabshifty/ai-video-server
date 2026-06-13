export const IMAGE_WORKBENCH_LIMITS = {
  maxImages: 4,
  maxReferenceImages: 4,
  maxReferenceImageBytes: 10 * 1024 * 1024,
  maxReferencePayloadBytes: 40 * 1024 * 1024,
  allowedMimeTypes: ['image/png', 'image/jpeg', 'image/webp']
}

export const DEFAULT_IMAGE_WORKBENCH_PARAMS = {
  size: 'auto',
  quality: 'auto',
  output_format: 'png',
  output_compression: null,
  n: 1
}

export const IMAGE_WORKBENCH_PREFERENCE_KEY = 'admin-image-workbench-preferences'

export function normalizeImageWorkbenchParams(input = {}) {
  const params = {
    ...DEFAULT_IMAGE_WORKBENCH_PARAMS,
    ...input
  }

  if (!['auto', '1024x1024', '1024x1536', '1536x1024'].includes(params.size)) {
    params.size = DEFAULT_IMAGE_WORKBENCH_PARAMS.size
  }
  if (!['auto', 'low', 'medium', 'high'].includes(params.quality)) {
    params.quality = DEFAULT_IMAGE_WORKBENCH_PARAMS.quality
  }
  if (!['png', 'jpeg', 'webp'].includes(params.output_format)) {
    params.output_format = DEFAULT_IMAGE_WORKBENCH_PARAMS.output_format
  }
  const n = Number(params.n)
  params.n = Number.isFinite(n) ? Math.min(Math.max(Math.trunc(n), 1), IMAGE_WORKBENCH_LIMITS.maxImages) : 1
  if (params.output_format === 'png') {
    params.output_compression = null
  } else if (params.output_compression !== null && params.output_compression !== undefined && params.output_compression !== '') {
    const compression = Number(params.output_compression)
    params.output_compression = Number.isFinite(compression) ? Math.min(Math.max(Math.trunc(compression), 0), 100) : null
  } else {
    params.output_compression = 82
  }
  return params
}

export function extensionForImageMime(mime) {
  if (mime === 'image/jpeg') return 'jpg'
  if (mime === 'image/webp') return 'webp'
  return 'png'
}

export function formatWorkbenchFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`
  return `${(value / 1024 / 1024).toFixed(1)} MiB`
}

export function validateReferenceImageFiles(files, limits = IMAGE_WORKBENCH_LIMITS) {
  const list = Array.from(files || [])
  if (list.length > limits.maxReferenceImages) {
    return { ok: false, message: `参考图最多 ${limits.maxReferenceImages} 张` }
  }
  let total = 0
  for (const file of list) {
    if (!limits.allowedMimeTypes.includes(file.type)) {
      return { ok: false, message: '仅支持 PNG、JPEG、WebP 参考图' }
    }
    if (file.size > limits.maxReferenceImageBytes) {
      return { ok: false, message: `单张参考图最多 ${formatWorkbenchFileSize(limits.maxReferenceImageBytes)}` }
    }
    total += file.size
    if (total > limits.maxReferencePayloadBytes) {
      return { ok: false, message: `参考图总大小最多 ${formatWorkbenchFileSize(limits.maxReferencePayloadBytes)}` }
    }
  }
  return { ok: true, message: '' }
}

export function dataUrlToFile(dataUrl, filename, fallbackType = 'image/png') {
  const [header, payload = ''] = String(dataUrl || '').split(',')
  const mimeMatch = header.match(/^data:([^;]+);base64$/)
  const mime = mimeMatch?.[1] || fallbackType
  const binary = atob(payload)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  return new File([bytes], filename, { type: mime })
}

export async function loadWorkbenchImage(dataUrl) {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片加载失败'))
    image.src = String(dataUrl || '')
  })
}

export async function ensureWorkbenchImagePng(dataUrl) {
  const normalized = String(dataUrl || '').trim()
  if (/^data:image\/png(?:[;,]|$)/i.test(normalized)) {
    return normalized
  }
  const image = await loadWorkbenchImage(normalized)
  const canvas = document.createElement('canvas')
  canvas.width = image.naturalWidth || 0
  canvas.height = image.naturalHeight || 0
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  ctx.drawImage(image, 0, 0, canvas.width, canvas.height)
  return canvas.toDataURL('image/png')
}

export async function validateWorkbenchMask(maskDataUrl, imageDataUrl) {
  const [maskImage, targetImage] = await Promise.all([
    loadWorkbenchImage(maskDataUrl),
    loadWorkbenchImage(imageDataUrl)
  ])
  if (
    maskImage.naturalWidth !== targetImage.naturalWidth ||
    maskImage.naturalHeight !== targetImage.naturalHeight
  ) {
    throw new Error('遮罩尺寸与目标参考图不一致，请重新绘制')
  }
  const canvas = document.createElement('canvas')
  canvas.width = maskImage.naturalWidth
  canvas.height = maskImage.naturalHeight
  const ctx = canvas.getContext('2d', { willReadFrequently: true })
  if (!ctx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  ctx.drawImage(maskImage, 0, 0)
  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
  let edited = 0
  let fullyTransparent = 0
  const total = imageData.data.length / 4
  for (let index = 3; index < imageData.data.length; index += 4) {
    if (imageData.data[index] < 255) edited += 1
    if (imageData.data[index] === 0) fullyTransparent += 1
  }
  if (edited === 0) return 'empty'
  if (fullyTransparent === total) return 'full'
  return 'partial'
}

export async function createWorkbenchMaskPreview(imageDataUrl, maskDataUrl) {
  const [image, mask] = await Promise.all([
    loadWorkbenchImage(imageDataUrl),
    loadWorkbenchImage(maskDataUrl)
  ])
  if (
    image.naturalWidth !== mask.naturalWidth ||
    image.naturalHeight !== mask.naturalHeight
  ) {
    throw new Error('遮罩尺寸与目标参考图不一致，请重新绘制')
  }
  const previewCanvas = document.createElement('canvas')
  previewCanvas.width = image.naturalWidth
  previewCanvas.height = image.naturalHeight
  const previewCtx = previewCanvas.getContext('2d')
  if (!previewCtx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  previewCtx.drawImage(image, 0, 0)

  const maskCanvas = document.createElement('canvas')
  maskCanvas.width = mask.naturalWidth
  maskCanvas.height = mask.naturalHeight
  const maskCtx = maskCanvas.getContext('2d')
  if (!maskCtx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  maskCtx.drawImage(mask, 0, 0)

  const overlayCanvas = document.createElement('canvas')
  overlayCanvas.width = previewCanvas.width
  overlayCanvas.height = previewCanvas.height
  const overlayCtx = overlayCanvas.getContext('2d')
  if (!overlayCtx) {
    throw new Error('当前浏览器不支持 Canvas')
  }
  overlayCtx.fillStyle = '#3b82f6'
  overlayCtx.fillRect(0, 0, overlayCanvas.width, overlayCanvas.height)
  overlayCtx.globalCompositeOperation = 'destination-out'
  overlayCtx.drawImage(maskCanvas, 0, 0)
  overlayCtx.globalCompositeOperation = 'source-over'

  previewCtx.globalAlpha = 0.58
  previewCtx.drawImage(overlayCanvas, 0, 0)
  previewCtx.globalAlpha = 1
  return previewCanvas.toDataURL('image/png')
}

function replaceImageFilenameExtension(name, nextExtension) {
  const normalized = String(name || '').trim()
  if (!normalized) return `reference.${nextExtension}`
  if (normalized.includes('.')) {
    return normalized.replace(/\.[^.]+$/, `.${nextExtension}`)
  }
  return `${normalized}.${nextExtension}`
}

export function createImageWorkbenchTask({ prompt, params, referenceImageIds, referenceSnapshots, outputImageIds, results, status, error, startedAt, finishedAt, mask }) {
  const createdAt = startedAt || Date.now()
  return {
    id: `img-task-${createdAt}-${Math.random().toString(36).slice(2, 8)}`,
    prompt: String(prompt || '').trim(),
    params: normalizeImageWorkbenchParams(params),
    referenceImageIds: referenceImageIds || [],
    referenceSnapshots: referenceSnapshots || [],
    outputImageIds: outputImageIds || [],
    results: results || [],
    mask: mask || null,
    status: status || 'done',
    error: error || '',
    createdAt,
    finishedAt: finishedAt || null
  }
}

export function buildReferenceImageSnapshots(referenceImages = []) {
  return (referenceImages || []).map((item, index) => {
    const snapshot = {
      image_id: item.id,
      name: item.name,
      mime: item.mime,
      slot_index: index,
      source_kind: item.sourceKind || 'browser_input',
      source_task_id: item.sourceTaskId || '',
      source_result_id: item.sourceResultId || ''
    }
    if (snapshot.source_kind === 'library_asset') {
      snapshot.source_image_id = item.sourceImageId || ''
      snapshot.source_title = item.sourceTitle || ''
      snapshot.source_status = item.sourceStatus || ''
      snapshot.source_active = item.sourceActive === undefined ? null : Boolean(item.sourceActive)
      snapshot.source_view_url = item.sourceViewUrl || ''
      snapshot.source_frozen_at = item.sourceFrozenAt || null
    }
    return snapshot
  })
}

export async function buildImageGenerationPayload(prompt, params, referenceImages, maskDraft = null) {
  const normalized = normalizeImageWorkbenchParams(params)
  const payloadReferences = (referenceImages || []).map((item, index) => ({
    name: item.name || `reference-${index + 1}.${extensionForImageMime(item.mime)}`,
    mime: item.mime,
    data_url: item.dataUrl
  }))
  let payloadMask = null
  if (maskDraft?.targetImageId && maskDraft?.maskDataUrl) {
    const targetIndex = (referenceImages || []).findIndex((item) => item.id === maskDraft.targetImageId)
    if (targetIndex < 0) {
      throw new Error('遮罩目标参考图已不存在，请重新选择')
    }
    const target = referenceImages[targetIndex]
    payloadReferences[targetIndex] = {
      ...payloadReferences[targetIndex],
      name: replaceImageFilenameExtension(target?.name, 'png'),
      mime: 'image/png',
      data_url: await ensureWorkbenchImagePng(target?.dataUrl)
    }
    payloadMask = {
      name: 'mask.png',
      mime: 'image/png',
      data_url: maskDraft.maskDataUrl,
      target_index: targetIndex
    }
  }
  return {
    prompt: String(prompt || '').trim(),
    ...normalized,
    reference_images: payloadReferences,
    ...(payloadMask ? { mask: payloadMask } : {})
  }
}

export function readWorkbenchPreferences(storage = globalThis.localStorage) {
  try {
    const raw = storage?.getItem(IMAGE_WORKBENCH_PREFERENCE_KEY)
    return normalizeImageWorkbenchParams(raw ? JSON.parse(raw) : {})
  } catch (_) {
    return { ...DEFAULT_IMAGE_WORKBENCH_PARAMS }
  }
}

export function writeWorkbenchPreferences(params, storage = globalThis.localStorage) {
  try {
    storage?.setItem(IMAGE_WORKBENCH_PREFERENCE_KEY, JSON.stringify(normalizeImageWorkbenchParams(params)))
  } catch (_) {
    // 偏好写入失败时保留当前会话状态。
  }
}
