<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  DataAnalysis,
  Film,
  UploadFilled,
  MagicStick,
  Avatar,
  PictureFilled,
  Files,
  User,
  List,
  Setting,
  SwitchButton,
  Menu
} from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const navItems = [
  { path: '/dashboard', label: '仪表盘', title: '系统仪表盘', subtitle: '管理员工作台', icon: DataAnalysis },
  { path: '/videos', label: '视频管理', title: '视频管理', subtitle: '管理员工作台', icon: Film },
  { path: '/tv-series', label: '电视剧管理', title: '电视剧管理', subtitle: '管理员工作台', icon: Files },
  { path: '/upload', label: '上传视频', title: '上传中心', subtitle: '管理员工作台', icon: UploadFilled },
  { path: '/scrape', label: '刮削管理', title: '刮削管理', subtitle: '管理员工作台', icon: MagicStick },
  { path: '/actors', label: '演员管理', title: '演员管理', subtitle: '管理员工作台', icon: Avatar },
  { path: '/collections', label: '合集管理', title: '合集管理', subtitle: '管理员工作台', icon: List },
  { path: '/images', label: '图片管理', title: '图片管理', subtitle: '管理员工作台', icon: PictureFilled },
  { path: '/image-collections', label: '图片合集', title: '图片合集', subtitle: '管理员工作台', icon: Files },
  { path: '/users', label: '用户管理', title: '用户管理', subtitle: '管理员工作台', icon: User },
  { path: '/tasks', label: '任务监控', title: '任务监控', subtitle: '管理员工作台', icon: List },
  { path: '/settings', label: '系统设置', title: '系统设置', subtitle: '管理员工作台', icon: Setting }
]

const mobileNavVisible = ref(false)
const shellContentRef = ref(null)
const mainContentRef = ref(null)

const sortedNavItems = [...navItems].sort((a, b) => b.path.length - a.path.length)

function findMatchedNavItem(path) {
  return sortedNavItems.find((item) => path === item.path || path.startsWith(`${item.path}/`))
}

const matchedNavItem = computed(() => findMatchedNavItem(route.path))
const active = computed(() => matchedNavItem.value?.path || route.path)
const pageMeta = computed(
  () =>
    matchedNavItem.value || {
      title: route.meta?.title || '管理后台',
      subtitle: '管理员工作台'
    }
)

watch(
  () => route.fullPath,
  () => {
    mobileNavVisible.value = false
  }
)

