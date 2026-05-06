<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import Layout from '../components/Layout.vue'
import {
  batchDeleteAdminVideos,
  captureAdminVideoThumbnail,
  deleteAdminVideoSubtitle,
  deleteAdminVideo,
  getAdminActors,
  getAdminCollections,
  getAdminImageCollections,
  getAdminVideoDetail,
  getAdminVideoPlayURL,
  getAdminVideoSubtitles,
  getAdminVideos,
  rescanAdminVideoSubtitles,
  retranscodeVideo,
  updateAdminVideo,
  updateAdminVideoSubtitle,
  uploadAdminVideoSubtitle
} from '../api/admin'
import { extractTvPendingDiagnostics, getVideoStatusMeta, tvPendingStageLabel } from './videoList.helpers'

const list = ref([])
const total = ref(0)
const tableRef = ref(null)
const detail = ref(null)
const detailVisible = ref(false)
const previewPlayerRef = ref(null)
const playURL = ref('')
const playExpiresAt = ref(0)
const loadingPlayURL = ref(false)
const capturingThumbnail = ref(false)
const actorOptions = ref([])
const loadingActors = ref(false)
const collectionOptions = ref([])
const loadingCollections = ref(false)
const imageCollectionOptions = ref([])
const loadingImageCollections = ref(false)
const subtitleItems = ref([])
const subtitleLoading = ref(false)
const subtitleUploadRef = ref(null)
const subtitleUploadFileList = ref([])
const subtitleUploading = ref(false)
const deletingBatch = ref(false)
const selectedRows = ref([])
const subtitleForm = reactive({
  language_code: '',
  label: '',
  is_default: false
})
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
const router = useRouter()

const query = reactive({ page: 1, page_size: 20, q: '', type: '', status: '' })
const retagTypeOptions = [
  { value: 'movie', label: '电影' },
  { value: 'episode', label: '剧集分集' },
  { value: 'av', label: 'AV' }
]

function statusLabel(status) {
  return getVideoStatusMeta(status).label
}

function typeLabel(type) {
  const map = {
    short: '短视频',
    movie: '电影',
    episode: '剧集分集',
    av: 'AV'
  }
  return map[type] || type || '-'
}

function normalizeVideoType(type) {
  return String(type || '')
    .trim()
    .toLowerCase()
}

function supportsSubtitleManage(type = detail.value?.type) {
  const normalized = normalizeVideoType(type)
  return normalized === 'movie' || normalized === 'episode' || normalized === 'av'
}

function isRetaggingShortVideo(video = detail.value) {
  if (!video || video.type !== 'short') {
    return false
  }
  const targetType = normalizeVideoType(video.target_type)
  return targetType === 'movie' || targetType === 'episode' || targetType === 'av'
}

function splitActorSelection(values) {
  const actorIDs = []
  const actorNames = []
  const seenID = new Set()
  const seenName = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value) continue
    if (uuidPattern.test(value)) {
      const key = value.toLowerCase()
      if (seenID.has(key)) continue
      seenID.add(key)
      actorIDs.push(value)
      continue
    }
    const key = value.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(value.replace(/\s+/g, ' '))
  }
  return { actorIDs, actorNames }
}

function mergeActorOptions(actors = []) {
  const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
  for (const actor of actors) {
    if (!actor?.id) continue
    optionMap.set(actor.id, { value: actor.id, label: actor.name || actor.id })
  }
  actorOptions.value = Array.from(optionMap.values())
}

function mergeCollectionOptions(collections = []) {
  const optionMap = new Map(collectionOptions.value.map((item) => [item.value, item]))
  for (const collection of collections) {
    if (!collection?.id) continue
    optionMap.set(collection.id, { value: collection.id, label: collection.name || collection.id })
  }
  collectionOptions.value = Array.from(optionMap.values())
}

function mergeImageCollectionOptions(collections = []) {
  const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
  for (const collection of collections) {
    if (!collection?.id) continue
    optionMap.set(collection.id, { value: collection.id, label: collection.name || collection.id })
  }
  imageCollectionOptions.value = Array.from(optionMap.values())
}

function normalizeCollectionSelection(values) {
  const out = []
  const seen = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value || !uuidPattern.test(value)) continue
    const key = value.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
  }
  return out
}

function statusTagType(status) {
  return getVideoStatusMeta(status).tagType
}

