<script setup>
import { Box } from '@element-plus/icons-vue'

defineProps({
  icon: {
    type: [String, Object],
    default: null
  },
  title: {
    type: String,
    required: true
  },
  description: {
    type: String,
    default: ''
  }
})
</script>

<template>
  <section class="empty-state">
    <div class="empty-state__icon">
      <el-icon>
        <component :is="icon || Box" />
      </el-icon>
    </div>
    <h2 class="empty-state__title">{{ title }}</h2>
    <p v-if="description" class="empty-state__description">{{ description }}</p>
    <div v-if="$slots.action" class="empty-state__action">
      <slot name="action" />
    </div>
  </section>
</template>

<style scoped>
.empty-state {
  display: grid;
  place-items: center;
  gap: var(--space-2);
  min-height: calc(var(--space-12) * 4 + var(--space-6) + var(--space-1));
  padding: var(--space-8);
  color: var(--text-secondary);
  text-align: center;
}

.empty-state__icon {
  display: grid;
  width: calc(var(--space-12) + var(--space-2));
  height: calc(var(--space-12) + var(--space-2));
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  color: var(--primary);
  background: var(--primary-soft);
  font-size: var(--text-display);
}

.empty-state__title {
  margin: var(--space-2) 0 0;
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.empty-state__description {
  max-width: 34rem;
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.empty-state__action {
  margin-top: var(--space-2);
}
</style>
