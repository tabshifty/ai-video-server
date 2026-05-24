<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import Toolbar from '../components/base/Toolbar.vue'
import { useRoute } from 'vue-router'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import {
  createAdminTvEpisode,
  createAdminTvSeason,
  createAdminTvSeries,
  deleteAdminTvEpisode,
  deleteAdminTvSeason,
  deleteAdminTvSeries,
  getAdminTvSeries,
  getAdminTvSeriesDetail,
  getAdminVideos,
  updateAdminTvEpisode,
  updateAdminTvSeason,
  updateAdminTvSeries
} from '../api/admin'
import {
  buildTvEpisodePayload,
  buildTvSeasonPayload,
  buildTvSeriesPayload,
  mapEpisodeVideoOption
} from './tvSeriesManage.helpers'

const loadingList = ref(false)
const loadingDetail = ref(false)
const savingKey = ref('')
const list = ref([])
const total = ref(0)
const selectedSeriesId = ref('')
const detail = ref(createEmptySeriesDetail())
const episodeVideoOptions = ref([])
const loadingVideoOptions = ref(false)
const activeSeasons = ref([])
const route = useRoute()

const query = reactive({
  page: 1,
  page_size: 12,
  q: '',
  active: '',
  has_playable: ''
})

function createEmptySeriesDetail() {
  return {
    id: '',
    title: '',
    overview: '',
    poster_url: '',
    backdrop_url: '',
    first_air_date: '',
    active: true,
    total_seasons: 0,
    total_episodes: 0,
    playable_episodes: 0,
    seasons: []
  }
}

function createNewSeasonDraft() {
  return {
    id: '',
    _temp_key: `season-${Date.now()}-${Math.random()}`,
    season_number: 1,
    title: '',
    overview: '',
    poster_url: '',
    air_date: '',
    episodes: []
  }
}

function createNewEpisodeDraft() {
  return {
    id: '',
    _temp_key: `episode-${Date.now()}-${Math.random()}`,
    episode_number: 1,
    title: '',
    overview: '',
    runtime: 0,
    air_date: '',
    still_url: '',
    video_id: '',
    video_title: '',
    video_status: '',
    playable: false
  }
}

function listParams() {
  const params = {
    page: query.page,
    page_size: query.page_size
  }
  if (query.q.trim()) params.q = query.q.trim()
  if (query.active !== '') params.active = query.active
  if (query.has_playable !== '') params.has_playable = query.has_playable
  return params
}

function mergeVideoOptions(options = []) {
  const optionMap = new Map(episodeVideoOptions.value.map((item) => [item.value, item]))
  for (const option of options) {
    optionMap.set(option.value, option)
  }
  episodeVideoOptions.value = Array.from(optionMap.values())
}

function syncBoundVideoOptions() {
  const bound = []
  for (const season of detail.value.seasons || []) {
    for (const episode of season.episodes || []) {
      if (!episode.video_id) continue
      bound.push(
        mapEpisodeVideoOption({
          id: episode.video_id,
          title: episode.video_title || episode.video_id,
          status: episode.video_status
        })
      )
    }
  }
  mergeVideoOptions(bound)
}

async function loadList() {
  loadingList.value = true
  try {
    const data = await getAdminTvSeries(listParams())
    list.value = data.items || []
    total.value = Number(data.total_count || 0)
    if (!selectedSeriesId.value && list.value.length > 0) {
      await selectSeries(list.value[0].id)
    }
  } finally {
    loadingList.value = false
  }
}

async function selectSeries(id) {
  selectedSeriesId.value = String(id)
  loadingDetail.value = true
  try {
    const data = await getAdminTvSeriesDetail(id)
    detail.value = {
      ...data,
      id: String(data.id),
      seasons: (data.seasons || []).map((season) => ({
        ...season,
        id: String(season.id),
        series_id: String(season.series_id),
        episodes: (season.episodes || []).map((episode) => ({
          ...episode,
          id: String(episode.id),
          season_id: String(episode.season_id)
        }))
      }))
    }
    activeSeasons.value = detail.value.seasons.map((season) => season.id)
    syncBoundVideoOptions()
  } finally {
    loadingDetail.value = false
  }
}

function openCreateSeries() {
  selectedSeriesId.value = ''
  detail.value = createEmptySeriesDetail()
  activeSeasons.value = []
}

async function saveSeries() {
  const payload = buildTvSeriesPayload(detail.value)
  if (!payload.title) {
    ElMessage.warning('请输入系列标题')
    return
  }
  savingKey.value = 'series'
  try {
    const saved = detail.value.id
      ? await updateAdminTvSeries(detail.value.id, payload)
      : await createAdminTvSeries(payload)
    ElMessage.success(detail.value.id ? '系列已更新' : '系列已创建')
    await loadList()
    await selectSeries(saved.id)
  } catch (error) {
    ElMessage.error(error?.message || '保存系列失败')
  } finally {
    savingKey.value = ''
  }
}

