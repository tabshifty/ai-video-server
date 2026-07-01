<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Back, Delete, Download, Plus, RefreshRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import {
  cleanAdminEd2kDownloadTaskFiles,
  createAdminEd2kDownloadTasks,
  deleteAdminEd2kDownloadTask,
  getAdminEd2kDownloadTasks,
  retryAdminEd2kDownloadCleanup,
  retryAdminEd2kDownloadTask
} from '../api/admin'
import {
  buildEd2kTaskFocusState,
  buildPendingEd2kInput,
  filterEd2kTasks,
  mergeEd2kCreateSession,
  mergeEd2kDraftEntries,
  normalizeEd2kCreateResults,
  parseEd2kCreateEntries,
  parseEd2kLinks,
  shouldKeepEd2kCreateDialogOpen
} from './toolbox.helpers'

const router = useRouter()
const createDialogVisible = ref(false)
const ed2kInput = ref('')
const createDraftEntries = ref([])
const createSessionResults = ref([])
const currentFilter = ref('all')
const tasks = ref([])
const selectedTaskID = ref('')
const pendingFocusTaskID = ref('')
const loadingTasks = ref(false)
const submitting = ref(false)

const statusOptions = [
  { value: 'all', label: '全部' },
  { value: 'queued', label: '排队中' },
  { value: 'running', label: '下载中' },
  { value: 'canceling', label: '取消中' },
  { value: 'cancelled', label: '已取消' },
  { value: 'completed', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'files_cleaned', label: '已下载但文件已清理' }
]

const taskStatusLabelMap = {
  queued: '排队中',
  running: '下载中',
  canceling: '取消中',
  cancelled: '已取消',
  completed: '已完成',
  failed: '失败',
  files_cleaned: '已下载但文件已清理'
}

const taskStatusToneMap = {
  queued: 'info',
  running: 'warning',
  canceling: 'warning',
  cancelled: 'info',
  completed: 'success',
  failed: 'danger',
  files_cleaned: 'success'
}

const parsedLinks = computed(() => parseEd2kLinks(ed2kInput.value))
const pendingCreateEntries = computed(() => parseEd2kCreateEntries(ed2kInput.value, createDraftEntries.value))
const canSubmit = computed(() => pendingCreateEntries.value.length > 0)
const linkCount = computed(() => parsedLinks.value.links.length)
const invalidCount = computed(() => parsedLinks.value.invalidCount)
const visibleTasks = computed(() => filterEd2kTasks(tasks.value, currentFilter.value))
const selectedTask = computed(() => visibleTasks.value.find((task) => task.id === selectedTaskID.value) || visibleTasks.value[0] || null)
const selectedTaskLabel = computed(() => selectedTask.value ? taskStatusLabelMap[selectedTask.value.status] || selectedTask.value.status : '暂无任务')
const selectedTaskTone = computed(() => selectedTask.value ? taskStatusToneMap[selectedTask.value.status] || 'info' : 'info')
const selectedTaskFiles = computed(() => selectedTask.value?.files || [])
const selectedTaskHasFiles = computed(() => selectedTaskFiles.value.length > 0)
const selectedTaskHistory = computed(() => selectedTask.value?.history || [])
const hasHistoryHit = computed(() => selectedTask.value?.history?.some((item) => item.kind === 'history') || false)
const selectedTaskProgressText = computed(() => selectedTask.value?.progressText || selectedTask.value?.progress_text || '等待执行器接管')
const selectedTaskErrorMessage = computed(() => selectedTask.value?.errorMessage || selectedTask.value?.error_message || '')
const canDeleteSelectedTask = computed(() => ['queued', 'running', 'canceling', 'failed'].includes(selectedTask.value?.status || ''))
const canCleanSelectedTaskFiles = computed(() => selectedTask.value?.status === 'completed')
const canRetrySelectedTask = computed(() => selectedTask.value?.status === 'files_cleaned')
const canRetrySelectedCleanup = computed(() => selectedTask.value?.status === 'cancelled' && Boolean(selectedTaskErrorMessage.value) && !selectedTask.value?.cleanedAt)
const selectedTaskFinishedAtLabel = computed(() => {
  const status = selectedTask.value?.status || ''
  if (status === 'completed' || status === 'files_cleaned') return '下载完成时间'
  if (status === 'failed') return '失败时间'
  if (status === 'cancelled') return '取消完成时间'
  return '结束时间'
})
const selectedTaskCleanedAtLabel = computed(() => {
  const status = selectedTask.value?.status || ''
  if (status === 'files_cleaned') return '文件清理时间'
  if (status === 'cancelled') return '残留清理完成时间'
  return '清理完成时间'
})
const selectedTaskDeleteActionLabel = computed(() => buildDeleteTaskActionCopy(selectedTask.value).actionLabel)

