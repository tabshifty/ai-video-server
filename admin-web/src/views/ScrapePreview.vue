
<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import Toolbar from '../components/base/Toolbar.vue'
import { scrapeConfirm, scrapePreview } from '../api/admin'
import {
  buildScrapeConfirmPayload,
  buildScrapePreviewPayload,
  resolveTVPreviewState,
  toPositiveInt
} from './scrapePreview.helpers'

const form = reactive({ video_id: '', title: '', year: null, type: 'movie', season_number: null, episode_number: null })
const edit = reactive({
  video_id: '',
  tmdb_id: 0,
  external_id: '',
  title: '',
  overview: '',
  poster_url: '',
  backdrop_url: '',
  release_date: '',
  metadata: {}
})
const candidates = ref([])
const selectedIndex = ref(-1)
const previewLoading = ref(false)
const saveLoading = ref(false)
const route = useRoute()

const selectedCandidate = computed(() => {
  if (selectedIndex.value < 0 || selectedIndex.value >= candidates.value.length) {
    return null
  }
  return candidates.value[selectedIndex.value]
})

const selectedMetadata = computed(() => {
  const metadata = selectedCandidate.value?.metadata
  if (!metadata || typeof metadata !== 'object') {
    return {}
  }
  return metadata
})

const prettyMetadata = computed(() => JSON.stringify(selectedMetadata.value, null, 2))

onMounted(() => {
  const videoID = typeof route.query.video_id === 'string' ? route.query.video_id.trim() : ''
  const title = typeof route.query.title === 'string' ? route.query.title.trim() : ''
  const type = typeof route.query.type === 'string' ? route.query.type.trim() : ''
  const year = toPositiveInt(route.query.year)
  const seasonNumber = toPositiveInt(route.query.season_number)
  const episodeNumber = toPositiveInt(route.query.episode_number)
  if (videoID) {
    form.video_id = videoID
  }
  if (title) {
    form.title = title
  }
  if (type === 'tv' || type === 'movie') {
    form.type = type
  }
  if (year > 0) {
    form.year = year
  }
  if (seasonNumber > 0) {
    form.season_number = seasonNumber
  }
  if (episodeNumber > 0) {
    form.episode_number = episodeNumber
  }
})

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') {
    return responseMsg.trim()
  }
  const message = error?.message
  if (typeof message === 'string' && message.trim() !== '') {
    return message.trim()
  }
  return fallback
}

function mediaTypeLabel(type) {
  if (type === 'tv') {
    return '剧集'
  }
  if (type === 'movie') {
    return '电影'
  }
  if (type === 'av') {
    return 'AV'
  }
  return type || '-'
}

function toText(value, fallback = '-') {
  if (value === null || typeof value === 'undefined') {
    return fallback
  }
  if (typeof value === 'number') {
    return String(value)
  }
  if (typeof value === 'string') {
    const trimmed = value.trim()
    return trimmed === '' ? fallback : trimmed
  }
  return fallback
}

function toNumberText(value, fallback = '-') {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return String(value)
  }
  return fallback
}

function toArray(value) {
  return Array.isArray(value) ? value : []
}

function joinArray(value) {
  const rows = toArray(value).map((item) => toText(item, '')).filter(Boolean)
  return rows.length > 0 ? rows.join(' / ') : '-'
}

function joinObjectField(value, field) {
  const rows = toArray(value)
    .map((item) => {
      if (!item || typeof item !== 'object') {
        return ''
      }
      return toText(item[field], '')
    })
    .filter(Boolean)
  return rows.length > 0 ? rows.join(' / ') : '-'
}

function displayGenres(item) {
  if (Array.isArray(item?.genres) && item.genres.length > 0) {
    return item.genres.join(' / ')
  }
  return joinObjectField(item?.metadata?.genres, 'name')
}

function candidateMediaType(item) {
  return item?.media_type_hint || form.type
}

function candidateKey(item, index) {
  return item?.external_id || item?.tmdb_id || String(index)
}

function candidateIDLabel(item) {
  return candidateMediaType(item) === 'av' ? 'AVID' : 'TMDB'
}

function candidateIDValue(item) {
  if (candidateMediaType(item) === 'av') {
    return toText(item?.external_id)
  }
  return toText(item?.tmdb_id)
}

function resolvePoster(urlOrPath) {
  const value = toText(urlOrPath, '')
  if (value === '') {
    return ''
  }
  if (value.startsWith('http://') || value.startsWith('https://')) {
    return value
  }
  return 'https://image.tmdb.org/t/p/w500' + value
}

