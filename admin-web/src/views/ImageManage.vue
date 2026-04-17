<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import {
  deleteAdminImage,
  getAdminActors,
  getAdminImageCollections,
  getAdminImageDetail,
  getAdminImageViewBlob,
  getAdminImages,
  updateAdminImage,
  uploadAdminImages
} from '../api/admin'

const loading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  status: '',
  active: '1',
  actor_id: '',
  collection_id: ''
})

const uploadRef = ref(null)
const uploadFileList = ref([])
const uploading = ref(false)
const uploadForm = reactive({
  description: '',
  actor_tokens: [],
  collection_ids: []
})
const uploadSummary = ref(null)

const detailVisible = ref(false)
const detail = ref(null)
const saving = ref(false)

const actorOptions = ref([])
const loadingActors = ref(false)
const imageCollectionOptions = ref([])
const loadingCollections = ref(false)
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

const preview = reactive({
  loading: false,
  url: '',
  error: '',
  w: 0,
  h: 0,
  fit: 'inside',
  q: 82,
  zoom: 100
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

function statusLabel(status) {
  const map = {
    ready: '可用',
    failed: '失败'
  }
  return map[status] || status || '-'
}

function statusTagType(status) {
  if (status === 'ready') return 'success'
  if (status === 'failed') return 'danger'
  return 'info'
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
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
    const cleaned = value.replace(/\s+/g, ' ')
    const key = cleaned.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(cleaned)
  }
  return { actorIDs, actorNames }
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

function mergeActorOptions(actors = []) {
  const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
  for (const actor of actors) {
    if (!actor?.id) continue
    optionMap.set(actor.id, { value: actor.id, label: actor.name || actor.id })
  }
  actorOptions.value = Array.from(optionMap.values())
}

function mergeCollectionOptions(collections = []) {
  const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
  for (const collection of collections) {
    if (!collection?.id) continue
    optionMap.set(collection.id, { value: collection.id, label: collection.name || collection.id })
  }
  imageCollectionOptions.value = Array.from(optionMap.values())
}

function buildListParams() {
  const params = {
    page: query.page,
    page_size: query.page_size,
    q: query.q,
    status: query.status
  }
  if (query.active === '1') {
    params.active = 1
  } else if (query.active === '0') {
    params.active = 0
  }
  if (query.actor_id) {
    params.actor_id = query.actor_id
  }
  if (query.collection_id) {
    params.collection_id = query.collection_id
  }
  return params
}

async function load() {
  loading.value = true
  try {
    const data = await getAdminImages(buildListParams())
    list.value = data.items || []
    total.value = data.total_count || 0
  } finally {
    loading.value = false
  }
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
    const optionMap = new Map(actorOptions.value.map((item) => [item.value, item]))
    for (const item of data.items || []) {
      optionMap.set(item.id, { value: item.id, label: item.name })
    }
    actorOptions.value = Array.from(optionMap.values())
  } finally {
    loadingActors.value = false
  }
}

async function searchImageCollections(keyword = '') {
  loadingCollections.value = true
  try {
    const data = await getAdminImageCollections({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 50
    })
    const optionMap = new Map(imageCollectionOptions.value.map((item) => [item.value, item]))
    for (const item of data.items || []) {
      optionMap.set(item.id, { value: item.id, label: item.name })
    }
    imageCollectionOptions.value = Array.from(optionMap.values())
  } finally {
    loadingCollections.value = false
  }
}

function resetUploadForm() {
  uploadForm.description = ''
  uploadForm.actor_tokens = []
  uploadForm.collection_ids = []
}

function handleUploadExceed() {
  ElMessage.warning('单次最多选择 100 张图片')
}

async function submitUpload() {
  if (uploadFileList.value.length === 0) {
    ElMessage.warning('请先选择图片文件')
    return
  }
  const formData = new FormData()
  for (const item of uploadFileList.value) {
    if (item?.raw instanceof File) {
      formData.append('files', item.raw)
    }
  }
  if (!formData.has('files')) {
    ElMessage.warning('未读取到可上传的图片文件')
    return
  }

  const { actorIDs, actorNames } = splitActorSelection(uploadForm.actor_tokens)
  const collectionIDs = normalizeCollectionSelection(uploadForm.collection_ids)
  if (uploadForm.description?.trim()) {
    formData.append('description', uploadForm.description.trim())
  }
  if (actorIDs.length > 0) {
    formData.append('actor_ids', JSON.stringify(actorIDs))
  }
  if (actorNames.length > 0) {
    formData.append('actor_names', JSON.stringify(actorNames))
  }
  if (collectionIDs.length > 0) {
    formData.append('collection_ids', JSON.stringify(collectionIDs))
  }

  uploading.value = true
  try {
    const result = await uploadAdminImages(formData)
    uploadSummary.value = result
    const success = Number(result?.success_count || 0)
    const failed = Number(result?.failed_count || 0)
    ElMessage.success(`上传完成：成功 ${success}，失败 ${failed}`)
    uploadRef.value?.clearFiles()
    uploadFileList.value = []
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '批量上传图片失败'))
  } finally {
    uploading.value = false
  }
}

