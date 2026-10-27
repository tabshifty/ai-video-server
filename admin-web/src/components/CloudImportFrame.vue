<script setup lang="ts">
import { ElDialog } from 'element-plus'

defineOptions({ inheritAttrs: false })
defineProps<{ inline?: boolean; modelValue?: boolean; title?: string }>()
</script>

<template>
  <section v-if="inline && modelValue" class="cloud-import-inline" :aria-label="title">
    <h2 v-if="title">{{ title }}</h2>
    <slot />
    <footer class="cloud-import-inline__actions"><slot name="footer" /></footer>
  </section>
  <ElDialog v-else-if="!inline" v-bind="$attrs" :model-value="modelValue" :title="title">
    <slot />
    <template #footer><slot name="footer" /></template>
  </ElDialog>
</template>

<style scoped>
.cloud-import-inline{display:grid;gap:var(--space-4);min-width:0;}
.cloud-import-inline h2{font-size:18px;margin:0;}
.cloud-import-inline__actions{display:flex;justify-content:flex-end;gap:var(--space-2);flex-wrap:wrap;}
</style>
