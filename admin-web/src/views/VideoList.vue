<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import {
  captureAdminVideoThumbnail,
  deleteAdminVideo,
  getAdminActors,
  getAdminCollections,
  getAdminImageCollections,
  getAdminVideoDetail,
  getAdminVideoPlayURL,
  getAdminVideos,
  retranscodeVideo,
  updateAdminVideo
} from '../api/admin'

const list = ref([])
const total = ref(0)
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
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

const query = reactive({ page: 1, page_size: 20, q: '', type: '', status: '' })
const retagTypeOptions = [
  { value: 'movie', label: '电影' },
  { value: 'episode', label: '剧集分集' },
  { value: 'av', label: 'AV' }
]

function statusLabel(status) {
  const map = {
    uploaded: '已上传',
    scraping: '刮削中',
    processing: '处理中',
    ready: '可播放',
    failed: '失败'
  }
  return map[status] || status || '-'
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

async function load() {
  const data = await getAdminVideos(query)
  list.value = data.items || []
  total.value = data.total_count || 0
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
  await Promise.all([searchActors(''), searchCollections(''), searchImageCollections('')])
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
  await load()
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
        </div>

        <div class="table-wrap">
          <el-table :data="list" border>
            <el-table-column prop="title" label="标题" min-width="220" />
            <el-table-column prop="type" label="类型" width="110">
              <template #default="{ row }">
                {{ typeLabel(row.type) }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.status === 'ready' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
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

@media (max-width: 992px) {
  .filter-actions {
    width: 100%;
    margin-left: 0;
  }
}
</style>