const restoreOverflow = { html: '', body: '' }
watch(mobileNavVisible, (visible) => {
  if (typeof window === 'undefined') {
    return
  }
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

onUnmounted(() => {
  if (typeof window === 'undefined') {
    return
  }
  document.documentElement.style.overflow = restoreOverflow.html
  document.body.style.overflow = restoreOverflow.body
})

function openMobileNav() {
  mobileNavVisible.value = true
}

function onMobileNavSelect() {
  mobileNavVisible.value = false
}

function onSkipLinkClick() {
  shellContentRef.value?.scrollTo({ top: 0, behavior: 'auto' })
  mainContentRef.value?.focus({ preventScroll: true })
}

async function onLogout() {
  await auth.logout()
  mobileNavVisible.value = false
  router.push('/login')
}
</script>

<template>
  <div class="admin-shell" :class="{ 'is-mobile-nav-open': mobileNavVisible }">
    <a class="skip-link" href="#main-content" @click.prevent="onSkipLinkClick">跳转到主要内容</a>

    <aside class="shell-aside" aria-label="主导航">
      <div class="brand-block">
        <div class="brand-badge">AV</div>
        <div>
          <div class="brand-title">视频管理后台</div>
          <div class="brand-subtitle">本地视频服务</div>
        </div>
      </div>

      <el-menu :default-active="active" router class="nav-menu">
        <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <section ref="shellContentRef" class="shell-content">
      <header class="shell-header">
        <div class="shell-header-left">
          <el-button class="mobile-nav-btn" text :icon="Menu" aria-label="打开导航菜单" @click="openMobileNav" />
          <div>
            <div class="shell-header-title">{{ pageMeta.title }}</div>
            <div class="shell-header-subtitle">{{ pageMeta.subtitle }}</div>
          </div>
        </div>
        <el-button plain type="danger" :icon="SwitchButton" @click="onLogout">退出登录</el-button>
      </header>

      <main id="main-content" ref="mainContentRef" class="shell-main" tabindex="-1">
        <slot />
      </main>
    </section>

    <el-drawer v-model="mobileNavVisible" direction="ltr" size="260px" :with-header="false" class="mobile-nav-drawer">
      <div class="brand-block brand-block--drawer">
        <div class="brand-badge">AV</div>
        <div>
          <div class="brand-title">视频管理后台</div>
          <div class="brand-subtitle">本地视频服务</div>
        </div>
      </div>

      <el-menu :default-active="active" router class="nav-menu nav-menu--drawer" @select="onMobileNavSelect">
        <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </el-drawer>
  </div>
</template>

<style scoped>
.admin-shell {
  --shell-bg: #f8fafc;
  --shell-title: #7f1d1d;
  --shell-subtitle: #6b7280;
  --shell-border: rgba(136, 19, 55, 0.12);
  min-height: 100vh;
  min-height: 100dvh;
  height: 100dvh;
  background: #f8fafc;
}

.skip-link {
  position: fixed;
  top: 10px;
  left: 10px;
  z-index: 2000;
  padding: 8px 12px;
  border-radius: 8px;
  color: #7f1d1d;
  background: #fff;
  border: 1px solid rgba(127, 29, 29, 0.28);
  transform: translateY(-160%);
  transition: transform 0.2s ease;
}

.skip-link:focus {
  transform: translateY(0);
}

.shell-aside {
  position: fixed;
  top: 0;
  left: 0;
  width: 240px;
  min-height: 100vh;
  min-height: 100dvh;
  height: 100dvh;
  overflow-y: auto;
  background: linear-gradient(180deg, #881337 0%, #be123c 45%, #1e3a8a 100%);
  border-right: 1px solid rgba(255, 255, 255, 0.15);
}

.shell-content {
  margin-left: 240px;
  min-height: 100vh;
  min-height: 100dvh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.admin-shell.is-mobile-nav-open .shell-content {
  overflow: hidden;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 16px 14px;
  color: #fff;
}

.brand-block--drawer {
  background: linear-gradient(180deg, #881337 0%, #be123c 45%, #1e3a8a 100%);
}

.brand-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-size: 14px;
  font-family: 'Fira Code', monospace;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.22);
  border: 1px solid rgba(255, 255, 255, 0.42);
}

.brand-title {
  font-weight: 700;
  letter-spacing: 0.2px;
}

.brand-subtitle {
  margin-top: 2px;
  font-size: 12px;
  opacity: 0.88;
}

.nav-menu {
  background: transparent;
  border-right: none;
}

:deep(.nav-menu.el-menu .el-menu-item) {
  margin: 4px 10px;
  border-radius: 10px;
  color: rgba(255, 255, 255, 0.8);
}

:deep(.nav-menu.el-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

:deep(.nav-menu.el-menu .el-menu-item.is-active) {
  color: #fff;
  background: rgba(255, 255, 255, 0.24);
}

.shell-header {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 64px;
  padding: 0 20px;
  border-bottom: 1px solid var(--shell-border);
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(8px);
}

.shell-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mobile-nav-btn {
  display: none;
  font-size: 18px;
}

.shell-header-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--shell-title);
}

.shell-header-subtitle {
  margin-top: 2px;
  font-size: 12px;
  color: var(--shell-subtitle);
}

.shell-main {
  flex: 1;
  padding: 18px;
  scroll-margin-top: 76px;
}

.shell-main:focus-visible {
  outline: 2px solid var(--shell-title);
  outline-offset: 2px;
}

.nav-menu--drawer {
  padding-top: 12px;
  border-right: none;
}

:deep(.nav-menu--drawer.el-menu .el-menu-item) {
  color: #374151;
}

:deep(.nav-menu--drawer.el-menu .el-menu-item:hover) {
  color: #111827;
  background: #f3f4f6;
}

:deep(.nav-menu--drawer.el-menu .el-menu-item.is-active) {
  color: #7f1d1d;
  background: rgba(136, 19, 55, 0.12);
}

:deep(.mobile-nav-drawer .el-drawer__body) {
  padding: 0;
}

@media (max-width: 992px) {
  .shell-aside {
    display: none;
  }

  .shell-content {
    margin-left: 0;
  }

  .shell-header {
    padding: 0 14px;
  }

  .mobile-nav-btn {
    display: inline-flex;
  }

  .shell-main {
    padding: 14px;
  }
}
</style>
