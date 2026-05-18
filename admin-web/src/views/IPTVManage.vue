<script setup>
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
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
  if (!value) return '暂无'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toLocaleString('zh-CN', { hour12: false })
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
    <div class="page page-shell">
      <section class="section-head">
        <div>
          <h1 class="page-title">IPTV 管理</h1>
          <p class="page-subtitle">维护 M3U 播放列表来源并预览频道解析结果</p>
        </div>
        <div class="section-head__actions">
          <el-button :icon="Refresh" :loading="loading" @click="loadPlaylist">重新加载</el-button>
        </div>
      </section>

      <section class="stats-grid">
        <div v-for="item in stats" :key="item.label" class="stat-card">
          <div class="stat-label">{{ item.label }}</div>
          <div class="stat-value">{{ item.value }}</div>
        </div>
      </section>

      <section class="ops-grid">
        <el-card class="soft-card content-card">
          <template #header>
            <div class="card-title">上传 M3U 文件</div>
          </template>
          <div class="upload-panel">
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
        </el-card>

        <el-card class="soft-card content-card">
          <template #header>
            <div class="card-title">远程 M3U URL</div>
          </template>
          <el-form label-position="top">
            <el-form-item label="播放列表地址">
              <el-input v-model="sourceUrl" clearable placeholder="请输入远程 M3U / M3U8 URL" />
            </el-form-item>
            <div class="source-actions">
              <el-button type="primary" :loading="saveSourceLoading" @click="saveSourceUrl">保存 URL</el-button>
              <el-button :icon="Refresh" :loading="refreshLoading" @click="refreshPlaylist">手动刷新</el-button>
            </div>
          </el-form>
          <div class="status-line">最后更新时间：{{ updatedAtText }}</div>
        </el-card>
      </section>

      <section>
        <el-card class="soft-card table-card">
          <template #header>
            <div class="table-head">
              <div>
                <div class="card-title">频道预览</div>
                <p>展示当前播放列表解析出的频道信息</p>
              </div>
              <el-tag type="info" effect="plain">{{ channels.length }} 个频道</el-tag>
            </div>
          </template>
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
        </el-card>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #7f1d1d;
}

.stat-card {
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: #fff;
  box-shadow: var(--shadow-soft);
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted);
}

.stat-value {
  margin-top: 6px;
  font-family: var(--font-code);
  font-size: 28px;
  font-weight: 600;
  color: #7f1d1d;
}

.ops-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 14px;
}

.upload-panel {
  display: grid;
  gap: 12px;
}

.upload-icon {
  color: #7f1d1d;
}

.source-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.status-line {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--line-soft);
  font-size: 13px;
  color: var(--text-muted);
}

.table-card {
  padding: 0;
}

.table-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.table-head p {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--text-muted);
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
  font-size: 12px;
}

@media (max-width: 900px) {
  .ops-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 700px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
