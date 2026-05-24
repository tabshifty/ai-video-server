<script setup>
defineProps({
  dense: {
    type: Boolean,
    default: false
  }
})
</script>

<template>
  <section class="admin-toolbar" :class="{ 'admin-toolbar--dense': dense }">
    <div v-if="$slots.filters" class="admin-toolbar__filters">
      <slot name="filters" />
    </div>
    <div v-if="$slots.actions" class="admin-toolbar__actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<style scoped>
.admin-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) 0;
  flex-wrap: wrap;
}

.admin-toolbar--dense {
  padding: var(--space-2) 0;
}

.admin-toolbar__filters,
.admin-toolbar__actions {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.admin-toolbar__actions {
  margin-left: auto;
}
</style>