function tvPendingDiagnostics(video = detail.value) {
  return extractTvPendingDiagnostics(video?.metadata)
}

function openTvPendingScrapeConfirm() {
  if (!detail.value?.id) {
    return
  }
  const diagnostics = tvPendingDiagnostics()
  detailVisible.value = false
  router.push({
    path: '/scrape',
    query: {
      video_id: detail.value.id,
      type: 'tv',
      title: diagnostics.parsedTitle || detail.value.title || '',
      season_number: diagnostics.parsedSeasonNumber || undefined,
      episode_number: diagnostics.parsedEpisodeNumber || undefined
    }
  })
}

function openTvSeriesManage() {
  const diagnostics = tvPendingDiagnostics()
  detailVisible.value = false
  router.push({
    path: '/tv-series',
    query: {
      q: diagnostics.parsedTitle || detail.value?.title || ''
    }
  })
}

async function load() {
  const data = await getAdminVideos(query)
  list.value = data.items || []
  total.value = data.total_count || 0
  clearSelection()
}

function applyFilters() {
  query.page = 1
  load()
}

function resetFilters() {
  query.q = ''
  query.type = ''
  query.status = ''
  query.page = 1
  load()
}

async function searchActors(keyword = '') {
  loadingActors.value = true
  try {
    const data = await getAdminActors({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 20
    })
    const options = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
    const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
    for (const item of options) {
      optionMap.set(item.value, item)
    }
    actorOptions.value = Array.from(optionMap.values())
  } finally {
    loadingActors.value = false
  }
}

async function searchCollections(keyword = '') {
  loadingCollections.value = true
  try {
    const data = await getAdminCollections({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 50
    })
    const options = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
    const optionMap = new Map(collectionOptions.value.map((item) => [item.value, item]))
    for (const item of options) {
      optionMap.set(item.value, item)
    }
    collectionOptions.value = Array.from(optionMap.values())
  } finally {
    loadingCollections.value = false
  }
}

async function searchImageCollections(keyword = '') {
  loadingImageCollections.value = true
  try {
    const data = await getAdminImageCollections({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 50
    })
    const options = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
    const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
    for (const item of options) {
      optionMap.set(item.value, item)
    }
    imageCollectionOptions.value = Array.from(optionMap.values())
  } finally {
    loadingImageCollections.value = false
  }
}

async function showDetail(row) {
  detail.value = await getAdminVideoDetail(row.id)
  detail.value.actor_tokens = (detail.value.actors || []).map((actor) => actor.id)
  detail.value.collection_ids = (detail.value.collections || []).map((collection) => collection.id)
  detail.value.image_collection_id = detail.value.image_collection?.id || ''
  detail.value.target_type = ''
  detail.value.auto_scrape = true
  detail.value.season_number = 1
  detail.value.episode_number = 1
  mergeActorOptions(detail.value.actors || [])
  mergeCollectionOptions(detail.value.collections || [])
  if (detail.value.image_collection) {
    mergeImageCollectionOptions([detail.value.image_collection])
  }
  detailVisible.value = true
  playURL.value = ''
  playExpiresAt.value = 0
  subtitleItems.value = []
  subtitleForm.language_code = ''
  subtitleForm.label = ''
  subtitleForm.is_default = false
  subtitleUploadFileList.value = []
  await Promise.all([
    searchActors(''),
    searchCollections(''),
    searchImageCollections(''),
    supportsSubtitleManage(detail.value?.type) ? loadSubtitles(detail.value.id) : Promise.resolve()
  ])
  if (detail.value?.status === 'ready') {
    await refreshPlayURL()
  }
}

async function refreshPlayURL() {
  if (!detail.value?.id) {
    return
  }
  if (detail.value.status !== 'ready') {
    playURL.value = ''
    playExpiresAt.value = 0
    return
  }
  loadingPlayURL.value = true
  try {
    const data = await getAdminVideoPlayURL(detail.value.id)
    playURL.value = data.signed_url || ''
    playExpiresAt.value = Number(data.expires_at || 0)
  } finally {
    loadingPlayURL.value = false
  }
}

