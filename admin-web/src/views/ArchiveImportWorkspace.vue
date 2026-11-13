<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Back, Refresh, Folder, Document, Plus, Delete, UploadFilled, MoreFilled } from '@element-plus/icons-vue'
import ArchiveWorkspaceConfig from '../components/ArchiveWorkspaceConfig.vue'
import { getAdminArchiveImportBatchDetail, getAdminArchiveWorkspace, saveAdminArchiveWorkspace, submitAdminArchiveWorkspace, generateAdminArchiveWorkspaceGroups, retryAdminArchiveImportExtract, getAdminCollections, getAdminImageCollections, createAdminCollection, createAdminImageCollection, getAdminVideoTags, getAdminPopularVideoTags } from '../api/admin'
import { cloneWorkspace, descendantIDs, workspaceTree, directoryBreadcrumbs, selectedWorkspaceFiles, summarizeWorkspaceImport, resolveWorkspaceValues, fileTitle, newWorkspaceID, workspaceGroupStatus } from './archiveWorkspace.helpers'

const route = useRoute(), router = useRouter(), batchID = String(route.params.id)
const data = ref<any>(null), draft = ref<any>(null), batch = ref<any>(null)
const loading = ref(true), errorText = ref(''), dirty = ref(false), saving = ref(false), conflict = ref(false), saveError = ref('')
const currentDirectory = ref(''), currentGroup = ref(''), selected = ref<any[]>([]), editor = ref<'group' | 'file' | 'batch'>('group'), fileID = ref('')
const directoryTreeVisible = ref(false)
const groupView = ref<'desktop' | 'members'>('desktop'), memberSelection = ref<string[]>([])
const groupMenu = reactive({ open: false, id: '', x: 0, y: 0 })
const operation = ref(''), busy = computed(() => !!operation.value), fileTable = ref<any>(null)
const bulkConfig = ref<Record<string, any>>({})
const search = ref(''), cut = ref<{ ids: string[]; dirs: string[] } | null>(null), targetGroup = ref('')
const newDirectory = reactive({ open: false, name: '' })
const groupDialog = reactive({ open: false, id: '', name: '', config: {} as Record<string, any> })
const collection = reactive({ open: false, kind: 'video', name: '', description: '', target: '' as string, targetKind: 'group-dialog' as 'group-dialog' | 'group' | 'file' | 'batch' })
const retry = reactive({ password: '', encoding_mode: 'auto' }), retrying = ref(false)
const contextMenu = reactive({ open: false, x: 0, y: 0, directory: '' })
const videoOptions = ref<any[]>([]), imageOptions = ref<any[]>([]), tagOptions = ref<string[]>([])
let requestEpoch = 0
let timer: ReturnType<typeof setTimeout> | undefined, poll: ReturnType<typeof setInterval> | undefined, muted = false, savePromise: Promise<boolean> | null = null, editSerial = 0
const activeIDs = computed(() => new Set((data.value?.jobs || []).filter((j: any) => ['queued', 'running'].includes(j.state)).map((j: any) => j.file_id)))
const jobs = computed(() => new Map((data.value?.jobs || []).map((j: any) => [j.file_id, j])))
const sources = computed(() => new Map((data.value?.files || []).map((f: any) => [f.id, f])))
const group = computed(() => draft.value?.groups.find((g: any) => g.id === currentGroup.value))
const editedFile = computed(() => draft.value?.files[fileID.value])
const fileLocked = (id: string) => activeIDs.value.has(id) || ['ready', 'existing', 'processing'].includes((sources.value.get(id) as any)?.status)
const groupLocked = (id: string) => !!id && ((data.value?.jobs || []).some((j: any) => ['queued', 'running'].includes(j.state) && (j.snapshot?.group_id || draft.value?.files[j.file_id]?.group_id) === id) || Object.entries(draft.value?.files || {}).some(([fid, f]: any) => f.group_id === id && activeIDs.value.has(fid)))
const fileMoveLocked = (id: string) => activeIDs.value.has(id) || (sources.value.get(id) as any)?.status === 'processing' || groupLocked(draft.value?.files[id]?.group_id)
const workspaceData = computed(() => ({ ...data.value, draft: draft.value, effective: effective.value }))
const groupStatus = (id: string) => draft.value ? workspaceGroupStatus(workspaceData.value, id) : ''
const ungroupedCount = computed(() => Object.values(draft.value?.files || {}).filter((f: any) => !f.group_id).length)
const directoryTree = computed(() => workspaceTree(draft.value?.directories || [], '', '根目录'))
const rows = computed(() => {
  if (!draft.value) return []
  const dirs = draft.value.directories.filter((d: any) => !search.value && d.parent_id === currentDirectory.value).map((d: any) => ({ ...d, rowID: `dir:${d.id}`, isDirectory: true }))
  const files = Object.entries(draft.value.files).filter(([, f]: any) => search.value ? f.name.toLowerCase().includes(search.value.toLowerCase()) : f.directory_id === currentDirectory.value).map(([id, f]: any) => ({ ...f, id, rowID: id, source: sources.value.get(id), job: jobs.value.get(id) }))
  return [...dirs, ...files]
})
const memberIDs = computed(() => {
  if (!draft.value) return []
  return Object.keys(draft.value.files).filter((id) => draft.value.files[id].group_id === currentGroup.value)
})
const selectedIDs = computed(() => draft.value ? selectedWorkspaceFiles(draft.value, data.value.files, selected.value.filter((r: any) => !r.isDirectory).map((r: any) => r.id), selected.value.filter((r: any) => r.isDirectory).map((r: any) => r.id)) : [])
const editorConfig = computed({ get: () => editor.value === 'batch' ? bulkConfig.value : editor.value === 'file' ? editedFile.value?.config || {} : group.value?.config || {}, set: (value) => { if (editor.value === 'batch') bulkConfig.value = value; else if (editor.value === 'file' && editedFile.value) editedFile.value.config = value; else if (group.value) group.value.config = value } })
const effective = computed(() => {
  const result: any = {}
  if (draft.value) for (const f of data.value.files) if (draft.value.files[f.id]) result[f.id] = resolveWorkspaceValues(draft.value, f.id, f.media_kind)
  return result
})
const directoryPath = computed(() => directoryBreadcrumbs(draft.value?.directories || [], currentDirectory.value))
const saveLabel = computed(() => conflict.value ? '保存冲突' : saving.value ? '正在保存…' : saveError.value ? '保存失败' : dirty.value ? '等待保存…' : '已自动保存')
const statusLabel = (row: any) => row.excluded ? '已排除' : row.job?.state === 'queued' ? '等待入库' : row.job?.state === 'running' ? '正在入库' : row.job?.state === 'failed' ? '入库失败' : ({ pending: '待入库', ready: '已入库', existing: '复用已有', processing: '处理中', failed: '失败', skipped: '不支持' } as any)[row.source?.status] || row.source?.status

