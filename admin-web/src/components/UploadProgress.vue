<script setup>
defineProps({
  percentage: { type: Number, default: 0 },
  statusText: { type: String, default: '' }
})
</script>

<template>
  <div class="upload-progress">
    <div class="progress-head">
      <span>上传进度</span>
      <span>{{ percentage }}%</span>
    </div>
    <el-progress :percentage="percentage" :stroke-width="16" />
    <div class="progress-note">{{ statusText }}</div>
  </div>
</template>

<style scoped>
.upload-progress {
  margin-top: 10px;
}

.progress-head {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  color: #7f1d1d;
  font-weight: 600;
}

.progress-note {
  margin-top: 6px;
  color: #6b7280;
  font-size: 13px;
}
</style>
