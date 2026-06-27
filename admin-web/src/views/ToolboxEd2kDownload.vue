<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Delete, Download, RefreshRight, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import { getEd2kLinkLabel, parseEd2kLinks } from './toolbox.helpers'

const STORAGE_KEY = 'admin-ed2k-download-workbench'

const router = useRouter()
const ed2kInput = ref('')
const titleInput = ref('')
const currentFilter = ref('all')
const tasks = ref([])
const selectedTaskID = ref('')

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'queued', label: '排队中' },
  { value: 'running', label: '下载中' },
  { value: 'completed', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'deleted', label: '已删除' }
]

const taskStatusLabelMap = {
  queued: '排队中',
  running: '下载中',
  completed: '已完成',
  failed: '失败',
  deleted: '已删除'
}

const taskStatusToneMap = {
  queued: 'info',
  running: 'warning',
  completed: 'success',
  failed: 'danger',
  deleted: 'info'
}

const parsedLinks = computed(() => parseEd2kLinks(ed2kInput.value))
const canSubmit = computed(() => parsedLinks.value.links.length > 0)
const linkCount = computed(() => parsedLinks.value.links.length)
const invalidCount = computed(() => parsedLinks.value.invalidCount)
const visibleTasks = computed(() => {
  const list = [...tasks.value]
  list.sort((a, b) => {
    const aActive = isActiveStatus(a.status) ? 0 : 1
    const bActive = isActiveStatus(b.status) ? 0 : 1
    if (aActive !== bActive) return aActive - bActive
    return Number(b.updatedAt || b.createdAt || 0) - Number(a.updatedAt || a.createdAt || 0)
  })
  if (currentFilter.value === 'all') return list
  return list.filter((task) => task.status === currentFilter.value)
})
const selectedTask = computed(() => visibleTasks.value.find((task) => task.id === selectedTaskID.value) || visibleTasks.value[0] || null)
const selectedTaskLabel = computed(() => selectedTask.value ? taskStatusLabelMap[selectedTask.value.status] || selectedTask.value.status : '暂无任务')
const selectedTaskTone = computed(() => selectedTask.value ? taskStatusToneMap[selectedTask.value.status] || 'info' : 'info')
const selectedTaskFiles = computed(() => selectedTask.value?.files || [])
const selectedTaskHasFiles = computed(() => selectedTaskFiles.value.length > 0)
const selectedTaskHistory = computed(() => selectedTask.value?.history || [])
const hasHistoryHit = computed(() => selectedTask.value?.history?.some((item) => item.kind === 'history') || false)

watch(
  tasks,
  (value) => persistTasks(value),
  { deep: true }
)

watch(
  visibleTasks,
  (value) => {
    if (!value.length) {
      selectedTaskID.value = ''
      return
    }
    if (!value.some((task) => task.id === selectedTaskID.value)) {
      selectedTaskID.value = value[0].id
    }
  },
  { immediate: true }
)

onMounted(() => {
  loadTasks()
})

function isActiveStatus(status) {
  return status === 'queued' || status === 'running'
}

function safeDecodeText(value) {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

function buildTaskHash(link) {
  const parts = String(link || '').trim().split('|')
  return parts[4] || parts[3] || String(link || '').trim()
}

function normalizeTaskFromLink(link, createdAt = Date.now()) {
  const href = String(link || '').trim()
  const title = getEd2kLinkLabel(href)
  const hash = buildTaskHash(href)
  const fileName = safeDecodeText(title || hash || href)
  return {
    id: `ed2k-task-${hash.toLowerCase()}-${createdAt}`,
    title: fileName,
    sourceLink: href,
    resourceHash: hash,
    filename: fileName,
    declaredSize: getEd2kDeclaredSize(href),
    status: 'queued',
    progressText: '等待外部下载引擎开始传输',
    reason: '',
    createdAt,
    updatedAt: createdAt,
    finishedAt: null,
    deletedAt: null,
    files: [],
    history: [
      {
        kind: 'created',
        label: '已创建',
        message: '任务已加入工作台'
      }
    ]
  }
}

function getEd2kDeclaredSize(link) {
  const parts = String(link || '').trim().split('|')
  const raw = Number(parts[3] || 0)
  return Number.isFinite(raw) ? raw : 0
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function persistTasks(value) {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value || []))
  } catch {
    // localStorage 不可用时只保留当前会话。
  }
}

function loadTasks() {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    tasks.value = raw ? JSON.parse(raw) : seedTasks()
  } catch {
    tasks.value = seedTasks()
  }
  if (tasks.value.length === 0) {
    tasks.value = seedTasks()
  }
  if (!selectedTaskID.value && tasks.value[0]) {
    selectedTaskID.value = tasks.value[0].id
  }
}

