<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import Layout from '../components/Layout.vue'
import { getAdminStats } from '../api/admin'

const stats = ref({})
const chartRef = ref()
let chart = null

async function load() {
  stats.value = await getAdminStats()
  await nextTick()
  renderChart()
}

function renderChart() {
  if (!chartRef.value || !stats.value.weekly_upload_trend) return

  if (!chart) {
    chart = echarts.init(chartRef.value)
    window.addEventListener('resize', handleResize)
  }

  const x = stats.value.weekly_upload_trend.map((item) => item.day)
  const y = stats.value.weekly_upload_trend.map((item) => item.count)

  chart.setOption({
    grid: { left: 36, right: 18, top: 30, bottom: 24 },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: x,
      axisLine: { lineStyle: { color: '#fda4af' } },
      axisLabel: { color: '#6b7280' }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: 'rgba(136,19,55,0.08)' } },
      axisLabel: { color: '#6b7280' }
    },
    series: [
      {
        type: 'line',
        smooth: true,
        data: y,
        symbolSize: 7,
        lineStyle: { width: 3, color: '#e11d48' },
        itemStyle: { color: '#2563eb' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(225, 29, 72, 0.32)' },
              { offset: 1, color: 'rgba(225, 29, 72, 0.04)' }
            ]
          }
        }
      }
    ]
  })
}

function handleResize() {
  chart?.resize()
}

onMounted(load)

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
  chart = null
})
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">系统概览</h1>
          <p class="page-subtitle">视频、用户、任务与存储状态一屏查看</p>
        </div>
      </div>

      <div class="kpi-grid">
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.short_videos || 0 }}</div>
          <div class="pill-label">短视频</div>
        </el-card>
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.movie_videos || 0 }}</div>
          <div class="pill-label">电影</div>
        </el-card>
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.episode_videos || 0 }}</div>
          <div class="pill-label">电视剧集</div>
        </el-card>
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.total_users || 0 }}</div>
          <div class="pill-label">总用户数</div>
        </el-card>
      </div>

      <div class="kpi-grid kpi-grid-3">
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.today_uploads || 0 }}</div>
          <div class="pill-label">今日上传</div>
        </el-card>
        <el-card class="soft-card">
          <div class="pill-value">{{ stats.queue_length || 0 }}</div>
          <div class="pill-label">转码队列长度</div>
        </el-card>
        <el-card class="soft-card">
          <div class="pill-value">{{ ((stats.disk_free_bytes || 0) / 1024 / 1024 / 1024).toFixed(2) }} GB</div>
          <div class="pill-label">磁盘剩余</div>
        </el-card>
      </div>

      <el-card class="soft-card">
        <template #header>近 7 天上传趋势</template>
        <div ref="chartRef" class="trend-chart" />
      </el-card>
    </div>
  </Layout>
</template>

<style scoped>
.kpi-grid-3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.trend-chart {
  height: 340px;
}

@media (max-width: 1200px) {
  .kpi-grid-3 {
    grid-template-columns: 1fr;
  }
}
</style>
