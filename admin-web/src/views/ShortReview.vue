<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft,
  ArrowRight,
  Delete,
  Headset,
  Mute,
  Refresh,
  Search,
  VideoCamera
} from '@element-plus/icons-vue'
import EmptyState from '../components/base/EmptyState.vue'
import PageHeader from '../components/base/PageHeader.vue'
import {
  deleteAdminVideo,
  getAdminVideoDetail,
  getAdminVideoPlayURL,
  getAdminVideos
} from '../api/admin'
import {
  buildShortReviewQuery,
  findVideoIndexByID,
  isReviewableShortVideo,
  readStoredShortReviewPosition,
  readStoredShortReviewSound,
  resolveInitialReviewIndex,
  resolveNextReviewStep,
  writeStoredShortReviewPosition,
  writeStoredShortReviewSound
} from './shortReview.helpers'

const PAGE_SIZE = 20

const items = ref([])
const total = ref(0)
const currentIndex = ref(-1)
const currentDetail = ref(null)
const playURL = ref('')
const listLoading = ref(false)
const loadingMore = ref(false)
const detailLoading = ref(false)
const playLoading = ref(false)
const deleting = ref(false)
const reachedEnd = ref(false)
const page = ref(1)
const loadedStartPage = ref(1)
const appliedKeyword = ref('')
const muted = ref(!readStoredShortReviewSound())
const videoRef = ref(null)
const requestSeq = ref(0)

const activeItem = computed(() => {
  if (currentIndex.value < 0) return null
  return items.value[currentIndex.value] || null
})
const loadedCount = computed(() => items.value.length)
const hasItems = computed(() => loadedCount.value > 0)
const canGoPrevious = computed(() => currentIndex.value > 0)
const hasMorePages = computed(() => loadedCount.value > 0 && page.value * PAGE_SIZE < total.value)
const progressText = computed(() => {
  if (!hasItems.value || currentIndex.value < 0) return '0 / 0'
  const absoluteIndex = (loadedStartPage.value - 1) * PAGE_SIZE + currentIndex.value + 1
  return `${Math.min(absoluteIndex, total.value || absoluteIndex)} / ${total.value || loadedCount.value}`
})
const activeTitle = computed(() => currentDetail.value?.title || activeItem.value?.title || '未命名短视频')
const emptyTitle = computed(() => {
  if (appliedKeyword.value) return '没有匹配的待审短视频'
  if (reachedEnd.value) return '已到待审列表末尾'
  return '暂无待审短视频'
})
const emptyDescription = computed(() => {
  if (appliedKeyword.value) return '可以清空关键词后重新加载，或稍后刷新新上传的视频。'
  if (reachedEnd.value) return '当前筛选下没有更多 ready 短视频。可以回到第一条复查，或刷新列表拉取新内容。'
  return '当前没有 ready 状态的短视频需要清理审核。'
})

const searchForm = reactive({
  q: ''
})

function nextRequestSeq() {
  requestSeq.value += 1
  return requestSeq.value
}

function isStale(seq) {
  return requestSeq.value !== seq
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
    minute: '2-digit'
  })
}

function formatDuration(seconds) {
  const totalSeconds = Number(seconds || 0)
  if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) return '--:--'
  const minutes = Math.floor(totalSeconds / 60)
  const rest = Math.floor(totalSeconds % 60)
  return `${minutes}:${String(rest).padStart(2, '0')}`
}

