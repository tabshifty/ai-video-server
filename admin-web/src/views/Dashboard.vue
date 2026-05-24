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
    <div class="page-shell dashboard-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">系统概览</h1>
          <p class="page-subtitle">视频、用户、任务与存储状态一屏查看</p>
        </div>
      </div>

      <section class="page-section">
        <div class="section-head">
          <div class="section-head__main">
            <h2 class="section-head__title">核心指标</h2>
            <p class="section-head__desc">覆盖视频存量、用户规模、当日上传与转码队列状态</p>
          </div>
        </div>

        <div class="stats-grid">
          <article class="content-card metric-card">
            <div class="metric-card__label">短视频</div>
            <div class="metric-card__value">{{ stats.short_videos || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">电影</div>
            <div class="metric-card__value">{{ stats.movie_videos || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">电视剧集</div>
            <div class="metric-card__value">{{ stats.episode_videos || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">AV</div>
            <div class="metric-card__value">{{ stats.av_videos || 0 }}</div>
          </article>
          <article class="content-card metric-card metric-card--accent">
            <div class="metric-card__label">总用户数</div>
            <div class="metric-card__value">{{ stats.total_users || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">今日上传</div>
            <div class="metric-card__value">{{ stats.today_uploads || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">转码队列长度</div>
            <div class="metric-card__value">{{ stats.queue_length || 0 }}</div>
          </article>
          <article class="content-card metric-card">
            <div class="metric-card__label">磁盘剩余</div>
            <div class="metric-card__value">{{ ((stats.disk_free_bytes || 0) / 1024 / 1024 / 1024).toFixed(2) }} GB</div>
          </article>
        </div>
      </section>

      <section class="table-panel trend-panel">
        <div class="section-head">
          <div class="section-head__main">
            <h2 class="section-head__title">近 7 天上传趋势</h2>
            <p class="section-head__desc">按天统计上传数量，辅助判断高峰与队列压力</p>
          </div>
        </div>
        <div ref="chartRef" class="trend-chart" />
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.dashboard-page {
  padding-bottom: 4px;
}

.metric-card {
  display: grid;
  gap: 8px;
  min-height: 110px;
}

.metric-card__label {
  color: var(--text-muted);
  font-size: 13px;
}

.metric-card__value {
  font-family: var(--font-mono);
  font-size: 30px;
  line-height: 1.1;
  font-weight: 600;
  color: var(--text-main);
  word-break: break-all;
}

.metric-card--accent {
  background: linear-gradient(145deg, #edf3ff 0%, #ffffff 70%);
  border-color: #cad8f5;
}

.trend-panel {
  display: grid;
  gap: 10px;
}

.trend-chart {
  height: 360px;
}

@media (max-width: 1200px) {
  .metric-card__value {
    font-size: 26px;
  }
}
</style>
