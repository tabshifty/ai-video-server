<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import { createAdminActor, getAdminActors, updateAdminActor } from '../api/admin'

const loading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 20,
  q: '',
  active: ''
})

const dialogVisible = ref(false)
const saving = ref(false)
const editingID = ref('')
const form = reactive(createEmptyForm())

function createEmptyForm() {
  return {
    name: '',
    aliases: [],
    gender: '',
    country: '',
    birth_date: '',
    avatar_url: '',
    source: 'manual',
    external_id: '',
    notes: '',
    active: true
  }
}

function normalizeAliases(raw) {
  const out = []
  const seen = new Set()
  for (const item of raw || []) {
    const cleaned = String(item || '').trim().replace(/\s+/g, ' ')
    if (!cleaned) continue
    const key = cleaned.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(cleaned)
  }
  return out
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

async function load() {
  loading.value = true
  try {
    const data = await getAdminActors(buildListParams())
    list.value = data.items || []
    total.value = data.total_count || 0
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, createEmptyForm())
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
    aliases: Array.isArray(row.aliases) ? [...row.aliases] : [],
    gender: row.gender || '',
    country: row.country || '',
    birth_date: row.birth_date || '',
    avatar_url: row.avatar_url || '',
    source: row.source || 'manual',
    external_id: row.external_id || '',
    notes: row.notes || '',
    active: !!row.active
  })
  dialogVisible.value = true
}

function buildPayloadFromForm() {
  return {
    name: form.name?.trim() || '',
    aliases: normalizeAliases(form.aliases),
    gender: form.gender || '',
    country: form.country?.trim() || '',
    birth_date: form.birth_date || '',
    avatar_url: form.avatar_url?.trim() || '',
    source: form.source || 'manual',
    external_id: form.external_id?.trim() || '',
    notes: form.notes?.trim() || '',
    active: !!form.active
  }
}

async function save() {
  if (!form.name || !form.name.trim()) {
    ElMessage.warning('请输入演员姓名')
    return
  }
  saving.value = true
  try {
    const payload = buildPayloadFromForm()
    if (editingID.value) {
      await updateAdminActor(editingID.value, payload)
      ElMessage.success('演员信息已更新')
    } else {
      await createAdminActor(payload)
      ElMessage.success('演员已创建')
    }
    dialogVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(error.message || '保存演员失败')
  } finally {
    saving.value = false
  }
}

async function toggleActive(row) {
  const payload = {
    name: row.name || '',
    aliases: Array.isArray(row.aliases) ? row.aliases : [],
    gender: row.gender || '',
    country: row.country || '',
    birth_date: row.birth_date || '',
    avatar_url: row.avatar_url || '',
    source: row.source || 'manual',
    external_id: row.external_id || '',
    notes: row.notes || '',
    active: !row.active
  }
  await updateAdminActor(row.id, payload)
  ElMessage.success(payload.active ? '演员已启用' : '演员已停用')
  await load()
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">演员管理</h1>
          <p class="page-subtitle">支持演员录入、启停用与别名维护</p>
        </div>
      </div>

      <el-card class="soft-card">
        <el-form inline class="filter-form">
          <el-form-item>
            <el-input v-model="query.q" placeholder="按姓名或别名搜索" clearable @keyup.enter="load" />
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
            <el-button type="success" @click="openCreate">新增演员</el-button>
          </el-form-item>
        </el-form>

        <div class="table-wrap">
          <el-table :data="list" border v-loading="loading">
            <el-table-column prop="name" label="姓名" min-width="180" />
            <el-table-column label="别名" min-width="220">
              <template #default="{ row }">
                <div class="alias-list">
                  <el-tag v-for="alias in row.aliases || []" :key="alias" size="small">{{ alias }}</el-tag>
                  <span v-if="!row.aliases || row.aliases.length === 0" class="muted">无</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="gender" label="性别" width="100" />
            <el-table-column prop="country" label="国家/地区" width="120" />
            <el-table-column prop="source" label="来源" width="130" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.active ? 'success' : 'info'">{{ row.active ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" width="180" />
            <el-table-column label="操作" width="220">
              <template #default="{ row }">
                <el-button size="small" @click="openEdit(row)">编辑</el-button>
                <el-button size="small" :type="row.active ? 'warning' : 'success'" @click="toggleActive(row)">
                  {{ row.active ? '停用' : '启用' }}
                </el-button>
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

    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑演员' : '新增演员'" width="680px">
      <el-form label-width="96px">
        <el-form-item label="姓名">
          <el-input v-model="form.name" placeholder="请输入演员姓名" />
        </el-form-item>
        <el-form-item label="别名">
          <el-select
            v-model="form.aliases"
            multiple
            filterable
            allow-create
            default-first-option
            clearable
            placeholder="可输入多个别名"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender" placeholder="请选择性别" clearable style="width: 180px">
            <el-option label="未知" value="" />
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="国家/地区">
          <el-input v-model="form.country" placeholder="例如：中国、日本" />
        </el-form-item>
        <el-form-item label="出生日期">
          <el-date-picker
            v-model="form.birth_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择日期"
            style="width: 220px"
          />
        </el-form-item>
        <el-form-item label="头像地址">
          <el-input v-model="form.avatar_url" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="来源">
          <el-select v-model="form.source" style="width: 220px">
            <el-option label="手动录入" value="manual" />
            <el-option label="TMDB 刮削" value="scrape_tmdb" />
            <el-option label="AV 刮削" value="scrape_av" />
          </el-select>
        </el-form-item>
        <el-form-item label="外部ID">
          <el-input v-model="form.external_id" placeholder="站点演员ID（可选）" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="3" placeholder="补充信息（可选）" />
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

<style scoped>
.alias-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.muted {
  color: #9ca3af;
}
</style>
