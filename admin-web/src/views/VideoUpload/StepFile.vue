<script setup>
import { ref } from 'vue'
import { UploadFilled } from '@element-plus/icons-vue'
import SectionCard from '../../components/base/SectionCard.vue'

defineProps({
  fileList: {
    type: Array,
    default: () => []
  },
  isMovieType: {
    type: Boolean,
    default: false
  },
  selectedCount: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['update:fileList', 'file-change'])
const uploadRef = ref(null)

function onFileListUpdate(next) {
  emit('update:fileList', next)
}

function onFileChange(file, nextFileList) {
  emit('file-change', file, nextFileList)
}

function clearFiles() {
  uploadRef.value?.clearFiles()
}

defineExpose({ clearFiles })
</script>

<template>
  <SectionCard>
    <template #title>选文件</template>
    <el-upload
      ref="uploadRef"
      :file-list="fileList"
      drag
      :auto-upload="false"
      :on-change="onFileChange"
      :multiple="!isMovieType"
      :limit="isMovieType ? 1 : 999"
      class="upload-drop"
      @update:file-list="onFileListUpdate"
    >
      <el-icon><UploadFilled /></el-icon>
      <div>拖拽文件到此，或点击选择文件</div>
    </el-upload>
    <div class="upload-tip">
      <span v-if="isMovieType">电影仅支持单文件上传</span>
      <span v-else>当前类型支持批量上传，可一次选择多个文件</span>
      <span>已选择 {{ selectedCount }} 个文件</span>
    </div>
  </SectionCard>
</template>

<style scoped>
.upload-drop :deep(.el-upload-dragger) {
  border-radius: var(--radius-lg);
  background: var(--bg-surface-muted);
}

.upload-tip {
  margin-top: var(--space-3);
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  width: 100%;
  color: var(--text-muted);
  font-size: var(--text-small);
}

@media (max-width: 768px) {
  .upload-tip {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
  }
}
</style>
