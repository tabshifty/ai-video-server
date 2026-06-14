<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Refresh, UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatCard from '../components/base/StatCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
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
    pageTitle: '安装包管理',
    pageSubtitle: '同一套分发模型，按客户端类型分轨管理手机端与 TV 端安装包。'
  },
  android_phone: {
    label: '手机端',
    shortLabel: '手机',
    packageName: 'com.chee.videos',
    supportsAbi: false,
    uploadTip: '仅接收 Release APK，上传首个 APK 会自动建档。',
    emptyTitle: '暂无手机端发布记录',
    emptyDesc: '先上传一个 APK 创建首条记录。',
    pageTitle: '安装包管理',
    pageSubtitle: '同一套分发模型，按客户端类型分轨管理手机端与 TV 端安装包。'
  }
}

const clientType = ref('android_tv')
const loading = ref(false)
const uploadLoading = ref(false)
const savingId = ref(0)
const actionId = ref(0)
const uploadFiles = ref([])
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
const visibleCount = computed(() => data.items.filter((item) => item.visible_to_family).length)
const draftCount = computed(() => data.items.filter((item) => item.publish_status === 'draft').length)
const latestItem = computed(() => data.items.find((item) => item.latest_recommended) || null)
const missingCount = computed(() => {
  if (clientMeta.value.supportsAbi) return data.items.filter((item) => !item.abi_complete).length
  return data.items.filter((item) => item.publish_status === 'draft').length
})

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') return responseMsg.trim()
  if (typeof error?.message === 'string' && error.message.trim() !== '') return error.message.trim()
  return fallback
}

