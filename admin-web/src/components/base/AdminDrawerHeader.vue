<script setup>
import { Close } from '@element-plus/icons-vue'

defineProps({
  title: { type: String, required: true },
  titleId: { type: String, required: true },
  titleClass: { type: [String, Array, Object], default: '' },
  close: { type: Function, required: true }
})
</script>

<template>
  <span :id="titleId" :class="titleClass" role="heading" aria-level="2">{{ title }}</span>
  <button
    type="button"
    class="el-drawer__close-btn"
    aria-label="关闭此对话框"
    title="关闭此对话框"
    @click="close"
  >
    <el-icon class="el-drawer__close" aria-hidden="true"><Close /></el-icon>
  </button>
</template>