function seedTasks() {
  return [
    {
      id: 'ed2k-task-seed-1',
      title: 'demo_episode_01.mkv',
      sourceLink: 'ed2k://|file|demo_episode_01.mkv|734003200|0123456789ABCDEF0123456789ABCDEF|/',
      resourceHash: '0123456789ABCDEF0123456789ABCDEF',
      filename: 'demo_episode_01.mkv',
      declaredSize: 734003200,
      status: 'running',
      progressText: '下载中但暂无进度',
      reason: '',
      createdAt: Date.now() - 3600000,
      updatedAt: Date.now() - 900000,
      finishedAt: null,
      deletedAt: null,
      files: [
        {
          name: 'demo_episode_01.mkv',
          path: 'Season 01/demo_episode_01.mkv',
          size: 524288000
        }
      ],
      history: [
        { kind: 'created', label: '已创建', message: '示例任务已加入工作台' },
        { kind: 'status', label: '下载中', message: '等待外部引擎同步状态' }
      ]
    }
  ]
}

function findTaskByHash(hash) {
  const normalizedHash = String(hash || '').trim().toLowerCase()
  return tasks.value.find((task) => String(task.resourceHash || '').trim().toLowerCase() === normalizedHash) || null
}

function ensureTaskSelected(task) {
  if (!task) return
  selectedTaskID.value = task.id
}

function markTaskDeleted(task) {
  task.status = 'deleted'
  task.deletedAt = Date.now()
  task.updatedAt = task.deletedAt
  task.progressText = '任务已永久删除'
  task.reason = ''
  task.history = [
    ...(task.history || []),
    {
      kind: 'deleted',
      label: '已删除',
      message: '管理员已永久删除该任务'
    }
  ]
}

async function submitLinks() {
  const links = parsedLinks.value.links
  if (!links.length) {
    ElMessage.warning('请先粘贴至少一条 ED2K 链接')
    return
  }

  const created = []
  const reused = []
  const rejected = []

  for (const link of links) {
    const hash = buildTaskHash(link.href)
    const existing = findTaskByHash(hash)
    if (existing) {
      reused.push({ link, existing })
      existing.updatedAt = Date.now()
      existing.history = [
        ...(existing.history || []),
        {
          kind: 'history',
          label: '历史命中',
          message: `已存在任务：${existing.title}`
        }
      ]
      ensureTaskSelected(existing)
      continue
    }

    const task = normalizeTaskFromLink(link.href)
    if (titleInput.value.trim()) {
      task.title = titleInput.value.trim()
    }
    task.history = [
      ...(task.history || []),
      {
        kind: 'queued',
        label: '排队中',
        message: '任务等待下载引擎接管'
      }
    ]
    tasks.value = [task, ...tasks.value]
    created.push(task)
    ensureTaskSelected(task)
  }

  if (created.length > 0) {
    ed2kInput.value = ''
    titleInput.value = ''
    ElMessage.success(`已创建 ${created.length} 条下载任务`)
  }
  if (reused.length > 0) {
    ElMessage.info(`${reused.length} 条链接命中历史任务，已直接定位到现有记录`)
  }
  if (rejected.length > 0) {
    ElMessage.warning(`${rejected.length} 条链接未通过校验`)
  }
}

function setFilter(status) {
  currentFilter.value = status
}

function selectTask(task) {
  selectedTaskID.value = task.id
}

function moveTaskStatus(task, status) {
  task.status = status
  task.updatedAt = Date.now()
  task.progressText = status === 'running' ? '下载中但暂无进度' : status === 'completed' ? '下载完成，等待导入' : status === 'failed' ? '等待管理员处理失败原因' : '等待外部下载引擎开始传输'
  task.reason = status === 'failed' ? '外部引擎未返回可执行结果' : ''
  if (status === 'completed') {
    task.files = task.files.length > 0 ? task.files : [
      {
        name: task.filename,
        path: task.filename,
        size: task.declaredSize || 0
      }
    ]
    task.finishedAt = task.finishedAt || Date.now()
  }
  task.history = [
    ...(task.history || []),
    {
      kind: 'status',
      label: taskStatusLabelMap[status] || status,
      message: `任务状态已切换为 ${taskStatusLabelMap[status] || status}`
    }
  ]
}

function resetTask(task) {
  task.status = 'queued'
  task.progressText = '等待外部下载引擎开始传输'
  task.reason = ''
  task.updatedAt = Date.now()
  task.deletedAt = null
  task.history = [
    ...(task.history || []),
    {
      kind: 'status',
      label: '排队中',
      message: '任务已重置为排队中'
    }
  ]
}

