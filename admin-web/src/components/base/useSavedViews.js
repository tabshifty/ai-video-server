import { computed, ref } from 'vue'
import {
  CUSTOM_VIEW_ID,
  createSavedViewId,
  parseSavedViewDocument,
  removeSavedView as removeSavedViewItem,
  serializeSavedViews,
  snapshotKey,
  upsertSavedView
} from './savedView.helpers'

function resolveStorage(storage) {
  if (storage !== undefined) return storage

  try {
    return typeof window === 'undefined' ? null : window.localStorage
  } catch (_) {
    return null
  }
}

export function useSavedViews({
  storageKey,
  builtInViews,
  normalizeSnapshot,
  getCurrentSnapshot,
  applySnapshot,
  refresh = () => {},
  storage,
  now = Date.now
}) {
  const targetStorage = resolveStorage(storage)
  const builtIns = Array.isArray(builtInViews) ? builtInViews : []
  const defaultView = builtIns[0] || null
  let initialViews = []

  try {
    initialViews = parseSavedViewDocument(targetStorage?.getItem(storageKey), normalizeSnapshot)
  } catch (_) {
    initialViews = []
  }

  const userViews = ref(initialViews)
  const selectedViewId = ref(defaultView?.id || '')
  const availableViews = computed(() => [...builtIns, ...userViews.value])
  const selectedView = computed(() =>
    availableViews.value.find((item) => item.id === selectedViewId.value) || defaultView
  )
  const currentSnapshot = computed(() => normalizeSnapshot(getCurrentSnapshot()))
  const activeViewId = computed(() => {
    if (!selectedView.value) return CUSTOM_VIEW_ID

    const sourceSnapshot = normalizeSnapshot(selectedView.value.snapshot)
    return snapshotKey(currentSnapshot.value) === snapshotKey(sourceSnapshot)
      ? selectedView.value.id
      : CUSTOM_VIEW_ID
  })
  const editableSourceId = computed(() =>
    selectedView.value && !selectedView.value.builtIn ? selectedView.value.id : ''
  )

  function persist() {
    try {
      targetStorage?.setItem(storageKey, serializeSavedViews(userViews.value))
    } catch (_) {
      // 浏览器禁用持久化时，当前会话中的响应式视图仍然可用。
    }
  }

  function nextViewId() {
    const baseId = createSavedViewId(now())
    const usedIds = new Set(availableViews.value.map((item) => item.id))
    if (!usedIds.has(baseId)) return baseId

    let suffix = 2
    while (usedIds.has(`${baseId}-${suffix}`)) suffix += 1
    return `${baseId}-${suffix}`
  }

  async function selectView(id) {
    const view = availableViews.value.find((item) => item.id === id)
    if (!view) return false

    selectedViewId.value = view.id
    applySnapshot(normalizeSnapshot(view.snapshot))
    await refresh()
    return true
  }

  function saveView(label) {
    const normalizedLabel = String(label ?? '').trim()
    if (!normalizedLabel) return null

    const view = {
      id: nextViewId(),
      label: normalizedLabel,
      snapshot: currentSnapshot.value
    }
    userViews.value = upsertSavedView(userViews.value, view)
    selectedViewId.value = view.id
    persist()
    return view.id
  }

  function updateView(id) {
    const source = userViews.value.find((item) => item.id === id)
    if (!source) return false

    userViews.value = upsertSavedView(userViews.value, {
      ...source,
      snapshot: currentSnapshot.value
    })
    selectedViewId.value = id
    persist()
    return true
  }

  function renameView({ id, label } = {}) {
    const source = userViews.value.find((item) => item.id === id)
    const normalizedLabel = String(label ?? '').trim()
    if (!source || !normalizedLabel) return false

    userViews.value = upsertSavedView(userViews.value, {
      ...source,
      label: normalizedLabel
    })
    persist()
    return true
  }

  async function removeView(id) {
    if (!userViews.value.some((item) => item.id === id) || !defaultView) return false

    userViews.value = removeSavedViewItem(userViews.value, id)
    persist()
    return selectView(defaultView.id)
  }

  return {
    userViews,
    availableViews,
    activeViewId,
    editableSourceId,
    selectView,
    saveView,
    updateView,
    renameView,
    removeView
  }
}
