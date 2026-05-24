<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowRight, Search } from '@element-plus/icons-vue'
import {
  registerCommandPaletteOpener,
  searchMenuItems
} from './commandPalette.helpers'

const router = useRouter()
const visible = ref(false)
const query = ref('')
const activeIndex = ref(0)
const inputRef = ref(null)

const results = computed(() => searchMenuItems(query.value))

watch(results, () => {
  activeIndex.value = 0
})

function open() {
  visible.value = true
  query.value = ''
  activeIndex.value = 0
  nextTick(() => inputRef.value?.focus())
}

function close() {
  visible.value = false
}

function move(delta) {
  if (!results.value.length) return
  activeIndex.value = (activeIndex.value + delta + results.value.length) % results.value.length
}

function select(item = results.value[activeIndex.value]) {
  if (!item) return
  close()
  router.push(item.path)
}

function onPanelKeydown(event) {
  if (event.key === 'Escape') {
    event.preventDefault()
    close()
    return
  }
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    move(1)
    return
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    move(-1)
    return
  }
  if (event.key === 'Enter') {
    event.preventDefault()
    select()
  }
}

function onGlobalKeydown(event) {
  if (event.code === 'KeyK' && (event.metaKey || event.ctrlKey)) {
    event.preventDefault()
    open()
  }
}

onMounted(() => {
  registerCommandPaletteOpener(open)
  window.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  registerCommandPaletteOpener(null)
  window.removeEventListener('keydown', onGlobalKeydown)
})
</script>

<template>
  <teleport to="body">
    <transition name="command-palette">
      <div v-if="visible" class="command-palette" role="dialog" aria-modal="true" aria-label="命令面板">
        <button class="command-palette__backdrop" type="button" aria-label="关闭命令面板" @click="close" />
        <section class="command-palette__panel" @keydown="onPanelKeydown">
          <label class="command-palette__search">
            <el-icon><Search /></el-icon>
            <input
              ref="inputRef"
              v-model="query"
              class="command-palette__input"
              type="search"
              placeholder="搜索页面、输入拼音首字母或路径"
              autocomplete="off"
            />
            <kbd>Esc</kbd>
          </label>
          <div class="command-palette__list" role="listbox">
            <button
              v-for="(item, index) in results"
              :key="item.path"
              class="command-palette__item"
              :class="{ 'is-active': index === activeIndex }"
              type="button"
              role="option"
              :aria-selected="index === activeIndex"
              @mouseenter="activeIndex = index"
              @click="select(item)"
            >
              <span>
                <strong>{{ item.label }}</strong>
                <small>{{ item.groupLabel }} · {{ item.path }}</small>
              </span>
              <el-icon><ArrowRight /></el-icon>
            </button>
          </div>
        </section>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.command-palette {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: grid;
  place-items: start center;
  padding-top: 12vh;
}

.command-palette__backdrop {
  position: fixed;
  inset: 0;
  border: 0;
  background: color-mix(in srgb, var(--slate-950) 28%, transparent);
}

.command-palette__panel {
  position: relative;
  z-index: 1;
  width: min(40rem, calc(100vw - var(--space-8)));
  overflow: hidden;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-2xl);
  background: var(--bg-surface);
  box-shadow: var(--shadow-lg);
}

.command-palette__search {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: 1px solid var(--line-soft);
  color: var(--text-muted);
}

.command-palette__input {
  min-width: 0;
  border: 0;
  outline: 0;
  color: var(--text-primary);
  background: transparent;
  font: inherit;
}

.command-palette__search kbd {
  padding: calc(var(--space-1) / 2) var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.command-palette__list {
  display: grid;
  max-height: 26.25rem;
  padding: var(--space-2);
  overflow: auto;
}

.command-palette__item {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3);
  border: 0;
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.command-palette__item.is-active,
.command-palette__item:hover {
  background: var(--primary-soft);
}

.command-palette__item strong,
.command-palette__item small {
  display: block;
}

.command-palette__item strong {
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
}

.command-palette__item small {
  margin-top: var(--space-1);
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.command-palette-enter-active,
.command-palette-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.command-palette-enter-from,
.command-palette-leave-to {
  opacity: 0;
}
</style>
