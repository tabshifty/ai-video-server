<script setup>
import { onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import Layout from '../components/Layout.vue'
import { getAdminStats } from '../api/admin'

const stats = ref({})
const chartRef = ref()

async function load() {
  stats.value = await getAdminStats()
  renderChart()
}

function renderChart() {
  if (!chartRef.value || !stats.value.weekly_upload_trend) return
  const chart = echarts.init(chartRef.value)
  const x = stats.value.weekly_upload_trend.map((x) => x.day)
  const y = stats.value.weekly_upload_trend.map((x) => x.count)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: x },
    yAxis: { type: 'value' },
    series: [{ type: 'line', smooth: true, data: y, areaStyle: {} }]
  })
}

onMounted(load)
</script>

<template>
  <Layout>
    <el-row :gutter="14">
      <el-col :span="6"><el-card>短视频: {{ stats.short_videos || 0 }}</el-card></el-col>
      <el-col :span="6"><el-card>电影: {{ stats.movie_videos || 0 }}</el-card></el-col>
      <el-col :span="6"><el-card>剧集: {{ stats.episode_videos || 0 }}</el-card></el-col>
      <el-col :span="6"><el-card>用户数: {{ stats.total_users || 0 }}</el-card></el-col>
    </el-row>
    <el-row :gutter="14" style="margin-top: 14px">
      <el-col :span="8"><el-card>今日上传: {{ stats.today_uploads || 0 }}</el-card></el-col>
      <el-col :span="8"><el-card>队列长度: {{ stats.queue_length || 0 }}</el-card></el-col>
      <el-col :span="8"><el-card>磁盘剩余: {{ ((stats.disk_free_bytes || 0) / 1024 / 1024 / 1024).toFixed(2) }} GB</el-card></el-col>
    </el-row>
    <el-card style="margin-top:14px">
      <template #header>近7天上传趋势</template>
      <div ref="chartRef" style="height: 320px" />
    </el-card>
  </Layout>
</template>
