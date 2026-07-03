export function isPendingDeleteShortVideo(video) {
  return normalizeVideoType(video?.type) === 'short' && normalizeVideoStatus(video?.status) === 'pending_delete'
}

export function resolveTotalPendingDeletePages(totalCount, pageSize = 20) {
  const total = Number(totalCount || 0)
  const size = Number(pageSize || 20)
  if (!Number.isFinite(total) || total <= 0 || !Number.isFinite(size) || size <= 0) return 1
  return Math.max(1, Math.ceil(total / size))
}

export function resolvePendingDeleteSelectionIndex(items, requestedIndex = 0) {
  const count = Array.isArray(items) ? items.length : 0
  if (count <= 0) return -1
  const index = Number(requestedIndex)
  if (!Number.isInteger(index) || index < 0) return 0
  return Math.min(index, count - 1)
}

export function formatShortVideoDuration(seconds) {
  const totalSeconds = Number(seconds || 0)
  if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) return '--:--'
  const minutes = Math.floor(totalSeconds / 60)
  const rest = Math.floor(totalSeconds % 60)
  return `${minutes}:${String(rest).padStart(2, '0')}`
}

function normalizeVideoType(type) {
  return String(type || '').trim().toLowerCase()
}

function normalizeVideoStatus(status) {
  return String(status || '').trim().toLowerCase()
}