async function removeSeries() {
  if (!detail.value.id) return
  await ElMessageBox.confirm(`确认删除「${detail.value.title}」？删除后会级联清理该系列下的季与集。`, '删除电视剧', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
  await deleteAdminTvSeries(detail.value.id)
  ElMessage.success('系列已删除')
  openCreateSeries()
  await loadList()
}

function addSeason() {
  detail.value.seasons.unshift(createNewSeasonDraft())
  activeSeasons.value = detail.value.seasons.map((season) => season.id || season._temp_key)
}

async function saveSeason(season) {
  const payload = buildTvSeasonPayload(season)
  savingKey.value = `season-${season.id || season._temp_key}`
  try {
    if (season.id) {
      await updateAdminTvSeason(season.id, payload)
      ElMessage.success('季度已更新')
    } else {
      if (!detail.value.id) {
        ElMessage.warning('请先保存系列，再新增季度')
        return
      }
      await createAdminTvSeason(detail.value.id, payload)
      ElMessage.success('季度已创建')
    }
    await selectSeries(detail.value.id)
  } catch (error) {
    ElMessage.error(error?.message || '保存季度失败')
  } finally {
    savingKey.value = ''
  }
}

async function removeSeason(season) {
  if (!season.id) {
    detail.value.seasons = detail.value.seasons.filter((item) => item !== season)
    return
  }
  await ElMessageBox.confirm(`确认删除第 ${season.season_number} 季？`, '删除季度', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
  await deleteAdminTvSeason(season.id)
  ElMessage.success('季度已删除')
  await selectSeries(detail.value.id)
}

function addEpisode(season) {
  season.episodes.unshift(createNewEpisodeDraft())
}

async function saveEpisode(season, episode) {
  const payload = buildTvEpisodePayload(episode)
  savingKey.value = `episode-${episode.id || episode._temp_key}`
  try {
    if (episode.id) {
      await updateAdminTvEpisode(episode.id, payload)
      ElMessage.success('分集已更新')
    } else {
      if (!season.id) {
        ElMessage.warning('请先保存季度，再新增分集')
        return
      }
      await createAdminTvEpisode(season.id, payload)
      ElMessage.success('分集已创建')
    }
    await selectSeries(detail.value.id)
  } catch (error) {
    ElMessage.error(error?.message || '保存分集失败')
  } finally {
    savingKey.value = ''
  }
}

async function removeEpisode(season, episode) {
  if (!episode.id) {
    season.episodes = season.episodes.filter((item) => item !== episode)
    return
  }
  await ElMessageBox.confirm(`确认删除第 ${episode.episode_number} 集？`, '删除分集', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
  await deleteAdminTvEpisode(episode.id)
  ElMessage.success('分集已删除')
  await selectSeries(detail.value.id)
}

async function searchEpisodeVideos(keyword = '') {
  loadingVideoOptions.value = true
  try {
    const data = await getAdminVideos({
      q: keyword,
      type: 'episode',
      page: 1,
      page_size: 20
    })
    mergeVideoOptions((data.items || []).map((item) => mapEpisodeVideoOption(item)))
  } finally {
    loadingVideoOptions.value = false
  }
}

onMounted(async () => {
  if (typeof route.query.q === 'string' && route.query.q.trim()) {
    query.q = route.query.q.trim()
  }
  await Promise.all([loadList(), searchEpisodeVideos('')])
})
</script>

<template>
  <Layout>
    <div class="page-shell page-shell--medium tv-manage-shell">
      <PageHeader title="电视剧管理" />

      <Toolbar>
        <template #filters>
          <el-input v-model="query.q" class="tv-search" placeholder="搜索系列标题" clearable @keyup.enter="query.page = 1; loadList()" />
          <el-select v-model="query.active" class="tv-filter" placeholder="启用状态" clearable>
            <el-option label="仅启用" value="1" />
            <el-option label="仅停用" value="0" />
          </el-select>
          <el-select v-model="query.has_playable" class="tv-filter" placeholder="可播分集" clearable>
            <el-option label="有可播分集" value="1" />
            <el-option label="无可播分集" value="0" />
          </el-select>
          <el-button @click="query.q = ''; query.active = ''; query.has_playable = ''; query.page = 1; loadList()">重置</el-button>
        </template>
        <template #actions>
          <el-button type="primary" :icon="Plus" @click="openCreateSeries">新建系列</el-button>
          <el-button :icon="Search" @click="query.page = 1; loadList()">筛选</el-button>
        </template>
      </Toolbar>

      <div class="tv-manage-grid">
        <SectionCard class="tv-series-list-card">
          <template #title>电视剧列表</template>
          <template #description>按系列维度管理季与分集。</template>
          <div v-loading="loadingList" class="series-list-shell">
            <EmptyState
              v-if="!list.length && !loadingList"
              title="暂无电视剧"
              description="先创建一个系列，再维护季与分集。"
            />
            <div v-else class="series-list">
              <div
                v-for="item in list"
                :key="item.id"
                class="series-card"
                :class="{ 'is-active': String(item.id) === selectedSeriesId }"
                @click="selectSeries(item.id)"
              >
                <div class="series-card__title">{{ item.title }}</div>
                <div class="series-card__meta">
                  <span>{{ item.total_seasons || 0 }} 季</span>
                  <span>{{ item.total_episodes || 0 }} 集</span>
                  <span>{{ item.playable_episodes || 0 }} 集可播</span>
                </div>
                <div class="series-card__status">
                  <el-tag :type="item.active ? 'success' : 'info'">{{ item.active ? '启用' : '停用' }}</el-tag>
                </div>
              </div>
            </div>
          </div>
          <div class="pager-wrap">
            <AdminTablePagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              layout="prev, pager, next"
              :total="total"
              @current-change="loadList"
            />
          </div>
        </SectionCard>

        <SectionCard class="tv-series-editor-card" v-loading="loadingDetail">
          <template #title>{{ detail.id ? '系列详情' : '新建电视剧系列' }}</template>
          <template #description>系列、季、集三级编辑，分集可绑定已有 type=episode 视频。</template>

          <SectionCard>
            <template #title>系列基础</template>
            <div class="editor-grid">
              <el-form label-position="top">
                <el-form-item label="系列标题">
                  <el-input v-model="detail.title" placeholder="例如：雾城档案" />
                </el-form-item>
                <el-form-item label="简介">
                  <el-input v-model="detail.overview" type="textarea" :rows="4" />
                </el-form-item>
                <el-form-item label="海报 URL / 路径">
                  <el-input v-model="detail.poster_url" />
                </el-form-item>
                <el-form-item label="横图 URL / 路径">
                  <el-input v-model="detail.backdrop_url" />
                </el-form-item>
              </el-form>

              <el-form label-position="top">
                <el-form-item label="首播日期">
                  <el-input v-model="detail.first_air_date" placeholder="YYYY-MM-DD" />
                </el-form-item>
                <el-form-item label="启用状态">
                  <el-switch v-model="detail.active" active-text="启用" inactive-text="停用" />
                </el-form-item>
                <el-descriptions :column="1" border class="series-stats">
                  <el-descriptions-item label="总季数">{{ detail.total_seasons || 0 }}</el-descriptions-item>
                  <el-descriptions-item label="总集数">{{ detail.total_episodes || 0 }}</el-descriptions-item>
                  <el-descriptions-item label="可播分集">{{ detail.playable_episodes || 0 }}</el-descriptions-item>
                </el-descriptions>
              </el-form>
            </div>
            <Toolbar dense>
              <template #actions>
                <el-button type="primary" :loading="savingKey === 'series'" @click="saveSeries">保存系列</el-button>
                <el-button v-if="detail.id" type="danger" plain @click="removeSeries">删除系列</el-button>
              </template>
            </Toolbar>
          </SectionCard>

          <SectionCard>
            <template #title>季度与分集</template>
            <template #description>优先在这里完成电视剧结构维护，通用视频页只做底层视频排查。</template>
            <div class="season-toolbar">
              <el-button type="primary" plain @click="addSeason">新增季度</el-button>
            </div>

            <EmptyState
              v-if="!detail.seasons.length"
              title="暂无季度"
              description="保存系列后可以继续创建季度与分集。"
            />

            <div v-else class="season-stack">
              <SectionCard v-for="season in detail.seasons" :key="season.id || season._temp_key">
                <template #title>第 {{ season.season_number }} 季</template>
                <template #description>{{ season.title || '未命名季度' }}</template>

                <div class="season-form-grid">
                  <el-form label-position="top">
                    <el-form-item label="季序号">
                      <el-input v-model="season.season_number" />
                    </el-form-item>
                    <el-form-item label="名称">
                      <el-input v-model="season.title" />
                    </el-form-item>
                    <el-form-item label="播出日期">
                      <el-input v-model="season.air_date" placeholder="YYYY-MM-DD" />
                    </el-form-item>
                  </el-form>
                  <el-form label-position="top">
                    <el-form-item label="简介">
                      <el-input v-model="season.overview" type="textarea" :rows="3" />
                    </el-form-item>
                    <el-form-item label="海报 URL / 路径">
                      <el-input v-model="season.poster_url" />
                    </el-form-item>
                  </el-form>
                </div>

                <div class="season-actions">
                  <el-button :loading="savingKey === 'season-' + (season.id || season._temp_key)" type="primary" @click="saveSeason(season)">保存季度</el-button>
                  <el-button plain @click="addEpisode(season)">新增分集</el-button>
                  <el-button type="danger" plain @click="removeSeason(season)">删除季度</el-button>
                </div>

                <div class="episode-stack">
                  <SectionCard v-for="episode in season.episodes" :key="episode.id || episode._temp_key">
                    <template #title>第 {{ episode.episode_number }} 集</template>
                    <template #description>
                      <el-tag :type="episode.playable ? 'success' : 'info'">{{ episode.playable ? '可播放' : '待绑定 / 未就绪' }}</el-tag>
                    </template>

                    <div class="episode-form-grid">
                      <el-form label-position="top">
                        <el-form-item label="集序号">
                          <el-input v-model="episode.episode_number" />
                        </el-form-item>
                        <el-form-item label="标题">
                          <el-input v-model="episode.title" />
                        </el-form-item>
                        <el-form-item label="简介">
                          <el-input v-model="episode.overview" type="textarea" :rows="3" />
                        </el-form-item>
                      </el-form>

                      <el-form label-position="top">
                        <el-form-item label="时长（秒）">
                          <el-input v-model="episode.runtime" />
                        </el-form-item>
                        <el-form-item label="播出日期">
                          <el-input v-model="episode.air_date" placeholder="YYYY-MM-DD" />
                        </el-form-item>
                        <el-form-item label="剧照 URL / 路径">
                          <el-input v-model="episode.still_url" />
                        </el-form-item>
                      </el-form>

                      <el-form label-position="top">
                        <el-form-item label="绑定视频">
                          <el-select
                            v-model="episode.video_id"
                            filterable
                            remote
                            clearable
                            reserve-keyword
                            placeholder="搜索已有剧集视频"
                            :remote-method="searchEpisodeVideos"
                            :loading="loadingVideoOptions"
                          >
                            <el-option
                              v-for="option in episodeVideoOptions"
                              :key="option.value"
                              :label="option.label"
                              :value="option.value"
                            />
                          </el-select>
                        </el-form-item>
                        <div class="episode-binding-tip">
                          当前绑定：{{ episode.video_title || '未绑定' }}
                          <span v-if="episode.video_status">（{{ episode.video_status }}）</span>
                        </div>
                      </el-form>
                    </div>

                    <Toolbar dense>
                      <template #actions>
                        <el-button :loading="savingKey === 'episode-' + (episode.id || episode._temp_key)" type="primary" @click="saveEpisode(season, episode)">保存分集</el-button>
                        <el-button type="danger" plain @click="removeEpisode(season, episode)">删除分集</el-button>
                      </template>
                    </Toolbar>
                  </SectionCard>
                </div>
              </SectionCard>
            </div>
          </SectionCard>
        </SectionCard>
      </div>
    </div>
  </Layout>
</template>

<style scoped>
.page-shell--medium {
  display: grid;
  gap: var(--space-5);
}

.tv-manage-shell {
  min-width: 0;
}

.tv-manage-grid {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.tv-search {
  width: min(18rem, 100%);
}

.tv-filter {
  width: 10rem;
}

.series-list-shell {
  min-height: 20rem;
}

.series-list {
  display: grid;
  gap: var(--space-3);
}

.series-card {
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
  cursor: pointer;
  background: var(--bg-surface);
  transition: transform 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.series-card:hover,
.series-card.is-active {
  transform: translateY(-1px);
  border-color: var(--primary);
  box-shadow: var(--shadow-xs);
}

.series-card__title {
  font-weight: 700;
  color: var(--text-primary);
}

.series-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: 8px;
  color: var(--text-muted);
  font-size: 12px;
}

.series-card__status {
  margin-top: 10px;
}

.pager-wrap {
  margin-top: var(--space-4);
  display: flex;
  justify-content: center;
}

.editor-grid,
.season-form-grid,
.episode-form-grid {
  display: grid;
  gap: var(--space-3);
}

.editor-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.season-form-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.episode-form-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.series-stats {
  margin-top: var(--space-3);
}

.season-toolbar {
  margin-bottom: var(--space-3);
}

.season-stack,
.episode-stack {
  display: grid;
  gap: var(--space-3);
}

.season-actions {
  margin-top: var(--space-3);
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.episode-binding-tip {
  color: var(--text-muted);
  font-size: 12px;
}

@media (max-width: 1280px) {
  .tv-manage-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 960px) {
  .editor-grid,
  .season-form-grid,
  .episode-form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