onMounted(() => {
  void loadTasks()
})

watch(
  visibleTasks,
  (value) => {
    if (!value.length) {
      if (!pendingFocusTaskID.value) {
        selectedTaskID.value = ''
      }
      return
    }
    if (pendingFocusTaskID.value && value.some((task) => task.id === pendingFocusTaskID.value)) {
      selectedTaskID.value = pendingFocusTaskID.value
      pendingFocusTaskID.value = ''
      return
    }
    if (!value.some((task) => task.id === selectedTaskID.value)) {
      selectedTaskID.value = value[0].id
    }
    pendingFocusTaskID.value = ''
  },
  { immediate: true }
)

function isActiveStatus(status) {
  return status === 'queued' || status === 'running' || status === 'canceling'
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
    cleanedAt: task.cleanedAt || task.cleaned_at || null,
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
      if (!pendingFocusTaskID.value) {
        selectedTaskID.value = ''
      }
      return
    }
    if (pendingFocusTaskID.value && tasks.value.some((task) => task.id === pendingFocusTaskID.value)) {
      selectedTaskID.value = pendingFocusTaskID.value
      pendingFocusTaskID.value = ''
      return
    }
    if (!tasks.value.some((task) => task.id === selectedTaskID.value)) {
      selectedTaskID.value = tasks.value[0].id
    }
    pendingFocusTaskID.value = ''
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '加载下载任务失败')
  } finally {
    loadingTasks.value = false
  }
}

function selectTask(task) {
  selectedTaskID.value = task.id
}

function applyTaskSnapshot(task) {
  if (!task?.id) {
    return
  }
  const nextTask = normalizeTask(task)
  const index = tasks.value.findIndex((item) => item.id === nextTask.id)
  if (index >= 0) {
    tasks.value = tasks.value.map((item, currentIndex) => currentIndex === index ? nextTask : item)
    return
  }
  tasks.value = [nextTask, ...tasks.value]
}

async function focusTask(task) {
  const { filter, selectedTaskID: nextSelectedTaskID } = buildEd2kTaskFocusState(currentFilter.value, task)
  if (!nextSelectedTaskID) {
    return
  }
  pendingFocusTaskID.value = nextSelectedTaskID
  selectedTaskID.value = nextSelectedTaskID
  if (filter !== currentFilter.value) {
    currentFilter.value = filter
    await loadTasks()
    return
  }
  if (!tasks.value.some((item) => item.id === nextSelectedTaskID)) {
    await loadTasks()
  }
}

async function submitLinks() {
  const entries = pendingCreateEntries.value
  if (!entries.length) {
    ElMessage.warning('请先粘贴至少一条 ED2K 链接')
    return
  }

  submitting.value = true
  try {
    const data = await createAdminEd2kDownloadTasks({
      entries: entries.map((entry) => ({
        line_number: entry.lineNumber,
        source_link: entry.sourceLink
      }))
    })
    const latestResults = normalizeEd2kCreateResults(data.results || []).map((item) => ({
      ...item,
      task: item.task ? normalizeTask(item.task) : null
    }))
    createSessionResults.value = mergeEd2kCreateSession(createSessionResults.value, latestResults)
    createDraftEntries.value = mergeEd2kDraftEntries(createDraftEntries.value, entries, createSessionResults.value)
    ed2kInput.value = buildPendingEd2kInput(createDraftEntries.value, createSessionResults.value)
    await loadTasks()
    const targetTask = createSessionResults.value.find((item) => item.task?.id && ['created', 'reused', 'enqueue_failed'].includes(item.status))?.task || null
    if (targetTask?.id) {
      await focusTask(targetTask)
    }
    const createdCount = latestResults.filter((item) => item.status === 'created').length
    const reusedCount = latestResults.filter((item) => item.status === 'reused').length
    const failedCount = latestResults.filter((item) => ['invalid', 'create_failed', 'enqueue_failed'].includes(item.status)).length
    if (createdCount > 0) {
      ElMessage.success(`已创建 ${createdCount} 条下载任务`)
    }
    if (reusedCount > 0) {
      ElMessage.info(`${reusedCount} 条链接命中历史任务，已直接定位到现有记录`)
    }
    if (failedCount > 0) {
      ElMessage.warning(`${failedCount} 条输入仍待修正，已保留在弹窗中`)
    }
    if (!shouldKeepEd2kCreateDialogOpen(createDraftEntries.value, createSessionResults.value)) {
      createDialogVisible.value = false
      createDraftEntries.value = []
      createSessionResults.value = []
      ed2kInput.value = ''
    }
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '创建下载任务失败')
  } finally {
    submitting.value = false
  }
}

function setFilter(status) {
  currentFilter.value = status
  void loadTasks()
}

