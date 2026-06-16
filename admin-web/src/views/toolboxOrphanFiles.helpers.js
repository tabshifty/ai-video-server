import { formatAdminDateTime } from '../utils/dateTime'

const ORPHAN_SCAN_STATUS_LABELS = {
  idle: '未扫描',
  pending: '排队中',
  running: '扫描中',
  completed: '已完成',
  failed: '失败',
  deleted: '已删除'
}

const ORPHAN_SCAN_STATUS_TYPES = {
  idle: 'info',
  pending: 'warning',
  running: 'warning',
  completed: 'success',
  failed: 'danger',
  deleted: 'info'
}

export function getOrphanScanStatusLabel(status) {
  return ORPHAN_SCAN_STATUS_LABELS[status] || '未知'
}

export function getOrphanScanStatusType(status) {
  return ORPHAN_SCAN_STATUS_TYPES[status] || 'info'
}

export function formatOrphanScanTime(value) {
  return formatAdminDateTime(value, '--')
}

export function formatOrphanScanBytes(bytes) {
  const value = Number(bytes)
  if (!Number.isFinite(value) || value < 0) {
    return '--'
  }
  if (value < 1024) {
    return `${value} B`
  }
  const units = ['KB', 'MB', 'GB', 'TB']
  let current = value / 1024
  let unitIndex = 0
  while (current >= 1024 && unitIndex < units.length - 1) {
    current /= 1024
    unitIndex += 1
  }
  return `${current.toFixed(current >= 100 ? 0 : 1)} ${units[unitIndex]}`
}

export function buildOrphanScanPromptKey(scan) {
  if (!scan || scan.status !== 'completed') {
    return ''
  }
  const scanId = Number(scan.id || 0)
  const orphanFiles = Number(scan.orphan_files || 0)
  const deletedFiles = Number(scan.deleted_files || 0)
  const revision = String(scan.finished_at || scan.updated_at || '')
  return [scanId, revision, orphanFiles, deletedFiles].join(':')
}

export function shouldPromptDeleteOrphanScan(scan, promptedKey) {
  const key = buildOrphanScanPromptKey(scan)
  if (!key || key === promptedKey) {
    return false
  }
  return Number(scan?.orphan_files || 0) > 0 && Number(scan?.deleted_files || 0) === 0
}