function syncEditFromCandidate(item) {
  const mediaType = item?.media_type_hint
  if (mediaType === 'movie' || mediaType === 'tv' || mediaType === 'av') {
    form.type = mediaType
  }
  if (mediaType === 'tv') {
    const parsedState = resolveTVPreviewState(item, form)
    if (parsedState.title) {
      form.title = parsedState.title
    }
    form.season_number = parsedState.season_number
    form.episode_number = parsedState.episode_number
  }
  const tmdbID = Number(item?.tmdb_id || 0)
  edit.video_id = item.video_id || form.video_id
  edit.tmdb_id = Number.isFinite(tmdbID) && tmdbID > 0 ? Math.trunc(tmdbID) : 0
  edit.external_id = item.external_id || ''
  if (mediaType === 'tv') {
    edit.title = ''
    edit.overview = ''
    edit.poster_url = ''
    edit.backdrop_url = ''
    edit.release_date = ''
  } else {
    edit.title = item.title || ''
    edit.overview = item.overview || ''
    edit.poster_url = item.poster_url || item.poster_path || ''
    edit.backdrop_url = item.backdrop_url || item.backdrop_path || ''
    edit.release_date = item.release_date || ''
  }
  edit.metadata = item.metadata || {}
}

async function doPreview() {
  previewLoading.value = true
  try {
    const data = await scrapePreview(buildScrapePreviewPayload(form))
    const list = Array.isArray(data?.candidates) ? data.candidates : []
    candidates.value = list
    if (list.length === 0) {
      selectedIndex.value = -1
      ElMessage.warning('未查询到匹配的刮削候选')
      return
    }
    selectedIndex.value = 0
    syncEditFromCandidate(list[0])
  } catch (error) {
    candidates.value = []
    selectedIndex.value = -1
    ElMessage.error(extractErrorMessage(error, '查询刮削信息失败'))
  } finally {
    previewLoading.value = false
  }
}

function choose(item, index) {
  selectedIndex.value = index
  syncEditFromCandidate(item)
}

async function doSave() {
  if (!edit.video_id) {
    ElMessage.warning('请先填写视频ID')
    return
  }
  if (form.type === 'av') {
    if (!edit.external_id) {
      ElMessage.warning('请先查询并选择一个 AV 候选结果')
      return
    }
  } else if (form.type === 'tv') {
    const seasonNumber = toPositiveInt(form.season_number)
    const episodeNumber = toPositiveInt(form.episode_number)
    const tmdbID = Number(edit.tmdb_id)
    if (!Number.isFinite(tmdbID) || tmdbID <= 0) {
      ElMessage.warning('请先查询并选择一个剧集候选结果')
      return
    }
    if (seasonNumber <= 0 || episodeNumber <= 0) {
      ElMessage.warning('剧集刮削必须填写季数和集数')
      return
    }
    edit.tmdb_id = Math.trunc(tmdbID)
  } else {
    const tmdbID = Number(edit.tmdb_id)
    if (!Number.isFinite(tmdbID) || tmdbID <= 0) {
      ElMessage.warning('请先查询并选择一个候选结果')
      return
    }
    edit.tmdb_id = Math.trunc(tmdbID)
  }
  saveLoading.value = true
  try {
    await scrapeConfirm(buildScrapeConfirmPayload(form, edit))
    ElMessage.success('刮削信息已保存')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存刮削信息失败'))
  } finally {
    saveLoading.value = false
  }
}
</script>

