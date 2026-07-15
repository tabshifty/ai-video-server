export const SHELL_PREFERENCE_VERSION = 1

function parseDocument(raw) {
  if (typeof raw !== 'string' || raw.trim() === '') return null
  try {
    const value = JSON.parse(raw)
    return value && value.version === SHELL_PREFERENCE_VERSION ? value : null
  } catch (_) {
    return null
  }
}

function knownUnique(values, validValues, limit = Number.POSITIVE_INFINITY) {
  const valid = new Set(validValues)
  const seen = new Set()
  const result = []
  for (const value of Array.isArray(values) ? values : []) {
    if (!valid.has(value) || seen.has(value)) continue
    seen.add(value)
    result.push(value)
    if (result.length >= limit) break
  }
  return result
}

export function parseRecentRoutes(raw, validPaths) {
  const document = parseDocument(raw)
  return document ? knownUnique(document.paths, validPaths, 3) : []
}

export function pushRecentRoute(paths, path, validPaths, limit = 3) {
  if (!validPaths.includes(path)) return knownUnique(paths, validPaths, limit)
  return knownUnique([path, ...(Array.isArray(paths) ? paths : [])], validPaths, limit)
}

export function serializeRecentRoutes(paths) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, paths })
}

export function parseExpandedGroupKeys(raw, validKeys) {
  const document = parseDocument(raw)
  return document ? knownUnique(document.keys, validKeys) : [...validKeys]
}

export function ensureActiveGroup(keys, activeKey, validKeys) {
  return knownUnique([...(Array.isArray(keys) ? keys : []), activeKey], validKeys)
}

export function serializeExpandedGroupKeys(keys) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, keys })
}