async function deleteTask(task) {
  try {
    await ElMessageBox.confirm(`确认永久删除「${task.title}」？删除后不会进入回收站。`, '删除下载任务', {
      confirmButtonText: '永久删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    markTaskDeleted(task)
    ElMessage.success('任务已永久删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error('删除任务失败')
    }
  }
}

function clearComposer() {
  ed2kInput.value = ''
  titleInput.value = ''
}

function returnToToolbox() {
  router.push('/toolbox')
}

function syncTitleFromSingleLink() {
  if (parsedLinks.value.links.length !== 1) return
  if (titleInput.value.trim()) return
  titleInput.value = parsedLinks.value.links[0]?.label || ''
}

watch(
  () => parsedLinks.value.links.length,
  () => syncTitleFromSingleLink()
)
</script>

<template>
  <main class="tool-workspace">
    <div class="tool-workspace__inner">
      <div class="tool-workspace__topbar">
        <el-button type="primary" plain :icon="Back" @click="returnToToolbox">返回工具箱</el-button>
      </div>

      <PageHeader
        title="ED2K 下载工作台"
        subtitle="管理员在这里粘贴 ED2K 链接、按资源哈希识别历史任务，并对任务做本地管理。"
      >
        <template #actions>
          <el-tag :type="selectedTaskTone" effect="plain">{{ selectedTaskLabel }}</el-tag>
          <el-button :icon="RefreshRight" @click="loadTasks">刷新本地记录</el-button>
        </template>
      </PageHeader>

      <SectionCard>
        <template #title>提交链接</template>
        <template #description>允许管理员直接贴入原始 ED2K 链接；任务标题默认取链接里的文件名。</template>
        <template #actions>
          <el-button :disabled="!ed2kInput && !titleInput" @click="clearComposer">清空</el-button>
        </template>

        <div class="composer">
          <el-input
            v-model="ed2kInput"
            type="textarea"
            :rows="6"
            resize="vertical"
            placeholder="每行一个 ed2k:// 链接"
            @change="syncTitleFromSingleLink"
          />
          <div class="composer__bar">
            <el-input
              v-model="titleInput"
              placeholder="任务标题（可不填，默认自动生成）"
            />
            <el-button type="primary" :icon="Download" :disabled="!canSubmit" @click="submitLinks">创建任务</el-button>
          </div>
          <div class="composer__meta">
            <span>有效链接：{{ linkCount }}</span>
            <span v-if="invalidCount > 0">已忽略 {{ invalidCount }} 行非 ED2K 文本</span>
            <span v-if="linkCount > 0">标题会优先使用文件名</span>
          </div>
        </div>
      </SectionCard>

      <section class="task-workspace">
        <SectionCard class="task-list-card">
          <template #title>下载任务</template>
          <template #description>历史任务存在就直接命中，不再允许重复创建同一资源。</template>
          <template #actions>
            <div class="status-filters">
              <el-button
                v-for="option in statusOptions"
                :key="option.value"
                :type="currentFilter === option.value ? 'primary' : ''"
                @click="setFilter(option.value)"
              >
                {{ option.label }}
              </el-button>
            </div>
          </template>

          <div v-if="visibleTasks.length > 0" class="task-list" aria-label="ED2K 下载任务列表">
            <button
              v-for="task in visibleTasks"
              :key="task.id"
              class="task-row"
              :class="{ 'is-active': selectedTask && selectedTask.id === task.id }"
              type="button"
              @click="selectTask(task)"
            >
              <span class="task-row__head">
                <strong>{{ task.title }}</strong>
                <span class="task-row__sub">{{ task.resourceHash }}</span>
              </span>
              <span class="task-row__meta">
                <el-tag size="small" :type="taskStatusToneMap[task.status] || 'info'" effect="plain">
                  {{ taskStatusLabelMap[task.status] || task.status }}
                </el-tag>
                <span>{{ formatFileSize(task.declaredSize) }}</span>
              </span>
            </button>
          </div>

          <EmptyState
            v-else
            title="暂无下载任务"
            description="先在上方粘贴 ED2K 链接创建任务。"
          />
        </SectionCard>

        <SectionCard v-if="selectedTask">
          <template #title>任务详情</template>
          <template #description>来源标识常驻，文件清单只展示当前任务已有结果。</template>
          <template #actions>
            <el-button :icon="RefreshRight" @click="resetTask(selectedTask)">重置</el-button>
            <el-button :icon="Search" @click="moveTaskStatus(selectedTask, 'running')">切到下载中</el-button>
            <el-button :icon="Delete" type="danger" plain @click="deleteTask(selectedTask)">永久删除</el-button>
          </template>

          <div class="detail-stack">
            <div class="source-block">
              <div class="source-block__row">
                <span class="source-block__label">原始链接</span>
                <code class="source-block__value">{{ selectedTask.sourceLink }}</code>
              </div>
              <div class="source-block__row">
                <span class="source-block__label">资源哈希</span>
                <code class="source-block__value">{{ selectedTask.resourceHash }}</code>
              </div>
              <div class="source-block__row">
                <span class="source-block__label">文件名</span>
                <span class="source-block__value">{{ selectedTask.filename }}</span>
              </div>
              <div class="source-block__row">
                <span class="source-block__label">声明大小</span>
                <span class="source-block__value">{{ formatFileSize(selectedTask.declaredSize) }}</span>
              </div>
            </div>

            <SectionCard>
              <template #title>状态反馈</template>
              <template #description>状态提示只表达当前走到哪一步。</template>
              <div class="feedback-panel">
                <el-tag :type="selectedTaskTone" effect="plain">{{ selectedTaskLabel }}</el-tag>
                <p>{{ selectedTask.progressText }}</p>
                <p v-if="selectedTask.reason">{{ selectedTask.reason }}</p>
              </div>
            </SectionCard>

            <SectionCard>
              <template #title>任务操作</template>
              <template #description>这里保留本地管理动作，后续可再接真实后端。</template>

              <div class="action-row">
                <el-button @click="moveTaskStatus(selectedTask, 'queued')">排队</el-button>
                <el-button @click="moveTaskStatus(selectedTask, 'running')">下载中</el-button>
                <el-button @click="moveTaskStatus(selectedTask, 'completed')">已完成</el-button>
                <el-button @click="moveTaskStatus(selectedTask, 'failed')">失败</el-button>
                <el-button @click="resetTask(selectedTask)">重新排队</el-button>
              </div>
            </SectionCard>

            <SectionCard>
              <template #title>文件区</template>
              <template #description>完成后显示最终文件，失败或进行中只显示已有快照。</template>

              <div v-if="selectedTaskHasFiles" class="file-list" aria-label="任务文件区">
                <article v-for="file in selectedTaskFiles" :key="file.path" class="file-item">
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
                description="当前任务还没有可展示的落盘文件。"
              />
            </SectionCard>

            <SectionCard>
              <template #title>历史记录</template>
              <template #description>只要历史任务存在，就算已经下载过。</template>

              <div v-if="selectedTaskHistory.length > 0" class="history-list">
                <article v-for="item in selectedTaskHistory" :key="`${selectedTask.id}-${item.kind}-${item.label}`" class="history-item">
                  <strong>{{ item.label }}</strong>
                  <span>{{ item.message }}</span>
                </article>
              </div>
              <EmptyState v-else title="暂无历史" description="该任务还没有历史记录。" />
              <p v-if="hasHistoryHit" class="history-hit">当前资源已命中历史任务，不允许重建下载任务。</p>
            </SectionCard>
          </div>
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
  width: min(100%, 80rem);
  margin: 0 auto;
  padding: var(--space-6);
  gap: var(--space-5);
}

.tool-workspace__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.composer {
  display: grid;
  gap: var(--space-3);
}

.composer__bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3);
}

