<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import {
  deleteAdminVideo,
  getAdminVideoDetail,
  getAdminVideoPlayURL,
  getAdminVideos,
  retranscodeVideo,
  updateAdminVideo
} from '../api/admin'

const list = ref([])
const total = ref(0)
const detail = ref(null)
const detailVisible = ref(false)
const playURL = ref('')
const playExpiresAt = ref(0)
const loadingPlayURL = ref(false)

const query = reactive({ page: 1, page_size: 20, q: '', type: '', status: '' })

async function load() {
  const data = await getAdminVideos(query)
  list.value = data.items || []
  total.value = data.total_count || 0
}

async function showDetail(row) {
  detail.value = await getAdminVideoDetail(row.id)
  detailVisible.value = true
  playURL.value = ''
  playExpiresAt.value = 0
  if (detail.value?.status === 'ready') {
    await refreshPlayURL()
  }
}

async function refreshPlayURL() {
  if (!detail.value?.id) {
    return
  }
  if (detail.value.status !== 'ready') {
    playURL.value = ''
    playExpiresAt.value = 0
    return
  }
  loadingPlayURL.value = true
  try {
    const data = await getAdminVideoPlayURL(detail.value.id)
    playURL.value = data.signed_url || ''
    playExpiresAt.value = Number(data.expires_at || 0)
  } finally {
    loadingPlayURL.value = false
  }
}

async function doDelete(row) {
  await ElMessageBox.confirm(`确认删除 ${row.title} ?`, '警告')
  await deleteAdminVideo(row.id)
  ElMessage.success('删除成功')
  await load()
}

async function doRetranscode(row) {
  await ElMessageBox.confirm('确认重新转码?', '提示')
  await retranscodeVideo(row.id)
  ElMessage.success('已加入转码队列')
  await load()
}

async function saveDetail() {
  const payload = {
    title: detail.value.title,
    description: detail.value.description,
    thumbnail: detail.value.thumbnail_path,
    tags: detail.value.tags || [],
    status: detail.value.status,
    metadata: detail.value.metadata || {}
  }
  await updateAdminVideo(detail.value.id, payload)
  ElMessage.success('保存成功')
  detailVisible.value = false
  await load()
}

onMounted(load)
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">视频管理</h1>
          <p class="page-subtitle">筛选、查看、重转码与元数据编辑</p>
        </div>
      </div>

      <el-card class="soft-card">
        <el-form inline class="filter-form">
          <el-form-item><el-input v-model="query.q" placeholder="标题/标签搜索" /></el-form-item>
          <el-form-item>
            <el-select v-model="query.type" placeholder="类型" clearable style="width:120px">
              <el-option label="short" value="short" />
              <el-option label="movie" value="movie" />
              <el-option label="episode" value="episode" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-select v-model="query.status" placeholder="状态" clearable style="width:120px">
              <el-option label="uploaded" value="uploaded" />
              <el-option label="scraping" value="scraping" />
              <el-option label="processing" value="processing" />
              <el-option label="ready" value="ready" />
              <el-option label="failed" value="failed" />
            </el-select>
          </el-form-item>
          <el-form-item><el-button type="primary" @click="load">查询</el-button></el-form-item>
        </el-form>

        <div class="table-wrap">
          <el-table :data="list" border>
            <el-table-column prop="title" label="标题" min-width="220" />
            <el-table-column prop="type" label="类型" width="110" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.status === 'ready' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="upload_user" label="上传用户" width="140" />
            <el-table-column prop="created_at" label="上传时间" width="180" />
            <el-table-column label="操作" width="300">
              <template #default="{ row }">
                <el-button size="small" @click="showDetail(row)">详情</el-button>
                <el-button size="small" @click="doRetranscode(row)">重转码</el-button>
                <el-button size="small" type="danger" @click="doDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="action-row">
          <el-pagination
            v-model:current-page="query.page"
            v-model:page-size="query.page_size"
            layout="total, prev, pager, next"
            :total="total"
            @current-change="load"
          />
        </div>
      </el-card>
    </div>

    <el-dialog v-model="detailVisible" title="视频详情" width="760px">
      <el-form v-if="detail" label-width="90px">
        <el-form-item label="标题"><el-input v-model="detail.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="detail.description" type="textarea" rows="4" /></el-form-item>
        <el-form-item label="封面"><el-input v-model="detail.thumbnail_path" /></el-form-item>

        <el-form-item label="播放预览">
          <div class="play-preview">
            <div class="play-actions">
              <el-button
                type="primary"
                plain
                size="small"
                :loading="loadingPlayURL"
                :disabled="detail.status !== 'ready'"
                @click="refreshPlayURL"
              >
                刷新播放链接
              </el-button>
              <span v-if="playExpiresAt > 0" class="play-expire">有效期至：{{ new Date(playExpiresAt * 1000).toLocaleString() }}</span>
            </div>

            <el-alert
              v-if="detail.status !== 'ready'"
              type="warning"
              :closable="false"
              title="当前视频未就绪，只有 ready 状态才可播放。"
            />

            <el-alert
              v-else-if="!playURL"
              type="info"
              :closable="false"
              title="点击“刷新播放链接”后可预览播放。"
            />

            <video
              v-else
              class="preview-player"
              :src="playURL"
              controls
              preload="metadata"
            />
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="detailVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDetail">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>

<style scoped>
.play-preview {
  width: 100%;
}

.play-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.play-expire {
  color: #6b7280;
  font-size: 12px;
}

.preview-player {
  width: 100%;
  max-height: 360px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  background: #000;
}
</style>
