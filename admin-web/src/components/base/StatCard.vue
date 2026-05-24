<script setup>
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  trend: {
    type: String,
    default: ''
  },
  delta: {
    type: String,
    default: ''
  }
})

function trendIcon() {
  if (props.trend === 'up') return ArrowUp
  if (props.trend === 'down') return ArrowDown
  return null
}
</script>

<template>
  <article class="stat-card">
    <div class="stat-card__label">{{ label }}</div>
    <div class="stat-card__value tabular-num">{{ value }}</div>
    <div v-if="delta" class="stat-card__delta" :class="`stat-card__delta--${trend || 'neutral'}`">
      <el-icon v-if="trendIcon()">
        <component :is="trendIcon()" />
      </el-icon>
      <span>{{ delta }}</span>
    </div>
  </article>
</template>

<style scoped>
.stat-card {
  display: grid;
  min-height: calc(var(--space-8) * 3 + var(--space-4));
  align-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.stat-card__label {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.stat-card__value {
  color: var(--text-primary);
  font-size: var(--text-display);
  line-height: var(--leading-display);
  font-weight: 600;
}

.stat-card__delta {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  gap: var(--space-1);
  padding: calc(var(--space-1) / 2) var(--space-2);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.stat-card__delta--up {
  color: var(--success-600);
  background: var(--success-50);
}

.stat-card__delta--down {
  color: var(--danger-600);
  background: var(--danger-50);
}
</style>
