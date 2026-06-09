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

export function createImageWorkbenchTask({ prompt, params, referenceImageIds, outputImageIds, results, status, error, startedAt, finishedAt }) {
  const createdAt = startedAt || Date.now()
  return {
    id: `img-task-${createdAt}-${Math.random().toString(36).slice(2, 8)}`,
    prompt: String(prompt || '').trim(),
    params: normalizeImageWorkbenchParams(params),
    referenceImageIds: referenceImageIds || [],
    outputImageIds: outputImageIds || [],
    results: results || [],
    status: status || 'done',
    error: error || '',
    createdAt,
    finishedAt: finishedAt || null
  }
}

export function buildImageGenerationPayload(prompt, params, referenceImages) {
  const normalized = normalizeImageWorkbenchParams(params)
  return {
    prompt: String(prompt || '').trim(),
    ...normalized,
    reference_images: (referenceImages || []).map((item) => ({
      name: item.name,
      mime: item.mime,
      data_url: item.dataUrl
    }))
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
