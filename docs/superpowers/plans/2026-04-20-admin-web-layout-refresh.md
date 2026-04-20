# Admin Web Layout Refresh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the admin web shell so the sidebar stays fixed on desktop, switches to a drawer on mobile, and unify the full admin interface under a tighter editorial dashboard visual system.

**Architecture:** Keep all routing and business API logic intact, and concentrate the behavior change in the shared shell and global theme layer. Use `Layout.vue` to own desktop/mobile navigation and scrolling behavior, use `theme.css` to define the visual system and common page primitives, then update each view template to consume those primitives without changing data flow.

**Tech Stack:** Vue 3 SFCs, Vue Router, Element Plus, Vite, shared CSS theme variables in `admin-web/src/assets/theme.css`

---

## File Structure

- `admin-web/src/components/Layout.vue`
  - Own the fixed desktop sidebar, sticky content header, mobile drawer navigation, skip link, and shell-level scroll container.
- `admin-web/src/assets/theme.css`
  - Own the design tokens, Element Plus theme overrides, and shared page primitives such as `page-shell`, `page-section`, `stats-grid`, `toolbar-row`, and `table-panel`.
- `admin-web/src/views/Dashboard.vue`
  - Act as the reference page for the new visual language.
- `admin-web/src/views/VideoList.vue`
  - Adopt the new toolbar and results panel structure.
- `admin-web/src/views/VideoUpload.vue`
  - Turn the upload flow into a clearer primary task page with grouped sections.
- `admin-web/src/views/TaskMonitor.vue`
  - Strengthen status readability and panel structure for live monitoring.
- `admin-web/src/views/ActorManage.vue`
- `admin-web/src/views/CollectionManage.vue`
- `admin-web/src/views/ImageManage.vue`
- `admin-web/src/views/ImageCollectionManage.vue`
- `admin-web/src/views/UserManage.vue`
- `admin-web/src/views/SystemSettings.vue`
- `admin-web/src/views/ScrapePreview.vue`
- `admin-web/src/views/Login.vue`
  - Align all remaining pages with the new shell, section, table, filter, and dialog styling.
- `plan.md`
  - Append implementation entries only; do not overwrite prior history.

## Constraints and Defaults

- Do not introduce new backend APIs or route changes.
- Do not add a frontend test runner in this task. `admin-web/package.json` has no test script or test dependencies, so verification stays on `npm run build` plus explicit manual QA.
- Keep Chinese copy intact unless a label must be adjusted to fit the new layout.
- Prefer global theme classes over page-local one-off CSS where the pattern repeats across pages.

### Task 1: Rebuild the Shared Shell

**Files:**
- Modify: `admin-web/src/components/Layout.vue`
- Verify: `admin-web/src/components/Layout.vue`

- [ ] **Step 1: Replace the shell script block with explicit nav state and route metadata**

```vue
<script setup>
import { computed, ref, watch } from 'vue'
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
  Menu,
  Monitor
} from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const mobileNavVisible = ref(false)

const navItems = [
  { index: '/dashboard', label: '仪表盘', icon: DataAnalysis, tone: 'overview' },
  { index: '/videos', label: '视频管理', icon: Film, tone: 'content' },
  { index: '/upload', label: '上传中心', icon: UploadFilled, tone: 'content' },
  { index: '/scrape', label: '刮削管理', icon: MagicStick, tone: 'ops' },
  { index: '/actors', label: '演员管理', icon: Avatar, tone: 'library' },
  { index: '/collections', label: '合集管理', icon: List, tone: 'library' },
  { index: '/images', label: '图片管理', icon: PictureFilled, tone: 'library' },
  { index: '/image-collections', label: '图片合集', icon: Files, tone: 'library' },
  { index: '/users', label: '用户管理', icon: User, tone: 'ops' },
  { index: '/tasks', label: '任务监控', icon: Monitor, tone: 'ops' },
  { index: '/settings', label: '系统设置', icon: Setting, tone: 'ops' }
]

const active = computed(() => route.path)
const pageMeta = computed(() => {
  const map = {
    '/dashboard': { title: '系统仪表盘', subtitle: '内容、用户与系统队列的统一视图' },
    '/videos': { title: '视频管理', subtitle: '筛选、编辑、预览与流程维护' },
    '/upload': { title: '上传中心', subtitle: '主任务流、进度与结果集中处理' },
    '/scrape': { title: '刮削管理', subtitle: '候选预览、元数据比对与确认保存' },
    '/actors': { title: '演员管理', subtitle: '演员主数据、别名与资料维护' },
    '/collections': { title: '合集管理', subtitle: '内容合集结构和启停状态管理' },
    '/images': { title: '图片管理', subtitle: '上传、归档、关联与预览查看' },
    '/image-collections': { title: '图片合集', subtitle: '图片归档集合与归属关系维护' },
    '/users': { title: '用户管理', subtitle: '账户角色与后台权限控制' },
    '/tasks': { title: '任务监控', subtitle: '转码任务实时状态与进度观测' },
    '/settings': { title: '系统设置', subtitle: '清理操作与系统日志查看' }
  }
  return map[route.path] || { title: '管理后台', subtitle: '后台控制台' }
})

watch(
  () => route.fullPath,
  () => {
    mobileNavVisible.value = false
  }
)

async function onLogout() {
  await auth.logout()
  router.push('/login')
}
</script>
```

