<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, FolderOpened, RefreshRight, UploadFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import {
  getAdminArchiveImportBatches,
  getAdminArchiveImportBatchDetail,
  getAdminArchiveImportFileDetail,
  processAdminArchiveImportBatch,
  processAdminArchiveImportFile,
  retryAdminArchiveImportExtract,
  updateAdminArchiveImportFile,
  uploadAdminArchiveImport
} from '../api/admin'

const router = useRouter()
const loading = ref(false)
const uploadLoading = ref(false)
const processingBatch = ref(false)
const processingFileID = ref('')
const batchList = ref([])
const selectedBatch = ref(null)
const selectedBatchFiles = ref([])
const selectedFile = ref(null)
const fileDetailLoading = ref(false)
const fileSaving = ref(false)
const uploadRef = ref(null)

const query = reactive({
  page: 1,
  page_size: 20
})

const uploadForm = reactive({
  title: '',
  default_description: '',
  default_tags: '',
  default_video_collection_ids: '',
  default_image_collection_ids: '',
  has_password: false,
  password: ''
})

const batches = computed(() => batchList.value || [])

function extractErrorMessage(error, fallback) {
  const responseMsg = error?.response?.data?.msg
  if (typeof responseMsg === 'string' && responseMsg.trim() !== '') return responseMsg.trim()
  const message = error?.message
  if (typeof message === 'string' && message.trim() !== '') return message.trim()
  return fallback
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function normalizeSelectionValues(values) {
  return Array.from(
    new Set(
      (Array.isArray(values) ? values : String(values || '').split(/[,，\n]/))
        .map((value) => String(value || '').trim())
        .filter(Boolean)
    )
  )
}

function normalizeUUIDSelection(values) {
  return normalizeSelectionValues(values).filter((value) =>
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  )
}

function batchStatusLabel(status) {
  const map = {
    uploaded: '已上传',
    ready: '待处理',
    processing: '处理中',
    partial: '部分完成',
    completed: '已完成',
    needs_password: '待密码',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function batchStatusType(status) {
  if (status === 'completed') return 'success'
  if (status === 'partial' || status === 'needs_password') return 'warning'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'info'
  return 'primary'
}

function fileStatusLabel(status) {
  const map = {
    pending: '待处理',
    processing: '处理中',
    ready: '已入库',
    existing: '已存在',
    skipped: '已跳过',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function fileStatusType(status) {
  if (status === 'ready' || status === 'existing') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'processing') return 'warning'
  if (status === 'skipped') return 'info'
  return 'primary'
}

async function loadBatches() {
  loading.value = true
  try {
    const data = await getAdminArchiveImportBatches({
      page: query.page,
      page_size: query.page_size
    })
    batchList.value = data.items || []
    if (selectedBatch.value) {
      const matched = batchList.value.find((item) => item.id === selectedBatch.value.id)
      if (matched) {
        selectedBatch.value = matched
      }
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载压缩包批次失败'))
  } finally {
    loading.value = false
  }
}

async function loadBatchDetail(batchID) {
  if (!batchID) return
  loading.value = true
  try {
    const data = await getAdminArchiveImportBatchDetail(batchID)
    selectedBatch.value = data
    selectedBatchFiles.value = data.files || []
    selectedFile.value = null
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载批次详情失败'))
  } finally {
    loading.value = false
  }
}

async function uploadArchive() {
  if (!uploadRef.value?.files?.length && !uploadRef.value?.uploadFiles?.length) {
    ElMessage.warning('请选择一个压缩包文件')
    return
  }
  const input = uploadRef.value?.uploadFiles?.[0]?.raw || uploadRef.value?.files?.[0]?.raw
  if (!(input instanceof File)) {
    ElMessage.warning('请选择一个压缩包文件')
    return
  }

  const formData = new FormData()
  formData.append('file', input)
  if (uploadForm.title.trim()) formData.append('title', uploadForm.title.trim())
  if (uploadForm.default_description.trim()) formData.append('default_description', uploadForm.default_description.trim())
  if (uploadForm.default_tags.trim()) formData.append('default_tags', uploadForm.default_tags.trim())
  if (uploadForm.default_video_collection_ids.trim()) {
    formData.append('default_video_collection_ids', uploadForm.default_video_collection_ids.trim())
  }
  if (uploadForm.default_image_collection_ids.trim()) {
    formData.append('default_image_collection_ids', uploadForm.default_image_collection_ids.trim())
  }
  formData.append('has_password', uploadForm.has_password ? '1' : '0')
  if (uploadForm.has_password && uploadForm.password.trim()) {
    formData.append('password', uploadForm.password.trim())
  }

  uploadLoading.value = true
  try {
    const batch = await uploadAdminArchiveImport(formData)
    ElMessage.success('压缩包已上传')
    await loadBatches()
    await loadBatchDetail(batch.id)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '上传压缩包失败'))
  } finally {
    uploadLoading.value = false
  }
}

async function refreshBatchDetail() {
  if (!selectedBatch.value?.id) return
  await loadBatchDetail(selectedBatch.value.id)
}

async function runBatchProcess() {
  if (!selectedBatch.value?.id) return
  processingBatch.value = true
  try {
    const data = await processAdminArchiveImportBatch(selectedBatch.value.id)
    selectedBatchFiles.value = data.items || []
    await refreshBatchDetail()
    ElMessage.success('已处理当前批次的待处理文件')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '处理批次失败'))
  } finally {
    processingBatch.value = false
  }
}

async function runRetryExtract() {
  if (!selectedBatch.value?.id) return
  const password = String(uploadForm.password || '').trim()
  if (!password) {
    ElMessage.warning('请输入密码后再重试')
    return
  }
  processingBatch.value = true
  try {
    await retryAdminArchiveImportExtract(selectedBatch.value.id, { password })
    ElMessage.success('已重新解包')
    await refreshBatchDetail()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '重试解包失败'))
  } finally {
    processingBatch.value = false
  }
}

