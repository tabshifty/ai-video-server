<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import { getAdminUsers, updateUserRole } from '../api/admin'

const list = ref([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20 })

async function load() {
  const data = await getAdminUsers(query)
  list.value = data.items || []
  total.value = data.total_count || 0
}

async function onRoleChange(row) {
  await updateUserRole(row.id, { role: row.role })
  ElMessage.success('角色已更新')
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

      <section class="page-section">
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
                  <el-select v-model="row.role" @change="onRoleChange(row)">
                    <el-option label="user" value="user" />
                    <el-option label="admin" value="admin" />
                  </el-select>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div class="action-row">
            <el-pagination v-model:current-page="query.page" :total="total" @current-change="load" />
          </div>
        </el-card>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.page-shell {
  gap: 16px;
}

.section-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
}

.page-section {
  display: grid;
  gap: 12px;
}

.table-panel :deep(.el-card__body) {
  display: grid;
  gap: 12px;
}

.toolbar-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

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
