<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import Toolbar from '../components/base/Toolbar.vue'
import {
  createAdminImageCollection,
  deleteAdminImageCollection,
  getAdminImages,
  getAdminImageViewBlob,
  getAdminImageCollections,
  updateAdminImageCollection
} from '../api/admin'
import { buildImageCollectionPayload, IMAGE_COLLECTION_PREVIEW_PARAMS, revokePreviewURLs } from './imageCollectionManage.helpers'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const editDrawerVisible = ref(false)
const editDrawerSaving = ref(false)
const editingID = ref('')
const imageDrawerVisible = ref(false)
const imageDrawerLoading = ref(false)
const imageDrawerCollection = ref(null)
const imageDrawerItems = ref([])
const imageDrawerTotal = ref(0)
const imageDrawerApplyingCoverID = ref('')
const imageDrawerPreviewURLs = ref({})
const imageDrawerPreviewErrors = ref({})
const imageDrawerCoverPreviewURL = ref('')
const viewportWidth = ref(readViewportWidth())

const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  active: '1'
})

const imageDrawerQuery = reactive({
  page: 1,
  page_size: 18
})

const form = reactive(createEmptyForm())
const editDrawerSnapshot = ref(createEmptyForm())
const editDrawerDirty = computed(() => (
  form.name !== editDrawerSnapshot.value.name
  || form.description !== editDrawerSnapshot.value.description
  || form.cover_url !== editDrawerSnapshot.value.cover_url
  || (form.cover_image_id ?? null) !== (editDrawerSnapshot.value.cover_image_id ?? null)
  || form.sort_order !== editDrawerSnapshot.value.sort_order
  || form.active !== editDrawerSnapshot.value.active
))
const editDrawerSize = computed(() => (viewportWidth.value < 1024 ? '100%' : '560px'))
const activeDrawerCoverSrc = computed(() => {
  const collection = imageDrawerCollection.value
  if (!collection) return ''
  if (collection.cover_image_id) {
    return imageDrawerCoverPreviewURL.value
  }
  return collection.cover_url || ''
})

function createEmptyForm() {
  return {
    name: '',
    description: '',
    cover_url: '',
    cover_image_id: null,
    sort_order: 0,
    active: true
  }
}

function readViewportWidth() {
  if (typeof window === 'undefined') {
    return 1440
  }
  return window.innerWidth
}

