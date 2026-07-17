<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, CircleCheck, Delete, Headset, Mute, Refresh, VideoCamera } from '@element-plus/icons-vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import {
  deleteAdminVideo,
  getAdminPendingDeleteShorts,
  getAdminVideoDetail,
  getAdminVideoPlayURL,
  keepAdminPendingDeleteShort
} from '../api/admin'
import {
  formatShortVideoDuration,
  isPendingDeleteShortVideo,
  resolvePendingDeleteSelectionIndex,
  resolveTotalPendingDeletePages
} from './pendingDeleteShorts.helpers'

const PAGE_SIZE = 20

const items = ref([])
const total = ref(0)
const page = ref(1)
const currentIndex = ref(-1)
const currentDetail = ref(null)
const playURL = ref('')
const muted = ref(true)
const listLoading = ref(true)
const listError = ref('')
const detailLoading = ref(false)
const playLoading = ref(false)
const actionBusy = ref(false)
const videoRef = ref(null)
const requestSeq = ref(0)

const hasItems = computed(() => items.value.length > 0)
const initialLoading = computed(() => listLoading.value && items.value.length === 0)
const activeItem = computed(() => (currentIndex.value >= 0 ? items.value[currentIndex.value] || null : null))
const activeTitle = computed(() => currentDetail.value?.title || activeItem.value?.title || '未命名短视频')
const totalPages = computed(() => resolveTotalPendingDeletePages(total.value, PAGE_SIZE))
const pageSummaryText = computed(() => `第 ${page.value} / ${totalPages.value} 页 · 共 ${total.value} 条`)
const progressText = computed(() => {
  if (!hasItems.value || currentIndex.value < 0) return '0 / 0'
  const absoluteIndex = (page.value - 1) * PAGE_SIZE + currentIndex.value + 1
  return `${Math.min(absoluteIndex, total.value || absoluteIndex)} / ${total.value || items.value.length}`
})
const canGoPrevious = computed(() => currentIndex.value > 0 || page.value > 1)

function nextRequestSeq() {
  requestSeq.value += 1
  return requestSeq.value
}

function isStale(seq) {
  return requestSeq.value !== seq
}

function resetPlayback() {
  const player = videoRef.value
  if (player) {
    player.pause?.()
    player.removeAttribute?.('src')
    player.load?.()
  }
  playURL.value = ''
}

async function loadPlayURL(videoID, seq) {
  if (!videoID) return
  playLoading.value = true
  resetPlayback()
  try {
    const data = await getAdminVideoPlayURL(videoID)
    if (isStale(seq) || activeItem.value?.id !== videoID) return
    playURL.value = data?.signed_url || ''
    await nextTick()
    const player = videoRef.value
    if (player && playURL.value) {
      player.muted = muted.value
      player.play?.().catch(() => {
        ElMessage.warning('浏览器阻止了自动播放，请手动点击播放')
      })
    }
  } catch (error) {
    if (!isStale(seq)) {
      ElMessage.error(error?.message || '播放链接加载失败')
    }
  } finally {
    if (!isStale(seq)) playLoading.value = false
  }
}

async function selectIndex(index) {
  const normalizedIndex = resolvePendingDeleteSelectionIndex(items.value, index)
  if (normalizedIndex < 0) {
    currentIndex.value = -1
    currentDetail.value = null
    resetPlayback()
    return
  }

  const item = items.value[normalizedIndex]
  const seq = nextRequestSeq()
  currentIndex.value = normalizedIndex
  currentDetail.value = null
  resetPlayback()
  detailLoading.value = true
  try {
    const detail = await getAdminVideoDetail(item.id)
    if (isStale(seq) || activeItem.value?.id !== item.id) return
    if (!isPendingDeleteShortVideo(detail)) {
      ElMessage.warning('该短视频已不在待删除列表中')
      await loadPage(page.value, { selectIndexAfterLoad: normalizedIndex })
      return
    }
    currentDetail.value = detail
    await loadPlayURL(item.id, seq)
  } catch (error) {
    if (!isStale(seq)) {
      ElMessage.error(error?.message || '短视频详情加载失败')
    }
  } finally {
    if (!isStale(seq)) detailLoading.value = false
  }
}