async function selectFile(row) {
  if (!row?.id) return
  fileDetailLoading.value = true
  processingFileID.value = row.id
  try {
    const data = await getAdminArchiveImportFileDetail(row.id)
    selectedFile.value = data
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载文件详情失败'))
  } finally {
    fileDetailLoading.value = false
    processingFileID.value = ''
  }
}

async function saveSelectedFile() {
  if (!selectedFile.value?.id) return
  fileSaving.value = true
  try {
    const payload = {
      title: selectedFile.value.title || '',
      description: selectedFile.value.description || '',
      tags: normalizeSelectionValues(selectedFile.value.tags),
      video_type: selectedFile.value.video_type || 'short',
      video_collection_ids: normalizeUUIDSelection(selectedFile.value.video_collection_ids),
      image_collection_ids: normalizeUUIDSelection(selectedFile.value.image_collection_ids)
    }
    const data = await updateAdminArchiveImportFile(selectedFile.value.id, payload)
    selectedFile.value = data
    await refreshBatchDetail()
    ElMessage.success('已保存文件信息')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存文件失败'))
  } finally {
    fileSaving.value = false
  }
}

async function processSelectedFile() {
  if (!selectedFile.value?.id) return
  processingFileID.value = selectedFile.value.id
  try {
    const data = await processAdminArchiveImportFile(selectedFile.value.id)
    selectedFile.value = data
    await refreshBatchDetail()
    ElMessage.success('已处理文件')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '处理文件失败'))
  } finally {
    processingFileID.value = ''
  }
}

function clearUploadForm() {
  uploadForm.title = ''
  uploadForm.default_description = ''
  uploadForm.default_tags = ''
  uploadForm.default_video_collection_ids = ''
  uploadForm.default_image_collection_ids = ''
  uploadForm.has_password = false
  uploadForm.password = ''
  uploadRef.value?.clearFiles?.()
}

function returnToToolbox() {
  router.push('/toolbox')
}

onMounted(async () => {
  await loadBatches()
})
</script>

