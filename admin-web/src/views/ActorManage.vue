<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import {
  createAdminActor,
  getAdminActors,
  scrapeAdminActorPreview,
  updateAdminActor
} from '../api/admin'

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
const scraping = ref(false)
const editingID = ref('')
const scrapeSource = ref('tmdb')
const scrapeCandidates = ref([])
const form = reactive(createEmptyForm())

const hasActors = computed(() => list.value.length > 0)

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

function sourceLabel(source) {
  if (source === 'tmdb') {
    return 'TMDB'
  }
  if (source === 'javdb') {
    return 'JavDB'
  }
  return source || '-'
}

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
}

function normalizeGenderValue(value) {
  if (value === '男' || value === 'male') {
    return 'male'
  }
  if (value === '女' || value === 'female') {
    return 'female'
  }
  if (value === '未知' || value === 'unknown') {
    return 'unknown'
  }
  return value || ''
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
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载演员列表失败'))
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, createEmptyForm())
}

function resetScrapeState() {
  scrapeSource.value = 'tmdb'
  scrapeCandidates.value = []
  scraping.value = false
}

function openCreate() {
  editingID.value = ''
  resetForm()
  resetScrapeState()
  dialogVisible.value = true
}

function openEdit(row) {
  editingID.value = row.id
  Object.assign(form, {
    name: row.name || '',
    aliases: Array.isArray(row.aliases) ? [...row.aliases] : [],
    gender: normalizeGenderValue(row.gender),
    country: row.country || '',
    birth_date: row.birth_date || '',
    avatar_url: row.avatar_url || '',
    source: row.source || 'manual',
    external_id: row.external_id || '',
    notes: row.notes || '',
    active: !!row.active
  })
  resetScrapeState()
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

async function scrapeByName() {
  const keyword = form.name?.trim() || ''
  if (!keyword) {
    ElMessage.warning('请先输入演员姓名')
    return
  }
  scraping.value = true
  try {
    const data = await scrapeAdminActorPreview({
      name: keyword,
      source: scrapeSource.value,
      limit: 10
    })
    const items = Array.isArray(data?.items) ? data.items : []
    scrapeCandidates.value = items
    if (items.length === 0) {
      ElMessage.warning('未查询到匹配演员候选')
      return
    }
    ElMessage.success(`查询到 ${items.length} 条候选，请选择一条回填`)
  } catch (error) {
    scrapeCandidates.value = []
    ElMessage.error(extractErrorMessage(error, '查询演员刮削信息失败'))
  } finally {
    scraping.value = false
  }
}

function applyCandidate(item) {
  const aliases = Array.isArray(item?.aliases) ? item.aliases : []
  form.name = item?.name?.trim() || form.name
  form.aliases = normalizeAliases([...(form.aliases || []), ...aliases])
  form.gender = normalizeGenderValue(item?.gender) || form.gender
  form.country = item?.country || form.country
  form.birth_date = item?.birth_date || form.birth_date
  form.avatar_url = item?.avatar_url || form.avatar_url
  form.external_id = item?.external_id || form.external_id
  form.notes = item?.notes || form.notes
  if (item?.source === 'tmdb') {
    form.source = 'scrape_tmdb'
  } else if (item?.source === 'javdb') {
    form.source = 'scrape_av'
  }
  ElMessage.success('候选信息已回填，请确认后保存')
}

async function openExistingActor(error) {
  const existing = error?.data?.existing_actor
  if (existing && existing.id) {
    openEdit(existing)
    return true
  }

  const existingID = error?.data?.existing_actor_id
  const existingName = error?.data?.existing_actor_name || form.name?.trim() || ''
  if (!existingID && !existingName) {
    return false
  }

  const result = await getAdminActors({
    page: 1,
    page_size: 20,
    q: existingName
  })
  const rows = Array.isArray(result?.items) ? result.items : []
  const target = rows.find((row) => row.id === existingID) || rows[0]
  if (!target) {
    return false
  }
  openEdit(target)
  return true
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
    const duplicateByMessage = typeof error?.message === 'string' && error.message.includes('演员名称已存在')
    const duplicateByReason = error?.data?.reason === 'duplicate_name'
    if (!editingID.value && (duplicateByReason || (error?.code === 1025 && duplicateByMessage))) {
      try {
        await ElMessageBox.confirm('同名演员已存在，是否打开现有演员进行编辑？', '演员已存在', {
          type: 'warning',
          confirmButtonText: '去编辑',
          cancelButtonText: '取消'
        })
        const opened = await openExistingActor(error)
        if (!opened) {
          ElMessage.warning('未定位到已有演员，请在列表中搜索后编辑')
        }
      } catch (_) {
        // ignore cancel action
      }
      return
    }
    ElMessage.error(extractErrorMessage(error, '保存演员失败'))
  } finally {
    saving.value = false
  }
}

async function toggleActive(row) {
  const nextActive = !row.active
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
    active: nextActive
  }

  try {
    await updateAdminActor(row.id, payload)
    await load()
    ElMessage.success(nextActive ? '演员已启用' : '演员已停用')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '切换演员状态失败'))
  }
}

