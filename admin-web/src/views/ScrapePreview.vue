<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import { scrapeConfirm, scrapePreview } from '../api/admin'

const form = reactive({ video_id: '', title: '', year: new Date().getFullYear(), type: 'movie' })
const edit = reactive({
  video_id: '',
  tmdb_id: 0,
  external_id: '',
  title: '',
  overview: '',
  poster_url: '',
  release_date: '',
  metadata: {}
})
const candidates = ref([])
const selectedIndex = ref(-1)
const previewLoading = ref(false)
const saveLoading = ref(false)

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
  return item?.external_id || item?.tmdb_id || `${index}`
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
  return `https://image.tmdb.org/t/p/w500${value}`
}

function syncEditFromCandidate(item) {
  const mediaType = item?.media_type_hint
  if (mediaType === 'movie' || mediaType === 'tv' || mediaType === 'av') {
    form.type = mediaType
  }
  edit.video_id = item.video_id || form.video_id
  edit.tmdb_id = item.tmdb_id || 0
  edit.external_id = item.external_id || ''
  edit.title = item.title || ''
  edit.overview = item.overview || ''
  edit.poster_url = item.poster_url || item.poster_path || ''
  edit.release_date = item.release_date || ''
  edit.metadata = item.metadata || {}
}

async function doPreview() {
  previewLoading.value = true
  try {
    const data = await scrapePreview({ ...form })
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
  } else if (!edit.tmdb_id) {
    ElMessage.warning('请先查询并选择一个候选结果')
    return
  }
  saveLoading.value = true
  try {
    await scrapeConfirm({ ...edit })
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
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">刮削管理</h1>
          <p class="page-subtitle">预览候选详情并确认保存，支持完整 metadata 检视</p>
        </div>
      </div>

      <el-card class="soft-card">
        <el-form inline class="filter-form">
          <el-form-item label="视频ID">
            <el-input v-model="form.video_id" style="width:300px" :disabled="previewLoading || saveLoading" />
          </el-form-item>
          <el-form-item label="标题">
            <el-input v-model="form.title" :disabled="previewLoading || saveLoading" />
          </el-form-item>
          <el-form-item label="年份">
            <el-input-number
              v-model="form.year"
              :min="1900"
              :disabled="previewLoading || saveLoading || form.type === 'av'"
            />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="form.type" style="width:120px" :disabled="previewLoading || saveLoading">
              <el-option label="电影" value="movie" />
              <el-option label="剧集" value="tv" />
              <el-option label="AV" value="av" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="previewLoading" @click="doPreview">查询预览</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-row :gutter="12" class="result-row">
        <el-col :xs="24" :lg="10">
          <el-card class="soft-card" v-loading="previewLoading">
            <template #header>候选列表</template>
            <el-empty v-if="!candidates.length" description="暂无候选数据" />
            <div v-else class="candidate-list">
              <div
                v-for="(item, index) in candidates"
                :key="candidateKey(item, index)"
                class="candidate-item"
                :class="{ active: index === selectedIndex }"
                @click="choose(item, index)"
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
          </el-card>
        </el-col>

        <el-col :xs="24" :lg="14">
          <el-card class="soft-card">
            <template #header>候选详情</template>
            <el-empty v-if="!selectedCandidate" description="请先查询并选择候选数据" />
            <div v-else class="detail-wrap">
              <div class="detail-head">
                <img
                  v-if="resolvePoster(selectedCandidate.poster_url || selectedCandidate.poster_path)"
                  :src="resolvePoster(selectedCandidate.poster_url || selectedCandidate.poster_path)"
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
          </el-card>
        </el-col>
      </el-row>

      <el-card class="soft-card">
        <template #header>编辑并确认</template>
        <el-form label-width="90px">
          <el-form-item v-if="form.type !== 'av'" label="TMDB ID">
            <el-input v-model="edit.tmdb_id" :disabled="saveLoading" />
          </el-form-item>
          <el-form-item v-else label="AVID">
            <el-input v-model="edit.external_id" :disabled="saveLoading" />
          </el-form-item>
          <el-form-item label="标题"><el-input v-model="edit.title" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="简介"><el-input v-model="edit.overview" type="textarea" rows="3" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="海报URL"><el-input v-model="edit.poster_url" :disabled="saveLoading" /></el-form-item>
          <el-form-item label="发布日期"><el-input v-model="edit.release_date" :disabled="saveLoading" /></el-form-item>
        </el-form>
        <el-button type="primary" :loading="saveLoading" @click="doSave">保存</el-button>
      </el-card>
    </div>
  </Layout>
</template>

<style scoped>
.result-row {
  margin-top: 12px;
}

.candidate-list {
  display: grid;
  gap: 10px;
}

.candidate-item {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.candidate-item:hover {
  border-color: #f43f5e;
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.08);
}

.candidate-item.active {
  border-color: #be123c;
  background: rgba(244, 63, 94, 0.05);
}

.candidate-title {
  font-weight: 700;
  color: #881337;
}

.candidate-subtitle {
  color: #6b7280;
  margin-top: 2px;
  font-size: 13px;
}

.candidate-meta {
  color: #6b7280;
  font-size: 12px;
  margin-top: 6px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.candidate-overview {
  font-size: 13px;
  line-height: 1.6;
  color: #374151;
  max-height: 72px;
  overflow: auto;
  margin-top: 8px;
}

.detail-wrap {
  display: grid;
  gap: 12px;
}

.detail-head {
  display: flex;
  gap: 12px;
}

.poster {
  width: 92px;
  height: 128px;
  object-fit: cover;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
}

.detail-head-info h3 {
  margin: 0;
  color: #111827;
}

.origin-name {
  margin: 6px 0;
  color: #6b7280;
}

.tag-line {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.detail-overview {
  margin: 0;
  color: #374151;
  line-height: 1.7;
}

.detail-table {
  margin-top: 4px;
}

.episode-box {
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  padding: 10px;
  background: #f8fafc;
}

.episode-title {
  font-weight: 600;
  color: #334155;
}

.episode-content {
  margin-top: 6px;
  color: #475569;
}

.json-box {
  margin: 0;
  padding: 12px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 10px;
  max-height: 320px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
