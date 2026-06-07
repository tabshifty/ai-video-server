<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Avatar,
  DataAnalysis,
  Expand,
  Files,
  Film,
  Fold,
  List,
  MagicStick,
  Menu,
  Monitor,
  PictureFilled,
  Search,
  Setting,
  SwitchButton,
  Tools,
  UploadFilled,
  User
} from '@element-plus/icons-vue'
import { profileApi } from '../api/auth'
import { useAuthStore } from '../stores/auth'
import CommandPalette from './base/CommandPalette.vue'
import PageHeader from './base/PageHeader.vue'
import {
  adminShellNavGroups,
  findAdminNavItemByPath,
  openCommandPalette
} from './base/commandPalette.helpers'

const SIDEBAR_COLLAPSE_KEY = 'admin-sidebar-collapsed'
const DRAWER_BREAKPOINT = 1024
const COLLAPSE_BREAKPOINT = 1280

const iconMap = {
  Avatar,
  DataAnalysis,
  Files,
  Film,
  List,
  MagicStick,
  Monitor,
  PictureFilled,
  Setting,
  Tools,
  UploadFilled,
  User
}

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const mobileNavVisible = ref(false)
const shellContentRef = ref(null)
const mainContentRef = ref(null)
const userCollapsed = ref(readStoredCollapsed())
const viewportWidth = ref(readViewportWidth())
const profile = ref({
  username: '',
  role: auth.role || 'admin'
})

const navGroups = adminShellNavGroups
const matchedNavItem = computed(() => findAdminNavItemByPath(route.path))
const pageTitle = computed(() => matchedNavItem.value?.title || route.meta?.title || '管理后台')
const showShellPageHeader = computed(() => !route.meta?.hideShellPageHeader)
const isDrawerMode = computed(() => viewportWidth.value < DRAWER_BREAKPOINT)
const isAutoCollapsed = computed(() => viewportWidth.value < COLLAPSE_BREAKPOINT)
const isSidebarCollapsed = computed(() => !isDrawerMode.value && (userCollapsed.value || isAutoCollapsed.value))
const profileName = computed(() => profile.value.username || 'Admin')
const profileRole = computed(() => profile.value.role || auth.role || 'admin')
const profileInitial = computed(() => profileName.value.trim().slice(0, 1).toUpperCase() || 'A')

function readStoredCollapsed() {
  if (typeof window === 'undefined') return false
  try {
    return window.localStorage.getItem(SIDEBAR_COLLAPSE_KEY) === '1'
  } catch (_) {
    // Safari Private Mode / 企业策略禁用 storage 时 localStorage 抛 SecurityError；
    // 静默返回默认展开形态，避免 Layout 在 setup 阶段抛出导致整页白屏
    return false
  }
}

function readViewportWidth() {
  if (typeof window === 'undefined') return 1440
  return window.innerWidth
}

function resolveIcon(iconName) {
  return iconMap[iconName] || List
}

function isActive(item) {
  return matchedNavItem.value?.path === item.path
}

function updateViewportWidth() {
  viewportWidth.value = readViewportWidth()
}

function toggleSidebar() {
  userCollapsed.value = !userCollapsed.value
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(SIDEBAR_COLLAPSE_KEY, userCollapsed.value ? '1' : '0')
  } catch (_) {
    // 隐私模式 / 配额溢出时 setItem 抛 SecurityError 或 QuotaExceededError；
    // 本次会话内偏好仍生效，仅放弃持久化，不影响点击交互
  }
}

function closeMobileNav() {
  mobileNavVisible.value = false
}

function openMobileNav() {
  mobileNavVisible.value = true
}

function onSkipLinkClick() {
  shellContentRef.value?.scrollTo({ top: 0, behavior: 'auto' })
  mainContentRef.value?.focus({ preventScroll: true })
}

async function loadProfile() {
  try {
    const data = await profileApi()
    profile.value = {
      username: data?.username || '',
      role: data?.role || auth.role || 'admin'
    }
  } catch (error) {
    // 401 已由 axios 拦截器 handleAuthExpired 跳转登录页；这里只为其它错误（500 / 网络 / CORS）兜底
    profile.value = {
      username: '',
      role: auth.role || 'admin'
    }
    if (error?.response?.status !== 401) {
      console.warn('[Layout] loadProfile failed:', error)
    }
  }
}

