const CLOUD_FILE_DELETION_STALE_MS = 5 * 60 * 1000
const CLOUD_FILE_DELETION_LABELS = {
  deleting: { label: '移入回收站中', type: 'warning' },
  deleted: { label: '已移入回收站', type: 'success' },
  failed: { label: '移入失败', type: 'danger' }
}
const TERMINAL_IMPORT_STATUSES = new Set(['imported', 'skipped_existing'])

function isStaleDeleting(row) {
  if (row?.file_deletion?.status !== 'deleting') return false
  const updatedAt = Date.parse(row.file_deletion.updated_at || '')
  return Number.isFinite(updatedAt) && Date.now() - updatedAt >= CLOUD_FILE_DELETION_STALE_MS
}

export function cloudFileDeletionMeta(row) {
  const status = String(row?.file_deletion?.status || '')
  if (status === 'deleting' && isStaleDeleting(row)) {
    return {
      label: '移入超时',
      type: 'warning',
      detail: '上次移入请求未完成，请先在 115 回收站确认后重试'
    }
  }
  const meta = CLOUD_FILE_DELETION_LABELS[status]
  if (!meta) return null
  const detail = status === 'failed'
    ? String(row?.file_deletion?.error_message || '').trim()
    : String(row?.file_deletion?.note || '').trim()
  return {
    ...meta,
    detail
  }
}

export function canDeleteCloudDownloadFile(row) {
  if (!row || row.status !== 'completed') return false
  const hasLocalFile = Boolean(String(row.local_path || '').trim())
  const imported = TERMINAL_IMPORT_STATUSES.has(String(row?.import?.status || ''))
  if (!hasLocalFile && !imported) return false
  const status = String(row?.file_deletion?.status || '')
  if (status === 'deleted') return false
  if (status === 'deleting') return isStaleDeleting(row)
  return true
}

export function cloudFileDeletionActionLabel(row) {
  const status = String(row?.file_deletion?.status || '')
  return status === 'failed' || (status === 'deleting' && isStaleDeleting(row))
    ? '重试移入回收站'
    : '移入 115 回收站'
}

export function hasActiveCloudFileDeletions(tasks) {
  return (tasks || []).some((task) => task?.file_deletion?.status === 'deleting')
}

export function cloudDownloadFileName(row) {
  const declared = String(row?.declared_name || '').trim()
  if (declared) return declared
  const localPath = String(row?.local_path || '').trim()
  if (!localPath) return '--'
  const segments = localPath.split(/[\\/]/).filter(Boolean)
  return segments[segments.length - 1] || '--'
}
