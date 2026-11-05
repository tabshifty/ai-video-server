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
  return document && Array.isArray(document.paths) ? knownUnique(document.paths, validPaths, 3) : []
}

export function pushRecentRoute(paths, path, validPaths, limit = 3) {
  const finiteLimit = Number.isFinite(limit) ? Math.trunc(limit) : 3
  const normalizedLimit = Math.min(3, Math.max(0, finiteLimit))
  if (normalizedLimit === 0) return []
  if (!validPaths.includes(path)) return knownUnique(paths, validPaths, normalizedLimit)
  return knownUnique([path, ...(Array.isArray(paths) ? paths : [])], validPaths, normalizedLimit)
}

export function serializeRecentRoutes(paths) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, paths })
}

export function parseExpandedGroupKeys(raw, validKeys) {
  const document = parseDocument(raw)
  if (!document || !Array.isArray(document.keys)) return [...validKeys]
  // 分组键变更后旧键会被白名单过滤。若存储里出现未知键，说明是旧版分组结构，
  // 直接回退为全部展开，避免升级后新分组默认折叠、用户找不到入口。
  if (document.keys.some((key) => !validKeys.includes(key))) return [...validKeys]
  return knownUnique(document.keys, validKeys)
}

export function ensureActiveGroup(keys, activeKey, validKeys) {
  return knownUnique([...(Array.isArray(keys) ? keys : []), activeKey], validKeys)
}

export function serializeExpandedGroupKeys(keys) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, keys })
}
