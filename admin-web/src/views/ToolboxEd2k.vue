<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Back, RefreshRight } from '@element-plus/icons-vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'

const router = useRouter()
const activeStatus = ref('queued')

const task = {
  source_link: 'ed2k://|file|demo_episode_01.mkv|734003200|0123456789ABCDEF0123456789ABCDEF|/',
  resource_hash: '0123456789ABCDEF0123456789ABCDEF',
  filename: 'demo_episode_01.mkv',
  declared_size: 734003200,
  progress_text: '下载中但暂无进度',
  status_reason: '等待外部引擎同步状态',
  expected_files: [
    {
      name: 'demo_episode_01.mkv',
      path: 'demo_episode_01.mkv',
      size: 734003200
    }
  ],
  running_files: [
    {
      name: 'demo_episode_01.mkv',
      path: 'Season 01/demo_episode_01.mkv',
      size: 524288000
    }
  ],
  failed_files: [
    {
      name: 'demo_episode_01.mkv',
      path: 'Season 01/demo_episode_01.mkv',
      size: 524288000
    }
  ],
  completed_files: [
    {
      name: 'demo_episode_01.mkv',
      path: 'Season 01/demo_episode_01.mkv',
      size: 734003200
    }
  ]
}

const statusOptions = [
  { value: 'queued', label: '排队中' },
  { value: 'running', label: '下载中' },
  { value: 'failed', label: '失败' },
  { value: 'completed', label: '已完成' }
]

const statusLabelMap = {
  queued: '排队中',
  running: '下载中',
  failed: '失败',
  completed: '已完成'
}

const statusToneMap = {
  queued: 'info',
  running: 'warning',
  failed: 'danger',
  completed: 'success'
}

const statusDescriptionMap = {
  queued: '任务已进入下载队列，等待外部引擎开始传输。',
  running: '任务正在下载，详情页先看来源与预期信息，再看进度和快照。',
  failed: '任务已失败，详情页优先展示失败原因和残留快照。',
  completed: '任务已完成，详情页展示最终结果清单。'
}

const currentStatusLabel = computed(() => statusLabelMap[activeStatus.value] || activeStatus.value)
const currentStatusTone = computed(() => statusToneMap[activeStatus.value] || 'info')
const currentStatusDescription = computed(() => statusDescriptionMap[activeStatus.value] || '')
const currentFiles = computed(() => {
  if (activeStatus.value === 'running') return task.running_files
  if (activeStatus.value === 'failed') return task.failed_files
  if (activeStatus.value === 'completed') return task.completed_files
  return []
})
const showExpectedFiles = computed(() => activeStatus.value === 'queued' || activeStatus.value === 'running')
const showFileSection = computed(() => activeStatus.value === 'running' || activeStatus.value === 'failed' || activeStatus.value === 'completed')
const fileSectionTitle = computed(() => {
  if (activeStatus.value === 'running') return '进行中落盘快照'
  if (activeStatus.value === 'failed') return '失败残留快照'
  if (activeStatus.value === 'completed') return '最终结果清单'
  return '文件区'
})

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function setStatus(status) {
  activeStatus.value = status
}

function returnToToolbox() {
  router.push('/toolbox')
}
</script>

