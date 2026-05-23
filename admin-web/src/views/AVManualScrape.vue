<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'
import Layout from '../components/Layout.vue'
import {
  avScrapeConfirm,
  avScrapePreview,
  getAdminVideoDetail,
  getAVScrapeConfig,
  updateAVScrapeConfig
} from '../api/admin'
import {
  AV_POSTER_CROP_MODE_OPTIONS,
  applyAVManualScrapeRouteQuery,
  applyAVScrapeConfig,
  AV_SITE_OPTIONS,
  buildAVManualScrapeConfirmPayload,
  buildAVManualScrapePreviewPayload,
  buildAVScrapeConfigPayload
} from './avManualScrape.helpers'

const route = useRoute()
const previewLoading = ref(false)
const saveLoading = ref(false)
const configLoading = ref(false)
const configSaving = ref(false)
const candidates = ref([])
const selectedIndex = ref(-1)
const recommendedSource = ref('')
const usedSource = ref('')
const enabledSources = ref([])

const form = reactive({
  video_id: '',
  title: '',
  site_category: '',
  site_source: '',
  bypass_cache: true
})

const edit = reactive({
  video_id: '',
  external_id: '',
  title: '',
  overview: '',
  poster_url: '',
  release_date: '',
  metadata: {}
})

const configForm = reactive({
  enabled_sites: [],
  fc2_order: '',
  western_order: '',
  japanese_order: '',
  poster_crop_enabled: true,
  poster_crop_mode: 'portrait_center'
})

const selectedCandidate = computed(() => {
  if (selectedIndex.value < 0 || selectedIndex.value >= candidates.value.length) {
    return null
  }
  return candidates.value[selectedIndex.value]
})

const prettyMetadata = computed(() => JSON.stringify(selectedCandidate.value?.metadata || {}, null, 2))

