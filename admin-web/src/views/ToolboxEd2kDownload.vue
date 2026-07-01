<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Delete, Download, Plus, RefreshRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import {
  createAdminEd2kDownloadTasks,
  deleteAdminEd2kDownloadTask,
  getAdminEd2kDownloadTasks,
  retryAdminEd2kDownloadTask
} from '../api/admin'
import { parseEd2kLinks } from './toolbox.helpers'

const router = useRouter()
const createDialogVisible = ref(false)
const ed2kInput = ref('')
const titleInput = ref('')
const currentFilter = ref('all')
const tasks = ref([])
const selectedTaskID = ref('')
const loadingTasks = ref(false)
const submitting = ref(false)

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
    return taskUpdatedAtValue(b) - taskUpdatedAtValue(a)
  })
  if (currentFilter.value === 'all') return list
  return list.filter((task) => String(task.status || '') === currentFilter.value)
})
const selectedTask = computed(() => visibleTasks.value.find((task) => task.id === selectedTaskID.value) || visibleTasks.value[0] || null)
const selectedTaskLabel = computed(() => selectedTask.value ? taskStatusLabelMap[selectedTask.value.status] || selectedTask.value.status : '暂无任务')
const selectedTaskTone = computed(() => selectedTask.value ? taskStatusToneMap[selectedTask.value.status] || 'info' : 'info')
const selectedTaskFiles = computed(() => selectedTask.value?.files || [])
const selectedTaskHasFiles = computed(() => selectedTaskFiles.value.length > 0)
const selectedTaskHistory = computed(() => selectedTask.value?.history || [])
const hasHistoryHit = computed(() => selectedTask.value?.history?.some((item) => item.kind === 'history') || false)
const selectedTaskProgressText = computed(() => selectedTask.value?.progressText || selectedTask.value?.progress_text || '等待执行器接管')
const selectedTaskErrorMessage = computed(() => selectedTask.value?.errorMessage || selectedTask.value?.error_message || '')

watch(
  currentFilter,
  () => {
    void loadTasks()
  },
  { immediate: true }
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

function isActiveStatus(status) {
  return status === 'queued' || status === 'running'
}

function taskUpdatedAtValue(task) {
  const raw = task?.updatedAt || task?.updated_at || task?.createdAt || task?.created_at || 0
  const time = new Date(raw).getTime()
  return Number.isFinite(time) ? time : 0
}

function formatFileSize(size) {
  const value = Number(size || 0)
  if (!Number.isFinite(value) || value <= 0) return '-'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function normalizeTask(task) {
  return {
    ...task,
    sourceLink: task.sourceLink || task.source_link || '',
    resourceHash: task.resourceHash || task.resource_hash || '',
    declaredSize: task.declaredSize || task.declared_size || 0,
    title: task.title || '',
    filename: task.filename || '',
    status: task.status || '',
    progressText: task.progressText || task.progress_text || '',
    errorMessage: task.errorMessage || task.error_message || '',
    outputDir: task.outputDir || task.output_dir || '',
    downloadedPath: task.downloadedPath || task.downloaded_path || '',
    retryCount: task.retryCount || task.retry_count || 0,
    files: Array.isArray(task.files) ? task.files : [],
    history: Array.isArray(task.history) ? task.history : [],
    createdAt: task.createdAt || task.created_at || '',
    updatedAt: task.updatedAt || task.updated_at || '',
    startedAt: task.startedAt || task.started_at || null,
    finishedAt: task.finishedAt || task.finished_at || null,
    deletedAt: task.deletedAt || task.deleted_at || null
  }
}

async function loadTasks() {
  loadingTasks.value = true
  try {
    const params = {
      page: 1,
      page_size: 100
    }
    if (currentFilter.value !== 'all') {
      params.status = currentFilter.value
    }
    const data = await getAdminEd2kDownloadTasks(params)
    tasks.value = (data.items || []).map((item) => normalizeTask(item))
    if (tasks.value.length === 0) {
      selectedTaskID.value = ''
      return
    }
    if (!tasks.value.some((task) => task.id === selectedTaskID.value)) {
      selectedTaskID.value = tasks.value[0].id
    }
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '加载下载任务失败')
  } finally {
    loadingTasks.value = false
  }
}

function selectTask(task) {
  selectedTaskID.value = task.id
}

async function submitLinks() {
  const links = parsedLinks.value.links
  if (!links.length) {
    ElMessage.warning('请先粘贴至少一条 ED2K 链接')
    return
  }

  submitting.value = true
  try {
    const data = await createAdminEd2kDownloadTasks({
      links: links.map((link) => link.href),
      title: titleInput.value.trim()
    })
    const created = (data.created || []).map((item) => normalizeTask(item))
    const reused = (data.reused || []).map((item) => normalizeTask(item))
    const rejected = data.rejected || []
    const focusTask = created[0] || reused[0] || null
    ed2kInput.value = ''
    titleInput.value = ''
    createDialogVisible.value = false
    await loadTasks()
    if (focusTask?.id) {
      selectedTaskID.value = focusTask.id
    }
    if (created.length > 0) {
      ElMessage.success(`已创建 ${created.length} 条下载任务`)
    }
    if (reused.length > 0) {
      ElMessage.info(`${reused.length} 条链接命中历史任务，已直接定位到现有记录`)
    }
    if (rejected.length > 0) {
      ElMessage.warning(`${rejected.length} 条链接未通过校验`)
    }
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '创建下载任务失败')
  } finally {
    submitting.value = false
  }
}

