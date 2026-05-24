<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { getSystemLogs, systemCleanup } from '../api/admin'

const logs = ref([])
const loading = ref(false)
const cleanupLoading = ref(false)

const hasLogs = computed(() => logs.value.length > 0)

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

async function loadLogs() {
  loading.value = true
  try {
    const data = await getSystemLogs({ lines: 300 })
    logs.value = data.lines || []
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '加载系统日志失败'))
  } finally {
    loading.value = false
  }
}

async function runCleanup() {
  cleanupLoading.value = true
  try {
    await systemCleanup({ older_than_hours: 24 })
    ElMessage.success('清理任务已执行')
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '执行清理任务失败'))
  } finally {
    cleanupLoading.value = false
  }
}
</script>

<template>
  <Layout>
    <div class="page-shell settings-page">
      <PageHeader title="系统设置" />

      <SectionCard>
        <template #title>控制中心</template>
        <template #description>执行清理任务并查看近期系统日志</template>
        <div class="ops-grid">
          <article class="ops-item">
            <div class="ops-name">临时文件清理</div>
            <p>清理超过 24 小时的临时文件，释放磁盘占用。</p>
            <el-button type="warning" :loading="cleanupLoading" @click="runCleanup">执行清理</el-button>
          </article>
          <article class="ops-item">
            <div class="ops-name">日志拉取</div>
            <p>拉取系统日志最近 300 行，用于排查上传与任务异常。</p>
            <el-button :loading="loading" @click="loadLogs">刷新日志</el-button>
          </article>
        </div>
      </SectionCard>

      <SectionCard>
        <template #title>系统日志</template>
        <template #description>按时间顺序展示最近日志输出</template>
        <EmptyState
          v-if="!hasLogs"
          title="暂无日志"
          description="点击右上角刷新按钮拉取最近日志"
        >
          <template #action>
            <el-button :loading="loading" @click="loadLogs">刷新</el-button>
          </template>
        </EmptyState>
        <el-scrollbar v-else max-height="60vh" class="log-box">
          <pre class="log-text">{{ logs.join('\n') }}</pre>
        </el-scrollbar>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.settings-page {
  display: grid;
  gap: var(--space-6);
}

.ops-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.ops-item {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.ops-name {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
}

.ops-item p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.log-box {
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--slate-950);
}

.log-text {
  margin: 0;
  padding: var(--space-3);
  white-space: pre-wrap;
  color: var(--slate-100);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
  font-family: var(--font-mono);
}

@media (max-width: 48rem) {
  .ops-grid {
    grid-template-columns: 1fr;
  }
}
</style>
