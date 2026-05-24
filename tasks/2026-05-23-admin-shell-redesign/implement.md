# Implement：admin 重设计 Phase 1 · 设计系统与外壳

- 日期：2026-05-23
- 关联 PRD：`prd.md`
- 任务三段执行流：当前阶段 = Implement

## 1. 实施顺序（TDD 友好）

1. **token 层先红后绿**：写 `themeTokens.spec.js` → 重写 `theme.css` → 跑 spec 绿
2. **Element Plus L2 覆写**：新增 `element-overrides.css` → `main.js` import → `npm run build` 通过
3. **wrapper 组件骨架**：7 个 `components/base/*.vue` + `commandPalette.helpers.js` + spec → 跑 spec 绿
4. **Layout 外壳重写**：`Layout.vue` + `Layout.spec.js` 源文 audit → 跑 spec 绿
5. **Login 重写**：套用新 token + PageHeader 范式
6. **截图 + 手测**：按 PRD H1–H15 子集走一遍
7. **提交**：一次 commit + plan.md 追加条目

## 2. 文件改动清单

### 2.1 `admin-web/src/assets/theme.css`（全文重写）

骨架（不含全部 token 列表，详见 Q3 / Q6 / Q8）：

```css
/* 字体引入 —— 仅 Inter 三档；删除 Fira Code 全部 weight */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap');

:root {
  /* ── 颜色 · slate ramp ─────────────────────────── */
  --slate-50: #f8fafc; --slate-100: #f1f5f9; /* ... */ --slate-900: #0f172a;

  /* ── 颜色 · blue ramp（主色） ─────────────────── */
  --blue-50: #eff6ff; /* ... */ --blue-600: #2563eb; /* ... */ --blue-900: #1e3a8a;

  /* ── 语义色（仅 600 + 50 两档） ──────────────── */
  --success-600: #059669; --success-50: #ecfdf5;
  --warning-600: #d97706; --warning-50: #fffbeb;
  --danger-600:  #dc2626; --danger-50:  #fef2f2;
  --info-600:    #0284c7; --info-50:    #f0f9ff;

  /* ── 语义别名 ──────────────────────────────── */
  --primary:        var(--blue-600);
  --primary-strong: var(--blue-700);
  --primary-soft:   var(--blue-50);
  --primary-ring:   color-mix(in srgb, var(--blue-600) 36%, transparent);

  --bg-canvas:        var(--slate-50);
  --bg-surface:       #ffffff;
  --bg-surface-muted: var(--slate-100);
  --bg-sidebar:       var(--slate-100);     /* Q9 决策 */

  --text-primary:   var(--slate-900);
  --text-secondary: var(--slate-600);
  --text-muted:     var(--slate-400);

  --line-soft:   var(--slate-200);
  --line-strong: var(--slate-300);
  --line-focus:  var(--primary-ring);

  /* ── 字体（Q6） ───────────────────────────── */
  --font-sans: 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, -apple-system, sans-serif;
  --font-mono: ui-monospace, 'SF Mono', 'Cascadia Code', Consolas, monospace;

  /* ── 字号（Q6 / 6 档） ───────────────────── */
  --text-display: 28px; --leading-display: 36px;
  --text-h1:      20px; --leading-h1:      28px;
  --text-h2:      15px; --leading-h2:      22px;
  --text-body:    14px; --leading-body:    22px;
  --text-small:   13px; --leading-small:   20px;
  --text-caption: 11px; --leading-caption: 16px;

  /* ── 圆角 / 阴影 / 间距 / 动效（Q8） ─────── */
  --radius-xs:  4px;  --radius-sm:  6px;  --radius-md: 8px;
  --radius-lg: 10px;  --radius-xl: 12px;  --radius-2xl: 16px;
  --shadow-xs: 0 1px 2px rgba(15, 23, 42, 0.04);
  --shadow-sm: 0 1px 3px rgba(15, 23, 42, 0.06), 0 1px 2px rgba(15, 23, 42, 0.04);
  --shadow-md: 0 4px 12px rgba(15, 23, 42, 0.08), 0 2px 4px rgba(15, 23, 42, 0.04);
  --shadow-lg: 0 12px 32px rgba(15, 23, 42, 0.10), 0 4px 8px rgba(15, 23, 42, 0.06);
  --space-1:4px; --space-2:8px; --space-3:12px; --space-4:16px;
  --space-5:20px; --space-6:24px; --space-8:32px; --space-12:48px;
  --motion-duration-fast: 120ms;
  --motion-duration-base: 180ms;
  --motion-duration-slow: 240ms;
  --motion-easing-standard:   cubic-bezier(0.2, 0, 0, 1);
  --motion-easing-accelerate: cubic-bezier(0.4, 0, 1, 1);

  /* ── 断点暴露给 JS 用（Q11） ─────────────── */
  --bp-sm: 768px;
  --bp-md: 1024px;
  --bp-lg: 1280px;
  --bp-xl: 1536px;

  /* ── Element Plus token 桥接 ─────────────── */
  --el-color-primary: var(--primary);
  --el-color-primary-light-9: var(--primary-soft);
  --el-color-primary-dark-2:  var(--primary-strong);
  --el-color-success: var(--success-600);
  --el-color-warning: var(--warning-600);
  --el-color-danger:  var(--danger-600);
  --el-color-info:    var(--info-600);
  --el-border-color:       var(--line-soft);
  --el-border-color-light: var(--line-soft);
  --el-text-color-primary:   var(--text-primary);
  --el-text-color-regular:   var(--text-secondary);
  --el-text-color-secondary: var(--text-muted);
  --el-fill-color-light:     var(--bg-surface-muted);
  --el-bg-color-overlay:     var(--bg-surface);
  --el-border-radius-base:   var(--radius-md);
  --el-font-family:          var(--font-sans);
}

/* dark 占位（Phase 1 不接 UI 切换；手动 `<html data-theme="dark">` 验证） */
:root[data-theme="dark"] {
  --bg-canvas:        var(--slate-900);
  --bg-surface:       var(--slate-800);
  --bg-surface-muted: var(--slate-700);
  --bg-sidebar:       var(--slate-950, #020617);
  --text-primary:     var(--slate-50);
  --text-secondary:   var(--slate-300);
  --text-muted:       var(--slate-400);
  --line-soft:        var(--slate-700);
  --line-strong:      var(--slate-600);
}

html, body, #app { margin: 0; min-height: 100%; }
:root { font-size: var(--text-body); }
body {
  font-family: var(--font-sans);
  color: var(--text-primary);
  background: var(--bg-canvas);
  font-feature-settings: 'cv11' on, 'ss03' on;
}
.tabular-num { font-variant-numeric: tabular-nums; }
```