async function onLogout() {
  await auth.logout()
  mobileNavVisible.value = false
  router.push('/login')
}

watch(
  () => route.fullPath,
  () => {
    mobileNavVisible.value = false
  }
)

// mobile drawer 打开时锁 body / html 滚动，避免触摸事件穿透到 main 区域；
// 离开本组件时无条件恢复，防止任何分支跳出未恢复
const restoreOverflow = { html: '', body: '' }
watch(mobileNavVisible, (visible) => {
  if (typeof document === 'undefined') return
  const html = document.documentElement
  const body = document.body
  if (visible) {
    restoreOverflow.html = html.style.overflow
    restoreOverflow.body = body.style.overflow
    html.style.overflow = 'hidden'
    body.style.overflow = 'hidden'
  } else {
    html.style.overflow = restoreOverflow.html
    body.style.overflow = restoreOverflow.body
  }
})

onMounted(() => {
  updateViewportWidth()
  loadProfile()
  window.addEventListener('resize', updateViewportWidth)
  // 部分 Android 浏览器横竖屏切换只发 orientationchange 不同步触发 resize，
  // 不监听会让 drawer 阈值（< 1024px）在转屏后滞留旧值
  window.addEventListener('orientationchange', updateViewportWidth)
})

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateViewportWidth)
    window.removeEventListener('orientationchange', updateViewportWidth)
  }
  if (typeof document !== 'undefined') {
    document.documentElement.style.overflow = restoreOverflow.html
    document.body.style.overflow = restoreOverflow.body
  }
})
</script>

<template>
  <div
    class="admin-shell"
    :class="{
      'is-collapsed': isSidebarCollapsed,
      'is-drawer': isDrawerMode,
      'is-mobile-nav-open': mobileNavVisible
    }"
  >
    <a class="skip-link" href="#main-content" @click.prevent="onSkipLinkClick">跳转到主要内容</a>

    <aside class="shell-aside" aria-label="分组导航">
      <div class="brand-block">
        <div class="brand-mark">VS</div>
        <div class="brand-copy">
          <strong>视频管理后台</strong>
          <span>Video Server</span>
        </div>
        <el-button class="collapse-button" text :aria-label="isSidebarCollapsed ? '展开侧栏' : '折叠侧栏'" @click="toggleSidebar">
          <el-icon><component :is="isSidebarCollapsed ? Expand : Fold" /></el-icon>
        </el-button>
      </div>

      <nav class="nav-groups" aria-label="分组导航">
        <section v-for="group in navGroups" :key="group.key" class="nav-group" :aria-label="`${group.label}分组`">
          <div class="nav-group__label">{{ group.label }}</div>
          <el-tooltip
            v-for="item in group.items"
            :key="item.path"
            :content="item.label"
            :disabled="!isSidebarCollapsed"
            placement="right"
          >
            <RouterLink
              class="nav-link"
              :class="{ 'is-active': isActive(item) }"
              :to="item.path"
              :aria-current="isActive(item) ? 'page' : undefined"
              @click="closeMobileNav"
            >
              <el-icon><component :is="resolveIcon(item.icon)" /></el-icon>
              <span class="nav-link__label">{{ item.label }}</span>
            </RouterLink>
          </el-tooltip>
        </section>
      </nav>

      <el-popover placement="top-start" trigger="click" popper-class="admin-profile-popper" :width="220">
        <template #reference>
          <button class="profile-chip" type="button">
            <span class="profile-chip__avatar">{{ profileInitial }}</span>
            <span class="profile-chip__copy">
              <strong>{{ profileName }}</strong>
              <em>{{ profileRole }}</em>
            </span>
          </button>
        </template>
        <div class="profile-panel">
          <div class="profile-panel__name">{{ profileName }}</div>
          <div class="profile-panel__role">admin</div>
          <el-button class="profile-panel__logout" type="danger" text :icon="SwitchButton" @click="onLogout">
            退出登录
          </el-button>
        </div>
      </el-popover>
    </aside>

    <section ref="shellContentRef" class="shell-content">
      <header class="shell-header">
        <div class="shell-header__left">
          <el-button class="mobile-nav-btn" text :icon="Menu" aria-label="打开导航菜单" @click="openMobileNav" />
          <PageHeader v-if="showShellPageHeader" class="shell-page-header" :title="pageTitle" />
        </div>
        <button class="command-trigger" type="button" @click="openCommandPalette">
          <el-icon><Search /></el-icon>
          <span>快速跳转</span>
          <kbd>⌘K</kbd>
        </button>
      </header>

      <main id="main-content" ref="mainContentRef" class="shell-main" tabindex="-1">
        <slot />
      </main>
    </section>

    <el-drawer
      v-model="mobileNavVisible"
      direction="ltr"
      size="calc(var(--admin-sidebar-width) + var(--space-8) + var(--space-2))"
      :with-header="false"
      class="mobile-nav-drawer"
    >
      <div class="drawer-brand">
        <div class="brand-mark">VS</div>
        <div>
          <strong>视频管理后台</strong>
          <span>分组导航</span>
        </div>
      </div>
      <nav class="drawer-nav" aria-label="移动端分组导航">
        <section v-for="group in navGroups" :key="group.key" class="drawer-nav__group">
          <div class="drawer-nav__label">{{ group.label }}</div>
          <RouterLink
            v-for="item in group.items"
            :key="item.path"
            class="drawer-nav__link"
            :class="{ 'is-active': isActive(item) }"
            :to="item.path"
            @click="closeMobileNav"
          >
            <el-icon><component :is="resolveIcon(item.icon)" /></el-icon>
            <span>{{ item.label }}</span>
          </RouterLink>
        </section>
      </nav>
    </el-drawer>

    <CommandPalette />
  </div>