- [ ] **Step 2: Replace the template with a fixed sidebar + independent scroll content shell**

```vue
<template>
  <div class="admin-shell">
    <a class="skip-link" href="#admin-main">跳到主内容</a>

    <aside class="shell-sidebar">
      <div class="brand-panel">
        <div class="brand-mark">AV</div>
        <div>
          <div class="brand-title">视频管理后台</div>
          <div class="brand-subtitle">Editorial Control Console</div>
        </div>
      </div>

      <div class="nav-section-label">Navigation</div>
      <el-menu :default-active="active" router class="nav-menu">
        <el-menu-item v-for="item in navItems" :key="item.index" :index="item.index" :class="`tone-${item.tone}`">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <div class="sidebar-footer-label">当前页面</div>
        <div class="sidebar-footer-value">{{ pageMeta.title }}</div>
      </div>
    </aside>

    <div class="shell-mobile-topbar">
      <button class="mobile-nav-trigger" type="button" @click="mobileNavVisible = true" aria-label="打开导航菜单">
        <el-icon><Menu /></el-icon>
      </button>
      <div>
        <div class="mobile-brand-title">视频管理后台</div>
        <div class="mobile-brand-subtitle">{{ pageMeta.title }}</div>
      </div>
      <el-button plain type="danger" :icon="SwitchButton" @click="onLogout">退出</el-button>
    </div>

    <el-drawer v-model="mobileNavVisible" direction="ltr" size="280px" :with-header="false" class="mobile-nav-drawer">
      <div class="drawer-inner">
        <div class="brand-panel brand-panel-drawer">
          <div class="brand-mark">AV</div>
          <div>
            <div class="brand-title">视频管理后台</div>
            <div class="brand-subtitle">Editorial Control Console</div>
          </div>
        </div>
        <el-menu :default-active="active" router class="nav-menu nav-menu-mobile">
          <el-menu-item v-for="item in navItems" :key="item.index" :index="item.index">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </el-menu-item>
        </el-menu>
      </div>
    </el-drawer>

    <div class="shell-content">
      <header class="shell-header">
        <div>
          <div class="shell-header-eyebrow">Admin Workspace</div>
          <div class="shell-header-title">{{ pageMeta.title }}</div>
          <div class="shell-header-subtitle">{{ pageMeta.subtitle }}</div>
        </div>
        <el-button plain type="danger" :icon="SwitchButton" class="desktop-logout" @click="onLogout">退出登录</el-button>
      </header>

      <main id="admin-main" class="shell-main">
        <slot />
      </main>
    </div>
  </div>
</template>
```

- [ ] **Step 3: Replace the scoped styles so desktop stays fixed and mobile switches to drawer mode**

