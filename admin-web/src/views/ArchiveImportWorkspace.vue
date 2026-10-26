<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Back, Refresh, Folder, Document, Plus, Delete, UploadFilled } from '@element-plus/icons-vue'
import ArchiveWorkspaceConfig from '../components/ArchiveWorkspaceConfig.vue'
import { getAdminArchiveImportBatchDetail, getAdminArchiveWorkspace, saveAdminArchiveWorkspace, submitAdminArchiveWorkspace, generateAdminArchiveWorkspaceGroups, retryAdminArchiveImportExtract, getAdminCollections, getAdminImageCollections, createAdminCollection, createAdminImageCollection, getAdminVideoTags, getAdminPopularVideoTags } from '../api/admin'
import { cloneWorkspace, descendantIDs, workspaceTree, selectedWorkspaceFiles, summarizeWorkspaceImport, resolveWorkspaceValues, fileTitle, newWorkspaceID, MAIN_GROUP } from './archiveWorkspace.helpers'

const route = useRoute(), router = useRouter(), batchID = String(route.params.id)
const data = ref<any>(null), draft = ref<any>(null), batch = ref<any>(null)
const loading = ref(true), errorText = ref(''), dirty = ref(false), saving = ref(false), conflict = ref(false), saveError = ref('')
const currentDirectory = ref(''), currentGroup = ref(MAIN_GROUP), selected = ref<any[]>([]), editor = ref<'group' | 'file' | 'batch'>('group'), fileID = ref('')
const operation = ref(''), busy = computed(() => !!operation.value), fileTable = ref<any>(null)
const bulkConfig = ref<Record<string, any>>({})
const search = ref(''), cut = ref<{ ids: string[]; dirs: string[] } | null>(null), targetGroup = ref(MAIN_GROUP)
const newDirectory = reactive({ open: false, name: '' }), newGroup = reactive({ open: false, name: '' })
const collection = reactive({ open: false, kind: 'video', name: '', description: '', target: '' as string, targetKind: 'group' as 'group' | 'file' | 'batch' })
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
const groupLocked = (id: string) => {
  const descendants = descendantIDs(draft.value?.groups || [], id)
  return Object.entries(draft.value?.files || {}).some(([fid, f]: any) => descendants.has(f.group_id) && fileLocked(fid))
}
const directoryTree = computed(() => workspaceTree(draft.value?.directories || []))
const groupTree = computed(() => [...workspaceTree(draft.value?.groups || [], MAIN_GROUP), { id: '', label: '未分组', children: [] }])
const rows = computed(() => {
  if (!draft.value) return []
  const dirs = draft.value.directories.filter((d: any) => !search.value && d.parent_id === currentDirectory.value).map((d: any) => ({ ...d, rowID: `dir:${d.id}`, isDirectory: true }))
  const files = Object.entries(draft.value.files).filter(([, f]: any) => search.value ? f.name.toLowerCase().includes(search.value.toLowerCase()) : f.directory_id === currentDirectory.value).map(([id, f]: any) => ({ ...f, id, rowID: id, source: sources.value.get(id), job: jobs.value.get(id) }))
  return [...dirs, ...files]
})
const memberIDs = computed(() => {
  if (!draft.value) return []
  const descendants = descendantIDs(draft.value.groups, currentGroup.value)
  return Object.keys(draft.value.files).filter((id) => currentGroup.value === '' ? !draft.value.files[id].group_id : descendants.has(draft.value.files[id].group_id))
})
const selectedIDs = computed(() => draft.value ? selectedWorkspaceFiles(draft.value, data.value.files, selected.value.filter((r: any) => !r.isDirectory).map((r: any) => r.id), selected.value.filter((r: any) => r.isDirectory).map((r: any) => r.id)) : [])
const editorConfig = computed({ get: () => editor.value === 'batch' ? bulkConfig.value : editor.value === 'file' ? editedFile.value?.config || {} : group.value?.config || {}, set: (value) => { if (editor.value === 'batch') bulkConfig.value = value; else if (editor.value === 'file' && editedFile.value) editedFile.value.config = value; else if (group.value) group.value.config = value } })
const effective = computed(() => {
  const result: any = {}
  if (draft.value) for (const f of data.value.files) if (draft.value.files[f.id]) result[f.id] = fileLocked(f.id) && data.value.effective[f.id] ? data.value.effective[f.id] : resolveWorkspaceValues(draft.value, f.id, f.media_kind)
  return result
})
const directoryLabel = computed(() => currentDirectory.value ? draft.value?.directories.find((d: any) => d.id === currentDirectory.value)?.name || '目录' : '根目录')
const saveLabel = computed(() => conflict.value ? '保存冲突' : saving.value ? '正在保存…' : saveError.value ? '保存失败' : dirty.value ? '等待保存…' : '已自动保存')
const statusLabel = (row: any) => row.excluded ? '已排除' : row.job?.state === 'queued' ? '等待入库' : row.job?.state === 'running' ? '正在入库' : row.job?.state === 'failed' ? '入库失败' : ({ pending: '待入库', ready: '已入库', existing: '复用已有', processing: '处理中', failed: '失败', skipped: '不支持' } as any)[row.source?.status] || row.source?.status

