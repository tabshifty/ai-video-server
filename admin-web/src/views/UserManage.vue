<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import AdminDrawerHeader from '../components/base/AdminDrawerHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { formatAdminDateTime } from '../utils/dateTime'
import { createAdminUser, getAdminUsers, updateUserRole } from '../api/admin'
import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const loaded = ref(false)
const loadError = ref('')
const dialogVisible = ref(false)
const saving = ref(false)
const query = reactive({ page: 1, page_size: 20 })
const roleUpdatingMap = reactive({})
const form = reactive(createEmptyForm())

const initialLoading = computed(() => shouldShowCrudCollectionSkeleton({
  loading: loading.value,
  rowCount: list.value.length
}))
const hasFilters = computed(() => false)

function createEmptyForm() {
  return {
    username: '',
    email: '',
    password: '',
    role: 'user'
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

async function load() {
  loadError.value = ''
  loading.value = true
  try {
    const data = await getAdminUsers(query)
    list.value = data.items || []
    total.value = data.total_count || 0
  } catch (error) {
    loadError.value = extractErrorMessage(error, '加载用户列表失败')
  } finally {
    loaded.value = true
    loading.value = false
  }
}

function isRoleUpdating(userID) {
  return !!roleUpdatingMap[userID]
}

async function onRoleChange(row, nextRole) {
  if (!row?.id || !nextRole || row.role === nextRole) {
    return
  }

  const prevRole = row.role
  row.role = nextRole
  roleUpdatingMap[row.id] = true
  try {
    await updateUserRole(row.id, { role: nextRole })
    ElMessage.success('角色已更新')
  } catch (error) {
    row.role = prevRole
    ElMessage.error(extractErrorMessage(error, '角色更新失败'))
  } finally {
    delete roleUpdatingMap[row.id]
  }
}

function openCreateDialog() {
  Object.assign(form, createEmptyForm())
  dialogVisible.value = true
}

async function saveUser() {
  if (!form.username.trim() || !form.email.trim() || !form.password.trim()) {
    ElMessage.warning('请填写用户名、邮箱和密码')
    return
  }
  saving.value = true
  try {
    await createAdminUser({
      username: form.username.trim(),
      email: form.email.trim(),
      password: form.password,
      role: form.role
    })
    dialogVisible.value = false
    ElMessage.success('用户已创建')
    await load()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '创建用户失败'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-button :loading="loading" @click="load">刷新</el-button>
      <el-button type="primary" @click="openCreateDialog">添加用户</el-button>
    </template>

    <div class="page-shell user-page" data-density="compact">

      <Toolbar>
        <template #filters>
          <el-tag effect="plain">共 {{ total }} 个账号</el-tag>
        </template>
      </Toolbar>

      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default>
          <el-button link type="primary" @click="load">重试</el-button>
        </template>
      </el-alert>

      <el-skeleton v-if="initialLoading" :rows="8" animated />

      <SectionCard v-else-if="!loadError || list.length > 0" dense>
        <template #title>账号列表</template>
        <template #description>可直接调整用户角色，变更即时生效。</template>
        <EmptyState
          v-if="list.length === 0"
          title="暂无用户"
          description="点击“添加用户”创建第一个账号"
        >
          <template #action>
            <el-button type="primary" @click="openCreateDialog">添加用户</el-button>
          </template>
        </EmptyState>
        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" :data="list" border>
              <el-table-column prop="username" label="用户名" min-width="160" />
              <el-table-column prop="email" label="邮箱" min-width="220" />
              <el-table-column label="注册时间" width="180">
                <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="角色" width="180">
                <template #default="{ row }">
                  <el-select
                    :model-value="row.role"
                    :aria-label="`调整用户 ${row.username || row.id} 的角色`"
                    :disabled="isRoleUpdating(row.id)"
                    @change="(value) => onRoleChange(row, value)"
                  >
                    <el-option label="user" value="user" />
                    <el-option label="admin" value="admin" />
                  </el-select>
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

    <el-drawer
      v-model="dialogVisible"
      class="crud-drawer"
      title="添加用户"
      direction="rtl"
      size="min(100vw, 560px)"
      destroy-on-close
      :show-close="false"
    >
      <template #header="{ close, titleId, titleClass }">
        <AdminDrawerHeader
          title="添加用户"
          :title-id="titleId"
          :title-class="titleClass"
          :close="close"
        />
      </template>
      <el-form label-width="88px" class="dialog-form">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" autocomplete="username" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" autocomplete="email" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" autocomplete="new-password" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" style="width: 100%">
            <el-option label="user" value="user" />
            <el-option label="admin" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveUser">创建</el-button>
      </template>
    </el-drawer>
  </Layout>
</template>

<style scoped>
.user-page {
  display: grid;
  gap: var(--space-4);
}

.action-row {
  padding-top: var(--space-2);
}
</style>