```css
<style scoped>
.admin-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  background: var(--shell-bg);
}

.skip-link {
  position: fixed;
  top: 14px;
  left: 16px;
  z-index: 90;
  transform: translateY(-180%);
}

.skip-link:focus {
  transform: translateY(0);
}

.shell-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 22px 16px 18px;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--sidebar-border);
}

.shell-content {
  min-width: 0;
  height: 100vh;
  overflow: auto;
}

.shell-header {
  position: sticky;
  top: 0;
  z-index: 30;
}

.shell-mobile-topbar {
  display: none;
}

@media (max-width: 1024px) {
  .admin-shell {
    display: block;
  }

  .shell-sidebar,
  .desktop-logout {
    display: none;
  }

  .shell-mobile-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .shell-content {
    height: auto;
    min-height: 100vh;
    overflow: visible;
  }
}
</style>
```

- [ ] **Step 4: Run the build to verify the new shell compiles before page work starts**

Run: `cd admin-web && npm run build`

Expected: `vite build` completes with `✓ built in ...` and no Vue template or CSS compilation errors.

- [ ] **Step 5: Commit the shell-only checkpoint**

```bash
git add admin-web/src/components/Layout.vue
git commit -m "feat: 重构管理端共享壳层"
```

### Task 2: Build the Global Visual System and Shared Page Primitives

**Files:**
- Modify: `admin-web/src/assets/theme.css`
- Verify: `admin-web/src/assets/theme.css`

- [ ] **Step 1: Replace the root tokens with a tighter editorial dashboard palette**

```css
:root {
  --shell-bg: #eef2f6;
  --shell-ambient: radial-gradient(circle at top left, rgba(236, 72, 153, 0.08), transparent 30%),
    radial-gradient(circle at top right, rgba(15, 23, 42, 0.08), transparent 28%),
    linear-gradient(180deg, #f8fafc 0%, #eef2f6 100%);
  --sidebar-bg: linear-gradient(180deg, #09090b 0%, #111827 48%, #1f2937 100%);
  --sidebar-border: rgba(255, 255, 255, 0.08);
  --surface-1: rgba(255, 255, 255, 0.92);
  --surface-2: #ffffff;
  --surface-3: #f8fafc;
  --border-soft: rgba(15, 23, 42, 0.08);
  --border-strong: rgba(15, 23, 42, 0.14);
  --text-main: #0f172a;
  --text-subtle: #475569;
  --text-soft: #64748b;
  --accent: #ec4899;
  --accent-strong: #db2777;
  --accent-soft: rgba(236, 72, 153, 0.12);
  --shadow-panel: 0 18px 42px rgba(15, 23, 42, 0.08);
  --shadow-hover: 0 22px 48px rgba(15, 23, 42, 0.12);
  --radius-xl: 24px;
  --radius-lg: 18px;
  --radius-md: 14px;
}
```

- [ ] **Step 2: Add shared shell/page primitives that all views can consume**

```css
.page-shell {
  display: grid;
  gap: 18px;
}

.page-section {
  display: grid;
  gap: 14px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.toolbar-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 12px;
}

.content-card,
.table-panel,
.soft-card.el-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-soft);
  background: var(--surface-1);
  box-shadow: var(--shadow-panel);
  backdrop-filter: blur(10px);
}
```

- [ ] **Step 3: Add consistent Element Plus overrides for tables, forms, buttons, dialogs, and drawers**

```css
.el-table {
  --el-table-header-bg-color: #f8fafc;
  --el-table-tr-bg-color: rgba(255, 255, 255, 0.8);
  --el-table-row-hover-bg-color: #fff7fb;
}

.el-button--primary {
  --el-button-bg-color: var(--accent);
  --el-button-border-color: var(--accent);
  --el-button-hover-bg-color: var(--accent-strong);
  --el-button-hover-border-color: var(--accent-strong);
}

.el-input__wrapper,
.el-select__wrapper,
.el-textarea__inner,
.el-dialog,
.el-drawer {
  border-radius: var(--radius-md);
}
```

- [ ] **Step 4: Add responsive rules for stats grids, page headers, and toolbars so mobile remains usable**

```css
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .stats-grid,
  .page-header,
  .section-head {
    grid-template-columns: 1fr;
    display: grid;
  }

  .toolbar-row,
  .filter-form {
    display: grid;
  }
}
```

- [ ] **Step 5: Run the build after the global theme rewrite**

Run: `cd admin-web && npm run build`

Expected: `vite build` completes with `✓ built in ...` and no CSS variable or selector errors.

