<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { getRouteTransitionName } from './router/transition'

const route = useRoute()
const transitionName = computed(() => getRouteTransitionName(route))
</script>

<template>
  <router-view v-slot="{ Component }">
    <transition v-if="transitionName" :name="transitionName" mode="out-in">
      <component :is="Component" :key="route.fullPath" />
    </transition>
    <component :is="Component" v-else :key="route.fullPath" />
  </router-view>
</template>

<style scoped>
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

@media (prefers-reduced-motion: reduce) {
  .fade-slide-enter-active,
  .fade-slide-leave-active {
    transition-duration: 0.01ms;
  }
}
</style>
