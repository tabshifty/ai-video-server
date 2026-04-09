<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Layout from '../components/Layout.vue'
import { scrapeConfirm, scrapePreview } from '../api/admin'

const form = reactive({ video_id: '', title: '', year: new Date().getFullYear(), type: 'movie' })
const edit = reactive({ video_id: '', tmdb_id: 0, title: '', overview: '', poster_url: '', release_date: '' })
const candidates = ref([])

async function doPreview() {
  const data = await scrapePreview(form)
  candidates.value = data.candidates || []
}

function choose(item) {
  edit.video_id = form.video_id
  edit.tmdb_id = item.tmdb_id
  edit.title = item.title
  edit.overview = item.overview
  edit.poster_url = item.poster_url || item.poster_path || ''
  edit.release_date = item.release_date || ''
}

async function doSave() {
  await scrapeConfirm(edit)
  ElMessage.success('刮削信息已保存')
}
</script>

<template>
  <Layout>
    <div class="page">
      <div class="page-header">
        <div>
          <h1 class="page-title">刮削管理</h1>
          <p class="page-subtitle">预览 TMDB 候选并手动修正元数据后确认保存</p>
        </div>
      </div>

    <el-card class="soft-card">
      <el-form inline class="filter-form">
        <el-form-item label="视频ID"><el-input v-model="form.video_id" style="width:300px" /></el-form-item>
        <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="年份"><el-input-number v-model="form.year" :min="1900" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width:120px">
            <el-option label="movie" value="movie" />
            <el-option label="tv" value="tv" />
          </el-select>
        </el-form-item>
        <el-form-item><el-button type="primary" @click="doPreview">预览</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12">
      <el-col v-for="item in candidates" :key="item.tmdb_id" :span="8">
        <el-card shadow="hover" class="soft-card interactive-card" @click="choose(item)">
          <div class="candidate-title">{{ item.title }}</div>
          <div class="candidate-score">评分: {{ item.vote_average || '-' }}</div>
          <div class="candidate-overview">{{ item.overview }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="soft-card">
      <template #header>编辑并确认</template>
      <el-form label-width="90px">
        <el-form-item label="TMDB ID"><el-input v-model="edit.tmdb_id" /></el-form-item>
        <el-form-item label="标题"><el-input v-model="edit.title" /></el-form-item>
        <el-form-item label="简介"><el-input v-model="edit.overview" type="textarea" rows="3" /></el-form-item>
        <el-form-item label="海报URL"><el-input v-model="edit.poster_url" /></el-form-item>
        <el-form-item label="发布日期"><el-input v-model="edit.release_date" /></el-form-item>
      </el-form>
      <el-button type="primary" @click="doSave">保存</el-button>
    </el-card>
    </div>
  </Layout>
</template>

<style scoped>
.candidate-title {
  font-weight: 700;
  color: #881337;
}

.candidate-score {
  color: #6b7280;
  margin: 6px 0;
}

.candidate-overview {
  font-size: 13px;
  line-height: 1.6;
  color: #374151;
  max-height: 82px;
  overflow: auto;
}
</style>
