// ED2K 批量入库的准入判定与标题清洗。
// 与「发送选中到115」不同：这里只处理单文件视频，并在提交前就确定入库标题。
import { VIDEO_EXTENSIONS, stripVideoExtension } from './videoUpload.filename'

export const ED2K_INGEST_VIDEO = 'video'
export const ED2K_INGEST_NOT_VIDEO = 'not_video'
export const ED2K_INGEST_UNKNOWN = 'unknown'

// 压缩包有独立的「压缩包批量入库」工作台，不能混进这条链路。
export const ARCHIVE_EXTENSIONS = [
  'rar',
  'zip',
  '7z',
  'tar',
  'gz',
  'tgz',
  'bz2',
  'xz',
  'iso',
  'cab',
  'lz4',
  'zst'
]

export const IMAGE_EXTENSIONS = [
  'jpg',
  'jpeg',
  'png',
  'gif',
  'webp',
  'bmp',
  'heic',
  'avif',
  'tif',
  'tiff'
]

// 明确不是视频、也不属于本功能的其它已知媒体/文档。
export const OTHER_MEDIA_EXTENSIONS = [
  'mp3',
  'flac',
  'wav',
  'aac',
  'm4a',
  'ogg',
  'ape',
  'srt',
  'ass',
  'ssa',
  'vtt',
  'sub',
  'idx',
  'pdf',
  'epub',
  'txt',
  'nfo'
]

export function getEd2kFilenameExtension(label) {
  const leaf = String(label || '')
    .trim()
    .replace(/\\/g, '/')
    .split('/')
    .pop()
  if (!leaf) {
    return ''
  }
  // 只认最后一段扩展名：`.tar.gz` 归为 gz；`.mp4`（点在开头）视为没有扩展名。
  const dot = leaf.lastIndexOf('.')
  if (dot <= 0 || dot === leaf.length - 1) {
    return ''
  }
  const extension = leaf.slice(dot + 1).trim()
  return /^[A-Za-z0-9]{1,12}$/.test(extension) ? extension.toLowerCase() : ''
}

// 括号标签分三态：
//   drop   —— 站点/语言类，连内容一起丢掉；
//   unwrap —— 画质/编码类，去掉括号但保留内容（ADR 要求"不删除画质标记"）；
//   keep   —— 有信息量的括号，原样保留。
const SITE_AND_LANGUAGE_WORDS = [
  'www',
  '.com',
  '.la',
  '.net',
  '.org',
  '.me',
  '中文字幕',
  '中字',
  '字幕',
  '无码',
  '有码',
  '国语',
  '粤语',
  '双语',
  '首发',
  '打包'
]

const QUALITY_WORDS = [
  '1080p',
  '720p',
  '2160p',
  '480p',
  '4k',
  '8k',
  'hdr',
  'x264',
  'x265',
  'hevc',
  'avc',
  'aac',
  'ac3',
  '10bit',
  '12bit',
  '60fps',
  'web-dl',
  'webdl',
  'bluray',
  'blu-ray',
  'hdrip',
  'bdrip',
  '蓝光',
  '高清',
  '超清',
  '标清'
]

function classifyBracketContent(inner) {
  const text = String(inner || '').trim().toLowerCase()
  if (!text) {
    return 'drop'
  }
  // 画质优先：`[中文字幕 1080p]` 必须去括号保留 1080p，而不是整段丢掉。
  if (QUALITY_WORDS.some((keyword) => text.includes(keyword))) {
    return 'unwrap'
  }
  if (SITE_AND_LANGUAGE_WORDS.some((keyword) => text.includes(keyword))) {
    return 'drop'
  }
  return 'keep'
}

// 【...】/[...]/(...) 三组成对标签，按内容分三态处理。
const BRACKET_PATTERN = /【([^】]*)】|\[([^\]]*)\]|\(([^)]*)\)/g
// 域名水印：www.98t.la@、98T.la@、hhd800.com@ 一类。
const WATERMARK_PATTERN = /(?:www\.)?[a-z0-9-]+\.(?:la|com|net|org)@/gi
// 重复下载后缀：`- 副本`、`(1)`、`（1）`，只在结尾处生效。
const DUPLICATE_SUFFIX_PATTERN =
  /(?:[\s_-]*副本\s*\d{0,3}|[(（]\s*\d{1,3}\s*[)）]|\s*_[2-9])$/