**严格删除**：

- ❌ `@import ... Fira+Code` 与 `@import ... Fira+Sans`
- ❌ `--font-code: 'Fira Code', ...`
- ❌ 任何裸 `#881337` / `#7f1d1d` / `#be123c` 色值
- ❌ 多套 radial-gradient body 背景（保留单色 `--bg-canvas`）

### 2.2 `admin-web/src/assets/element-overrides.css`（新增 ≤ 400 行）

```css
/* :where() 包裹保持选择器低权重 */
:where(.el-button) {
  border-radius: var(--radius-md);
  font-weight: 500;
  transition: background var(--motion-duration-base) var(--motion-easing-standard),
              border-color var(--motion-duration-base) var(--motion-easing-standard);
}
:where(.el-button--primary) { background: var(--primary); border-color: var(--primary); }
:where(.el-button--primary:hover) { background: var(--primary-strong); border-color: var(--primary-strong); }
:where(.el-button:focus-visible) { box-shadow: 0 0 0 3px var(--primary-ring); }

:where(.el-input__wrapper, .el-select .el-input__wrapper) {
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px var(--line-soft) inset;
  transition: box-shadow var(--motion-duration-base);
}
:where(.el-input__wrapper.is-focus, .el-input__wrapper:focus-within) {
  box-shadow: 0 0 0 1px var(--primary) inset, 0 0 0 3px var(--primary-ring);
}

:where(.el-dialog) {
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
}
:where(.el-dialog__header) { padding: var(--space-5) var(--space-6); border-bottom: 1px solid var(--line-soft); }
:where(.el-dialog__body)   { padding: var(--space-5) var(--space-6); }
:where(.el-dialog__footer) { padding: var(--space-4) var(--space-6); border-top: 1px solid var(--line-soft); }

:where(.el-drawer) { border-radius: 0; }
:where(.el-drawer__header) { padding: var(--space-5) var(--space-6); border-bottom: 1px solid var(--line-soft); margin: 0; }
:where(.el-drawer__body)   { padding: var(--space-5) var(--space-6); }

:where(.el-table) {
  --el-table-border-color: var(--line-soft);
  --el-table-header-bg-color: var(--bg-surface-muted);
  --el-table-row-hover-bg-color: var(--bg-surface-muted);
  font-size: var(--text-small);
}
:where(.el-table th.el-table__cell) { color: var(--text-secondary); font-weight: 600; }
:where(.el-table .cell) { line-height: var(--leading-small); }

:where(.el-tag) {
  border-radius: var(--radius-sm);
  font-size: var(--text-caption);
  padding: 2px var(--space-2);
  height: auto;
  line-height: var(--leading-caption);
}

:where(.el-pagination .btn-prev, .el-pagination .btn-next, .el-pager li) {
  border-radius: var(--radius-md);
  min-width: 32px;
  height: 32px;
}

:where(.el-message)       { border-radius: var(--radius-lg); box-shadow: var(--shadow-md); }
:where(.el-notification)  { border-radius: var(--radius-lg); box-shadow: var(--shadow-md); }

:where(.el-dropdown-menu, .el-popper) {
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  padding: var(--space-1);
}
```

