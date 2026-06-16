<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatCard from '../components/base/StatCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import {
  getAdminIPTVPlaylist,
  refreshAdminIPTVPlaylist,
  updateAdminIPTVSource,
  uploadAdminIPTVPlaylist
} from '../api/admin'

const loading = ref(false)
const uploadLoading = ref(false)
const saveSourceLoading = ref(false)
const refreshLoading = ref(false)
const uploadFiles = ref([])
const sourceUrl = ref('')
const playlist = ref(createEmptyPlaylist())

const stats = computed(() => [
  { label: '频道数', value: Number(playlist.value.channel_count || 0) },
  { label: '跳过数', value: Number(playlist.value.skipped_count || 0) },
  { label: '分组数', value: groupCount.value },
  { label: '预览数', value: channels.value.length }
])
const channels = computed(() => (Array.isArray(playlist.value.channels) ? playlist.value.channels : []))
const groupCount = computed(() => {
  if (Array.isArray(playlist.value.groups)) {
    return playlist.value.groups.length
  }
  return new Set(channels.value.map((item) => item.group).filter(Boolean)).size
})
const updatedAtText = computed(() => formatDateTime(playlist.value.updated_at))
const hasChannels = computed(() => channels.value.length > 0)

function createEmptyPlaylist() {
  return {
    source_url: '',
    updated_at: '',
    channel_count: 0,
    skipped_count: 0,
    groups: [],
    channels: []
  }
}

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

function applyPlaylist(data) {
  playlist.value = {
    ...createEmptyPlaylist(),
    ...(data || {}),
    groups: Array.isArray(data?.groups) ? data.groups : [],
    channels: Array.isArray(data?.channels) ? data.channels : []
  }
  sourceUrl.value = playlist.value.source_url || ''
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '暂无')
}

async function loadPlaylist() {
  loading.value = true
  try {
    applyPlaylist(await getAdminIPTVPlaylist())
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载 IPTV 状态失败'))
  } finally {
    loading.value = false
  }
}

function onUploadChange(file, files) {
  uploadFiles.value = files.slice(-1)
}

function onUploadRemove(file, files) {
  uploadFiles.value = files
}

function beforeUpload(file) {
  const lowerName = String(file.name || '').toLowerCase()
  const valid = lowerName.endsWith('.m3u') || lowerName.endsWith('.m3u8')
  if (!valid) {
    ElMessage.warning('请选择 .m3u 或 .m3u8 文件')
  }
  return valid
}

async function uploadPlaylist() {
  const rawFile = uploadFiles.value[0]?.raw
  if (!rawFile) {
    ElMessage.warning('请先选择 M3U 文件')
    return
  }
  if (!beforeUpload(rawFile)) {
    return
  }
  const formData = new FormData()
  formData.append('file', rawFile)
  uploadLoading.value = true
  try {
    applyPlaylist(await uploadAdminIPTVPlaylist(formData))
    uploadFiles.value = []
    ElMessage.success('M3U 文件已上传并替换')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '上传 M3U 文件失败'))
  } finally {
    uploadLoading.value = false
  }
}

async function saveSourceUrl() {
  const trimmed = sourceUrl.value.trim()
  if (!trimmed) {
    ElMessage.warning('请输入远程 M3U URL')
    return
  }
  saveSourceLoading.value = true
  try {
    applyPlaylist(await updateAdminIPTVSource({ source_url: trimmed }))
    ElMessage.success('远程 M3U URL 已保存')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存远程 M3U URL 失败'))
  } finally {
    saveSourceLoading.value = false
  }
}

async function refreshPlaylist() {
  refreshLoading.value = true
  try {
    applyPlaylist(await refreshAdminIPTVPlaylist())
    ElMessage.success('IPTV 播放列表已刷新')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '刷新 IPTV 播放列表失败'))
  } finally {
    refreshLoading.value = false
  }
}

onMounted(loadPlaylist)
</script>