function applyResponse(result: any, replaceDraft = true) {
  if (data.value && result.version < data.value.version) return
  data.value = result; batch.value = result.batch
  if (replaceDraft) { muted = true; draft.value = cloneWorkspace(result.draft); muted = false }
}
function clearWorkspace() {
  muted = true; draft.value = null; muted = false
  data.value = null; selected.value = []; currentDirectory.value = ''; currentGroup.value = ''; editor.value = 'group'; fileID.value = ''
  contextMenu.directory = ''; contextMenu.open = false; cut.value = null; targetGroup.value = ''
  groupView.value = 'desktop'; groupMenu.open = false; memberSelection.value = []; newDirectory.open = false; groupDialog.open = false; collection.open = false; bulkConfig.value = {}
}
async function load(initial = false) {
  if (busy.value || (!initial && (saving.value || dirty.value || conflict.value))) return
  const serial = editSerial, epoch = ++requestEpoch
  try {
    const detail = await getAdminArchiveImportBatchDetail(batchID)
    batch.value = detail
    if (['uploaded', 'extracting', 'needs_password', 'needs_encoding'].includes(detail.status) || (detail.status === 'failed' && !detail.files?.length)) {
      if (!detail.files?.length && !dirty.value && !saving.value && !conflict.value && epoch === requestEpoch) clearWorkspace()
      errorText.value = detail.last_error || ''; return
    }
    const result = await getAdminArchiveWorkspace(batchID)
    if (epoch !== requestEpoch || serial !== editSerial || saving.value || dirty.value || conflict.value) return
    applyResponse(result, initial || !dirty.value)
    errorText.value = ''
  } catch (e: any) { errorText.value = e.message || '加载失败' } finally { loading.value = false }
}
watch(draft, () => {
  if (muted || !draft.value) return
  editSerial++; dirty.value = true
  clearTimeout(timer); timer = setTimeout(() => { void save() }, 600)
}, { deep: true, flush: 'sync' })
async function save(internal = false): Promise<boolean> {
  if (busy.value && !internal) return false
  clearTimeout(timer)
  if (savePromise) { const ok = await savePromise; return ok && dirty.value ? save(internal) : ok }
  if (!dirty.value) return !conflict.value
  if (conflict.value) return false
  ++requestEpoch
  const serial = editSerial, sent = cloneWorkspace(draft.value)
  saving.value = true; saveError.value = ''
  savePromise = (async () => {
    try {
      const response = await saveAdminArchiveWorkspace(batchID, { version: data.value.version, draft: sent })
      // Keep edits made while this request was in flight; only advance the saved version.
      if (!data.value || response.version >= data.value.version) { data.value = response; batch.value = response.batch }
      dirty.value = serial !== editSerial
      return true
    } catch (e: any) { saveError.value = e.message || '保存失败'; conflict.value = e.code === 409; return false }
    finally { saving.value = false; savePromise = null }
  })()
  const ok = await savePromise
  if (ok && dirty.value) { timer = setTimeout(() => { void save() }, 200) }
  return ok && !dirty.value
}
function exportDraft() {
  const url = URL.createObjectURL(new Blob([JSON.stringify(draft.value, null, 2)], { type: 'application/json' }))
  const link = document.createElement('a'); link.href = url; link.download = `压缩包草稿-${batchID}.json`; link.click(); URL.revokeObjectURL(url)
}
async function reloadDraft() {
  try { await ElMessageBox.confirm('将重新加载服务器已保存的草稿。可先下载当前草稿以保留本地编辑。', '重新加载', { confirmButtonText: '重新加载', cancelButtonText: '取消', type: 'warning' }) } catch { return }
  if (busy.value || saving.value) return
  operation.value = 'reload'; clearTimeout(timer); ++requestEpoch
  try {
    const result = await getAdminArchiveWorkspace(batchID)
    applyResponse(result)
    conflict.value = false; dirty.value = false; saveError.value = ''; errorText.value = ''
  } catch (e: any) { ElMessage.error(e.message || '重新加载失败，当前草稿已保留') }
  finally { operation.value = '' }
}
function chooseGroup(node: any) { currentGroup.value = node.id; editor.value = 'group'; groupView.value = 'members'; memberSelection.value = []; collection.open = false }
function editGroup(id = '') {
  if (!draft.value || busy.value || (id && groupLocked(id))) return
  const target = draft.value.groups.find((g: any) => g.id === id)
  if (id && !target) return
  groupDialog.id = id; groupDialog.name = target?.name || ''
  groupDialog.config = cloneWorkspace(target?.config || draft.value.template || {})
  groupDialog.open = true
}
function showGroupMenu(event: MouseEvent, id: string) { event.preventDefault(); event.stopPropagation(); groupMenu.id = id; groupMenu.x = Math.max(0, Math.min(event.clientX, window.innerWidth - 210)); groupMenu.y = Math.max(0, Math.min(event.clientY, window.innerHeight - 150)); groupMenu.open = true }
function toggleMember(id: string, checked: boolean) { memberSelection.value = checked ? [...new Set([...memberSelection.value, id])] : memberSelection.value.filter((f) => f !== id) }

