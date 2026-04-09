<script setup>
defineProps({
  percentage: { type: Number, default: 0 },
  statusText: { type: String, default: '' }
})
</script>

<template>
  <div style="margin-top: 10px">
    <div style="display:flex;justify-content:space-between;margin-bottom:6px">
      <span>上传进度</span>
      <span>{{ percentage }}%</span>
    </div>
    <el-progress :percentage="percentage" :stroke-width="16" />
    <div style="margin-top: 6px; color: #64748b">{{ statusText }}</div>
  </div>
</template>
