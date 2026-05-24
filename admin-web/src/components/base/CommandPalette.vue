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
const panelRef = ref(null)
let previouslyFocusedElement = null

function isTypingInEditableTarget(target) {
  if (!target || target.nodeType !== 1) return false
  if (target.isContentEditable) return true
  const tag = target.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true
  return false
}

const results = computed(() => searchMenuItems(query.value))

watch(results, () => {
  activeIndex.value = 0
})

function open() {
  if (visible.value) return
  if (typeof document !== 'undefined') {
    previouslyFocusedElement = document.activeElement
  }
  visible.value = true
  query.value = ''
  activeIndex.value = 0
  nextTick(() => inputRef.value?.focus())
}

function close() {
  visible.value = false
  // 关闭时把焦点还给打开 palette 前的可聚焦节点，避免焦点丢到 body 上
  nextTick(() => {
    const target = previouslyFocusedElement
    previouslyFocusedElement = null
    if (target && typeof target.focus === 'function') {
      try {
        target.focus({ preventScroll: true })
      } catch (_) {
        // 节点已卸载等情况忽略
      }
    }
  })
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
  if (event.key === 'Enter' && !event.isComposing) {
    event.preventDefault()
    select()
    return
  }
  if (event.key === 'Tab') {
    // focus trap：Tab / Shift+Tab 在 panel 内部循环，不允许逃逸到底层页面
    const focusables = panelRef.value?.querySelectorAll(
      'input, button:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
    if (!focusables || focusables.length === 0) return
    const first = focusables[0]
    const last = focusables[focusables.length - 1]
    const activeEl = typeof document !== 'undefined' ? document.activeElement : null
    if (event.shiftKey && activeEl === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && activeEl === last) {
      event.preventDefault()
      first.focus()
    }
  }
}

function onGlobalKeydown(event) {
  // Esc 关 palette 提到 document 级，保证 backdrop 焦点 / 任意子节点焦点状态下都能关闭
  if (event.key === 'Escape' && visible.value) {
    event.preventDefault()
    close()
    return
  }
  if (event.code !== 'KeyK' || !(event.metaKey || event.ctrlKey)) return
  // 用户在 input / textarea / contenteditable 内打字时不抢键；IME 拼音输入过程中也不抢键
  if (event.isComposing || event.repeat) return
  if (isTypingInEditableTarget(event.target)) return
  event.preventDefault()
  open()
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
        <section ref="panelRef" class="command-palette__panel" @keydown="onPanelKeydown">
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