function chooseFile(row: any) { if (row.isDirectory) { currentDirectory.value = row.id; search.value = ''; return } fileID.value = row.id; editor.value = 'file'; collection.open = false }
function showContext(event: MouseEvent, row?: any) {
  event.preventDefault(); event.stopPropagation(); contextMenu.open = true; contextMenu.x = Math.min(event.clientX, window.innerWidth - 210); contextMenu.y = Math.min(event.clientY, window.innerHeight - 250)
  contextMenu.directory = row?.isDirectory ? row.id : ''
  if (row && !selected.value.some((s: any) => s.rowID === row.rowID)) { fileTable.value?.clearSelection(); fileTable.value?.toggleRowSelection(row, true); selected.value = [row] }
}
function setExcluded(excluded: boolean) {
  const ids = selectedIDs.value.filter((id: string) => !fileLocked(id))
  for (const id of ids) draft.value.files[id].excluded = excluded
  contextMenu.open = false
}
function assignToGroup(id = targetGroup.value, ids = selectedIDs.value) {
  if (id && !draft.value.groups.some((g: any) => g.id === id)) { targetGroup.value = ''; ElMessage.warning('目标分组已不存在，请重新选择'); return }
  if (id && groupLocked(id)) { ElMessage.warning('目标分组正在入库，暂不能加入'); return }
  let changed = 0
  for (const fid of ids) if (!fileMoveLocked(fid) && ['video', 'image'].includes((sources.value.get(fid) as any)?.media_kind)) { draft.value.files[fid].group_id = id; changed++ }
  if (!changed) ElMessage.info('没有可调整归属的文件')
  contextMenu.open = false
}
function startDrag(event: DragEvent, row: any) {
  const current = selected.value.some((s: any) => s.rowID === row.rowID) ? selected.value : [row]
  const ids = selectedWorkspaceFiles(draft.value, data.value.files, current.filter((r: any) => !r.isDirectory).map((r: any) => r.id), current.filter((r: any) => r.isDirectory).map((r: any) => r.id))
  event.dataTransfer?.setData('application/x-archive-files', JSON.stringify({ batch_id: batchID, ids }))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}