- [ ] **Step 6: Commit the global visual system checkpoint**

```bash
git add admin-web/src/assets/theme.css
git commit -m "feat: 统一管理端全局视觉系统"
```

### Task 3: Refresh the Reference Pages and Primary Workflows

**Files:**
- Modify: `admin-web/src/views/Dashboard.vue`
- Modify: `admin-web/src/views/VideoList.vue`
- Modify: `admin-web/src/views/VideoUpload.vue`
- Modify: `admin-web/src/views/TaskMonitor.vue`

- [ ] **Step 1: Rebuild `Dashboard.vue` into the reference page for the new system**

```vue
<template>
  <Layout>
    <div class="page page-shell dashboard-page">
      <section class="page-header page-section">
        <div>
          <h1 class="page-title">系统概览</h1>
          <p class="page-subtitle">视频、用户、任务与存储状态一屏查看</p>
        </div>
      </section>

      <section class="stats-grid stats-grid-primary">
        <el-card v-for="item in summaryCards" :key="item.label" class="soft-card stat-card">
          <div class="stat-card-label">{{ item.label }}</div>
          <div class="stat-card-value">{{ item.value }}</div>
          <div class="stat-card-note">{{ item.note }}</div>
        </el-card>
      </section>

      <section class="page-section dashboard-grid">
        <el-card class="soft-card trend-card">
          <template #header>近 7 天上传趋势</template>
          <div ref="chartRef" class="trend-chart" />
        </el-card>
      </section>
    </div>
  </Layout>
</template>
```

- [ ] **Step 2: Convert `VideoList.vue` to the new toolbar + table panel structure without changing script behavior**

```vue
<template>
  <Layout>
    <div class="page page-shell">
      <div class="page-header">
        <div>
          <h1 class="page-title">视频管理</h1>
          <p class="page-subtitle">筛选、查看、重转码与元数据编辑</p>
        </div>
      </div>

      <el-card class="soft-card content-card">
        <el-form inline class="filter-form toolbar-row">
          <!-- 保留现有筛选控件 -->
        </el-form>
      </el-card>

      <el-card class="soft-card table-panel">
        <div class="table-wrap">
          <el-table :data="list" border>
            <!-- 保留现有列定义 -->
          </el-table>
        </div>
        <div class="action-row">
          <el-pagination v-model:current-page="query.page" v-model:page-size="query.page_size" :total="total" @current-change="load" />
        </div>
      </el-card>
    </div>
  </Layout>
</template>
```

- [ ] **Step 3: Turn `VideoUpload.vue` into grouped workflow sections and preserve the existing upload logic**

```vue
<template>
  <Layout>
    <div class="page page-shell upload-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">上传中心</h1>
          <p class="page-subtitle">支持分片上传、秒传检测与上传取消</p>
        </div>
      </div>

      <el-card class="soft-card content-card upload-stage-card">
        <!-- 文件选择与核心表单 -->
      </el-card>

      <div class="upload-grid">
        <el-card class="soft-card content-card progress-card">
          <template #header>上传进度</template>
          <UploadProgress :percentage="progress" :status-text="`当前文件：${currentFileName || '-'} ｜ 哈希计算 ${hashProgress}%`" />
        </el-card>

        <el-card class="soft-card content-card result-card">
          <template #header>批次结果</template>
          <!-- 保留现有表格和统计 -->
        </el-card>
      </div>
    </div>
  </Layout>
</template>
```

- [ ] **Step 4: Update `TaskMonitor.vue` to match the monitoring page treatment**

```vue
<template>
  <Layout>
    <div class="page page-shell task-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">任务监控</h1>
          <p class="page-subtitle">每 5 秒自动刷新转码任务状态</p>
        </div>
      </div>

      <el-card class="soft-card table-panel">
        <template #header>
          <div class="section-head">
            <span>任务队列</span>
            <span class="panel-note">自动刷新中</span>
          </div>
        </template>
        <div class="table-wrap">
          <el-table :data="list" border>
            <!-- 保留现有列定义 -->
          </el-table>
        </div>
      </el-card>
    </div>
  </Layout>
</template>
```

