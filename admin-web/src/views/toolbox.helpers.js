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
