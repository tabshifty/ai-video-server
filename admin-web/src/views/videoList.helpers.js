export function getVideoStatusMeta(status) {
  const normalized = String(status || '').trim().toLowerCase()
  const map = {
    uploaded: { label: '已上传', tagType: 'info' },
    scraping: { label: '刮削中', tagType: 'info' },
    tv_pending: { label: '待绑定', tagType: 'warning' },
    processing: { label: '处理中', tagType: 'info' },
    ready: { label: '可播放', tagType: 'success' },
    failed: { label: '失败', tagType: 'danger' }
  }
  return map[normalized] || { label: normalized || '-', tagType: 'info' }
}

const manualVideoStatusValues = ['uploaded', 'scraping', 'tv_pending', 'ready', 'failed']

export function getManualVideoStatusOptions() {
  return manualVideoStatusValues.map((status) => ({
    value: status,
    label: getVideoStatusMeta(status).label
  }))
}

export function canManuallyEditVideoStatus(status) {
  return normalizeVideoStatus(status) !== 'processing'
}

export function getManualVideoStatusValue(status) {
  const normalized = normalizeVideoStatus(status)
  if (!canManuallyEditVideoStatus(normalized) || !manualVideoStatusValues.includes(normalized)) {
    return ''
  }
  return normalized
}

export function buildAVManualScrapeRoute(video) {
  const videoID = toText(video?.id)
  const externalID = extractAVManualScrapeExternalID(video)
  return {
    path: '/av-scrape',
    query: {
      video_id: videoID,
      external_id: externalID,
      title: externalID || toText(video?.title)
    }
  }
}

export function getVideoThumbnailURL(video) {
  const videoID = toText(video?.id)
  return videoID ? `/api/v1/videos/${videoID}/thumbnail` : ''
}

export function shouldShowVideoThumbnail(video) {
  return normalizeVideoStatus(video?.status) === 'ready' && getVideoThumbnailURL(video) !== ''
}

export function getVideoThumbnailPlaceholder(video) {
  return shouldShowVideoThumbnail(video) ? '暂无封面' : '未就绪'
}

export function extractTvPendingDiagnostics(metadata) {
  const source = metadata && typeof metadata === 'object' ? metadata : {}
  return {
    error: toText(source.scrape_error),
    stage: toText(source.scrape_stage),
    parsedTitle: toText(source.parsed_title),
    parsedSeasonNumber: toPositiveInt(source.parsed_season_number),
    parsedEpisodeNumber: toPositiveInt(source.parsed_episode_number),
    candidateCount: toPositiveInt(source.candidate_count),
    candidatePreview: Array.isArray(source.candidate_preview) ? source.candidate_preview : []
  }
}

export function tvPendingStageLabel(stage) {
  const normalized = String(stage || '').trim().toLowerCase()
  const map = {
    parse_failed: '文件名解析失败',
    candidate_ambiguous: '候选结果不唯一',
    candidate_not_found: '未找到可靠候选',
    tmdb_missing_episode: 'TMDB 缺少目标季集',
    api_error: '远端接口失败'
  }
  return map[normalized] || (normalized || '-')
}

export function teardownPreviewPlayer(player) {
  if (!player || typeof player !== 'object') {
    return
  }

  try {
    if (typeof player.pause === 'function') {
      player.pause()
    }
  } catch (_) {}

  try {
    if (typeof player.removeAttribute === 'function') {
      player.removeAttribute('src')
    }
    if ('src' in player) {
      player.src = ''
    }
  } catch (_) {}

  try {
    if (typeof player.load === 'function') {
      player.load()
    }
  } catch (_) {}
}

function toText(value) {
  if (typeof value === 'string') {
    const trimmed = value.trim()
    return trimmed === '' ? '' : trimmed
  }
  if (typeof value === 'number' && Number.isFinite(value)) {
    return String(value)
  }
  return ''
}

function toPositiveInt(value) {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return 0
  }
  return parsed
}

function normalizeVideoStatus(status) {
  return String(status || '').trim().toLowerCase()
}

function extractAVManualScrapeExternalID(video) {
  const metadata = video && typeof video.metadata === 'object' && video.metadata !== null ? video.metadata : {}
  return toText(video?.external_id) || toText(video?.av_code) || toText(metadata.external_id) || toText(metadata.av_code) || toText(metadata.number)
}