.composer__meta,
.history-hit {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.task-workspace {
  display: grid;
  grid-template-columns: minmax(0, 24rem) minmax(0, 1fr);
  gap: var(--space-4);
  align-items: start;
}

.task-list-card,
.detail-stack {
  min-width: 0;
}

.task-list {
  display: grid;
  gap: var(--space-2);
}

.task-row {
  display: grid;
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
  text-align: left;
}

.task-row.is-active {
  border-color: var(--primary);
  background: var(--bg-surface);
}

.task-row__head,
.task-row__meta {
  display: flex;
  justify-content: space-between;
  gap: var(--space-2);
  min-width: 0;
}

.task-row__head strong,
.task-row__sub {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-row__sub {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.task-row__meta {
  align-items: center;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.detail-stack {
  display: grid;
  gap: var(--space-4);
}

.source-block,
.history-list {
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
}

.source-block__value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: var(--text-small);
}

.feedback-panel {
  display: grid;
  gap: var(--space-2);
}

.feedback-panel p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.file-list {
  display: grid;
  gap: var(--space-2);
}

.file-item {
  display: flex;
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

.file-item__main strong,
.file-item__main span {
  overflow-wrap: anywhere;
}

.file-item__meta {
  flex: 0 0 auto;
  color: var(--text-muted);
  font-size: var(--text-small);
}

.history-item {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface-muted);
}

.history-item span {
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.status-filters {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

@media (max-width: 64rem) {
  .task-workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .composer__bar {
    grid-template-columns: 1fr;
  }

  .file-item {
    flex-direction: column;
  }
}
</style>
