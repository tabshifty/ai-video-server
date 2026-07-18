<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import QRCode from 'qrcode'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, MoreFilled, Refresh, UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import MetricStrip from '../components/base/MetricStrip.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import Toolbar from '../components/base/Toolbar.vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'
import { buildTVAppDownloadPageURL, getTVAppDownloadQRCodeTitle } from './tvAppManage.qr'
import {
  buildTVAppReleaseRequest,
  isCurrentTVAppReleaseRequest,
  selectTVAppReleaseCache
} from './tvAppManage.requestState'
import {
  deleteAdminTVAppReleaseDraft,
  downloadAdminTVAppReleaseURL,
  getAdminTVAppReleases,
  offlineAdminTVAppRelease,
  publishAdminTVAppRelease,
  restoreAdminTVAppRelease,
  updateAdminTVAppRelease,
  uploadAdminTVAppAPK
} from '../api/admin'

const CLIENTS = {
  android_tv: {
    label: 'TV 端',
    shortLabel: 'TV',
    packageName: 'com.chee.videos.tv',
    supportsAbi: true,
    uploadTip: '仅接收 Release APK，自动按 ABI 建档或补包。',
    emptyTitle: '暂无 TV 发布记录',
    emptyDesc: '先上传一个 APK 创建首条记录。',
  },
  android_phone: {
    label: '手机端',
    shortLabel: '手机',
    packageName: 'com.chee.videos',
    supportsAbi: false,
    uploadTip: '仅接收 Release APK，上传首个 APK 会自动建档。',
    emptyTitle: '暂无手机端发布记录',
    emptyDesc: '先上传一个 APK 创建首条记录。',
  }
}

const clientType = ref('android_tv')
const loading = ref(true)
const loadError = ref('')
const uploadLoading = ref(false)
const savingId = ref(0)
const actionId = ref(0)
const uploadFiles = ref([])
const downloadQRCodeDataURL = ref('')
const activeRequestKey = ref('')
const cachedRequestKey = ref(null)
let loadSequence = 0
let latestRequest = { sequence: 0, key: '' }
const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  status: '',
  abi_completeness: '',
  current_published: false
})
const data = reactive({
  items: [],
  total_count: 0,
  page: 1,
  page_size: 20
})

const clientMeta = computed(() => CLIENTS[clientType.value] || CLIENTS.android_tv)
const downloadQRCodeTitle = computed(() => getTVAppDownloadQRCodeTitle(clientType.value))
const currentCollection = computed(() => selectTVAppReleaseCache({
  activeRequestKey: activeRequestKey.value,
  cachedRequestKey: cachedRequestKey.value,
  items: data.items,
  totalCount: data.total_count
}))
const hasCurrentResult = computed(() => currentCollection.value.hasCurrentResult)
const currentItems = computed(() => currentCollection.value.items)
const currentTotalCount = computed(() => currentCollection.value.totalCount)
const visibleCount = computed(() => currentItems.value.filter((item) => item.visible_to_family).length)
const latestItem = computed(() => currentItems.value.find((item) => item.latest_recommended) || null)
const missingCount = computed(() => {
  if (clientMeta.value.supportsAbi) return currentItems.value.filter((item) => !item.abi_complete).length
  return currentItems.value.filter((item) => item.publish_status === 'draft').length
})
const initialLoading = computed(() => shouldShowCrudCollectionSkeleton({
  loading: loading.value,
  rowCount: currentItems.value.length
}))
const summaryMetrics = computed(() => [
  { key: 'total', label: '记录总数', value: currentTotalCount.value, scope: '当前筛选·全部页' },
  { key: 'visible', label: '家庭可见', value: visibleCount.value, scope: '当前页' },
  {
    key: 'incomplete',
    label: clientMeta.value.supportsAbi ? '缺少 ABI' : '草稿',
    value: missingCount.value,
    scope: '当前页'
  },
  { key: 'recommended', label: '推荐版本', value: latestItem.value ? `${latestItem.value.version_name} (${latestItem.value.version_code})` : '暂无', scope: '当前页' }
])

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') return responseMsg.trim()
  if (typeof error?.message === 'string' && error.message.trim() !== '') return error.message.trim()
  return fallback
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '暂无')
}

