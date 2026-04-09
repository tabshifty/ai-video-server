<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Layout from '../components/Layout.vue'
import { deleteAdminVideo, getAdminVideoDetail, getAdminVideos, retranscodeVideo, updateAdminVideo } from '../api/admin'

const list = ref([])
const total = ref(0)
const detail = ref(null)
const detailVisible = ref(false)

const query = reactive({ page: 1, page_size: 20, q: '', type: '', status: '' })

async function load() {
  const data = await getAdminVideos(query)
  list.value = data.items || []
  total.value = data.total_count || 0
}

async function showDetail(row) {
  detail.value = await getAdminVideoDetail(row.id)
  detailVisible.value = true
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
}

async function saveDetail() {
  await updateAdminVideo(detail.value.id, detail.value)
  ElMessage.success('保存成功')
  detailVisible.value = false
  await load()
}

onMounted(load)
</script>

<template>
  <Layout>
    <el-card>
      <el-form inline>
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

      <el-table :data="list" border>
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column prop="type" label="类型" width="110" />
        <el-table-column prop="status" label="状态" width="120" />
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

      <div style="margin-top: 12px; display:flex; justify-content:flex-end">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          layout="total, prev, pager, next"
          :total="total"
          @current-change="load"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailVisible" title="视频详情" width="760px">
      <el-form v-if="detail" label-width="90px">
        <el-form-item label="标题"><el-input v-model="detail.title" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="detail.description" type="textarea" rows="4" /></el-form-item>
        <el-form-item label="封面"><el-input v-model="detail.thumbnail" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="detailVisible=false">取消</el-button>
        <el-button type="primary" @click="saveDetail">保存</el-button>
      </template>
    </el-dialog>
  </Layout>
</template>
