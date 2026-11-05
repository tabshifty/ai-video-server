<script>
const STATUS_TONES = new Set(['neutral', 'success', 'warning', 'danger', 'info'])

function isNonEmptyStatusLabel(label) {
  return typeof label === 'string' && label.trim().length > 0
}

function isStatusTone(tone) {
  return STATUS_TONES.has(tone)
}
</script>

<script setup>
defineProps({
  label: { type: String, required: true, validator: isNonEmptyStatusLabel },
  tone: { type: String, default: 'neutral', validator: isStatusTone },
  icon: { type: [Object, Function], default: null }
})
</script>

<template>
  <span
    class="status-indicator"
    :class="`status-indicator--${STATUS_TONES.has(tone) ? tone : 'neutral'}`"
    :aria-label="label"
  >
    <el-icon v-if="icon" aria-hidden="true"><component :is="icon" /></el-icon>
    <span v-else class="status-indicator__dot" aria-hidden="true" />
    <span class="status-indicator__label">{{ label }}</span>
  </span>
</template>

<style scoped>
.status-indicator {
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  gap: var(--space-1);
  padding: 1px var(--space-2);
  border: 1px solid color-mix(in srgb, currentColor 24%, transparent);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: color-mix(in srgb, currentColor 8%, var(--bg-surface));
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.status-indicator__dot {
  width: 6px;
  height: 6px;
  flex: 0 0 auto;
  border: 0;
  border-radius: 50%;
  background: currentColor;
}

.status-indicator__label {
  min-width: 0;
  overflow-wrap: anywhere;
  white-space: normal;
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
