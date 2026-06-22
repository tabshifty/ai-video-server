<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, CopyDocument, Delete, Edit, Key, Plus, Refresh, Search, View } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import Toolbar from '../components/base/Toolbar.vue'
import {
  createAdminPasswordVaultEntry,
  deleteAdminPasswordVaultEntry,
  getAdminPasswordVaultEntries,
  getAdminPasswordVaultPassword,
  updateAdminPasswordVaultEntry
} from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const passwordLoading = ref(false)
const deletingId = ref('')
const list = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const formMode = ref('create')
const editingId = ref('')
const passwordDialogVisible = ref(false)
const passwordDialogTitle = ref('')
const passwordValue = ref('')
const query = reactive({
  page: 1,
  page_size: 20,
  q: ''
})
const form = reactive(createEmptyForm())

const hasEntries = computed(() => list.value.length > 0)
const dialogTitle = computed(() => (formMode.value === 'create' ? '新增密码条目' : '修改密码条目'))
const passwordPlaceholder = computed(() => (formMode.value === 'create' ? '请输入密码' : '留空则保持原密码'))

function createEmptyForm() {
  return {
    name: '',
    account: '',
    password: '',
    url: '',
    note: ''
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

function formatDateTime(value) {
  return formatAdminDateTime(value, '--')
}

function resetForm() {
  Object.assign(form, createEmptyForm())
}

function returnToToolbox() {
  router.push('/toolbox')
}

async function loadEntries() {
  loading.value = true
  try {
    const data = await getAdminPasswordVaultEntries({
      page: query.page,
      page_size: query.page_size,
      q: query.q
    })
    list.value = data.items || []
    total.value = Number(data.total_count || 0)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载密码条目失败'))
  } finally {
    loading.value = false
  }
}

function applySearch() {
  query.page = 1
  void loadEntries()
}

function clearSearch() {
  query.q = ''
  applySearch()
}

function openCreateDialog() {
  formMode.value = 'create'
  editingId.value = ''
  resetForm()
  dialogVisible.value = true
}

function openEditDialog(row) {
  formMode.value = 'edit'
  editingId.value = row.id
  Object.assign(form, {
    name: row.name || '',
    account: row.account || '',
    password: '',
    url: row.url || '',
    note: row.note || ''
  })
  dialogVisible.value = true
}

function buildPayload() {
  return {
    name: form.name.trim(),
    account: form.account.trim(),
    password: form.password,
    url: form.url.trim(),
    note: form.note.trim()
  }
}

async function saveEntry() {
  const payload = buildPayload()
  if (!payload.name) {
    ElMessage.warning('请填写名称')
    return
  }
  if (formMode.value === 'create' && !payload.password.trim()) {
    ElMessage.warning('请填写密码')
    return
  }
  saving.value = true
  try {
    if (formMode.value === 'create') {
      await createAdminPasswordVaultEntry(payload)
      ElMessage.success('密码条目已新增')
    } else {
      await updateAdminPasswordVaultEntry(editingId.value, payload)
      ElMessage.success('密码条目已更新')
    }
    dialogVisible.value = false
    await loadEntries()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存密码条目失败'))
  } finally {
    saving.value = false
  }
}

async function fetchPassword(row) {
  const data = await getAdminPasswordVaultPassword(row.id)
  return String(data?.password || '')
}

async function showPassword(row) {
  passwordLoading.value = true
  try {
    passwordValue.value = await fetchPassword(row)
    passwordDialogTitle.value = `密码：${row.name || row.account || row.id}`
    passwordDialogVisible.value = true
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '读取密码失败'))
  } finally {
    passwordLoading.value = false
  }
}

async function copyText(text, successMessage) {
  if (String(text || '') === '') {
    ElMessage.warning('没有可复制的内容')
    return
  }
  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else if (typeof document !== 'undefined') {
      const input = document.createElement('textarea')
      input.value = text
      input.setAttribute('readonly', 'readonly')
      input.style.position = 'fixed'
      input.style.opacity = '0'
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    } else {
      throw new Error('当前环境不支持复制')
    }
    ElMessage.success(successMessage)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '复制失败'))
  }
}

async function copyAccount(row) {
  await copyText(row.account || '', '账号已复制')
}

async function copyPassword(row) {
  passwordLoading.value = true
  try {
    const password = await fetchPassword(row)
    await copyText(password, '密码已复制')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '复制密码失败'))
  } finally {
    passwordLoading.value = false
  }
}

async function deleteEntry(row) {
  try {
    await ElMessageBox.confirm(`确认删除「${row.name || row.account || row.id}」？此操作不可恢复。`, '删除密码条目', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch (_) {
    return
  }

  deletingId.value = row.id
  try {
    await deleteAdminPasswordVaultEntry(row.id)
    ElMessage.success('密码条目已删除')
    await loadEntries()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '删除密码条目失败'))
  } finally {
    deletingId.value = ''
  }
}