function updateViewportWidth() {
  viewportWidth.value = readViewportWidth()
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

function buildListParams() {
  const params = {
    page: query.page,
    page_size: query.page_size,
    q: query.q
  }
  if (query.active === '1') {
    params.active = 1
  } else if (query.active === '0') {
    params.active = 0
  }
  return params
}

function resetForm() {
  Object.assign(form, createEmptyForm())
  editDrawerSnapshot.value = createEmptyForm()
}

function captureFormSnapshot() {
  editDrawerSnapshot.value = {
    name: form.name,
    description: form.description,
    cover_url: form.cover_url,
    cover_image_id: form.cover_image_id ?? null,
    sort_order: form.sort_order,
    active: form.active
  }
}

async function load() {
  loading.value = true
  try {
    const data = await getAdminImageCollections(buildListParams())
    list.value = data.items || []
    total.value = data.total_count || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingID.value = ''
  resetForm()
  captureFormSnapshot()
  editDrawerVisible.value = true
}

function openEdit(row) {
  editingID.value = row.id
  Object.assign(form, {
    name: row.name || '',
    description: row.description || '',
    cover_url: row.manual_cover_url || '',
    cover_image_id: row.cover_image_id || null,
    sort_order: typeof row.sort_order === 'number' ? row.sort_order : 0,
    active: !!row.active
  })
  captureFormSnapshot()
  editDrawerVisible.value = true
}

async function confirmEditDrawerClose() {
  if (!editDrawerDirty.value) {
    return true
  }
  try {
    await ElMessageBox.confirm(
      '未保存的修改将会丢失，确认关闭吗？',
      '确认关闭',
      {
        type: 'warning',
        confirmButtonText: '确认丢弃',
        cancelButtonText: '继续编辑'
      }
    )
    return true
  } catch (_) {
    return false
  }
}

async function requestEditDrawerClose() {
  if (await confirmEditDrawerClose()) {
    // 先把 snapshot 对齐到 form，避免 v-model flip 后 el-drawer before-close 再次弹同款确认
    captureFormSnapshot()
    editDrawerVisible.value = false
  }
}

function handleEditDrawerBeforeClose(done) {
  if (!editDrawerDirty.value) {
    done()
    return
  }
  ElMessageBox.confirm(
    '未保存的修改将会丢失，确认关闭吗？',
    '确认关闭',
    {
      type: 'warning',
      confirmButtonText: '确认丢弃',
      cancelButtonText: '继续编辑'
    }
  )
    .then(() => {
      done()
    })
    .catch(() => {})
}

async function save() {
  if (!form.name || !form.name.trim()) {
    ElMessage.warning('请输入图片合集名称')
    return
  }
  editDrawerSaving.value = true
  try {
    const payload = buildImageCollectionPayload(form)
    if (editingID.value) {
      const updated = await updateAdminImageCollection(editingID.value, payload)
      replaceCollectionRow(updated)
      if (imageDrawerCollection.value?.id === updated.id) {
        imageDrawerCollection.value = updated
      }
      ElMessage.success('图片合集已更新')
    } else {
      await createAdminImageCollection(payload)
      ElMessage.success('图片合集已创建')
    }
    // 保存成功后 snapshot 对齐 form，避免随后 visible flip 触发 before-close 的「未保存」假弹
    captureFormSnapshot()
    editDrawerVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存图片合集失败'))
  } finally {
    editDrawerSaving.value = false
  }
}

async function doDelete(row) {
  const title = row?.name || '该图片合集'
  try {
    await ElMessageBox.confirm(
      `确认删除「${title}」？删除后仅解除关联，不会删除图片。`,
      '删除图片合集',
      {
        type: 'warning',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      }
    )
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    ElMessage.error('删除确认失败，请重试')
    return
  }

  try {
    const result = await deleteAdminImageCollection(row.id)
    const detached = Number(result?.detached_images || 0)
    ElMessage.success(`已删除图片合集，解除 ${detached} 张图片关联（图片未删除）`)
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除图片合集失败'))
  }
}

function replaceCollectionRow(updated) {
  if (!updated?.id) return
  let replaced = false
  list.value = list.value.map((item) => {
    if (item.id !== updated.id) return item
    replaced = true
    return updated
  })
  if (!replaced) {
    return
  }
}

function clearImageDrawerPreviews() {
  imageDrawerPreviewURLs.value = revokePreviewURLs(imageDrawerPreviewURLs.value)
  imageDrawerPreviewErrors.value = {}
  if (imageDrawerCoverPreviewURL.value) {
    URL.revokeObjectURL(imageDrawerCoverPreviewURL.value)
    imageDrawerCoverPreviewURL.value = ''
  }
}

async function readPreviewBlob(blob, fallback) {
  if (!blob?.type?.includes('application/json')) {
    return blob
  }
  const text = await blob.text()
  try {
    const payload = JSON.parse(text)
    throw new Error(payload?.msg || fallback)
  } catch (error) {
    if (error instanceof Error) {
      throw error
    }
    throw new Error(fallback)
  }
}

async function loadImageDrawerPreviews(items, collectionID) {
  clearImageDrawerPreviews()
  if (!items.length) {
    return
  }
  const nextURLs = {}
  const nextErrors = {}
  await Promise.all(
    items.map(async (item) => {
      try {
        const blob = await getAdminImageViewBlob(item.id, IMAGE_COLLECTION_PREVIEW_PARAMS)
        const imageBlob = await readPreviewBlob(blob, '加载缩略图失败')
        nextURLs[item.id] = URL.createObjectURL(imageBlob)
      } catch (error) {
        nextErrors[item.id] = extractErrorMessage(error, '加载缩略图失败')
      }
    })
  )
  if (imageDrawerCollection.value?.id !== collectionID) {
    revokePreviewURLs(nextURLs)
    return
  }
  imageDrawerPreviewURLs.value = nextURLs
  imageDrawerPreviewErrors.value = nextErrors
}

async function loadImageDrawerCoverPreview(collection) {
  if (imageDrawerCoverPreviewURL.value) {
    URL.revokeObjectURL(imageDrawerCoverPreviewURL.value)
    imageDrawerCoverPreviewURL.value = ''
  }
  if (!collection?.cover_image_id) {
    return
  }
  try {
    const blob = await getAdminImageViewBlob(collection.cover_image_id, IMAGE_COLLECTION_PREVIEW_PARAMS)
    const imageBlob = await readPreviewBlob(blob, '加载合集封面失败')
    if (imageDrawerCollection.value?.id !== collection.id || imageDrawerCollection.value?.cover_image_id !== collection.cover_image_id) {
      return
    }
    imageDrawerCoverPreviewURL.value = URL.createObjectURL(imageBlob)
  } catch (_) {
    imageDrawerCoverPreviewURL.value = ''
  }
}

async function loadImageDrawerImages() {
  const collection = imageDrawerCollection.value
  if (!collection?.id) return
  imageDrawerLoading.value = true
  const collectionID = collection.id
  try {
    const data = await getAdminImages({
      page: imageDrawerQuery.page,
      page_size: imageDrawerQuery.page_size,
      collection_id: collection.id,
      active: 1
    })
    if (imageDrawerCollection.value?.id !== collectionID) {
      return
    }
    imageDrawerItems.value = data.items || []
    imageDrawerTotal.value = Number(data.total_count || 0)
    await Promise.all([
      loadImageDrawerPreviews(imageDrawerItems.value, collectionID),
      loadImageDrawerCoverPreview(imageDrawerCollection.value)
    ])
  } finally {
    if (imageDrawerCollection.value?.id === collectionID) {
      imageDrawerLoading.value = false
    }
  }
}

async function openImageDrawer(row) {
  imageDrawerCollection.value = { ...row }
  imageDrawerItems.value = []
  imageDrawerTotal.value = 0
  imageDrawerQuery.page = 1
  imageDrawerVisible.value = true
  await loadImageDrawerImages()
}

async function applyImageDrawerCover(row) {
  const collection = imageDrawerCollection.value
  if (!collection?.id || !row?.id) return
  imageDrawerApplyingCoverID.value = row.id
  try {
    const updated = await updateAdminImageCollection(
      collection.id,
      buildImageCollectionPayload({
        ...collection,
        cover_url: collection.manual_cover_url || collection.cover_url || '',
        cover_image_id: row.id
      })
    )
    imageDrawerCollection.value = updated
    replaceCollectionRow(updated)
    await loadImageDrawerCoverPreview(updated)
    ElMessage.success('已更新图片合集封面')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '设置图片合集封面失败'))
  } finally {
    imageDrawerApplyingCoverID.value = ''
  }
}

