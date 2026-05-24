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
  margin-bottom: var(--space-1);
  color: var(--primary);
  font-weight: 600;
}

.progress-note {
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--text-small);
}
</style>
