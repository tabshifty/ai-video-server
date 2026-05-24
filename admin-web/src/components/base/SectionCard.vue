<script setup>
import { ref } from 'vue'
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'

defineProps({
  collapsible: {
    type: Boolean,
    default: false
  }
})

const expanded = ref(true)
</script>

<template>
  <section class="section-card">
    <header class="section-card__header">
      <button v-if="collapsible" class="section-card__toggle" type="button" @click="expanded = !expanded">
        <el-icon>
          <component :is="expanded ? ArrowDown : ArrowRight" />
        </el-icon>
      </button>
      <div class="section-card__copy">
        <h2 v-if="$slots.title" class="section-card__title">
          <slot name="title" />
        </h2>
        <p v-if="$slots.description" class="section-card__description">
          <slot name="description" />
        </p>
      </div>
      <div v-if="$slots.actions" class="section-card__actions">
        <slot name="actions" />
      </div>
    </header>
    <div v-show="expanded" class="section-card__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.section-card {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
}

.section-card__header {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: var(--space-2);
}

.section-card__copy {
  min-width: 0;
}

.section-card__title,
.section-card__description {
  margin: 0;
}

.section-card__title {
  color: var(--text-primary);
  font-size: var(--text-h2);
  line-height: var(--leading-h2);
  font-weight: 600;
}

.section-card__description {
  margin-top: var(--space-1);
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.section-card__actions {
  display: inline-flex;
  margin-left: auto;
  gap: var(--space-2);
}

.section-card__toggle {
  display: inline-grid;
  width: calc(var(--space-6) + var(--space-1));
  height: calc(var(--space-6) + var(--space-1));
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  background: var(--bg-surface);
  cursor: pointer;
}

.section-card__body {
  min-width: 0;
}
</style>
