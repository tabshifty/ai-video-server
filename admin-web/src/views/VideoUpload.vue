<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import UploadProgress from '../components/UploadProgress.vue'
import { checkUpload, uploadAbort, uploadChunk, uploadComplete, uploadInit } from '../api/video'
import { getAdminActors } from '../api/admin'
import { sha256File } from '../utils/hash'

const file = ref(null)
const progress = ref(0)
const hashProgress = ref(0)
const uploading = ref(false)
const result = ref(null)
const abortController = ref(null)
const sessionId = ref('')
const actorOptions = ref([])
const loadingActors = ref(false)

const form = reactive({
  type: 'short',
  title: '',
  description: '',
  tags: [],
  actors: []
})
const presetTags = ['action', 'comedy', 'drama', 'documentary', 'anime', 'music', 'sports', 'family']
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

function onFileChange(raw) {
  file.value = raw.raw
}

function splitActorSelection(values) {
  const actorIDs = []
  const actorNames = []
  const seenID = new Set()
  const seenName = new Set()
  for (const item of values || []) {
    const value = String(item || '').trim()
    if (!value) continue
    if (uuidPattern.test(value)) {
      const key = value.toLowerCase()
      if (seenID.has(key)) continue
      seenID.add(key)
      actorIDs.push(value)
      continue
    }
    const key = value.toLowerCase()
    if (seenName.has(key)) continue
    seenName.add(key)
    actorNames.push(value.replace(/\s+/g, ' '))
  }
  return { actorIDs, actorNames }
}

async function searchActors(keyword = '') {
  loadingActors.value = true
  try {
    const data = await getAdminActors({
      q: keyword,
      active: 1,
      page: 1,
      page_size: 20
    })
    actorOptions.value = (data.items || []).map((item) => ({
      value: item.id,
      label: item.name
    }))
  } finally {
    loadingActors.value = false
  }
}

async function submit() {
  if (!file.value) {
    ElMessage.warning('请选择文件')
    return
  }
  uploading.value = true
  progress.value = 0
  hashProgress.value = 0
  abortController.value = new AbortController()
  try {
    const hash = await sha256File(file.value, (p) => (hashProgress.value = p))
    const check = await checkUpload({ hash, file_size: file.value.size })
    if (check.exists) {
      ElMessage.success('秒传命中，文件已存在')
      result.value = check
      return
    }

    const normalizedTags = Array.from(
      new Map(
        (form.tags || [])
          .map((tag) => String(tag).trim())
          .filter((tag) => tag !== '')
          .map((tag) => [tag.toLowerCase(), tag.toLowerCase()])
      ).values()
    )
    const { actorIDs, actorNames } = splitActorSelection(form.actors)

    const chunkSize = 4 * 1024 * 1024
    const totalChunks = Math.ceil(file.value.size / chunkSize)
    const initResp = await uploadInit({
      filename: file.value.name,
      file_size: file.value.size,
      chunk_size: chunkSize,
      total_chunks: totalChunks,
      hash,
      type: form.type,
      title: form.title,
      description: form.description,
      tags: normalizedTags,
      actor_ids: actorIDs,
      actor_names: actorNames
    })
    sessionId.value = initResp.upload_session_id

    for (let i = 0; i < totalChunks; i += 1) {
      const start = i * chunkSize
      const end = Math.min(start + chunkSize, file.value.size)
      const chunk = file.value.slice(start, end)
      await uploadChunk(sessionId.value, i, chunk, abortController.value.signal)
      progress.value = Math.round(((i + 1) / totalChunks) * 100)
    }

    result.value = await uploadComplete(sessionId.value)
    ElMessage.success('上传成功')
  } catch (e) {
    ElMessage.error(e.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

async function cancelUpload() {
  abortController.value?.abort()
  if (sessionId.value) {
    await uploadAbort(sessionId.value)
  }
  ElMessage.warning('已取消上传')
}

onMounted(() => {
  searchActors().catch(() => {})
})
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">上传中心</h1>
          <p class="page-subtitle">支持分片上传、秒传检测与上传取消</p>
        </div>
      </div>

      <el-card class="soft-card">
      <el-form label-width="100px">
        <el-form-item label="视频文件">
          <el-upload drag :auto-upload="false" :on-change="onFileChange" :limit="1" class="upload-drop">
            <el-icon><UploadFilled /></el-icon>
            <div>拖拽文件到此，或点击选择文件</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 220px">
            <el-option label="短视频" value="short" />
            <el-option label="电影" value="movie" />
            <el-option label="剧集分集" value="episode" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
        <el-form-item label="视频标签">
          <el-select
            v-model="form.tags"
            multiple
            filterable
            allow-create
            default-first-option
            clearable
            placeholder="可选择或输入标签"
            style="width: 100%"
          >
            <el-option v-for="tag in presetTags" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="关联演员">
          <el-select
            v-model="form.actors"
            multiple
            filterable
            remote
            reserve-keyword
            allow-create
            default-first-option
            clearable
            :remote-method="searchActors"
            :loading="loadingActors"
            placeholder="可搜索演员，也可直接输入新演员姓名"
            style="width: 100%"
          >
            <el-option v-for="actor in actorOptions" :key="actor.value" :label="actor.label" :value="actor.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="uploading" @click="submit">开始上传</el-button>
          <el-button v-if="uploading" type="danger" @click="cancelUpload">取消上传</el-button>
        </el-form-item>
      </el-form>

      <UploadProgress :percentage="progress" :status-text="`哈希计算 ${hashProgress}%`" />
      <div v-if="result" class="upload-result">视频ID: {{ result.video_id }}</div>
    </el-card>
    </div>
  </Layout>
</template>

<style scoped>
.upload-drop :deep(.el-upload-dragger) {
  border-radius: 14px;
  background: linear-gradient(160deg, #fff1f2 0%, #fff 100%);
}

.upload-result {
  margin-top: 10px;
  color: #881337;
  font-family: 'Fira Code', monospace;
}
</style>
