<script setup>
defineProps({
  label: { type: String, required: true },
  tone: { type: String, default: 'neutral' },
  icon: { type: [Object, Function], default: null }
})
</script>

<template>
  <span class="status-indicator" :class="`status-indicator--${tone}`" :aria-label="label">
    <el-icon v-if="icon" aria-hidden="true"><component :is="icon" /></el-icon>
    <span v-else class="status-indicator__dot" aria-hidden="true" />
    <span>{{ label }}</span>
  </span>
</template>

<style scoped>
.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  white-space: nowrap;
}

.status-indicator__dot {
  width: 8px;
  height: 8px;
  flex: 0 0 auto;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.status-indicator--success {
  color: var(--success-600);
}

.status-indicator--warning {
  color: var(--warning-600);
}

.status-indicator--danger {
  color: var(--danger-600);
}

.status-indicator--info {
  color: var(--info-600);
}
</style>