onMounted(loadEntries)
</script>

<template>
  <main class="tool-workspace password-vault-tool">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader title="密码管理" subtitle="记录外部服务、网站和设备的账号密码。" />

      <SectionCard>
        <template #title>密码库</template>
        <template #description>支持新增、修改、删除、显示密码和复制账号密码。</template>

        <Toolbar dense>
          <template #filters>
            <el-input
              v-model="query.q"
              class="password-vault-tool__search"
              clearable
              :prefix-icon="Search"
              placeholder="搜索名称、账号、网址、备注"
              @clear="clearSearch"
              @keyup.enter="applySearch"
            />
            <el-button :icon="Search" @click="applySearch">搜索</el-button>
          </template>
          <template #actions>
            <el-button :icon="Refresh" :loading="loading" @click="loadEntries">刷新</el-button>
            <el-button type="primary" :icon="Plus" @click="openCreateDialog">新增条目</el-button>
          </template>
        </Toolbar>

        <EmptyState
          v-if="!loading && !hasEntries"
          title="暂无密码条目"
          description="点击“新增条目”记录第一个外部账号密码"
        >
          <template #action>
            <el-button type="primary" :icon="Plus" @click="openCreateDialog">新增条目</el-button>
          </template>
        </EmptyState>

        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="list" border>
              <el-table-column prop="name" label="名称" min-width="160" />
              <el-table-column prop="account" label="账号" min-width="180" show-overflow-tooltip />
              <el-table-column label="网址" min-width="220" show-overflow-tooltip>
                <template #default="{ row }">
                  <a v-if="row.url" class="password-vault-tool__url" :href="row.url" target="_blank" rel="noopener noreferrer">
                    {{ row.url }}
                  </a>
                  <span v-else>--</span>
                </template>
              </el-table-column>
              <el-table-column prop="note" label="备注" min-width="220" show-overflow-tooltip />
              <el-table-column label="更新时间" width="180">
                <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="330" fixed="right">
                <template #default="{ row }">
                  <div class="password-vault-tool__actions">
                    <el-button size="small" :icon="View" :loading="passwordLoading" @click="showPassword(row)">显示密码</el-button>
                    <el-button size="small" :icon="CopyDocument" @click="copyAccount(row)">复制账号</el-button>
                    <el-button size="small" :icon="Key" :loading="passwordLoading" @click="copyPassword(row)">复制密码</el-button>
                    <el-button size="small" :icon="Edit" @click="openEditDialog(row)">修改</el-button>
                    <el-button
                      size="small"
                      type="danger"
                      :icon="Delete"
                      :loading="deletingId === row.id"
                      @click="deleteEntry(row)"
                    >
                      删除
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="password-vault-tool__pagination">
            <AdminTablePagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              layout="total, sizes, prev, pager, next"
              :total="total"
              @current-change="loadEntries"
              @size-change="loadEntries"
            />
          </div>
        </template>
      </SectionCard>
    </div>

    <el-dialog v-model="dialogVisible" class="crud-dialog" :title="dialogTitle" width="min(94vw, 560px)">
      <el-form label-width="88px" class="dialog-form">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="例如：家庭路由器" autocomplete="off" />
        </el-form-item>
        <el-form-item label="账号">
          <el-input v-model="form.account" placeholder="请输入账号" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password :placeholder="passwordPlaceholder" autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="网址">
          <el-input v-model="form.url" placeholder="https://example.com" autocomplete="url" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.note" type="textarea" :rows="3" resize="vertical" placeholder="补充说明" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEntry">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="passwordDialogVisible" class="crud-dialog" :title="passwordDialogTitle" width="min(94vw, 520px)">
      <el-input :model-value="passwordValue" readonly autocomplete="off" />
      <template #footer>
        <el-button @click="passwordDialogVisible = false">关闭</el-button>
        <el-button type="primary" :icon="CopyDocument" @click="copyText(passwordValue, '密码已复制')">复制密码</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<style scoped>
.tool-workspace {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.tool-workspace__inner {
  display: grid;
  width: min(100%, 88rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.password-vault-tool__search {
  width: min(100%, 24rem);
}

.password-vault-tool__url {
  overflow: hidden;
  color: var(--primary);
  text-overflow: ellipsis;
  white-space: nowrap;
  text-decoration: none;
}

.password-vault-tool__url:hover {
  text-decoration: underline;
}

.password-vault-tool__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.password-vault-tool__actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.password-vault-tool__pagination {
  padding-top: var(--space-3);
}

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }
}
</style>
