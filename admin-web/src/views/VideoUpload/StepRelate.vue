<script setup>
import UploadProgress from '../../components/UploadProgress.vue'
import SectionCard from '../../components/base/SectionCard.vue'

defineProps({
  form: {
    type: Object,
    required: true
  },
  isShortType: {
    type: Boolean,
    default: false
  },
  tagOptions: {
    type: Array,
    default: () => []
  },
  loadingTags: {
    type: Boolean,
    default: false
  },
  collectionOptions: {
    type: Array,
    default: () => []
  },
  loadingCollections: {
    type: Boolean,
    default: false
  },
  imageCollectionOptions: {
    type: Array,
    default: () => []
  },
  loadingImageCollections: {
    type: Boolean,
    default: false
  },
  actorOptions: {
    type: Array,
    default: () => []
  },
  loadingActors: {
    type: Boolean,
    default: false
  },
  selectedCount: {
    type: Number,
    default: 0
  },
  uploading: {
    type: Boolean,
    default: false
  },
  canClearSelectedFiles: {
    type: Boolean,
    default: false
  },
  canClearRecords: {
    type: Boolean,
    default: false
  },
  progress: {
    type: Number,
    default: 0
  },
  hashProgress: {
    type: Number,
    default: 0
  },
  currentFileName: {
    type: String,
    default: ''
  },
  completedCount: {
    type: Number,
    default: 0
  },
  totalFiles: {
    type: Number,
    default: 0
  },
  successCount: {
    type: Number,
    default: 0
  },
  failedCount: {
    type: Number,
    default: 0
  },
  cancelledCount: {
    type: Number,
    default: 0
  },
  batchProgress: {
    type: Number,
    default: 0
  },
  uploadResults: {
    type: Array,
    default: () => []
  }
})

defineEmits([
  'load-tags',
  'load-collections',
  'load-image-collections',
  'search-actors',
  'submit',
  'cancel-upload',
  'clear-selected-files',
  'clear-upload-records'
])
</script>

<template>
  <div class="relate-step">
    <SectionCard>
      <template #title>关联信息</template>
      <el-form label-width="100px">
        <el-form-item label="视频标签">
          <el-select
            v-model="form.tags"
            multiple
            filterable
            remote
            reserve-keyword
            :remote-method="(keyword) => $emit('load-tags', keyword)"
            allow-create
            default-first-option
            clearable
            placeholder="可选择或输入标签"
            style="width: 100%"
            :loading="loadingTags"
          >
            <el-option v-for="tag in tagOptions" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>
        <el-form-item label="图片图集">
          <el-select
            v-model="form.imageCollectionID"
            filterable
            remote
            reserve-keyword
            clearable
            :remote-method="(keyword) => $emit('load-image-collections', keyword)"
            :loading="loadingImageCollections"
            placeholder="可选，仅可关联一个图片图集"
            style="width: 100%"
          >
            <el-option
              v-for="collection in imageCollectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isShortType" label="所属合集">
          <el-select
            v-model="form.collections"
            multiple
            filterable
            remote
            reserve-keyword
            clearable
            collapse-tags
            collapse-tags-tooltip
            :remote-method="(keyword) => $emit('load-collections', keyword)"
            :loading="loadingCollections"
            placeholder="可选，可多选"
            style="width: 100%"
          >
            <el-option
              v-for="collection in collectionOptions"
              :key="collection.value"
              :label="collection.label"
              :value="collection.value"
            />
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
            :remote-method="(keyword) => $emit('search-actors', keyword)"
            :loading="loadingActors"
            placeholder="可搜索演员，也可直接输入新演员姓名"
            style="width: 100%"
          >
            <el-option
              v-for="actor in actorOptions"
              :key="actor.value"
              :label="actor.label"
              :value="actor.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="upload-actions">
        <el-button type="primary" :loading="uploading" :disabled="selectedCount === 0" @click="$emit('submit')">开始上传</el-button>
        <el-button v-if="uploading" type="danger" @click="$emit('cancel-upload')">取消上传</el-button>
        <el-button :disabled="!canClearSelectedFiles" @click="$emit('clear-selected-files')">清空已选文件</el-button>
        <el-button :disabled="!canClearRecords" @click="$emit('clear-upload-records')">清空上传记录</el-button>
      </div>
    </SectionCard>

    <SectionCard>
      <template #title>进度区</template>
      <UploadProgress
        :percentage="progress"
        :status-text="`当前文件：${currentFileName || '-'} ｜ 哈希计算 ${hashProgress}%`"
      />
      <div class="batch-progress">
        <div class="batch-summary">
          批次进度：{{ completedCount }}/{{ totalFiles || selectedCount }}，成功 {{ successCount }}，失败 {{ failedCount }}，取消 {{ cancelledCount }}
        </div>
        <el-progress :percentage="batchProgress" />
      </div>
    </SectionCard>

    <SectionCard>
      <template #title>结果区</template>
      <div v-if="uploadResults.length" class="table-wrap upload-result">
        <el-table :data="uploadResults" size="small" border>
          <el-table-column prop="name" label="文件名" min-width="280" show-overflow-tooltip />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">
                {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : '已取消' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="信息" min-width="180" show-overflow-tooltip />
          <el-table-column prop="videoId" label="视频ID" min-width="260" show-overflow-tooltip />
        </el-table>
      </div>
      <div v-else class="upload-result-empty">暂无上传记录，开始上传后将在这里展示结果。</div>
    </SectionCard>
  </div>
</template>

<style scoped>
.relate-step {
  display: grid;
  gap: var(--space-4);
}

.upload-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  padding-left: 100px;
}

.batch-progress {
  display: grid;
  gap: var(--space-2);
}

.batch-summary {
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.upload-result-empty {
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--line-strong);
  color: var(--text-muted);
  font-size: var(--text-small);
  background: var(--bg-surface-muted);
}

@media (max-width: 768px) {
  .upload-actions {
    padding-left: 0;
  }
}
</style>