<template>
  <Layout>
    <div class="page-shell page-shell--medium">
      <PageHeader title="通用刮削" />

      <Toolbar>
        <template #filters>
          <el-form inline class="scrape-filter-form">
            <el-form-item label="视频ID">
              <el-input v-model="form.video_id" style="width: 300px" :disabled="previewLoading || saveLoading" @keyup.enter="doPreview" />
            </el-form-item>
            <el-form-item label="标题">
              <el-input v-model="form.title" :disabled="previewLoading || saveLoading" @keyup.enter="doPreview" />
            </el-form-item>
            <el-form-item label="年份">
              <el-input-number
                v-model="form.year"
                :min="1900"
                :disabled="previewLoading || saveLoading || form.type === 'av'"
              />
            </el-form-item>
            <el-form-item label="类型">
              <el-select v-model="form.type" style="width: 120px" :disabled="previewLoading || saveLoading">
                <el-option label="电影" value="movie" />
                <el-option label="剧集" value="tv" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="form.type === 'tv'" label="季/集">
              <div class="tv-episode-fields">
                <el-input-number
                  v-model="form.season_number"
                  :min="1"
                  :step="1"
                  controls-position="right"
                  :disabled="previewLoading || saveLoading"
                />
                <span class="tv-episode-divider">季</span>
                <el-input-number
                  v-model="form.episode_number"
                  :min="1"
                  :step="1"
                  controls-position="right"
                  :disabled="previewLoading || saveLoading"
                />
                <span class="tv-episode-divider">集</span>
              </div>
            </el-form-item>
          </el-form>
        </template>
        <template #actions>
          <el-button type="primary" :icon="Search" :loading="previewLoading" @click="doPreview">查询预览</el-button>
        </template>
      </Toolbar>

      <SectionCard>
        <template #title>候选列表</template>
        <template #description>输入视频信息后检索候选，支持电影与剧集两种类型。</template>
        <div v-loading="previewLoading" class="candidate-list-shell">
          <EmptyState
            v-if="!candidates.length"
            title="暂无候选数据"
            description="先查询预览，再从候选中选择一个结果。"
          />
          <div v-else class="candidate-list">
            <div
              v-for="(item, index) in candidates"
              :key="candidateKey(item, index)"
              class="candidate-item"
              :class="{ active: index === selectedIndex }"
              role="button"
              tabindex="0"
              :aria-pressed="index === selectedIndex"
              @click="choose(item, index)"
              @keydown.enter.prevent="choose(item, index)"
              @keydown.space.prevent="choose(item, index)"
            >
              <div class="candidate-title">{{ toText(item.title) }}</div>
              <div class="candidate-subtitle">{{ toText(item.original_title) }}</div>
              <div class="candidate-meta">
                <span>类型：{{ mediaTypeLabel(candidateMediaType(item)) }}</span>
                <span>评分：{{ toText(item.vote_average) }}</span>
              </div>
              <div class="candidate-meta">日期：{{ toText(item.release_date) }}</div>
              <div class="candidate-overview">{{ toText(item.overview) }}</div>
            </div>
          </div>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>候选详情</template>
        <template #description>检查海报、日期与原始 metadata，再确认是否写入视频。</template>
        <EmptyState
          v-if="!selectedCandidate"
          title="请先选择候选数据"
          description="从左侧候选列表选择一个结果后，这里会显示完整详情。"
        />
        <div v-else class="detail-wrap">
          <div class="detail-head">
            <img
              v-if="resolvePoster(selectedCandidate.poster_url || selectedCandidate.poster_path)"
              :src="resolvePoster(selectedCandidate.poster_url || selectedCandidate.poster_path)"
              :alt="'海报：' + toText(selectedCandidate.title, '未命名')"
              class="poster"
            />
            <div class="detail-head-info">
              <h3>{{ toText(selectedCandidate.title) }}</h3>
              <p class="origin-name">原名：{{ toText(selectedCandidate.original_title) }}</p>
              <div class="tag-line">
                <el-tag size="small" type="danger">{{ mediaTypeLabel(candidateMediaType(selectedCandidate)) }}</el-tag>
                <el-tag size="small">{{ candidateIDLabel(selectedCandidate) }}: {{ candidateIDValue(selectedCandidate) }}</el-tag>
                <el-tag size="small" type="warning">评分: {{ toText(selectedCandidate.vote_average) }}</el-tag>
              </div>
            </div>
          </div>

          <p class="detail-overview">{{ toText(selectedCandidate.overview) }}</p>

          <div v-if="candidateMediaType(selectedCandidate) === 'tv' && form.season_number && form.episode_number" class="episode-box">
            <div class="episode-title">目标分集</div>
            <div class="episode-content">
              第 {{ toNumberText(form.season_number) }} 季 · 第 {{ toNumberText(form.episode_number) }} 集
            </div>
          </div>

          <el-descriptions border :column="2" size="small" class="detail-table">
            <el-descriptions-item v-if="candidateMediaType(selectedCandidate) === 'av'" label="番号">
              {{ toText(selectedCandidate.av_code) }}
            </el-descriptions-item>
            <el-descriptions-item label="上映/首播日期">{{ toText(selectedCandidate.release_date || selectedMetadata.first_air_date) }}</el-descriptions-item>
            <el-descriptions-item label="题材">{{ displayGenres(selectedCandidate) }}</el-descriptions-item>
            <el-descriptions-item v-if="candidateMediaType(selectedCandidate) === 'av'" label="演员">
              {{ joinArray(selectedCandidate.actors || selectedMetadata.actors) }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">{{ toText(selectedMetadata.status) }}</el-descriptions-item>
            <el-descriptions-item label="节目类型">{{ toText(selectedMetadata.type) }}</el-descriptions-item>
            <el-descriptions-item label="季数">{{ toNumberText(selectedMetadata.number_of_seasons) }}</el-descriptions-item>
            <el-descriptions-item label="总集数">{{ toNumberText(selectedMetadata.number_of_episodes) }}</el-descriptions-item>
            <el-descriptions-item label="语言">{{ joinArray(selectedMetadata.languages) }}</el-descriptions-item>
            <el-descriptions-item label="原产国">{{ joinArray(selectedMetadata.origin_country) }}</el-descriptions-item>
            <el-descriptions-item label="平台">{{ joinObjectField(selectedMetadata.networks, 'name') }}</el-descriptions-item>
            <el-descriptions-item label="制作公司">{{ joinObjectField(selectedMetadata.production_companies, 'name') }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="selectedMetadata.last_episode_to_air" class="episode-box">
            <div class="episode-title">最近播出集</div>
            <div class="episode-content">
              {{ toText(selectedMetadata.last_episode_to_air.name) }}
              （第 {{ toNumberText(selectedMetadata.last_episode_to_air.episode_number) }} 集，{{ toText(selectedMetadata.last_episode_to_air.air_date) }}）
            </div>
          </div>

          <el-collapse>
            <el-collapse-item title="查看原始 metadata JSON" name="metadata-json">
              <pre class="json-box">{{ prettyMetadata }}</pre>
            </el-collapse-item>
          </el-collapse>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>编辑并确认</template>
        <template #description>把确认后的 metadata 写入当前视频。</template>
        <el-form label-width="90px">
          <el-form-item v-if="form.type !== 'av'" label="TMDB ID">
            <el-input-number
              v-model="edit.tmdb_id"
              :min="1"
              :step="1"
              :precision="0"
              controls-position="right"
              :disabled="saveLoading"
              style="width: 220px"
            />
          </el-form-item>
          <el-form-item v-else label="AVID">
            <el-input v-model="edit.external_id" :disabled="saveLoading" />
          </el-form-item>
          <el-alert
            v-if="form.type === 'tv'"
            type="info"
            :closable="false"
            title="电视剧确认会按上方季/集绑定具体分集；标题、简介、海报和日期留空时将自动使用目标分集元数据。"
            style="margin-bottom: 16px"
          />
          <el-form-item label="标题"><el-input v-model="edit.title" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="简介"><el-input v-model="edit.overview" type="textarea" rows="3" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="海报URL"><el-input v-model="edit.poster_url" :disabled="saveLoading" /></el-form-item>
          <el-form-item v-if="form.type === 'movie'" label="横向背景"><el-input v-model="edit.backdrop_url" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="发布日期"><el-input v-model="edit.release_date" :disabled="saveLoading" /></el-form-item>
        </el-form>
        <Toolbar dense>
          <template #actions>
            <el-button type="primary" :loading="saveLoading" @click="doSave">保存</el-button>
          </template>
        </Toolbar>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.page-shell--medium {
  display: grid;
  gap: var(--space-5);
}

.scrape-filter-form {
  display: inline-flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.tv-episode-fields {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.tv-episode-divider {
  color: var(--text-muted);
  font-size: 13px;
}

.candidate-list-shell {
  min-height: 12rem;
}

.candidate-list {
  display: grid;
  gap: var(--space-3);
}

.candidate-item {
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
  background: var(--bg-surface);
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.candidate-item:hover,
.candidate-item.active {
  transform: translateY(-1px);
  border-color: var(--primary);
  box-shadow: var(--shadow-xs);
}

.candidate-item:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

.candidate-title {
  font-weight: 700;
  color: var(--text-primary);
}

.candidate-subtitle {
  color: var(--text-muted);
  margin-top: 2px;
  font-size: 13px;
}

.candidate-meta {
  color: var(--text-muted);
  font-size: 12px;
  margin-top: 6px;
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.candidate-overview {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
  max-height: 72px;
  overflow: auto;
  margin-top: 8px;
}

.detail-wrap {
  display: grid;
  gap: var(--space-3);
}

.detail-head {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
}

.poster {
  width: 92px;
  height: 128px;
  object-fit: cover;
  border-radius: var(--radius-md);
  border: 1px solid var(--line-soft);
}

.detail-head-info h3 {
  margin: 0;
  color: var(--text-primary);
}

.origin-name {
  margin: 6px 0;
  color: var(--text-muted);
}

.tag-line {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.detail-overview {
  margin: 0;
  color: var(--text-secondary);
  line-height: 1.7;
}

.detail-table {
  margin-top: 4px;
}

.episode-box {
  border: 1px dashed var(--line-soft);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  background: var(--primary-soft);
}

.episode-title {
  font-weight: 600;
  color: var(--text-primary);
}

.episode-content {
  margin-top: 6px;
  color: var(--text-secondary);
}

.json-box {
  margin: 0;
  padding: 12px;
  background: var(--bg-inverse);
  color: var(--text-on-inverse, #e2e8f0);
  border-radius: var(--radius-md);
  max-height: 320px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
}

@media (max-width: 900px) {
  .detail-head {
    flex-direction: column;
  }
}
</style>
