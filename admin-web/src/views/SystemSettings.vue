<script setup>
import { computed, ref } from 'vue'
import { RefreshRight } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { getSystemLogs, systemCleanup } from '../api/admin'

const logs = ref([])
const loading = ref(false)
const cleanupLoading = ref(false)
const logsError = ref('')
const cleanupError = ref('')

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
  logsError.value = ''
  try {
    const data = await getSystemLogs({ lines: 300 })
    logs.value = data.lines || []
  } catch (error) {
    logsError.value = extractErrorMessage(error, '加载系统日志失败')
    ElMessage.error(logsError.value)
  } finally {
    loading.value = false
  }
}

async function runCleanup() {
  cleanupLoading.value = true
  cleanupError.value = ''
  try {
    await systemCleanup({ older_than_hours: 24 })
    ElMessage.success('清理任务已执行')
  } catch (error) {
    cleanupError.value = extractErrorMessage(error, '执行清理任务失败')
    ElMessage.error(cleanupError.value)
  } finally {
    cleanupLoading.value = false
  }
}
</script>

<template>
  <Layout>
    <div class="page-shell settings-page" data-density="form">
      <p class="page-context-note">执行临时文件清理并查看最近系统日志。</p>

      <SectionCard>
        <template #title>临时文件清理</template>
        <template #description>清理超过 24 小时的临时文件，释放磁盘占用。</template>
        <template #actions>
          <el-button type="warning" :loading="cleanupLoading" @click="runCleanup">执行清理</el-button>
        </template>
        <el-alert
          v-if="cleanupError"
          class="section-feedback"
          :title="cleanupError"
          type="error"
          show-icon
          closable
          @close="cleanupError = ''"
        />
        <p class="section-note">清理仅作用于上传暂存目录，不会删除业务库中的媒体资源。</p>
      </SectionCard>

      <SectionCard>
        <template #title>系统日志</template>
        <template #description>按时间顺序展示最近日志输出</template>
        <template #actions>
          <el-button :loading="loading" @click="loadLogs">
            <el-icon><RefreshRight /></el-icon>
            <span>刷新日志</span>
          </el-button>
        </template>
        <el-alert
          v-if="logsError"
          class="section-feedback"
          :title="logsError"
          type="error"
          show-icon
          closable
          @close="logsError = ''"
        />
        <EmptyState
          v-if="!hasLogs && !logsError"
          title="暂无日志"
          description="点击刷新日志按钮拉取最近日志"
        >
          <template #action>
            <el-button :loading="loading" @click="loadLogs">刷新</el-button>
          </template>
        </EmptyState>
        <el-scrollbar v-else-if="hasLogs" max-height="60vh" class="log-box">
          <pre class="log-text">{{ logs.join('\n') }}</pre>
        </el-scrollbar>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.settings-page {
  display: grid;
  min-width: 0;
  gap: var(--space-5);
  overflow-x: clip;
}

.settings-page :deep(.section-card) {
  border-radius: var(--radius-md);
  box-shadow: none;
}

.section-feedback {
  margin-bottom: var(--space-3);
}

.section-note {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.log-box {
  overflow: hidden;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
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
</style>
