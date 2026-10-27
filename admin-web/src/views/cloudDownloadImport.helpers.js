// 115 云下载任务入媒体库的纯函数集合。
// 这里只处理元数据、表单归一化和状态判定，不发起请求，便于单独测试。

export const ACTIVE_IMPORT_STATUSES = ['queued', 'hashing', 'copying', 'creating']
export const VIDEO_IMPORT_TYPES = ['short', 'movie', 'episode', 'av']

const ACTIVE_IMPORT_STATUS_SET = new Set(ACTIVE_IMPORT_STATUSES)
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

function toSelectionArray(values) {
  if (Array.isArray(values)) return values
  if (typeof values === 'string') return values.split(/[,，\n]/)
  return []
}

export function normalizeImportTags(values) {
  const out = []
  const seen = new Set()
  for (const raw of toSelectionArray(values)) {
    const tag = String(raw ?? '')
      .trim()
      .replace(/\s+/g, ' ')
      .toLowerCase()
    if (!tag || seen.has(tag)) continue
    seen.add(tag)
    out.push(tag)
  }
  return out
}

export function normalizeImportUUIDs(values) {
  const out = []
  const seen = new Set()
  for (const raw of toSelectionArray(values)) {
    const value = String(raw ?? '').trim()
    if (!value || !UUID_PATTERN.test(value)) continue
    const key = value.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
  }
  return out
}

export function splitImportActors(values) {
  const actorIDs = []
  const actorNames = []
  const seenID = new Set()
  const seenName = new Set()
  for (const raw of toSelectionArray(values)) {
    const value = String(raw ?? '').trim()
    if (!value) continue
    if (UUID_PATTERN.test(value)) {
      const key = value.toLowerCase()
      if (seenID.has(key)) continue
      seenID.add(key)
      actorIDs.push(value)
      continue
    }
    const name = value.replace(/\s+/g, ' ')
    const key = name.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(name)
  }
  return { actorIDs, actorNames }
}

export function basenameWithoutExtension(filePath) {
  const normalized = String(filePath ?? '').replace(/\\/g, '/').replace(/\/+$/, '')
  if (!normalized) return ''
  const base = normalized.slice(normalized.lastIndexOf('/') + 1)
  const dot = base.lastIndexOf('.')
  if (dot <= 0) return base
  return base.slice(0, dot)
}

export function resolveCloudImportDefaultTitle(task, preview) {
  return basenameWithoutExtension(task?.local_path || preview?.name || task?.declared_name || '')
}

export function formatImportFileSize(bytes) {
  const size = Number(bytes)
  if (!Number.isFinite(size) || size <= 0) return '--'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 100 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

export function buildVideoImportPayload(form = {}) {
  const type = VIDEO_IMPORT_TYPES.includes(String(form.type)) ? String(form.type) : 'short'
  const { actorIDs, actorNames } = splitImportActors(form.actors)
  return {
    kind: 'video',
    type,
    site_category: type === 'av'
      ? String(form.site_category || '').trim() || 'japanese'
      : '',
    title: String(form.title || '').trim(),
    description: String(form.description || '').trim(),
    tags: normalizeImportTags(form.tags),
    actor_ids: actorIDs,
    actor_names: actorNames,
    collection_ids: type === 'short' ? normalizeImportUUIDs(form.collections) : [],
    image_collection_id: normalizeImportUUIDs([form.image_collection_id])[0] || null
  }
}

export function buildArchiveImportPayload(form = {}) {
  const hasPassword = Boolean(form.has_password)
  return {
    kind: 'archive',
    title: String(form.title || '').trim(),
    default_description: String(form.default_description || '').trim(),
    default_tags: normalizeImportTags(form.default_tags),
    default_video_collection_ids: normalizeImportUUIDs(form.default_video_collection_ids),
    default_image_collection_ids: normalizeImportUUIDs(form.default_image_collection_ids),
    has_password: hasPassword,
    password: hasPassword ? String(form.password || '').trim() : ''
  }
}

export function validateVideoImportForm(form = {}) {
  if (!VIDEO_IMPORT_TYPES.includes(String(form.type || ''))) {
    return '请选择视频类型'
  }
  return ''
}

export function validateArchiveImportForm(form = {}) {
  if (form.has_password && !String(form.password || '').trim()) {
    return '请输入压缩包密码'
  }
  return ''
}

export const STALE_IMPORT_MS = 10 * 60 * 1000

export function isImportActive(task) {
  return ACTIVE_IMPORT_STATUS_SET.has(String(task?.import?.status || ''))
}

export function isStaleActiveImport(task, now = Date.now()) {
  if (!isImportActive(task)) return false
  const createdAt = Date.parse(task?.import?.created_at || '')
  if (!Number.isFinite(createdAt)) return false
  return now - createdAt >= STALE_IMPORT_MS
}

export function hasActiveImportTasks(tasks) {
  return Array.isArray(tasks) && tasks.some((task) => isImportActive(task))
}

export function resolveCloudImportAction(task, now = Date.now()) {
  if (!task?.import) {
    if (task?.status === 'completed') {
      return String(task?.local_path || '').trim() ? 'import' : 'missing'
    }
    if (task?.status === 'failed' || task?.status === 'blocked') return 'retry-download'
    return 'none'
  }
  const status = String(task.import.status || '')
  if (ACTIVE_IMPORT_STATUS_SET.has(status)) {
    // A queued/hashing record with no live asynq task would otherwise stay
    // "导入中" forever. After the grace period allow an explicit re-submit;
    // the backend reuses the same attempt/TaskID so a live task is not
    // duplicated.
    return isStaleActiveImport(task, now) ? 'retry-import' : 'importing'
  }
  if (status === 'imported') return 'imported'
  if (status === 'skipped_existing') return 'existing'
  if (status === 'failed') return 'retry-import'
  return 'none'
}

// 导入批次一旦创建即可进入工作台，排队、失败及源文件移除均不影响定位。
export function importDestination(task) {
  const info = task?.import
  const batchID = String(info?.archive_batch_id || '').trim()
  if (info?.kind === 'archive' && UUID_PATTERN.test(batchID)) return `/toolbox/archive-import/${batchID}`
  const action = resolveCloudImportAction(task)
  if (action === 'imported') return info?.kind === 'archive' ? '/toolbox/archive-import' : '/videos'
  if (action === 'existing') return '/videos'
  return ''
}

export function cloudArchiveImportPath(taskID) {
  const id = String(taskID || '').trim()
  return /^[1-9]\d*$/.test(id) ? `/cloud-downloads/${id}/archive-import` : ''
}

export function cloudArchiveImportState(task) {
  const info = task?.import
  if (info?.kind && info.kind !== 'archive') return 'unsupported'
  if (info?.archive_batch_id && importDestination(task).startsWith('/toolbox/archive-import/')) return 'workspace'
  if (isImportActive(task)) return 'waiting'
  if (info?.status === 'imported') return 'missing-batch'
  return 'form'
}