- [ ] **Step 5: Run the build after the primary workflow pages are updated**

Run: `cd admin-web && npm run build`

Expected: `vite build` completes with `✓ built in ...` and no template mismatch errors.

- [ ] **Step 6: Commit the reference page checkpoint**

```bash
git add admin-web/src/views/Dashboard.vue admin-web/src/views/VideoList.vue admin-web/src/views/VideoUpload.vue admin-web/src/views/TaskMonitor.vue
git commit -m "feat: 升级管理端核心页面骨架"
```

### Task 4: Apply the New System to the Remaining Admin Pages

**Files:**
- Modify: `admin-web/src/views/ActorManage.vue`
- Modify: `admin-web/src/views/CollectionManage.vue`
- Modify: `admin-web/src/views/ImageManage.vue`
- Modify: `admin-web/src/views/ImageCollectionManage.vue`
- Modify: `admin-web/src/views/UserManage.vue`
- Modify: `admin-web/src/views/SystemSettings.vue`
- Modify: `admin-web/src/views/ScrapePreview.vue`
- Modify: `admin-web/src/views/Login.vue`

- [ ] **Step 1: Normalize the CRUD pages to the shared page-shell + toolbar + table-panel structure**

```vue
<div class="page page-shell">
  <div class="page-header">
    <div>
      <h1 class="page-title">页面标题</h1>
      <p class="page-subtitle">页面说明</p>
    </div>
    <el-button type="primary">主操作</el-button>
  </div>

  <el-card class="soft-card content-card">
    <el-form inline class="filter-form toolbar-row">
      <!-- 保留现有筛选和搜索控件 -->
    </el-form>
  </el-card>

  <el-card class="soft-card table-panel">
    <div class="table-wrap">
      <el-table :data="list" border>
        <!-- 保留现有列 -->
      </el-table>
    </div>
    <div class="action-row">
      <el-pagination ... />
    </div>
  </el-card>
</div>
```

Apply this pattern to:
- `ActorManage.vue`
- `CollectionManage.vue`
- `ImageManage.vue`
- `ImageCollectionManage.vue`
- `UserManage.vue`

- [ ] **Step 2: Refresh `SystemSettings.vue` so logs and maintenance actions read like a control center**

```vue
<template>
  <Layout>
    <div class="page page-shell settings-page">
      <div class="page-header">
        <div>
          <h1 class="page-title">系统设置</h1>
          <p class="page-subtitle">执行清理任务并查看近期系统日志</p>
        </div>
      </div>

      <el-card class="soft-card content-card settings-actions-card">
        <div class="toolbar-row">
          <el-button type="warning" @click="runCleanup">清理临时文件</el-button>
          <el-button @click="loadLogs" :loading="loading">刷新日志</el-button>
        </div>
      </el-card>

      <el-card class="soft-card content-card settings-log-card">
        <template #header>系统日志</template>
        <el-scrollbar height="520px" class="log-box">
          <pre class="log-text">{{ logs.join('\n') }}</pre>
        </el-scrollbar>
      </el-card>
    </div>
  </Layout>
</template>
```

- [ ] **Step 3: Refresh `ScrapePreview.vue` and `Login.vue` so they visually belong to the same system**

```vue
<!-- ScrapePreview.vue -->
<div class="page page-shell scrape-page">
  <div class="page-header">
    <div>
      <h1 class="page-title">刮削管理</h1>
      <p class="page-subtitle">预览候选详情并确认保存，支持完整 metadata 检视</p>
    </div>
  </div>
  <!-- 保留现有功能卡片，但统一 section/card class -->
</div>

<!-- Login.vue -->
<div class="login-page editorial-login">
  <div class="login-grid">
    <section class="hero-panel hero-panel-editorial">
      <h1 class="hero-title">家用视频服务器</h1>
      <p class="hero-subtitle">管理内容、用户与系统任务的统一控制台</p>
    </section>
    <el-card class="soft-card login-card login-card-editorial">
      <!-- 保留现有登录表单 -->
    </el-card>
  </div>
</div>
```

- [ ] **Step 4: Run the full build after all remaining pages adopt the new system**

Run: `cd admin-web && npm run build`

Expected: `vite build` completes with `✓ built in ...` and no route view compilation errors.

