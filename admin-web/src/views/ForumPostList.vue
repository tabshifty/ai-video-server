<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import AdminTablePagination from '../components/AdminTablePagination.vue'
import Layout from '../components/Layout.vue'
import EmptyState from '../components/base/EmptyState.vue'
import SectionCard from '../components/base/SectionCard.vue'
import StatusIndicator from '../components/base/StatusIndicator.vue'
import Toolbar from '../components/base/Toolbar.vue'
import { getAdminForumPosts } from '../api/admin'
import { formatAdminDateTime } from '../utils/dateTime'
import { getEd2kLinkLabel } from './toolbox.helpers'

const list = ref([])
const total = ref(0)
const loading = ref(false)
const loaded = ref(false)
const loadError = ref('')
const query = reactive({ page: 1, page_size: 20, q: '' })
const searchDraft = ref('')
let loadSequence = 0

const initialLoading = computed(() => loading.value && !loaded.value)
const hasAppliedSearch = computed(() => query.q.length > 0)

function extractErrorMessage(error, fallback) {
  const responseMessage = error?.response?.data?.msg
  if (typeof responseMessage === 'string' && responseMessage.trim()) {
    return responseMessage.trim()
  }
  if (typeof error?.message === 'string' && error.message.trim()) {
    return error.message.trim()
  }
  return fallback
}

async function load() {
  const sequence = ++loadSequence
  loading.value = true
  try {
    const data = await getAdminForumPosts({
      page: query.page,
      page_size: query.page_size,
      q: query.q
    })
    if (sequence !== loadSequence) return

    loadError.value = ''
    list.value = Array.isArray(data?.items) ? data.items : []
    total.value = Number(data?.total_count || 0)
  } catch (error) {
    if (sequence !== loadSequence) return
    loadError.value = extractErrorMessage(error, '加载论坛资源失败')
  } finally {
    if (sequence === loadSequence) {
      loaded.value = true
      loading.value = false
    }
  }
}

function setPage(page) {
  query.page = page
  list.value = []
  loaded.value = false
  loadError.value = ''
  load()
}

function applySearch() {
  query.q = searchDraft.value.trim()
  query.page = 1
  list.value = []
  loaded.value = false
  loadError.value = ''
  load()
}

function clearSearch() {
  searchDraft.value = ''
  applySearch()
}

onMounted(load)
</script>

<template>
  <Layout>
    <template #header-actions>
      <el-tooltip content="刷新论坛资源" placement="bottom">
        <el-button
          class="forum-refresh"
          circle
          :icon="Refresh"
          :loading="loading"
          aria-label="刷新论坛资源"
          title="刷新论坛资源"
          @click="load"
        />
      </el-tooltip>
    </template>

    <div class="page-shell forum-resource-page" data-density="compact">
      <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
        <template #default>
          <el-button link type="primary" @click="load">重试</el-button>
        </template>
      </el-alert>

      <el-skeleton v-if="initialLoading" :rows="10" animated />

      <SectionCard v-else-if="!loadError || list.length > 0" dense>
        <template #title>资源列表</template>
        <template #actions>
          <el-tag effect="plain">共 {{ total }} 条</el-tag>
        </template>

        <Toolbar dense>
          <template #filters>
            <el-input
              v-model="searchDraft"
              class="forum-search"
              clearable
              :maxlength="200"
              :prefix-icon="Search"
              placeholder="按标题搜索"
              aria-label="按标题搜索论坛资源"
              @keyup.enter="applySearch"
              @clear="clearSearch"
            />
            <el-button :icon="Search" :loading="loading" @click="applySearch">搜索</el-button>
          </template>
        </Toolbar>

        <EmptyState
          v-if="list.length === 0"
          :title="hasAppliedSearch ? '未找到匹配的论坛资源' : '暂无论坛资源'"
          :description="hasAppliedSearch ? '请调整标题关键词后重试' : '当前没有可展示的论坛资源'"
        />

        <template v-else>
          <div class="table-wrap">
            <el-table v-loading="loading" class="forum-table" :data="list" row-key="id" border>
              <el-table-column label="标题" min-width="280">
                <template #default="{ row }">
                  <div class="forum-title-cell">
                    <a
                      class="forum-link forum-link--title"
                      :href="row.url"
                      target="_blank"
                      rel="noopener noreferrer"
                      :title="row.title"
                    >
                      {{ row.title }}
                    </a>
                    <StatusIndicator
                      v-if="row.inspection_status === 'restricted'"
                      label="受限"
                      tone="warning"
                    />
                  </div>
                </template>
              </el-table-column>

              <el-table-column label="文件链接" min-width="180">
                <template #default="{ row }">
                  <div v-if="row.attachments?.length" class="resource-links">
                    <a
                      v-for="(attachment, index) in row.attachments"
                      :key="`${row.id}:attachment:${index}`"
                      class="forum-link"
                      :href="attachment"
                      target="_blank"
                      rel="noopener noreferrer"
                      :title="attachment"
                    >
                      文件 {{ index + 1 }}
                    </a>
                  </div>
                  <span v-else class="empty-resource">无</span>
                </template>
              </el-table-column>

              <el-table-column label="ED2K" min-width="280">
                <template #default="{ row }">
                  <div v-if="row.ed2k_links?.length" class="resource-links">
                    <a
                      v-for="(ed2kLink, index) in row.ed2k_links"
                      :key="`${row.id}:ed2k:${index}`"
                      class="forum-link"
                      :href="ed2kLink"
                      :title="ed2kLink"
                    >
                      {{ getEd2kLinkLabel(ed2kLink) }}
                    </a>
                  </div>
                  <span v-else class="empty-resource">无</span>
                </template>
              </el-table-column>

              <el-table-column label="发现时间" width="168">
                <template #default="{ row }">
                  <time class="forum-time tabular-num" :datetime="row.observed_at">
                    {{ formatAdminDateTime(row.observed_at) }}
                  </time>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="action-row">
            <AdminTablePagination
              :current-page="query.page"
              :page-size="query.page_size"
              layout="total, prev, pager, next"
              :total="total"
              :disabled="loading"
              @current-change="setPage"
            />
          </div>
        </template>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.forum-resource-page {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-4);
}

.forum-table {
  min-width: 56rem;
}

.forum-table :deep(.el-table__cell) {
  vertical-align: top;
}

.forum-search {
  width: min(24rem, 70vw);
}

.forum-title-cell {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: var(--space-2);
}

.forum-title-cell .status-indicator {
  flex: 0 0 auto;
  margin-top: 0.125rem;
}

.resource-links {
  display: grid;
  min-width: 0;
  gap: var(--space-2);
}

.forum-link {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  color: var(--primary);
  line-height: var(--leading-body);
  overflow-wrap: anywhere;
  text-decoration: none;
  transition: color 160ms ease;
}

.forum-link--title {
  min-width: 0;
  color: var(--text-primary);
  font-weight: 600;
}

.forum-link:hover {
  color: var(--primary-strong);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.forum-link:focus-visible {
  border-radius: var(--radius-sm);
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}

.empty-resource,
.forum-time {
  color: var(--text-muted);
}

.action-row {
  padding-top: var(--space-2);
}

@media (max-width: 63.9375rem) {
  .forum-refresh {
    width: 44px;
    height: 44px;
    min-width: 44px;
    min-height: 44px;
  }

  .forum-link {
    min-height: 44px;
    align-items: center;
  }
}

@media (prefers-reduced-motion: reduce) {
  .forum-link {
    transition: none;
  }
}
</style>