async function deleteTask(task) {
  const actionCopy = buildDeleteTaskActionCopy(task)
  try {
    await ElMessageBox.confirm(actionCopy.confirmMessage(task), actionCopy.dialogTitle, {
      confirmButtonText: actionCopy.confirmButtonText,
      cancelButtonText: '取消',
      type: 'warning'
    })
    const result = await deleteAdminEd2kDownloadTask(task.id)
    if (result?.deleted) {
      await loadTasks()
      ElMessage.success('任务已删除')
      return
    }
    const updatedTask = normalizeTask(result || {})
    applyTaskSnapshot(updatedTask)
    await focusTask(updatedTask)
    if (task.status === 'queued' && updatedTask.status === 'files_cleaned') {
      ElMessage.success('已撤销本次重新下载，历史任务已恢复')
      return
    }
    if (updatedTask.status === 'cancelled' && updatedTask.errorMessage) {
      ElMessage.warning(updatedTask.progressText || '任务已取消，仍有残留待清理')
      return
    }
    if (updatedTask.status === 'cancelled') {
      ElMessage.success('任务已取消')
      return
    }
    if (updatedTask.status === 'canceling') {
      ElMessage.warning(updatedTask.progressText || '取消失败，请重试')
      return
    }
    ElMessage.success(actionCopy.successMessage)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.response?.data?.msg || error?.message || '删除任务失败')
    }
  }
}

async function cleanTaskFiles(task) {
  try {
    await ElMessageBox.confirm(`确认删除「${task.title}」的暂存文件？这不会改变已下载判定。`, '删除暂存文件', {
      confirmButtonText: '删除暂存文件',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const item = normalizeTask(await cleanAdminEd2kDownloadTaskFiles(task.id))
    applyTaskSnapshot(item)
    await focusTask(item)
    ElMessage.success('暂存文件已清理')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.response?.data?.msg || error?.message || '删除暂存文件失败')
    }
  }
}

async function retryTask(task) {
  try {
    const item = normalizeTask(await retryAdminEd2kDownloadTask(task.id))
    applyTaskSnapshot(item)
    await focusTask(item)
    ElMessage.success('任务已重新加入下载队列')
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '重新下载失败')
  }
}

async function retryCleanup(task) {
  try {
    const item = normalizeTask(await retryAdminEd2kDownloadCleanup(task.id))
    applyTaskSnapshot(item)
    await focusTask(item)
    ElMessage.success('残留清理已重试完成')
  } catch (error) {
    ElMessage.error(error?.response?.data?.msg || error?.message || '重试清理失败')
  }
}

function clearComposer() {
  ed2kInput.value = ''
  createDraftEntries.value = []
  createSessionResults.value = []
}

function openCreateDialog() {
  clearComposer()
  createDialogVisible.value = true
}

function returnToToolbox() {
  router.push('/toolbox')
}

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

async function handleCreateDialogBeforeClose(done) {
  if (!ed2kInput.value.trim() && createSessionResults.value.length === 0) {
    done()
    return
  }
  try {
    await ElMessageBox.confirm('未保存的修改将会丢失，确认关闭？', '关闭新建任务', {
      confirmButtonText: '丢弃并关闭',
      cancelButtonText: '继续编辑',
      type: 'warning'
    })
    clearComposer()
    done()
  } catch {
    // keep dialog open
  }
}