async function captureCurrentFrameThumbnail() {
  if (!detail.value?.id) {
    return
  }
  if (detail.value.status !== 'ready') {
    ElMessage.warning('当前视频未就绪，无法截帧')
    return
  }
  if (!playURL.value) {
    ElMessage.warning('请先刷新播放链接并加载视频')
    return
  }
  const player = previewPlayerRef.value
  if (!player) {
    ElMessage.warning('播放器未就绪，请稍后重试')
    return
  }
  const currentTime = Number(player.currentTime || 0)
  if (!Number.isFinite(currentTime) || currentTime < 0) {
    ElMessage.warning('无法读取当前播放时间')
    return
  }

  capturingThumbnail.value = true
  try {
    const data = await captureAdminVideoThumbnail(detail.value.id, { time_seconds: currentTime })
    detail.value.thumbnail_path = data.thumbnail_path || detail.value.thumbnail_path
    const usedSeconds = Number(data?.captured_at_seconds)
    const displaySeconds = Number.isFinite(usedSeconds) ? usedSeconds : currentTime
    ElMessage.success(`已截取 ${displaySeconds.toFixed(2)} 秒画面并更新封面`)
  } catch (error) {
    ElMessage.error(error?.message || '截帧更新封面失败')
  } finally {
    capturingThumbnail.value = false
  }
}

async function doDelete(row) {
  await ElMessageBox.confirm(`确认删除 ${row.title} ?`, '警告')
  await deleteAdminVideo(row.id)
  ElMessage.success('删除成功')
  await reloadAfterDeletion(1)
}

function onSelectionChange(rows) {
  selectedRows.value = Array.isArray(rows) ? rows : []
}

function clearSelection() {
  selectedRows.value = []
  tableRef.value?.clearSelection?.()
}

function selectedVideoIDs() {
  return selectedRows.value
    .map((item) => String(item?.id || '').trim())
    .filter((item) => item !== '')
}

function buildBatchDeleteMessage(data) {
  const successCount = Number(data?.success_count || 0)
  const failureCount = Number(data?.failure_count || 0)
  const failures = Array.isArray(data?.results) ? data.results.filter((item) => !item?.deleted) : []
  if (failureCount <= 0) {
    return `已删除 ${successCount} 条视频`
  }
  const failureSummary = failures
    .slice(0, 3)
    .map((item) => item?.message || item?.video_id || '未知错误')
    .join('；')
  return `删除完成，成功 ${successCount} 条，失败 ${failureCount} 条${failureSummary ? `：${failureSummary}` : ''}`
}

async function reloadAfterDeletion(deletedCount = 0) {
  const currentCount = list.value.length
  if (query.page > 1 && deletedCount > 0 && deletedCount >= currentCount) {
    query.page -= 1
  }
  await load()
}

async function doBatchDelete() {
  const videoIDs = selectedVideoIDs()
  if (videoIDs.length === 0 || deletingBatch.value) {
    return
  }
  await ElMessageBox.confirm(`确认批量删除已勾选的 ${videoIDs.length} 条视频？此操作不可恢复。`, '警告', {
    type: 'warning'
  })
  deletingBatch.value = true
  try {
    const data = await batchDeleteAdminVideos({ video_ids: videoIDs })
    const successCount = Number(data?.success_count || 0)
    const failureCount = Number(data?.failure_count || 0)
    if (failureCount > 0) {
      ElMessage.warning(buildBatchDeleteMessage(data))
    } else {
      ElMessage.success(buildBatchDeleteMessage(data))
    }
    await reloadAfterDeletion(successCount)
  } catch (error) {
    ElMessage.error(error?.message || '批量删除失败')
  } finally {
    deletingBatch.value = false
  }
}

async function doRetranscode(row) {
  await ElMessageBox.confirm('确认重新转码?', '提示')
  await retranscodeVideo(row.id)
  ElMessage.success('已加入转码队列')
  await load()
}