async function loadPage(targetPage = page.value, { selectIndexAfterLoad = 0 } = {}) {
  listError.value = ''
  listLoading.value = true
  try {
    const data = await getAdminPendingDeleteShorts({
      page: Math.max(1, Number(targetPage) || 1),
      page_size: PAGE_SIZE
    })
    items.value = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total_count || 0)
    page.value = Number(data?.page || targetPage || 1)
    if (items.value.length === 0) {
      currentIndex.value = -1
      currentDetail.value = null
      resetPlayback()
      return false
    }
    await selectIndex(selectIndexAfterLoad)
    return true
  } catch (error) {
    listError.value = error?.response?.data?.msg || error?.message || '待删除短视频加载失败'
    ElMessage.error(listError.value)
    return false
  } finally {
    listLoading.value = false
  }
}

async function refreshList() {
  await loadPage(page.value, { selectIndexAfterLoad: Math.max(currentIndex.value, 0) })
}

async function handlePageChange(targetPage) {
  const normalizedPage = Math.min(Math.max(Number(targetPage) || 1, 1), totalPages.value)
  if (normalizedPage === page.value) return
  await loadPage(normalizedPage, { selectIndexAfterLoad: 0 })
}

async function goPrevious() {
  if (!canGoPrevious.value) return
  if (currentIndex.value > 0) {
    await selectIndex(currentIndex.value - 1)
    return
  }
  await loadPage(page.value - 1, { selectIndexAfterLoad: PAGE_SIZE - 1 })
}

function nextTargetPageAfterMutation() {
  const nextTotal = Math.max(0, total.value - 1)
  return Math.max(1, Math.min(page.value, resolveTotalPendingDeletePages(nextTotal, PAGE_SIZE)))
}

async function reloadAfterMutation(removedIndex) {
  const targetPage = nextTargetPageAfterMutation()
  await loadPage(targetPage, { selectIndexAfterLoad: removedIndex })
}

async function keepCurrent() {
  const item = activeItem.value
  if (!item || actionBusy.value) return
  actionBusy.value = true
  try {
    const removedIndex = currentIndex.value
    await keepAdminPendingDeleteShort(item.id)
    ElMessage.success('已保留短视频')
    await reloadAfterMutation(removedIndex)
  } catch (error) {
    ElMessage.error(error?.message || '保留失败')
  } finally {
    actionBusy.value = false
  }
}

async function deleteCurrent() {
  const item = activeItem.value
  if (!item || actionBusy.value) return
  try {
    await ElMessageBox.confirm(`确认最终删除「${item.title || '未命名短视频'}」？此操作不可恢复。`, '最终删除短视频', {
      type: 'warning',
      confirmButtonText: '最终删除',
      cancelButtonText: '取消'
    })
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    ElMessage.error(error?.message || '最终删除确认失败')
    return
  }

  actionBusy.value = true
  try {
    const removedIndex = currentIndex.value
    await deleteAdminVideo(item.id)
    ElMessage.success('已最终删除短视频')
    await reloadAfterMutation(removedIndex)
  } catch (error) {
    ElMessage.error(error?.message || '最终删除失败')
  } finally {
    actionBusy.value = false
  }
}

function toggleSound() {
  muted.value = !muted.value
  if (videoRef.value) {
    videoRef.value.muted = muted.value
  }
}

function dimensionText(video = currentDetail.value) {
  const width = Number(video?.width || 0)
  const height = Number(video?.height || 0)
  if (width <= 0 || height <= 0) return '-- x --'
  return `${width} x ${height}`
}

onMounted(() => {
  loadPage(1, { selectIndexAfterLoad: 0 })
})

