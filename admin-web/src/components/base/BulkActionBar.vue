<script setup>
defineProps({
  count: {
    type: Number,
    required: true
  },
  actions: {
    type: Array,
    default: () => []
  }
})
</script>

<template>
  <transition name="bulk-action">
    <aside v-if="count > 0" class="bulk-action-bar">
      <div class="bulk-action-bar__copy">已选择 <strong>{{ count }}</strong> 项</div>
      <div class="bulk-action-bar__actions">
        <el-button
          v-for="action in actions"
          :key="action.label"
          :type="action.type || 'primary'"
          :icon="action.icon"
          @click="action.onClick"
        >
          {{ action.label }}
        </el-button>
      </div>
    </aside>
  </transition>
</template>

<style scoped>
.bulk-action-bar {
  position: sticky;
  bottom: var(--space-4);
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  background: var(--bg-surface);
  box-shadow: var(--shadow-lg);
}

.bulk-action-bar__copy {
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.bulk-action-bar__copy strong {
  color: var(--text-primary);
  font-weight: 600;
}

.bulk-action-bar__actions {
  display: inline-flex;
  gap: var(--space-2);
}

.bulk-action-enter-active,
.bulk-action-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--motion-easing-standard);
}

.bulk-action-enter-from,
.bulk-action-leave-to {
  opacity: 0;
  transform: translateY(var(--space-2));
}
</style>