完整版包含 button-variant / input states / dialog max-width / drawer width 等更多覆写，规模控制在 ≤ 400 行。

### 2.3 `admin-web/src/main.js`（修改 import 顺序）

```js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import './assets/theme.css'
import './assets/element-overrides.css'   // ← 必须 import 在 theme.css 之后
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.mount('#app')
```

### 2.4 `admin-web/index.html`（修改头部 preconnect）

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
```

删除既有任何 `Fira` 相关 preconnect / preload。

### 2.5 `admin-web/src/components/Layout.vue`（全文重写）

主要变化：

- **侧栏底色** `--bg-sidebar`（slate-100）+ 右侧 1px slate-200 分隔线
- **分组导航**：5 组（仪表盘 / 媒体库 / 录入处理 / 服务 / 系统），每组上方 11px uppercase 小标题
- **可折叠**：顶部一个 chevron 按钮切换 240px ↔ 60px，状态写 `localStorage` key `admin-sidebar-collapsed`
- **profile chip**：侧栏底部固定，含首字母圆形头像 + 用户名 + admin 徽标；点击弹 popper 含"退出登录"
- **顶栏**：删除 `subtitle = '管理员工作台'`，删除"退出登录"按钮，右侧只保留命令面板触发按钮（图标 + `⌘K` 标记）
- **响应式**：≤ 1279px 自动折叠（但尊重用户既有展开偏好）；≤ 1023px 侧栏改 drawer
- **不再用** `<el-menu>` 自带样式——直接用 `<router-link>` + scoped CSS（Element Plus 菜单的样式覆盖在 5 分组场景下成本高于自写）

### 2.6 七个 `components/base/*.vue`

| 文件 | 接口（简） |
|------|-----------|
| `PageHeader.vue` | props: `title`, `subtitle?` (admin 默认 omitted)；slot: `actions` |
| `Toolbar.vue` | slot: `filters`, `actions`；prop `dense?` |
| `EmptyState.vue` | props: `icon?`, `title`, `description?`；slot: `action` |
| `StatCard.vue` | props: `label`, `value`, `trend?` (`up`/`down`/null), `delta?`；自动加 `.tabular-num` |
| `SectionCard.vue` | slots: `title`, `description?`, `actions?`, default body；prop `collapsible?` |
| `CommandPalette.vue` | 全局注册；内部维护 visible state；通过 `useEventBus` 或 provide/inject 暴露 open()；监听 `⌘K` / `Ctrl+K` |
| `BulkActionBar.vue` | props: `count`, `actions: Array<{label, icon?, type?, onClick}>`；自动浮起 + fade in 动画 |

每个组件 ≤ 200 行，使用 `<script setup>` + scoped CSS，全部 token 而非裸字面量。

### 2.7 `components/base/commandPalette.helpers.js` + `.spec.js`

```js
// commandPalette.helpers.js
export function matchMenuItem(query, item) {
  if (!query) return { score: 0, matched: true }
  const q = query.toLowerCase().trim()
  if (item.path.toLowerCase().startsWith(q)) return { score: 100, matched: true }
  if (item.label.includes(q)) return { score: 80, matched: true }
  if (item.alias?.toLowerCase()?.includes(q)) return { score: 60, matched: true }
  if (matchPinyinInitial(item.label, q)) return { score: 40, matched: true }
  return { score: 0, matched: false }
}

// 简单中文 → 拼音首字母映射表（不引入 pinyin-pro 全量包，自维护 200 项常用）
const PINYIN_INITIAL = {
  '视': 's', '频': 'p', '电': 'd', '剧': 'j', '上': 's', '传': 'c',
  // ...
}
export function matchPinyinInitial(label, q) {
  const initials = [...label].map(ch => PINYIN_INITIAL[ch] || ch.toLowerCase()).join('')
  return initials.includes(q)
}
```

`spec` 测试真值表：

| 输入 | 期望命中 | 优先级 |
|------|---------|--------|
| `'上传'` | 「上传视频」 | label 模糊 |
| `'sj'` | 「视频管理」 | 拼音首字母 |
| `'/videos'` | 「视频管理」 | 路径前缀 |
| `'upload'` | 「上传视频」 | alias |
| `''` | 所有项 | matched=true score=0 |

### 2.8 `views/Login.vue`（全文重写）

居中卡片，包含 logo + 标题 + 用户名/密码 + 登录按钮；套用新 token；不再依赖 `Layout.vue`（Login 是独立外壳）。

### 2.9 测试文件

`themeTokens.spec.js`：

```js
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

const css = readFileSync('src/assets/theme.css', 'utf8')

describe('theme tokens', () => {
  it('exports core color tokens', () => {
    expect(css).toContain('--primary: var(--blue-600)')
    expect(css).toContain('--text-primary: var(--slate-900)')
    expect(css).toContain('--bg-canvas: var(--slate-50)')
    expect(css).toContain('--bg-sidebar: var(--slate-100)')
  })
  it('exports 6 typography tiers', () => {
    expect(css).toMatch(/--text-display:\s*28px/)
    expect(css).toMatch(/--text-h1:\s*20px/)
    expect(css).toMatch(/--text-h2:\s*15px/)
    expect(css).toMatch(/--text-body:\s*14px/)
    expect(css).toMatch(/--text-small:\s*13px/)
    expect(css).toMatch(/--text-caption:\s*11px/)
  })
  it('removes legacy rose color tokens', () => {
    expect(css).not.toMatch(/#881337/i)
    expect(css).not.toMatch(/#be123c/i)
    expect(css).not.toMatch(/#7f1d1d/i)
    expect(css).not.toContain('Fira Code')
    expect(css).not.toContain('Fira Sans')
    expect(css).not.toContain('--font-code')
  })
})
```

`Layout.spec.js`（源文 audit）：

```js
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

const layout = readFileSync('src/components/Layout.vue', 'utf8')

describe('Layout shell', () => {
  it('uses grouped navigation', () => {
    expect(layout).toContain('分组')
    expect(layout).toContain('媒体库')
    expect(layout).toContain('录入处理')
    expect(layout).toContain('系统')
  })
  it('binds sidebar collapse to localStorage', () => {
    expect(layout).toContain('admin-sidebar-collapsed')
  })
  it('renders profile chip and command palette trigger', () => {
    expect(layout).toMatch(/profile/i)
    expect(layout).toMatch(/CommandPalette|command-palette/i)
  })
  it('removes legacy rose gradient and dual color titles', () => {
    expect(layout).not.toContain('#881337')
    expect(layout).not.toContain('#7f1d1d')
    expect(layout).not.toContain('linear-gradient(180deg, #881337')
    expect(layout).not.toContain("'管理员工作台'")
  })
})
```

## 3. plan.md 追加模板

```markdown
## 2026-05-23 · admin Phase 1：设计系统与外壳

- 摘要：admin-web 全面重设计第 1 阶段——重写 theme.css 引入完整设计 token，新增 element-overrides.css 全局覆写，重写 Layout.vue 实现浅灰侧栏 / 5 组导航 / 可折叠 / profile chip / 命令面板触发，新增 7 个 base wrapper 组件，重写 Login，引入 Inter 字体并删除 Fira 系。
- 受影响文件：admin-web/src/assets/theme.css、admin-web/src/assets/element-overrides.css、admin-web/src/components/Layout.vue、admin-web/src/components/base/*.vue、admin-web/src/views/Login.vue、admin-web/src/main.js、admin-web/index.html、新增 spec 三处
- 验证：cd admin-web && npm run build 通过；cd admin-web && npm test 全绿；H1–H6 / H13 / H14 / H15 手测通过；前后截图归档于 tasks/2026-05-23-admin-shell-redesign/screenshots/
```

## 4. 风险与回退

- **风险 1**：Element Plus 弹窗 / 抽屉的 mask z-index 与新覆写冲突——对策：覆写文件全用 `:where()` 降权，view 内 scoped 样式可继续 override
- **风险 2**：Inter 字体加载失败时 PingFang SC 兜底，可能在 mac 之外平台首屏中文字重偏轻——对策：字体栈 `'Microsoft YaHei'` 跟在 PingFang 之后；可选地把 Inter 自托管 woff2 进 `public/fonts/`
- **风险 3**：14 个业务视图引用了既有 `.content-card` / `.metric-card` / `.page-shell` 等 class——本期保留这些 class 名为别名（在新 theme.css 内提供软兼容），Phase 2 才统一替换为 SectionCard / PageHeader
- **回退**：所有改动是「新增 + 替换 token」纯加法改动；删除新增 7 个 base 组件 + 还原 Layout.vue + 还原 theme.css 即可回到 Phase 0