- [ ] **Step 5: Commit the remaining page refresh checkpoint**

```bash
git add admin-web/src/views/ActorManage.vue admin-web/src/views/CollectionManage.vue admin-web/src/views/ImageManage.vue admin-web/src/views/ImageCollectionManage.vue admin-web/src/views/UserManage.vue admin-web/src/views/SystemSettings.vue admin-web/src/views/ScrapePreview.vue admin-web/src/views/Login.vue
git commit -m "feat: 统一管理端页面视觉风格"
```

### Task 5: Manual QA, Documentation, and Final Verification

**Files:**
- Modify: `plan.md`
- Verify: `admin-web/src/components/Layout.vue`
- Verify: `admin-web/src/assets/theme.css`
- Verify: `admin-web/src/views/*.vue`

- [ ] **Step 1: Run the final build on the finished branch**

Run: `cd admin-web && npm run build`

Expected: `vite build` completes with `✓ built in ...`. The existing large chunk warning may remain, but there must be no compile failures.

- [ ] **Step 2: Perform the explicit manual QA pass in the browser**

Use this checklist and do not skip any item:

```text
Desktop 1440px:
- Sidebar stays fixed while scrolling VideoList, ImageManage, and VideoUpload.
- Header stays sticky only inside the content column.
- Active menu item remains highlighted after route changes.

Tablet 768px:
- Sidebar is hidden.
- Top mobile bar is visible.
- Drawer opens, route switch works, drawer closes after navigation.

Mobile 375px:
- Page headers wrap cleanly.
- Filter forms stack vertically.
- Tables remain contained within card panels without shell breakage.

Visual consistency:
- Dashboard cards, upload panels, monitor panels, and CRUD tables share the same surface and spacing language.
- Login page matches the editorial palette and typography of the admin shell.
```

- [ ] **Step 3: Append the implementation record to `plan.md`**

```md
### [YYYY-MM-DD HH:MM] 管理端固定侧栏与全站视觉提质
- Type: `implementation`
- Summary:
  - 重构后台共享壳层，实现桌面端固定侧栏、移动端抽屉导航与内容区独立滚动。
  - 重写全局主题变量与通用页面骨架，统一后台卡片、筛选区、表格区、按钮和弹窗视觉语言。
  - 升级 Dashboard、VideoList、VideoUpload、TaskMonitor 以及其余管理页与登录页，使其接入统一的后台视觉系统。
- Changed Files:
  - `admin-web/src/components/Layout.vue`
  - `admin-web/src/assets/theme.css`
  - `admin-web/src/views/Dashboard.vue`
  - `admin-web/src/views/VideoList.vue`
  - `admin-web/src/views/VideoUpload.vue`
  - `admin-web/src/views/TaskMonitor.vue`
  - `admin-web/src/views/ActorManage.vue`
  - `admin-web/src/views/CollectionManage.vue`
  - `admin-web/src/views/ImageManage.vue`
  - `admin-web/src/views/ImageCollectionManage.vue`
  - `admin-web/src/views/UserManage.vue`
  - `admin-web/src/views/SystemSettings.vue`
  - `admin-web/src/views/ScrapePreview.vue`
  - `admin-web/src/views/Login.vue`
  - `plan.md`
- Verification:
  - `cd admin-web && npm run build` passed.
  - `桌面/平板/移动端手工验收` passed.
- Rollback:
  - `git revert <commit>`
```

- [ ] **Step 4: Commit the final documentation entry**

```bash
git add plan.md
git commit -m "docs: 记录管理端布局与视觉提质实施"
```

## Self-Review Checklist

- Spec coverage:
  - Fixed sidebar: Task 1
  - Mobile drawer navigation: Task 1
  - Global visual system: Task 2
  - Reference page upgrade: Task 3
  - Remaining page alignment: Task 4
  - Build + manual acceptance: Task 5
- Placeholder scan:
  - The only bracketed placeholders are the final `plan.md` timestamp and `<commit>` rollback marker, both intentionally matching repository conventions.
- Type consistency:
  - `mobileNavVisible`, `navItems`, `page-shell`, `content-card`, `table-panel`, and `toolbar-row` are referenced consistently across tasks.