</template>

<style scoped>
.admin-shell {
  --shell-current-sidebar-width: var(--admin-sidebar-width);
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-canvas);
}

.admin-shell.is-collapsed {
  --shell-current-sidebar-width: var(--admin-sidebar-collapsed-width);
}

.admin-shell.is-drawer {
  --shell-current-sidebar-width: 0;
}

.skip-link {
  position: fixed;
  top: var(--space-3);
  left: var(--space-3);
  z-index: 2000;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
  transform: translateY(calc(-1 * var(--space-12)));
  transition: transform var(--motion-duration-base) var(--motion-easing-standard);
}

.skip-link:focus {
  transform: translateY(0);
}

.shell-aside {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 60;
  display: grid;
  width: var(--shell-current-sidebar-width);
  grid-template-rows: auto minmax(0, 1fr) auto;
  border-right: 1px solid var(--line-soft);
  background: var(--bg-sidebar);
  transition: width var(--motion-duration-slow) var(--motion-easing-standard);
}

.admin-shell.is-drawer .shell-aside {
  display: none;
}

.brand-block {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
}

.brand-mark {
  display: grid;
  width: calc(var(--space-8) + var(--space-2));
  height: calc(var(--space-8) + var(--space-2));
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
  font-weight: 600;
}

.brand-copy,
.profile-chip__copy {
  display: grid;
  min-width: 0;
}

