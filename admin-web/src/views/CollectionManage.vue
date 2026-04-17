<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import { createAdminCollection, deleteAdminCollection, getAdminCollections, updateAdminCollection } from '../api/admin'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const saving = ref(false)
const editingID = ref('')

const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  active: '1'
})

const form = reactive(createEmptyForm())

function createEmptyForm() {
  return {
    name: '',
    description: '',
    cover_url: '',
    sort_order: 0,
    active: true
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
}

function normalizeSortOrder(raw) {
  const value = Number(raw)
  if (!Number.isFinite(value)) return 0
  return Math.trunc(value)
}

function buildPayloadFromForm() {
  return {
    name: form.name?.trim() || '',
    description: form.description?.trim() || '',
    cover_url: form.cover_url?.trim() || '',
    sort_order: normalizeSortOrder(form.sort_order),
    active: !!form.active
  }
}

async function load() {
  loading.value = true
  try {
    const data = await getAdminCollections(buildListParams())
    list.value = data.items || []
    total.value = data.total_count || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingID.value = ''
  resetForm()
  dialogVisible.value = true
}

function openEdit(row) {
  editingID.value = row.id
  Object.assign(form, {
    name: row.name || '',
    description: row.description || '',
    cover_url: row.cover_url || '',
    sort_order: typeof row.sort_order === 'number' ? row.sort_order : 0,
    active: !!row.active
  })
  dialogVisible.value = true
}

async function save() {
  if (!form.name || !form.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }
  saving.value = true
  try {
    const payload = buildPayloadFromForm()
    if (editingID.value) {
      await updateAdminCollection(editingID.value, payload)
      ElMessage.success('合集已更新')
    } else {
      await createAdminCollection(payload)
      ElMessage.success('合集已创建')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存合集失败'))
  } finally {
    saving.value = false
  }
}

async function doDelete(row) {
  const title = row?.name || '该合集'
  await ElMessageBox.confirm(
    `确认删除「${title}」？删除后仅解除关联，不会删除视频。`,
    '删除合集',
    {
      type: 'warning',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    }
  )
  const result = await deleteAdminCollection(row.id)
  const detached = Number(result?.detached_videos || 0)
  ElMessage.success(`已删除合集，解除 ${detached} 个视频关联（视频未删除）`)
  await load()
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">合集管理</h1>
          <p class="page-subtitle">支持短视频合集的新增、编辑、停用与删除</p>
        </div>
      </div>

      <el-card class="soft-card">
        <el-form inline class="filter-form">
          <el-form-item>
            <el-input v-model="query.q" placeholder="按合集名称搜索" clearable @keyup.enter="load" />
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.active" style="width: 150px" clearable placeholder="状态筛选">
              <el-option label="全部状态" value="" />
              <el-option label="仅启用" value="1" />
              <el-option label="仅停用" value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="load">查询</el-button>
          </el-form-item>
          <el-form-item>
            <el-button type="success" @click="openCreate">新增合集</el-button>
          </el-form-item>
        </el-form>

        <div class="table-wrap">
          <el-table :data="list" border v-loading="loading">
            <el-table-column prop="name" label="合集名称" min-width="180" />
            <el-table-column prop="description" label="简介" min-width="260" show-overflow-tooltip />
            <el-table-column prop="cover_url" label="封面地址" min-width="240" show-overflow-tooltip />
            <el-table-column prop="sort_order" label="排序" width="90" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" width="180" />
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button size="small" @click="openEdit(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑合集' : '新增合集'" width="680px">
      <el-form label-width="100px">
        <el-form-item label="合集名称">
          <el-input v-model="form.name" placeholder="请输入合集名称" />
        </el-form-item>
        <el-form-item label="合集简介">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="可选，简介用于后台识别" />
        </el-form-item>
        <el-form-item label="封面地址">
          <el-input v-model="form.cover_url" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="排序值">
          <el-input-number v-model="form.sort_order" :step="1" :precision="0" style="width: 200px" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="form.active" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>