function dropGroup(event: DragEvent, id: string) {
  event.preventDefault()
  try { const payload = JSON.parse(event.dataTransfer?.getData('application/x-archive-files') || '{}'); if (payload.batch_id === batchID && Array.isArray(payload.ids)) assignToGroup(id, payload.ids) } catch { ElMessage.warning('无法识别拖入的文件') }
}
function applyBulkConfig(clear = false) {
  const ids = selectedIDs.value.filter((id: string) => !fileLocked(id))
  for (const id of ids) draft.value.files[id].config = clear ? {} : { ...draft.value.files[id].config, ...cloneWorkspace(bulkConfig.value) }
  ElMessage.success(`已更新 ${ids.length} 个文件，未设置的批量字段保留原值`)
}
function archiveTitleToDescription() {
  const ids = editor.value === 'file' ? [fileID.value] : selectedIDs.value
  const eligible = ids.filter((id: string) => !fileLocked(id) && (sources.value.get(id) as any)?.media_kind === 'video')
  for (const id of eligible) { if (!fileTitle(draft.value.files[id].name)) { ElMessage.warning('存在清理后标题为空的文件，本次未执行'); return } }
  for (const id of eligible) {
    const old = effective.value[id], title = fileTitle(draft.value.files[id].name), previous = String(old.title || '').trim()
    const description = previous && previous !== title ? `${previous}${old.description ? `\n${old.description}` : ''}` : old.description
    draft.value.files[id].config = { ...draft.value.files[id].config, title_mode: 'uniform', title, description }
  }
  ElMessage.success(`已为 ${eligible.length} 个视频替换标题并保留原说明`)
}
function cutSelection() { cut.value = { ids: selected.value.filter((r: any) => !r.isDirectory).map((r: any) => r.id), dirs: selected.value.filter((r: any) => r.isDirectory).map((r: any) => r.id) }; contextMenu.open = false }
function pasteSelection(target = currentDirectory.value) {
  if (!cut.value) return
  for (const id of cut.value.dirs) {
    if (descendantIDs(draft.value.directories, id).has(target)) { ElMessage.warning('不能将目录移入自身或子目录'); return }
    const ids = selectedWorkspaceFiles(draft.value, data.value.files, [], [id]); if (ids.some((fid: string) => activeIDs.value.has(fid))) { ElMessage.warning('目录包含处理中文件，暂不能移动'); return }
  }
  for (const id of cut.value.ids) if (!activeIDs.value.has(id)) draft.value.files[id].directory_id = target
  for (const id of cut.value.dirs) { const dir = draft.value.directories.find((d: any) => d.id === id); if (dir) dir.parent_id = target }
  cut.value = null; contextMenu.open = false
}
function addDirectory() {
  const name = newDirectory.name.trim(); if (!name) return
  if (draft.value.directories.some((d: any) => d.parent_id === currentDirectory.value && d.name.toLowerCase() === name.toLowerCase()) || Object.values(draft.value.files).some((f: any) => f.directory_id === currentDirectory.value && f.name.toLowerCase() === name.toLowerCase())) { ElMessage.warning('同目录已有此名称'); return }
  draft.value.directories.push({ id: newWorkspaceID(), parent_id: currentDirectory.value, name }); newDirectory.open = false; newDirectory.name = ''
}
async function saveGroup() {
  if (!draft.value || busy.value || saving.value || conflict.value) return
  const name = groupDialog.name.trim()
  if (!name) { ElMessage.warning('请填写分组名称'); return }
  if (groupDialog.id && groupLocked(groupDialog.id)) { ElMessage.warning('分组正在入库，暂不能修改'); return }
  if (draft.value.groups.some((g: any) => g.id !== groupDialog.id && g.name.toLowerCase() === name.toLowerCase())) { ElMessage.warning('分组不能重名'); return }
  let target = draft.value.groups.find((g: any) => g.id === groupDialog.id)
  if (groupDialog.id && !target) { ElMessage.warning('分组已不存在，请关闭后重新选择'); return }
  if (!target) {
    target = { id: newWorkspaceID(), parent_id: '', name, config: {} }
    draft.value.groups.push(target)
    groupDialog.id = target.id
    target = draft.value.groups.find((g: any) => g.id === groupDialog.id)
  }
  target.name = name; target.config = cloneWorkspace(groupDialog.config)
  if (await save()) { groupDialog.open = false; ElMessage.success('分组已保存') }
}
async function deleteGroup(id = currentGroup.value) {
  if (!id || groupLocked(id)) return
  try { await ElMessageBox.confirm('删除分组后成员回到未分组。源文件、已入库媒体、合集及历史关联保留。', '删除分组', { type: 'warning' }) } catch { return }
  if (groupLocked(id)) return
  for (const file of Object.values(draft.value.files) as any[]) if (file.group_id === id) file.group_id = ''
  draft.value.groups = draft.value.groups.filter((g: any) => g.id !== id)
  if (targetGroup.value === id) targetGroup.value = ''
  if (currentGroup.value === id) { currentGroup.value = ''; groupView.value = 'desktop'; editor.value = 'group' }
}
async function generateGroups() {
  if (busy.value || saving.value) return
  operation.value = 'generate'; contextMenu.open = false
  try {
    if (!await save(true)) return
    ++requestEpoch
    const result = await generateAdminArchiveWorkspaceGroups(batchID, { version: data.value.version })
    applyResponse(result)
    ElMessage.success('已补充缺失的目录分组，已有归属保留')
  } catch (e: any) { if (e.code === 409) conflict.value = true; ElMessage.error(e.message) }
  finally { ++requestEpoch; operation.value = '' }
}
function renameSelection() { const row = selected.value[0]; if (!row) return; if (row.isDirectory) { contextMenu.directory = row.id } else chooseFile(row); contextMenu.open = false }
function openCollection(kind: string) { collection.kind = kind; collection.target = editor.value === 'file' ? fileID.value : currentGroup.value; collection.targetKind = groupDialog.open ? 'group-dialog' : editor.value; collection.name = ''; collection.description = ''; collection.open = true }
async function createCollection() {
  if (!collection.name.trim() || busy.value) return
  operation.value = 'collection'
  try {
    const result = await (collection.kind === 'video' ? createAdminCollection : createAdminImageCollection)({ name: collection.name.trim(), description: collection.description })
    const target = collection.targetKind === 'group-dialog' ? (groupDialog.open ? groupDialog : null) : collection.targetKind === 'batch' ? { config: bulkConfig.value } : collection.targetKind === 'file' ? draft.value.files[collection.target] : draft.value.groups.find((g: any) => g.id === collection.target)
    if (!target || (collection.targetKind === 'file' && fileLocked(collection.target))) { ElMessage.warning('合集已创建，当前编辑对象已改变，请重新选择'); collection.open = false; return }
    const key = collection.kind === 'video' ? 'video_collection_ids' : 'image_collection_ids'; target.config[key] = [...new Set([...(target.config[key] || []), result.id])]
    const list = collection.kind === 'video' ? videoOptions : imageOptions; list.value = [...list.value, { id: result.id, name: collection.name.trim() }]
    collection.open = false; ElMessage.success('合集已创建并选中；取消入库仍会保留')
  } catch (e: any) { ElMessage.error(e.message) }
  finally { operation.value = '' }
}
let videoQuery = 0, imageQuery = 0, tagQuery = 0
async function searchVideo(q = '') { const n = ++videoQuery; try { const r = await getAdminCollections({ q, page: 1, page_size: q ? 12 : 5 }); if (n === videoQuery) videoOptions.value = r.items || [] } catch { /* keep the selected options */ } }
async function searchImage(q = '') { const n = ++imageQuery; try { const r = await getAdminImageCollections({ q, page: 1, page_size: q ? 12 : 5 }); if (n === imageQuery) imageOptions.value = r.items || [] } catch { /* keep the selected options */ } }
async function searchTags(q = '') { const n = ++tagQuery; try { const r = await (q ? getAdminVideoTags({ q, limit: 12 }) : getAdminPopularVideoTags({ limit: 5 })); if (n === tagQuery) tagOptions.value = (r.items || r || []).map((t: any) => typeof t === 'string' ? t : t.name || t.tag) } catch { /* retain current choices */ } }
async function submit(ids: string[]) {
  if (busy.value || saving.value) { ElMessage.info('正在处理，请稍后提交'); return }
  operation.value = 'submit'; contextMenu.open = false
  try {
    if (!await save(true)) { ElMessage.warning('请先解决保存失败或冲突'); return }
    const confirmedSerial = editSerial
    const confirmedVersion = data.value.version
    const summary = summarizeWorkspaceImport({ ...data.value, draft: draft.value, effective: effective.value }, ids)
    if (!summary.ids.length) { ElMessage.info('没有可入库文件'); return }
    if (summary.missingVideoCollections || summary.missingImageCollections) { ElMessage.warning('请先补齐分组的目标合集：视频需要视频合集，图片需要图片合集'); return }
    if (summary.emptyTitles.length) { ElMessage.warning(`以下文件标题为空：${summary.emptyTitles.join('、')}`); return }
    const selectedGroups = draft.value.groups.filter((g: any) => summary.ids.some((id: string) => draft.value.files[id].group_id === g.id))
    const pairs = selectedGroups.map((g: any) => `${g.name}：${(g.config.video_collection_ids || []).length} 个视频合集 × ${(g.config.image_collection_ids || []).length} 个图片合集`)
    const warnings = [`未分组 ${ungroupedCount.value} 个文件不参与入库`, ...pairs, '以上合集将两两追加关联，对合集及原有成员生效；已入库媒体只补关系，不覆盖内容']
    try { await ElMessageBox.confirm(`将入库 ${summary.videos} 个视频、${summary.images} 张图片。${warnings.join('；')}。确认后后台持续处理，离开页面不中断。`, '入库预览', { confirmButtonText: '确认入库', cancelButtonText: '取消' }) } catch { return }
    if (confirmedSerial !== editSerial || confirmedVersion !== data.value.version || dirty.value || saving.value) { ElMessage.warning('配置已变化，请重新预览并提交'); return }
    ++requestEpoch
    try { const result = await submitAdminArchiveWorkspace(batchID, { version: data.value.version, file_ids: summary.ids }); applyResponse(result, false); ElMessage.success('已提交，后台将持续处理') } catch (e: any) { if (e.code === 409) conflict.value = true; ElMessage.error(e.message) }
  } finally { ++requestEpoch; operation.value = '' }
}
async function retryExtract() {
  if (busy.value || saving.value) return
  operation.value = 'retry'; contextMenu.open = false
  let refreshAfterAttempt = false
  try {
    try { await ElMessageBox.confirm('重新解压会重置虚拟目录、分组和文件配置。有成功文件或正在入库时不允许执行。', '重新解压', { type: 'warning' }) } catch { return }
    if (draft.value && !await save(true)) { ElMessage.warning('请先解决保存失败或冲突'); return }
    retrying.value = true; clearTimeout(timer); ++requestEpoch; refreshAfterAttempt = true
    await retryAdminArchiveImportExtract(batchID, retry)
    clearWorkspace(); dirty.value = false; conflict.value = false; saveError.value = ''
  } catch (e: any) { ElMessage.error(e.message) }
  finally { ++requestEpoch; operation.value = ''; retrying.value = false }
  if (refreshAfterAttempt) await load(true)
}
const canRetryExtract = computed(() => batch.value && !['extracting', 'processing'].includes(batch.value.status) && !activeIDs.value.size && !(data.value?.files || []).some((f: any) => ['ready', 'existing', 'processing'].includes(f.status) || f.linked_video_id || f.linked_image_id))
function beforeUnload(e: BeforeUnloadEvent) { if (dirty.value || saving.value) { e.preventDefault(); e.returnValue = '' } }
onBeforeRouteLeave(async () => { if (!dirty.value && !saving.value) return true; if (await save()) return true; try { await ElMessageBox.confirm('当前草稿尚未保存，离开会丢失本地编辑。可取消并下载草稿。', '离开工作台', { type: 'warning' }); return true } catch { return false } })
onMounted(() => { void load(true); void searchVideo(); void searchImage(); void searchTags(); poll = setInterval(() => { void load() }, 4000); window.addEventListener('beforeunload', beforeUnload) })
onUnmounted(() => { clearTimeout(timer); clearInterval(poll); window.removeEventListener('beforeunload', beforeUnload) })
</script>

