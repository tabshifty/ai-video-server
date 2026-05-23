export function getVideoStatusMeta(status) {
  const normalized = String(status || '').trim().toLowerCase()
  const map = {
    uploaded: { label: '已上传', tagType: 'info' },
    scraping: { label: '刮削中', tagType: 'info' },
    tv_pending: { label: '待绑定', tagType: 'warning' },
    av_scrape_pending: { label: '欧美 AV 待确认', tagType: 'warning' },
    processing: { label: '处理中', tagType: 'info' },
    ready: { label: '可播放', tagType: 'success' },
    failed: { label: '失败', tagType: 'danger' }
  }
  return map[normalized] || { label: normalized || '-', tagType: 'info' }
}

export const subtitleUploadAccept = '.srt,.vtt,.ass,.ssa'

const manualVideoStatusValues = ['uploaded', 'scraping', 'tv_pending', 'av_scrape_pending', 'ready', 'failed']

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

export function buildMovieManualScrapeRoute(video) {
  const query = {
    video_id: toText(video?.id),
    type: 'movie',
    title: toText(video?.title)
  }
  const year = extractMovieManualScrapeYear(video)
  if (year > 0) {
    query.year = year
  }
  return {
    path: '/scrape',
    query
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

export function extractAVScrapePendingState(metadata) {
  const source = metadata && typeof metadata === 'object' ? metadata : {}
  return {
    candidates: Array.isArray(source.scrape_preview) ? source.scrape_preview : [],
    attempt: source.scrape_attempt && typeof source.scrape_attempt === 'object' ? source.scrape_attempt : {},
    skipped: source.scrape_skipped === true,
    skipReason: toText(source.scrape_skip_reason)
  }
}

export function avMatchSourceLabel(value) {
  const normalized = toText(value).toLowerCase()
  const map = {
    oshash: 'hash 命中',
    'keyword:scenes': '场景关键字',
    'keyword:movies': '影片关键字',
    manual_retry: '手动重搜'
  }
  return map[normalized] || (normalized || '-')
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

function extractMovieManualScrapeYear(video) {
  const metadata = video && typeof video.metadata === 'object' && video.metadata !== null ? video.metadata : {}
  const tmdb = metadata.tmdb && typeof metadata.tmdb === 'object' ? metadata.tmdb : {}
  return parseReleaseYear(metadata.release_date) || parseReleaseYear(tmdb.release_date)
}

function parseReleaseYear(value) {
  const text = toText(value)
  const match = text.match(/^(\d{4})-\d{2}-\d{2}$/)
  if (!match) {
    return 0
  }
  return toPositiveInt(match[1])
}