function buildStatusLabel(row) {
  return row.active ? '启用' : '停用'
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page-shell actor-page">
      <PageHeader title="演员管理" subtitle="管理演员资料、头像与来源信息">
        <template #actions>
          <el-button :loading="loading" @click="load">刷新</el-button>
          <el-button type="primary" @click="openCreate">创建演员</el-button>
        </template>
      </PageHeader>

      <Toolbar>
        <template #filters>
          <el-input v-model="query.q" class="toolbar-search" placeholder="按演员姓名搜索" clearable @keyup.enter="load" />
          <el-select v-model="query.active" class="toolbar-select" clearable placeholder="状态筛选">
            <el-option label="全部状态" value="" />
            <el-option label="仅启用" value="1" />
            <el-option label="仅停用" value="0" />
          </el-select>
        </template>
        <template #actions>
          <el-button :loading="loading" @click="load">查询</el-button>
          <el-button type="primary" @click="openCreate">创建演员</el-button>
        </template>
      </Toolbar>

      <SectionCard>
        <template #title>演员列表</template>
        <template #description>支持演员资料的新增、编辑、停用与刮削回填</template>
        <EmptyState
          v-if="!hasActors"
          title="暂无演员"
          description="点击“创建演员”添加第一位演员"
        >
          <template #action>
            <el-button type="primary" @click="openCreate">创建演员</el-button>
          </template>
        </EmptyState>
        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="list" border>
              <el-table-column prop="name" label="演员姓名" min-width="160" />
              <el-table-column prop="aliases" label="别名" min-width="220">
                <template #default="{ row }">
                  {{ Array.isArray(row.aliases) && row.aliases.length > 0 ? row.aliases.join(' / ') : '暂无' }}
                </template>
              </el-table-column>
              <el-table-column prop="gender" label="性别" width="100" />
              <el-table-column prop="country" label="国家/地区" width="130" />
              <el-table-column prop="source" label="来源" width="110">
                <template #default="{ row }">{{ sourceLabel(row.source) }}</template>
              </el-table-column>
              <el-table-column prop="active" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.active ? 'success' : 'info'">{{ buildStatusLabel(row) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="更新时间" width="180">
                <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="220">
                <template #default="{ row }">
                  <el-button size="small" @click="openEdit(row)">编辑</el-button>
                  <el-button size="small" @click="toggleActive(row)">{{ row.active ? '停用' : '启用' }}</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="action-row">
            <AdminTablePagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="total"
              @current-change="load"
            />
          </div>
        </template>
      </SectionCard>
    </div>

    <el-dialog
      v-model="dialogVisible"
      class="crud-dialog"
      :title="editingID ? '编辑演员' : '创建演员'"
      width="min(94vw, 860px)"
    >
      <div class="dialog-body">
        <el-form label-width="100px" class="dialog-form">
          <el-form-item label="演员姓名">
            <el-input v-model="form.name" placeholder="请输入演员姓名" />
          </el-form-item>
          <el-form-item label="别名">
            <el-select
              v-model="form.aliases"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="输入后回车"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="性别">
            <el-select v-model="form.gender" clearable style="width: 100%">
              <el-option label="男" value="male" />
              <el-option label="女" value="female" />
              <el-option label="未知" value="unknown" />
            </el-select>
          </el-form-item>
          <el-form-item label="国家/地区">
            <el-input v-model="form.country" placeholder="例如：日本 / 中国 / 美国" />
          </el-form-item>
          <el-form-item label="生日">
            <el-date-picker v-model="form.birth_date" type="date" value-format="YYYY-MM-DD" placeholder="选择生日" style="width: 100%" />
          </el-form-item>
          <el-form-item label="头像地址">
            <el-input v-model="form.avatar_url" placeholder="https://..." />
          </el-form-item>
          <el-form-item label="外部来源">
            <el-input v-model="form.external_id" placeholder="可选的 TMDB / JavDB ID" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="form.notes" type="textarea" :rows="3" placeholder="可选备注" />
          </el-form-item>
          <el-form-item label="启用状态">
            <el-switch v-model="form.active" active-text="启用" inactive-text="停用" />
          </el-form-item>
        </el-form>

        <SectionCard>
          <template #title>刮削预览</template>
          <template #description>可先查询候选并回填到当前表单</template>
          <div class="scrape-panel">
            <div class="scrape-panel__head">
              <el-select v-model="scrapeSource" style="width: 140px">
                <el-option label="TMDB" value="tmdb" />
                <el-option label="JavDB" value="javdb" />
              </el-select>
              <el-button :loading="scraping" @click="scrapeByName">查询候选</el-button>
            </div>
            <EmptyState
              v-if="scrapeCandidates.length === 0"
              title="暂无候选"
              description="输入演员姓名后可查询刮削候选"
            />
            <div v-else class="scrape-grid">
              <button
                v-for="item in scrapeCandidates"
                :key="`${item.source || 'unknown'}-${item.external_id || item.name}`"
                type="button"
                class="scrape-card"
                @click="applyCandidate(item)"
              >
                <strong>{{ item.name }}</strong>
                <span>{{ sourceLabel(item.source) }}</span>
              </button>
            </div>
          </div>
        </SectionCard>
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.actor-page {
  display: grid;
  gap: var(--space-4);
}

.toolbar-search {
  width: min(20rem, 100%);
}

.toolbar-select {
  width: 9.5rem;
}

.dialog-body {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(280px, 0.85fr);
  gap: var(--space-4);
}

.scrape-panel {
  display: grid;
  gap: var(--space-3);
}

.scrape-panel__head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.scrape-grid {
  display: grid;
  gap: var(--space-2);
}

.scrape-card {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  text-align: left;
  cursor: pointer;
}

.scrape-card strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.scrape-card span {
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.action-row {
  padding-top: var(--space-2);
}

@media (max-width: 64rem) {
  .dialog-body {
    grid-template-columns: 1fr;
  }
}
</style>