function formatBytes(bytes) {
  const size = Number(bytes || 0)
  if (!Number.isFinite(size) || size <= 0) return '--'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 100 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function statusText(status) {
  if (clientMeta.value.supportsAbi) {
    if (status === 'draft') return '草稿'
    if (status === 'published_complete') return '已发布-完整'
    if (status === 'published_missing_abi') return '已发布-缺少 ABI'
    if (status === 'offline') return '已下线'
  } else {
    if (status === 'draft') return '草稿'
    if (status === 'published_complete') return '已发布'
    if (status === 'offline') return '已下线'
  }
  return status || '--'
}

function statusTone(status) {
  if (!clientMeta.value.supportsAbi) {
    if (status === 'published_complete') return 'success'
    if (status === 'offline') return 'info'
    return 'neutral'
  }
  if (status === 'published_complete') return 'success'
  if (status === 'published_missing_abi') return 'warning'
  if (status === 'offline') return 'info'
  return 'neutral'
}

function abiLine(item) {
  if (!clientMeta.value.supportsAbi) {
    return item.abi_items?.length ? '已上传 APK' : '暂无 APK'
  }
  const uploaded = Array.isArray(item.uploaded_abis) ? item.uploaded_abis.join(' / ') : ''
  const missing = Array.isArray(item.missing_abis) ? item.missing_abis.join(' / ') : ''
  if (uploaded && missing) return `已上传：${uploaded}；缺失：${missing}`
  if (uploaded) return `已上传：${uploaded}`
  if (missing) return `缺失：${missing}`
  return '暂无 APK'
}

function createDraft(item) {
  return {
    release_notes: item.release_notes || '',
    remarks: item.remarks || ''
  }
}

function applyResult(result) {
  data.items = Array.isArray(result?.items) ? result.items.map((item) => ({ ...item, draft: createDraft(item) })) : []
  data.total_count = Number(result?.total_count || 0)
  data.page = Number(result?.page || query.page)
  data.page_size = Number(result?.page_size || query.page_size)
}

async function load() {
  const request = buildTVAppReleaseRequest({
    query,
    clientType: clientType.value,
    supportsAbi: clientMeta.value.supportsAbi
  })
  const token = { sequence: ++loadSequence, key: request.key }
  latestRequest = token
  activeRequestKey.value = request.key
  loadError.value = ''
  loading.value = true
  try {
    const result = await getAdminTVAppReleases(request.params)
    if (!isCurrentTVAppReleaseRequest(token, latestRequest)) return
    applyResult(result)
    cachedRequestKey.value = request.key
  } catch (error) {
    if (!isCurrentTVAppReleaseRequest(token, latestRequest)) return
    loadError.value = extractErrorMessage(error, '加载安装包列表失败')
  } finally {
    if (isCurrentTVAppReleaseRequest(token, latestRequest)) {
      loading.value = false
    }
  }
}

function resetQuery() {
  query.page = 1
  query.q = ''
  query.status = ''
  query.abi_completeness = ''
  query.current_published = false
}

async function refreshDownloadQRCode() {
  downloadQRCodeDataURL.value = await QRCode.toDataURL(
    buildTVAppDownloadPageURL(clientType.value, {
      currentOrigin: window.location.origin,
      isDev: import.meta.env.DEV,
      apiProxyTarget: import.meta.env.VITE_API_PROXY_TARGET
    }),
    {
      margin: 1,
      width: 220
    }
  )
}

function changeClientType(nextType) {
  if (!CLIENTS[nextType]) return
  if (clientType.value === nextType) return
  clientType.value = nextType
  resetQuery()
  uploadFiles.value = []
  refreshDownloadQRCode()
  load()
}

function onUploadChange(file, files) {
  uploadFiles.value = files.slice(-1)
}

function onUploadRemove(file, files) {
  uploadFiles.value = files
}

function beforeUpload(file) {
  const valid = String(file?.name || '').toLowerCase().endsWith('.apk')
  if (!valid) {
    ElMessage.warning('请选择 APK 文件')
  }
  return valid
}

async function uploadAPK(replaceExisting = false) {
  const rawFile = uploadFiles.value[0]?.raw
  if (!rawFile) {
    ElMessage.warning('请先选择 APK 文件')
    return
  }
  if (!beforeUpload(rawFile)) return
  const formData = new FormData()
  formData.append('file', rawFile)
  if (replaceExisting) formData.append('replace_existing', 'true')
  uploadLoading.value = true
  try {
    await uploadAdminTVAppAPK(formData, clientType.value)
    uploadFiles.value = []
    query.page = 1
    ElMessage.success(replaceExisting ? 'APK 已替换' : 'APK 已上传')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, replaceExisting ? '替换 APK 失败' : '上传 APK 失败'))
  } finally {
    uploadLoading.value = false
  }
}