async function saveDetail() {
  if (!detail.value?.id) {
    return
  }
  const { actorIDs, actorNames } = splitActorSelection(detail.value.actor_tokens)
  const payload = {
    title: detail.value.title,
    description: detail.value.description,
    thumbnail: detail.value.thumbnail_path,
    tags: detail.value.tags || [],
    actor_ids: actorIDs,
    actor_names: actorNames,
    metadata: detail.value.metadata || {},
    image_collection_id: String(detail.value.image_collection_id || '').trim()
  }
  let enqueueAutoScrape = false
  if (detail.value.type === 'short') {
    payload.collection_ids = normalizeCollectionSelection(detail.value.collection_ids)
    const targetType = normalizeVideoType(detail.value.target_type)
    if (targetType === 'movie' || targetType === 'episode' || targetType === 'av') {
      payload.type = targetType
      payload.auto_scrape = detail.value.auto_scrape !== false
      payload.collection_ids = []
      enqueueAutoScrape = payload.auto_scrape === true
      if (targetType === 'episode') {
        const seasonNumber = Number(detail.value.season_number)
        const episodeNumber = Number(detail.value.episode_number)
        if (!Number.isInteger(seasonNumber) || seasonNumber <= 0 || !Number.isInteger(episodeNumber) || episodeNumber <= 0) {
          ElMessage.warning('切换为剧集分集时，请填写季号和集号（正整数）')
          return
        }
        payload.season_number = seasonNumber
        payload.episode_number = episodeNumber
      }
    }
  }
  await updateAdminVideo(detail.value.id, payload)
  ElMessage.success(enqueueAutoScrape ? '保存成功，已加入自动刮削队列' : '保存成功')
  detailVisible.value = false
  await load()
}

async function loadSubtitles(videoID = detail.value?.id) {
  if (!videoID || !supportsSubtitleManage(detail.value?.type)) {
    subtitleItems.value = []
    return
  }
  subtitleLoading.value = true
  try {
    const data = await getAdminVideoSubtitles(videoID)
    subtitleItems.value = data.items || []
  } finally {
    subtitleLoading.value = false
  }
}

async function uploadSubtitle() {
  if (!detail.value?.id || subtitleUploading.value) {
    return
  }
  if (subtitleUploadFileList.value.length === 0) {
    ElMessage.warning('请先选择字幕文件')
    return
  }
  const file = subtitleUploadFileList.value[0]?.raw
  if (!file) {
    ElMessage.warning('字幕文件读取失败')
    return
  }
  const formData = new FormData()
  formData.append('file', file)
  if (subtitleForm.language_code?.trim()) {
    formData.append('language_code', subtitleForm.language_code.trim())
  }
  if (subtitleForm.label?.trim()) {
    formData.append('label', subtitleForm.label.trim())
  }
  formData.append('is_default', subtitleForm.is_default ? 'true' : 'false')
  subtitleUploading.value = true
  try {
    await uploadAdminVideoSubtitle(detail.value.id, formData)
    ElMessage.success('字幕上传成功')
    subtitleUploadRef.value?.clearFiles()
    subtitleUploadFileList.value = []
    subtitleForm.language_code = ''
    subtitleForm.label = ''
    subtitleForm.is_default = false
    await loadSubtitles(detail.value.id)
  } catch (error) {
    ElMessage.error(error?.message || '字幕上传失败')
  } finally {
    subtitleUploading.value = false
  }
}

async function rescanSubtitles() {
  if (!detail.value?.id) {
    return
  }
  subtitleLoading.value = true
  try {
    await rescanAdminVideoSubtitles(detail.value.id)
    ElMessage.success('内嵌字幕重扫完成')
    await loadSubtitles(detail.value.id)
  } catch (error) {
    ElMessage.error(error?.message || '内嵌字幕重扫失败')
  } finally {
    subtitleLoading.value = false
  }
}

function subtitleSourceLabel(sourceType) {
  return sourceType === 'embedded' ? '内嵌' : '外挂'
}

function subtitleDefaultLabel(item) {
  if (item?.is_default) {
    return '默认'
  }
  return item?.source_type === 'embedded' ? '媒体候选' : '-'
}

async function toggleSubtitleDefault(row) {
  if (!detail.value?.id || row?.source_type !== 'uploaded') {
    return
  }
  try {
    await updateAdminVideoSubtitle(detail.value.id, row.id, {
      language_code: row.language_code || '',
      label: row.label || '',
      sort_order: Number(row.sort_order || 0),
      is_default: !row.is_default
    })
    ElMessage.success(!row.is_default ? '已设为默认字幕' : '已取消默认字幕')
    await loadSubtitles(detail.value.id)
  } catch (error) {
    ElMessage.error(error?.message || '默认字幕更新失败')
  }
}