function buildDeleteTaskActionCopy(task) {
  const status = task?.status || ''
  if (status === 'running') {
    return {
      actionLabel: '取消任务',
      dialogTitle: '取消下载任务',
      confirmButtonText: '确认取消',
      confirmMessage: (currentTask) => `确认取消「${currentTask.title}」？这会停止下载并清理已落下的临时文件，任务历史会保留为已取消。`,
      successMessage: '任务已取消'
    }
  }
  if (status === 'canceling') {
    return {
      actionLabel: '继续取消',
      dialogTitle: '继续取消下载任务',
      confirmButtonText: '继续取消',
      confirmMessage: (currentTask) => `确认继续取消「${currentTask.title}」？系统会再次尝试停止下载并清理残留。`,
      successMessage: '已重新尝试取消任务'
    }
  }
  if (status === 'failed') {
    return {
      actionLabel: '删除任务',
      dialogTitle: '删除下载任务',
      confirmButtonText: '永久删除',
      confirmMessage: (currentTask) => `确认删除「${currentTask.title}」？失败任务会同时清理数据库记录和已落下的下载残留。`,
      successMessage: '任务已删除'
    }
  }
  return {
    actionLabel: '删除任务',
    dialogTitle: '删除下载任务',
    confirmButtonText: '永久删除',
    confirmMessage: (currentTask) => `确认删除「${currentTask.title}」？这会永久删除当前下载任务。`,
    successMessage: '任务已删除'
  }
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
            <el-button v-if="canDeleteSelectedTask" :icon="Delete" type="danger" plain @click="deleteTask(selectedTask)">{{ selectedTaskDeleteActionLabel }}</el-button>
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
              <template #title>预期文件信息</template>
              <template #description>链接声明信息始终保留，用来核对这条任务原本打算下载什么。</template>
              <div class="source-block">
                <div class="source-block__row">
                  <span class="source-block__label">预期文件名</span>
                  <span class="source-block__value">{{ selectedTask.filename }}</span>
                </div>
                <div class="source-block__row">
                  <span class="source-block__label">声明大小</span>
                  <span class="source-block__value">{{ formatFileSize(selectedTask.declaredSize) }}</span>
                </div>
              </div>
            </SectionCard>

            <SectionCard>
              <template #title>状态反馈</template>
              <template #description>状态提示只表达当前走到哪一步。</template>
              <div class="feedback-panel">
                <el-tag :type="selectedTaskTone" effect="plain">{{ selectedTaskLabel }}</el-tag>
                <p>{{ selectedTaskProgressText }}</p>
                <p v-if="selectedTaskErrorMessage">{{ selectedTaskErrorMessage }}</p>
                <p v-if="selectedTask.finishedAt && !['queued', 'running', 'canceling'].includes(selectedTask.status)">{{ selectedTaskFinishedAtLabel }}：{{ formatDateTime(selectedTask.finishedAt) }}</p>
                <p v-if="selectedTask.cleanedAt && ['files_cleaned', 'cancelled'].includes(selectedTask.status)">{{ selectedTaskCleanedAtLabel }}：{{ formatDateTime(selectedTask.cleanedAt) }}</p>
              </div>
            </SectionCard>

            <SectionCard>
              <template #title>任务操作</template>
              <template #description>这里保留后端管理动作：删除排队/失败任务、取消运行中任务、继续重试取消中的任务，已完成任务可清理暂存文件，已清理历史可重新下载。</template>

              <div class="action-row">
                <el-button :disabled="!canDeleteSelectedTask" @click="deleteTask(selectedTask)">{{ selectedTaskDeleteActionLabel }}</el-button>
                <el-button :disabled="!canCleanSelectedTaskFiles" @click="cleanTaskFiles(selectedTask)">删除暂存文件</el-button>
                <el-button :disabled="!canRetrySelectedTask" @click="retryTask(selectedTask)">重新下载</el-button>
                <el-button :disabled="!canRetrySelectedCleanup" @click="retryCleanup(selectedTask)">重试清理</el-button>
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

  <el-dialog
    v-model="createDialogVisible"
    class="crud-dialog"
    title="新建下载任务"
    width="min(94vw, 720px)"
    destroy-on-close
    :before-close="handleCreateDialogBeforeClose"
  >
    <div class="composer">
      <p class="composer__intro">允许管理员直接贴入原始 ED2K 链接；任务标题默认取链接里的文件名。</p>
      <el-input
        v-model="ed2kInput"
        type="textarea"
        :rows="8"
        resize="vertical"
        placeholder="每行一个 ed2k:// 链接"
      />
      <div class="composer__bar">
        <el-button type="primary" :icon="Download" :loading="submitting" :disabled="!canSubmit" @click="submitLinks">创建任务</el-button>
      </div>
      <div class="composer__meta">
        <span>有效链接：{{ linkCount }}</span>
        <span v-if="invalidCount > 0">检测到 {{ invalidCount }} 行非 ED2K 文件链接，提交后会在结果区逐行回显</span>
        <span v-if="pendingCreateEntries.length > 0">任务标题会直接使用链接里的文件名</span>
      </div>
      <SectionCard v-if="createSessionResults.length > 0">
        <template #title>结果区</template>
        <template #description>逐行回显本次打开期间的提交结果，已收口输入会自动从文本框移除。</template>
        <div class="history-list" aria-label="ED2K 创建结果区">
          <article v-for="item in createSessionResults" :key="`${item.lineNumber}:${item.sourceLink}`" class="history-item">
            <strong>第 {{ item.lineNumber }} 行 · {{ item.message }}</strong>
            <span>{{ item.sourceLink }}</span>
            <div class="action-row">
              <el-tag size="small" :type="item.status === 'created' ? 'success' : item.status === 'reused' ? 'info' : item.status === 'duplicate' ? 'warning' : 'danger'" effect="plain">
                {{ item.status }}
              </el-tag>
              <el-button v-if="item.task?.id" text type="primary" @click="focusTask(item.task)">定位任务</el-button>
            </div>
          </article>
        </div>
      </SectionCard>
    </div>
    <template #footer>
      <div class="composer__footer">
        <el-button :disabled="!ed2kInput && createSessionResults.length === 0" @click="clearComposer">清空</el-button>
        <el-button @click="handleCreateDialogBeforeClose(() => { createDialogVisible = false })">取消</el-button>
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
