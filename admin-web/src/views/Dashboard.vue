<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as echarts from 'echarts'
import { Refresh } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatCard from '../components/base/StatCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { getAdminStats } from '../api/admin'

const stats = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const chartRef = ref(null)
let chart = null

const trendPoints = computed(() => {
  const points = stats.value?.weekly_upload_trend
  return Array.isArray(points) ? points : []
})

const metricCards = computed(() => [
  { label: '短视频', value: stats.value?.short_videos || 0 },
  { label: '电影', value: stats.value?.movie_videos || 0 },
  { label: '电视剧集', value: stats.value?.episode_videos || 0 },
  { label: 'AV', value: stats.value?.av_videos || 0 },
  { label: '总用户数', value: stats.value?.total_users || 0 },
  { label: '今日上传', value: stats.value?.today_uploads || 0 },
  { label: '转码队列长度', value: stats.value?.queue_length || 0 },
  {
    label: '磁盘剩余',
    value: `${((stats.value?.disk_free_bytes || 0) / 1024 / 1024 / 1024).toFixed(2)} GB`
  }
])

function resolveColor(token) {
  if (typeof window === 'undefined' || !document?.documentElement) return ''
  const value = getComputedStyle(document.documentElement).getPropertyValue(token).trim()
  return value || ''
}

function withAlpha(color, alpha) {
  const normalized = String(color || '').trim()
  const alphaValue = Math.max(0, Math.min(1, Number(alpha)))
  if (normalized.startsWith('#')) {
    const hex = normalized.slice(1)
    const pairs =
      hex.length === 3
        ? hex.split('').map((ch) => ch + ch)
        : hex.length === 6
          ? [hex.slice(0, 2), hex.slice(2, 4), hex.slice(4, 6)]
          : null
    if (pairs) {
      const [r, g, b] = pairs.map((part) => Number.parseInt(part, 16))
      return `rgba(${r}, ${g}, ${b}, ${alphaValue})`
    }
  }
  const rgbMatch = normalized.match(/^rgb\(\s*([0-9]+)\s*,\s*([0-9]+)\s*,\s*([0-9]+)\s*\)$/i)
  if (rgbMatch) {
    const [, r, g, b] = rgbMatch
    return `rgba(${r}, ${g}, ${b}, ${alphaValue})`
  }
  return normalized
}

function renderChart() {
  if (!chartRef.value) return
  if (!trendPoints.value.length) {
    if (chart) {
      chart.clear()
    }
    return
  }

  if (!chart) {
    chart = echarts.init(chartRef.value)
    window.addEventListener('resize', handleResize)
  }

  const primary = resolveColor('--primary')
  const primarySoft = resolveColor('--primary-soft')
  const textSecondary = resolveColor('--text-secondary')
  const lineSoft = resolveColor('--line-soft')

  chart.setOption({
    grid: { left: 36, right: 18, top: 30, bottom: 24 },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: trendPoints.value.map((item) => item.day),
      axisLine: { lineStyle: { color: lineSoft } },
      axisLabel: { color: textSecondary }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: lineSoft } },
      axisLabel: { color: textSecondary }
    },
    series: [
      {
        type: 'line',
        smooth: true,
        data: trendPoints.value.map((item) => item.count),
        symbolSize: 7,
        lineStyle: { width: 3, color: primary },
        itemStyle: { color: primary },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: withAlpha(primary, 0.28) },
            { offset: 1, color: withAlpha(primarySoft, 0.22) }
          ])
        }
      }
    ]
  })
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    stats.value = await getAdminStats()
    await nextTick()
    renderChart()
  } catch (error) {
    stats.value = null
    errorMessage.value = error?.message || '加载仪表盘失败'
  } finally {
    loading.value = false
  }
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
      <PageHeader title="系统仪表盘">
        <template #actions>
          <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
        </template>
      </PageHeader>

      <EmptyState
        v-if="errorMessage"
        title="加载失败"
        :description="errorMessage"
      >
        <template #action>
          <el-button type="primary" :icon="Refresh" :loading="loading" @click="load">重试</el-button>
        </template>
      </EmptyState>

      <template v-else>
        <section class="metric-grid">
          <StatCard
            v-for="item in metricCards"
            :key="item.label"
            :label="item.label"
            :value="item.value"
          />
        </section>

        <SectionCard>
          <template #title>近 7 天上传趋势</template>
          <template #description>按天统计上传数量，辅助判断高峰与队列压力</template>
          <div v-if="trendPoints.length" ref="chartRef" class="trend-chart" />
          <EmptyState
            v-else
            title="暂无趋势数据"
            description="后端暂未返回最近 7 天上传趋势"
          />
        </SectionCard>
      </template>
    </div>
  </Layout>
</template>

<style scoped>
.dashboard-page {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-1);
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.trend-chart {
  height: 360px;
}

@media (max-width: 75rem) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 46rem) {
  .metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
