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
      label: getEd2kLinkLabel(href)
    })
  })

  return { links, invalidCount }
}

export function parseEd2kCreateEntries(input, previousEntries = []) {
  const previous = Array.isArray(previousEntries) ? previousEntries : []
  const previousBySource = new Map(previous.map((item) => [item.sourceLink, item]))
  const reusedPreviousLines = new Set()
  const lines = String(input || '').split(/\r?\n/)
  let maxLineNumber = previous.reduce((max, item) => Math.max(max, Number(item?.lineNumber || 0)), 0)
  return lines.reduce((acc, rawLine) => {
    const sourceLink = String(rawLine || '').trim()
    if (!sourceLink) {
      return acc
    }
    const previousEntry = previousBySource.get(sourceLink)
    if (previousEntry && !reusedPreviousLines.has(previousEntry.lineNumber)) {
      reusedPreviousLines.add(previousEntry.lineNumber)
      acc.push({
        lineNumber: previousEntry.lineNumber,
        sourceLink
      })
      return acc
    }
    maxLineNumber += 1
    acc.push({
      lineNumber: maxLineNumber,
      sourceLink
    })
    return acc
  }, [])
}

export function normalizeEd2kCreateResults(results) {
  return (Array.isArray(results) ? results : []).map((item) => ({
    lineNumber: Number(item?.line_number || item?.lineNumber || 0),
    sourceLink: String(item?.source_link || item?.sourceLink || '').trim(),
    resourceHash: String(item?.resource_hash || item?.resourceHash || '').trim(),
    status: String(item?.status || '').trim(),
    message: String(item?.message || '').trim(),
    task: item?.task || null
  }))
}

export function mergeEd2kCreateSession(previousResults, latestResults) {
  const previous = Array.isArray(previousResults) ? previousResults : []
  const latest = Array.isArray(latestResults) ? latestResults : []
  const byLine = new Map(previous.map((item) => [item.lineNumber, item]))
  latest.forEach((item) => {
    if (item.lineNumber > 0) {
      byLine.set(item.lineNumber, item)
    }
  })
  return [...byLine.values()].sort((a, b) => a.lineNumber - b.lineNumber)
}

export function mergeEd2kDraftEntries(previousEntries, latestEntries, results) {
  const doneStatuses = new Set(['created', 'reused', 'duplicate'])
  const latestResultsByLine = new Map((Array.isArray(results) ? results : []).map((item) => [item.lineNumber, item]))
  const latestLineNumbers = new Set((Array.isArray(latestEntries) ? latestEntries : []).map((entry) => entry.lineNumber))
  const nextEntriesByLine = new Map()
  ;(Array.isArray(previousEntries) ? previousEntries : []).forEach((entry) => {
    if (!latestLineNumbers.has(entry.lineNumber)) {
      return
    }
    const result = latestResultsByLine.get(entry.lineNumber)
    if (doneStatuses.has(result?.status)) {
      return
    }
    nextEntriesByLine.set(entry.lineNumber, entry)
  })
  ;(Array.isArray(latestEntries) ? latestEntries : []).forEach((entry) => {
    const result = latestResultsByLine.get(entry.lineNumber)
    if (doneStatuses.has(result?.status)) {
      return
    }
    nextEntriesByLine.set(entry.lineNumber, entry)
  })
  return [...nextEntriesByLine.values()].sort((a, b) => a.lineNumber - b.lineNumber)
}

export function buildPendingEd2kInput(allEntries, results) {
  const doneStatuses = new Set(['created', 'reused', 'duplicate'])
  const latestByLine = new Map((Array.isArray(results) ? results : []).map((item) => [item.lineNumber, item]))
  return (Array.isArray(allEntries) ? allEntries : [])
    .filter((entry) => !doneStatuses.has(latestByLine.get(entry.lineNumber)?.status))
    .sort((a, b) => a.lineNumber - b.lineNumber)
    .map((entry) => entry.sourceLink)
    .join('\n')
}

export function shouldKeepEd2kCreateDialogOpen(entries, results) {
  const activeLineNumbers = new Set((Array.isArray(entries) ? entries : []).map((entry) => entry.lineNumber))
  if (activeLineNumbers.size === 0) {
    return false
  }
  return (Array.isArray(results) ? results : []).some((item) =>
    activeLineNumbers.has(item.lineNumber) &&
    ['invalid', 'create_failed', 'enqueue_failed'].includes(item.status)
  )
}

export function filterEd2kTasks(tasks, statusFilter) {
  const list = (Array.isArray(tasks) ? tasks : []).filter((task) => String(task?.status || '') !== 'deleted')
  if (!statusFilter || statusFilter === 'all') {
    return list
  }
  return list.filter((task) => String(task?.status || '') === statusFilter)
}

export function buildEd2kTaskFocusState(currentFilter, task) {
  const normalizedFilter = String(currentFilter || 'all').trim() || 'all'
  const taskID = String(task?.id || '').trim()
  const taskStatus = String(task?.status || '').trim()
  if (!taskID) {
    return {
      filter: normalizedFilter,
      selectedTaskID: ''
    }
  }
  if (!taskStatus || normalizedFilter === 'all' || normalizedFilter === taskStatus) {
    return {
      filter: normalizedFilter,
      selectedTaskID: taskID
    }
  }
  return {
    filter: taskStatus,
    selectedTaskID: taskID
  }
}
