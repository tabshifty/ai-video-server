export const SAVED_VIEW_SCHEMA_VERSION = 1
export const CUSTOM_VIEW_ID = 'custom'

function canonicalize(value) {
  if (Array.isArray(value)) return value.map(canonicalize)
  if (!value || typeof value !== 'object') return value

  return Object.keys(value).sort().reduce((result, key) => {
    result[key] = canonicalize(value[key])
    return result
  }, {})
}

function isUserViewId(id) {
  return /^user-\S+$/u.test(id)
}

export function snapshotKey(snapshot) {
  const value = snapshot && typeof snapshot === 'object' ? snapshot : {}
  return JSON.stringify(canonicalize(value))
}

export function parseSavedViewDocument(raw, normalizeSnapshot) {
  try {
    const document = JSON.parse(String(raw || ''))
    if (
      document?.version !== SAVED_VIEW_SCHEMA_VERSION ||
      !Array.isArray(document.items) ||
      typeof normalizeSnapshot !== 'function'
    ) {
      return []
    }

    const seen = new Set()
    return document.items.flatMap((item) => {
      try {
        const id = String(item?.id || '').trim()
        const label = String(item?.label || '').trim()
        if (!isUserViewId(id) || !label || seen.has(id)) return []

        const snapshot = normalizeSnapshot(item?.snapshot)
        seen.add(id)
        return [{ id, label, snapshot }]
      } catch (_) {
        return []
      }
    })
  } catch (_) {
    return []
  }
}

export function serializeSavedViews(items) {
  return JSON.stringify({
    version: SAVED_VIEW_SCHEMA_VERSION,
    items: Array.isArray(items) ? items : []
  })
}

export function upsertSavedView(items, view) {
  const next = (Array.isArray(items) ? items : []).filter((item) => item.id !== view.id)
  return [...next, view]
}

export function removeSavedView(items, id) {
  return (Array.isArray(items) ? items : []).filter((item) => item.id !== id)
}

export function createSavedViewId(now = Date.now()) {
  return `user-${now}`
}