function formatDateTime(value) {
  if (!value) return '暂无'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toLocaleString('zh-CN', { hour12: false })
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

function statusTagType(status) {
  if (!clientMeta.value.supportsAbi) {
    if (status === 'published_complete') return 'success'
    if (status === 'offline') return 'info'
    return ''
  }
  if (status === 'published_complete') return 'success'
  if (status === 'published_missing_abi') return 'warning'
  if (status === 'offline') return 'info'
  return ''
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
  loading.value = true
  try {
    const params = {
      page: query.page,
      page_size: query.page_size,
      current_published: query.current_published ? 1 : 0,
      client_type: clientType.value
    }
    if (query.q.trim()) params.q = query.q.trim()
    if (query.status) params.status = query.status
    if (query.abi_completeness && clientMeta.value.supportsAbi) params.abi_completeness = query.abi_completeness
    applyResult(await getAdminTVAppReleases(params))
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载安装包列表失败'))
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.page = 1
  query.q = ''
  query.status = ''
  query.abi_completeness = ''
  query.current_published = false
}

function changeClientType(nextType) {
  if (!CLIENTS[nextType]) return
  if (clientType.value === nextType) return
  clientType.value = nextType
  resetQuery()
  uploadFiles.value = []
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
  await ElMessageBox.confirm(
    `确认${actionText} ${item.version_name} (${item.version_code}) 吗？`,
    `${actionText}确认`,
    {
      type: action === 'delete' ? 'warning' : 'info',
      confirmButtonText: '确认',
      cancelButtonText: '取消'
    }
  )
  await runDangerAction(item, action)
}

function downloadHref(item, abi) {
  return downloadAdminTVAppReleaseURL(item.id, abi, clientType.value)
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page-shell app-package-page">
      <PageHeader :title="clientMeta.pageTitle" :subtitle="clientMeta.pageSubtitle">
        <template #actions>
          <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
        </template>
      </PageHeader>

      <Toolbar>
        <template #filters>
          <el-segmented :model-value="clientType" :options="[
            { label: 'TV 端', value: 'android_tv' },
            { label: '手机端', value: 'android_phone' }
          ]" @update:modelValue="changeClientType" />
          <el-input v-model="query.q" clearable :placeholder="clientMeta.supportsAbi ? '搜索版本号 / versionName / 时间' : '搜索版本号 / versionName / 时间'" style="width: 240px" @keyup.enter="load" />
          <el-select v-model="query.status" clearable placeholder="状态" style="width: 180px" @change="load">
            <el-option label="草稿" value="draft" />
            <el-option v-if="clientMeta.supportsAbi" label="已发布-完整" value="published_complete" />
            <el-option v-if="clientMeta.supportsAbi" label="已发布-缺少 ABI" value="published_missing_abi" />
            <el-option v-else label="已发布" value="published_complete" />
            <el-option label="已下线" value="offline" />
          </el-select>
          <el-select v-if="clientMeta.supportsAbi" v-model="query.abi_completeness" clearable placeholder="ABI 完整性" style="width: 180px" @change="load">
            <el-option label="完整" value="complete" />
            <el-option label="缺少 ABI" value="missing" />
            <el-option label="空记录" value="empty" />
          </el-select>
          <el-switch v-model="query.current_published" active-text="只看家庭可见" inactive-text="查看全部" @change="load" />
        </template>
        <template #actions>
          <el-button type="primary" :icon="UploadFilled" :loading="uploadLoading" @click="uploadAPK(false)">上传 APK</el-button>
        </template>
      </Toolbar>

      <section class="stat-grid">
        <StatCard :label="clientMeta.supportsAbi ? '当前可见记录' : '当前可见记录'" :value="visibleCount" />
        <StatCard :label="clientMeta.supportsAbi ? '缺少 ABI' : '草稿'" :value="missingCount" />
        <StatCard label="草稿" :value="draftCount" />
        <StatCard label="推荐版本" :value="latestItem ? `${latestItem.version_name} (${latestItem.version_code})` : '暂无'" />
      </section>

      <SectionCard>
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

      <SectionCard>
        <template #title>发布记录</template>
        <template #description>{{ clientMeta.supportsAbi ? '默认查看全部记录，可切换为只看当前家庭可见记录，并支持按状态、版本和 ABI 完整性叠加筛选。' : '默认查看全部记录，可切换为只看当前家庭可见记录，并支持按状态和版本筛选。' }}</template>
        <EmptyState v-if="!data.total_count" :title="clientMeta.emptyTitle" :description="clientMeta.emptyDesc" />
        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="data.items" border row-key="id">
              <el-table-column label="版本" min-width="220">
                <template #default="{ row }">
                  <div class="version-cell">
                    <div class="version-main">
                      <strong>{{ row.version_name }}</strong>
                      <span>({{ row.version_code }})</span>
                      <el-tag v-if="row.latest_recommended" size="small" type="success">推荐</el-tag>
                    </div>
                    <div class="version-sub">{{ statusText(row.publish_status) }}</div>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="安装包状态" min-width="280">
                <template #default="{ row }">
                  <el-tag :type="statusTagType(row.publish_status)" effect="plain">{{ statusText(row.publish_status) }}</el-tag>
                  <div class="abi-line">{{ abiLine(row) }}</div>
                  <div v-if="clientMeta.supportsAbi" class="abi-size">
                    <span v-for="abi in row.abi_items" :key="abi.id">{{ abi.abi }} {{ formatBytes(abi.file_size) }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="版本说明" min-width="320">
                <template #default="{ row }">
                  <el-input
                    v-model="row.draft.release_notes"
                    type="textarea"
                    :rows="2"
                    resize="none"
                    placeholder="给家庭成员看的简短版本说明"
                  />
                  <el-input
                    v-model="row.draft.remarks"
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
              <el-table-column label="操作" width="240" fixed="right">
                <template #default="{ row }">
                  <div class="action-stack">
                    <el-button
                      v-if="row.publish_status === 'draft'"
                      size="small"
                      type="primary"
                      :loading="actionId === row.id"
                      @click="confirmAction(row, 'publish')"
                    >
                      发布
                    </el-button>
                    <el-button
                      v-if="row.publish_status === 'published_complete' || row.publish_status === 'published_missing_abi'"
                      size="small"
                      :loading="actionId === row.id"
                      @click="confirmAction(row, 'offline')"
                    >
                      下线
                    </el-button>
                    <el-button
                      v-if="row.publish_status === 'offline'"
                      size="small"
                      :loading="actionId === row.id"
                      @click="confirmAction(row, 'restore')"
                    >
                      恢复发布
                    </el-button>
                    <el-button
                      v-if="row.publish_status === 'draft'"
                      size="small"
                      type="danger"
                      :loading="actionId === row.id"
                      @click="confirmAction(row, 'delete')"
                    >
                      删除草稿
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="toolbar-row toolbar-row--end">
            <AdminTablePagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="data.total_count"
              @current-change="load"
              @size-change="load"
            />
          </div>
        </template>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.app-package-page {
  display: grid;
  gap: var(--space-4);
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.upload-panel {
  display: grid;
  gap: 16px;
}

.upload-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.version-cell {
  display: grid;
  gap: 6px;
}

.version-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.version-sub,
.abi-line,
.abi-size,
.remark-input {
  margin-top: 8px;
}

.abi-size {
  display: grid;
  gap: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.inline-actions {
  margin-top: 8px;
}

.download-list,
.action-stack {
  display: grid;
  gap: 8px;
}

.download-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--el-color-primary);
  text-decoration: none;
}

@media (max-width: 48rem) {
  .stat-grid {
    grid-template-columns: 1fr;
  }
}
</style>
