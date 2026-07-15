<script setup>
defineProps({
  items: { type: Array, required: true },
  ariaLabel: { type: String, default: '指标摘要' }
})
</script>

<template>
  <section class="metric-strip" :aria-label="ariaLabel">
    <div
      v-for="item in items"
      :key="item.key || item.label"
      class="metric-strip__item"
      :class="`metric-strip__item--${item.tone || 'neutral'}`"
    >
      <div class="metric-strip__label">
        <span>{{ item.label }}</span>
        <em v-if="item.scope">{{ item.scope }}</em>
      </div>
      <strong class="metric-strip__value tabular-num">{{ item.value }}</strong>
    </div>
  </section>
</template>

<style scoped>
.metric-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-block: 1px solid var(--line-soft);
  background: var(--bg-surface);
}

.metric-strip__item {
  min-width: 0;
  padding: var(--space-3);
  border-right: 1px solid var(--line-soft);
}

.metric-strip__item:last-child {
  border-right: 0;
}

.metric-strip__label {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.metric-strip__label span,
.metric-strip__label em,
.metric-strip__value {
  min-width: 0;
  overflow-wrap: anywhere;
}

.metric-strip__label em {
  color: var(--text-muted);
  font-style: normal;
}

.metric-strip__value {
  display: block;
  margin-top: var(--space-1);
  color: var(--text-primary);
  font-size: var(--text-kpi);
  line-height: var(--leading-kpi);
  font-weight: 600;
}

.metric-strip__item--success .metric-strip__value {
  color: var(--success-600);
}

.metric-strip__item--warning .metric-strip__value {
  color: var(--warning-600);
}

.metric-strip__item--danger .metric-strip__value {
  color: var(--danger-600);
}

.metric-strip__item--info .metric-strip__value {
  color: var(--info-600);
}

@media (max-width: 63.9375rem) {
  .metric-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .metric-strip__item:nth-child(2n) {
    border-right: 0;
  }
}

@media (max-width: 36rem) {
  .metric-strip {
    grid-template-columns: 1fr;
  }

  .metric-strip__item {
    border-right: 0;
  }
}
</style>
