export const SHORT_REVIEW_POSITION_KEY = 'admin-short-review-position'
export const SHORT_REVIEW_SOUND_KEY = 'admin-short-review-sound-enabled'

export function normalizeVideoID(value) {
  return String(value || '').trim()
}

export function normalizeReviewPosition(value) {
  if (!value) return { videoID: '', page: 1 }
  if (typeof value === 'string') {
    const raw = value.trim()
    if (!raw) return { videoID: '', page: 1 }
    if (raw.startsWith('{')) {
      try {
        return normalizeReviewPosition(JSON.parse(raw))
      } catch (_) {
        return { videoID: raw, page: 1 }
      }
    }
    return { videoID: raw, page: 1 }
  }
  const page = Math.max(1, Number.parseInt(String(value.page || 1), 10) || 1)
  return {
    videoID: normalizeVideoID(value.video_id || value.videoID || value.id),
    page
  }
}

export function isReviewableShortVideo(video) {
  return String(video?.type || '').toLowerCase() === 'short' && String(video?.status || '').toLowerCase() === 'ready'
}

export function findVideoIndexByID(items, videoID) {
  const id = normalizeVideoID(videoID)
  if (!id) return -1
  return (items || []).findIndex((item) => normalizeVideoID(item?.id) === id)
}

export function resolveInitialReviewIndex(items, savedVideoID) {
  const savedIndex = findVideoIndexByID(items, savedVideoID)
  if (savedIndex >= 0) return savedIndex
  return (items || []).length > 0 ? 0 : -1
}

export function resolveTotalReviewPages(totalCount, pageSize = 20) {
  const total = Number(totalCount || 0)
  const size = Number(pageSize || 20)
  if (!Number.isFinite(total) || total <= 0 || !Number.isFinite(size) || size <= 0) return 1
  return Math.max(1, Math.ceil(total / size))
}

export function hasMoreReviewPages({ currentPage = 1, pageSize = 20, totalCount, loadingMore = false }) {
  if (loadingMore) return false
  const page = Number(currentPage || 1)
  const size = Number(pageSize || 20)
  const total = Number(totalCount || 0)
  return Number.isFinite(page) && Number.isFinite(size) && page > 0 && size > 0 && page * size < total
}

export function resolveNextReviewStep({
  currentIndex,
  loadedCount,
  currentPage = 1,
  pageSize = 20,
  totalCount,
  loadingMore = false
}) {
  const index = Number(currentIndex)
  const loaded = Number(loadedCount || 0)
  if (!Number.isFinite(index) || index < 0) {
    return loaded > 0 ? { type: 'select', index: 0 } : { type: 'end' }
  }
  if (index + 1 < loaded) {
    return { type: 'select', index: index + 1 }
  }
  if (hasMoreReviewPages({ currentPage, pageSize, totalCount, loadingMore })) {
    return { type: 'load-more' }
  }
  return { type: 'end' }
}

export function buildShortReviewQuery({ page = 1, pageSize = 20, keyword = '' } = {}) {
  return {
    page,
    page_size: pageSize,
    q: String(keyword || '').trim(),
    type: 'short',
    status: 'ready'
  }
}

export function readStoredShortReviewPosition(storage = globalThis.localStorage) {
  try {
    return normalizeReviewPosition(storage?.getItem(SHORT_REVIEW_POSITION_KEY))
  } catch (_) {
    return { videoID: '', page: 1 }
  }
}

export function writeStoredShortReviewPosition(position, storage = globalThis.localStorage) {
  const normalized = normalizeReviewPosition(position)
  try {
    if (normalized.videoID) {
      storage?.setItem(SHORT_REVIEW_POSITION_KEY, JSON.stringify({
        video_id: normalized.videoID,
        page: normalized.page
      }))
    } else {
      storage?.removeItem(SHORT_REVIEW_POSITION_KEY)
    }
  } catch (_) {
    // 持久化失败只影响恢复位置，不影响当前审核流。
  }
}

export function readStoredShortReviewSound(storage = globalThis.sessionStorage) {
  try {
    return storage?.getItem(SHORT_REVIEW_SOUND_KEY) === '1'
  } catch (_) {
    return false
  }
}

export function writeStoredShortReviewSound(enabled, storage = globalThis.sessionStorage) {
  try {
    storage?.setItem(SHORT_REVIEW_SOUND_KEY, enabled ? '1' : '0')
  } catch (_) {
    // sessionStorage 不可用时仅保留内存态。
  }
}
