<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import { getAdminUsers, updateUserRole } from '../api/admin'

const list = ref([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20 })
const roleUpdatingMap = reactive({})

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

async function load() {
  const data = await getAdminUsers(query)
  list.value = data.items || []
  total.value = data.total_count || 0
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

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page page-shell">
      <section class="section-head">
        <div>
          <h1 class="page-title">用户管理</h1>
          <p class="page-subtitle">管理用户角色与权限</p>
        </div>
      </section>

      <section>
        <el-card class="soft-card content-card table-panel">
          <div class="toolbar-row">
            <div class="toolbar-copy">
              <div class="toolbar-title">账号列表</div>
              <p>可直接调整用户角色，变更即时生效。</p>
            </div>
          </div>

          <div class="table-wrap">
            <el-table :data="list" border>
              <el-table-column prop="username" label="用户名" />
              <el-table-column prop="email" label="邮箱" />
              <el-table-column prop="created_at" label="注册时间" width="180" />
              <el-table-column label="角色" width="180">
                <template #default="{ row }">
                  <el-select
                    :model-value="row.role"
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
        </el-card>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.toolbar-copy p {
  margin: 4px 0 0;
  font-size: 12px;
  color: #6b7280;
}

.toolbar-title {
  font-size: 15px;
  font-weight: 600;
  color: #7f1d1d;
}
</style>