function onImageDrawerClosed() {
  clearImageDrawerPreviews()
  imageDrawerCollection.value = null
  imageDrawerItems.value = []
  imageDrawerTotal.value = 0
  imageDrawerApplyingCoverID.value = ''
  imageDrawerLoading.value = false
}

onMounted(() => {
  load()
  updateViewportWidth()
  window.addEventListener('resize', updateViewportWidth)
})

onBeforeUnmount(() => {
  clearImageDrawerPreviews()
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateViewportWidth)
  }
})
</script>

<template>
  <Layout>
    <div class="page-shell page-shell--medium">
      <PageHeader title="图片合集" />

      <Toolbar>
        <template #filters>
          <el-input v-model="query.q" class="collection-search" placeholder="按图片合集名称搜索" clearable @keyup.enter="load" />
          <el-select v-model="query.active" class="collection-status" clearable placeholder="状态筛选">
            <el-option label="全部状态" value="" />
            <el-option label="仅启用" value="1" />
            <el-option label="仅停用" value="0" />
          </el-select>
          <el-button :icon="Search" @click="load">查询</el-button>
        </template>
        <template #actions>
          <el-button type="primary" :icon="Plus" @click="openCreate">创建合集</el-button>
        </template>
      </Toolbar>

      <SectionCard v-loading="loading">
        <template #title>合集列表</template>
        <template #description>维护图片合集并统一管理图片归档关系。</template>

        <el-table v-if="list.length" :data="list" border>
          <el-table-column prop="name" label="图片合集名称" min-width="180" />
          <el-table-column prop="description" label="简介" min-width="260" show-overflow-tooltip />
          <el-table-column label="封面" min-width="240">
            <template #default="{ row }">
              <div class="cover-cell">
                <el-tag v-if="row.cover_image_id" type="success" size="small">已绑定图片封面</el-tag>
                <span class="cover-text">{{ row.cover_url || '-' }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="90" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" width="180" />
          <el-table-column label="操作" width="260">
            <template #default="{ row }">
              <el-button size="small" type="primary" plain @click="openImageDrawer(row)">查看图片</el-button>
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="doDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <EmptyState
          v-else-if="!loading"
          title="暂无图片合集"
          description="先创建一个图片合集，再为图片建立归档关系。"
        >
          <template #action>
            <el-button type="primary" :icon="Plus" @click="openCreate">创建合集</el-button>
          </template>
        </EmptyState>

        <div class="collection-footer">
          <AdminTablePagination
            v-model:current-page="query.page"
            v-model:page-size="query.page_size"
            layout="total, prev, pager, next"
            :total="total"
            @current-change="load"
          />
        </div>
      </SectionCard>
    </div>

    <el-drawer
      v-model="editDrawerVisible"
      :before-close="handleEditDrawerBeforeClose"
      :size="editDrawerSize"
      destroy-on-close
      direction="rtl"
      :title="editingID ? '编辑图片合集' : '创建图片合集'"
      class="collection-edit-drawer"
    >
      <el-form label-width="110px" class="collection-edit-form">
        <SectionCard>
          <template #title>基础信息</template>
          <el-form-item label="图片合集名称">
            <el-input v-model="form.name" placeholder="请输入图片合集名称" />
          </el-form-item>
          <el-form-item label="图片合集简介">
            <el-input v-model="form.description" type="textarea" :rows="3" placeholder="可选，简介用于后台识别" />
          </el-form-item>
        </SectionCard>

        <SectionCard>
          <template #title>封面兜底</template>
          <el-form-item label="封面地址">
            <el-input v-model="form.cover_url" placeholder="https://..." />
          </el-form-item>
          <p class="form-tip">未选择封面图片时，回退使用这里的封面地址。</p>
        </SectionCard>

        <SectionCard>
          <template #title>发布设置</template>
          <el-form-item label="排序值">
            <el-input-number v-model="form.sort_order" :step="1" :precision="0" style="width: 200px" />
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="form.active" active-text="启用" inactive-text="停用" />
          </el-form-item>
        </SectionCard>
      </el-form>

      <div class="collection-edit-drawer__footer">
        <Toolbar dense>
          <template #actions>
            <el-button @click="requestEditDrawerClose">取消</el-button>
            <el-button type="primary" :loading="editDrawerSaving" @click="save">保存</el-button>
          </template>
        </Toolbar>
      </div>
    </el-drawer>

    <el-drawer
      v-model="imageDrawerVisible"
      title="合集关联图片"
      size="920px"
      destroy-on-close
      @closed="onImageDrawerClosed"
    >
      <div v-if="imageDrawerCollection" class="image-drawer">
        <div class="drawer-hero">
          <div class="drawer-cover-card">
            <div class="drawer-cover-media">
              <img v-if="activeDrawerCoverSrc" :src="activeDrawerCoverSrc" alt="合集封面" class="drawer-cover-image" />
              <div v-else class="drawer-cover-empty">暂无封面</div>
            </div>
            <div class="drawer-cover-meta">
              <div class="drawer-cover-title">{{ imageDrawerCollection.name }}</div>
              <div class="drawer-cover-desc">{{ imageDrawerCollection.description || '未填写合集简介' }}</div>
              <div class="drawer-cover-note">
                当前共关联 {{ imageDrawerTotal }} 张图片，点击下方缩略图可将任意关联图片设为合集封面。
              </div>
            </div>
          </div>
        </div>

        <div class="drawer-grid" v-loading="imageDrawerLoading">
          <div v-if="!imageDrawerItems.length && !imageDrawerLoading" class="drawer-empty">该合集下暂无关联图片。</div>
          <div v-else class="thumb-grid">
            <article v-for="item in imageDrawerItems" :key="item.id" class="thumb-card">
              <div class="thumb-media">
                <img v-if="imageDrawerPreviewURLs[item.id]" :src="imageDrawerPreviewURLs[item.id]" :alt="item.title || item.id" class="thumb-image" />
                <div v-else class="thumb-placeholder">
                  {{ imageDrawerPreviewErrors[item.id] ? '加载失败' : '加载中' }}
                </div>
              </div>
              <div class="thumb-body">
                <div class="thumb-title">{{ item.title || item.id }}</div>
                <div class="thumb-meta">{{ item.width }} × {{ item.height }}</div>
                <div class="thumb-actions">
                  <el-tag v-if="imageDrawerCollection.cover_image_id === item.id" type="success">当前封面</el-tag>
                  <el-button
                    v-else
                    size="small"
                    type="primary"
                    :loading="imageDrawerApplyingCoverID === item.id"
                    @click="applyImageDrawerCover(item)"
                  >
                    设为封面
                  </el-button>
                </div>
              </div>
            </article>
          </div>
        </div>

        <div class="action-row drawer-pagination">
          <AdminTablePagination
            v-model:current-page="imageDrawerQuery.page"
            v-model:page-size="imageDrawerQuery.page_size"
            layout="total, prev, pager, next"
            :total="imageDrawerTotal"
            @current-change="loadImageDrawerImages"
          />
        </div>
      </div>
    </el-drawer>
  </Layout>
</template>

<style scoped>
.page-shell--medium {
  display: grid;
  gap: var(--space-5);
}

.collection-search {
  width: min(24rem, 100%);
}

.collection-status {
  width: 10rem;
}

.collection-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-4);
}

