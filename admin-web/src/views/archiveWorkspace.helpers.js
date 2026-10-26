export const MAIN_GROUP = 'main'
export const cloneWorkspace = (value) => JSON.parse(JSON.stringify(value))
export const fileTitle = (name) => String(name || '').replace(/\.[^.]*$/, '').replace(/www\.98T\.la@/gi, '').replace(/\s+/g, ' ').trim()

export function descendantIDs(nodes, id, parentField = 'parent_id') {
  const found = new Set([id])
  let changed = true
  while (changed) {
    changed = false
    for (const node of nodes || []) {
      if (found.has(node[parentField]) && !found.has(node.id)) { found.add(node.id); changed = true }
    }
  }
  return found
}

export function workspaceTree(nodes, rootID = '', rootLabel = '全部文件') {
  const byID = new Map((nodes || []).map((node) => [node.id, { ...node, label: node.name, children: [] }]))
  const root = { id: rootID, label: rootLabel, children: [] }
  for (const node of byID.values()) {
    if (node.id === rootID) continue
    const parent = byID.get(node.parent_id)
    if (parent) parent.children.push(node)
    else root.children.push(node)
  }
  if (byID.has(rootID)) return [{ ...byID.get(rootID) }]
  return [root]
}

export function selectedWorkspaceFiles(draft, files, ids = [], directoryIDs = []) {
  const selected = new Set(ids)
  const directories = new Set()
  for (const id of directoryIDs) for (const child of descendantIDs(draft.directories, id)) directories.add(child)
  return (files || []).filter((f) => draft.files[f.id] && (selected.has(f.id) || directories.has(draft.files[f.id].directory_id))).map((f) => f.id)
}

export function importableWorkspaceFiles(data, ids) {
  const active = new Set((data.jobs || []).filter((j) => ['queued', 'running'].includes(j.state)).map((j) => j.file_id))
  const byID = new Map((data.files || []).map((f) => [f.id, f]))
  return [...new Set(ids)].filter((id) => {
    const f = byID.get(id), node = data.draft.files[id]
    return f && node && !node.excluded && ['video', 'image'].includes(f.media_kind) && ['pending', 'failed'].includes(f.status) && !active.has(id)
  })
}

export function summarizeWorkspaceImport(data, ids) {
  const selected = importableWorkspaceFiles(data, ids)
  const summary = { ids: selected, videos: 0, images: 0, missingVideoCollections: 0, missingImageCollections: 0, missingVideoImageCollections: 0, emptyTitles: [] }
  const byID = new Map(data.files.map((f) => [f.id, f]))
  for (const id of selected) {
    const f = byID.get(id), values = data.effective[id] || {}
    if (!String(values.title || '').trim()) summary.emptyTitles.push(data.draft.files[id].name)
    if (f.media_kind === 'video') {
      summary.videos++
      if (!values.video_collection_ids?.length) summary.missingVideoCollections++
      if (!values.image_collection_ids?.length) summary.missingVideoImageCollections++
    } else {
      summary.images++
      if (!values.image_collection_ids?.length) summary.missingImageCollections++
    }
  }
  return summary
}

export function resolveWorkspaceValues(draft, fileID, kind) {
  const file = draft.files[fileID], byID = new Map(draft.groups.map((g) => [g.id, g]))
  const chain = [], seen = new Set()
  let id = file?.group_id || MAIN_GROUP
  while (id && byID.has(id) && !seen.has(id)) { seen.add(id); const group = byID.get(id); chain.unshift(group); id = group.parent_id }
  let values = { title_mode: 'uniform', title: '', description: '', tags: [], video_type: 'short', video_collection_ids: [], image_collection_ids: [] }
  for (const group of chain) Object.assign(values, group.config)
  if (kind === 'image') values.title_mode = 'filename'
  Object.assign(values, file?.config)
  if (file?.config?.title !== undefined && file?.config?.title_mode === undefined) values.title_mode = 'uniform'
  if (values.title_mode === 'filename') values.title = fileTitle(file?.name)
  values.title = String(values.title || '').trim()
  return values
}

// randomUUID is unavailable on the LAN HTTP origin used by the admin UI.
export function newWorkspaceID(cryptoProvider = globalThis.crypto) {
  if (typeof cryptoProvider?.randomUUID === 'function') return cryptoProvider.randomUUID()
  const bytes = new Uint8Array(16)
  if (cryptoProvider?.getRandomValues) cryptoProvider.getRandomValues(bytes)
  else for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256)
  bytes[6] = (bytes[6] & 15) | 64
  bytes[8] = (bytes[8] & 63) | 128
  const hex = [...bytes].map((b) => b.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}