onMounted(async () => {
  applyAVManualScrapeRouteQuery(form, edit, route.query)
  await loadConfig()
  await loadPendingPreviewFromVideo()
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

function toText(value, fallback = '-') {
  if (value === null || typeof value === 'undefined') {
    return fallback
  }
  const text = String(value).trim()
  return text === '' ? fallback : text
}

function resolvePoster(urlOrPath) {
  const value = toText(urlOrPath, '')
  if (value === '') {
    return ''
  }
  if (value.startsWith('http://') || value.startsWith('https://')) {
    return value
  }
  return value
}

async function loadPendingPreviewFromVideo() {
  if (!form.video_id) {
    return
  }
  try {
    const video = await getAdminVideoDetail(form.video_id)
    const metadata = video?.metadata && typeof video.metadata === 'object' ? video.metadata : {}
    if (video?.status !== 'av_scrape_pending' || !Array.isArray(metadata.scrape_preview)) {
      return
    }
    if (!form.title) {
      form.title = video.title || ''
    }
    if (!form.site_category && typeof metadata.site_category === 'string') {
      form.site_category = metadata.site_category
    }
    candidates.value = metadata.scrape_preview
    if (candidates.value.length > 0) {
      selectedIndex.value = 0
      syncEditFromCandidate(candidates.value[0])
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载待确认候选失败'))
  }
}

async function loadConfig() {
  configLoading.value = true
  try {
    const payload = await getAVScrapeConfig()
    applyAVScrapeConfig(configForm, payload || {})
    enabledSources.value = Array.isArray(payload?.enabled_sites) ? payload.enabled_sites : []
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载 AV 刮削配置失败'))
  } finally {
    configLoading.value = false
  }
}

async function saveConfig() {
  configSaving.value = true
  try {
    const payload = buildAVScrapeConfigPayload(configForm)
    await updateAVScrapeConfig(payload)
    enabledSources.value = payload.enabled_sites
    ElMessage.success('AV 刮削配置已保存')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存 AV 刮削配置失败'))
  } finally {
    configSaving.value = false
  }
}

function syncEditFromCandidate(item) {
  edit.video_id = form.video_id
  edit.external_id = item?.external_id || ''
  edit.title = item?.title || ''
  edit.overview = item?.overview || ''
  edit.poster_url = item?.poster_url || ''
  edit.release_date = item?.release_date || ''
  edit.metadata = item?.metadata || {}
}

function candidateMatchSource(item) {
  return item?.match_source || item?.metadata?.match_source || ''
}

function matchSourceLabel(value) {
  const normalized = String(value || '').trim().toLowerCase()
  const map = {
    oshash: 'hash 命中',
    'keyword:scenes': '场景关键字',
    'keyword:movies': '影片关键字',
    manual_retry: '手动重搜'
  }
  return map[normalized] || (normalized || '-')
}

function chooseCandidate(item, index) {
  selectedIndex.value = index
  form.site_source = item?.scrape_source || form.site_source
  syncEditFromCandidate(item)
}

async function doPreview() {
  previewLoading.value = true
  try {
    const payload = buildAVManualScrapePreviewPayload(form)
    const data = await avScrapePreview(payload)
    candidates.value = Array.isArray(data?.candidates) ? data.candidates : []
    recommendedSource.value = data?.recommended_source || ''
    usedSource.value = data?.used_source || ''
    enabledSources.value = Array.isArray(data?.enabled_sources) ? data.enabled_sources : enabledSources.value
    if (data?.site_category) {
      form.site_category = data.site_category
    }
    if (candidates.value.length === 0) {
      selectedIndex.value = -1
      ElMessage.warning('未查询到匹配的 AV 候选')
      return
    }
    selectedIndex.value = 0
    syncEditFromCandidate(candidates.value[0])
  } catch (error) {
    candidates.value = []
    selectedIndex.value = -1
    ElMessage.error(extractErrorMessage(error, '查询 AV 刮削预览失败'))
  } finally {
    previewLoading.value = false
  }
}

async function doSave() {
  if (!form.video_id) {
    ElMessage.warning('请先填写视频 ID')
    return
  }
  if (!edit.external_id) {
    ElMessage.warning('请先查询并选择一个候选结果')
    return
  }
  saveLoading.value = true
  try {
    await avScrapeConfirm(buildAVManualScrapeConfirmPayload(form, edit))
    ElMessage.success('AV 刮削结果已保存')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存 AV 刮削结果失败'))
  } finally {
    saveLoading.value = false
  }
}
</script>

<template>
  <Layout>
    <div class="page page-shell">
      <section class="section-head">
        <div>
          <h1 class="page-title">AV 手动刮削</h1>
          <p class="page-subtitle">每次都按当前标题与站点策略在线重抓，保存后直接覆盖标题、简介、海报与 metadata，不触发重新转码。</p>
        </div>
      </section>

      <section>
        <el-card class="soft-card content-card" v-loading="configLoading">
          <template #header>
            <div class="panel-head">
              <div class="panel-title">AV 刮削配置</div>
              <p>自动刮削与手动刮削共用这份站点策略和海报裁剪设置。</p>
            </div>
          </template>
          <el-form label-width="140px">
            <el-form-item label="启用站点">
              <el-select v-model="configForm.enabled_sites" multiple filterable clearable placeholder="选择可参与 AV 刮削的站点" style="width: 100%">
                <el-option v-for="site in AV_SITE_OPTIONS" :key="site" :label="site" :value="site" />
              </el-select>
            </el-form-item>
            <el-form-item label="FC2 默认顺序">
              <el-input v-model="configForm.fc2_order" placeholder="如：fc2, fc2club, fc2hub" />
            </el-form-item>
            <el-form-item label="欧美默认顺序">
              <el-input v-model="configForm.western_order" placeholder="如：theporndb, javdb" />
            </el-form-item>
            <el-form-item label="日系默认顺序">
              <el-input v-model="configForm.japanese_order" placeholder="如：javdb, javbus, javlibrary" />
            </el-form-item>
            <el-form-item label="海报裁剪">
              <el-switch v-model="configForm.poster_crop_enabled" active-text="启用裁剪" inactive-text="仅保留原图" />
            </el-form-item>
            <el-form-item label="裁剪模式">
              <el-select
                v-model="configForm.poster_crop_mode"
                :disabled="!configForm.poster_crop_enabled"
                placeholder="选择海报裁剪锚点"
                style="width: 100%"
              >
                <el-option
                  v-for="mode in AV_POSTER_CROP_MODE_OPTIONS"
                  :key="mode"
                  :label="mode"
                  :value="mode"
                />
              </el-select>
            </el-form-item>
          </el-form>
          <el-button type="primary" :loading="configSaving" @click="saveConfig">保存 AV 配置</el-button>
        </el-card>
      </section>

      <section>
        <el-card class="soft-card content-card">
          <template #header>
            <div class="panel-head">
              <div class="panel-title">手动预览</div>
              <p>默认不走缓存，会先按标题自动推荐站点，也可以手动切换站点后重新预览。</p>
            </div>
          </template>
          <el-form inline class="filter-form av-filter-form">
            <el-form-item label="视频 ID">
              <el-input v-model="form.video_id" style="width: 320px" :disabled="previewLoading || saveLoading" />
            </el-form-item>
            <el-form-item label="标题">
              <el-input v-model="form.title" style="width: 320px" :disabled="previewLoading || saveLoading" />
            </el-form-item>
            <el-form-item label="站点分类">
              <el-select v-model="form.site_category" clearable placeholder="按标题自动判断" style="width: 180px">
                <el-option label="自动判断" value="" />
                <el-option label="FC2" value="fc2" />
                <el-option label="欧美" value="western" />
                <el-option label="日系" value="japanese" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标站点">
              <el-select v-model="form.site_source" clearable filterable placeholder="使用自动推荐" style="width: 200px">
                <el-option label="使用自动推荐" value="" />
                <el-option v-for="site in enabledSources.length ? enabledSources : AV_SITE_OPTIONS" :key="site" :label="site" :value="site" />
              </el-select>
            </el-form-item>
            <el-form-item label="绕过缓存">
              <el-switch v-model="form.bypass_cache" active-text="始终重抓" inactive-text="允许缓存" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="previewLoading" @click="doPreview">查询预览</el-button>
            </el-form-item>
          </el-form>

          <div class="source-summary">
            <el-tag size="small" type="danger">自动推荐：{{ toText(recommendedSource) }}</el-tag>
            <el-tag size="small">实际使用：{{ toText(usedSource) }}</el-tag>
            <el-tag size="small" type="info">当前分类：{{ toText(form.site_category || '自动') }}</el-tag>
          </div>
        </el-card>
      </section>

      <section>
        <el-row :gutter="12" class="result-row">
          <el-col :xs="24" :lg="10">
            <el-card class="soft-card content-card" v-loading="previewLoading">
              <template #header>候选列表</template>
              <el-empty v-if="!candidates.length" description="暂无候选数据" />
              <div v-else class="candidate-list">
                <div
                  v-for="(item, index) in candidates"
                  :key="item.external_id || `${index}`"
                  class="candidate-item"
                  :class="{ active: index === selectedIndex }"
                  @click="chooseCandidate(item, index)"
                >
                  <div class="candidate-title">{{ toText(item.title) }}</div>
                  <div class="candidate-subtitle">{{ toText(item.av_code) }}</div>
                  <div class="candidate-meta">
                    <span>站点：{{ toText(item.scrape_source) }}</span>
                    <span>日期：{{ toText(item.release_date) }}</span>
                    <el-tag
                      v-if="candidateMatchSource(item)"
                      size="small"
                      :type="candidateMatchSource(item) === 'oshash' ? 'success' : 'info'"
                    >
                      {{ matchSourceLabel(candidateMatchSource(item)) }}
                    </el-tag>
                  </div>
                  <div class="candidate-overview">{{ toText(item.overview) }}</div>
                </div>
              </div>
            </el-card>
          </el-col>

          <el-col :xs="24" :lg="14">
            <el-card class="soft-card content-card">
              <template #header>候选详情</template>
              <el-empty v-if="!selectedCandidate" description="请先查询并选择候选数据" />
              <div v-else class="detail-wrap">
                <div class="detail-head">
                  <img
                    v-if="resolvePoster(selectedCandidate.poster_url)"
                    :src="resolvePoster(selectedCandidate.poster_url)"
                    :alt="`海报：${toText(selectedCandidate.title, '未命名')}`"
                    class="poster"
                  />
                  <div class="detail-head-info">
                    <h3>{{ toText(selectedCandidate.title) }}</h3>
                    <p class="origin-name">站点：{{ toText(selectedCandidate.scrape_source) }}</p>
                    <div class="tag-line">
                      <el-tag size="small">{{ toText(selectedCandidate.external_id) }}</el-tag>
                      <el-tag size="small" type="warning">番号：{{ toText(selectedCandidate.av_code) }}</el-tag>
                    </div>
                  </div>
                </div>
                <p class="detail-overview">{{ toText(selectedCandidate.overview) }}</p>
                <el-collapse>
                  <el-collapse-item title="查看原始 metadata JSON" name="metadata-json">
                    <pre class="json-box">{{ prettyMetadata }}</pre>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </section>

      <section>
        <el-card class="soft-card content-card">
          <template #header>覆盖保存</template>
          <el-form label-width="90px">
            <el-form-item label="AVID"><el-input v-model="edit.external_id" :disabled="saveLoading" /></el-form-item>
            <el-form-item label="标题"><el-input v-model="edit.title" :disabled="saveLoading" /></el-form-item>
            <el-form-item label="简介"><el-input v-model="edit.overview" type="textarea" rows="3" :disabled="saveLoading" /></el-form-item>
            <el-form-item label="海报 URL"><el-input v-model="edit.poster_url" :disabled="saveLoading" /></el-form-item>
            <el-form-item label="发布日期"><el-input v-model="edit.release_date" :disabled="saveLoading" /></el-form-item>
          </el-form>
          <el-button type="primary" :loading="saveLoading" @click="doSave">保存到 AV</el-button>
        </el-card>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.panel-head p {
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 12px;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #7f1d1d;
}

.av-filter-form {
  width: 100%;
}

.source-summary {
  margin-top: 14px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.result-row {
  margin-top: 0;
}

.candidate-list {
  display: grid;
  gap: 12px;
}

.candidate-item {
  padding: 14px 16px;
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 18px;
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.candidate-item:hover,
.candidate-item.active {
  transform: translateY(-1px);
  border-color: rgba(225, 29, 72, 0.4);
  box-shadow: 0 16px 32px rgba(127, 29, 29, 0.08);
}

.candidate-title {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.candidate-subtitle,
.candidate-meta,
.candidate-overview,
.origin-name,
.detail-overview {
  color: #475569;
}

.candidate-meta {
  display: flex;
  gap: 10px;
  margin-top: 8px;
  font-size: 12px;
  flex-wrap: wrap;
}

.candidate-overview {
  margin-top: 10px;
  font-size: 13px;
  line-height: 1.6;
}

.detail-wrap {
  display: grid;
  gap: 16px;
}

.detail-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.poster {
  width: 180px;
  max-width: 40vw;
  aspect-ratio: 2 / 3;
  object-fit: cover;
  border-radius: 18px;
  background: #e2e8f0;
}

.detail-head-info h3 {
  margin: 0;
  font-size: 24px;
  color: #0f172a;
}

.tag-line {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.json-box {
  margin: 0;
  padding: 16px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 16px;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
}

@media (max-width: 900px) {
  .detail-head {
    flex-direction: column;
  }
}
</style>