function applyResponse(result: any, replaceDraft = true) {
  if (data.value && result.version < data.value.version) return
  data.value = result; batch.value = result.batch
  if (replaceDraft) { muted = true; draft.value = cloneWorkspace(result.draft); muted = false }
}
function clearWorkspace() {
  muted = true; draft.value = null; muted = false
  data.value = null; selected.value = []; currentDirectory.value = ''; currentGroup.value = MAIN_GROUP; editor.value = 'group'; fileID.value = ''
  contextMenu.directory = ''; contextMenu.open = false; cut.value = null; targetGroup.value = MAIN_GROUP
  newDirectory.open = false; newGroup.open = false; collection.open = false; bulkConfig.value = {}
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
function chooseGroup(node: any) { currentGroup.value = node.id; editor.value = 'group' }
function chooseFile(row: any) { if (row.isDirectory) { currentDirectory.value = row.id; search.value = ''; return } fileID.value = row.id; editor.value = 'file' }
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
  let changed = 0
  for (const fid of ids) if (!fileLocked(fid) && ['video', 'image'].includes((sources.value.get(fid) as any)?.media_kind)) { draft.value.files[fid].group_id = id; changed++ }
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
function addGroup() { const name = newGroup.name.trim(); if (!name) return; const parent = currentGroup.value || MAIN_GROUP; if (draft.value.groups.some((g: any) => g.parent_id === parent && g.name.toLowerCase() === name.toLowerCase())) { ElMessage.warning('同级分组不能重名'); return }; const id = newWorkspaceID(); draft.value.groups.push({ id, parent_id: parent, name, config: {} }); currentGroup.value = id; editor.value = 'group'; newGroup.open = false; newGroup.name = '' }
async function deleteGroup() {
  if (currentGroup.value === MAIN_GROUP || !currentGroup.value || groupLocked(currentGroup.value)) return
  try { await ElMessageBox.confirm('将删除本组及全部子组，成员回到未分组并保留单文件覆盖，源文件不变。', '删除分组', { type: 'warning' }) } catch { return }
  const ids = descendantIDs(draft.value.groups, currentGroup.value)
  for (const file of Object.values(draft.value.files) as any[]) if (ids.has(file.group_id)) file.group_id = ''
  draft.value.groups = draft.value.groups.filter((g: any) => !ids.has(g.id)); currentGroup.value = MAIN_GROUP
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
function openCollection(kind: string) { collection.kind = kind; collection.target = editor.value === 'file' ? fileID.value : currentGroup.value; collection.targetKind = editor.value; collection.name = ''; collection.description = ''; collection.open = true }
async function createCollection() {
  if (!collection.name.trim() || busy.value) return
  operation.value = 'collection'
  try {
    const result = await (collection.kind === 'video' ? createAdminCollection : createAdminImageCollection)({ name: collection.name.trim(), description: collection.description })
    const target = collection.targetKind === 'batch' ? { config: bulkConfig.value } : collection.targetKind === 'file' ? draft.value.files[collection.target] : draft.value.groups.find((g: any) => g.id === collection.target)
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
    if (summary.emptyTitles.length) { ElMessage.warning(`以下文件标题为空：${summary.emptyTitles.join('、')}`); return }
    const warnings = [`${summary.missingVideoCollections} 个视频未选视频合集`, `${summary.missingImageCollections} 张图片未选图片合集`, `${summary.missingVideoImageCollections} 个视频未关联图片合集`]
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
  <main class="archive-workspace" @click="contextMenu.open = false">
    <header class="workspace-header">
      <div class="header-title"><el-button :icon="Back" circle aria-label="返回批次列表" @click="router.push('/toolbox/archive-import')" /><div><small>压缩包入库工作台</small><h1>{{ batch?.title || '正在加载批次' }}</h1></div></div>
      <div class="header-actions"><el-tag v-if="draft" :type="saveError || conflict ? 'danger' : dirty || saving ? 'warning' : 'success'">{{ saveLabel }}</el-tag><el-button v-if="draft" :loading="saving" :disabled="conflict || busy" @click="save()">保存 / 重试</el-button><el-button :icon="Refresh" :disabled="dirty || saving || busy" @click="load(true)">刷新</el-button><el-button v-if="draft" type="primary" :icon="UploadFilled" :disabled="saving || conflict || busy" :loading="operation === 'submit'" @click="submit(Object.keys(draft.files))">全部入库</el-button></div>
    </header>
    <el-alert v-if="saveError || conflict" :title="saveError || '保存冲突'" type="error" :closable="false" show-icon><el-button text @click="exportDraft">下载当前草稿</el-button><el-button text @click="reloadDraft">重新加载服务器草稿</el-button></el-alert>
    <div v-if="!draft" v-loading="loading" class="workspace-waiting"><el-icon :size="42"><Folder /></el-icon><h2>{{ batch?.status === 'extracting' ? '正在解压，完成后自动打开工作台' : '等待完成解压' }}</h2><p v-if="errorText">{{ errorText }}</p><div v-if="canRetryExtract" class="retry-form"><el-input v-model="retry.password" type="password" show-password placeholder="压缩包密码（可选）" /><el-select v-model="retry.encoding_mode" aria-label="文件名编码"><el-option label="自动识别编码" value="auto" /><el-option label="UTF-8" value="utf8" /><el-option label="GBK" value="gbk" /></el-select><el-button :loading="retrying" @click="retryExtract">重新解压</el-button></div></div>
    <div v-else class="workspace-columns" :inert="busy" :aria-busy="busy">
      <section class="files-pane" aria-label="虚拟文件管理器">
        <div class="pane-heading"><div><h2>文件管理器</h2><small>仅整理虚拟关系，原有物理文件保持不变</small></div><el-button :icon="Plus" @click="newDirectory.open = !newDirectory.open">新建文件夹</el-button></div>
        <div v-if="newDirectory.open" class="inline-create"><el-input v-model="newDirectory.name" maxlength="200" placeholder="文件夹名称" @keyup.enter="addDirectory" /><el-button type="primary" @click="addDirectory">创建</el-button></div>
        <div class="file-toolbar"><el-input v-model="search" clearable placeholder="搜索文件名称" aria-label="搜索文件" /><el-button :disabled="!selected.length" @click="cutSelection">剪切</el-button><el-button :disabled="!cut" @click="pasteSelection()">粘贴到此目录</el-button></div>
        <div class="file-browser">
          <el-tree class="directory-tree" :data="directoryTree" node-key="id" default-expand-all highlight-current :expand-on-click-node="false" @node-click="(node: any) => { currentDirectory = node.id; search = '' }" />
          <div class="file-list" @contextmenu="showContext($event)"><div class="breadcrumb"><el-button text @click="currentDirectory = ''; search = ''">根目录</el-button><span v-if="currentDirectory"> / {{ directoryLabel }}</span></div>
            <el-table ref="fileTable" :data="rows" row-key="rowID" height="100%" @selection-change="selected = $event" @row-dblclick="chooseFile" @row-contextmenu="(row: any, _c: any, event: MouseEvent) => showContext(event, row)">
              <el-table-column type="selection" width="42" reserve-selection />
              <el-table-column label="名称" min-width="200"><template #default="{ row }"><div class="file-name" draggable="true" @dragstart="startDrag($event, row)"><el-icon><Folder v-if="row.isDirectory" /><Document v-else /></el-icon><button class="file-link" @click="chooseFile(row)">{{ row.name }}</button></div></template></el-table-column>
              <el-table-column label="分组" min-width="110"><template #default="{ row }">{{ row.isDirectory ? '文件夹' : draft.groups.find((g: any) => g.id === row.group_id)?.name || '未分组' }}</template></el-table-column>
              <el-table-column label="状态" min-width="100"><template #default="{ row }"><el-tag v-if="!row.isDirectory" :type="row.job?.state === 'failed' ? 'danger' : row.excluded ? 'info' : undefined" size="small">{{ statusLabel(row) }}</el-tag></template></el-table-column>
            </el-table>
          </div>
        </div>
        <div v-if="contextMenu.directory && draft.directories.some((d: any) => d.id === contextMenu.directory)" class="inline-create"><span>目录名称</span><el-input v-model="draft.directories.find((d: any) => d.id === contextMenu.directory).name" :disabled="selectedWorkspaceFiles(draft, data.files, [], [contextMenu.directory]).some((id: string) => activeIDs.has(id))" maxlength="200" /><el-button @click="contextMenu.directory = ''">完成</el-button></div>
        <div class="selection-toolbar"><span>已选 {{ selectedIDs.length }} 个文件</span><el-select v-model="targetGroup" aria-label="批量目标分组"><el-option label="未分组" value="" /><el-option v-for="g in draft.groups" :key="g.id" :label="g.name" :value="g.id" /></el-select><el-button :disabled="!selectedIDs.length" @click="assignToGroup()">移入分组</el-button><el-button :disabled="!selectedIDs.length" @click="editor = 'batch'; bulkConfig = {}">批量属性</el-button><el-button :disabled="!selectedIDs.length" @click="setExcluded(true)">排除</el-button><el-button :disabled="!selectedIDs.length" @click="setExcluded(false)">恢复</el-button><el-button type="primary" :disabled="!selectedIDs.length || conflict" @click="submit(selectedIDs)">入库所选</el-button></div>
      </section>
      <section class="groups-pane" aria-label="入库分组">
        <div class="pane-heading"><div><h2>入库分组</h2><small>拖入文件或文件夹，按媒体类型进入所选合集</small></div><el-button :icon="Plus" @click="newGroup.open = !newGroup.open">新建分组</el-button></div>
        <div v-if="newGroup.open" class="inline-create"><el-input v-model="newGroup.name" maxlength="200" placeholder="新建子分组名称" @keyup.enter="addGroup" /><el-button type="primary" @click="addGroup">创建</el-button></div>
        <el-tree class="group-tree" :data="groupTree" node-key="id" default-expand-all highlight-current :expand-on-click-node="false" @node-click="chooseGroup"><template #default="{ data: node }"><div class="group-drop" @dragover.prevent @drop="dropGroup($event, node.id)"><el-icon><Folder /></el-icon><span>{{ node.label }}</span><small>{{ Object.values(draft.files).filter((f: any) => f.group_id === node.id).length }}</small></div></template></el-tree>
        <div class="group-actions"><el-button @click="generateGroups">按文件夹补建</el-button><el-button :disabled="!currentGroup || currentGroup === MAIN_GROUP || groupLocked(currentGroup)" :icon="Delete" @click="deleteGroup">删除分组</el-button><el-button type="primary" :disabled="conflict" @click="submit(memberIDs)">入库本组</el-button></div>
        <div class="properties">
          <div class="properties-heading"><h3>{{ editor === 'batch' ? '批量文件属性' : editor === 'file' ? '文件属性' : currentGroup ? '分组配置' : '未分组' }}</h3><el-button v-if="editor !== 'group'" text @click="editor = 'group'">返回分组</el-button></div>
          <template v-if="editor === 'file' && editedFile"><el-input v-model="editedFile.name" :disabled="activeIDs.has(fileID)" maxlength="200" aria-label="虚拟文件名称" /><p class="muted">原路径：{{ (sources.get(fileID) as any)?.relative_path }}</p><el-alert v-if="fileLocked(fileID)" title="已入库或处理中：入库配置锁定" :closable="false" type="info" /><p v-if="(jobs.get(fileID) as any)?.last_error" class="error-text">{{ (jobs.get(fileID) as any).last_error }}</p><p class="resolved-title">有效标题：{{ effective[fileID]?.title || '（空，入库前须补填）' }}</p><el-button v-if="!fileLocked(fileID)" @click="archiveTitleToDescription">按文件名替换标题并转存原标题</el-button></template>
          <template v-else-if="editor === 'batch'"><p class="muted">作用于当前所选 {{ selectedIDs.length }} 个文件，跳过已入库和处理中成员。勾选字段后编辑，点击应用；未勾选字段不修改。</p><el-button type="primary" @click="applyBulkConfig()">应用批量字段</el-button><el-button @click="applyBulkConfig(true)">所选文件恢复全部继承</el-button><el-button @click="archiveTitleToDescription">按文件名替换标题并转存原标题</el-button></template>
          <template v-else-if="group"><el-input v-model="group.name" maxlength="200" aria-label="分组显示名称" /><p class="muted">{{ currentGroup === MAIN_GROUP ? '来自上传时填写的内容，普通分组逐字段继承。' : '勾选“单独设置”可覆盖，取消勾选恢复继承。集合字段覆盖会整项替换。' }}</p><el-select v-if="currentGroup !== MAIN_GROUP" v-model="group.parent_id" aria-label="父分组"><el-option v-for="g in draft.groups.filter((g: any) => !descendantIDs(draft.groups, currentGroup).has(g.id))" :key="g.id" :label="g.name" :value="g.id" /></el-select></template>
          <p v-else class="muted">未分组文件继承主分组默认值。</p>
          <ArchiveWorkspaceConfig v-if="editor === 'batch' || (editor === 'file' && editedFile) || group" v-model="editorConfig" :root="editor === 'group' && currentGroup === MAIN_GROUP" :image="editor === 'file' && (sources.get(fileID) as any)?.media_kind === 'image'" :disabled="editor === 'file' && fileLocked(fileID)" :video-options="videoOptions" :image-options="imageOptions" :tag-options="tagOptions" @search-video="searchVideo" @search-image="searchImage" @search-tags="searchTags" @create="openCollection" />
          <div v-if="collection.open" class="collection-inline"><h4>新建{{ collection.kind === 'video' ? '视频' : '图片' }}合集</h4><el-input v-model="collection.name" placeholder="合集名称" maxlength="200" /><el-input v-model="collection.description" type="textarea" placeholder="合集说明（可选）" /><p class="muted">创建后即正式保存，取消本次入库也会保留。</p><el-button @click="collection.open = false">取消</el-button><el-button type="primary" @click="createCollection">创建并选中</el-button></div>
          <div v-if="memberIDs.length" class="member-preview"><h4>当前分组及子组 · {{ memberIDs.length }} 个文件</h4><button v-for="id in memberIDs" :key="id" class="member-row" @click="fileID = id; editor = 'file'"><span>{{ draft.files[id].name }}</span><small>{{ effective[id]?.title }}</small></button></div>
          <details v-if="canRetryExtract" class="reextract"><summary>重新解压并重置整理</summary><el-input v-model="retry.password" type="password" show-password placeholder="密码（可选）" /><el-button type="danger" plain :loading="retrying" @click="retryExtract">确认重置并重新解压</el-button></details>
        </div>
      </section>
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
.breadcrumb{height:34px;font-size:12px;padding:0 8px;display:flex;align-items:center;}
.file-name{display:flex;align-items:center;gap:7px;min-width:0;}
.file-name .el-icon{color:var(--primary);}
.file-link{background:none;border:0;text-align:left;color:inherit;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding:7px 0;font-size:13px;}
.selection-toolbar{border-top:1px solid var(--line-soft);font-size:12px;}
.selection-toolbar .el-select{width:130px;}
.group-tree{min-height:90px;max-height:24vh;overflow:auto;padding:10px 12px;background:var(--bg-surface-muted);}
.group-drop{display:flex;align-items:center;gap:8px;width:100%;padding:6px 0;min-height:24px;}
.group-drop small{margin-left:auto;margin-right:12px;color:var(--text-muted);}
.group-drop:hover{background:var(--primary-soft);}
.group-actions{border-bottom:1px solid var(--line-soft);}
.properties{flex:1;min-height:0;overflow:auto;padding:14px 18px;}
.properties-heading{display:flex;align-items:center;justify-content:space-between;}
.properties h3{font-size:14px;margin:0 0 12px;}
.properties :deep(.el-select){width:100%;}
.properties>.el-input,.properties>.el-select{margin-bottom:10px;}
.resolved-title{font-size:12px;background:var(--bg-surface-muted);padding:10px;border-radius:6px;}
.collection-inline{margin-top:18px;padding:14px;background:var(--primary-soft);border:1px solid var(--line-soft);border-radius:8px;}
.collection-inline .el-input{margin-bottom:10px;}
.member-preview{border-top:1px solid var(--line-soft);margin-top:18px;}
.member-preview h4{font-size:12px;color:var(--text-muted);}
.member-row{display:flex;justify-content:space-between;gap:8px;align-items:center;width:100%;border:0;border-bottom:1px solid var(--line-soft);background:var(--bg-surface);padding:9px 0;text-align:left;cursor:pointer;color:inherit;}
.member-row span,.member-row small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:50%;}
.member-row small{color:var(--text-muted);}
.context-menu{position:fixed;z-index:2500;width:205px;padding:6px;background:var(--bg-surface);border:1px solid var(--line-soft);border-radius:8px;box-shadow:none;}
.context-menu button{display:block;width:100%;background:none;border:0;padding:9px 12px;text-align:left;cursor:pointer;color:inherit;}
.context-menu button:hover{background:var(--primary-soft);}
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
