<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import UploadProgress from '../components/UploadProgress.vue'
import { checkUpload, uploadAbort, uploadChunk, uploadComplete, uploadInit } from '../api/video'
import { sha256File } from '../utils/hash'

const file = ref(null)
const progress = ref(0)
const hashProgress = ref(0)
const uploading = ref(false)
const result = ref(null)
const abortController = ref(null)
const sessionId = ref('')

const form = reactive({
  type: 'short',
  title: '',
  description: '',
  tags: []
})

function onFileChange(raw) {
  file.value = raw.raw
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
      tags: form.tags
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
</script>

<template>
  <Layout>
    <el-card>
      <el-form label-width="100px">
        <el-form-item label="视频文件">
          <el-upload drag :auto-upload="false" :on-change="onFileChange" :limit="1">
            <el-icon><UploadFilled /></el-icon>
            <div>拖拽文件到此，或点击选择文件</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 220px">
            <el-option label="short" value="short" />
            <el-option label="movie" value="movie" />
            <el-option label="episode" value="episode" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="uploading" @click="submit">开始上传</el-button>
          <el-button v-if="uploading" type="danger" @click="cancelUpload">取消上传</el-button>
        </el-form-item>
      </el-form>

      <UploadProgress :percentage="progress" :status-text="`哈希计算 ${hashProgress}%`" />
      <div v-if="result" style="margin-top: 10px">视频ID: {{ result.video_id }}</div>
    </el-card>
  </Layout>
</template>