async function saveNotes(item) {
  savingId.value = item.id
  try {
    await updateAdminTVAppRelease(item.id, {
      release_notes: item.draft.release_notes,
      remarks: item.draft.remarks
    }, clientType.value)
    ElMessage.success('版本说明已保存')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存版本说明失败'))
  } finally {
    savingId.value = 0
  }
}

async function runDangerAction(item, action) {
  actionId.value = item.id
  try {
    if (action === 'publish') {
      await publishAdminTVAppRelease(item.id, {
        release_notes: item.draft.release_notes,
        remarks: item.draft.remarks
      }, clientType.value)
    } else if (action === 'offline') {
      await offlineAdminTVAppRelease(item.id, clientType.value)
    } else if (action === 'restore') {
      await restoreAdminTVAppRelease(item.id, clientType.value)
    } else if (action === 'delete') {
      await deleteAdminTVAppReleaseDraft(item.id, clientType.value)
    }
    ElMessage.success('操作已完成')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '操作失败'))
  } finally {
    actionId.value = 0
  }
}

async function confirmAction(item, action) {
  const actionText = action === 'offline'
    ? '下线'
    : action === 'restore'
      ? '恢复发布'
      : action === 'delete'
        ? '删除草稿'
        : '发布'
  if (action === 'publish') {
    await runDangerAction(item, action)
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认${actionText} ${item.version_name} (${item.version_code}) 吗？`,
      `${actionText}确认`,
      {
        type: action === 'delete' ? 'warning' : 'info',
        confirmButtonText: '确认',
        cancelButtonText: '取消'
      }
    )
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    ElMessage.error('操作确认失败，请重试')
    return
  }
  await runDangerAction(item, action)
}

function downloadHref(item, abi) {
  return downloadAdminTVAppReleaseURL(item.id, abi, clientType.value)
}

onMounted(() => {
  load()
  refreshDownloadQRCode()
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
      <el-button type="primary" :icon="UploadFilled" :loading="uploadLoading" @click="uploadAPK(false)">上传 APK</el-button>
    </template>

    <div class="page-shell app-package-page" data-density="compact">
      <Toolbar dense>
        <template #filters>
          <el-segmented class="client-type-switch" :model-value="clientType" :options="[
            { label: 'TV 端', value: 'android_tv' },
            { label: '手机端', value: 'android_phone' }
          ]" @update:modelValue="changeClientType" />
          <el-input v-model="query.q" aria-label="安装包版本筛选" clearable :placeholder="clientMeta.supportsAbi ? '搜索版本号 / versionName / 时间' : '搜索版本号 / versionName / 时间'" style="width: 240px" @keyup.enter="load" />
          <el-select v-model="query.status" aria-label="安装包状态筛选" clearable placeholder="状态" style="width: 180px" @change="load">
            <el-option label="草稿" value="draft" />
            <el-option v-if="clientMeta.supportsAbi" label="已发布-完整" value="published_complete" />
            <el-option v-if="clientMeta.supportsAbi" label="已发布-缺少 ABI" value="published_missing_abi" />
            <el-option v-else label="已发布" value="published_complete" />
            <el-option label="已下线" value="offline" />
          </el-select>
          <el-select v-if="clientMeta.supportsAbi" v-model="query.abi_completeness" aria-label="ABI 完整性筛选" clearable placeholder="ABI 完整性" style="width: 180px" @change="load">
            <el-option label="完整" value="complete" />
            <el-option label="缺少 ABI" value="missing" />
            <el-option label="空记录" value="empty" />
          </el-select>
          <el-switch v-model="query.current_published" aria-label="家庭可见筛选" active-text="只看家庭可见" inactive-text="查看全部" @change="load" />
        </template>
      </Toolbar>

      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default><el-button link type="primary" @click="load">重试</el-button></template>
      </el-alert>

      <SectionCard dense>
        <template #title>{{ downloadQRCodeTitle }}</template>
        <div class="download-qr-card">
          <img v-if="downloadQRCodeDataURL" class="download-qr-image" :src="downloadQRCodeDataURL" :alt="downloadQRCodeTitle" width="220" height="220">
        </div>
      </SectionCard>

      <SectionCard dense>
        <template #title>上传 APK</template>
        <template #description>{{ clientMeta.packageName }} · {{ clientMeta.uploadTip }}</template>
        <div class="upload-panel">
          <el-upload
            v-model:file-list="uploadFiles"
            drag
            :auto-upload="false"
            :limit="1"
            accept=".apk"
            :on-change="onUploadChange"
            :on-remove="onUploadRemove"
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽 APK 到此处，或点击选择</div>
            <template #tip>
              <div class="el-upload__tip">
                <template v-if="clientMeta.supportsAbi">同 ABI 已存在时，服务端会拒绝静默覆盖；若需要替换，请在下线后选择“替换上传”。</template>
                <template v-else>同版本已存在时，服务端会拒绝静默覆盖；若需要替换，请在下线后选择“替换上传”。</template>
              </div>
            </template>
          </el-upload>
          <div class="upload-actions">
            <el-button type="primary" :loading="uploadLoading" @click="uploadAPK(false)">上传并建档</el-button>
            <el-button :loading="uploadLoading" @click="uploadAPK(true)">替换上传</el-button>
          </div>
        </div>
      </SectionCard>

      <el-skeleton v-if="initialLoading" :rows="8" animated />

      <template v-else-if="!loadError || hasCurrentResult">
        <MetricStrip :items="summaryMetrics" aria-label="安装包摘要" />

        <SectionCard dense>
          <template #title>发布记录</template>
          <template #description>{{ clientMeta.supportsAbi ? '默认查看全部记录，可切换为只看当前家庭可见记录，并支持按状态、版本和 ABI 完整性叠加筛选。' : '默认查看全部记录，可切换为只看当前家庭可见记录，并支持按状态和版本筛选。' }}</template>
          <EmptyState v-if="currentTotalCount === 0" :title="clientMeta.emptyTitle" :description="clientMeta.emptyDesc" />
          <template v-else>
            <div class="table-wrap">
              <el-table v-loading="loading" class="package-table" :data="currentItems" border row-key="id">
                <el-table-column label="版本" min-width="220">
                  <template #default="{ row }">
                    <div class="version-cell">
                      <div class="version-main">
                        <el-tooltip :content="row.version_name || '--'" placement="top">
                          <strong
                            class="compact-text version-name"
                            tabindex="0"
                            :aria-label="`版本名称：${row.version_name || '--'}`"
                          >
                            {{ row.version_name }}
                          </strong>
                        </el-tooltip>
                        <div class="version-meta">
                          <el-tooltip :content="`版本号：${row.version_code ?? '--'}`" placement="top">
                            <span
                              class="compact-text version-code"
                              tabindex="0"
                              :aria-label="`版本号：${row.version_code ?? '--'}`"
                            >
                              ({{ row.version_code }})
                            </span>
                          </el-tooltip>
                          <el-tag v-if="row.latest_recommended" size="small" type="success">推荐</el-tag>
                        </div>
                      </div>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="安装包状态" min-width="280">
                  <template #default="{ row }">
                    <StatusIndicator :label="statusText(row.publish_status)" :tone="statusTone(row.publish_status)" />
                    <el-tooltip :content="abiLine(row)" placement="top">
                      <div
                        class="abi-line compact-text"
                        tabindex="0"
                        :aria-label="`ABI 概览：${abiLine(row)}`"
                      >
                        {{ abiLine(row) }}
                      </div>
                    </el-tooltip>
                    <div v-if="clientMeta.supportsAbi" class="abi-size">
                      <el-tooltip
                        v-for="abi in row.abi_items"
                        :key="abi.id"
                        :content="`${abi.abi} ${formatBytes(abi.file_size)}`"
                        placement="top"
                      >
                        <span
                          class="abi-entry compact-text"
                          tabindex="0"
                          :aria-label="`ABI 文件：${abi.abi}，大小 ${formatBytes(abi.file_size)}`"
                        >
                          {{ abi.abi }} {{ formatBytes(abi.file_size) }}
                        </span>
                      </el-tooltip>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="版本说明" min-width="320">
                  <template #default="{ row }">
                    <el-input
                      v-model="row.draft.release_notes"
                      :aria-label="`${row.version_name} 版本说明`"
                      type="textarea"
                      :rows="2"
                      resize="none"
                      placeholder="给家庭成员看的简短版本说明"
                    />
                    <el-input
                      v-model="row.draft.remarks"
                      :aria-label="`${row.version_name} 管理端备注`"
                      class="remark-input"
                      placeholder="备注（仅管理端）"
                    />
                    <div class="inline-actions">
                      <el-button size="small" :loading="savingId === row.id" @click="saveNotes(row)">保存说明</el-button>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="时间" min-width="220">
                  <template #default="{ row }">
                    <div>上传：{{ formatDateTime(row.original_uploaded_at) }}</div>
                    <div>发布：{{ formatDateTime(row.published_at) }}</div>
                    <div>状态变更：{{ formatDateTime(row.last_status_changed_at) }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="下载" min-width="220">
                  <template #default="{ row }">
                    <div v-if="row.abi_items?.length" class="download-list">
                      <a
                        v-for="abi in row.abi_items"
                        :key="abi.id"
                        class="download-link"
                        :href="downloadHref(row, abi.abi)"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <el-icon><Download /></el-icon>
                        {{ clientMeta.supportsAbi ? `下载 ${abi.abi}` : '下载 APK' }}
                      </a>
                    </div>
                    <span v-else>暂无可下载 APK</span>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="132" fixed="right">
                  <template #default="{ row }">
                    <el-dropdown trigger="click">
                      <el-button :icon="MoreFilled" size="small" :loading="actionId === row.id">更多操作</el-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item
                            v-if="row.publish_status === 'draft'"
                            @click="confirmAction(row, 'publish')"
                          >
                            发布
                          </el-dropdown-item>
                          <el-dropdown-item
                            v-if="row.publish_status === 'published_complete' || row.publish_status === 'published_missing_abi'"
                            @click="confirmAction(row, 'offline')"
                          >
                            下线
                          </el-dropdown-item>
                          <el-dropdown-item
                            v-if="row.publish_status === 'offline'"
                            @click="confirmAction(row, 'restore')"
                          >
                            恢复发布
                          </el-dropdown-item>
                          <el-dropdown-item
                            v-if="row.publish_status === 'draft'"
                            divided
                            @click="confirmAction(row, 'delete')"
                          >
                            删除草稿
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <div class="toolbar-row toolbar-row--end">
              <AdminTablePagination
                v-model:current-page="query.page"
                v-model:page-size="query.page_size"
                layout="total, prev, pager, next"
                :total="currentTotalCount"
                @current-change="load"
                @size-change="load"
              />
            </div>
          </template>
        </SectionCard>
      </template>
    </div>
  </Layout>
</template>

<style scoped>
.app-package-page {
  display: grid;
  gap: var(--space-4);
}

.upload-panel {
  display: grid;
  gap: var(--space-4);
}

.download-qr-card {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 252px;
}

.download-qr-image {
  width: 220px;
  height: 220px;
  object-fit: contain;
}

.upload-actions {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.version-cell {
  display: grid;
  gap: var(--space-2);
}

.version-main {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.version-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}

.version-code {
  flex: 1 1 auto;
}

.version-meta :deep(.el-tag) {
  flex: 0 0 auto;
}

.compact-text {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.abi-line,
.abi-size,
.remark-input {
  margin-top: var(--space-2);
}

.abi-size {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
  color: var(--el-text-color-secondary);
  font-size: var(--text-caption);
}

.inline-actions {
  margin-top: var(--space-2);
}

.download-list {
  display: grid;
  gap: var(--space-2);
}

.download-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--el-color-primary);
  text-decoration: none;
}

.download-link:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.package-table :deep(.el-table__row) {
  height: var(--table-row-height);
}

@media (max-width: 63.9375rem) {
  .client-type-switch :deep(.el-segmented__item) {
    min-height: 44px;
  }
}
</style>