async function removeSubtitle(row) {
  if (!detail.value?.id || row?.source_type !== 'uploaded') {
    return
  }
  await ElMessageBox.confirm(`确认删除字幕 ${row.label || row.language_code || row.id} ?`, '提示')
  await deleteAdminVideoSubtitle(detail.value.id, row.id)
  ElMessage.success('字幕已删除')
  await loadSubtitles(detail.value.id)
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page-shell video-list-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">视频管理</h1>
          <p class="page-subtitle">筛选、查看、重转码与元数据编辑</p>
        </div>
      </div>

      <section class="content-card filter-panel">
        <div class="section-head">
          <div class="section-head__main">
            <h2 class="section-head__title">筛选区</h2>
            <p class="section-head__desc">按标题、类型与状态定位目标视频</p>
          </div>
        </div>

        <el-form class="filter-form" @submit.prevent>
          <el-form-item>
            <el-input v-model="query.q" placeholder="标题/标签搜索" clearable />
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.type" placeholder="类型" clearable style="width: 160px">
              <el-option label="短视频" value="short" />
              <el-option label="电影" value="movie" />
              <el-option label="剧集分集" value="episode" />
              <el-option label="AV" value="av" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.status" placeholder="状态" clearable style="width: 160px">
              <el-option label="已上传" value="uploaded" />
              <el-option label="刮削中" value="scraping" />
              <el-option label="待绑定" value="tv_pending" />
              <el-option label="处理中" value="processing" />
              <el-option label="可播放" value="ready" />
              <el-option label="失败" value="failed" />
            </el-select>
          </el-form-item>
          <div class="toolbar-row toolbar-row--start filter-actions">
            <el-button type="primary" @click="applyFilters">查询</el-button>
            <el-button @click="resetFilters">重置</el-button>
          </div>
        </el-form>
      </section>

      <section class="table-panel result-panel">
        <div class="section-head">
          <div class="section-head__main">
            <h2 class="section-head__title">结果区</h2>
            <p class="section-head__desc">当前共 {{ total }} 条视频记录</p>
          </div>
          <div class="section-head__actions">
            <el-button
              type="danger"
              plain
              :disabled="selectedRows.length === 0"
              :loading="deletingBatch"
              @click="doBatchDelete"
            >
              批量删除
            </el-button>
          </div>
        </div>

        <div class="table-wrap">
          <el-table ref="tableRef" :data="list" border @selection-change="onSelectionChange">
            <el-table-column type="selection" width="52" />
            <el-table-column prop="title" label="标题" min-width="220" />
            <el-table-column prop="type" label="类型" width="110">
              <template #default="{ row }">
                {{ typeLabel(row.type) }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">
                  {{ statusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="upload_user" label="上传用户" width="140" />
            <el-table-column prop="created_at" label="上传时间" width="180" />
            <el-table-column label="操作" width="300">
              <template #default="{ row }">
                <el-button size="small" @click="showDetail(row)">详情</el-button>
                <el-button size="small" @click="doRetranscode(row)">重转码</el-button>
                <el-button size="small" type="danger" @click="doDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="toolbar-row toolbar-row--end">
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.page_size"
            layout="total, prev, pager, next"
            :total="total"
            @current-change="load"
          />
        </div>
      </section>
    </div>

    <el-dialog v-model="detailVisible" title="视频详情" width="760px">
      <el-form v-if="detail" label-width="90px">
        <el-form-item v-if="detail.type === 'episode' && detail.status === 'tv_pending'" label="待处理">
          <div class="tv-pending-panel">
            <el-alert
              type="warning"
              :closable="false"
              title="该剧集未自动并入电视剧树，当前处于待确认/待绑定状态。"
            />
            <div class="tv-pending-actions">
              <el-button type="primary" plain @click="openTvPendingScrapeConfirm">去刮削确认</el-button>
              <el-button plain @click="openTvSeriesManage">去电视剧管理</el-button>
            </div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="失败阶段">
                {{ tvPendingStageLabel(tvPendingDiagnostics().stage) }}
              </el-descriptions-item>
              <el-descriptions-item label="解析剧名">
                {{ tvPendingDiagnostics().parsedTitle || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="解析季集">
                {{
                  tvPendingDiagnostics().parsedSeasonNumber > 0 && tvPendingDiagnostics().parsedEpisodeNumber > 0
                    ? `S${String(tvPendingDiagnostics().parsedSeasonNumber).padStart(2, '0')}E${String(tvPendingDiagnostics().parsedEpisodeNumber).padStart(2, '0')}`
                    : '-'
                }}
              </el-descriptions-item>
              <el-descriptions-item label="候选数量">
                {{ tvPendingDiagnostics().candidateCount || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="错误信息">
                {{ tvPendingDiagnostics().error || '-' }}
              </el-descriptions-item>
            </el-descriptions>
            <div v-if="tvPendingDiagnostics().candidatePreview.length > 0" class="tv-pending-candidates">
              <div class="tv-pending-candidates__title">候选摘要</div>
              <div
                v-for="candidate in tvPendingDiagnostics().candidatePreview"
                :key="`${candidate.tmdb_id || candidate.title}-${candidate.release_date || ''}`"
                class="tv-pending-candidate"
              >
                <span>{{ candidate.title || '-' }}</span>
                <span class="tv-pending-candidate__meta">TMDB {{ candidate.tmdb_id || '-' }} / {{ candidate.release_date || '-' }}</span>
              </div>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="detail.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="detail.description" type="textarea" rows="4" /></el-form-item>
        <el-form-item label="封面"><el-input v-model="detail.thumbnail_path" /></el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="detail.tags"
            multiple
            filterable
            allow-create
            default-first-option
            clearable
            placeholder="可新增标签"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="演员">
          <el-select
            v-model="detail.actor_tokens"
            multiple
            filterable
            remote
            reserve-keyword
            allow-create
            default-first-option
            clearable
            :remote-method="searchActors"
            :loading="loadingActors"
            placeholder="可搜索演员，也可输入新演员姓名"
            style="width: 100%"
          >
            <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片图集">
          <el-select
            v-model="detail.image_collection_id"
            filterable
            remote
            reserve-keyword
            clearable
            :remote-method="searchImageCollections"
            :loading="loadingImageCollections"
            placeholder="可选，仅可关联一个图片图集"
            style="width: 100%"
          >
            <el-option
              v-for="collection in imageCollectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="detail.type === 'short'" label="改为类型">
          <el-select v-model="detail.target_type" clearable placeholder="保持短视频" style="width: 100%">
            <el-option label="保持短视频" value="" />
            <el-option v-for="item in retagTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="detail.type === 'short' && normalizeVideoType(detail.target_type) === 'episode'" label="季/集">
          <div class="episode-fields">
            <el-input-number v-model="detail.season_number" :min="1" :step="1" controls-position="right" />
            <span class="episode-divider">季</span>
            <el-input-number v-model="detail.episode_number" :min="1" :step="1" controls-position="right" />
            <span class="episode-divider">集</span>
          </div>
        </el-form-item>
        <el-form-item v-if="isRetaggingShortVideo(detail)" label="自动刮削">
          <el-alert
            type="info"
            :closable="false"
            title="保存后将自动清空短视频合集，并在后台按标题自动刮削（默认选第一个候选）。"
          />
        </el-form-item>
        <el-form-item v-if="detail.type === 'short' && !isRetaggingShortVideo(detail)" label="所属合集">
          <el-select
            v-model="detail.collection_ids"
            multiple
            filterable
            remote
            reserve-keyword
            clearable
            collapse-tags
            collapse-tags-tooltip
            :remote-method="searchCollections"
            :loading="loadingCollections"
            placeholder="可选，可多选"
            style="width: 100%"
          >
            <el-option
              v-for="collection in collectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="播放预览">
          <div class="play-preview">
            <div class="play-actions">
              <el-button
                type="primary"
                plain
                size="small"
                :loading="loadingPlayURL"
                :disabled="detail.status !== 'ready'"
                @click="refreshPlayURL"
              >
                刷新播放链接
              </el-button>
              <el-button
                type="success"
                plain
                size="small"
                :loading="capturingThumbnail"
                :disabled="detail.status !== 'ready' || !playURL"
                @click="captureCurrentFrameThumbnail"
              >
                设为封面（当前帧）
              </el-button>
              <span v-if="playExpiresAt > 0" class="play-expire">有效期至：{{ new Date(playExpiresAt * 1000).toLocaleString() }}</span>
            </div>

            <el-alert
              v-if="detail.status !== 'ready'"
              type="warning"
              :closable="false"
              title="当前视频未就绪，只有“可播放”状态才可预览。"
            />

            <el-alert
              v-else-if="!playURL"
              type="info"
              :closable="false"
              title="点击“刷新播放链接”后可预览播放。"
            />

            <video
              ref="previewPlayerRef"
              v-else
              class="preview-player"
              :src="playURL"
              controls
              preload="metadata"
            />
          </div>
        </el-form-item>

        <el-form-item v-if="supportsSubtitleManage(detail.type)" label="字幕管理">
          <div class="subtitle-panel">
            <div class="subtitle-panel__actions">
              <el-button
                type="primary"
                plain
                size="small"
                :loading="subtitleLoading"
                @click="loadSubtitles(detail.id)"
              >
                刷新字幕列表
              </el-button>
              <el-button
                plain
                size="small"
                :loading="subtitleLoading"
                @click="rescanSubtitles"
              >
                重扫内嵌字幕
              </el-button>
            </div>

            <el-table :data="subtitleItems" border size="small" v-loading="subtitleLoading">
              <el-table-column label="来源" width="84">
                <template #default="{ row }">
                  {{ subtitleSourceLabel(row.source_type) }}
                </template>
              </el-table-column>
              <el-table-column prop="language_code" label="语言" width="120" />
              <el-table-column prop="label" label="名称" min-width="180" />
              <el-table-column prop="format" label="格式" width="90" />
              <el-table-column label="默认" width="90">
                <template #default="{ row }">
                  {{ subtitleDefaultLabel(row) }}
                </template>
              </el-table-column>
              <el-table-column label="附加信息" min-width="160">
                <template #default="{ row }">
                  <span v-if="row.source_type === 'embedded'">索引 {{ row.embedded_index || '-' }}</span>
                  <span v-else>{{ Math.round(Number(row.file_size || 0) / 1024) }} KB</span>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="180">
                <template #default="{ row }">
                  <el-button
                    v-if="row.source_type === 'uploaded'"
                    size="small"
                    type="primary"
                    plain
                    @click="toggleSubtitleDefault(row)"
                  >
                    {{ row.is_default ? '取消默认' : '设默认' }}
                  </el-button>
                  <el-button
                    v-if="row.source_type === 'uploaded'"
                    size="small"
                    type="danger"
                    plain
                    @click="removeSubtitle(row)"
                  >
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div class="subtitle-upload">
              <el-upload
                ref="subtitleUploadRef"
                v-model:file-list="subtitleUploadFileList"
                :auto-upload="false"
                :limit="1"
                accept=".srt,.vtt"
              >
                <el-button size="small">选择字幕文件</el-button>
              </el-upload>
              <el-input
                v-model="subtitleForm.language_code"
                placeholder="语言代码，如 zh-CN / en"
                clearable
              />
              <el-input
                v-model="subtitleForm.label"
                placeholder="字幕名称，可选"
                clearable
              />
              <el-checkbox v-model="subtitleForm.is_default">设为默认字幕</el-checkbox>
              <el-button type="success" :loading="subtitleUploading" @click="uploadSubtitle">上传字幕</el-button>
            </div>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDetail">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.video-list-page {
  padding-bottom: 4px;
}

.filter-panel,
.result-panel {
  display: grid;
  gap: 12px;
}

.filter-form {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.filter-form .el-form-item {
  margin: 0;
}

.filter-actions {
  margin-left: auto;
}

.result-panel :deep(.el-button + .el-button) {
  margin-left: 8px;
}

.play-preview {
  width: 100%;
}

.tv-pending-panel {
  width: 100%;
  display: grid;
  gap: 12px;
}

.tv-pending-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tv-pending-candidates {
  display: grid;
  gap: 8px;
}

.tv-pending-candidates__title {
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.tv-pending-candidate {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 10px;
  border-radius: 10px;
  background: rgba(245, 158, 11, 0.12);
  color: var(--el-text-color-primary);
}

.tv-pending-candidate__meta {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.episode-fields {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.episode-divider {
  color: #6b7280;
  font-size: 13px;
}

.play-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.play-expire {
  color: #6b7280;
  font-size: 12px;
}

.preview-player {
  width: 100%;
  max-height: 360px;
  border-radius: 12px;
  border: 1px solid var(--line-soft);
  background: #000;
}

.subtitle-panel {
  width: 100%;
  display: grid;
  gap: 12px;
}

.subtitle-panel__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.subtitle-upload {
  display: grid;
  gap: 10px;
  padding: 12px;
  border-radius: 12px;
  border: 1px solid var(--line-soft);
  background: var(--surface-muted);
}

@media (max-width: 992px) {
  .filter-actions {
    width: 100%;
    margin-left: 0;
  }
}
</style>
