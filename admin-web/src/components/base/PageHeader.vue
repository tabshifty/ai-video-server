<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  }
})
</script>

<template>
  <section class="page-header-shell">
    <div class="page-header-shell__copy">
      <h1 class="page-header-shell__title">{{ title }}</h1>
      <p v-if="subtitle" class="page-header-shell__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header-shell__actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<style scoped>
.page-header-shell {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.page-header-shell__copy {
  min-width: 0;
}

.page-header-shell__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
  font-weight: 600;
}

.page-header-shell__subtitle {
  margin: var(--space-1) 0 0;
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.page-header-shell__actions {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--space-2);
}

@media (max-width: 48rem) {
  .page-header-shell {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