.collection-edit-form {
  display: grid;
  gap: var(--space-4);
}

.collection-edit-drawer :deep(.el-drawer__body) {
  display: grid;
  gap: var(--space-4);
  padding-bottom: 0;
}

.collection-edit-drawer__footer {
  position: sticky;
  bottom: 0;
  padding-top: var(--space-3);
  background: linear-gradient(180deg, transparent, var(--bg-canvas) 28%);
}

.form-tip {
  margin-top: 6px;
  color: var(--text-muted);
  font-size: 12px;
  line-height: 1.5;
}

.cover-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cover-text {
  color: var(--text-muted);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-all;
}

.image-drawer {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 100%;
}

.drawer-hero {
  padding-right: 12px;
}

.drawer-cover-card {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(17, 24, 39, 0.96), rgba(30, 41, 59, 0.92));
}

.drawer-cover-media {
  aspect-ratio: 1 / 1;
  border-radius: 14px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.08);
}

.drawer-cover-image,
.thumb-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.drawer-cover-empty,
.thumb-placeholder,
.drawer-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.72);
  text-align: center;
}

.drawer-cover-empty,
.thumb-placeholder {
  width: 100%;
  height: 100%;
  background: rgba(255, 255, 255, 0.08);
}

.drawer-cover-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  justify-content: center;
}

.drawer-cover-title {
  color: #fff;
  font-size: 22px;
  font-weight: 700;
}

.drawer-cover-desc {
  color: rgba(255, 255, 255, 0.78);
  line-height: 1.7;
}

.drawer-cover-note {
  color: rgba(255, 255, 255, 0.68);
  font-size: 13px;
}

.drawer-grid {
  min-height: 240px;
}

.thumb-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.thumb-card {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  overflow: hidden;
  background: var(--el-bg-color);
}

.thumb-media {
  aspect-ratio: 1 / 1;
  background: #0f172a;
}

.thumb-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
}

.thumb-title {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  color: var(--el-text-color-primary);
  word-break: break-word;
}

.thumb-meta {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.thumb-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  min-height: 32px;
}

.drawer-empty {
  min-height: 240px;
  border: 1px dashed var(--el-border-color);
  border-radius: 18px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
}

.drawer-pagination {
  padding-bottom: 8px;
}

@media (max-width: 900px) {
  .drawer-cover-card {
    grid-template-columns: 1fr;
  }
}
</style>