async function showDetail(row) {
  const data = await getAdminImageDetail(row.id)
  data.actor_tokens = (data.actors || []).map((item) => item.id)
  data.collection_ids = (data.collections || []).map((item) => item.id)
  data.metadata_text = JSON.stringify(data.metadata || {}, null, 2)
  detail.value = data
  preview.w = 0
  preview.h = 0
  preview.fit = 'inside'
  preview.q = 82
  preview.zoom = 100
  mergeActorOptions(data.actors || [])
  mergeCollectionOptions(data.collections || [])
  detailVisible.value = true
  await Promise.all([searchActors(''), searchImageCollections('')])
  await loadPreview()
}

function parseMetadata(text) {
  const raw = String(text || '').trim()
  if (!raw) return {}
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('metadata 必须是 JSON 对象')
  }
  return parsed
}

async function saveDetail() {
  if (!detail.value?.id) return
  if (!detail.value.title || !detail.value.title.trim()) {
    ElMessage.warning('图片标题不能为空')
    return
  }
  let metadata
  try {
    metadata = parseMetadata(detail.value.metadata_text)
  } catch (error) {
    ElMessage.error(error.message || 'metadata JSON 格式错误')
    return
  }

  const { actorIDs, actorNames } = splitActorSelection(detail.value.actor_tokens)
  const payload = {
    title: detail.value.title.trim(),
    description: detail.value.description || '',
    active: !!detail.value.active,
    metadata,
    actor_ids: actorIDs,
    actor_names: actorNames,
    collection_ids: normalizeCollectionSelection(detail.value.collection_ids)
  }

  saving.value = true
  try {
    await updateAdminImage(detail.value.id, payload)
    ElMessage.success('图片信息已更新')
    detailVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存图片信息失败'))
  } finally {
    saving.value = false
  }
}

async function toggleActive(row) {
  await updateAdminImage(row.id, { active: !row.active })
  ElMessage.success(!row.active ? '图片已启用' : '图片已停用')
  await load()
}

async function doDelete(row) {
  await ElMessageBox.confirm(`确认删除图片「${row.title || row.id}」？`, '删除图片', {
    type: 'warning',
    confirmButtonText: '确认删除',
    cancelButtonText: '取消'
  })
  await deleteAdminImage(row.id)
  ElMessage.success('图片已删除')
  await load()
}

function currentPreviewParams() {
  const params = {}
  if (Number(preview.w) > 0) {
    params.w = Number(preview.w)
  }
  if (Number(preview.h) > 0) {
    params.h = Number(preview.h)
  }
  if (preview.fit) {
    params.fit = preview.fit
  }
  if (Number(preview.q) > 0) {
    params.q = Number(preview.q)
  }
  return params
}

async function loadPreview() {
  if (!detail.value?.id) return
  if (preview.url) {
    URL.revokeObjectURL(preview.url)
    preview.url = ''
  }
  preview.error = ''
  preview.loading = true
  try {
    const blob = await getAdminImageViewBlob(detail.value.id, currentPreviewParams())
    if (blob?.type?.includes('application/json')) {
      const text = await blob.text()
      let payload
      try {
        payload = JSON.parse(text)
      } catch (_) {
        payload = null
      }
      throw new Error(payload?.msg || '加载图片预览失败')
    }
    preview.url = URL.createObjectURL(blob)
  } catch (error) {
    preview.error = extractErrorMessage(error, '加载图片预览失败')
  } finally {
    preview.loading = false
  }
}

function clearPreview() {
  if (preview.url) {
    URL.revokeObjectURL(preview.url)
    preview.url = ''
  }
  preview.error = ''
}