onBeforeUnmount(() => {
  resetPlayback()
  nextRequestSeq()
})
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-button :icon="Refresh" :loading="listLoading" @click="refreshList">刷新列表</el-button>
    </template>

    <div class="page-shell pending-delete-page" data-density="compact">
      <p class="page-context-note">逐条复核手机端加入待删除列表的短视频。</p>

      <el-alert v-if="listError" type="error" :closable="false" :title="listError">
        <template #default><el-button link type="primary" :loading="listLoading" @click="refreshList">重试</el-button></template>
      </el-alert>

      <el-skeleton v-if="initialLoading" class="pending-delete-skeleton" :rows="12" animated />

      <section
        v-else-if="!listError || hasItems"
        class="pending-delete-workbench"
        :class="{ 'is-empty': !hasItems }"
        aria-label="待删除短视频队列"
      >
        <aside class="pending-delete-queue" aria-label="待删除短视频列表">
          <div class="pending-delete-queue__head">
            <div>
              <h2>待处理队列</h2>
              <p>按队列顺序排列</p>
            </div>
            <span>{{ pageSummaryText }}</span>
          </div>

          <div class="pending-delete-queue__body">
            <div v-loading="listLoading" class="pending-delete-queue__list">
              <button
                v-for="(item, index) in items"
                :key="item.id"
                class="pending-delete-item"
                :class="{ 'is-active': index === currentIndex }"
                :aria-current="index === currentIndex ? 'true' : undefined"
                type="button"
                @click="selectIndex(index)"
              >
                <span class="pending-delete-item__thumb">
                  <el-icon aria-hidden="true">
                    <CircleCheck v-if="index === currentIndex" />
                    <VideoCamera v-else />
                  </el-icon>
                </span>
                <span class="pending-delete-item__copy">
                  <strong>{{ item.title || '未命名短视频' }}</strong>
                </span>
              </button>

              <EmptyState
                v-if="!listLoading && items.length === 0"
                class="pending-delete-queue__empty"
                :icon="VideoCamera"
                title="暂无待删除短视频"
                description="当前队列为空。"
              />
            </div>

            <div class="pending-delete-queue__pagination">
              <AdminTablePagination
                :current-page="page"
                :page-size="PAGE_SIZE"
                :total="total"
                layout="prev, pager, next"
                :disabled="listLoading || actionBusy"
                @current-change="handlePageChange"
              />
            </div>
          </div>
        </aside>

        <main class="pending-delete-player-panel" aria-label="待删除短视频播放器">
          <EmptyState
            v-if="!hasItems"
            :icon="VideoCamera"
            title="暂无待删除短视频"
            description="处理完成后队列会停留在这里，不会跳转到其他页面。"
          >
            <template #action>
              <el-button type="primary" :icon="Refresh" :loading="listLoading" @click="refreshList">刷新列表</el-button>
            </template>
          </EmptyState>

          <template v-else>
            <section class="pending-delete-stage" v-loading="detailLoading || playLoading">
              <div class="pending-delete-video-frame">
                <video
                  ref="videoRef"
                  class="pending-delete-video"
                  :src="playURL"
                  :muted="muted"
                  playsinline
                  controls
                  autoplay
                  loop
                  preload="metadata"
                />
              </div>
              <div class="pending-delete-stage__top">
                <span>{{ progressText }}</span>
              </div>
            </section>

            <section class="pending-delete-detail">
              <div class="pending-delete-detail__copy">
                <h2>{{ activeTitle }}</h2>
                <div class="pending-delete-detail__facts">
                  <span>时长 {{ formatShortVideoDuration(currentDetail?.duration_seconds) }}</span>
                  <span>{{ dimensionText(currentDetail) }}</span>
                </div>
              </div>

              <div class="pending-delete-detail__controls">
                <el-button :icon="ArrowLeft" :disabled="!canGoPrevious || actionBusy" @click="goPrevious">上一条</el-button>
                <el-button :icon="muted ? Mute : Headset" :disabled="actionBusy" @click="toggleSound">
                  {{ muted ? '开启声音' : '静音' }}
                </el-button>
                <el-button
                  type="primary"
                  :icon="CircleCheck"
                  :loading="actionBusy"
                  @click="keepCurrent"
                >
                  保留
                </el-button>
                <el-button
                  type="danger"
                  :icon="Delete"
                  :loading="actionBusy"
                  @click="deleteCurrent"
                >
                  最终删除
                </el-button>
              </div>
            </section>
          </template>
        </main>
      </section>
    </div>
  </Layout>
</template>

<style scoped>
.pending-delete-page {
  display: flex;
  height: calc(100dvh - var(--admin-header-height) - var(--space-12));
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: var(--space-3);
  overflow: hidden;
}

.page-context-note {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.pending-delete-skeleton {
  min-height: 0;
  flex: 1;
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
}

.pending-delete-workbench {
  display: grid;
  min-height: 0;
  flex: 1;
  grid-template-columns: minmax(280px, 340px) minmax(0, 1fr);
  gap: var(--space-4);
}

.pending-delete-queue,
.pending-delete-player-panel {
  min-width: 0;
  min-height: 0;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
}

.pending-delete-queue {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  overflow: hidden;
}

.pending-delete-queue__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: 1px solid var(--line-soft);
}