<template>
  <Layout>
    <div class="page-shell iptv-page">
      <PageHeader title="IPTV 管理" subtitle="维护 M3U 播放列表来源并预览频道解析结果">
        <template #actions>
          <el-button :icon="Refresh" :loading="loading" @click="loadPlaylist">刷新</el-button>
        </template>
      </PageHeader>

      <Toolbar>
        <template #filters>
          <el-tag effect="plain">最后更新时间：{{ updatedAtText }}</el-tag>
        </template>
        <template #actions>
          <el-button :icon="Refresh" :loading="refreshLoading" @click="refreshPlaylist">远程拉取</el-button>
          <el-button type="primary" :icon="UploadFilled" :loading="uploadLoading" @click="uploadPlaylist">上传 M3U</el-button>
        </template>
      </Toolbar>

      <section class="stat-grid">
        <StatCard v-for="item in stats" :key="item.label" :label="item.label" :value="item.value" />
      </section>

      <SectionCard>
        <template #title>播放列表来源</template>
        <template #description>上传本地文件或更新远程 M3U URL</template>
        <div class="source-grid">
          <article class="source-panel">
            <div class="source-panel__title">本地 M3U 文件</div>
            <p class="source-panel__desc">选择 .m3u / .m3u8 文件后点击上传并替换。</p>
            <div class="source-panel__actions">
              <el-upload
                v-model:file-list="uploadFiles"
                drag
                :auto-upload="false"
                :limit="1"
                accept=".m3u,.m3u8"
                :on-change="onUploadChange"
                :on-remove="onUploadRemove"
              >
                <el-icon class="upload-icon"><UploadFilled /></el-icon>
                <div class="el-upload__text">拖拽文件到此处，或点击选择</div>
                <template #tip>
                  <div class="el-upload__tip">支持 .m3u / .m3u8，上传后会替换当前播放列表。</div>
                </template>
              </el-upload>
              <el-button type="primary" :loading="uploadLoading" @click="uploadPlaylist">上传并替换</el-button>
            </div>
          </article>

          <article class="source-panel">
            <div class="source-panel__title">远程 M3U URL</div>
            <p class="source-panel__desc">保存后可手动拉取最新频道列表。</p>
            <el-form label-position="top">
              <el-form-item label="播放列表地址">
                <el-input v-model="sourceUrl" clearable placeholder="请输入远程 M3U / M3U8 URL" />
              </el-form-item>
              <div class="source-panel__actions">
                <el-button type="primary" :loading="saveSourceLoading" @click="saveSourceUrl">保存 URL</el-button>
                <el-button :loading="refreshLoading" @click="refreshPlaylist">手动刷新</el-button>
              </div>
            </el-form>
          </article>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>频道预览</template>
        <template #description>展示当前播放列表解析出的频道信息</template>
        <EmptyState
          v-if="!hasChannels"
          title="暂无频道"
          description="上传 M3U 文件或填入远程地址后刷新"
        />
        <div v-else class="table-wrap">
          <el-table v-loading="loading" :data="channels" border stripe class="channel-table" empty-text="暂无频道数据">
            <el-table-column prop="name" label="频道名" min-width="180" show-overflow-tooltip />
            <el-table-column prop="group" label="分组" min-width="130" show-overflow-tooltip>
              <template #default="{ row }">{{ row.group || '未分组' }}</template>
            </el-table-column>
            <el-table-column label="台标" width="96" align="center">
              <template #default="{ row }">
                <el-image v-if="row.logo_url" class="logo-image" :src="row.logo_url" fit="contain" lazy>
                  <template #error>
                    <span class="logo-empty">无</span>
                  </template>
                </el-image>
                <span v-else class="logo-empty">无</span>
              </template>
            </el-table-column>
            <el-table-column label="播放地址" min-width="260" show-overflow-tooltip>
              <template #default="{ row }">
                <el-link v-if="row.url" :href="row.url" target="_blank" type="primary">{{ row.url }}</el-link>
                <span v-else>暂无</span>
              </template>
            </el-table-column>
            <el-table-column prop="tvg_id" label="tvg-id" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">{{ row.tvg_id || '暂无' }}</template>
            </el-table-column>
          </el-table>
        </div>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.iptv-page {
  display: grid;
  gap: var(--space-4);
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.source-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
}

.source-panel {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.source-panel__title {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
}

.source-panel__desc {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.source-panel__actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.upload-icon {
  color: var(--primary);
}

.channel-table {
  width: 100%;
}

.logo-image {
  width: 44px;
  height: 28px;
  vertical-align: middle;
}

.logo-empty {
  color: var(--text-muted);
  font-size: var(--text-caption);
}

@media (max-width: 75rem) {
  .stat-grid,
  .source-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 46rem) {
  .stat-grid,
  .source-grid {
    grid-template-columns: 1fr;
  }
}
</style>
