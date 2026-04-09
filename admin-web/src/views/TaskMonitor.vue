<script setup>
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import Layout from '../components/Layout.vue'
import { getAdminTasks } from '../api/admin'

const list = ref([])
const total = ref(0)
const query = reactive({ page: 1, page_size: 20 })
let timer = null

async function load() {
  const data = await getAdminTasks(query)
  list.value = data.items || []
  total.value = data.total_count || 0
}

onMounted(async () => {
  await load()
  timer = setInterval(load, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">任务监控</h1>
          <p class="page-subtitle">每 5 秒自动刷新转码任务状态</p>
        </div>
      </div>

    <el-card class="soft-card">
      <div class="table-wrap">
      <el-table :data="list" border>
        <el-table-column prop="id" label="任务ID" width="90" />
        <el-table-column prop="video_id" label="视频ID" min-width="220" />
        <el-table-column prop="status" label="状态" width="120" />
        <el-table-column prop="retry_count" label="重试" width="80" />
        <el-table-column prop="error" label="错误信息" min-width="280" />
        <el-table-column prop="started_at" label="开始时间" width="180" />
      </el-table>
      </div>
      <div class="action-row">
        <el-pagination v-model:current-page="query.page" :total="total" @current-change="load" />
      </div>
    </el-card>
    </div>
  </Layout>
</template>
