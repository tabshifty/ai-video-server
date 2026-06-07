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

const ED2K_SCHEME_RE = /^ed2k:\/\//i

function pad2(value) {
  return String(value).padStart(2, '0')
}

function safeDecodeEd2kText(value) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

export function getOrphanScanStatusLabel(status) {
  return ORPHAN_SCAN_STATUS_LABELS[status] || '未知'
}

export function getOrphanScanStatusType(status) {
  return ORPHAN_SCAN_STATUS_TYPES[status] || 'info'
}

export function formatOrphanScanTime(value) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`
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

export function getEd2kLinkLabel(href) {
  const normalized = String(href || '').trim()
  if (!normalized) {
    return ''
  }
  const parts = normalized.split('|')
  if (parts.length >= 5 && parts[1]?.toLowerCase() === 'file' && parts[2]) {
    return safeDecodeEd2kText(parts[2])
  }
  return normalized
}

export function parseEd2kLinks(input) {
  const lines = String(input || '').split(/\r?\n/)
  const links = []
  let invalidCount = 0

  lines.forEach((line, index) => {
    const href = line.trim()
    if (!href) {
      return
    }
    if (!ED2K_SCHEME_RE.test(href)) {
      invalidCount += 1
      return
    }
    links.push({
      id: `${index + 1}:${href}`,
      lineNumber: index + 1,
      href,
      label: getEd2kLinkLabel(href)
    })
  })

  return { links, invalidCount }
}
