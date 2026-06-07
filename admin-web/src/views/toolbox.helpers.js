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