function onDetailClosed() {
  clearPreview()
  detail.value = null
}

onMounted(async () => {
  await Promise.all([load(), searchActors(''), searchImageCollections('')])
})

onBeforeUnmount(() => {
  clearPreview()
})
</script>

<template>
  <Layout>
    <div class="page image-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">图片管理</h1>
          <p class="page-subtitle">支持批量上传、演员关联、图片合集归档和图片缩放查看</p>
        </div>
      </div>

      <el-card class="soft-card">
        <template #header>批量上传</template>
        <el-form label-width="100px" class="upload-form">
          <el-form-item label="批量选图">
            <el-upload
              ref="uploadRef"
              v-model:file-list="uploadFileList"
              drag
              multiple
              :auto-upload="false"
              :limit="100"
              accept="image/*"
              @exceed="handleUploadExceed"
            >
              <div class="el-upload__text">将图片拖到这里，或<em>点击选择文件</em></div>
              <template #tip>
                <div class="el-upload__tip">
                  支持批量上传；非 GIF 文件将自动转换并保存为 WebP，源文件会在转换后删除。
                </div>
              </template>
            </el-upload>
          </el-form-item>
          <el-form-item label="默认演员">
            <el-select
              v-model="uploadForm.actor_tokens"
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
          <el-form-item label="默认合集">
            <el-select
              v-model="uploadForm.collection_ids"
              multiple
              filterable
              remote
              reserve-keyword
              clearable
              collapse-tags
              collapse-tags-tooltip
              :remote-method="searchImageCollections"
              :loading="loadingCollections"
              placeholder="可选，可多选"
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
          <el-form-item label="默认备注">
            <el-input
              v-model="uploadForm.description"
              type="textarea"
              :rows="2"
              placeholder="可选，批量上传时会作为所有图片的描述"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="uploading" @click="submitUpload">开始上传</el-button>
            <el-button @click="resetUploadForm">清空默认参数</el-button>
          </el-form-item>
        </el-form>

        <div v-if="uploadSummary" class="upload-summary">
          <el-alert
            type="success"
            :closable="false"
            :title="`上传结果：总计 ${uploadSummary.total_count || 0}，成功 ${uploadSummary.success_count || 0}，失败 ${uploadSummary.failed_count || 0}`"
          />
          <el-table :data="uploadSummary.items || []" size="small" border>
            <el-table-column prop="filename" label="文件名" min-width="220" show-overflow-tooltip />
            <el-table-column label="结果" width="120">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="image_id" label="图片ID" min-width="220" />
            <el-table-column prop="error" label="失败原因" min-width="200" show-overflow-tooltip />
          </el-table>
        </div>
      </el-card>

      <el-card class="soft-card">
        <template #header>图片列表</template>
        <el-form inline class="filter-form">
          <el-form-item>
            <el-input v-model="query.q" placeholder="按标题或描述搜索" clearable @keyup.enter="load" />
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.status" placeholder="状态筛选" clearable style="width: 120px">
              <el-option label="可用" value="ready" />
              <el-option label="失败" value="failed" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.active" placeholder="启用状态" clearable style="width: 120px">
              <el-option label="全部状态" value="" />
              <el-option label="仅启用" value="1" />
              <el-option label="仅停用" value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select
              v-model="query.actor_id"
              filterable
              remote
              clearable
              reserve-keyword
              placeholder="按演员筛选"
              style="width: 180px"
              :remote-method="searchActors"
              :loading="loadingActors"
            >
              <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select
              v-model="query.collection_id"
              filterable
              remote
              clearable
              reserve-keyword
              placeholder="按图片合集筛选"
              style="width: 190px"
              :remote-method="searchImageCollections"
              :loading="loadingCollections"
            >
              <el-option
                v-for="collection in imageCollectionOptions"
                :key="collection.value"
                :label="collection.label"
                :value="collection.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="load">查询</el-button>
          </el-form-item>
        </el-form>

        <div class="table-wrap">
          <el-table :data="list" border v-loading="loading">
            <el-table-column prop="title" label="标题" min-width="220" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="启用" width="90">
              <template #default="{ row }">
                <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? '是' : '否' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="stored_mime" label="格式" width="110" />
            <el-table-column label="尺寸" width="130">
              <template #default="{ row }">{{ row.width || 0 }} x {{ row.height || 0 }}</template>
            </el-table-column>
            <el-table-column label="文件大小" width="130">
              <template #default="{ row }">{{ formatFileSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column prop="created_at" label="上传时间" width="180" />
            <el-table-column label="操作" width="250">
              <template #default="{ row }">
                <el-button size="small" @click="showDetail(row)">详情</el-button>
                <el-button size="small" :type="row.active ? 'warning' : 'success'" @click="toggleActive(row)">
                  {{ row.active ? '停用' : '启用' }}
                </el-button>
                <el-button size="small" type="danger" @click="doDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="action-row">
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.page_size"
            layout="total, prev, pager, next"
            :total="total"
            @current-change="load"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailVisible" title="图片详情" width="980px" @closed="onDetailClosed">
      <div v-if="detail" class="detail-grid">
        <el-card class="detail-card" shadow="never">
          <el-form label-width="100px">
            <el-form-item label="标题">
              <el-input v-model="detail.title" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="detail.description" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="启用状态">
              <el-switch v-model="detail.active" active-text="启用" inactive-text="停用" />
            </el-form-item>
            <el-form-item label="关联演员">
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
            <el-form-item label="关联合集">
              <el-select
                v-model="detail.collection_ids"
                multiple
                filterable
                remote
                reserve-keyword
                clearable
                collapse-tags
                collapse-tags-tooltip
                :remote-method="searchImageCollections"
                :loading="loadingCollections"
                placeholder="可选，可多选"
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
            <el-form-item label="Metadata">
              <el-input v-model="detail.metadata_text" type="textarea" :rows="8" />
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="detail-card" shadow="never">
          <template #header>图片查看（支持缩放接口）</template>
          <div class="preview-controls">
            <el-input-number v-model="preview.w" :min="0" :step="50" controls-position="right" placeholder="宽" />
            <el-input-number v-model="preview.h" :min="0" :step="50" controls-position="right" placeholder="高" />
            <el-select v-model="preview.fit" style="width: 120px">
              <el-option label="inside" value="inside" />
              <el-option label="cover" value="cover" />
              <el-option label="contain" value="contain" />
            </el-select>
            <el-input-number v-model="preview.q" :min="1" :max="100" :step="1" controls-position="right" />
            <el-button :loading="preview.loading" type="primary" plain @click="loadPreview">应用缩放接口</el-button>
          </div>
          <div class="zoom-row">
            <span>前端缩放</span>
            <el-slider v-model="preview.zoom" :min="20" :max="300" :step="5" style="flex: 1" />
            <span class="zoom-label">{{ preview.zoom }}%</span>
          </div>
          <el-alert v-if="preview.error" type="error" :title="preview.error" :closable="false" />
          <div class="preview-canvas" v-loading="preview.loading">
            <img
              v-if="preview.url"
              :src="preview.url"
              alt="preview"
              class="preview-image"
              :style="{ transform: `scale(${preview.zoom / 100})` }"
            />
            <div v-else class="preview-placeholder">暂无图片预览</div>
          </div>
          <div class="preview-meta">
            <span>原始尺寸：{{ detail.width || 0 }} x {{ detail.height || 0 }}</span>
            <span>存储格式：{{ detail.stored_mime || '-' }}</span>
            <span>文件大小：{{ formatFileSize(detail.file_size) }}</span>
          </div>
        </el-card>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveDetail">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.image-page {
  gap: 14px;
}

.upload-form {
  max-width: 980px;
}

.upload-summary {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1.1fr 1fr;
  gap: 12px;
}

.detail-card {
  border-radius: 12px;
  border: 1px solid rgba(136, 19, 55, 0.12);
}

.preview-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 10px;
}

.zoom-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.zoom-label {
  width: 50px;
  text-align: right;
  font-family: 'Fira Code', monospace;
  color: #7f1d1d;
}

.preview-canvas {
  min-height: 340px;
  max-height: 440px;
  border: 1px dashed rgba(136, 19, 55, 0.2);
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(255, 241, 242, 0.8), rgba(248, 250, 252, 0.9));
  overflow: auto;
  display: grid;
  place-items: center;
  padding: 16px;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  transform-origin: center center;
  transition: transform 160ms ease;
}

.preview-placeholder {
  color: #6b7280;
}

.preview-meta {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #4b5563;
  font-size: 12px;
}

@media (max-width: 1200px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
