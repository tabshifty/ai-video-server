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

export function directoryBreadcrumbs(directories, currentID) {
  const byID = new Map((directories || []).map((directory) => [directory.id, directory]))
  const path = [], seen = new Set()
  let id = currentID
  while (id && byID.has(id) && !seen.has(id)) {
    seen.add(id)
    const directory = byID.get(id)
    path.unshift({ id: directory.id, name: directory.name })
    id = directory.parent_id
  }
  return [{ id: '', name: '根目录' }, ...path]
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
    if (!f || !node || node.excluded || !['video', 'image'].includes(f.media_kind) || active.has(id)) return false
    if (data.draft.layout_version >= 2) {
      if (!node.group_id || !data.draft.groups.some((g) => g.id === node.group_id)) return false
      if ((data.jobs || []).some((j) => ['queued', 'running'].includes(j.state) && (j.snapshot?.group_id || data.draft.files[j.file_id]?.group_id) === node.group_id)) return false
      if (['pending', 'failed'].includes(f.status)) return true
      const job = (data.jobs || []).find((j) => j.file_id === id)
      if (job?.state === 'failed') return true
      return ['ready', 'existing'].includes(f.status) && !workspaceRelationsApplied(job?.state === 'completed' ? job.snapshot : null, resolveWorkspaceValues(data.draft, id, f.media_kind))
    }
    return ['pending', 'failed'].includes(f.status)
  })
}

export function summarizeWorkspaceImport(data, ids) {
  const selected = importableWorkspaceFiles(data, ids)
  const summary = { ids: selected, videos: 0, images: 0, missingVideoCollections: 0, missingImageCollections: 0, missingVideoImageCollections: 0, emptyTitles: [] }
  const byID = new Map(data.files.map((f) => [f.id, f]))
  for (const id of selected) {
    const f = byID.get(id), values = data.effective[id] || {}
    if (!['ready', 'existing'].includes(f.status) && !String(values.title || '').trim()) summary.emptyTitles.push(data.draft.files[id].name)
    if (f.media_kind === 'video') {
      summary.videos++
      if (!values.video_collection_ids?.length) summary.missingVideoCollections++
      if (!values.image_collection_ids?.length) summary.missingVideoImageCollections++
    } else {
      summary.images++
      if (!values.image_collection_ids?.length) summary.missingImageCollections++
    }
  }
  if (data.draft.layout_version >= 2) {
    const groups = new Set(selected.map((id) => data.draft.files[id].group_id))
    for (const f of data.files) {
      const node = data.draft.files[f.id]
      if (!node || node.excluded || !groups.has(node.group_id)) continue
      const values = resolveWorkspaceValues(data.draft, f.id, f.media_kind)
      if (f.media_kind === 'video' && !values.video_collection_ids.length) summary.missingVideoCollections = Math.max(1, summary.missingVideoCollections)
      if (f.media_kind === 'image' && !values.image_collection_ids.length) summary.missingImageCollections = Math.max(1, summary.missingImageCollections)
    }
  }
  return summary
}

export function resolveWorkspaceValues(draft, fileID, kind) {
  const file = draft.files[fileID], byID = new Map(draft.groups.map((g) => [g.id, g]))
  const chain = [], seen = new Set()
  let id = file?.group_id || (draft.layout_version >= 2 ? '' : MAIN_GROUP)
  while (id && byID.has(id) && !seen.has(id)) { seen.add(id); const group = byID.get(id); chain.unshift(group); id = group.parent_id }
  let values = { title_mode: 'uniform', title: '', description: '', tags: [], video_type: 'short', video_collection_ids: [], image_collection_ids: [] }
  for (const group of chain) Object.assign(values, group.config)
  if (kind === 'image') values.title_mode = 'filename'
  const config = { ...file?.config }
  if (draft.layout_version >= 2) { delete config.video_collection_ids; delete config.image_collection_ids }
  Object.assign(values, config)
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


export function workspaceRelationsApplied(have, want) {
  return !!have?.group_relations && ['video_collection_ids', 'image_collection_ids'].every((key) => (want[key] || []).every((id) => (have[key] || []).includes(id)))
}

export function workspaceGroupStatus(data, groupID) {
  const ids = Object.keys(data.draft.files).filter((id) => data.draft.files[id].group_id === groupID && !data.draft.files[id].excluded && ['video', 'image'].includes(data.files.find((f) => f.id === id)?.media_kind))
  if ((data.jobs || []).some((j) => ['queued', 'running'].includes(j.state) && (j.snapshot?.group_id || data.draft.files[j.file_id]?.group_id) === groupID)) return '处理中'
  if (!ids.length) return '空分组'
  const relevant = (data.jobs || []).filter((j) => ids.includes(j.file_id))
  if (relevant.some((j) => j.state === 'failed')) return '有失败'
  const pending = importableWorkspaceFiles(data, ids)
  if (!pending.length) return '已完成'
  return pending.some((id) => ['ready', 'existing'].includes(data.files.find((f) => f.id === id)?.status)) ? '待补充关联' : '待入库'
}