function setFilter(status) {
  currentFilter.value = status
}

async function deleteTask(task) {
  try {
    await ElMessageBox.confirm(`确认删除「${task.title}」？排队中的任务会被永久移除。`, '删除下载任务', {
      confirmButtonText: '永久删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteAdminEd2kDownloadTask(task.id)
    await loadTasks()
    ElMessage.success('任务已删除')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.response?.data?.msg || error?.message || '删除任务失败')
    }
  }
}

async function retryTask(task) {
  try {
    await retryAdminEd2kDownloadTask(task.id)
    await loadTasks()
    ElMessage.success('任务已重新排队')
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '重试任务失败')
  }
}

function clearComposer() {
  ed2kInput.value = ''
  titleInput.value = ''
}

function openCreateDialog() {
  clearComposer()
  createDialogVisible.value = true
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
        subtitle="管理员在这里粘贴 ED2K 链接、按资源哈希识别历史任务，并在后端工作台里管理下载任务。"
      >
        <template #actions>
          <el-button type="primary" :icon="Plus" @click="openCreateDialog">新建任务</el-button>
          <el-tag :type="selectedTaskTone" effect="plain">{{ selectedTaskLabel }}</el-tag>
          <el-button :icon="RefreshRight" :loading="loadingTasks" @click="loadTasks">刷新任务</el-button>
        </template>
      </PageHeader>

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
            <el-button :icon="RefreshRight" :disabled="selectedTask.status !== 'failed'" @click="retryTask(selectedTask)">重试任务</el-button>
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
                <p>{{ selectedTaskProgressText }}</p>
                <p v-if="selectedTaskErrorMessage">{{ selectedTaskErrorMessage }}</p>
              </div>
            </SectionCard>

            <SectionCard>
              <template #title>任务操作</template>
              <template #description>这里保留后端管理动作：重试失败任务、永久删除排队任务。</template>

              <div class="action-row">
                <el-button :disabled="selectedTask.status !== 'failed'" @click="retryTask(selectedTask)">重新排队</el-button>
                <el-button :disabled="selectedTask.status !== 'queued' && selectedTask.status !== 'running'" @click="deleteTask(selectedTask)">删除任务</el-button>
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

  <el-dialog v-model="createDialogVisible" class="crud-dialog" title="新建下载任务" width="min(94vw, 720px)" destroy-on-close>
    <div class="composer">
      <p class="composer__intro">允许管理员直接贴入原始 ED2K 链接；任务标题默认取链接里的文件名。</p>
      <el-input
        v-model="ed2kInput"
        type="textarea"
        :rows="8"
        resize="vertical"
        placeholder="每行一个 ed2k:// 链接"
        @change="syncTitleFromSingleLink"
      />
      <div class="composer__bar">
        <el-input
          v-model="titleInput"
          placeholder="任务标题（可不填，默认自动生成）"
        />
        <el-button type="primary" :icon="Download" :loading="submitting" :disabled="!canSubmit" @click="submitLinks">创建任务</el-button>
      </div>
      <div class="composer__meta">
        <span>有效链接：{{ linkCount }}</span>
        <span v-if="invalidCount > 0">已忽略 {{ invalidCount }} 行非 ED2K 文本</span>
        <span v-if="linkCount > 0">标题会优先使用文件名</span>
      </div>
    </div>
    <template #footer>
      <div class="composer__footer">
        <el-button :disabled="!ed2kInput && !titleInput" @click="clearComposer">清空</el-button>
        <el-button @click="createDialogVisible = false">取消</el-button>
      </div>
    </template>
  </el-dialog>
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

.composer__intro {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
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

.composer__footer {
  display: inline-flex;
  gap: var(--space-2);
}

.task-workspace {
  display: grid;
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
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
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
  gap: var(--space-2);
  min-width: 0;
}

.task-row__head {
  align-items: baseline;
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
  justify-content: flex-end;
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

@media (max-width: 48rem) {
  .tool-workspace__inner {
    padding: var(--space-4);
  }

  .task-row {
    grid-template-columns: 1fr;
  }

  .task-row__head,
  .task-row__meta {
    justify-content: flex-start;
  }

  .composer__bar {
    grid-template-columns: 1fr;
  }

  .file-item {
    flex-direction: column;
  }
}
</style>
