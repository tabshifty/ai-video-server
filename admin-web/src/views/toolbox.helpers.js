const ED2K_SCHEME_RE = /^ed2k:\/\//i

function safeDecodeEd2kText(value) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
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

export function getEd2kInfoHash(link) {
  const raw = String(link || '').trim()
  if (!/^ed2k:\/\//i.test(raw)) {
    return ''
  }
  const parts = raw.slice('ed2k://'.length).split('|')
  if (parts[0] === '') {
    parts.shift()
  }
  if (parts.length < 4 || parts[0]?.toLowerCase() !== 'file') {
    return ''
  }
  const hash = String(parts[3] || '').trim()
  return /^[0-9a-fA-F]{32}$/.test(hash) ? hash.toLowerCase() : ''
}

export const CLOUD_DOWNLOAD_ACTIVE_STATUSES = new Set([
  'pending',
  'submitting',
  'queued',
  'ready',
  'downloading'
])

export function indexCloudBatchItems(items) {
  const index = {}
  for (const item of Array.isArray(items) ? items : []) {
    const infoHash = String(item?.info_hash || '').trim().toLowerCase()
    const link = String(item?.link || '').trim()
    // Backend invalid items have no info_hash. Register the link and the hash
    // that the frontend parsed from it, so the row lookup finds the outcome.
    const keys = new Set([infoHash, link, getEd2kInfoHash(link)].filter(Boolean))
    for (const key of keys) {
      index[key] = item
    }
  }
  return index
}

export function summarizeEd2kBatchSelection(uniqueLinks, duplicates = [], statuses = {}) {
  const invalidLinks = []
  const submittableLinks = []
  for (const link of uniqueLinks) {
    if (link.infoHash) {
      submittableLinks.push(link)
    } else {
      invalidLinks.push(link)
    }
  }

  const counts = { fresh: 0, retry: 0, completed: 0, active: 0 }
  const hashes = []
  for (const link of submittableLinks) {
    hashes.push(link.infoHash)
    const status = statuses?.[link.infoHash]?.status
    if (status === 'completed') {
      counts.completed += 1
    } else if (CLOUD_DOWNLOAD_ACTIVE_STATUSES.has(status)) {
      counts.active += 1
    } else if (status === 'failed' || status === 'blocked') {
      counts.retry += 1
    } else {
      counts.fresh += 1
    }
  }

  return {
    invalidLinks,
    submittableLinks,
    hashes: [...new Set(hashes)],
    duplicateCount: Array.isArray(duplicates) ? duplicates.length : 0,
    ...counts
  }
}

export function dedupeEd2kLinksByInfoHash(links) {
  const seen = new Set()
  const unique = []
  const duplicates = []

  for (const link of links) {
    const key = link.infoHash || link.href
    if (seen.has(key)) {
      duplicates.push(link)
      continue
    }
    seen.add(key)
    unique.push(link)
  }

  return { links: unique, duplicates }
}

export function extractApiErrorMessage(error, fallback = '操作失败') {
  if (error?.response?.status === 404) {
    return '云下载功能未启用'
  }
  const responseMessage = error?.response?.data?.msg
  if (typeof responseMessage === 'string' && responseMessage.trim()) {
    return responseMessage.trim()
  }
  if (typeof error?.message === 'string' && error.message.trim()) {
    return error.message.trim()
  }
  return fallback
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
    const parts = href.split('|')
    const isEd2kFileLink = ED2K_SCHEME_RE.test(href) && parts.length >= 6 && parts[1]?.toLowerCase() === 'file'
    if (!isEd2kFileLink) {
      invalidCount += 1
      return
    }
    links.push({
      id: `${index + 1}:${href}`,
      lineNumber: index + 1,
      href,
      label: getEd2kLinkLabel(href),
      infoHash: getEd2kInfoHash(href)
    })
  })

  return { links, invalidCount }
}