<template>
  <main class="archive-workspace" @click="contextMenu.open = false; groupMenu.open = false">
    <header class="workspace-header">
      <div class="header-title"><el-button :icon="Back" circle aria-label="返回批次列表" @click="router.push('/toolbox/archive-import')" /><div><small>压缩包入库工作台</small><h1>{{ batch?.title || '正在加载批次' }}</h1></div></div>
      <div class="header-actions"><el-tag v-if="draft" :type="saveError || conflict ? 'danger' : dirty || saving ? 'warning' : 'success'">{{ saveLabel }}</el-tag><el-button v-if="draft" :loading="saving" :disabled="conflict || busy" @click="save()">保存 / 重试</el-button><el-button :icon="Refresh" :disabled="dirty || saving || busy" @click="load(true)">刷新</el-button><el-button v-if="draft" type="primary" :icon="UploadFilled" :disabled="saving || conflict || busy" :loading="operation === 'submit'" @click="submit(Object.keys(draft.files))">全部入库</el-button></div>
    </header>
    <el-alert v-if="saveError || conflict" :title="saveError || '保存冲突'" type="error" :closable="false" show-icon><el-button text @click="exportDraft">下载当前草稿</el-button><el-button text @click="reloadDraft">重新加载服务器草稿</el-button></el-alert>
    <div v-if="!draft" v-loading="loading" class="workspace-waiting"><el-icon :size="42"><Folder /></el-icon><h2>{{ batch?.status === 'extracting' ? '正在解压，完成后自动打开工作台' : '等待完成解压' }}</h2><p v-if="errorText">{{ errorText }}</p><div v-if="canRetryExtract" class="retry-form"><el-input v-model="retry.password" type="password" show-password placeholder="压缩包密码（可选）" /><el-select v-model="retry.encoding_mode" aria-label="文件名编码"><el-option label="自动识别编码" value="auto" /><el-option label="UTF-8" value="utf8" /><el-option label="GBK" value="gbk" /></el-select><el-button :loading="retrying" @click="retryExtract">重新解压</el-button></div></div>
    <el-alert v-if="draft && draft.layout_version < 2" title="旧批次正在执行任务，结束后将自动转换为平级分组" type="info" :closable="false" />
    <div v-if="draft" class="workspace-columns" :inert="busy || draft.layout_version < 2" :aria-busy="busy">
      <section class="files-pane" aria-label="虚拟文件管理器">
        <div class="pane-heading"><div><h2>文件管理器</h2><small>仅整理虚拟关系，原有物理文件保持不变</small></div><el-button :icon="Plus" @click="newDirectory.open = !newDirectory.open">新建文件夹</el-button></div>
        <div v-if="newDirectory.open" class="inline-create"><el-input v-model="newDirectory.name" maxlength="200" placeholder="文件夹名称" @keyup.enter="addDirectory" /><el-button type="primary" @click="addDirectory">创建</el-button></div>
        <div class="file-toolbar"><el-button :icon="Folder" :aria-expanded="directoryTreeVisible" aria-controls="archive-directory-tree" :type="directoryTreeVisible ? 'primary' : undefined" @click="directoryTreeVisible = !directoryTreeVisible">目录</el-button><el-input v-model="search" clearable placeholder="搜索文件名称" aria-label="搜索文件" /><el-button :disabled="!selected.length" @click="cutSelection">剪切</el-button><el-button :disabled="!cut" @click="pasteSelection()">粘贴到此目录</el-button></div>
        <div class="file-browser">
          <el-tree v-if="directoryTreeVisible" id="archive-directory-tree" class="directory-tree" aria-label="目录导航" :current-node-key="currentDirectory" :data="directoryTree" node-key="id" default-expand-all highlight-current :expand-on-click-node="false" @node-click="(node: any) => { currentDirectory = node.id; search = '' }" />
          <div class="file-list" @contextmenu="showContext($event)"><nav class="breadcrumb" aria-label="当前目录路径"><template v-for="(directory, index) in directoryPath" :key="directory.id"><span v-if="index" aria-hidden="true">/</span><el-button text :aria-current="directory.id === currentDirectory ? 'page' : undefined" @click="currentDirectory = directory.id; search = ''">{{ directory.name }}</el-button></template></nav>
            <el-table ref="fileTable" :data="rows" row-key="rowID" height="100%" @selection-change="selected = $event" @row-dblclick="chooseFile" @row-contextmenu="(row: any, _c: any, event: MouseEvent) => showContext(event, row)">
              <el-table-column type="selection" width="42" reserve-selection />
              <el-table-column label="名称" min-width="200"><template #default="{ row }"><div class="file-name" draggable="true" @dragstart="startDrag($event, row)"><el-icon><Folder v-if="row.isDirectory" /><Document v-else /></el-icon><button class="file-link" @click="chooseFile(row)">{{ row.name }}</button></div></template></el-table-column>
              <el-table-column label="分组" min-width="110"><template #default="{ row }">{{ row.isDirectory ? '文件夹' : draft.groups.find((g: any) => g.id === row.group_id)?.name || '未分组' }}</template></el-table-column>
              <el-table-column label="状态" min-width="100"><template #default="{ row }"><el-tag v-if="!row.isDirectory" :type="row.job?.state === 'failed' ? 'danger' : row.excluded ? 'info' : undefined" size="small">{{ statusLabel(row) }}</el-tag></template></el-table-column>
            </el-table>
          </div>
        </div>
        <div v-if="contextMenu.directory && draft.directories.some((d: any) => d.id === contextMenu.directory)" class="inline-create"><span>目录名称</span><el-input v-model="draft.directories.find((d: any) => d.id === contextMenu.directory).name" :disabled="selectedWorkspaceFiles(draft, data.files, [], [contextMenu.directory]).some((id: string) => activeIDs.has(id))" maxlength="200" /><el-button @click="contextMenu.directory = ''">完成</el-button></div>
        <div class="selection-toolbar"><span>已选 {{ selectedIDs.length }} 个文件</span><el-select v-model="targetGroup" aria-label="批量目标分组"><el-option label="未分组" value="" /><el-option v-for="g in draft.groups" :key="g.id" :label="g.name" :value="g.id" /></el-select><el-button :disabled="!selectedIDs.length" @click="assignToGroup()">加入分组</el-button><el-button :disabled="!selectedIDs.length" @click="editor = 'batch'; bulkConfig = {}">批量属性</el-button><el-button :disabled="!selectedIDs.length" @click="setExcluded(true)">排除</el-button><el-button :disabled="!selectedIDs.length" @click="setExcluded(false)">恢复</el-button><el-button type="primary" :disabled="!selectedIDs.length || conflict" @click="submit(selectedIDs)">入库所选</el-button></div>
      </section>
      <section class="groups-pane" aria-label="入库分组">
        <div class="pane-heading"><div><h2>入库分组</h2><small>拖入文件或文件夹，按媒体类型进入所选合集</small></div><el-button :icon="Plus" @click="editGroup()">新建分组</el-button></div>
        <div v-if="editor === 'group' && groupView === 'desktop'" class="group-desktop">
          <article class="group-folder" @dblclick="chooseGroup({ id: '' })" @dragover.prevent @drop="dropGroup($event, '')">
            <button class="folder-open" @dblclick.stop="chooseGroup({ id: '' })" @keydown.enter="chooseGroup({ id: '' })"><el-icon :size="42"><Folder /></el-icon><strong>未分组</strong><small>{{ ungroupedCount }} 个文件</small></button>
            <el-button text @click="chooseGroup({ id: '' })">打开</el-button>
          </article>
          <article v-for="g in draft.groups" :key="g.id" class="group-folder" :class="{ 'is-locked': groupLocked(g.id) }" @dblclick="chooseGroup(g)" @contextmenu="showGroupMenu($event, g.id)" @dragover.prevent @drop="dropGroup($event, g.id)">
            <button class="folder-open" @dblclick.stop="chooseGroup(g)" @keydown.enter="chooseGroup(g)"><el-icon :size="42"><Folder /></el-icon><strong :title="g.name">{{ g.name }}</strong><small>{{ Object.values(draft.files).filter((f: any) => f.group_id === g.id).length }} 个文件 · {{ groupStatus(g.id) }}</small></button>
            <div class="folder-actions"><el-button text @click="chooseGroup(g)">打开</el-button><el-button text :icon="MoreFilled" :aria-label="`${g.name}的操作菜单`" @click.stop="showGroupMenu($event, g.id)">操作</el-button></div>
          </article>
        </div>
        <div v-if="editor === 'group' && groupView === 'members'" class="group-members">
          <div class="group-actions"><el-button :icon="Back" @click="groupView = 'desktop'">返回分组</el-button><strong>{{ group?.name || '未分组' }}</strong><el-button v-if="group" :disabled="groupLocked(currentGroup)" @click="editGroup(currentGroup)">编辑</el-button><el-button v-if="group" type="primary" :disabled="conflict || groupLocked(currentGroup)" @click="submit(memberIDs)">入库本组</el-button></div>
          <p v-if="!group" class="muted member-note">先从左侧选择文件并加入分组，再执行入库。</p>
          <div class="group-actions"><span>{{ memberIDs.length }} 个文件 · 已选 {{ memberSelection.length }} 个</span><el-button v-if="group" :disabled="!memberSelection.length || groupLocked(currentGroup)" @click="assignToGroup('', memberSelection); memberSelection = []">移出本组</el-button></div>
          <div class="member-list"><div v-for="id in memberIDs" :key="id" class="member-row"><el-checkbox :model-value="memberSelection.includes(id)" :disabled="fileMoveLocked(id)" :aria-label="`选择${draft.files[id].name}`" @change="toggleMember(id, Boolean($event))" /><button class="file-link" @click="chooseFile({ id })">{{ draft.files[id].name }}</button><small>{{ statusLabel({ ...draft.files[id], source: sources.get(id), job: jobs.get(id) }) }}</small></div><p v-if="!memberIDs.length" class="muted member-note">暂无文件，可从左侧拖入或通过“加入分组”添加。</p></div>
        </div>
        <div v-if="editor !== 'group'" class="properties">

          <div class="properties-heading"><h3>{{ editor === 'batch' ? '批量文件属性' : editor === 'file' ? '文件属性' : currentGroup ? '分组配置' : '未分组' }}</h3><el-button v-if="editor !== 'group'" text @click="editor = 'group'; groupView = 'members'">返回分组</el-button></div>
          <template v-if="editor === 'file' && editedFile"><el-input v-model="editedFile.name" :disabled="activeIDs.has(fileID)" maxlength="200" aria-label="虚拟文件名称" /><p class="muted">原路径：{{ (sources.get(fileID) as any)?.relative_path }}</p><el-alert v-if="fileLocked(fileID)" title="已入库或处理中：入库配置锁定" :closable="false" type="info" /><p v-if="(jobs.get(fileID) as any)?.last_error" class="error-text">{{ (jobs.get(fileID) as any).last_error }}</p><p class="resolved-title">有效标题：{{ effective[fileID]?.title || '（空，入库前须补填）' }}</p><el-button v-if="!fileLocked(fileID)" @click="archiveTitleToDescription">按文件名替换标题并转存原标题</el-button></template>
          <template v-else-if="editor === 'batch'"><p class="muted">作用于当前所选 {{ selectedIDs.length }} 个文件，跳过已入库和处理中成员。勾选字段后编辑，点击应用；未勾选字段不修改。</p><el-button type="primary" @click="applyBulkConfig()">应用批量字段</el-button><el-button @click="applyBulkConfig(true)">所选文件恢复全部继承</el-button><el-button @click="archiveTitleToDescription">按文件名替换标题并转存原标题</el-button></template>
          <ArchiveWorkspaceConfig v-if="editor === 'batch' || (editor === 'file' && editedFile)" v-model="editorConfig" :root="editor === 'group'" :content-only="editor !== 'group'" :image="editor === 'file' && (sources.get(fileID) as any)?.media_kind === 'image'" :disabled="editor === 'file' ? fileLocked(fileID) : editor === 'group' && groupLocked(currentGroup)" :video-options="videoOptions" :image-options="imageOptions" :tag-options="tagOptions" @search-video="searchVideo" @search-image="searchImage" @search-tags="searchTags" @create="openCollection" />
          <details v-if="canRetryExtract" class="reextract"><summary>重新解压并重置整理</summary><el-input v-model="retry.password" type="password" show-password placeholder="密码（可选）" /><el-button type="danger" plain :loading="retrying" @click="retryExtract">确认重置并重新解压</el-button></details>
        </div>
      </section>
    </div>
    <el-dialog v-model="groupDialog.open" :title="groupDialog.id ? '编辑分组' : '新建分组'" width="min(820px, 94vw)" append-to-body destroy-on-close :close-on-click-modal="false" :close-on-press-escape="!saving && !collection.open" :show-close="!saving && !collection.open" :before-close="(done: () => void) => { if (!saving && !collection.open) done() }">
      <div class="group-dialog-form" :inert="busy || saving || collection.open">
        <el-form label-position="top"><el-form-item label="分组名称" required><el-input v-model="groupDialog.name" maxlength="200" aria-label="分组名称" placeholder="请输入分组名称" :disabled="groupLocked(groupDialog.id)" /></el-form-item></el-form>
        <p class="muted">本组配置独立保存。视频与图片分别加入对应合集，入库时两类合集会两两追加关联。</p>
        <el-alert v-if="groupLocked(groupDialog.id)" title="分组正在入库，暂不能修改" type="info" :closable="false" />
        <ArchiveWorkspaceConfig v-model="groupDialog.config" root :disabled="groupLocked(groupDialog.id)" :video-options="videoOptions" :image-options="imageOptions" :tag-options="tagOptions" @search-video="searchVideo" @search-image="searchImage" @search-tags="searchTags" @create="openCollection" />
        <el-alert v-if="saveError || conflict" :title="saveError || '保存冲突，请关闭弹窗并重新加载草稿'" type="error" :closable="false" />
      </div>
      <template #footer><el-button :disabled="saving || busy || collection.open" @click="groupDialog.open = false">取消</el-button><el-button type="primary" :loading="saving" :disabled="busy || conflict || collection.open || groupLocked(groupDialog.id)" @click="saveGroup">保存分组</el-button></template>
    </el-dialog>
    <el-dialog v-model="collection.open" :title="`新建${collection.kind === 'video' ? '视频' : '图片'}合集`" width="min(520px, 94vw)" append-to-body destroy-on-close :close-on-click-modal="false" :close-on-press-escape="!busy" :show-close="!busy" :before-close="(done: () => void) => { if (!busy) done() }">
      <el-form label-position="top" :disabled="busy">
        <el-form-item label="合集名称" required><el-input v-model="collection.name" placeholder="请输入合集名称" maxlength="200" @keyup.enter="createCollection" /></el-form-item>
        <el-form-item label="合集说明"><el-input v-model="collection.description" type="textarea" :rows="3" placeholder="合集说明（可选）" /></el-form-item>
      </el-form>
      <p class="muted">创建后即正式保存，取消分组编辑仍会保留此合集。</p>
      <template #footer><el-button :disabled="busy" @click="collection.open = false">取消</el-button><el-button type="primary" :loading="operation === 'collection'" :disabled="!collection.name.trim() || busy" @click="createCollection">创建并选中</el-button></template>
    </el-dialog>
    <div v-if="groupMenu.open && !busy" class="context-menu" role="menu" :style="{ left: `${groupMenu.x}px`, top: `${groupMenu.y}px` }" @click.stop>
      <button role="menuitem" :disabled="groupLocked(groupMenu.id)" @click="editGroup(groupMenu.id); groupMenu.open = false">编辑</button><button role="menuitem" :disabled="groupLocked(groupMenu.id)" @click="deleteGroup(groupMenu.id); groupMenu.open = false">删除</button><button role="menuitem" :disabled="conflict || groupLocked(groupMenu.id)" @click="submit(Object.keys(draft.files).filter((id) => draft.files[id].group_id === groupMenu.id)); groupMenu.open = false">入库</button>
    </div>
    <div v-if="contextMenu.open && !busy" class="context-menu" :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }" @click.stop>
      <button @click="renameSelection">重命名 / 文件属性</button><button @click="cutSelection">剪切</button><button :disabled="!cut" @click="pasteSelection(contextMenu.directory || currentDirectory)">粘贴到此目录</button><button @click="assignToGroup(targetGroup)">移入所选目标分组</button><button @click="setExcluded(true)">排除入库</button><button @click="setExcluded(false)">恢复入库</button><button @click="newDirectory.open = true; contextMenu.open = false">新建文件夹</button>
    </div>
  </main>
