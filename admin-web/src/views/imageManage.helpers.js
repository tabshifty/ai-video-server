export const DEFAULT_IMAGE_ACTIVE = '1'

export function normalizeImageViewSnapshot(snapshot) {
  return {
    q: String(snapshot?.q || ''),
    status: String(snapshot?.status || ''),
    active: ['0', '1'].includes(String(snapshot?.active)) ? String(snapshot.active) : '',
    actor_id: String(snapshot?.actor_id || ''),
    collection_id: String(snapshot?.collection_id || ''),
    viewMode: snapshot?.viewMode === 'list' ? 'list' : 'grid'
  }
}

export function createImageBuiltInViews(defaultActive = DEFAULT_IMAGE_ACTIVE, defaultViewMode = 'grid') {
  const snapshot = (status) => normalizeImageViewSnapshot({
    status,
    active: defaultActive,
    viewMode: defaultViewMode
  })

  return [
    { id: 'builtin-all', label: '全部图片', builtIn: true, snapshot: snapshot('') },
    { id: 'builtin-ready', label: '可用', builtIn: true, snapshot: snapshot('ready') },
    { id: 'builtin-failed', label: '失败', builtIn: true, snapshot: snapshot('failed') }
  ]
}

export function hasImageActiveFilters(snapshot) {
  return String(snapshot?.q || '').trim() !== ''
    || String(snapshot?.status || '') !== ''
    || String(snapshot?.active) !== DEFAULT_IMAGE_ACTIVE
    || String(snapshot?.actor_id || '') !== ''
    || String(snapshot?.collection_id || '') !== ''
}