<template>
  <main class="tool-workspace archive-import-tool">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader title="压缩包导入" subtitle="管理员上传 zip、rar、7z 压缩包后解包、审核并导入视频与图片。" />

      <div class="archive-workspace">
        <div class="archive-workspace__hero">
          <SectionCard class="archive-hero-card">
            <template #title>上传入口</template>
            <template #description>支持密码包，上传时如果勾选“有密码”，请由管理员填写密码。压缩包内不允许嵌套压缩包。</template>
            <template #actions>
              <el-button :icon="FolderOpened" @click="returnToToolbox">返回工具箱</el-button>
              <el-button :icon="RefreshRight" :loading="loading" @click="loadBatches">刷新批次</el-button>
            </template>

            <div class="archive-upload-grid">
              <div class="archive-upload-grid__file">
                <el-upload
                  ref="uploadRef"
                  drag
                  :auto-upload="false"
                  :limit="1"
                  :show-file-list="true"
                  accept=".zip,.rar,.7z"
                >
                  <el-icon class="archive-upload-grid__icon"><UploadFilled /></el-icon>
                  <div class="archive-upload-grid__hint">拖拽或点击选择压缩包文件</div>
                  <div class="archive-upload-grid__subhint">仅支持 zip、rar、7z</div>
                </el-upload>
              </div>

              <el-form label-width="120px" class="archive-upload-form">
                <el-form-item label="批次标题"><el-input v-model="uploadForm.title" placeholder="可不填，默认取文件名" /></el-form-item>
                <el-form-item label="默认说明"><el-input v-model="uploadForm.default_description" type="textarea" :rows="2" /></el-form-item>
                <el-form-item label="默认标签"><el-input v-model="uploadForm.default_tags" placeholder="JSON 数组或逗号分隔" /></el-form-item>
                <el-form-item label="默认视频合集"><el-input v-model="uploadForm.default_video_collection_ids" placeholder="合集 ID JSON 数组或逗号分隔" /></el-form-item>
                <el-form-item label="默认图片合集"><el-input v-model="uploadForm.default_image_collection_ids" placeholder="合集 ID JSON 数组或逗号分隔" /></el-form-item>
                <el-form-item label="有密码">
                  <el-switch v-model="uploadForm.has_password" />
                </el-form-item>
                <el-form-item v-if="uploadForm.has_password" label="压缩包密码">
                  <el-input v-model="uploadForm.password" type="password" show-password placeholder="由管理员填写密码" />
                </el-form-item>
                <el-form-item>
                  <div class="archive-actions">
                    <el-button type="primary" :loading="uploadLoading" @click="uploadArchive">上传压缩包</el-button>
                    <el-button :disabled="uploadLoading" @click="clearUploadForm">清空表单</el-button>
                  </div>
                </el-form-item>
              </el-form>
            </div>
          </SectionCard>

          <SectionCard class="archive-stats-card" dense>
            <template #title>当前批次概览</template>
            <template #description>从最近上传到当前选中批次的处理进度，便于先扫一眼再进入细节。</template>

            <div class="archive-stats-grid">
              <div class="archive-stat">
                <span>批次总数</span>
                <strong class="tabular-num">{{ batches.length }}</strong>
              </div>
              <div class="archive-stat">
                <span>当前已选</span>
                <strong>{{ selectedBatch ? (selectedBatch.title || selectedBatch.original_filename || '已选中') : '未选择' }}</strong>
              </div>
              <div class="archive-stat">
                <span>待处理</span>
                <strong class="tabular-num">{{ selectedBatch ? selectedBatch.processable_entries : 0 }}</strong>
              </div>
              <div class="archive-stat">
                <span>已完成</span>
                <strong class="tabular-num">{{ selectedBatch ? selectedBatch.processed_entries : 0 }}</strong>
              </div>
            </div>
          </SectionCard>
        </div>

        <div class="archive-workspace__grid">
          <SectionCard class="archive-batch-panel">
            <template #title>批次列表</template>
            <template #description>选择一个批次查看文件清单，并按文件级别处理视频、图片和跳过项。</template>

            <EmptyState v-if="!batches.length" title="暂无压缩包批次" description="上传一个压缩包后，批次会出现在这里。" />

            <div v-else class="archive-batch-list">
              <button
                v-for="batch in batches"
                :key="batch.id"
                type="button"
                class="archive-batch-item"
                :class="{ 'is-active': selectedBatch?.id === batch.id }"
                @click="loadBatchDetail(batch.id)"
              >
                <strong>{{ batch.title || batch.original_filename }}</strong>
                <span>{{ batch.original_filename }}</span>
                <span>
                  <el-tag effect="plain" size="small" :type="batchStatusType(batch.status)">{{ batchStatusLabel(batch.status) }}</el-tag>
                </span>
                <span class="tabular-num">{{ batch.processed_entries }}/{{ batch.processable_entries }}</span>
              </button>
            </div>
          </SectionCard>

          <SectionCard v-if="selectedBatch" class="archive-detail-panel">
            <template #title>批次详情</template>
            <template #description>查看当前批次文件清单，逐条修正标题、标签、合集和视频类型后再处理。</template>
            <template #actions>
              <el-button :loading="processingBatch" @click="refreshBatchDetail">刷新详情</el-button>
              <el-button type="primary" :loading="processingBatch" @click="runBatchProcess">处理待处理文件</el-button>
              <el-button v-if="selectedBatch.status === 'needs_password'" :loading="processingBatch" @click="runRetryExtract">重试解包</el-button>
            </template>

            <div class="archive-batch-summary">
              <div><span>标题</span><strong>{{ selectedBatch.title }}</strong></div>
              <div><span>格式</span><strong>{{ selectedBatch.archive_format }}</strong></div>
              <div><span>状态</span><strong>{{ batchStatusLabel(selectedBatch.status) }}</strong></div>
              <div><span>总项</span><strong class="tabular-num">{{ selectedBatch.total_entries }}</strong></div>
              <div><span>可处理</span><strong class="tabular-num">{{ selectedBatch.processable_entries }}</strong></div>
              <div><span>已处理</span><strong class="tabular-num">{{ selectedBatch.processed_entries }}</strong></div>
              <div><span>跳过</span><strong class="tabular-num">{{ selectedBatch.skipped_entries }}</strong></div>
              <div><span>失败</span><strong class="tabular-num">{{ selectedBatch.failed_entries }}</strong></div>
            </div>
            <p v-if="selectedBatch.last_error" class="archive-batch-error">{{ selectedBatch.last_error }}</p>

            <div class="archive-file-layout">
              <div class="archive-file-list">
                <button
                  v-for="file in selectedBatchFiles || []"
                  :key="file.id"
                  type="button"
                  class="archive-file-item"
                  :class="{ 'is-active': selectedFile?.id === file.id }"
                  @click="selectFile(file)"
                >
                  <strong>{{ file.relative_path }}</strong>
                  <span>{{ file.media_kind }} · {{ formatFileSize(file.file_size) }}</span>
                  <span><el-tag size="small" effect="plain" :type="fileStatusType(file.status)">{{ fileStatusLabel(file.status) }}</el-tag></span>
                </button>
              </div>

              <SectionCard dense>
                <template #title>文件详情</template>
                <template #description>选中一个文件后可以修改标题、说明、标签和合集，再单独处理。</template>

                <EmptyState
                  v-if="!selectedFile"
                  title="请选择一个文件"
                  description="左侧列表中选择文件后，这里会显示可编辑内容。"
                />

                <div v-else class="archive-file-detail">
                  <el-form label-width="104px">
                    <el-form-item label="相对路径"><el-input :model-value="selectedFile.relative_path" disabled /></el-form-item>
                    <el-form-item label="类型"><el-input :model-value="selectedFile.media_kind" disabled /></el-form-item>
                    <el-form-item label="视频类型" v-if="selectedFile.media_kind === 'video'">
                      <el-select v-model="selectedFile.video_type" style="width: 100%">
                        <el-option label="短视频" value="short" />
                        <el-option label="电影" value="movie" />
                        <el-option label="剧集分集" value="episode" />
                        <el-option label="AV" value="av" />
                      </el-select>
                    </el-form-item>
                    <el-form-item label="标题"><el-input v-model="selectedFile.title" /></el-form-item>
                    <el-form-item label="说明"><el-input v-model="selectedFile.description" type="textarea" :rows="3" /></el-form-item>
                    <el-form-item label="标签" v-if="selectedFile.media_kind === 'video'">
                      <el-select v-model="selectedFile.tags" multiple filterable allow-create default-first-option style="width: 100%" />
                    </el-form-item>
                    <el-form-item label="视频合集" v-if="selectedFile.media_kind === 'video'">
                      <el-select
                        v-model="selectedFile.video_collection_ids"
                        multiple
                        filterable
                        allow-create
                        default-first-option
                        style="width: 100%"
                        placeholder="合集 ID，支持输入或粘贴"
                      />
                    </el-form-item>
                    <el-form-item label="图片合集" v-if="selectedFile.media_kind === 'image'">
                      <el-select
                        v-model="selectedFile.image_collection_ids"
                        multiple
                        filterable
                        allow-create
                        default-first-option
                        style="width: 100%"
                        placeholder="合集 ID，支持输入或粘贴"
                      />
                    </el-form-item>
                    <el-form-item>
                      <div class="archive-actions">
                        <el-button type="primary" :loading="fileSaving" @click="saveSelectedFile">保存文件</el-button>
                        <el-button :loading="fileDetailLoading || processingFileID === selectedFile.id" @click="processSelectedFile">处理文件</el-button>
                      </div>
                    </el-form-item>
                  </el-form>
                </div>
              </SectionCard>
            </div>
          </SectionCard>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.archive-import-tool {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.archive-workspace {
  display: grid;
  gap: var(--space-5);
}

.archive-workspace__hero {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) minmax(22rem, 0.7fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-hero-card {
  min-width: 0;
}

.archive-stats-card {
  min-width: 0;
}

.archive-stats-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.archive-stat {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.archive-stat span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.archive-stat strong {
  min-width: 0;
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.archive-workspace__grid {
  display: grid;
  grid-template-columns: minmax(20rem, 26rem) minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-batch-panel,
.archive-detail-panel {
  min-width: 0;
}

.archive-upload-grid {
  display: grid;
  grid-template-columns: minmax(18rem, 20rem) minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.archive-upload-grid__file :deep(.el-upload) {
  width: 100%;
}

.archive-upload-grid__file :deep(.el-upload-dragger) {
  width: 100%;
  min-height: 14rem;
  border: 1px dashed var(--line-strong);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, var(--bg-surface) 0%, var(--bg-surface-muted) 100%);
  box-shadow: var(--shadow-xs);
}

.archive-upload-grid__icon {
  margin-top: var(--space-2);
  font-size: 2.25rem;
  color: var(--primary);
}

.archive-upload-grid__hint {
  margin-top: var(--space-3);
  color: var(--text-primary);
  font-size: var(--text-body);
  font-weight: 600;
}

.archive-upload-grid__subhint {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.archive-upload-form {
  min-width: 0;
  padding-top: var(--space-1);
}

.archive-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.archive-batch-list {
  display: grid;
  gap: var(--space-2);
}

.archive-batch-item,
.archive-file-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: var(--space-3);
  align-items: center;
  width: 100%;
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
  color: var(--text-primary);
  text-align: left;
  cursor: pointer;
  transition: border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.archive-batch-item.is-active,
.archive-file-item.is-active {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.archive-batch-item:hover,
.archive-batch-item:focus-visible,
.archive-file-item:hover,
.archive-file-item:focus-visible {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.archive-batch-item:focus-visible,
.archive-file-item:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 2px;
}

.archive-batch-item strong,
.archive-file-item strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.archive-batch-item span,
.archive-file-item span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.archive-batch-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.archive-batch-summary div {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.archive-batch-summary span {
  color: var(--text-muted);
  font-size: var(--text-caption);
}

.archive-batch-summary strong,
.archive-file-detail strong {
  min-width: 0;
}

.archive-batch-error {
  margin: 0 0 var(--space-3);
  color: var(--danger);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  padding: var(--space-3) var(--space-4);
  border: 1px solid color-mix(in srgb, var(--danger) 20%, var(--line-soft));
  border-radius: var(--radius-md);
  background: var(--danger-50);
}

.archive-file-layout {
  display: grid;
  grid-template-columns: minmax(18rem, 24rem) minmax(0, 1fr);
  gap: var(--space-4);
}

.archive-file-list {
  display: grid;
  gap: var(--space-2);
  align-content: start;
}

.archive-file-detail {
  min-width: 0;
  padding-top: var(--space-1);
}

@media (max-width: 64rem) {
  .archive-workspace__hero,
  .archive-workspace__grid,
  .archive-upload-grid,
  .archive-file-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 48rem) {
  .archive-workspace__hero,
  .archive-workspace__grid {
    gap: var(--space-3);
  }

  .archive-stats-grid {
    grid-template-columns: 1fr;
  }

  .archive-batch-item,
  .archive-file-item {
    grid-template-columns: minmax(0, 1fr) auto;
    row-gap: var(--space-2);
  }

  .archive-batch-item span:nth-child(2),
  .archive-file-item span:nth-child(2) {
    grid-column: 1;
  }

  .archive-batch-item span:nth-child(3),
  .archive-file-item span:nth-child(3) {
    justify-self: end;
  }

  .archive-batch-item span:last-child,
  .archive-file-item span:last-child {
    grid-column: 2;
    justify-self: end;
  }

  .archive-upload-grid__file :deep(.el-upload-dragger) {
    min-height: 12rem;
  }
}
</style>