<template>
  <main class="tool-workspace">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader
        title="ED2K 下载工作台"
        subtitle="先看来源标识、预期文件信息、状态反馈和文件区，后续再接入真实下载任务。"
      >
        <template #actions>
          <el-tag :type="currentStatusTone" effect="plain">{{ currentStatusLabel }}</el-tag>
          <el-button :icon="RefreshRight" @click="setStatus('queued')">重置为排队中</el-button>
        </template>
      </PageHeader>

      <SectionCard>
        <template #title>任务状态</template>
        <template #description>当前只展示下载工作流骨架，状态切换用于校验详情区块顺序。</template>

        <div class="status-switcher" role="tablist" aria-label="ED2K 任务状态切换">
          <el-button
            v-for="option in statusOptions"
            :key="option.value"
            :type="activeStatus === option.value ? 'primary' : ''"
            @click="setStatus(option.value)"
          >
            {{ option.label }}
          </el-button>
        </div>
      </SectionCard>

      <section class="detail-stack">
        <SectionCard>
          <template #title>来源标识</template>
          <template #description>原始 ED2K 链接与资源哈希始终常驻。</template>

          <div class="source-block">
            <div class="source-block__row">
              <span class="source-block__label">原始链接</span>
              <code class="source-block__value">{{ task.source_link }}</code>
            </div>
            <div class="source-block__row">
              <span class="source-block__label">资源哈希</span>
              <code class="source-block__value">{{ task.resource_hash }}</code>
            </div>
            <div class="source-block__row">
              <span class="source-block__label">文件名</span>
              <span class="source-block__value">{{ task.filename }}</span>
            </div>
            <div class="source-block__row">
              <span class="source-block__label">声明大小</span>
              <span class="source-block__value">{{ formatFileSize(task.declared_size) }}</span>
            </div>
          </div>
        </SectionCard>

        <SectionCard v-if="showExpectedFiles">
          <template #title>预期文件信息</template>
          <template #description>即使还没有落盘快照，也先展示链接解析出的目标文件信息。</template>

          <div class="file-list" aria-label="预期文件信息">
            <article v-for="file in task.expected_files" :key="file.path" class="file-item">
              <div class="file-item__main">
                <strong>{{ file.name }}</strong>
                <span>{{ file.path }}</span>
              </div>
              <div class="file-item__meta">{{ formatFileSize(file.size) }}</div>
            </article>
          </div>
        </SectionCard>

        <SectionCard>
          <template #title>状态反馈</template>
          <template #description>状态反馈只回答“现在到哪一步了”。</template>

          <div class="feedback-panel">
            <el-tag :type="currentStatusTone" effect="plain">{{ currentStatusLabel }}</el-tag>
            <p>{{ currentStatusDescription }}</p>
            <p v-if="activeStatus === 'running'" class="feedback-panel__progress">{{ task.progress_text }}</p>
            <p v-else-if="activeStatus === 'failed'" class="feedback-panel__progress">{{ task.status_reason }}</p>
          </div>
        </SectionCard>

        <SectionCard v-if="showFileSection">
          <template #title>{{ fileSectionTitle }}</template>
          <template #description>这里展示进行中快照、失败残留快照或最终结果清单。</template>

          <div v-if="currentFiles.length > 0" class="file-list" :aria-label="fileSectionTitle">
            <article v-for="file in currentFiles" :key="file.path" class="file-item">
              <div class="file-item__main">
                <strong>{{ file.name }}</strong>
                <span>{{ file.path }}</span>
              </div>
              <div class="file-item__meta">{{ formatFileSize(file.size) }}</div>
            </article>
          </div>

          <EmptyState
            v-else
            title="暂无文件"
            description="当前状态下没有可展示的落盘文件。"
          />
        </SectionCard>
      </section>
    </div>
  </main>
</template>

<style scoped>
.tool-workspace {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.tool-workspace__inner {
  display: grid;
  width: min(100%, 72rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.status-switcher {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.detail-stack {
  display: grid;
  gap: var(--space-4);
}

.source-block {
  display: grid;
  gap: var(--space-3);
}

.source-block__row {
  display: grid;
  gap: var(--space-1);
}

.source-block__label {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.source-block__value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.file-list {
  display: grid;
  gap: var(--space-2);
}

.file-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.file-item__main {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.file-item__main strong {
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
}

.file-item__main span {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.file-item__meta {
  flex: 0 0 auto;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.feedback-panel {
  display: grid;
  gap: var(--space-2);
}

.feedback-panel p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.feedback-panel__progress {
  color: var(--text-primary);
}

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .file-item {
    flex-direction: column;
  }
}
</style>