export function cleanEd2kIngestTitle(label) {
  const base = stripVideoExtension(String(label || '').trim())
  let text = base.replace(WATERMARK_PATTERN, ' ')
  text = text.replace(BRACKET_PATTERN, (match, curly, square, round) => {
    const inner = curly ?? square ?? round
    const verdict = classifyBracketContent(inner)
    if (verdict === 'drop') {
      return ' '
    }
    if (verdict === 'unwrap') {
      return ` ${String(inner).trim()} `
    }
    return match
  })
  text = text.replace(DUPLICATE_SUFFIX_PATTERN, ' ')
  text = text
    .replace(/[._]+/g, ' ')
    .replace(/[-_]{2,}/g, ' ')
    .replace(/\s+/g, ' ')
    .replace(/^[\s\-_.]+|[\s\-_.]+$/g, '')
    .trim()
  return text || base
}

export function classifyEd2kIngestLink(label) {
  const extension = getEd2kFilenameExtension(label)
  if (!extension) {
    return { kind: ED2K_INGEST_UNKNOWN, extension, reason: '文件名没有扩展名' }
  }
  if (VIDEO_EXTENSIONS.includes(extension)) {
    return { kind: ED2K_INGEST_VIDEO, extension, reason: '' }
  }
  if (ARCHIVE_EXTENSIONS.includes(extension)) {
    return {
      kind: ED2K_INGEST_NOT_VIDEO,
      extension,
      reason: '压缩包请改用「压缩包批量入库」'
    }
  }
  if (IMAGE_EXTENSIONS.includes(extension)) {
    return { kind: ED2K_INGEST_NOT_VIDEO, extension, reason: '图片不支持批量入库' }
  }
  if (OTHER_MEDIA_EXTENSIONS.includes(extension)) {
    return {
      kind: ED2K_INGEST_NOT_VIDEO,
      extension,
      reason: `.${extension} 不是视频文件`
    }
  }
  return {
    kind: ED2K_INGEST_UNKNOWN,
    extension,
    reason: `无法识别扩展名 .${extension}`
  }
}

// 把已解析的 ED2K 链接整理成入库计划：可入库、明确排除、待裁决。
export function buildEd2kIngestPlan(links) {
  const videos = []
  const notVideos = []
  const unknown = []
  for (const link of Array.isArray(links) ? links : []) {
    const verdict = classifyEd2kIngestLink(link?.label)
    const item = {
      ...link,
      kind: verdict.kind,
      extension: verdict.extension,
      reason: verdict.reason,
      title: cleanEd2kIngestTitle(link?.label)
    }
    if (verdict.kind === ED2K_INGEST_VIDEO) {
      videos.push(item)
      continue
    }
    if (verdict.kind === ED2K_INGEST_NOT_VIDEO) {
      notVideos.push(item)
      continue
    }
    unknown.push(item)
  }
  return { videos, notVideos, unknown }
}

export function ed2kIngestKey(link) {
  return String(link?.infoHash || link?.href || '')
}

// 最终提交条目：视频全部提交；无法识别的只有被强制放行才提交。
export function buildEd2kIngestLinks(plan, { forcedKeys = new Set() } = {}) {
  const items = []
  for (const link of plan?.videos || []) {
    items.push({ link: link.href, label: link.label, infoHash: link.infoHash, title: link.title, forced: false })
  }
  for (const link of plan?.unknown || []) {
    if (forcedKeys.has(ed2kIngestKey(link))) {
      items.push({ link: link.href, label: link.label, infoHash: link.infoHash, title: link.title, forced: true })
    }
  }
  return items
}

export const CLOUD_INGEST_TITLE_MAX_RUNES = 200

export const CLOUD_INGEST_TITLE_MODES = [
  { value: 'filename', label: '按文件名' },
  { value: 'uniform', label: '统一标题' }
]

export const CLOUD_INGEST_VIDEO_TYPES = [
  { value: 'short', label: '短视频' },
  { value: 'movie', label: '电影' },
  { value: 'episode', label: '剧集分集' },
  { value: 'av', label: 'AV' }
]

export const CLOUD_INGEST_SITE_CATEGORIES = [
  { value: 'japanese', label: '日本' },
  { value: 'western', label: '欧美' },
  { value: 'fc2', label: 'FC2' }
]

export function runeLength(text) {
  return Array.from(String(text ?? '')).length
}

export function truncateTitle(text, max = CLOUD_INGEST_TITLE_MAX_RUNES) {
  const chars = Array.from(String(text ?? ''))
  if (chars.length <= max) {
    return chars.join('')
  }
  return `${chars.slice(0, max - 1).join('')}…`
}