function videoMetaLine(item) {
  const parts = []
  if (item?.upload_user) parts.push(item.upload_user)
  if (item?.created_at) parts.push(formatDateTime(item.created_at))
  return parts.join(' · ') || '暂无上传信息'
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

function resolveItemPage(index = currentIndex.value) {
  const normalized = Math.max(0, Number(index || 0))
  return loadedStartPage.value + Math.floor(normalized / PAGE_SIZE)
}

function updateCurrentPosition(videoID = activeItem.value?.id, index = currentIndex.value) {
  writeStoredShortReviewPosition({ videoID: videoID || '', page: resolveItemPage(index) })
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

async function selectIndex(index, { persist = true } = {}) {
  if (index < 0 || index >= items.value.length) {
    currentIndex.value = -1
    currentDetail.value = null
    resetPlayback()
    return
  }
  const item = items.value[index]
  const seq = nextRequestSeq()
  currentIndex.value = index
  reachedEnd.value = false
  currentDetail.value = null
  resetPlayback()
  if (persist) updateCurrentPosition(item.id, index)

  detailLoading.value = true
  try {
    const detail = await getAdminVideoDetail(item.id)
    if (isStale(seq) || activeItem.value?.id !== item.id) return
    if (!isReviewableShortVideo(detail)) {
      ElMessage.warning('该视频已不在待审集合中，已跳到下一条')
      if (!isStale(seq)) detailLoading.value = false
      await advanceReview({ persistCurrent: false })
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

async function restoreDetachedVideo(videoID) {
  if (!videoID) return false
  const seq = nextRequestSeq()
  detailLoading.value = true
  try {
    const detail = await getAdminVideoDetail(videoID)
    if (isStale(seq) || !isReviewableShortVideo(detail)) return false
    if (findVideoIndexByID(items.value, videoID) < 0) {
      items.value = [detail, ...items.value]
      total.value = Math.max(total.value, items.value.length)
    }
    currentIndex.value = findVideoIndexByID(items.value, videoID)
    currentDetail.value = detail
    reachedEnd.value = false
    updateCurrentPosition(videoID, currentIndex.value)
    await loadPlayURL(videoID, seq)
    return true
  } catch (_) {
    return false
  } finally {
    if (!isStale(seq)) detailLoading.value = false
  }
}

async function loadPage(targetPage, { append = false, restoreVideoID = '', allowDetachedRestore = false } = {}) {
  const isAppend = append
  if (isAppend) {
    loadingMore.value = true
  } else {
    listLoading.value = true
    reachedEnd.value = false
  }
  try {
    const data = await getAdminVideos(buildShortReviewQuery({
      page: targetPage,
      pageSize: PAGE_SIZE,
      keyword: appliedKeyword.value
    }))
    const nextItems = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total_count || 0)
    page.value = Number(data?.page || targetPage)
    if (isAppend) {
      if (items.value.length === 0) {
        loadedStartPage.value = page.value
      }
      const known = new Set(items.value.map((item) => String(item.id)))
      items.value = [
        ...items.value,
        ...nextItems.filter((item) => !known.has(String(item.id)))
      ]
    } else {
      loadedStartPage.value = page.value
      items.value = nextItems
    }

    if (!isAppend) {
      const index = resolveInitialReviewIndex(items.value, restoreVideoID)
      if (restoreVideoID && findVideoIndexByID(items.value, restoreVideoID) < 0 && allowDetachedRestore) {
        if (await restoreDetachedVideo(restoreVideoID)) {
          return true
        }
        currentIndex.value = -1
        currentDetail.value = null
        resetPlayback()
        return false
      }
      if (index >= 0) {
        await selectIndex(index, { persist: true })
        return true
      } else {
        currentIndex.value = -1
        currentDetail.value = null
        resetPlayback()
        return false
      }
    }
    return true
  } catch (error) {
    ElMessage.error(error?.message || '待审短视频加载失败')
    return false
  } finally {
    listLoading.value = false
    loadingMore.value = false
  }
}

async function loadInitial({ restorePosition = true } = {}) {
  const savedPosition = restorePosition ? readStoredShortReviewPosition() : { videoID: '', page: 1 }
  page.value = restorePosition ? savedPosition.page || 1 : 1
  const restored = await loadPage(page.value, {
    restoreVideoID: savedPosition.videoID,
    allowDetachedRestore: restorePosition
  })
  if (!restored && restorePosition && savedPosition.videoID) {
    page.value = 1
    await loadPage(1)
  }
}

async function loadNextPageAndSelect() {
  if (!hasMorePages.value || loadingMore.value) {
    reachedEnd.value = true
    return
  }
  const beforeCount = items.value.length
  await loadPage(page.value + 1, { append: true })
  if (items.value.length > beforeCount) {
    await selectIndex(beforeCount)
  } else {
    reachedEnd.value = true
  }
}

async function loadFollowingPageAfterLocalEmpty() {
  const currentPage = page.value
  await loadPage(currentPage, { append: false })
  if (items.value.length > 0) {
    return
  }
  if (currentPage > 1) {
    page.value = currentPage - 1
    await loadPage(page.value, { append: false })
    if (items.value.length > 0) {
      return
    }
  }
  reachedEnd.value = true
}

async function advanceReview({ persistCurrent = true } = {}) {
  if (persistCurrent) updateCurrentPosition(activeItem.value?.id)
  const step = resolveNextReviewStep({
    currentIndex: currentIndex.value,
    loadedCount: loadedCount.value,
    totalCount: total.value,
    loadingMore: loadingMore.value
  })
  if (step.type === 'select') {
    await selectIndex(step.index)
    return
  }
  if (step.type === 'load-more') {
    await loadNextPageAndSelect()
    return
  }
  reachedEnd.value = true
}

async function goPrevious() {
  if (!canGoPrevious.value) return
  await selectIndex(currentIndex.value - 1)
}

async function applySearch() {
  appliedKeyword.value = String(searchForm.q || '').trim()
  await loadInitial({ restorePosition: false })
}

async function resetSearch() {
  searchForm.q = ''
  appliedKeyword.value = ''
  await loadInitial({ restorePosition: false })
}

async function refreshList() {
  await loadInitial()
}

async function restartFromFirst() {
  if (items.value.length === 0) {
    await loadInitial()
    return
  }
  reachedEnd.value = false
  await selectIndex(0)
}

function toggleSound() {
  muted.value = !muted.value
  writeStoredShortReviewSound(!muted.value)
  if (videoRef.value) {
    videoRef.value.muted = muted.value
  }
}

async function deleteAndNext() {
  const item = activeItem.value
  if (!item || deleting.value) return
  await ElMessageBox.confirm(`确认删除「${item.title || '未命名短视频'}」？此操作不可恢复。`, '删除短视频', {
    type: 'warning',
    confirmButtonText: '删除并下一条',
    cancelButtonText: '取消'
  })
  deleting.value = true
  try {
    const totalBeforeDelete = total.value
    const hadMorePagesBeforeDelete = page.value * PAGE_SIZE < totalBeforeDelete
    await deleteAdminVideo(item.id)
    ElMessage.success('已删除短视频')
    const deletedIndex = currentIndex.value
    items.value.splice(deletedIndex, 1)
    total.value = Math.max(0, total.value - 1)
    currentDetail.value = null
    resetPlayback()
    if (items.value.length === 0 && total.value > 0) {
      currentIndex.value = -1
      await loadFollowingPageAfterLocalEmpty()
      return
    }
    if (deletedIndex >= items.value.length && hadMorePagesBeforeDelete) {
      await loadNextPageAndSelect()
      return
    }
    if (items.value.length === 0) {
      currentIndex.value = -1
      reachedEnd.value = true
      writeStoredShortReviewPosition({ videoID: '' })
      return
    }
    const nextIndex = Math.min(deletedIndex, items.value.length - 1)
    await selectIndex(nextIndex)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.message || '删除失败')
    }
  } finally {
    deleting.value = false
  }
}

watch(muted, (value) => {
  if (videoRef.value) {
    videoRef.value.muted = value
  }
})

onMounted(() => {
  searchForm.q = appliedKeyword.value
  loadInitial()
})

onBeforeUnmount(() => {
  resetPlayback()
  nextRequestSeq()
})
</script>

<template>
  <div class="page-shell short-review-page">
    <PageHeader
      title="短视频审核"
      subtitle="连续查看 ready 短视频，保留需要的视频，删除不需要的视频。"
    >
      <template #actions>
        <el-button :icon="Refresh" :loading="listLoading" @click="refreshList">刷新列表</el-button>
      </template>
    </PageHeader>

    <section class="short-review-toolbar" aria-label="短视频审核筛选">
      <el-input
        v-model="searchForm.q"
        class="short-review-toolbar__search"
        clearable
        placeholder="搜索短视频标题、描述或标签"
        :prefix-icon="Search"
        @keyup.enter="applySearch"
        @clear="resetSearch"
      />
      <el-button type="primary" :icon="Search" @click="applySearch">搜索</el-button>
      <el-button @click="resetSearch">重置</el-button>
      <span class="short-review-toolbar__meta">待审 {{ total }} 条 · 当前 {{ progressText }}</span>
    </section>

    <section class="review-workbench" :class="{ 'is-empty': !hasItems }">
      <aside class="review-queue" aria-label="待审短视频队列">
        <div class="review-queue__head">
          <div>
            <h2>待审队列</h2>
            <p>{{ appliedKeyword ? `关键词：${appliedKeyword}` : 'ready 短视频 · 上传时间倒序' }}</p>
          </div>
          <el-tag type="info" effect="plain">{{ loadedCount }}/{{ total }}</el-tag>
        </div>

        <div v-loading="listLoading" class="review-queue__list">
          <button
            v-for="(item, index) in items"
            :key="item.id"
            class="queue-item"
            :class="{ 'is-active': index === currentIndex }"
            type="button"
            @click="selectIndex(index)"
          >
            <span class="queue-item__thumb">
              <img
                v-if="item.thumbnail"
                :src="`/api/v1/videos/${item.id}/thumbnail`"
                :alt="`${item.title || '短视频'}封面`"
              />
              <el-icon v-else><VideoCamera /></el-icon>
            </span>
            <span class="queue-item__copy">
              <strong>{{ item.title || '未命名短视频' }}</strong>
              <em>{{ videoMetaLine(item) }}</em>
            </span>
          </button>

          <div v-if="loadingMore" class="review-queue__loading">正在加载下一页...</div>
          <EmptyState
            v-if="!listLoading && items.length === 0"
            class="review-queue__empty"
            :icon="VideoCamera"
            title="暂无队列"
            description="当前筛选下没有可审核的 ready 短视频。"
          />
        </div>
      </aside>

      <main class="review-player-panel" aria-label="短视频审核播放器">
        <EmptyState
          v-if="!hasItems || reachedEnd"
          :icon="VideoCamera"
          :title="emptyTitle"
          :description="emptyDescription"
        >
          <template #action>
            <div class="empty-actions">
              <el-button v-if="hasItems" :icon="ArrowLeft" @click="restartFromFirst">回到第一条</el-button>
              <el-button type="primary" :icon="Refresh" @click="refreshList">刷新列表</el-button>
            </div>
          </template>
        </EmptyState>

        <template v-else>
          <section class="player-stage" v-loading="detailLoading || playLoading">
            <video
              ref="videoRef"
              class="review-video"
              :src="playURL"
              :muted="muted"
              playsinline
              controls
              autoplay
              loop
              preload="metadata"
            />
            <div class="player-stage__top">
              <el-tag effect="dark">短视频</el-tag>
              <span>{{ progressText }}</span>
            </div>
          </section>

          <section class="review-detail">
            <div class="review-detail__copy">
              <h2>{{ activeTitle }}</h2>
              <p>{{ videoMetaLine(activeItem) }}</p>
              <div class="review-detail__facts">
                <span>时长 {{ formatDuration(currentDetail?.duration_seconds) }}</span>
                <span>{{ currentDetail?.width || '--' }} x {{ currentDetail?.height || '--' }}</span>
                <span>状态 ready</span>
              </div>
            </div>

            <div class="review-detail__controls">
              <el-button :icon="ArrowLeft" :disabled="!canGoPrevious" @click="goPrevious">上一条</el-button>
              <el-button :icon="muted ? Mute : Headset" @click="toggleSound">
                {{ muted ? '开启声音' : '静音' }}
              </el-button>
              <el-button
                type="primary"
                :icon="ArrowRight"
                :disabled="deleting"
                @click="advanceReview()"
              >
                保留并下一条
              </el-button>
              <el-button
                type="danger"
                :icon="Delete"
                :loading="deleting"
                @click="deleteAndNext"
              >
                删除并下一条
              </el-button>
            </div>
          </section>
        </template>
      </main>
    </section>
  </div>
</template>

<style scoped>
.short-review-page {
  min-width: 0;
}

.short-review-toolbar {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.short-review-toolbar__search {
  width: min(420px, 100%);
}

.short-review-toolbar__meta {
  margin-left: auto;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.review-workbench {
  display: grid;
  min-height: min(760px, calc(100vh - 190px));
  grid-template-columns: minmax(280px, 340px) minmax(0, 1fr);
  gap: var(--space-4);
}

.review-queue,
.review-player-panel {
  min-width: 0;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.review-queue {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  overflow: hidden;
}

.review-queue__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: 1px solid var(--line-soft);
}

.review-queue__head h2 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.review-queue__head p {
  margin: var(--space-1) 0 0;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.review-queue__list {
  min-height: 0;
  overflow: auto;
  padding: var(--space-2);
}

.queue-item {
  display: grid;
  width: 100%;
  min-width: 0;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: var(--space-3);
  align-items: center;
  padding: var(--space-2);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: inherit;
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: background var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.queue-item:hover,
.queue-item:focus-visible {
  border-color: var(--line-strong);
  background: var(--bg-surface-muted);
  outline: none;
}

.queue-item.is-active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.queue-item__thumb {
  display: grid;
  width: 72px;
  aspect-ratio: 9 / 16;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  background: var(--bg-surface-muted);
}

.queue-item__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.queue-item__copy {
  display: grid;
  min-width: 0;
  gap: var(--space-1);
}

.queue-item__copy strong {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-small);
  font-weight: 600;
  line-height: var(--leading-small);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-item__copy em {
  display: -webkit-box;
  overflow: hidden;
  color: var(--text-muted);
  font-size: var(--text-caption);
  font-style: normal;
  line-height: var(--leading-caption);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.review-queue__loading {
  padding: var(--space-3);
  color: var(--text-secondary);
  text-align: center;
  font-size: var(--text-small);
}

.review-queue__empty {
  min-height: 280px;
}

.review-player-panel {
  display: grid;
  min-height: 0;
  grid-template-rows: minmax(0, 1fr) auto;
  overflow: hidden;
}

.player-stage {
  position: relative;
  display: grid;
  min-height: 480px;
  place-items: center;
  padding: var(--space-4);
  background: #020617;
}

.review-video {
  width: min(100%, 460px);
  height: min(100%, calc(100vh - 310px));
  min-height: 420px;
  border-radius: var(--radius-md);
  background: #000000;
  object-fit: contain;
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.32);
}

.player-stage__top {
  position: absolute;
  top: var(--space-3);
  right: var(--space-3);
  left: var(--space-3);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  color: #f8fafc;
  font-size: var(--text-small);
}

.review-detail {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border-top: 1px solid var(--line-soft);
}

.review-detail__copy {
  min-width: 0;
}

.review-detail__copy h2 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
  font-weight: 600;
}

.review-detail__copy p {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.review-detail__facts {
  display: flex;
  margin-top: var(--space-2);
  gap: var(--space-2);
  flex-wrap: wrap;
}

.review-detail__facts span {
  padding: 2px var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.review-detail__controls {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.empty-actions {
  display: inline-flex;
  justify-content: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

@media (max-width: 1023px) {
  .review-workbench {
    min-height: 0;
    grid-template-columns: 1fr;
  }

  .review-queue {
    max-height: 300px;
  }

  .review-detail {
    flex-direction: column;
  }

  .review-detail__controls {
    width: 100%;
    justify-content: flex-start;
  }

  .review-video {
    height: min(70vh, 620px);
  }
}

@media (max-width: 767px) {
  .short-review-toolbar__search,
  .short-review-toolbar .el-button {
    width: 100%;
  }

  .short-review-toolbar__meta {
    width: 100%;
    margin-left: 0;
  }

  .player-stage {
    min-height: 420px;
    padding: var(--space-3);
  }

  .review-video {
    width: 100%;
    min-height: 360px;
  }

  .review-detail__controls .el-button {
    flex: 1 1 calc(50% - var(--space-2));
  }
}
</style>