.brand-copy strong,
.profile-chip__copy strong {
  overflow: hidden;
  color: var(--text-primary);
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-copy span,
.profile-chip__copy em {
  overflow: hidden;
  color: var(--text-muted);
  font-size: var(--text-caption);
  font-style: normal;
  line-height: var(--leading-caption);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collapse-button {
  margin-left: auto;
}

.admin-shell.is-collapsed .brand-copy,
.admin-shell.is-collapsed .collapse-button,
.admin-shell.is-collapsed .nav-group__label,
.admin-shell.is-collapsed .nav-link__label,
.admin-shell.is-collapsed .profile-chip__copy {
  display: none;
}

.nav-groups {
  min-width: 0;
  overflow-y: auto;
  padding: 0 var(--space-2) var(--space-4);
}

.nav-group {
  display: grid;
  gap: var(--space-1);
  margin-top: var(--space-3);
}

.nav-group__label {
  padding: 0 var(--space-2);
  color: var(--text-muted);
  font-size: var(--text-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: var(--leading-caption);
  text-transform: uppercase;
}

.nav-link,
.drawer-nav__link {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
  font-weight: 500;
  text-decoration: none;
}

.nav-link {
  justify-content: flex-start;
  min-height: calc(var(--space-8) + var(--space-2));
  padding: 0 var(--space-3);
}

.admin-shell.is-collapsed .nav-link {
  justify-content: center;
  padding: 0;
}

.nav-link:hover,
.drawer-nav__link:hover {
  color: var(--text-primary);
  background: var(--bg-surface);
}

.nav-link.is-active,
.drawer-nav__link.is-active {
  color: var(--primary);
  background: var(--primary-soft);
}

.profile-chip {
  display: flex;
  width: calc(100% - var(--space-4));
  min-width: 0;
  align-items: center;
  gap: var(--space-3);
  margin: var(--space-2);
  padding: var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-xl);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
  cursor: pointer;
}

.profile-chip__avatar {
  display: grid;
  width: var(--space-8);
  height: var(--space-8);
  flex: 0 0 auto;
  place-items: center;
  border-radius: var(--radius-lg);
  color: var(--bg-surface);
  background: var(--primary);
  font-size: var(--text-small);
  font-weight: 600;
}

.admin-shell.is-collapsed .profile-chip {
  justify-content: center;
  width: auto;
  padding: var(--space-2) 0;
}

.shell-content {
  min-height: 100vh;
  min-height: 100dvh;
  margin-left: var(--shell-current-sidebar-width);
  transition: margin-left var(--motion-duration-slow) var(--motion-easing-standard);
}

.shell-header {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  min-height: var(--admin-header-height);
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: 0 var(--space-6);
  border-bottom: 1px solid var(--line-soft);
  background: color-mix(in srgb, var(--bg-surface) 92%, transparent);
  backdrop-filter: blur(var(--space-3));
}

.shell-header__left {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: var(--space-3);
}

.shell-page-header {
  min-width: 0;
}

.mobile-nav-btn {
  display: none;
}

.command-trigger {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xs);
  cursor: pointer;
}

.command-trigger kbd {
  padding: calc(var(--space-1) / 2) var(--space-2);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  background: var(--bg-surface-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.shell-main {
  min-width: 0;
  padding: var(--space-6);
}

.shell-main:focus-visible {
  outline: calc(var(--space-1) / 2) solid var(--primary);
  outline-offset: calc(var(--space-1) / 2);
}

.drawer-brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  border-bottom: 1px solid var(--line-soft);
}

.drawer-brand strong,
.drawer-brand span {
  display: block;
}

.drawer-brand strong {
  color: var(--text-primary);
  font-weight: 600;
}

.drawer-brand span {
  color: var(--text-muted);
  font-size: var(--text-caption);
}

.drawer-nav {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
}

.drawer-nav__group {
  display: grid;
  gap: var(--space-1);
}

.drawer-nav__label {
  color: var(--text-muted);
  font-size: var(--text-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.drawer-nav__link {
  min-height: calc(var(--space-8) + var(--space-2));
  padding: 0 var(--space-3);
}

:global(.admin-profile-popper) {
  padding: var(--space-2);
}

:global(.admin-profile-popper .profile-panel) {
  display: grid;
  gap: var(--space-2);
}

:global(.admin-profile-popper .profile-panel__name) {
  color: var(--text-primary);
  font-weight: 600;
}

:global(.admin-profile-popper .profile-panel__role) {
  display: inline-flex;
  width: fit-content;
  padding: calc(var(--space-1) / 2) var(--space-2);
  border-radius: var(--radius-sm);
  color: var(--primary);
  background: var(--primary-soft);
  font-size: var(--text-caption);
}

:global(.admin-profile-popper .profile-panel__logout) {
  justify-content: flex-start;
}

:deep(.mobile-nav-drawer .el-drawer__body) {
  padding: 0;
}

@media (max-width: 63.9375rem) {
  .mobile-nav-btn {
    display: inline-flex;
  }

  .shell-header {
    padding: 0 var(--space-4);
  }

  .shell-main {
    padding: var(--space-4);
  }
}

@media (max-width: 48rem) {
  .command-trigger span {
    display: none;
  }
}
</style>
