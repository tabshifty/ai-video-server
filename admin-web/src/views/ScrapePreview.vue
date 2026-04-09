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
    <el-card>
      <el-form inline>
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

    <el-row :gutter="12" style="margin-top:12px">
      <el-col v-for="item in candidates" :key="item.tmdb_id" :span="8">
        <el-card shadow="hover" @click="choose(item)">
          <div style="font-weight:700">{{ item.title }}</div>
          <div style="color:#64748b;margin:6px 0">评分: {{ item.vote_average || '-' }}</div>
          <div style="font-size:13px;line-height:1.5">{{ item.overview }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top:12px">
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
  </Layout>
</template>