.pending-delete-queue__head h2 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.pending-delete-queue__head p,
.pending-delete-queue__head span {
  margin: var(--space-1) 0 0;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.pending-delete-queue__body {
  display: grid;
  min-height: 0;
  grid-template-rows: minmax(0, 1fr) auto;
}

.pending-delete-queue__list {
  min-height: 0;
  overflow: auto;
  padding: var(--space-2);
}

.pending-delete-item {
  display: grid;
  width: 100%;
  height: var(--media-row-height);
  min-width: 0;
  grid-template-columns: 26px minmax(0, 1fr);
  gap: var(--space-2);
  align-items: center;
  padding: 3px var(--space-2);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: inherit;
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: background var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.pending-delete-item:hover,
.pending-delete-item:focus-visible {
  border-color: var(--line-strong);
  background: var(--bg-surface-muted);
}

.pending-delete-item:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: -2px;
}

.pending-delete-item.is-active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.pending-delete-item__thumb {
  display: grid;
  height: calc(var(--media-row-height) - 8px);
  aspect-ratio: 9 / 16;
  justify-self: center;
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  background: var(--bg-surface-muted);
}

.pending-delete-item__copy {
  display: flex;
  min-width: 0;
  min-height: 0;
  align-items: center;
}

.pending-delete-item__copy strong {
  display: -webkit-box;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-small);
  font-weight: 600;
  line-height: var(--leading-small);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.pending-delete-queue__empty {
  min-height: 280px;
}

.pending-delete-queue__pagination {
  padding: var(--space-3);
  border-top: 1px solid var(--line-soft);
  background: var(--bg-surface);
}

.pending-delete-queue__pagination :deep(.admin-table-pagination) {
  justify-content: flex-start;
  gap: var(--space-2);
}

.pending-delete-player-panel {
  display: grid;
  min-height: 0;
  grid-template-rows: minmax(0, 1fr) auto;
  overflow: hidden;
}

.pending-delete-stage {
  position: relative;
  display: grid;
  min-height: 0;
  place-items: center;
  padding: var(--space-4);
  background: var(--slate-950);
  overflow: hidden;
}

.pending-delete-video-frame {
  width: auto;
  max-width: min(100%, 430px);
  height: min(100%, 760px);
  max-height: calc(100% - var(--space-8));
  aspect-ratio: 9 / 16;
  padding: var(--space-2);
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  background: var(--slate-900);
}

.pending-delete-video {
  width: 100%;
  height: 100%;
  min-height: 0;
  border-radius: var(--radius-md);
  background: var(--slate-950);
  object-fit: contain;
}

.pending-delete-stage__top {
  position: absolute;
  top: var(--space-3);
  right: var(--space-3);
  left: var(--space-3);
  display: flex;
  justify-content: flex-end;
  color: var(--slate-50);
  font-size: var(--text-small);
}

.pending-delete-detail {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border-top: 1px solid var(--line-soft);
}

.pending-delete-detail__copy {
  min-width: 0;
}

.pending-delete-detail__copy h2 {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
  font-weight: 600;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.pending-delete-detail__facts {
  display: flex;
  margin-top: var(--space-2);
  gap: var(--space-2);
  flex-wrap: wrap;
}

.pending-delete-detail__facts span {
  padding: 2px var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.pending-delete-detail__controls {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

@media (max-width: 1023px) {
  .pending-delete-page {
    height: auto;
    min-height: calc(100dvh - var(--admin-header-height) - var(--space-8));
    overflow: visible;
  }

  .pending-delete-workbench {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(420px, 1fr);
    overflow: visible;
  }

  .pending-delete-queue {
    max-height: min(440px, calc(100dvh - var(--admin-header-height) - var(--space-8)));
  }

  .pending-delete-player-panel {
    min-height: 420px;
  }

  .pending-delete-detail {
    flex-direction: column;
    gap: var(--space-3);
  }

  .pending-delete-detail__controls {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 767px) {
  .pending-delete-stage {
    padding: var(--space-3);
  }

  .pending-delete-video-frame {
    max-width: 100%;
    max-height: calc(100% - var(--space-6));
    padding: var(--space-2);
  }

  .pending-delete-detail__controls .el-button {
    flex: 1 1 calc(50% - var(--space-2));
  }
}
</style>
