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
    <el-card>
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
      <div style="margin-top:12px;display:flex;justify-content:flex-end">
        <el-pagination v-model:current-page="query.page" :total="total" @current-change="load" />
      </div>
    </el-card>
  </Layout>
</template>