// 提交前校验，与后端 normalizeCloudIngestBatchInput 保持同一口径。
export function validateEd2kIngestForm(form) {
  const titleMode = form?.titleMode === 'uniform' ? 'uniform' : 'filename'
  if (titleMode === 'uniform' && !String(form?.uniformTitle || '').trim()) {
    return '统一标题不能为空'
  }
  if (titleMode === 'uniform' && runeLength(form?.uniformTitle) > CLOUD_INGEST_TITLE_MAX_RUNES) {
    return `统一标题最多 ${CLOUD_INGEST_TITLE_MAX_RUNES} 字`
  }
  if (String(form?.videoType || 'short') !== 'short' && (form?.collectionIDs || []).length > 0) {
    return '仅短视频支持视频合集'
  }
  return ''
}

export function buildEd2kIngestPayload({ links = [], form = {} } = {}) {
  const titleMode = form.titleMode === 'uniform' ? 'uniform' : 'filename'
  return {
    title_mode: titleMode,
    uniform_title: titleMode === 'uniform' ? String(form.uniformTitle || '').trim() : '',
    video_type: String(form.videoType || 'short'),
    site_category: String(form.videoType || 'short') === 'av' ? String(form.siteCategory || 'japanese') : '',
    description: String(form.description || '').trim(),
    tags: (form.tags || []).map((tag) => String(tag).trim()).filter(Boolean),
    actor_ids: form.actorIDs || [],
    actor_names: (form.actorNames || []).map((name) => String(name).trim()).filter(Boolean),
    collection_ids: String(form.videoType || 'short') === 'short' ? form.collectionIDs || [] : [],
    image_collection_id: form.imageCollectionID || null,
    links: links.map((item) => ({
      link: item.link,
      title: titleMode === 'filename' ? truncateTitle(String(item.title || '').trim()) : '',
      forced: Boolean(item.forced)
    }))
  }
}

// 入库批成员在任务列表里的派生状态。没有导入记录不等于失败：
// 它可能只是还没落盘，或刚落盘等待自动入库。
export function resolveIngestStatusMeta(row) {
  if (!row?.ingest_batch_id) {
    return null
  }
  const status = String(row.status || '')
  const importStatus = String(row.import?.status || '')
  if (['pending', 'submitting', 'queued', 'ready', 'downloading'].includes(status)) {
    return { label: '等待落盘', tone: 'info' }
  }
  if (status === 'failed' || status === 'blocked') {
    return { label: '下载失败', tone: 'danger' }
  }
  switch (importStatus) {
    case 'queued':
    case 'hashing':
    case 'copying':
    case 'creating':
      return { label: '入库中', tone: 'info' }
    case 'imported':
      return { label: '已入库', tone: 'success' }
    case 'skipped_existing':
      return { label: '已在库', tone: 'success' }
    case 'failed':
      return { label: '入库失败', tone: 'danger' }
    default:
      return { label: '待自动入库', tone: 'warning' }
  }
}

export const INGEST_VIDEO_TYPE_LABELS = {
  short: '短视频',
  movie: '电影',
  episode: '剧集分集',
  av: 'AV'
}

// 批 id 一律按字符串处理：el-select 清空时可能给 undefined/null。
export function normalizedIngestBatchID(value) {
  return value == null ? '' : String(value)
}

export function ingestBatchKey(batch) {
  return String(batch?.id || '')
}

export function ingestBatchWhen(iso) {
  const date = iso ? new Date(iso) : null
  if (!date || Number.isNaN(date.getTime())) {
    return ''
  }
  const pad = (value) => String(value).padStart(2, '0')
  return `${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

// 下拉选项文案。批详情未知时也只给中文占位，绝不把 UUID 露出来。
export function ingestBatchLabel(batch) {
  if (batch?.placeholder) {
    return '入库批（当前批）'
  }
  const type = INGEST_VIDEO_TYPE_LABELS[batch?.video_type] || batch?.video_type || ''
  const members = batch?.counts?.members || 0
  const when = ingestBatchWhen(batch?.created_at)
  return ['入库批', type, `${members} 条`, when].filter(Boolean).join(' · ')
}

// 选中项显示文案：从批量入库跳转过来时，选项可能还没到、或该批不在最近若干条里。
// 这里保证任何分支都返回中文，永远不回退成 id。
export function ingestBatchSelectedLabel(value, options = [], loading = false) {
  const selected = normalizedIngestBatchID(value)
  if (!selected) {
    return '全部入库批'
  }
  const found = (options || []).find((batch) => ingestBatchKey(batch) === selected)
  if (found) {
    return ingestBatchLabel(found)
  }
  return loading ? '入库批（加载中…）' : '入库批（当前批）'
}