</template>

<style scoped>
.archive-workspace{height:100dvh;display:flex;flex-direction:column;background:var(--bg-canvas);color:var(--text-primary);padding:18px;gap:14px;box-sizing:border-box;overflow:hidden;}
.workspace-header{display:flex;align-items:center;justify-content:space-between;gap:16px;}
.header-title{display:flex;align-items:center;gap:14px;min-width:0;}
.header-title small{color:var(--text-muted);font-size:12px;letter-spacing:0;}
.header-title h1{font-size:21px;line-height:1.4;margin:3px 0 0;max-width:50vw;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.header-actions{display:flex;align-items:center;gap:8px;flex-wrap:wrap;}
.workspace-columns{display:grid;grid-template-columns:minmax(0,1.2fr) minmax(360px,1fr);gap:14px;flex:1;min-height:0;}
.files-pane,.groups-pane{background:var(--bg-surface);border:1px solid var(--line-soft);border-radius:8px;display:flex;flex-direction:column;min-width:0;min-height:0;overflow:hidden;}
.pane-heading{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:17px;border-bottom:1px solid var(--line-soft);}
.pane-heading h2{font-size:16px;margin:0 0 4px;}
.pane-heading small,.muted{font-size:12px;color:var(--text-muted);line-height:1.6;}
.file-toolbar,.inline-create,.selection-toolbar,.group-actions{display:flex;align-items:center;gap:8px;padding:10px 14px;flex-wrap:wrap;}
.file-toolbar .el-input{flex:1;min-width:130px;}
.inline-create .el-input{flex:1;min-width:120px;}
.file-browser{display:flex;flex:1;min-height:0;}
.directory-tree{width:155px;flex-shrink:0;overflow:auto;padding:10px 4px;background:var(--bg-surface-muted);border-right:1px solid var(--line-soft);}
.file-list{flex:1;min-width:0;display:flex;flex-direction:column;}
.breadcrumb{min-height:34px;flex-shrink:0;font-size:12px;padding:0 8px;display:flex;align-items:center;overflow-x:auto;white-space:nowrap;}
.breadcrumb .el-button{flex-shrink:0;margin-left:0;}
.file-name{display:flex;align-items:center;gap:7px;min-width:0;}
.file-name .el-icon{color:var(--text-muted);}
.file-link{background:none;border:0;text-align:left;color:inherit;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding:7px 0;font-size:13px;}
.selection-toolbar{border-top:1px solid var(--line-soft);font-size:12px;}
.selection-toolbar .el-select{width:130px;}
.group-tree{min-height:90px;max-height:24vh;overflow:auto;padding:10px 12px;background:var(--bg-surface-muted);}
.group-drop{display:flex;align-items:center;gap:8px;width:100%;padding:6px 0;min-height:24px;}
.group-drop small{margin-left:auto;margin-right:12px;color:var(--text-muted);}
.group-drop:hover{background:var(--bg-surface-strong);}
.group-actions{border-bottom:1px solid var(--line-soft);}
.group-desktop{display:grid;grid-template-columns:repeat(auto-fill,minmax(145px,1fr));gap:12px;padding:16px;overflow:auto;align-content:start;}
.group-folder{border:1px solid var(--line-soft);border-radius:8px;padding:10px;min-width:0;}
.group-folder:hover{background:var(--bg-surface-muted);}
.folder-open{display:flex;flex-direction:column;align-items:center;gap:8px;width:100%;border:0;background:none;color:inherit;cursor:pointer;padding:8px 0;}
.folder-open .el-icon{color:var(--text-muted);}
.folder-open strong{max-width:100%;overflow-wrap:anywhere;}
.folder-open small{font-size:12px;color:var(--text-muted);}
.folder-actions{display:flex;justify-content:center;}
.group-members{display:flex;flex-direction:column;min-height:0;flex:1;}
.member-list{overflow:auto;padding:0 14px;}
.member-note{padding:0 14px;}
.member-row .file-link{flex:1;min-width:0;}
.properties{flex:1;min-height:0;overflow:auto;padding:14px 18px;}
.properties-heading{display:flex;align-items:center;justify-content:space-between;}
.properties h3{font-size:14px;margin:0 0 12px;}
.properties :deep(.el-select){width:100%;}
.properties>.el-input,.properties>.el-select{margin-bottom:10px;}
.resolved-title{font-size:12px;background:var(--bg-surface-muted);padding:10px;border-radius:6px;}
.group-dialog-form{max-height:65dvh;overflow:auto;padding:0 4px;}
.group-dialog-form>.el-alert{margin:12px 0;}
.member-preview{border-top:1px solid var(--line-soft);margin-top:18px;}
.member-preview h4{font-size:12px;color:var(--text-muted);}
.member-row{display:flex;justify-content:space-between;gap:8px;align-items:center;width:100%;border:0;border-bottom:1px solid var(--line-soft);background:var(--bg-surface);padding:9px 0;text-align:left;cursor:pointer;color:inherit;}
.member-row span,.member-row small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:50%;}
.member-row small{color:var(--text-muted);}
.context-menu{position:fixed;z-index:2500;width:205px;padding:6px;background:var(--bg-surface);border:1px solid var(--line-soft);border-radius:8px;box-shadow:none;}
.context-menu button{display:block;width:100%;background:none;border:0;padding:9px 12px;text-align:left;cursor:pointer;color:inherit;}
.context-menu button:hover{background:var(--bg-surface-strong);}
.context-menu button:disabled{color:var(--text-muted);}
.workspace-waiting{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;}
.workspace-waiting h2{font-size:18px;}
.retry-form{display:flex;gap:10px;max-width:650px;}
.error-text{color:var(--danger-600);font-size:12px;}
.reextract{margin-top:24px;color:var(--text-muted);font-size:12px;}
.reextract .el-input{margin:12px 0;}

@media(max-width:1000px){.archive-workspace{padding:10px;overflow:auto;height:auto;min-height:100dvh;}
.workspace-header{align-items:flex-start;flex-direction:column;}
.workspace-columns{grid-template-columns:1fr;flex:none;}
.files-pane{height:65vh;min-height:420px;}
.groups-pane{min-height:600px;}
.properties{max-height:none;overflow:visible;}
.directory-tree{width:130px;}
.group-tree{max-height:260px;}
.retry-form{flex-direction:column;}
;}

</style>
