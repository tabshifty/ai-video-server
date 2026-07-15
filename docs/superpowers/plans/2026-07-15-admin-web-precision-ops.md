# PC 管理端 Precision Ops 实施计划

> **供代理执行：** 必须使用 `subagent-driven-development`（推荐）或 `executing-plans` 逐任务实施；每个行为变更严格执行测试先行，每个任务完成后独立复核。

**目标：** 在不改变路由、权限、API、数据库和业务流程的前提下，把 `admin-web` 全站升级为高密度、高可扫描、可键盘操作的 Precision Ops 管理端。

**架构：** 继续使用 `theme.css` 语义 token、`element-overrides.css` 低特异性映射和 `components/base/` 结构组件三层架构。先稳定任务密度、壳层和本地偏好接口，再迁移 Dashboard、TaskMonitor、VideoList、ImageManage 四个样板页，之后按资源集合页、表单/工具页推广；复杂 SFC 只调整模板组织、共享组件接入和样式，保留现有请求、Drawer、选择和脏数据逻辑。

**技术栈：** Vue 3 `<script setup>`、Vue Router、Pinia、Element Plus、ECharts、Vitest、Vite。

## 全局约束

- 权威规格：`docs/superpowers/specs/2026-07-15-admin-web-precision-ops-design.md`。
- 所有 Markdown、界面文案和提交信息使用中文，提交前检查 U+FFFD 乱码替换字符。
- 不修改现有 API 路径、参数、响应结构、权限、路由目标、数据库、依赖或业务流程。
- 不新增 UI 框架、图标库、字体依赖或测试框架；继续使用 `@element-plus/icons-vue` 和现有 Vitest。
- 桌面侧栏为 224/56px，工作区页头 52px，桌面主区边距 20px；`<1024px` 使用 Drawer 导航和至少 44px 点击目标。
- 紧凑集合页控件 32px、表头 36px、文字行 40px、媒体行 52px；监控任务行 44px、进度条 6px；表单控件 36px。
- 普通区块圆角不超过 8px且不使用常驻阴影；浮层和批量操作条可使用阴影。
- 颜色不能是状态的唯一载体；纯图标按钮必须有中文 `aria-label` 和 tooltip；焦点始终可见。
- 本地偏好必须带 schema 版本，解析或存储失败时回退且不能让页面白屏。
- 每次视图切换、筛选、分页、刷新或 reload 后，批量选择继续按现有当前页契约清空或同步。
- 保留 `VideoList.vue`、`ImageManage.vue`、`ToolboxArchiveImport.vue` 和 `ToolboxImageWorkbench.vue` 的业务脚本边界，不借样式升级拆分其请求和编辑流程。
- 每个任务只精确暂存其生产/测试文件以及本任务新增的 `plan.md`、`CONTEXT.md` 记录；保留工作区中与本任务无关的用户改动。
- 每个任务开始、RED、实现和验证阶段都在 `plan.md` 顶部追加反向时间记录；修改生产代码或界面行为的任务必须同时在 `CONTEXT.md` 追加一条长期有效的 Precision Ops 契约，不能写临时进度。
- 管理端 Task 1-22 无论各任务步骤是否重复列出，都必须在提交前依次运行定向 Vitest、完整 `npm test` 和 `npm run build`；最终阶段再完成 375/768/1024/1440px 浏览器验收。

## 通用任务门禁

以下步骤适用于 Task 1-22，并由每个 task brief 的分派提示原文携带：

1. 在 `plan.md` 记录任务范围和待执行验证。
2. 先写定向测试并观察与缺失行为一致的 RED；把命令和关键失败写入 `plan.md`。
3. 完成最小实现后运行任务列出的定向 Vitest。
4. 运行 `cd admin-web && npm test`，要求所有 Vitest 通过。
5. 运行 `cd admin-web && npm run build`，只允许既有 chunk-size 警告，不允许新增 warning/error。
6. 运行 `git diff --check` 和 U+FFFD 扫描；修改生产代码或界面行为时，把本任务形成的长期约束追加到 `CONTEXT.md`。
7. 在 `plan.md` 追加验证结果，精确暂存本任务文件与本任务新增的账本/技术沉淀片段，再使用任务指定的中文提交信息提交。

## 文件与职责

| 文件 | 职责 |
|---|---|
| `admin-web/src/assets/theme.css` | Precision Ops 颜色、字号、间距、圆角、壳层和任务密度 token |
| `admin-web/src/assets/element-overrides.css` | 只在显式任务密度容器内映射 Element Plus 尺寸与表格行高 |
| `admin-web/src/components/Layout.vue` | 224/56px 侧栏、52px 工作区页头、最近访问、分组折叠和页面主操作 slot |
| `admin-web/src/components/adminShellPreferences.js` | 壳层最近访问与展开分组的版本化纯函数 |
| `admin-web/src/components/base/MetricStrip.vue` | 口径明确的紧凑指标条 |
| `admin-web/src/components/base/StatusIndicator.vue` | 状态点、文字和语义色的统一组合 |
| `admin-web/src/components/base/SavedViewTabs.vue` | 内置视图、自定义态和用户视图操作 |
| `admin-web/src/components/base/savedView.helpers.js` | 保存视图解析、快照比较、增删改与版本回退 |
| `admin-web/src/components/base/useSavedViews.js` | 保存视图持久化、选中态、自定义态和快照生命周期 composable |
| `admin-web/src/views/precisionOpsRollout.spec.js` | 25 个视图的阶段归属、页头迁移和密度静态契约 |
| 四个样板页及其测试 | 验证概览、监控、表格集合和媒体网格四种范式 |
| 其余 21 个视图 | 按阶段接入已验证的壳层、密度和状态模式 |

## 阶段一：共享基础与四个样板页

### Task 1: Precision Ops token 与显式任务密度

**Files:**
- Modify: `admin-web/src/assets/theme.css`
- Modify: `admin-web/src/assets/element-overrides.css`
- Modify: `admin-web/src/assets/themeTokens.spec.js`
- Modify: `admin-web/src/components/base/SectionCard.vue`
- Modify: `admin-web/src/components/base/EmptyState.vue`
- Modify: `admin-web/src/components/base/BulkActionBar.vue`

**Interfaces:**
- Produces: `--admin-sidebar-width: 224px`、`--admin-sidebar-collapsed-width: 56px`、`--admin-header-height: 52px`。
- Produces: `data-density="compact"`、`data-density="monitor"`、`data-density="form"` 三种显式作用域。
- Produces: `--control-height`、`--table-head-height`、`--table-row-height`、`--media-row-height`、`--section-padding`。

- [ ] **Step 1: 写 token 与密度作用域红灯测试**

在 `themeTokens.spec.js` 增加：

```js
const overrides = readFileSync(new URL('./element-overrides.css', import.meta.url), 'utf8')
const mainSource = readFileSync(new URL('../main.js', import.meta.url), 'utf8')
const densityTokenContracts = [
  {
    density: 'compact',
    tokens: [
      ['--control-height', '32px'],
      ['--table-head-height', '36px'],
      ['--table-row-height', '40px'],
      ['--media-row-height', '52px'],
      ['--section-padding', '12px']
    ]
  },
  {
    density: 'monitor',
    tokens: [
      ['--control-height', '32px'],
      ['--table-head-height', '36px'],
      ['--table-row-height', '44px'],
      ['--media-row-height', '44px'],
      ['--section-padding', '12px']
    ]
  },
  {
    density: 'form',
    tokens: [
      ['--control-height', '36px'],
      ['--section-padding', '16px']
    ]
  }
]

it('exports the approved Precision Ops shell and semantic tokens', () => {
  expect(css).toContain('--admin-sidebar-width: 224px')
  expect(css).toContain('--admin-sidebar-collapsed-width: 56px')
  expect(css).toContain('--admin-header-height: 52px')
  expect(css).toContain('--bg-canvas: #f7f8fa')
  expect(css).toContain('--text-primary: #172033')
  expect(css).toContain('--text-muted: #607085')
  expect(css).toContain('--success-600: #047857')
  expect(css).toContain('--warning-600: #b45309')
  expect(css).toContain('--danger-600: #c81e1e')
  expect(css).toContain('--info-600: #0369a1')
})

it('scopes compact sizing instead of applying it to every form', () => {
  expect(css).toContain('[data-density="compact"]')
  expect(css).toContain('[data-density="monitor"]')
  expect(css).toContain('[data-density="form"]')
  expect(overrides).toContain('min-width: 44px')
  expect(overrides).toContain('.el-button.is-circle')
  expect(overrides).toContain('.el-checkbox')
  expect(overrides).not.toMatch(/^:where\(\.el-button\)\s*\{[^}]*min-height:\s*32px/m)
})

it.each(densityTokenContracts)('exports every $density density token', ({ density, tokens }) => {
  const densityRule = css.match(new RegExp(`\\[data-density="${density}"\\]\\s*\\{([^}]*)\\}`))

  expect(densityRule).not.toBeNull()
  tokens.forEach(([token, value]) => {
    expect(densityRule?.[1]).toContain(`${token}: ${value}`)
  })
})

it('keeps component specificity in density overrides', () => {
  expect(overrides).not.toMatch(/:where\(\[data-density[^\n]*\)\s+:where\(/)

  const selectors = [
    ':where([data-density="compact"], [data-density="monitor"]) .el-select__wrapper',
    ':where([data-density="compact"], [data-density="monitor"]) .el-table th.el-table__cell',
    ':where([data-density="compact"], [data-density="monitor"]) .el-table td.el-table__cell',
    ':where([data-density="compact"]) .has-media-rows .el-table td.el-table__cell',
    ':where([data-density="form"]) .el-select__wrapper',
    ':where([data-density]) .el-pagination .btn-prev',
    ':where([data-density]) .el-pagination .btn-next',
    ':where([data-density]) .el-pager li',
    ':where([data-density]) .el-button.is-circle'
  ]

  selectors.forEach((selector) => {
    expect(overrides).toContain(selector)
  })
})

it('matches Element Plus pagination specificity in both narrow-screen rules', () => {
  expect(overrides).not.toContain(':where([data-density]) .el-pagination button')

  const paginationRules = [...overrides.matchAll(
    /:where\(\[data-density\]\) \.el-pagination \.btn-prev,\s*:where\(\[data-density\]\) \.el-pagination \.btn-next,\s*:where\(\[data-density\]\) \.el-pager li\s*\{([^}]*)\}/g
  )]

  expect(paginationRules).toHaveLength(2)
  paginationRules.forEach(([, body]) => {
    expect(body).toMatch(/min-height:\s*44px/)
  })
  expect(paginationRules.some(([, body]) => /min-width:\s*44px/.test(body))).toBe(true)
})

it('loads density overrides after Element Plus defaults and theme tokens', () => {
  const elementPlusStyles = mainSource.indexOf("import 'element-plus/dist/index.css'")
  const themeStyles = mainSource.indexOf("import './assets/theme.css'")
  const densityOverrides = mainSource.indexOf("import './assets/element-overrides.css'")

  expect(elementPlusStyles).toBeGreaterThan(-1)
  expect(themeStyles).toBeGreaterThan(elementPlusStyles)
  expect(densityOverrides).toBeGreaterThan(themeStyles)
})

it('exports the approved typography scale', () => {
  expect(css).toMatch(/--text-h1:\s*20px/)
  expect(css).toMatch(/--text-h2:\s*14px/)
  expect(css).toMatch(/--text-body:\s*14px/)
  expect(css).toMatch(/--text-small:\s*13px/)
  expect(css).toMatch(/--text-caption:\s*12px/)
  expect(css).toMatch(/--text-kpi:\s*24px/)
})
```

用上述字号测试替换现有仍断言 `--text-h2: 15px` 和 `--text-caption: 11px` 的旧测试块。

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/assets/themeTokens.spec.js`

Expected: FAIL，旧壳层仍为 240/60/64px，且没有三个 `data-density` 作用域。

- [ ] **Step 3: 替换核心 token 并保留兼容别名**

在 `theme.css` 用以下值替换对应定义；旧页面继续通过现有别名读取新 token：

```css
:root {
  --success-600: #047857;
  --warning-600: #b45309;
  --danger-600: #c81e1e;
  --info-600: #0369a1;

  --bg-canvas: #f7f8fa;
  --bg-surface: #ffffff;
  --bg-surface-muted: #f1f3f5;
  --bg-sidebar: #f1f3f5;
  --text-primary: #172033;
  --text-secondary: #475569;
  --text-muted: #607085;
  --text-on-inverse: #ffffff;
  --bg-inverse: #0f172a;
  --line-soft: #e2e8f0;
  --line-strong: #cbd5e1;

  --text-h1: 20px;
  --leading-h1: 28px;
  --text-h2: 14px;
  --leading-h2: 20px;
  --text-body: 14px;
  --leading-body: 20px;
  --text-small: 13px;
  --leading-small: 18px;
  --text-caption: 12px;
  --leading-caption: 16px;
  --text-kpi: 24px;
  --leading-kpi: 30px;
  --text-display: var(--text-kpi);
  --leading-display: var(--leading-kpi);

  --radius-xs: 4px;
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 8px;
  --radius-xl: 8px;
  --radius-2xl: 8px;

  --admin-sidebar-width: 224px;
  --admin-sidebar-collapsed-width: 56px;
  --admin-header-height: 52px;
}

[data-density="compact"] {
  --control-height: 32px;
  --table-head-height: 36px;
  --table-row-height: 40px;
  --media-row-height: 52px;
  --section-padding: 12px;
}

[data-density="monitor"] {
  --control-height: 32px;
  --table-head-height: 36px;
  --table-row-height: 44px;
  --media-row-height: 44px;
  --section-padding: 12px;
}

[data-density="form"] {
  --control-height: 36px;
  --section-padding: 16px;
}

body {
  letter-spacing: 0;
}
```

- [ ] **Step 4: 增加低特异性密度祖先和可覆盖的组件映射**

把以下规则加入 `element-overrides.css`，不改变未声明密度的页面。只用 `:where()` 清零密度祖先的特异性，组件目标必须保留自身 class/tag 特异性；`element-overrides.css` 在 Element Plus CSS 后加载，因此等于或高于默认选择器的目标规则可以通过正常级联获胜，不使用 `!important`：

```css
:where([data-density="compact"], [data-density="monitor"]) .el-button,
:where([data-density="compact"], [data-density="monitor"]) .el-input__wrapper,
:where([data-density="compact"], [data-density="monitor"]) .el-select__wrapper,
:where([data-density="compact"], [data-density="monitor"]) .el-segmented,
:where([data-density="compact"], [data-density="monitor"]) .el-radio-button__inner {
  min-height: var(--control-height);
}

:where([data-density="compact"], [data-density="monitor"]) .el-table th.el-table__cell {
  height: var(--table-head-height);
  padding-block: 0;
}

:where([data-density="compact"], [data-density="monitor"]) .el-table td.el-table__cell {
  height: var(--table-row-height);
  padding-block: 0;
}

:where([data-density="compact"]) .has-media-rows .el-table td.el-table__cell {
  height: var(--media-row-height);
}

:where([data-density="form"]) .el-button,
:where([data-density="form"]) .el-input__wrapper,
:where([data-density="form"]) .el-select__wrapper {
  min-height: var(--control-height);
}

@media (max-width: 63.9375rem) {
  :where([data-density]) .el-button,
  :where([data-density]) .el-input__wrapper,
  :where([data-density]) .el-select__wrapper,
  :where([data-density]) .el-radio-button__inner,
  :where([data-density]) .el-pagination .btn-prev,
  :where([data-density]) .el-pagination .btn-next,
  :where([data-density]) .el-pager li {
    min-height: 44px;
  }

  :where([data-density]) .el-button.is-circle,
  :where([data-density]) .el-checkbox,
  :where([data-density]) .el-radio,
  :where([data-density]) .el-pagination .btn-prev,
  :where([data-density]) .el-pagination .btn-next,
  :where([data-density]) .el-pager li {
    min-width: 44px;
    min-height: 44px;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

同时让 `SectionCard` 移除常驻阴影、`EmptyState` 最小高度收为 160px、`BulkActionBar` 圆角改为 `var(--radius-md)`；浮条保留 `var(--shadow-lg)`。

- [ ] **Step 5: 运行测试、构建并提交**

Run: `cd admin-web && npm test -- src/assets/themeTokens.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功，0 个错误。

Commit message: `样式：建立 Precision Ops 设计令牌`

---

### Task 2: 壳层偏好纯函数与五组权威导航

**Files:**
- Create: `admin-web/src/components/adminShellPreferences.js`
- Create: `admin-web/src/components/adminShellPreferences.spec.js`
- Modify: `admin-web/src/components/base/commandPalette.helpers.js`
- Modify: `admin-web/src/components/base/commandPalette.helpers.spec.js`
- Modify: `CONTEXT.md`

**Interfaces:**
- Produces: `SHELL_PREFERENCE_VERSION = 1`。
- Produces: `parseRecentRoutes(raw, validPaths)`、`pushRecentRoute(paths, path, validPaths, limit)`、`serializeRecentRoutes(paths)`。
- Produces: `parseExpandedGroupKeys(raw, validKeys)`、`ensureActiveGroup(keys, activeKey, validKeys)`、`serializeExpandedGroupKeys(keys)`。
- Changes: `adminShellNavGroups` 固定为 `overview`、`media`、`ingest`、`service-tools`、`system` 五组。

- [ ] **Step 1: 写壳层偏好和导航分组红灯测试**

创建测试：

```js
import { describe, expect, it } from 'vitest'
import {
  ensureActiveGroup,
  parseExpandedGroupKeys,
  parseRecentRoutes,
  pushRecentRoute,
  serializeExpandedGroupKeys,
  serializeRecentRoutes
} from './adminShellPreferences'
import { adminShellNavGroups } from './base/commandPalette.helpers'

const paths = ['/dashboard', '/videos', '/tasks', '/toolbox']
const groups = ['overview', 'media', 'ingest', 'service-tools', 'system']

describe('admin shell preferences', () => {
  it('keeps three unique recent known routes with newest first', () => {
    expect(pushRecentRoute(['/videos', '/dashboard'], '/tasks', paths)).toEqual(['/tasks', '/videos', '/dashboard'])
    expect(pushRecentRoute(['/tasks', '/videos', '/dashboard'], '/videos', paths)).toEqual(['/videos', '/tasks', '/dashboard'])
    expect(pushRecentRoute(['/tasks'], '/unknown', paths)).toEqual(['/tasks'])
    expect(pushRecentRoute(['/videos', '/tasks', '/toolbox'], '/dashboard', paths, 4)).toEqual(['/dashboard', '/videos', '/tasks'])
    expect(pushRecentRoute(['/videos'], '/dashboard', paths, 0)).toEqual([])
    expect(pushRecentRoute(['/videos', '/tasks'], '/dashboard', paths, 2.9)).toEqual(['/dashboard', '/videos'])
    expect(pushRecentRoute(['/videos', '/tasks'], '/dashboard', paths, Number.POSITIVE_INFINITY)).toEqual(['/dashboard', '/videos', '/tasks'])
  })

  it('falls back safely for corrupt or incompatible documents', () => {
    expect(parseRecentRoutes('{bad', paths)).toEqual([])
    expect(parseRecentRoutes('{"version":2,"paths":["/videos"]}', paths)).toEqual([])
    expect(parseExpandedGroupKeys('{bad', groups)).toEqual(groups)
    expect(parseExpandedGroupKeys('{"version":1}', groups)).toEqual(groups)
    expect(parseExpandedGroupKeys('{"version":1,"keys":null}', groups)).toEqual(groups)
    expect(parseExpandedGroupKeys('{"version":1,"keys":"overview"}', groups)).toEqual(groups)
    expect(parseExpandedGroupKeys('{"version":1,"keys":[]}', groups)).toEqual([])
  })

  it('round trips versioned documents and forces the active group open', () => {
    expect(parseRecentRoutes(serializeRecentRoutes(['/videos']), paths)).toEqual(['/videos'])
    expect(parseExpandedGroupKeys(serializeExpandedGroupKeys(['overview']), groups)).toEqual(['overview'])
    expect(ensureActiveGroup(['overview'], 'media', groups)).toEqual(['overview', 'media'])
  })

  it('exposes exactly five navigation groups', () => {
    expect(adminShellNavGroups.map((group) => [group.key, group.label])).toEqual([
      ['overview', '仪表盘'],
      ['media', '媒体库'],
      ['ingest', '录入处理'],
      ['service-tools', '服务与工具'],
      ['system', '系统']
    ])
    const serviceTools = adminShellNavGroups.find((group) => group.key === 'service-tools')
    expect(serviceTools.items.map(({ path, title, icon }) => ({ path, title, icon }))).toEqual([
      { path: '/iptv', title: 'IPTV 管理', icon: 'Monitor' },
      { path: '/tasks', title: '任务监控', icon: 'List' },
      { path: '/toolbox', title: '工具箱', icon: 'Tools' }
    ])
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/components/adminShellPreferences.spec.js src/components/base/commandPalette.helpers.spec.js`

Expected: FAIL，偏好模块不存在，导航仍有“服务”和“工具箱”两个分组。

- [ ] **Step 3: 实现版本化偏好纯函数**

创建 `adminShellPreferences.js`：

```js
export const SHELL_PREFERENCE_VERSION = 1

function parseDocument(raw) {
  if (typeof raw !== 'string' || raw.trim() === '') return null
  try {
    const value = JSON.parse(raw)
    return value && value.version === SHELL_PREFERENCE_VERSION ? value : null
  } catch (_) {
    return null
  }
}

function knownUnique(values, validValues, limit = Number.POSITIVE_INFINITY) {
  const valid = new Set(validValues)
  const seen = new Set()
  const result = []
  for (const value of Array.isArray(values) ? values : []) {
    if (!valid.has(value) || seen.has(value)) continue
    seen.add(value)
    result.push(value)
    if (result.length >= limit) break
  }
  return result
}

export function parseRecentRoutes(raw, validPaths) {
  const document = parseDocument(raw)
  return document && Array.isArray(document.paths) ? knownUnique(document.paths, validPaths, 3) : []
}

export function pushRecentRoute(paths, path, validPaths, limit = 3) {
  const finiteLimit = Number.isFinite(limit) ? Math.trunc(limit) : 3
  const normalizedLimit = Math.min(3, Math.max(0, finiteLimit))
  if (normalizedLimit === 0) return []
  if (!validPaths.includes(path)) return knownUnique(paths, validPaths, normalizedLimit)
  return knownUnique([path, ...(Array.isArray(paths) ? paths : [])], validPaths, normalizedLimit)
}

export function serializeRecentRoutes(paths) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, paths })
}

export function parseExpandedGroupKeys(raw, validKeys) {
  const document = parseDocument(raw)
  return document && Array.isArray(document.keys) ? knownUnique(document.keys, validKeys) : [...validKeys]
}

export function ensureActiveGroup(keys, activeKey, validKeys) {
  return knownUnique([...(Array.isArray(keys) ? keys : []), activeKey], validKeys)
}

export function serializeExpandedGroupKeys(keys) {
  return JSON.stringify({ version: SHELL_PREFERENCE_VERSION, keys })
}
```

- [ ] **Step 4: 合并“服务与工具”分组并补长期 key 契约**

在 `commandPalette.helpers.js` 用一个分组替换原 `service` 和 `tools`：

```js
{
  key: 'service-tools',
  label: '服务与工具',
  items: [
    { path: '/iptv', label: 'IPTV 管理', title: 'IPTV 管理', icon: 'Monitor', alias: 'iptv live' },
    { path: '/tasks', label: '任务监控', title: '任务监控', icon: 'List', alias: 'task tasks jobs rw' },
    {
      path: '/toolbox',
      label: '工具箱',
      title: '工具箱',
      icon: 'Tools',
      alias: 'toolbox tools ed2k orphan scan orphan-files 孤儿文件扫描 archive archive-import archive import zip rar 7z 压缩包导入 压缩包 password vault credentials 密码 密码库 密码管理 gjx'
    }
  ]
}
```

在 `CONTEXT.md` 的壳层术语中记录：最近访问 key 为 `admin-recent-routes-v1`，分组展开 key 为 `admin-nav-groups-v1`，两者文档结构均为 `{ version: 1, ... }`；旧值、损坏 JSON、字段语义损坏或 storage 异常回退默认。合法空 `keys` 保留全部收起语义，`parseRecentRoutes` 固定截断到 3 条，`pushRecentRoute` 将调用方 limit 归一化到 0..3。

- [ ] **Step 5: 运行测试并提交**

Run: `cd admin-web && npm test -- src/components/adminShellPreferences.spec.js src/components/base/commandPalette.helpers.spec.js`

Expected: PASS。

Commit message: `功能：统一管理端导航与壳层偏好`

---

### Task 3: 52px 工作区页头、最近访问与折叠分组

**Files:**
- Modify: `admin-web/src/components/Layout.vue`
- Modify: `admin-web/src/components/Layout.spec.js`

**Interfaces:**
- Consumes: Task 2 的导航和平稳回退函数。
- Produces: Layout 命名 slot `header-actions`。
- Produces: `activeGroupKey`、`recentNavItems`、`isGroupExpanded(key)`、`toggleGroup(key)`。
- Compatibility: `route.meta.hideShellPageHeader === true` 时继续隐藏工作区身份区，供未迁移页面保留自身 `PageHeader`。

- [ ] **Step 1: 写 Layout 壳层红灯测试**

替换或扩展 `Layout.spec.js` 的契约：

```js
it('renders the Precision Ops workspace header contract', () => {
  expect(layout).toContain('admin-recent-routes-v1')
  expect(layout).toContain('admin-nav-groups-v1')
  expect(layout).toContain('name="header-actions"')
  expect(layout).toContain('matchedNavItem.value?.groupLabel')
  expect(layout).toContain('recentNavItems')
  expect(layout).toContain('toggleGroup(group.key)')
  expect(layout).toContain('aria-expanded')
  expect(layout).toContain('var(--admin-header-height)')
})

it('keeps the migration compatibility boundary', () => {
  expect(layout).toContain('const showShellPageHeader = computed(() => !route.meta?.hideShellPageHeader)')
  expect(layout).toContain('v-if="showShellPageHeader"')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/components/Layout.spec.js`

Expected: FAIL，Layout 尚无最近访问、折叠分组和 `header-actions` slot。

- [ ] **Step 3: 接入壳层偏好状态**

在 `Layout.vue` 导入 Task 2 函数并增加以下状态；所有 storage 读写继续使用 `try/catch`：

```js
import {
  adminShellNavGroups,
  adminShellNavItems,
  findAdminNavItemByPath,
  openCommandPalette
} from './base/commandPalette.helpers'
import {
  ensureActiveGroup,
  parseExpandedGroupKeys,
  parseRecentRoutes,
  pushRecentRoute,
  serializeExpandedGroupKeys,
  serializeRecentRoutes
} from './adminShellPreferences'

const RECENT_ROUTES_KEY = 'admin-recent-routes-v1'
const NAV_GROUPS_KEY = 'admin-nav-groups-v1'
const validNavPaths = adminShellNavItems.map((item) => item.path)
const validGroupKeys = adminShellNavGroups.map((group) => group.key)
const recentRoutePaths = ref(readStoredRecentRoutes())
const expandedGroupKeys = ref(readStoredExpandedGroups())
const activeGroupKey = computed(() => matchedNavItem.value?.groupKey || '')
const pageGroupLabel = computed(() => matchedNavItem.value?.groupLabel || '管理后台')
const recentNavItems = computed(() => recentRoutePaths.value
  .map((path) => findAdminNavItemByPath(path))
  .filter(Boolean))

function readStoredRecentRoutes() {
  if (typeof window === 'undefined') return []
  try {
    return parseRecentRoutes(window.localStorage.getItem(RECENT_ROUTES_KEY), validNavPaths)
  } catch (_) {
    return []
  }
}

function readStoredExpandedGroups() {
  if (typeof window === 'undefined') return [...validGroupKeys]
  try {
    return parseExpandedGroupKeys(window.localStorage.getItem(NAV_GROUPS_KEY), validGroupKeys)
  } catch (_) {
    return [...validGroupKeys]
  }
}

function persistShellPreference(key, value) {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(key, value)
  } catch (_) {}
}

function isGroupExpanded(key) {
  return expandedGroupKeys.value.includes(key)
}

function toggleGroup(key) {
  expandedGroupKeys.value = isGroupExpanded(key)
    ? expandedGroupKeys.value.filter((item) => item !== key)
    : [...expandedGroupKeys.value, key]
  persistShellPreference(NAV_GROUPS_KEY, serializeExpandedGroupKeys(expandedGroupKeys.value))
}
```

将现有 route watch 扩展为：

```js
watch(
  () => route.fullPath,
  () => {
    const path = route.path
    mobileNavVisible.value = false
    recentRoutePaths.value = pushRecentRoute(recentRoutePaths.value, path, validNavPaths, 3)
    expandedGroupKeys.value = ensureActiveGroup(expandedGroupKeys.value, activeGroupKey.value, validGroupKeys)
    persistShellPreference(RECENT_ROUTES_KEY, serializeRecentRoutes(recentRoutePaths.value))
    persistShellPreference(NAV_GROUPS_KEY, serializeExpandedGroupKeys(expandedGroupKeys.value))
  },
  { immediate: true }
)
```

- [ ] **Step 4: 改造桌面/Drawer 导航和工作区页头**

工作区页头使用以下结构，页面操作只通过 slot 接入：

```vue
<header class="shell-header">
  <div class="shell-header__left">
    <el-button class="mobile-nav-btn" text :icon="Menu" aria-label="打开导航菜单" @click="openMobileNav" />
    <div v-if="showShellPageHeader" class="workspace-identity">
      <span>{{ pageGroupLabel }}</span>
      <h1>{{ pageTitle }}</h1>
    </div>
  </div>
  <button class="command-trigger" type="button" @click="openCommandPalette">
    <el-icon><Search /></el-icon>
    <span>快速跳转</span>
    <kbd>⌘K</kbd>
  </button>
  <div v-if="$slots['header-actions']" class="shell-header__actions">
    <slot name="header-actions" />
  </div>
</header>
```

在 `.nav-groups` 顶部渲染最多 3 个 `recentNavItems`；每个正式分组标题改成带 `aria-expanded` 的按钮，列表使用 `v-show="isGroupExpanded(group.key)"`。折叠侧栏只显示图标和 tooltip，不渲染分组按钮文字；Drawer 使用同一组 `navGroups` 和展开状态。CSS 固定页头高度 52px、主区桌面 padding 20px、`<1024px` 16px、`<768px` 12px，所有导航目标在窄屏至少 44px。

桌面最近访问和正式分组的 RouterLink 都会在折叠态隐藏 `.nav-link__label`，因此两处必须直接用权威菜单标签提供可访问名称；tooltip 只作为视觉提示，不能替代链接 name。Drawer 的链接文字始终可见，不重复增加该属性。

```vue
<RouterLink
  class="nav-link"
  :class="{ 'is-active': isActive(item) }"
  :to="item.path"
  :aria-current="isActive(item) ? 'page' : undefined"
  :aria-label="item.label"
  @click="closeMobileNav"
>
  <el-icon><component :is="resolveIcon(item.icon)" /></el-icon>
  <span class="nav-link__label">{{ item.label }}</span>
</RouterLink>
```

把 `.nav-group__label` 和 `.drawer-nav__label` 的 `letter-spacing: 0.08em` 都改为 `letter-spacing: 0`，与全站字距契约一致。

- [ ] **Step 5: 运行壳层测试、全量测试和构建**

Run: `cd admin-web && npm test -- src/components/Layout.spec.js src/components/adminShellPreferences.spec.js src/components/base/commandPalette.helpers.spec.js`

Expected: PASS。

Run: `cd admin-web && npm test && npm run build`

Expected: 全部测试和 Vite 构建通过。

Commit message: `样式：升级管理端工作区壳层`

---

### Task 4: 紧凑指标与可访问状态组件

**Files:**
- Create: `admin-web/src/components/base/MetricStrip.vue`
- Create: `admin-web/src/components/base/StatusIndicator.vue`
- Create: `admin-web/src/components/base/precisionOpsComponents.spec.js`

**Interfaces:**
- `MetricStrip.items`: `Array<{ key?: string, label: string, value: string|number, scope?: string, tone?: 'neutral'|'success'|'warning'|'danger'|'info' }>`。
- `StatusIndicator.label`: string；`tone`: `neutral|success|warning|danger|info`；`icon`: 可选 Element Plus 图标对象。

- [ ] **Step 1: 写组件静态契约红灯测试**

```js
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const metricStrip = readFileSync(new URL('./MetricStrip.vue', import.meta.url), 'utf8')
const statusIndicator = readFileSync(new URL('./StatusIndicator.vue', import.meta.url), 'utf8')

describe('Precision Ops base components', () => {
  it('renders labeled tabular metrics without cards', () => {
    expect(metricStrip).toContain('metric-strip')
    expect(metricStrip).toContain('tabular-num')
    expect(metricStrip).toContain('item.scope')
    expect(metricStrip).not.toContain('box-shadow')
  })

  it('combines text, shape and semantic tone', () => {
    expect(statusIndicator).toContain('status-indicator__dot')
    expect(statusIndicator).toContain('{{ label }}')
    expect(statusIndicator).toContain('aria-label')
    expect(statusIndicator).toContain('status-indicator--danger')
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/components/base/precisionOpsComponents.spec.js`

Expected: FAIL，两个组件文件尚不存在。

- [ ] **Step 3: 实现 MetricStrip**

```vue
<script setup>
defineProps({
  items: { type: Array, required: true },
  ariaLabel: { type: String, default: '指标摘要' }
})
</script>

<template>
  <section class="metric-strip" :aria-label="ariaLabel">
    <div v-for="item in items" :key="item.key || item.label" class="metric-strip__item" :class="`metric-strip__item--${item.tone || 'neutral'}`">
      <div class="metric-strip__label">
        <span>{{ item.label }}</span>
        <em v-if="item.scope">{{ item.scope }}</em>
      </div>
      <strong class="metric-strip__value tabular-num">{{ item.value }}</strong>
    </div>
  </section>
</template>

<style scoped>
.metric-strip { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); border-block: 1px solid var(--line-soft); background: var(--bg-surface); }
.metric-strip__item { min-width: 0; padding: var(--space-3); border-right: 1px solid var(--line-soft); }
.metric-strip__item:last-child { border-right: 0; }
.metric-strip__label { display: flex; align-items: baseline; justify-content: space-between; gap: var(--space-2); color: var(--text-secondary); font-size: var(--text-caption); }
.metric-strip__label em { color: var(--text-muted); font-style: normal; }
.metric-strip__value { display: block; margin-top: var(--space-1); color: var(--text-primary); font-size: var(--text-kpi); line-height: var(--leading-kpi); font-weight: 600; }
.metric-strip__item--success .metric-strip__value { color: var(--success-600); }
.metric-strip__item--warning .metric-strip__value { color: var(--warning-600); }
.metric-strip__item--danger .metric-strip__value { color: var(--danger-600); }
.metric-strip__item--info .metric-strip__value { color: var(--info-600); }
@media (max-width: 63.9375rem) { .metric-strip { grid-template-columns: repeat(2, minmax(0, 1fr)); } .metric-strip__item:nth-child(2n) { border-right: 0; } }
@media (max-width: 36rem) { .metric-strip { grid-template-columns: 1fr; } .metric-strip__item { border-right: 0; } }
</style>
```

- [ ] **Step 4: 实现 StatusIndicator**

```vue
<script setup>
defineProps({
  label: { type: String, required: true },
  tone: { type: String, default: 'neutral' },
  icon: { type: [Object, Function], default: null }
})
</script>

<template>
  <span class="status-indicator" :class="`status-indicator--${tone}`" :aria-label="label">
    <el-icon v-if="icon"><component :is="icon" /></el-icon>
    <span v-else class="status-indicator__dot" aria-hidden="true" />
    <span>{{ label }}</span>
  </span>
</template>

<style scoped>
.status-indicator { display: inline-flex; align-items: center; gap: var(--space-1); color: var(--text-secondary); font-size: var(--text-small); white-space: nowrap; }
.status-indicator__dot { width: 8px; height: 8px; flex: 0 0 auto; border: 2px solid currentColor; border-radius: 50%; }
.status-indicator--success { color: var(--success-600); }
.status-indicator--warning { color: var(--warning-600); }
.status-indicator--danger { color: var(--danger-600); }
.status-indicator--info { color: var(--info-600); }
</style>
```

- [ ] **Step 5: 运行测试并提交**

Run: `cd admin-web && npm test -- src/components/base/precisionOpsComponents.spec.js`

Expected: PASS。

Commit message: `组件：增加紧凑指标与状态指示`

---

### Task 5: 保存视图纯函数与标签组件

**Files:**
- Create: `admin-web/src/components/base/savedView.helpers.js`
- Create: `admin-web/src/components/base/savedView.helpers.spec.js`
- Create: `admin-web/src/components/base/useSavedViews.js`
- Create: `admin-web/src/components/base/useSavedViews.spec.js`
- Create: `admin-web/src/components/base/SavedViewTabs.vue`
- Modify: `admin-web/src/components/base/precisionOpsComponents.spec.js`

**Interfaces:**
- Produces: `SAVED_VIEW_SCHEMA_VERSION = 1`、`CUSTOM_VIEW_ID = 'custom'`。
- Produces: `snapshotKey(snapshot)`、`parseSavedViewDocument(raw, normalizeSnapshot)`、`serializeSavedViews(items)`、`upsertSavedView(items, view)`、`removeSavedView(items, id)`、`createSavedViewId(now)`。
- Produces: `useSavedViews(options)`，统一管理 storage 容错、用户视图、选中来源、自定义态、save/update/rename/remove 和切换后的刷新。
- `SavedViewTabs`: props `items`、`activeId`、`editableSourceId`；在组件内完成名称 prompt 和删除 confirm 后 emits `select`、`save(label)`、`update(id)`、`rename({ id, label })`、`remove(id)`。

- [ ] **Step 1: 写保存视图红灯测试**

```js
import { describe, expect, it } from 'vitest'
import {
  createSavedViewId,
  parseSavedViewDocument,
  removeSavedView,
  serializeSavedViews,
  snapshotKey,
  upsertSavedView
} from './savedView.helpers'

const normalize = (snapshot) => ({
  q: String(snapshot?.q || ''),
  status: String(snapshot?.status || ''),
  columns: Array.isArray(snapshot?.columns) ? snapshot.columns : []
})

describe('saved view helpers', () => {
  it('compares snapshots independent of object key order', () => {
    expect(snapshotKey({ status: 'failed', q: '' })).toBe(snapshotKey({ q: '', status: 'failed' }))
  })

  it('round trips versioned user views', () => {
    const items = [{ id: 'user-1', label: '失败处理', snapshot: normalize({ status: 'failed' }) }]
    expect(parseSavedViewDocument(serializeSavedViews(items), normalize)).toEqual(items)
  })

  it('drops corrupt, incompatible and duplicate records', () => {
    expect(parseSavedViewDocument('{bad', normalize)).toEqual([])
    expect(parseSavedViewDocument('{"version":2,"items":[]}', normalize)).toEqual([])
    const raw = JSON.stringify({ version: 1, items: [
      { id: 'user-1', label: 'A', snapshot: {} },
      { id: 'user-1', label: 'B', snapshot: {} },
      { id: '', label: 'C', snapshot: {} }
    ] })
    expect(parseSavedViewDocument(raw, normalize)).toHaveLength(1)
  })

  it('upserts, removes and creates stable ids', () => {
    expect(upsertSavedView([], { id: 'user-1', label: 'A', snapshot: {} })).toHaveLength(1)
    expect(upsertSavedView([{ id: 'user-1', label: 'A', snapshot: {} }], { id: 'user-1', label: 'B', snapshot: {} })[0].label).toBe('B')
    expect(removeSavedView([{ id: 'user-1' }], 'user-1')).toEqual([])
    expect(createSavedViewId(123)).toBe('user-123')
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/components/base/savedView.helpers.spec.js`

Expected: FAIL，helper 文件尚不存在。

- [ ] **Step 3: 实现保存视图纯函数**

```js
export const SAVED_VIEW_SCHEMA_VERSION = 1
export const CUSTOM_VIEW_ID = 'custom'

function canonicalize(value) {
  if (Array.isArray(value)) return value.map(canonicalize)
  if (!value || typeof value !== 'object') return value
  return Object.keys(value).sort().reduce((result, key) => {
    result[key] = canonicalize(value[key])
    return result
  }, {})
}

export function snapshotKey(snapshot) {
  return JSON.stringify(canonicalize(snapshot && typeof snapshot === 'object' ? snapshot : {}))
}

export function parseSavedViewDocument(raw, normalizeSnapshot) {
  try {
    const document = JSON.parse(String(raw || ''))
    if (document?.version !== SAVED_VIEW_SCHEMA_VERSION || !Array.isArray(document.items)) return []
    const seen = new Set()
    return document.items.flatMap((item) => {
      const id = String(item?.id || '').trim()
      const label = String(item?.label || '').trim()
      if (!id.startsWith('user-') || !label || seen.has(id)) return []
      seen.add(id)
      return [{ id, label, snapshot: normalizeSnapshot(item.snapshot) }]
    })
  } catch (_) {
    return []
  }
}

export function serializeSavedViews(items) {
  return JSON.stringify({ version: SAVED_VIEW_SCHEMA_VERSION, items })
}

export function upsertSavedView(items, view) {
  const next = (Array.isArray(items) ? items : []).filter((item) => item.id !== view.id)
  return [...next, view]
}

export function removeSavedView(items, id) {
  return (Array.isArray(items) ? items : []).filter((item) => item.id !== id)
}

export function createSavedViewId(now = Date.now()) {
  return `user-${now}`
}
```

- [ ] **Step 4: 写 composable 红灯测试并确认失败**

创建 `useSavedViews.spec.js`，使用响应式页面快照和内存 storage 验证共享控制器，而不是在两个页面复制状态机：

```js
import { reactive } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import { CUSTOM_VIEW_ID } from './savedView.helpers'
import { useSavedViews } from './useSavedViews'

const normalize = (snapshot) => ({ q: String(snapshot?.q || '') })
const builtInViews = [
  { id: 'builtin-all', label: '全部', builtIn: true, snapshot: normalize({}) }
]

function createStorage(raw = '') {
  let value = raw
  return {
    getItem: vi.fn(() => value),
    setItem: vi.fn((_, next) => { value = next })
  }
}

describe('useSavedViews', () => {
  it('owns custom state and the complete saved-view lifecycle', async () => {
    const page = reactive({ q: '' })
    const storage = createStorage()
    const refresh = vi.fn()
    const controller = useSavedViews({
      storageKey: 'test-saved-views-v1',
      builtInViews,
      normalizeSnapshot: normalize,
      getCurrentSnapshot: () => page,
      applySnapshot: (snapshot) => { page.q = snapshot.q },
      refresh,
      storage,
      now: () => 123
    })

    expect(controller.activeViewId.value).toBe('builtin-all')
    page.q = '失败'
    expect(controller.activeViewId.value).toBe(CUSTOM_VIEW_ID)
    const id = controller.saveView('  失败处理  ')
    expect(id).toBe('user-123')
    expect(controller.activeViewId.value).toBe(id)
    expect(controller.renameView({ id, label: '待处理' })).toBe(true)
    page.q = ''
    await controller.selectView(id)
    expect(page.q).toBe('失败')
    expect(refresh).toHaveBeenCalledTimes(1)
    page.q = '重试'
    expect(controller.updateView(id)).toBe(true)
    await controller.removeView(id)
    expect(controller.activeViewId.value).toBe('builtin-all')
    expect(storage.setItem).toHaveBeenCalled()
  })

  it('keeps the session usable when storage is unavailable', () => {
    const storage = {
      getItem: () => { throw new Error('denied') },
      setItem: () => { throw new Error('denied') }
    }
    const controller = useSavedViews({
      storageKey: 'test-saved-views-v1',
      builtInViews,
      normalizeSnapshot: normalize,
      getCurrentSnapshot: () => ({ q: '' }),
      applySnapshot: () => {},
      storage,
      now: () => 456
    })
    expect(controller.saveView('会话视图')).toBe('user-456')
    expect(controller.userViews.value).toHaveLength(1)
  })
})
```

Run: `cd admin-web && npm test -- src/components/base/useSavedViews.spec.js`

Expected: FAIL，composable 尚不存在。

- [ ] **Step 5: 实现共享 useSavedViews 控制器**

创建 `useSavedViews.js`。页面必须通过 `getCurrentSnapshot` 提供可追踪快照，通过 `applySnapshot` 负责页面字段、列偏好和选择清理，通过 `refresh` 负责业务加载；composable 不知道任何页面业务字段：

```js
import { computed, ref } from 'vue'
import {
  CUSTOM_VIEW_ID,
  createSavedViewId,
  parseSavedViewDocument,
  removeSavedView as removeSavedViewItem,
  serializeSavedViews,
  snapshotKey,
  upsertSavedView
} from './savedView.helpers'

function resolveStorage(storage) {
  if (storage !== undefined) return storage
  try {
    return typeof window === 'undefined' ? null : window.localStorage
  } catch (_) {
    return null
  }
}

export function useSavedViews({
  storageKey,
  builtInViews,
  normalizeSnapshot,
  getCurrentSnapshot,
  applySnapshot,
  refresh = () => {},
  storage,
  now = Date.now
}) {
  const targetStorage = resolveStorage(storage)
  const defaultView = builtInViews[0] || null
  let initialViews = []
  try {
    initialViews = parseSavedViewDocument(targetStorage?.getItem(storageKey), normalizeSnapshot)
  } catch (_) {}

  const userViews = ref(initialViews)
  const selectedViewId = ref(defaultView?.id || '')
  const availableViews = computed(() => [...builtInViews, ...userViews.value])
  const selectedView = computed(() => availableViews.value.find((item) => item.id === selectedViewId.value) || defaultView)
  const currentSnapshot = computed(() => normalizeSnapshot(getCurrentSnapshot()))
  const activeViewId = computed(() => selectedView.value && snapshotKey(currentSnapshot.value) === snapshotKey(normalizeSnapshot(selectedView.value.snapshot))
    ? selectedView.value.id
    : CUSTOM_VIEW_ID)
  const editableSourceId = computed(() => selectedView.value && !selectedView.value.builtIn ? selectedView.value.id : '')

  function persist() {
    try {
      targetStorage?.setItem(storageKey, serializeSavedViews(userViews.value))
    } catch (_) {}
  }

  async function selectView(id) {
    const view = availableViews.value.find((item) => item.id === id)
    if (!view) return false
    selectedViewId.value = view.id
    applySnapshot(normalizeSnapshot(view.snapshot))
    await refresh()
    return true
  }

  function saveView(label) {
    const normalizedLabel = String(label || '').trim()
    if (!normalizedLabel) return null
    const view = { id: createSavedViewId(now()), label: normalizedLabel, snapshot: currentSnapshot.value }
    userViews.value = upsertSavedView(userViews.value, view)
    selectedViewId.value = view.id
    persist()
    return view.id
  }

  function updateView(id) {
    const source = userViews.value.find((item) => item.id === id)
    if (!source) return false
    userViews.value = upsertSavedView(userViews.value, { ...source, snapshot: currentSnapshot.value })
    selectedViewId.value = id
    persist()
    return true
  }

  function renameView({ id, label }) {
    const source = userViews.value.find((item) => item.id === id)
    const normalizedLabel = String(label || '').trim()
    if (!source || !normalizedLabel) return false
    userViews.value = upsertSavedView(userViews.value, { ...source, label: normalizedLabel })
    persist()
    return true
  }

  async function removeView(id) {
    if (!userViews.value.some((item) => item.id === id) || !defaultView) return false
    userViews.value = removeSavedViewItem(userViews.value, id)
    persist()
    return selectView(defaultView.id)
  }

  return {
    userViews,
    availableViews,
    activeViewId,
    editableSourceId,
    selectView,
    saveView,
    updateView,
    renameView,
    removeView
  }
}
```

- [ ] **Step 6: 实现 SavedViewTabs 的完整命令面**

组件必须：用 tabs 切换视图；当前快照偏离所选视图时显示临时“自定义”；自定义态可“另存为视图”；来源为用户视图时额外显示“更新视图”；用户视图可重命名和删除；内置视图不可覆盖、重命名或删除。名称输入和删除确认只能存在于此组件，页面和 composable 不得调用 `ElMessageBox`：

```vue
<script setup>
import { computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { EditPen, MoreFilled, Plus, RefreshRight, Delete } from '@element-plus/icons-vue'
import { CUSTOM_VIEW_ID } from './savedView.helpers'

const props = defineProps({
  items: { type: Array, required: true },
  activeId: { type: String, required: true },
  editableSourceId: { type: String, default: '' }
})
const emit = defineEmits(['select', 'save', 'update', 'rename', 'remove'])
const visibleItems = computed(() => props.activeId === CUSTOM_VIEW_ID
  ? [...props.items, { id: CUSTOM_VIEW_ID, label: '自定义', builtIn: true, transient: true }]
  : props.items)
const activeItem = computed(() => props.items.find((item) => item.id === props.activeId) || null)
const editableSource = computed(() => props.items.find((item) => item.id === props.editableSourceId && !item.builtIn) || null)

function isDismissed(error) {
  return error === 'cancel' || error === 'close'
}

async function requestSave() {
  try {
    const { value } = await ElMessageBox.prompt('请输入视图名称', '保存视图', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (text) => String(text || '').trim() !== '' || '请输入视图名称'
    })
    emit('save', String(value).trim())
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}

async function requestRename() {
  if (!activeItem.value || activeItem.value.builtIn) return
  try {
    const { value } = await ElMessageBox.prompt('请输入新的视图名称', '重命名视图', {
      inputValue: activeItem.value.label,
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (text) => String(text || '').trim() !== '' || '请输入视图名称'
    })
    emit('rename', { id: activeItem.value.id, label: String(value).trim() })
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}

async function requestRemove() {
  if (!activeItem.value || activeItem.value.builtIn) return
  try {
    await ElMessageBox.confirm('确认删除这个保存视图？', '删除视图', { type: 'warning' })
    emit('remove', activeItem.value.id)
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}
</script>

<template>
  <section class="saved-view-tabs" aria-label="保存视图">
    <el-tabs :model-value="activeId" class="saved-view-tabs__tabs" @tab-change="(id) => emit('select', id)">
      <el-tab-pane v-for="item in visibleItems" :key="item.id" :name="item.id" :label="item.label" />
    </el-tabs>
    <div class="saved-view-tabs__actions">
      <el-button v-if="activeId === CUSTOM_VIEW_ID" :icon="Plus" @click="requestSave">另存为视图</el-button>
      <el-button v-if="activeId === CUSTOM_VIEW_ID && editableSource" :icon="RefreshRight" @click="emit('update', editableSource.id)">更新视图</el-button>
      <el-dropdown v-if="activeItem && !activeItem.builtIn" trigger="click">
        <el-tooltip content="视图操作" placement="top">
          <el-button :icon="MoreFilled" circle aria-label="视图操作" />
        </el-tooltip>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :icon="EditPen" @click="requestRename">重命名</el-dropdown-item>
            <el-dropdown-item :icon="Delete" divided @click="requestRemove">删除</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </section>
</template>
```

CSS 使用一条下边框和 32px 操作按钮，不添加卡片背景；`<1024px` 时允许 tabs 自身横向滚动且按钮至少 44px。把组件静态契约加入 `precisionOpsComponents.spec.js`。

```vue
<style scoped>
.saved-view-tabs { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: var(--space-3); border-bottom: 1px solid var(--line-soft); }
.saved-view-tabs__tabs { min-width: 0; flex: 1 1 auto; }
.saved-view-tabs__actions { display: inline-flex; flex: 0 0 auto; align-items: center; gap: var(--space-2); }
:deep(.el-tabs__header) { margin: 0; }
@media (max-width: 63.9375rem) {
  .saved-view-tabs { align-items: stretch; flex-direction: column; }
  .saved-view-tabs__tabs { overflow-x: auto; }
  .saved-view-tabs__actions :deep(.el-button) { min-height: 44px; }
}
</style>
```

- [ ] **Step 7: 运行测试并提交**

在 `precisionOpsComponents.spec.js` 加入：

```js
const savedViewTabs = readFileSync(new URL('./SavedViewTabs.vue', import.meta.url), 'utf8')

it('owns confirmed saved-view naming and deletion commands', () => {
  expect(savedViewTabs).toContain("ElMessageBox.prompt('请输入视图名称'")
  expect(savedViewTabs).toContain("ElMessageBox.prompt('请输入新的视图名称'")
  expect(savedViewTabs).toContain("ElMessageBox.confirm('确认删除这个保存视图？'")
  expect(savedViewTabs).toContain("emit('save', String(value).trim())")
  expect(savedViewTabs).toContain("emit('rename', { id: activeItem.value.id, label: String(value).trim() })")
  expect(savedViewTabs).toContain("emit('remove', activeItem.value.id)")
  expect(savedViewTabs).not.toContain('localStorage')
})
```

Run: `cd admin-web && npm test -- src/components/base/savedView.helpers.spec.js src/components/base/useSavedViews.spec.js src/components/base/precisionOpsComponents.spec.js`

Expected: PASS。

Commit message: `功能：增加集合保存视图基础能力`

---

### Task 6: Dashboard 运行摘要、内容库存与快捷入口

**Files:**
- Create: `admin-web/src/views/dashboard.helpers.js`
- Create: `admin-web/src/views/dashboard.helpers.spec.js`
- Create: `admin-web/src/views/dashboardPage.spec.js`
- Modify: `admin-web/src/views/Dashboard.vue`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Consumes: Task 3 的 `header-actions` slot 和 Task 4 的 `MetricStrip`。
- Produces: `buildDashboardMetricGroups(stats)`，返回 `{ runtime, inventory }`。
- Produces: `dashboardQuickActions`，只包含 `/upload`、`/tasks`、`/videos`、`/images` 四个现有路由。
- Migration: `/dashboard` 移除 `meta.hideShellPageHeader`，页面不再导入或渲染 `PageHeader`、`StatCard`。

- [ ] **Step 1: 写数据诚实性和页面结构红灯测试**

创建 helper 测试：

```js
import { describe, expect, it } from 'vitest'
import { buildDashboardMetricGroups, dashboardQuickActions } from './dashboard.helpers'

describe('dashboard helpers', () => {
  it('uses only fields supplied by the current stats API', () => {
    const groups = buildDashboardMetricGroups({
      short_videos: 12,
      movie_videos: 3,
      episode_videos: 8,
      av_videos: 4,
      total_users: 5,
      today_uploads: 6,
      queue_length: 2,
      disk_free_bytes: 10737418240
    })
    expect(groups.runtime.map((item) => [item.label, item.value])).toEqual([
      ['今日上传', 6],
      ['转码队列', 2],
      ['磁盘剩余', '10.00 GB'],
      ['总用户数', 5]
    ])
    expect(groups.inventory.map((item) => item.value)).toEqual([12, 3, 8, 4])
    expect(JSON.stringify(groups)).not.toMatch(/失败率|健康|告警/)
  })

  it('links only to existing admin destinations', () => {
    expect(dashboardQuickActions.map((item) => item.path)).toEqual(['/upload', '/tasks', '/videos', '/images'])
  })
})
```

创建页面静态测试：

```js
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./Dashboard.vue', import.meta.url), 'utf8')

describe('Precision Ops dashboard', () => {
  it('uses the workspace header and compact metric hierarchy', () => {
    expect(source).toContain('<template #header-actions>')
    expect(source).toContain('<MetricStrip')
    expect(source).toContain('dashboard-quick-actions')
    expect(source).toContain('height: 240px')
    expect(source).not.toContain('<PageHeader')
    expect(source).not.toContain('<StatCard')
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/dashboard.helpers.spec.js src/views/dashboardPage.spec.js`

Expected: FAIL，helper 和页面测试文件尚不存在，Dashboard 仍为 8 张同权 `StatCard`。

- [ ] **Step 3: 实现 Dashboard 纯映射**

```js
function count(value) {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

function formatDiskFree(value) {
  return `${(count(value) / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function buildDashboardMetricGroups(stats = {}) {
  return {
    runtime: [
      { key: 'today', label: '今日上传', value: count(stats.today_uploads), tone: 'info' },
      { key: 'queue', label: '转码队列', value: count(stats.queue_length) },
      { key: 'disk', label: '磁盘剩余', value: formatDiskFree(stats.disk_free_bytes) },
      { key: 'users', label: '总用户数', value: count(stats.total_users) }
    ],
    inventory: [
      { key: 'short', label: '短视频', value: count(stats.short_videos) },
      { key: 'movie', label: '电影', value: count(stats.movie_videos) },
      { key: 'episode', label: '电视剧集', value: count(stats.episode_videos) },
      { key: 'av', label: 'AV', value: count(stats.av_videos) }
    ]
  }
}

export const dashboardQuickActions = [
  { path: '/upload', label: '上传视频' },
  { path: '/tasks', label: '查看任务' },
  { path: '/videos', label: '管理视频' },
  { path: '/images', label: '管理图片' }
]
```

- [ ] **Step 4: 重组 Dashboard 模板并保留 ECharts 生命周期**

导入 `MetricStrip`、`buildDashboardMetricGroups` 和 `dashboardQuickActions`，删除 `PageHeader`、`StatCard` import。保留现有 `renderChart`、resize 和 dispose 逻辑；指标与加载改为：

```js
const metricGroups = computed(() => buildDashboardMetricGroups(stats.value || {}))

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const nextStats = await getAdminStats()
    stats.value = nextStats
    await nextTick()
    renderChart()
  } catch (error) {
    errorMessage.value = error?.message || '加载仪表盘失败'
  } finally {
    loading.value = false
  }
}
```

刷新失败时不清空已有 `stats`，因此旧数据继续可见并同时显示行内错误。核心模板使用：

```vue
<Layout>
  <template #header-actions>
    <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
  </template>
  <div class="page-shell dashboard-page" data-density="compact">
    <el-alert v-if="errorMessage" type="error" :closable="false" :title="errorMessage">
      <template #default><el-button link type="primary" @click="load">重试</el-button></template>
    </el-alert>
    <el-skeleton v-if="loading && !stats" :rows="7" animated />
    <template v-else-if="stats">
      <MetricStrip :items="metricGroups.runtime" aria-label="运行摘要" />
      <div class="dashboard-workspace">
        <div class="dashboard-workspace__main">
          <section class="dashboard-section">
            <h2>内容库存</h2>
            <MetricStrip :items="metricGroups.inventory" aria-label="内容库存" />
          </section>
          <SectionCard dense>
            <template #title>近 7 天上传趋势</template>
            <div v-if="trendPoints.length" ref="chartRef" class="trend-chart" />
            <EmptyState v-else title="暂无趋势数据" description="后端暂未返回最近 7 天上传趋势" />
          </SectionCard>
        </div>
        <nav class="dashboard-quick-actions" aria-label="常用入口">
          <h2>常用入口</h2>
          <RouterLink v-for="item in dashboardQuickActions" :key="item.path" :to="item.path">{{ item.label }}</RouterLink>
        </nav>
      </div>
    </template>
  </div>
</Layout>
```

CSS 在桌面使用 `grid-template-columns: minmax(0, 1fr) 240px`，趋势图 `height: 240px`；`<1024px` 改成单列。不得增加健康分、失败率、告警卡或拖拽布局。

- [ ] **Step 5: 移除 Dashboard 兼容 meta 并验证**

把路由改为：

```js
{ path: '/dashboard', component: Dashboard }
```

在 `router/index.spec.js` 断言 `/dashboard` 行不含 `hideShellPageHeader`。

Run: `cd admin-web && npm test -- src/views/dashboard.helpers.spec.js src/views/dashboardPage.spec.js src/router/index.spec.js src/components/Layout.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级管理端仪表盘`

---

### Task 7: TaskMonitor 诚实统计与非阻断刷新

**Files:**
- Modify: `admin-web/src/views/TaskMonitor.vue`
- Modify: `admin-web/src/views/taskMonitorPage.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Consumes: `MetricStrip`、`StatusIndicator`、Layout `header-actions`。
- Preserves: `loadSeq` 旧响应保护和 5 秒 `skipIfLoading` 自动刷新。
- Produces: `initialLoading`、`backgroundRefreshing`、`loadError`、`summaryMetrics`。
- Migration: `/tasks` 移除 `meta.hideShellPageHeader`；移除 `successRate` 和 4 张 `StatCard`。

- [ ] **Step 1: 扩展 TaskMonitor 红灯测试**

```js
it('removes the mixed-scope success rate and labels every summary scope', () => {
  expect(taskMonitor).not.toContain('successRate')
  expect(taskMonitor).not.toContain('<StatCard')
  expect(taskMonitor).toContain("label: '任务总量'")
  expect(taskMonitor).toContain("scope: '全局'")
  expect(taskMonitor).toContain("scope: query.status ? '当前筛选·全部页' : '全局'")
  expect(taskMonitor).toContain("label: '排队'")
  expect(taskMonitor).toContain("scope: '本页'")
})

it('keeps existing rows visible during background refresh', () => {
  expect(taskMonitor).toContain('const initialLoading = computed(() => loading.value && !loaded.value)')
  expect(taskMonitor).toContain('const backgroundRefreshing = computed(() => loading.value && loaded.value)')
  expect(taskMonitor).toContain('if (seq === loadSeq) loaded.value = true')
  expect(taskMonitor).not.toContain('<el-table v-loading="loading"')
  expect(taskMonitor).not.toContain('if (hadRows) ElMessage.error')
  expect(taskMonitor).toContain("loadError.value = error?.message || '加载任务失败'")
})

it('offers all five status filters and compact task rows', () => {
  expect(taskMonitor).toContain("{ label: '已完成', value: 'success' }")
  expect(taskMonitor).toContain('data-density="monitor"')
  expect(taskMonitor).toContain(':stroke-width="6"')
  expect(taskMonitor).toContain('class="task-error"')
  expect(taskMonitor).toContain('tabindex="0"')
  expect(taskMonitor).toContain(':aria-label="row.error || \'无错误\'"')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/taskMonitorPage.spec.js`

Expected: FAIL，页面仍计算成功率、使用 4 张卡并用全表 loading 遮罩刷新。

- [ ] **Step 3: 重写派生状态而不改变 API 参数**

删除 `PageHeader`、`StatCard` import，新增 `../components/base/MetricStrip.vue` 和 `../components/base/StatusIndicator.vue` import。删除 `successCount` 和 `successRate`，增加：

```js
const loaded = ref(false)
const loadError = ref('')
const initialLoading = computed(() => loading.value && !loaded.value)
const backgroundRefreshing = computed(() => loading.value && loaded.value)
const hasStatusFilter = computed(() => query.status !== '')
const statusOptions = [
  { label: '全部', value: '' },
  { label: '排队', value: 'pending' },
  { label: '处理中', value: 'running' },
  { label: '已完成', value: 'success' },
  { label: '失败', value: 'failed' }
]
const summaryMetrics = computed(() => [
  { key: 'total', label: '任务总量', value: total.value, scope: query.status ? '当前筛选·全部页' : '全局' },
  { key: 'queued', label: '排队', value: queuedCount.value, scope: '本页', tone: 'info' },
  { key: 'running', label: '处理中', value: runningCount.value, scope: '本页', tone: 'warning' },
  { key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' }
])

function taskStatusTone(status) {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'warning'
  if (status === 'pending') return 'info'
  return 'neutral'
}

async function load(options = {}) {
  const { skipIfLoading = false } = options
  if (skipIfLoading && loading.value) return
  const seq = ++loadSeq
  loading.value = true
  loadError.value = ''
  try {
    const params = { page: query.page, page_size: query.page_size }
    if (query.status) params.status = query.status
    const data = await getAdminTasks(params)
    if (seq !== loadSeq) return
    list.value = data.items || []
    total.value = data.total_count || 0
  } catch (error) {
    if (seq !== loadSeq) return
    loadError.value = error?.message || '加载任务失败'
  } finally {
    if (seq === loadSeq) {
      loaded.value = true
      loading.value = false
    }
  }
}
```

上述 `load` 继续使用 `loadSeq`，不会让旧请求覆盖新结果；`loaded` 只由最新请求在 finally 置为 true，因此首次成功为空后自动刷新仍保留空态而不反复切回骨架。首次和后台刷新失败都写入 `loadError`，已有列表继续在错误提示下方渲染，不能只用瞬时 toast。API 的 `total_count` 会随状态筛选变化，因此筛选时范围必须显示“当前筛选·全部页”，不能误称全局。

- [ ] **Step 4: 改造模板、状态和错误单元格**

```vue
<Layout>
  <template #header-actions>
    <el-button :loading="loading" @click="load">立即刷新</el-button>
  </template>
  <div class="page-shell task-monitor-page" data-density="monitor">
    <Toolbar dense>
      <template #filters>
        <el-segmented :model-value="query.status" :options="statusOptions" @update:model-value="setStatus" />
      </template>
      <template #actions>
        <StatusIndicator :label="backgroundRefreshing ? '正在刷新' : '每 5 秒自动刷新'" :tone="backgroundRefreshing ? 'info' : 'neutral'" />
      </template>
    </Toolbar>
    <MetricStrip :items="summaryMetrics" aria-label="任务摘要" />
    <el-alert v-if="loadError" type="error" :closable="false" :title="loadError">
      <template #default><el-button link type="primary" @click="load">重试</el-button></template>
    </el-alert>
    <el-skeleton v-if="initialLoading" :rows="12" animated />
    <SectionCard v-else-if="!loadError || list.length > 0" dense>
      <template #title>任务列表</template>
      <EmptyState
        v-if="list.length === 0"
        :title="hasStatusFilter ? '当前筛选无结果' : '暂无任务'"
        :description="hasStatusFilter ? '清除状态筛选后查看全部任务' : '任务创建后会显示在这里'"
      >
        <template v-if="hasStatusFilter" #action><el-button @click="setStatus('')">清除筛选</el-button></template>
      </EmptyState>
      <div v-else class="table-wrap">
        <el-table :data="list" border>
          <el-table-column prop="video_title" label="任务" min-width="260">
            <template #default="{ row }">
              <div class="task-cell">
                <strong>{{ taskTitle(row) }}</strong>
                <span>任务 ID：{{ row.id }} · 视频 ID：{{ row.video_id || '--' }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="112">
            <template #default="{ row }">
              <StatusIndicator :label="statusLabel(row.status)" :tone="taskStatusTone(row.status)" />
            </template>
          </el-table-column>
          <el-table-column label="进度" min-width="190">
            <template #default="{ row }">
              <el-progress :stroke-width="6" :percentage="resolveProgress(row)" :status="progressStatus(row)" />
            </template>
          </el-table-column>
          <el-table-column label="剩余时间" width="112"><template #default="{ row }">{{ formatRemaining(row) }}</template></el-table-column>
          <el-table-column label="已耗时" width="112"><template #default="{ row }">{{ formatElapsed(row) }}</template></el-table-column>
          <el-table-column prop="retry_count" label="重试" width="72" />
          <el-table-column prop="error" label="错误" min-width="220">
            <template #default="{ row }">
              <el-tooltip :content="row.error || '无错误'" placement="top">
                <span class="task-error" tabindex="0" :aria-label="row.error || '无错误'">{{ row.error || '--' }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="开始时间" width="168"><template #default="{ row }">{{ formatDateTime(row.started_at) }}</template></el-table-column>
          <el-table-column label="进度更新时间" width="168"><template #default="{ row }">{{ formatDateTime(row.progress_updated_at) }}</template></el-table-column>
        </el-table>
      </div>
    </SectionCard>
  </div>
</Layout>
```

实施时保留现有完整任务、状态、进度、剩余时间、已耗时、重试、开始时间和进度更新时间列，不删除任何列。状态列使用 `StatusIndicator`；任务标题与 ID 保持两行，行高 44px；`.task-error` 使用 `display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap`，tooltip 提供 hover 全文，`tabindex` 与 `aria-label` 提供键盘焦点和完整可访问名称。

- [ ] **Step 5: 移除路由 meta、运行测试并提交**

把路由改为：

```js
{ path: '/tasks', component: TaskMonitor }
```

Run: `cd admin-web && npm test -- src/views/taskMonitorPage.spec.js src/router/index.spec.js src/components/base/precisionOpsComponents.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级任务监控信息层级`

---

### Task 8: VideoList 保存视图、紧凑媒体表格与行操作菜单

**Files:**
- Modify: `admin-web/src/views/videoList.helpers.js`
- Modify: `admin-web/src/views/videoList.helpers.spec.js`
- Create: `admin-web/src/views/videoListPage.spec.js`
- Modify: `admin-web/src/views/VideoList.vue`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`
- Modify: `CONTEXT.md`

**Interfaces:**
- Consumes: `SavedViewTabs`、`StatusIndicator`、Task 5 `useSavedViews`、Layout `header-actions`。
- Produces: `createVideoBuiltInViews(defaultColumns)`、`normalizeVideoViewSnapshot(snapshot, allowedColumns, defaultColumns)`。
- Storage: `admin-videolist-saved-views-v1`，文档 `{ version: 1, items }`；继续同步现有 `admin-videolist-columns`。
- Preserves: Shift 当前页区间选择、批量编辑/删除、1280px 次要列隐藏、560px Drawer、字幕和脏数据守卫。

- [ ] **Step 1: 写视图快照和页面结构红灯测试**

在 `videoList.helpers.spec.js` 增加：

```js
import { createVideoBuiltInViews, normalizeVideoViewSnapshot } from './videoList.helpers'

it('builds only API-supported video views and excludes pagination', () => {
  const columns = ['title', 'thumbnail', 'status', 'operations']
  expect(createVideoBuiltInViews(columns).map((item) => [item.id, item.label, item.snapshot.status])).toEqual([
    ['builtin-all', '全部视频', ''],
    ['builtin-processing', '处理中', 'processing'],
    ['builtin-failed', '失败', 'failed']
  ])
  expect(normalizeVideoViewSnapshot({ q: 'A', type: 'movie', status: 'ready', page: 9, columns }, columns, columns)).toEqual({
    q: 'A', type: 'movie', status: 'ready', columns
  })
})
```

创建 `videoListPage.spec.js`：

```js
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./VideoList.vue', import.meta.url), 'utf8')

describe('Precision Ops video list', () => {
  it('uses saved views and the merged workspace header', () => {
    expect(source).toContain('admin-videolist-saved-views-v1')
    expect(source).toContain('useSavedViews')
    expect(source).toContain('<SavedViewTabs')
    expect(source).toContain('<template #header-actions>')
    expect(source).not.toContain('<PageHeader')
    expect(source).not.toContain('function persistUserViews')
    expect(source).not.toContain("ElMessageBox.prompt('请输入视图名称'")
  })

  it('keeps detail visible and moves low-frequency actions into a menu', () => {
    expect(source).toContain('aria-label="更多视频操作"')
    expect(source).toContain('>详情</el-button>')
    expect(source).toContain('command="retranscode"')
    expect(source).toContain('command="delete"')
    expect(source).toContain('width="108"')
  })

  it('uses fixed compact media geometry', () => {
    expect(source).toContain('data-density="compact"')
    expect(source).toContain('has-media-rows')
    expect(source).toContain('width: 72px')
    expect(source).toContain('height: 40px')
    expect(source).not.toMatch(/rgba?\(\s*\d/)
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/videoList.helpers.spec.js src/views/videoListPage.spec.js`

Expected: FAIL，页面没有保存视图，操作列仍为 300px 且三个按钮全部常驻。

- [ ] **Step 3: 实现视频视图快照纯函数**

在 `videoList.helpers.js` 导出：

```js
export function normalizeVideoViewSnapshot(snapshot, allowedColumns, defaultColumns) {
  const allowed = new Set(allowedColumns)
  const requested = Array.isArray(snapshot?.columns) ? snapshot.columns : defaultColumns
  const columns = requested.filter((key, index) => allowed.has(key) && requested.indexOf(key) === index)
  if (allowed.has('operations') && !columns.includes('operations')) columns.push('operations')
  return {
    q: String(snapshot?.q || ''),
    type: String(snapshot?.type || ''),
    status: String(snapshot?.status || ''),
    columns: columns.length > 0 ? columns : [...defaultColumns]
  }
}

export function createVideoBuiltInViews(defaultColumns) {
  const snapshot = (status) => ({ q: '', type: '', status, columns: [...defaultColumns] })
  return [
    { id: 'builtin-all', label: '全部视频', builtIn: true, snapshot: snapshot('') },
    { id: 'builtin-processing', label: '处理中', builtIn: true, snapshot: snapshot('processing') },
    { id: 'builtin-failed', label: '失败', builtIn: true, snapshot: snapshot('failed') }
  ]
}
```

- [ ] **Step 4: 通过 useSavedViews 接入版本化保存视图**

在 `VideoList.vue` 的查询、列状态和列常量就绪后实例化共享 composable。页面只定义 storage key、内置视图、快照规范、应用快照和刷新；名称输入、删除确认、storage 解析、选中态、自定义态和 CRUD 生命周期不得在页面重复实现：

```js
const SAVED_VIEWS_KEY = 'admin-videolist-saved-views-v1'
const listError = ref('')
const hasActiveFilters = computed(() => String(query.q || '').trim() !== '' || query.type !== '' || query.status !== '')
const allowedColumnKeys = ALL_COLUMNS.map((item) => item.key)
const builtInViews = createVideoBuiltInViews(DEFAULT_VISIBLE_COLUMNS)

const {
  availableViews,
  activeViewId,
  editableSourceId,
  selectView,
  saveView,
  updateView,
  renameView,
  removeView
} = useSavedViews({
  storageKey: SAVED_VIEWS_KEY,
  builtInViews,
  normalizeSnapshot: (snapshot) => normalizeVideoViewSnapshot(
    snapshot,
    allowedColumnKeys,
    DEFAULT_VISIBLE_COLUMNS
  ),
  getCurrentSnapshot: () => ({
    q: query.q,
    type: query.type,
    status: query.status,
    columns: columnVisibility.value
  }),
  applySnapshot: (snapshot) => {
    const next = normalizeVideoViewSnapshot(snapshot, allowedColumnKeys, DEFAULT_VISIBLE_COLUMNS)
    query.q = next.q
    query.type = next.type
    query.status = next.status
    query.page = 1
    columnVisibility.value = [...next.columns]
    persistColumns()
    clearSelection()
  },
  refresh: load
})

function videoStatusTone(status) {
  return getVideoStatusMeta(status).tagType || 'neutral'
}

async function load() {
  listLoading.value = true
  listError.value = ''
  try {
    const data = await getAdminVideos(query)
    list.value = data.items || []
    total.value = data.total_count || 0
    clearSelection()
  } catch (error) {
    listError.value = error?.message || '加载视频列表失败'
  } finally {
    listLoading.value = false
  }
}
```

从 Task 5 只导入 `useSavedViews`，同时导入 `SavedViewTabs.vue` 和 `StatusIndicator.vue`，删除 `PageHeader` import。页面不得导入保存视图 helper，不得出现 `readUserViews`、`persistUserViews`、`selectedViewID` 或保存视图专用 `ElMessageBox.prompt/confirm`。切换保存视图通过 `applySnapshot` 调用现有 `clearSelection()`；页码、选择和 Drawer 状态不得进入快照。
- [ ] **Step 5: 重组集合工具条、加载状态和行操作**

页面根容器增加 `data-density="compact"`，移除 `PageHeader`。在 Layout 头部放列设置和上传入口，在集合区依次放保存视图、搜索/筛选工具条、表格、分页和批量条：

```vue
<Layout>
  <template #header-actions>
    <el-popover trigger="click" :width="240">
      <template #reference><el-button :icon="Setting">列设置</el-button></template>
      <el-checkbox-group :model-value="columnVisibility" class="column-settings" @update:model-value="onColumnVisibilityChange">
        <el-checkbox v-for="column in ALL_COLUMNS" :key="column.key" :value="column.key" :disabled="column.required">{{ column.label }}</el-checkbox>
      </el-checkbox-group>
    </el-popover>
    <el-button type="primary" @click="router.push('/upload')">上传视频</el-button>
  </template>
  <div class="page-shell video-list-page" data-density="compact">
    <SavedViewTabs
      :items="availableViews"
      :active-id="activeViewId"
      :editable-source-id="editableSourceId"
      @select="selectView"
      @save="saveView"
      @update="updateView"
      @rename="renameView"
      @remove="removeView"
    />
    <Toolbar dense>
      <template #filters>
        <el-input
          v-model="query.q"
          class="quick-search"
          placeholder="标题/标签搜索"
          clearable
          :prefix-icon="Search"
          @keyup.enter="applyFilters"
          @clear="applyFilters"
        />
        <el-tag v-for="chip in activeFilterChips" :key="chip.key" closable @close="removeFilter(chip.key)">
          {{ chip.label }}：{{ chip.value }}
        </el-tag>
        <el-button plain @click="filterDrawerVisible = true">更多筛选</el-button>
      </template>
    </Toolbar>
    <el-alert v-if="listError" type="error" :closable="false" :title="listError"><template #default><el-button link type="primary" @click="load">重试</el-button></template></el-alert>
    <el-skeleton v-if="listLoading && list.length === 0" :rows="10" animated />
    <SectionCard v-else-if="!listError || list.length > 0" dense>
      <template #title>视频列表</template>
      <template #actions><span class="result-total">共 {{ total }} 条</span></template>
      <EmptyState
        v-if="list.length === 0"
        :title="hasActiveFilters ? '当前筛选无结果' : '暂无视频'"
        :description="hasActiveFilters ? '清除筛选后查看全部视频' : '上传视频后会显示在这里'"
      >
        <template v-if="hasActiveFilters" #action><el-button @click="resetFilters">清除筛选</el-button></template>
      </EmptyState>
      <div v-else class="table-wrap has-media-rows">
        <el-table
          ref="tableRef"
          :data="list"
          row-key="id"
          border
          @selection-change="onSelectionChange"
          @select="onRowSelectionSelect"
          @select-all="onSelectionSelectAll"
        >
          <el-table-column type="selection" width="44" />
          <el-table-column v-if="isColumnVisible('title')" prop="title" label="标题" min-width="220" show-overflow-tooltip />
          <el-table-column v-if="isColumnVisible('thumbnail')" label="封面" width="96">
            <template #default="{ row }">
              <div class="video-cover-cell">
                <el-image v-if="shouldShowVideoThumbnail(row)" :src="getVideoThumbnailURL(row)" fit="cover" class="video-cover-image">
                  <template #error><div class="video-cover-placeholder">暂无封面</div></template>
                </el-image>
                <div v-else class="video-cover-placeholder">{{ getVideoThumbnailPlaceholder(row) }}</div>
              </div>
            </template>
          </el-table-column>
          <el-table-column v-if="isColumnVisible('type')" prop="type" label="类型" width="96">
            <template #default="{ row }">{{ typeLabel(row.type) }}</template>
          </el-table-column>
          <el-table-column v-if="isColumnVisible('status')" prop="status" label="状态" width="120">
            <template #default="{ row }"><StatusIndicator :label="statusLabel(row.status)" :tone="videoStatusTone(row.status)" /></template>
          </el-table-column>
          <el-table-column v-if="isColumnVisible('upload_user')" prop="upload_user" label="上传用户" width="128" show-overflow-tooltip />
          <el-table-column v-if="isColumnVisible('created_at')" label="上传时间" width="168">
            <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
          </el-table-column>
          <el-table-column v-if="isColumnVisible('operations')" label="操作" width="108" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="showDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(command) => command === 'retranscode' ? doRetranscode(row) : doDelete(row)">
                <el-tooltip content="更多视频操作" placement="top">
                  <el-button :icon="MoreFilled" circle aria-label="更多视频操作" />
                </el-tooltip>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="retranscode">重新转码</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="toolbar-row toolbar-row--end">
        <AdminTablePagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          layout="total, prev, pager, next"
          :total="total"
          @current-change="load"
        />
      </div>
    </SectionCard>
    <BulkActionBar :count="selectedRows.length" :actions="bulkActions" />
  </div>
</Layout>
```

CSS 固定 `.video-cover-cell` 和 `.video-cover-image` 为 `72px × 40px`，媒体行由 `.has-media-rows` 使用 52px token；现有直接数字 `rgba(...)` 状态底色改为对应语义 token 的 `color-mix(...)`。详情 Drawer 保持 560px/窄屏 100%、现有脏数据确认和固定 footer；只用分组标题、预览层级和间距整理，不改保存 payload。

- [ ] **Step 6: 移除路由 meta、记录 key、验证并提交**

把路由改为 `{ path: '/videos', component: VideoList }`。在 `CONTEXT.md` 的保存视图术语补充两个 key 及同步规则：`admin-videolist-saved-views-v1` 保存版本化视图，应用视图时同步 `admin-videolist-columns`；手动改列只进入“自定义”，显式更新后才覆盖快照。

Run: `cd admin-web && npm test -- src/views/videoList.helpers.spec.js src/views/videoListPage.spec.js src/components/base/savedView.helpers.spec.js src/components/base/useSavedViews.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级视频资源集合页`

---

### Task 9: ImageManage 保存视图、高密度资产网格与检查器

**Files:**
- Create: `admin-web/src/views/imageManage.helpers.js`
- Create: `admin-web/src/views/imageManage.helpers.spec.js`
- Create: `admin-web/src/views/imageManagePage.spec.js`
- Modify: `admin-web/src/views/ImageManage.vue`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`
- Modify: `CONTEXT.md`

**Interfaces:**
- Consumes: `SavedViewTabs`、`MetricStrip`、`StatusIndicator`、Task 5 `useSavedViews`、Layout `header-actions`。
- Produces: `DEFAULT_IMAGE_ACTIVE = '1'`、`createImageBuiltInViews(defaultActive, defaultViewMode)`、`normalizeImageViewSnapshot(snapshot)`、`hasImageActiveFilters(snapshot)`。
- Storage: `admin-imagemanage-saved-views-v1`；应用视图时同步既有 `admin-imagemanage-view`。
- Preserves: 上传队列、秒传预检、批量启停/删除、详情/上传脏数据守卫、560px Drawer、路由 query 打开详情。

- [ ] **Step 1: 写图片快照和页面结构红灯测试**

```js
import { describe, expect, it } from 'vitest'
import {
  DEFAULT_IMAGE_ACTIVE,
  createImageBuiltInViews,
  hasImageActiveFilters,
  normalizeImageViewSnapshot
} from './imageManage.helpers'

describe('image manage view helpers', () => {
  it('builds supported built-ins and excludes page state', () => {
    expect(createImageBuiltInViews(DEFAULT_IMAGE_ACTIVE, 'list').map((item) => [item.label, item.snapshot.status, item.snapshot.active, item.snapshot.viewMode])).toEqual([
      ['全部图片', '', '1', 'list'],
      ['可用', 'ready', '1', 'list'],
      ['失败', 'failed', '1', 'list']
    ])
    expect(normalizeImageViewSnapshot({ q: 'A', status: 'ready', active: '0', actor_id: 'a', collection_id: 'c', viewMode: 'list', page: 4 })).toEqual({
      q: 'A', status: 'ready', active: '0', actor_id: 'a', collection_id: 'c', viewMode: 'list'
    })
  })

  it('treats the existing active default as baseline instead of a user filter', () => {
    const baseline = { q: '', status: '', active: DEFAULT_IMAGE_ACTIVE, actor_id: '', collection_id: '' }
    expect(hasImageActiveFilters(baseline)).toBe(false)
    expect(hasImageActiveFilters({ ...baseline, active: '0' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, active: '' })).toBe(true)
  })
})
```

```js
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ImageManage.vue', import.meta.url), 'utf8')

describe('Precision Ops image manage', () => {
  it('uses saved views, metrics and workspace actions', () => {
    expect(source).toContain('admin-imagemanage-saved-views-v1')
    expect(source).toContain('useSavedViews')
    expect(source).toContain('<SavedViewTabs')
    expect(source).toContain('<MetricStrip')
    expect(source).toContain('<template #header-actions>')
    expect(source).not.toContain('<PageHeader')
    expect(source).not.toContain('function persistUserViews')
    expect(source).not.toContain("ElMessageBox.prompt('请输入视图名称'")
    expect(source).toContain("label: '结果总数'")
    expect(source).toContain("scope: '当前条件·全部页'")
  })

  it('keeps asset controls visible and preserves full images', () => {
    expect(source).toContain('aria-label="图片操作"')
    expect(source).toContain(':aria-label="`选择图片：${item.title || item.id}`"')
    expect(source).toContain('minmax(184px, 1fr)')
    expect(source).toContain('object-fit: contain')
    expect(source).not.toMatch(/image-grid-card__actions[\s\S]{0,160}opacity:\s*0/)
  })
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/imageManage.helpers.spec.js src/views/imageManagePage.spec.js`

Expected: FAIL，helper 不存在，网格仍为 220px 且操作只在 hover 后出现。

- [ ] **Step 3: 实现图片视图纯函数**

```js
export const DEFAULT_IMAGE_ACTIVE = '1'

export function normalizeImageViewSnapshot(snapshot) {
  return {
    q: String(snapshot?.q || ''),
    status: String(snapshot?.status || ''),
    active: ['0', '1'].includes(String(snapshot?.active)) ? String(snapshot.active) : '',
    actor_id: String(snapshot?.actor_id || ''),
    collection_id: String(snapshot?.collection_id || ''),
    viewMode: snapshot?.viewMode === 'list' ? 'list' : 'grid'
  }
}

export function createImageBuiltInViews(defaultActive = DEFAULT_IMAGE_ACTIVE, defaultViewMode = 'grid') {
  const snapshot = (status) => normalizeImageViewSnapshot({
    status,
    active: defaultActive,
    viewMode: defaultViewMode
  })
  return [
    { id: 'builtin-all', label: '全部图片', builtIn: true, snapshot: snapshot('') },
    { id: 'builtin-ready', label: '可用', builtIn: true, snapshot: snapshot('ready') },
    { id: 'builtin-failed', label: '失败', builtIn: true, snapshot: snapshot('failed') }
  ]
}

export function hasImageActiveFilters(snapshot) {
  return String(snapshot?.q || '').trim() !== ''
    || String(snapshot?.status || '') !== ''
    || String(snapshot?.active) !== DEFAULT_IMAGE_ACTIVE
    || String(snapshot?.actor_id || '') !== ''
    || String(snapshot?.collection_id || '') !== ''
}
```

- [ ] **Step 4: 通过 useSavedViews 接入保存视图并同步既有网格偏好**

保留 `query.active` 的现有初始语义，不扩大默认结果集：把查询初始值、`resetFilters` 和 `removeFilter('active')` 的 `'1'` 统一替换为 `DEFAULT_IMAGE_ACTIVE`，并以当前 `active` 和既有 `admin-imagemanage-view` 模式作为内置视图基线。页面只提供 storage key、内置视图、快照规范、应用快照和业务刷新：

```js
const SAVED_VIEWS_KEY = 'admin-imagemanage-saved-views-v1'
const builtInViews = createImageBuiltInViews(query.active, viewMode.value)

const {
  availableViews,
  activeViewId,
  editableSourceId,
  selectView,
  saveView,
  updateView,
  renameView,
  removeView
} = useSavedViews({
  storageKey: SAVED_VIEWS_KEY,
  builtInViews,
  normalizeSnapshot: normalizeImageViewSnapshot,
  getCurrentSnapshot: () => ({
    q: query.q,
    status: query.status,
    active: query.active,
    actor_id: query.actor_id,
    collection_id: query.collection_id,
    viewMode: viewMode.value
  }),
  applySnapshot: applyImageViewSnapshot,
  refresh: load
})

function applyImageViewSnapshot(snapshot) {
  const next = normalizeImageViewSnapshot(snapshot)
  query.q = next.q
  query.status = next.status
  query.active = next.active
  query.actor_id = next.actor_id
  query.collection_id = next.collection_id
  query.page = 1
  setViewMode(next.viewMode)
  clearImageSelection()
}
```

从 Task 5 只导入 `useSavedViews`；导入 `SavedViewTabs.vue`、`MetricStrip.vue`、`StatusIndicator.vue`，在 Element 图标 import 中加入 `MoreFilled`，并删除 `PageHeader` import。页面不得导入保存视图 helper，不得出现 `readUserViews`、`persistUserViews`、`selectedViewID` 或保存视图专用 `ElMessageBox.prompt/confirm`。`setViewMode` 继续写 `admin-imagemanage-view`。手动切换网格/列表只改变当前快照并使 tab 进入“自定义”；只有用户触发保存或更新时才由 composable 写 `admin-imagemanage-saved-views-v1`。

- [ ] **Step 5: 重组指标、工具条、空态和资产卡操作**

移除 `PageHeader`，根容器增加 `data-density="compact"`；上传按钮进入 `header-actions`。增加明确的列表错误状态和筛选判定：

```js
const listError = ref('')
const hasActiveFilters = computed(() => hasImageActiveFilters(query))

async function load() {
  loading.value = true
  listError.value = ''
  try {
    const data = await getAdminImages(buildListParams())
    list.value = data.items || []
    total.value = data.total_count || 0
    clearImageSelection()
  } catch (error) {
    listError.value = extractErrorMessage(error, '加载图片列表失败')
  } finally {
    loading.value = false
  }
}
```

用以下指标替换手写 `stats-strip`：

```js
const summaryMetrics = computed(() => [
  { key: 'total', label: '结果总数', value: total.value, scope: '当前条件·全部页' },
  { key: 'ready', label: '可用', value: readyCount.value, scope: '本页', tone: 'success' },
  { key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' },
  { key: 'inactive', label: '停用', value: inactiveCount.value, scope: '本页', tone: 'warning' }
])
```

模板顺序固定为 `SavedViewTabs → MetricStrip → Toolbar → 行内错误/骨架 → 网格或表格 → 分页 → BulkActionBar`：

```vue
<Layout>
  <template #header-actions>
    <el-button :icon="Upload" type="primary" @click="openUploadDialog">上传图片</el-button>
  </template>
  <div class="page-shell image-page" data-density="compact">
    <SavedViewTabs
      :items="availableViews"
      :active-id="activeViewId"
      :editable-source-id="editableSourceId"
      @select="selectView"
      @save="saveView"
      @update="updateView"
      @rename="renameView"
      @remove="removeView"
    />
    <MetricStrip :items="summaryMetrics" aria-label="图片摘要" />
    <Toolbar dense>
      <template #filters>
        <el-input v-model="query.q" class="quick-search" placeholder="搜索图片" clearable :prefix-icon="Search" @keyup.enter="load" @clear="load" />
        <el-tag v-for="chip in activeFilterChips" :key="chip.key" closable @close="removeFilter(chip.key)">{{ chip.label }}：{{ chip.value }}</el-tag>
        <el-button plain @click="filterDrawerVisible = true">更多筛选</el-button>
      </template>
      <template #actions>
        <el-segmented
          :model-value="viewMode"
          :options="[{ label: '网格', value: 'grid' }, { label: '列表', value: 'list' }]"
          @update:model-value="setViewMode"
        />
      </template>
    </Toolbar>
    <el-alert v-if="listError" type="error" :closable="false" :title="listError"><template #default><el-button link type="primary" @click="load">重试</el-button></template></el-alert>
    <el-skeleton v-if="loading && list.length === 0" :rows="12" animated />
    <SectionCard v-else-if="!listError || list.length > 0" dense>
      <EmptyState
        v-if="list.length === 0"
        :title="hasActiveFilters ? '当前筛选无结果' : '暂无图片'"
        :description="hasActiveFilters ? '清除筛选后查看全部图片' : '上传图片后会显示在这里'"
      >
        <template #action>
          <el-button v-if="hasActiveFilters" @click="resetFilters">清除筛选</el-button>
          <el-button v-else type="primary" :icon="Upload" @click="openUploadDialog">上传图片</el-button>
        </template>
      </EmptyState>
      <div v-else-if="viewMode === 'grid'" class="image-grid">
        <article v-for="item in list" :key="item.id" class="image-grid-card" :class="{ 'is-selected': isGridSelected(item) }">
          <el-checkbox class="image-grid-card__select" :aria-label="`选择图片：${item.title || item.id}`" :model-value="isGridSelected(item)" @update:model-value="(checked) => toggleGridSelection(item, checked)" />
          <div class="image-grid-card__preview">
            <img v-if="item.view_url || item.url || item.thumbnail_url" :src="item.view_url || item.url || item.thumbnail_url" :alt="item.title || '图片预览'" />
            <span v-else>{{ item.title || '图片' }}</span>
          </div>
          <div class="image-grid-card__body">
            <strong>{{ item.title || item.id }}</strong>
            <span>{{ item.width || 0 }} × {{ item.height || 0 }} · {{ formatFileSize(item.file_size) }}</span>
          </div>
          <div class="image-grid-card__meta">
            <StatusIndicator :label="statusLabel(item.status)" :tone="item.status === 'failed' ? 'danger' : 'success'" />
            <StatusIndicator :label="item.active ? '启用' : '停用'" :tone="item.active ? 'success' : 'warning'" />
          </div>
          <div class="image-grid-card__actions">
            <el-button link type="primary" @click="showDetail(item)">详情</el-button>
            <el-dropdown trigger="click" @command="(command) => command === 'toggle' ? toggleActive(item) : doDelete(item)">
              <el-tooltip content="图片操作" placement="top"><el-button :icon="MoreFilled" circle aria-label="图片操作" /></el-tooltip>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="toggle">{{ item.active ? '停用' : '启用' }}</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </article>
      </div>
      <div v-else class="table-wrap">
        <el-table ref="tableRef" :data="list" border @selection-change="onImageSelectionChange">
          <el-table-column type="selection" width="44" />
          <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="104"><template #default="{ row }"><StatusIndicator :label="statusLabel(row.status)" :tone="row.status === 'failed' ? 'danger' : 'success'" /></template></el-table-column>
          <el-table-column label="启用" width="88"><template #default="{ row }">{{ row.active ? '是' : '否' }}</template></el-table-column>
          <el-table-column prop="stored_mime" label="格式" width="104" />
          <el-table-column label="尺寸" width="120"><template #default="{ row }">{{ row.width || 0 }} × {{ row.height || 0 }}</template></el-table-column>
          <el-table-column label="文件大小" width="120"><template #default="{ row }">{{ formatFileSize(row.file_size) }}</template></el-table-column>
          <el-table-column label="上传时间" width="168"><template #default="{ row }">{{ formatDateTime(row.created_at) }}</template></el-table-column>
          <el-table-column label="操作" width="108" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="showDetail(row)">详情</el-button><el-dropdown trigger="click" @command="(command) => command === 'toggle' ? toggleActive(row) : doDelete(row)"><el-tooltip content="图片操作" placement="top"><el-button :icon="MoreFilled" circle aria-label="图片操作" /></el-tooltip><template #dropdown><el-dropdown-menu><el-dropdown-item command="toggle">{{ row.active ? '停用' : '启用' }}</el-dropdown-item><el-dropdown-item command="delete" divided>删除</el-dropdown-item></el-dropdown-menu></template></el-dropdown></template></el-table-column>
        </el-table>
      </div>
      <div class="toolbar-row toolbar-row--end">
        <AdminTablePagination v-model:current-page="query.page" v-model:page-size="query.page_size" layout="total, prev, pager, next" :total="total" @current-change="load" />
      </div>
    </SectionCard>
    <BulkActionBar :count="selectedImageRows.length" :actions="bulkActions" />
  </div>
</Layout>
```

真正空态提供上传，筛选零结果提供 `resetFilters`。图片卡动作始终可见；上面的网格和列表都使用同一详情入口与更多菜单。

```css
.image-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(184px, 1fr)); gap: 12px; }
.image-grid-card { padding: 8px; border-radius: var(--radius-md); }
.image-grid-card__preview { aspect-ratio: 4 / 3; background-color: var(--bg-surface-muted); background-image: linear-gradient(45deg, var(--line-soft) 25%, transparent 25%), linear-gradient(-45deg, var(--line-soft) 25%, transparent 25%), linear-gradient(45deg, transparent 75%, var(--line-soft) 75%), linear-gradient(-45deg, transparent 75%, var(--line-soft) 75%); background-size: 16px 16px; background-position: 0 0, 0 8px, 8px -8px, -8px 0; }
.image-grid-card__preview img { width: 100%; height: 100%; object-fit: contain; }
.image-grid-card__actions { display: flex; align-items: center; justify-content: space-between; gap: var(--space-2); opacity: 1; }
```

列表模式的详情按钮始终可见，停用/启用和删除收入同类更多菜单。Drawer 保持 560px、预览中密度、现有 zoom/fit 参数、上传队列和脏数据保护。

- [ ] **Step 6: 移除路由 meta、记录 key、验证并提交**

把路由改为 `{ path: '/images', component: ImageManage }`。在 `CONTEXT.md` 记录 `admin-imagemanage-saved-views-v1` 与 `admin-imagemanage-view` 的同步关系，以及手动模式切换不静默覆盖保存视图。

Run: `cd admin-web && npm test -- src/views/imageManage.helpers.spec.js src/views/imageManagePage.spec.js src/components/base/savedView.helpers.spec.js src/components/base/useSavedViews.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级图片资产集合页`

---

### Task 10: 第一阶段集成门禁

**Files:**
- Create: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `plan.md`

**Interfaces:**
- Verifies: 第一阶段四个样板页均迁移工作区页头并声明正确密度。
- Verifies: 其余 13 个 shell 页面仍保留 `PageHeader + hideShellPageHeader` 兼容边界。

- [ ] **Step 1: 写阶段覆盖测试**

```js
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const readView = (name) => readFileSync(new URL(`./${name}`, import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const migrated = {
  'Dashboard.vue': 'compact',
  'TaskMonitor.vue': 'monitor',
  'VideoList.vue': 'compact',
  'ImageManage.vue': 'compact'
}
const pendingShellViews = [
  'AVManualScrape.vue', 'ActorManage.vue', 'CollectionManage.vue', 'IPTVManage.vue',
  'ImageCollectionManage.vue', 'PendingDeleteShorts.vue', 'ScrapePreview.vue',
  'SystemSettings.vue', 'Toolbox.vue', 'TvAppManage.vue', 'TvSeriesManage.vue',
  'UserManage.vue', 'VideoUpload.vue'
]
const compatibilityMeta = (component) => new RegExp(
  `component:\\s*${component},\\s*meta:\\s*\\{\\s*hideShellPageHeader:\\s*true\\s*\\}`
)

describe('Precision Ops phase one rollout', () => {
  Object.entries(migrated).forEach(([file, density]) => {
    it(`${file} uses the merged workspace header`, () => {
      const source = readView(file)
      expect(source).toContain('<Layout>')
      expect(source).toContain(`data-density="${density}"`)
      expect(source).not.toContain('<PageHeader')
      expect(router).not.toMatch(compatibilityMeta(file.replace('.vue', '')))
    })
  })

  it('keeps all 13 pending shell pages on the compatibility boundary', () => {
    expect(pendingShellViews).toHaveLength(13)
    pendingShellViews.forEach((file) => {
      expect(readView(file)).toContain('<PageHeader')
      expect(router).toMatch(compatibilityMeta(file.replace('.vue', '')))
    })
  })
})
```

- [ ] **Step 2: 运行管理端全量自动验证**

Run: `cd admin-web && npm test`

Expected: 所有 Vitest 文件通过，0 个失败。

Run: `cd admin-web && npm run build`

Expected: Vite 生产构建成功，0 个错误。

- [ ] **Step 3: 完成四视口浏览器验收**

Run: `cd admin-web && npm run dev -- --host 127.0.0.1 --port 4173`

使用真实管理员账号逐页检查 Dashboard、TaskMonitor、VideoList、ImageManage 的 375、768、1024、1440px：

```text
Dashboard 1440×900：运行摘要、内容库存、240px 趋势和四个快捷入口全部在首屏。
TaskMonitor 1440×900：至少 12 条任务可见，后台刷新不遮住旧内容，成功筛选存在且无成功率。
VideoList 1440×900：至少 10 行可见，封面 72×40，详情 1 次操作，低频动作最多 2 次。
ImageManage 1440×900：至少 5×3 张图片可见，选择与更多操作常显，透明图背景清晰。
四页 375/768/1024px：页面级横向滚动为 0，表格只在自身容器滚动，所有点击目标至少 44px。
键盘：侧栏、命令面板、保存视图、筛选、列设置、行/卡片操作、分页和 Drawer 均可达且焦点可见。
状态：首次加载、后台刷新、真正空态、筛选零结果、读取失败和危险确认逐项可辨。
控制台：无新增 error/warning，图片和网络请求无非预期失败。
```

- [ ] **Step 4: 记录阶段结果并提交**

在 `plan.md` 顶部追加第一阶段完成记录，列出上述 `npm test`、`npm run build` 和四视口结果；若任何浏览器项失败，先修复并重新执行 Step 2-3，不进入阶段二。

Commit message: `验证：完成 Precision Ops 第一阶段验收`

---

## 阶段二：资源集合与批量页

### Task 11: ActorManage、CollectionManage、UserManage 基础 CRUD 集合

**Files:**
- Modify: `admin-web/src/views/ActorManage.vue`
- Modify: `admin-web/src/views/CollectionManage.vue`
- Modify: `admin-web/src/views/UserManage.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Consumes: Layout `header-actions`、`StatusIndicator`、紧凑集合密度。
- Preserves: 三页既有 API、分页、角色更新、演员刮削、启停和删除语义。
- Migration: 三个主编辑表单从居中 Dialog 改为 560px/窄屏全宽 Drawer；演员候选预览随主表单进入同一 Drawer。

- [ ] **Step 1: 扩展阶段覆盖红灯测试**

在 `precisionOpsRollout.spec.js` 合并以下条目：

```js
Object.assign(migrated, {
  'ActorManage.vue': 'compact',
  'CollectionManage.vue': 'compact',
  'UserManage.vue': 'compact'
})
```

再增加：

```js
it('moves base CRUD editors into contextual drawers', () => {
  for (const file of ['ActorManage.vue', 'CollectionManage.vue', 'UserManage.vue']) {
    const source = readView(file)
    expect(source).toContain('<el-drawer')
    expect(source).toContain('size="min(100vw, 560px)"')
  }
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js`

Expected: FAIL，三页仍渲染自身 `PageHeader`，主编辑器仍为 Dialog。

- [ ] **Step 3: 迁移三页工作区动作和集合密度**

三页分别使用以下壳层片段，删除 `PageHeader` import 和模板：

ActorManage 使用：

```vue
<Layout>
  <template #header-actions>
    <el-button :loading="loading" @click="load">刷新</el-button>
    <el-button type="primary" @click="openCreate">创建演员</el-button>
  </template>
  <div class="page-shell actor-page" data-density="compact">
```

CollectionManage 使用：

```vue
<Layout>
  <template #header-actions>
    <el-button :loading="loading" @click="load">刷新</el-button>
    <el-button type="primary" @click="openCreate">新增合集</el-button>
  </template>
  <div class="page-shell collection-page" data-density="compact">
```

UserManage 使用：

```vue
<Layout>
  <template #header-actions>
    <el-button :loading="loading" @click="load">刷新</el-button>
    <el-button type="primary" @click="openCreateDialog">添加用户</el-button>
  </template>
  <div class="page-shell user-page" data-density="compact">
```

工具条只保留搜索、状态筛选、查询/重置和结果总数，不重复放已进入页头的创建按钮。表格状态列改用 `StatusIndicator`；详情或编辑始终 1 次操作，删除/停用保持显式文字与确认。

- [ ] **Step 4: 区分首次加载、真正空态、零结果和读取失败**

三页采用同一状态字段，但各自保留原查询结构：

```js
const loaded = ref(false)
const loadError = ref('')
const initialLoading = computed(() => loading.value && !loaded.value)
const hasFilters = computed(() => String(query.q || '').trim() !== '' || String(query.active || '') !== '')
```

每个 `load` 开始清空 `loadError`，成功后写列表/总量，catch 设置中文错误，finally 设置 `loaded = true` 并结束 loading。模板顺序为行内 `el-alert`、首次骨架、`SectionCard dense`；空列表时 `hasFilters` 为 true 显示“当前筛选无结果”与重置动作，否则显示真正空态和创建主操作。

UserManage 当前没有搜索筛选，改用 `const hasFilters = computed(() => false)`，不得为本次样式升级新增 API 参数。

- [ ] **Step 5: 把主表单改为 Drawer**

保留三个页面现有 `dialogVisible`、`form`、校验、保存和 footer，只替换各自的打开标签与匹配的关闭标签。

ActorManage：

```vue
<el-drawer
  v-model="dialogVisible"
  class="crud-drawer"
  :title="editingID ? '编辑演员' : '创建演员'"
  direction="rtl"
  size="min(100vw, 560px)"
  destroy-on-close
>
```

CollectionManage：

```vue
<el-drawer
  v-model="dialogVisible"
  class="crud-drawer"
  :title="editingID ? '编辑合集' : '新增合集'"
  direction="rtl"
  size="min(100vw, 560px)"
  destroy-on-close
>
```

UserManage：

```vue
<el-drawer
  v-model="dialogVisible"
  class="crud-drawer"
  title="添加用户"
  direction="rtl"
  size="min(100vw, 560px)"
  destroy-on-close
>
```

三个原 `</el-dialog>` 均改为 `</el-drawer>`。Actor/Collection footer 继续调用 `save`，User footer 继续调用 `saveUser`；三页都继续使用 `saving`，不创建第二套保存入口或 payload。

- [ ] **Step 6: 移除三个路由 meta、验证并提交**

路由改为：

```js
{ path: '/actors', component: ActorManage },
{ path: '/collections', component: CollectionManage },
{ path: '/users', component: UserManage }
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/router/index.spec.js src/views/adminDateTimeDisplay.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级基础资源集合页`

---

### Task 12: PendingDeleteShorts 与 ImageCollectionManage 媒体集合

**Files:**
- Modify: `admin-web/src/views/PendingDeleteShorts.vue`
- Modify: `admin-web/src/views/ImageCollectionManage.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Preserves: 待删除队列当前项、保留/删除确认、分页与播放状态；图片合集的编辑/图片选择 Drawer 和脏数据语义。
- Produces: 紧凑媒体复核工作台和紧凑图片合集表格。

- [ ] **Step 1: 扩展阶段覆盖红灯测试**

```js
Object.assign(migrated, {
  'PendingDeleteShorts.vue': 'compact',
  'ImageCollectionManage.vue': 'compact'
})

it('keeps media review actions explicit and keyboard reachable', () => {
  const pending = readView('PendingDeleteShorts.vue')
  const collections = readView('ImageCollectionManage.vue')
  expect(pending).toContain('aria-label="待删除短视频队列"')
  expect(pending).toContain('刷新列表')
  expect(pending).not.toMatch(/\.pending-delete-queue,\s*\.pending-delete-player-panel\s*\{[^}]*box-shadow:/s)
  expect(pending).not.toMatch(/\.pending-delete-video-frame\s*\{[^}]*box-shadow:/s)
  expect(collections).toContain('创建合集')
  expect(collections).toContain('<el-drawer')
  expect(collections).not.toMatch(/border-radius:\s*(?:14|16|18)px/)
  expect(collections).not.toMatch(/(?:linear|radial)-gradient\(/)
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/pendingDeleteShorts.helpers.spec.js src/views/imageCollectionManage.helpers.spec.js`

Expected: FAIL，两页仍有自身 `PageHeader` 且未声明紧凑密度。

- [ ] **Step 3: 迁移 PendingDeleteShorts 壳层并收紧复核工作台**

```vue
<Layout>
  <template #header-actions>
    <el-button :icon="Refresh" :loading="listLoading" @click="refreshList">刷新列表</el-button>
  </template>
  <div class="page-shell pending-delete-page" data-density="compact">
    <p class="page-context-note">逐条复核手机端加入待删除列表的短视频。</p>
    <section class="pending-delete-workbench" :class="{ 'is-empty': !hasItems }" aria-label="待删除短视频队列">
```

保留现有队列和详情 DOM 顺序、键盘按钮、播放器清理、保留/最终删除函数。把副标题改成内容区单行范围说明；队列项固定紧凑高度，长标题两行省略；操作区不依赖 hover。移除队列、播放器面板和视频框的常驻阴影。首次加载使用稳定骨架，读取失败保留页头和重试，队列为空使用真正空态。

- [ ] **Step 4: 迁移 ImageCollectionManage 壳层和媒体缩略图密度**

```vue
<Layout>
  <template #header-actions>
    <el-button type="primary" :icon="Plus" @click="openCreate">创建合集</el-button>
  </template>
  <div class="page-shell image-collection-page" data-density="compact">
```

保留现有筛选、表格、两个 Drawer 和图片关联操作。表格使用 40px 文字行；图片选择 Drawer 内网格使用 `repeat(auto-fill, minmax(184px, 1fr))`、12px gap、4:3 稳定预览和 `object-fit: contain`。缩略图选择与详情入口始终可见，状态使用 `StatusIndicator`。读取失败、真正空态和筛选零结果使用 Task 11 状态模式。

把该页现有 `color: #fff` 和 `background: #0f172a` 分别替换为 `var(--text-on-inverse)` 与 `var(--bg-inverse)`；其它直接 `rgba(...)` 改为语义 token 或 `color-mix(in srgb, <semantic-token> <percentage>, transparent)`。把 14/16/18px 局部圆角全部收敛为 `var(--radius-md)`，把装饰性 `linear-gradient` 替换为对应的 `var(--bg-canvas)`、`var(--bg-inverse)` 或 `var(--bg-surface-muted)` 单色表面。不得在视图 CSS 留直接 hex、直接数字 rgba 或装饰渐变。

- [ ] **Step 5: 移除路由 meta、更新旧路由测试并提交**

```js
{ path: '/short-pending-delete', component: PendingDeleteShorts },
{ path: '/image-collections', component: ImageCollectionManage }
```

把 `router/index.spec.js` 旧断言从“注册并隐藏壳层重复标题”替换为以下精确断言：

```js
expect(routerSource).toContain("{ path: '/short-pending-delete', component: PendingDeleteShorts },")
expect(routerSource).toContain("{ path: '/image-collections', component: ImageCollectionManage },")
expect(pendingSource).not.toContain('<PageHeader')
expect(collectionSource).not.toContain('<PageHeader')
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/router/index.spec.js src/views/pendingDeleteShorts.helpers.spec.js src/views/imageCollectionManage.helpers.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级媒体复核集合页`

---

### Task 13: IPTVManage 与 TvAppManage 服务资源集合

**Files:**
- Modify: `admin-web/src/views/IPTVManage.vue`
- Modify: `admin-web/src/views/TvAppManage.vue`
- Modify: `admin-web/src/views/tvAppManagePage.spec.js`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Consumes: `MetricStrip`、`StatusIndicator`、Layout `header-actions`。
- Preserves: IPTV 上传/远程拉取/频道预览；TV/手机安装包切换、状态/ABI 筛选、上传、发布、下线、二维码和下载。

- [ ] **Step 1: 写服务资源页红灯契约**

```js
Object.assign(migrated, {
  'IPTVManage.vue': 'compact',
  'TvAppManage.vue': 'compact'
})

it('uses metric strips instead of repeated stat cards on service pages', () => {
  for (const file of ['IPTVManage.vue', 'TvAppManage.vue']) {
    const source = readView(file)
    expect(source).toContain('<MetricStrip')
    expect(source).not.toContain('<StatCard')
  }
})
```

在 `tvAppManagePage.spec.js` 增加：

```js
it('uses the compact workspace without dropping package commands', () => {
  expect(tvAppManage).toContain('<template #header-actions>')
  expect(tvAppManage).toContain('data-density="compact"')
  expect(tvAppManage).not.toContain('<PageHeader')
  expect(tvAppManage).toContain('@click="uploadAPK(false)"')
  expect(tvAppManage).toContain("@click=\"confirmAction(row, 'publish')\"")
  expect(tvAppManage).toContain("@click=\"confirmAction(row, 'offline')\"")
  expect(tvAppManage).toContain("@click=\"confirmAction(row, 'delete')\"")
  expect(tvAppManage).toContain('下载 APK')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/tvAppManagePage.spec.js`

Expected: FAIL，两页仍渲染 `PageHeader` 和 `StatCard`。

- [ ] **Step 3: 迁移 IPTVManage**

两页都删除 `PageHeader`、`StatCard` import，并导入 `MetricStrip.vue`、`StatusIndicator.vue`。

```vue
<Layout>
  <template #header-actions>
    <el-button :icon="Refresh" :loading="loading" @click="loadPlaylist">刷新</el-button>
  </template>
  <div class="page-shell iptv-page" data-density="compact">
    <Toolbar dense>
      <template #filters><StatusIndicator :label="`最后更新时间：${updatedAtText}`" tone="neutral" /></template>
      <template #actions>
        <el-button :icon="Refresh" :loading="refreshLoading" @click="refreshPlaylist">远程拉取</el-button>
        <el-button type="primary" :icon="UploadFilled" :loading="uploadLoading" @click="uploadPlaylist">上传 M3U</el-button>
      </template>
    </Toolbar>
    <MetricStrip :items="stats" aria-label="IPTV 摘要" />
```

把现有 `stats` 映射为 `{ key, label, value, scope }`，不得生成健康度或告警。播放列表来源区使用两个并列的无阴影 section，频道表使用紧凑行、行内读取失败和真正空态；频道 URL 继续 overflow tooltip。

- [ ] **Step 4: 迁移 TvAppManage**

```vue
<Layout>
  <template #header-actions>
    <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
    <el-button type="primary" :icon="UploadFilled" :loading="uploadLoading" @click="uploadAPK(false)">上传 APK</el-button>
  </template>
  <div class="page-shell app-package-page" data-density="compact">
```

保留客户端 segmented、搜索、状态、ABI 完整性和家庭可见开关，删除工具条中重复上传按钮。四张 `StatCard` 改为 `MetricStrip`，各项直接标注当前筛选或全局口径；列表继续直显已上传/缺失 ABI、文件大小、版本说明和三类时间。详情/下载始终可见，发布、下线、替换和删除放入文字明确的更多菜单并保留现有确认。

- [ ] **Step 5: 移除路由 meta、验证并提交**

```js
{ path: '/iptv', component: IPTVManage },
{ path: '/tv-app', component: TvAppManage }
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/tvAppManagePage.spec.js src/views/tvAppManage.qr.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级服务资源集合页`

---

### Task 14: 第二阶段集成门禁

**Files:**
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `plan.md`

- [ ] **Step 1: 锁定阶段二 7 页完整清单**

```js
const phaseTwoFiles = [
  'PendingDeleteShorts.vue',
  'ActorManage.vue',
  'CollectionManage.vue',
  'ImageCollectionManage.vue',
  'UserManage.vue',
  'IPTVManage.vue',
  'TvAppManage.vue'
]

it('covers every phase two resource view', () => {
  expect(phaseTwoFiles.every((file) => migrated[file] === 'compact')).toBe(true)
})
```

- [ ] **Step 2: 运行全量测试和构建**

Run: `cd admin-web && npm test && npm run build`

Expected: 全部测试和生产构建通过，0 个失败。

- [ ] **Step 3: 浏览器验收阶段二页面**

在 375/768/1024/1440px 逐页验证：表格自身滚动但页面不横向滚动；主操作位于工作区页头；筛选、分页、详情/编辑、危险确认和 Drawer 均可键盘完成；长演员名、合集名、频道 URL、版本说明和 ABI 标签不遮挡；读取失败保留筛选上下文。

- [ ] **Step 4: 记录结果并提交**

在 `plan.md` 顶部追加第二阶段验证记录，列出 7 页、自动命令和四视口结果。任何阻塞项修复并复跑后再提交。

Commit message: `验证：完成 Precision Ops 第二阶段验收`

---

## 阶段三：表单、编辑器与工具页

### Task 15: VideoUpload 中密度上传流程

**Files:**
- Modify: `admin-web/src/views/VideoUpload.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Preserves: 文件选择、类型差异、远程拉取、上传进度、结果表、演员/合集/图片合集选择和现有 payload。
- Migration: `/upload` 移除 `hideShellPageHeader`；页面声明 `data-density="form"`，不新增顶部重复提交按钮。

- [ ] **Step 1: 写上传页红灯契约**

```js
Object.assign(migrated, { 'VideoUpload.vue': 'form' })

it('keeps the upload workflow inside one medium-density workspace', () => {
  const source = readView('VideoUpload.vue')
  expect(source).toContain('data-density="form"')
  expect(source).toContain('uploadFileList')
  expect(source).toContain('uploadResults')
  expect(source).not.toContain('<PageHeader')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/videoUpload.remote.spec.js`

Expected: FAIL，上传页仍渲染自身 `PageHeader` 且没有中密度作用域。

- [ ] **Step 3: 迁移模板层级**

把页面开头改为：

```vue
<Layout>
  <div class="page-shell upload-page" data-density="form">
    <p class="page-context-note">选择文件与媒体类型，补充关联信息后开始上传。</p>
```

删除 `PageHeader` import。保留现有“文件与基础信息 → 关联信息 → 上传配置 → 进度 → 结果”顺序；每个 `SectionCard` 使用 16px 内边距和 14px 区块标题，不增加嵌套卡片。上传结果表包在 `.table-wrap` 内；375/768px 时内联字段改为单列，文件名和错误使用 overflow tooltip，按钮和上传目标至少 44px。

- [ ] **Step 4: 移除路由 meta、运行测试并提交**

```js
{ path: '/upload', component: VideoUpload }
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/videoUpload.remote.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级视频上传工作区`

---

### Task 16: ScrapePreview 与 AVManualScrape 刮削工作台

**Files:**
- Modify: `admin-web/src/views/ScrapePreview.vue`
- Modify: `admin-web/src/views/AVManualScrape.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`

**Interfaces:**
- Preserves: route query 预填、预览/保存请求、候选选择、站点配置、缓存绕过、图片与元数据表单。
- Produces: 中密度筛选表单、结果区和工作区页头主操作。

- [ ] **Step 1: 写两页迁移红灯契约**

```js
Object.assign(migrated, {
  'ScrapePreview.vue': 'form',
  'AVManualScrape.vue': 'form'
})

it('moves scrape preview commands into the workspace header', () => {
  for (const file of ['ScrapePreview.vue', 'AVManualScrape.vue']) {
    const source = readView(file)
    expect(source).toContain('<template #header-actions>')
    expect(source).toContain('查询预览')
    expect(source).not.toContain('<PageHeader')
  }
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/scrapePreview.helpers.spec.js src/views/avManualScrape.helpers.spec.js`

Expected: FAIL，两页仍有自身标题，主查询动作仍位于表单工具条。

- [ ] **Step 3: 迁移通用刮削页**

```vue
<Layout>
  <template #header-actions>
    <el-button type="primary" :icon="Search" :loading="previewLoading" @click="doPreview">查询预览</el-button>
  </template>
  <div class="page-shell scrape-preview-page" data-density="form">
```

保留视频 ID、标题、年份、类型、季/集全部字段；工具条 actions 移除重复“查询预览”，筛选表单在 1440px 单行、`<1024px` 两列、`<768px` 单列。候选结果、详情编辑和保存区不嵌套新卡片；加载、无候选、读取失败和保存失败保持各自内容区上下文。

把现有 `var(--text-on-inverse, #e2e8f0)` 改为 `var(--text-on-inverse)`；AVManualScrape 同类 fallback 一并移除。

- [ ] **Step 4: 迁移 AV 手动刮削页**

```vue
<Layout>
  <template #header-actions>
    <el-button type="primary" :icon="Search" :loading="previewLoading" @click="doPreview">查询预览</el-button>
  </template>
  <div class="page-shell av-manual-scrape-page" data-density="form">
```

保留视频 ID、标题、站点分类、目标站点、绕过缓存和站点配置；结果双栏在 900px 以下改单列。所有候选按钮可键盘聚焦，选中态使用边框、文字和图标/状态，不只依赖颜色。

- [ ] **Step 5: 移除路由 meta、验证并提交**

```js
{ path: '/scrape', component: ScrapePreview },
{ path: '/av-scrape', component: AVManualScrape }
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/scrapePreview.helpers.spec.js src/views/avManualScrape.helpers.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级媒体刮削工作台`

---

### Task 17: TvSeriesManage、SystemSettings 与 Toolbox 壳层页

**Files:**
- Modify: `admin-web/src/views/TvSeriesManage.vue`
- Modify: `admin-web/src/views/SystemSettings.vue`
- Modify: `admin-web/src/views/Toolbox.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/router/index.js`
- Modify: `admin-web/src/router/index.spec.js`
- Verify: `admin-web/src/views/tvSeriesManage.helpers.spec.js`
- Verify: `admin-web/src/views/toolboxPage.spec.js`

**Interfaces:**
- Preserves: 电视剧系列/季/集编辑、系统清理/日志、工具箱新标签页链接和独立工具边界。
- Migration: 三页移除重复 PageHeader；TvSeries 使用表单密度，Settings/Toolbox 使用表单密度和无阴影区块。

- [ ] **Step 1: 写三页红灯契约**

```js
Object.assign(migrated, {
  'TvSeriesManage.vue': 'form',
  'SystemSettings.vue': 'form',
  'Toolbox.vue': 'form'
})

it('keeps toolbox destinations in standalone tabs', () => {
  const source = readView('Toolbox.vue')
  expect(source).toContain('target="_blank"')
  expect(source).toContain('rel="noopener noreferrer"')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/tvSeriesManage.helpers.spec.js src/views/toolboxPage.spec.js`

Expected: FAIL，三页仍有自身 PageHeader 且没有显式表单密度。

- [ ] **Step 3: 迁移 TvSeriesManage**

```vue
<Layout>
  <template #header-actions>
    <el-button type="primary" :icon="Plus" @click="openCreateSeries">新建系列</el-button>
    <el-button :icon="Search" @click="query.page = 1; loadList()">筛选</el-button>
  </template>
  <div class="page-shell tv-manage-shell" data-density="form">
```

工具条只保留搜索、启用状态、可播分集和重置。保持左侧系列列表、右侧系列/季/集编辑边界；左侧列表内部声明 `data-density="compact"`，右侧表单沿用 36px 控件和 16px 字段间距。删除嵌套 `SectionCard` 的装饰边框：季和分集使用标题分隔线与无框字段组，但不移动或重命名现有 `v-model` 和保存函数。

- [ ] **Step 4: 迁移 SystemSettings 与 Toolbox**

两页开头分别改为：

```vue
<Layout>
  <div class="page-shell settings-page" data-density="form">
    <p class="page-context-note">执行临时文件清理并查看最近系统日志。</p>
```

```vue
<Layout>
  <div class="page-shell toolbox-page" data-density="form">
    <p class="page-context-note">工具会在独立标签页打开，并保持当前管理端上下文。</p>
```

SystemSettings 保留清理警示和日志刷新动作，错误显示在对应区块；日志使用等宽字体和自身滚动。Toolbox 只把真正重复的工具入口保留为 8px 卡片，卡片图标、名称、说明和“新标签页打开”状态完整；不得把独立工具路由并回主壳层。

- [ ] **Step 5: 移除路由 meta、验证并提交**

```js
{ path: '/tv-series', component: TvSeriesManage },
{ path: '/toolbox', component: Toolbox },
{ path: '/settings', component: SystemSettings }
```

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/tvSeriesManage.helpers.spec.js src/views/toolboxPage.spec.js src/router/index.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级管理端表单与工具壳层`

---

### Task 18: 轻量独立工具与登录页

**Files:**
- Modify: `admin-web/src/views/ToolboxEd2k.vue`
- Modify: `admin-web/src/views/ToolboxOrphanFiles.vue`
- Modify: `admin-web/src/views/ToolboxPasswordVault.vue`
- Modify: `admin-web/src/views/Login.vue`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Verify: `admin-web/src/views/ToolboxEd2k.spec.js`
- Verify: `admin-web/src/views/toolboxOrphanFiles.helpers.spec.js`

**Interfaces:**
- Preserves: 四页不渲染 `Layout`，继续使用自己的 `PageHeader` 和返回工具箱/登录边界。
- Preserves: ED2K 解析、孤儿文件扫描/删除、密码增删改查/显示复制、登录鉴权。
- Produces: 独立工具中密度外壳；密码列表内部使用紧凑密度。

- [ ] **Step 1: 写独立工作区边界红灯测试**

在 `precisionOpsRollout.spec.js` 增加：

```js
const standaloneViews = {
  'ToolboxEd2k.vue': 'form',
  'ToolboxOrphanFiles.vue': 'form',
  'ToolboxPasswordVault.vue': 'form',
  'Login.vue': 'form'
}

Object.entries(standaloneViews).forEach(([file, density]) => {
  it(`${file} keeps an independent titled workspace`, () => {
    const source = readView(file)
    expect(source).toContain(`data-density="${density}"`)
    expect(source).toContain('<PageHeader')
    expect(source).not.toContain('<Layout')
  })
})

it('removes the decorative login gradient', () => {
  const source = readView('Login.vue')
  expect(source).not.toMatch(/(?:linear|radial)-gradient\(/)
  expect(source).not.toMatch(/\.login-card\s*\{[^}]*box-shadow:/s)
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/ToolboxEd2k.spec.js src/views/toolboxOrphanFiles.helpers.spec.js`

Expected: FAIL，四页尚未声明任务密度。

- [ ] **Step 3: 为独立工作区增加中密度边界**

只修改根节点属性，不把页面搬入 Layout：

```vue
<main class="tool-workspace" data-density="form">
```

```vue
<main class="tool-workspace orphan-tool" data-density="form">
```

```vue
<main class="tool-workspace password-vault-tool" data-density="form">
```

```vue
<main class="login-page" data-density="form">
```

ED2K 输入/结果、孤儿扫描状态/表格和登录表单保留原结构；普通区块阴影清零、圆角不超过 8px。Login 的装饰性 `linear-gradient` 改为 `var(--bg-canvas)`/`var(--bg-surface)` 单色层级，并移除登录面板常驻阴影。密码页的“密码库” `SectionCard` 额外添加 `data-density="compact"`，使搜索、表格和分页使用集合尺寸；创建/编辑和显示密码 Dialog 保持中密度和既有确认语义。

- [ ] **Step 4: 运行测试、构建并提交**

Run: `cd admin-web && npm test -- src/views/precisionOpsRollout.spec.js src/views/ToolboxEd2k.spec.js src/views/toolboxOrphanFiles.helpers.spec.js src/views/adminDateTimeDisplay.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：统一轻量工具与登录工作区`

---

### Task 19: ToolboxEd2kDownload 独立下载工作台

**Files:**
- Modify: `admin-web/src/views/ToolboxEd2kDownload.vue`
- Modify: `admin-web/src/views/ToolboxEd2kDownload.spec.js`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`

**Interfaces:**
- Preserves: 新建任务 Dialog、任务选择、状态轮询、下载引擎状态、历史、文件列表、危险操作和返回工具箱。
- Produces: 中密度工作台外壳、紧凑任务列表/文件表和明确状态指示。

- [ ] **Step 1: 写下载工作台红灯契约**

```js
standaloneViews['ToolboxEd2kDownload.vue'] = 'form'

it('scopes compact density to ED2K task collections', () => {
  const source = readView('ToolboxEd2kDownload.vue')
  expect(source).toContain('class="task-list-card" data-density="compact"')
  expect(source).toContain('<StatusIndicator')
  expect(source).toContain('<PageHeader')
  expect(source).not.toContain('<Layout')
})
```

在现有 `ToolboxEd2kDownload.spec.js` 增加：

```js
it('keeps the primary task commands and history', () => {
  expect(source).toContain('@click="openCreateDialog"')
  expect(source).toContain('@click="manualRefresh"')
  expect(source).toContain('@click="deleteTask(selectedTask)"')
  expect(source).toContain('@click="retryTask(selectedTask)"')
  expect(source).toContain('@click="retryCleanup(selectedTask)"')
  expect(source).toContain('<template #title>历史记录</template>')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/ToolboxEd2kDownload.spec.js src/views/precisionOpsRollout.spec.js`

Expected: FAIL，任务集合没有显式紧凑作用域，状态仍只用 tag。

- [ ] **Step 3: 分离工作台与集合密度**

导入 `../components/base/StatusIndicator.vue`，不新增状态映射文件。

根节点改为：

```vue
<main class="tool-workspace" data-density="form">
```

任务列表区改为：

```vue
<SectionCard class="task-list-card" data-density="compact">
  <template #title>下载任务</template>
  <template #description>按最近状态变更排序，选择任务后在右侧查看详情。</template>
```

下载引擎、所选任务和文件状态使用 `StatusIndicator`，文字保留现有 `selectedTaskLabel` 和状态 helper。左侧任务行固定高度，标题/哈希两行省略；右侧详情保持中密度；`<1024px` 双栏改单列。新建任务仍使用现有居中 Dialog，因为它是短流程录入例外。

- [ ] **Step 4: 运行测试、构建并提交**

Run: `cd admin-web && npm test -- src/views/ToolboxEd2kDownload.spec.js src/views/precisionOpsRollout.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级 ED2K 下载工作台`

---

### Task 20: ToolboxArchiveImport 高复杂度批次工作台

**Files:**
- Modify: `admin-web/src/views/ToolboxArchiveImport.vue`
- Modify: `admin-web/src/views/ToolboxArchiveImport.spec.js`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Verify: `admin-web/src/views/toolboxArchiveImport.helpers.spec.js`

**Interfaces:**
- Preserves: 压缩包上传、批次轮询、密码重试、文件/目录分组、单项/批量编辑、集合新建、选择、处理、重试和所有 Dialog/Drawer dirty 逻辑。
- Produces: 中密度独立工作台、紧凑批次/文件集合和非卡片化指标条。

- [ ] **Step 1: 写压缩包工作台红灯契约**

```js
standaloneViews['ToolboxArchiveImport.vue'] = 'form'

it('keeps archive business boundaries while applying collection density', () => {
  const source = readView('ToolboxArchiveImport.vue')
  expect(source).toContain('class="archive-batch-panel" data-density="compact"')
  expect(source).toContain('class="archive-file-panel" data-density="compact"')
  expect(source).toContain('<MetricStrip')
  expect(source).toContain('<BulkActionBar')
  expect(source).toContain('<PageHeader')
  expect(source).not.toContain('<Layout')
  expect(source).not.toMatch(/(?:linear|radial)-gradient\(/)
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/ToolboxArchiveImport.spec.js src/views/toolboxArchiveImport.helpers.spec.js src/views/precisionOpsRollout.spec.js`

Expected: FAIL，概览仍为独立卡片，批次和文件区没有显式集合密度。

- [ ] **Step 3: 只调整模板组织和样式作用域**

导入 `../components/base/MetricStrip.vue`，删除手写概览卡所需的纯装饰结构，不改 `overviewCards` 的统计来源。

根节点改为：

```vue
<main class="tool-workspace archive-import-tool" data-density="form">
```

把现有 `overviewCards` 直接交给：

```vue
<MetricStrip :items="overviewCards" aria-label="压缩包批次摘要" />
```

让 computed 输出 `{ key: label, label, value }`，不改统计来源；原 `hint` 合并为指标条下方一行 `.archive-overview-note`，在窄屏允许自然换行，不能塞入 `scope` 撑宽指标。为批次和文件区分别增加：

```vue
<SectionCard class="archive-batch-panel" data-density="compact">
<SectionCard class="archive-file-panel" data-density="compact">
```

批次列表、分组网格、文件清单、详情 Drawer、批量浮条和 5 类现有 Dialog 的事件、props、函数名、请求 payload 全部保持不变。移除普通面板阴影和装饰性 `linear-gradient`/`radial-gradient`，用 `var(--bg-canvas)`、`var(--bg-surface)`、`var(--bg-surface-muted)` 建立单色层级；表格/文件行采用 40px 或 52px 媒体行；错误文本单行省略并保留 tooltip；375/768px 时 Dialog/Drawer 不溢出。

- [ ] **Step 4: 运行既有大页回归、全量测试和构建**

Run: `cd admin-web && npm test -- src/views/ToolboxArchiveImport.spec.js src/views/toolboxArchiveImport.helpers.spec.js src/views/precisionOpsRollout.spec.js`

Expected: PASS。

Run: `cd admin-web && npm test && npm run build`

Expected: 全部测试和 Vite 构建通过。

Commit message: `样式：升级压缩包导入工作台`

---

### Task 21: ToolboxImageWorkbench 与 MaskEditor 媒体编辑工作台

**Files:**
- Modify: `admin-web/src/views/ToolboxImageWorkbench.vue`
- Modify: `admin-web/src/views/ImageWorkbenchMaskEditor.vue`
- Modify: `admin-web/src/views/imageWorkbench.helpers.spec.js`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`

**Interfaces:**
- Preserves: IndexedDB 历史、生成任务、参考图/媒体库选择、参数、结果导入/保存、mask canvas 指针和缩放逻辑。
- Produces: 中密度工作台、184px 结果/媒体网格、稳定 mask 工具栏和窄屏无重叠布局。

- [ ] **Step 1: 写图像工作台红灯契约**

```js
standaloneViews['ToolboxImageWorkbench.vue'] = 'form'

it('keeps the mask editor component-only boundary', () => {
  const source = readView('ImageWorkbenchMaskEditor.vue')
  expect(source).toContain('data-density="form"')
  expect(source).toContain('mask-editor__toolbar')
  expect(source).toContain("const MASK_OPAQUE_COLOR = '#ffffff'")
  expect(source).toContain("overlayCtx.fillStyle = resolveCanvasColor('--primary')")
  expect(source).not.toMatch(/rgba?\(\s*\d/)
  expect(source).not.toMatch(/\.mask-editor__canvas\s*\{[^}]*box-shadow:/s)
  expect(source).not.toContain('<Layout')
  expect(source).not.toContain('<PageHeader')
})

it('uses stable media-grid geometry in the image workbench', () => {
  const source = readView('ToolboxImageWorkbench.vue')
  expect(source).toContain('minmax(184px, 1fr)')
  expect(source).toContain('object-fit: contain')
  expect(source).toContain('<PageHeader')
})
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `cd admin-web && npm test -- src/views/imageWorkbench.helpers.spec.js src/views/precisionOpsRollout.spec.js`

Expected: FAIL，工作台和 mask editor 没有任务密度，媒体网格未统一到 184px。

- [ ] **Step 3: 迁移图像生成工作台**

导入 `../components/base/StatusIndicator.vue`，继续使用现有状态 label/tone 派生值。

```vue
<main class="image-workbench" data-density="form">
```

保留顶部返回、打开媒体库和状态 tag；状态 tag 改用 `StatusIndicator`。输入/参数区保持中密度，结果和媒体库选择区增加 `data-density="compact"`。`.result-grid` 与 `.library-grid` 使用 `repeat(auto-fill, minmax(184px, 1fr))`、12px gap、稳定预览比例和 `object-fit: contain`；导入/保存/编辑操作始终可键盘访问，不只在 hover 出现。

- [ ] **Step 4: 迁移 MaskEditor 固定格式控件**

Dialog 根内容增加：

```vue
<div class="mask-editor" data-density="form">
```

Mask 的纯白像素是导出 PNG 的业务数据编码，不能耦合到 UI 主题；overlay 才使用界面语义色。在脚本中增加：

```js
const MASK_OPAQUE_COLOR = '#ffffff'

function resolveCanvasColor(token) {
  if (typeof window === 'undefined' || typeof document === 'undefined') return ''
  return window.getComputedStyle(document.documentElement).getPropertyValue(token).trim()
}
```

`fillWhiteMask` 以及 brush/eraser 两个分支的 `ctx.fillStyle`/`ctx.strokeStyle` 统一使用 `MASK_OPAQUE_COLOR`；brush 的 `destination-out` 只依赖不透明 alpha，不再保留直接黑色 `rgba(...)`。把 overlay 的 `#3b82f6` 改为 `resolveCanvasColor('--primary')`。调用点只在组件 mounted 且 canvas context 已建立后运行，因此不得加入第二套界面色 fallback。

棋盘背景是透明媒体识别能力，可保留两层 `linear-gradient`，但其中的直接数字 rgba 改为 `color-mix(in srgb, var(--text-primary) 4%, transparent)`；移除 `.mask-editor__canvas` 常驻 `var(--shadow-sm)`。最终 direct-color 审计只对上述精确 `MASK_OPAQUE_COLOR` 声明做域例外，其它直接 hex/数字 rgba 仍失败。

工具栏用稳定 grid/flex 尺寸：图标按钮 36×36px，模式使用 segmented/radio，数值缩放使用 slider，画布容器保持明确 `min-height`、`max-width` 和 aspect constraints。`<768px` 工具栏换行但不覆盖画布；不修改 pointer 事件、mask 数据结构或 confirm/cancel emits。

- [ ] **Step 5: 运行测试、构建并提交**

Run: `cd admin-web && npm test -- src/views/imageWorkbench.helpers.spec.js src/views/precisionOpsRollout.spec.js`

Expected: PASS。

Run: `cd admin-web && npm run build`

Expected: Vite 构建成功。

Commit message: `样式：升级图像生成与遮罩工作台`

---

## 阶段四：全站收尾与完成门禁

### Task 22: 25 视图覆盖、响应式、无障碍与最终验证

**Files:**
- Create: `admin-web/src/views/precisionOpsAudit.spec.js`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js`
- Modify: `admin-web/src/assets/themeTokens.spec.js`
- Modify: `plan.md`
- Modify only if implementation introduced a new long-term decision: `CONTEXT.md`

**Interfaces:**
- Verifies: 17 个 shell 页面全部使用 Layout 合并页头，7 个独立页面保留 PageHeader，MaskEditor 保持组件边界。
- Verifies: 所有 25 个 `admin-web/src/views/*.vue` 都有唯一推广归属。
- Verifies: 普通业务路由不再包含 `hideShellPageHeader`。

- [ ] **Step 1: 把覆盖测试扩展为完整清单**

```js
import { readdirSync } from 'node:fs'

const shellViews = [
  'AVManualScrape.vue', 'ActorManage.vue', 'CollectionManage.vue', 'Dashboard.vue',
  'IPTVManage.vue', 'ImageCollectionManage.vue', 'ImageManage.vue', 'PendingDeleteShorts.vue',
  'ScrapePreview.vue', 'SystemSettings.vue', 'TaskMonitor.vue', 'Toolbox.vue',
  'TvAppManage.vue', 'TvSeriesManage.vue', 'UserManage.vue', 'VideoList.vue', 'VideoUpload.vue'
]
const standalonePageViews = [
  'Login.vue', 'ToolboxArchiveImport.vue', 'ToolboxEd2k.vue', 'ToolboxEd2kDownload.vue',
  'ToolboxImageWorkbench.vue', 'ToolboxOrphanFiles.vue', 'ToolboxPasswordVault.vue'
]
const componentViews = ['ImageWorkbenchMaskEditor.vue']

it('assigns all 25 Vue views to exactly one boundary', () => {
  const actual = readdirSync(new URL('.', import.meta.url)).filter((name) => name.endsWith('.vue')).sort()
  const assigned = [...shellViews, ...standalonePageViews, ...componentViews].sort()
  expect(actual).toEqual(assigned)
  expect(assigned).toHaveLength(25)
})

it('fully removes the ordinary-page compatibility meta', () => {
  const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
  expect(router).not.toContain('hideShellPageHeader')
})
```

- [ ] **Step 2: 增加全站静态审计**

创建 `precisionOpsAudit.spec.js`：

```js
import { readFileSync, readdirSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const directory = new URL('.', import.meta.url)
const viewFiles = readdirSync(directory).filter((name) => name.endsWith('.vue'))
const MASK_DATA_COLOR_DECLARATION = "const MASK_OPAQUE_COLOR = '#ffffff'"
const functionalGradientBlocks = {
  'ImageManage.vue': /\.image-grid-card__preview\s*\{[^{}]*\}/g,
  'ImageWorkbenchMaskEditor.vue': /\.mask-editor__stage\s*\{[^{}]*\}/g
}

function colorAuditSource(file, source) {
  return file === 'ImageWorkbenchMaskEditor.vue'
    ? source.replace(MASK_DATA_COLOR_DECLARATION, '')
    : source
}

function gradientAuditSource(file, source) {
  const block = functionalGradientBlocks[file]
  return block ? source.replace(block, '') : source
}

function persistentPanelShadowSelectors(source) {
  return [...source.matchAll(/([^{}]+)\{([^{}]*box-shadow:\s*var\(--shadow-[a-z0-9-]+\)[^{}]*)\}/gi)]
    .map((match) => match[1].trim())
    .filter((selector) => !/:(hover|focus|focus-visible|active)\b/.test(selector))
    .filter((selector) => !/\.(?:is-)?(?:active|selected)\b/.test(selector))
    .filter((selector) => !/(drawer|dialog|popover|popper|tooltip|dropdown|bulk-action|floating|overlay|modal)/i.test(selector))
}

function oversizedDirectRadii(source) {
  return [...source.matchAll(/border-radius:\s*([0-9]+(?:\.[0-9]+)?)px/g)]
    .map((match) => Number(match[1]))
    .filter((value) => value > 8)
}

describe('Precision Ops final audit', () => {
  viewFiles.forEach((file) => {
    const source = readFileSync(new URL(file, directory), 'utf8')
    it(`${file} avoids nonzero tracking and direct view colors`, () => {
      const auditedSource = colorAuditSource(file, source)
      expect(source).not.toMatch(/letter-spacing:\s*(?!0(?:px|rem|em)?\s*;)[^;]+;/)
      expect(auditedSource).not.toMatch(/#[0-9a-f]{3,8}\b/i)
      expect(auditedSource).not.toMatch(/rgba?\(\s*\d/i)
    })
    it(`${file} avoids persistent panel shadows`, () => {
      expect(persistentPanelShadowSelectors(source)).toEqual([])
    })
    it(`${file} avoids oversized direct panel radii`, () => {
      expect(oversizedDirectRadii(source)).toEqual([])
    })
    it(`${file} avoids decorative gradients`, () => {
      expect(gradientAuditSource(file, source)).not.toMatch(/(?:linear|radial)-gradient\(/)
    })
  })

  it('keeps image card actions visible without hover dependency', () => {
    const imageManage = readFileSync(new URL('./ImageManage.vue', import.meta.url), 'utf8')
    expect(imageManage).not.toMatch(/image-grid-card__actions[\s\S]{0,180}opacity:\s*0/)
  })

  it('keeps shell letter spacing at zero', () => {
    const layout = readFileSync(new URL('../components/Layout.vue', import.meta.url), 'utf8')
    expect(layout).not.toMatch(/letter-spacing:\s*(?!0(?:px|rem|em)?\s*;)[^;]+;/)
  })
})
```

在 `themeTokens.spec.js` 增加以下最终锁定；继续保留旧玫红和视图直接 hex 审计：

```js
it('locks the final Precision Ops geometry and semantic palette', () => {
  expect(css).toContain('--admin-sidebar-width: 224px')
  expect(css).toContain('--admin-sidebar-collapsed-width: 56px')
  expect(css).toContain('--admin-header-height: 52px')
  expect(css).toContain('--text-muted: #607085')
  expect(css).toContain('--success-600: #047857')
  expect(css).toContain('--warning-600: #b45309')
  expect(css).toContain('--danger-600: #c81e1e')
  expect(css).toContain('--info-600: #0369a1')
  expect(css).toMatch(/\[data-density="compact"\][\s\S]*--control-height:\s*32px/)
  expect(css).toMatch(/\[data-density="monitor"\][\s\S]*--table-row-height:\s*44px/)
  expect(css).toMatch(/\[data-density="form"\][\s\S]*--control-height:\s*36px/)
  expect(css).toMatch(/--radius-md:\s*8px/)
})
```

- [ ] **Step 3: 运行全量自动验证**

Run: `cd admin-web && npm test`

Expected: 所有 Vitest 测试通过，0 个失败。

Run: `cd admin-web && npm run build`

Expected: Vite 生产构建成功，0 个错误。

Run: `git diff --check`

Expected: 无输出，退出码 0。

Run:

```bash
replacement_status=0
replacement_output="$(LC_ALL=C rg -n $'\xEF\xBF\xBD' CONTEXT.md plan.md docs/superpowers admin-web/src 2>&1)" || replacement_status=$?
test "$replacement_status" -eq 1
test -z "$replacement_output"
```

Expected: 两个 `test` 均通过；严格区分“无匹配”的退出码 1 与命令错误。

- [ ] **Step 4: 完成 25 视图四档视口验收**

Run: `cd admin-web && npm run dev -- --host 127.0.0.1 --port 4173`

用真实管理员数据逐页验证 375、768、1024、1440px，并保存验收记录：

```text
壳层：侧栏 224/56px，页头 52px，最近访问最多 3 项，当前分组自动展开，移动 Drawer 可关闭且不穿透滚动。
密度：集合/监控/表单分别命中 32/44/36px 规格；<1024px 点击目标至少 44px。
量化：Dashboard 1440×900 完整首屏；VideoList ≥10 行；TaskMonitor ≥12 行；ImageManage ≥15 张。
滚动：页面级横向滚动为 0；宽表格只在自身容器横向滚动。
键盘：导航、命令面板、保存视图、筛选、列设置、资源选择、更多菜单、分页、Dialog 和 Drawer 100% 可达。
对比与动效：正文和交互文字对比度至少 4.5:1；状态不只依赖颜色；启用 prefers-reduced-motion 时无非必要过渡。
状态：首次加载、后台刷新、真正空态、筛选零结果、读取失败、危险确认均不丢失上下文。
媒体：透明图、加载失败图、长标题、长 ID、长错误和长中文按钮不遮挡。
独立页：7 个独立工具/登录页无 Layout，返回工具箱、登录和新标签页边界不变。
控制台：无新增 error/warning，网络和图片无非预期失败。
```

- [ ] **Step 5: 对照规格逐项复核业务边界**

Run:

```bash
MERGE_BASE="$(git merge-base master HEAD)"
git diff --name-only "$MERGE_BASE"
git ls-files --others --exclude-standard
git status --short
```

Expected: 只出现 `admin-web/`、`CONTEXT.md`、`plan.md` 和本计划明确的测试/文档文件；不出现 Go、Android、migration、依赖锁文件或 API 契约改动。

Run: `git diff "$(git merge-base master HEAD)" -- admin-web/src/api admin-web/package.json admin-web/package-lock.json`

Expected: 无输出。

逐项确认：Dashboard 无推测健康指标；TaskMonitor 无混合口径成功率；保存视图不含页码/选择/Drawer；VideoList/ImageManage 现有批量和 dirty guard 测试通过；25 页路由目标和权限不变。

- [ ] **Step 6: 记录最终结果并提交**

在 `plan.md` 顶部追加最终记录，写明完整测试、构建、静态检查、25 页四视口和业务边界结果。只有执行中产生规格未覆盖但长期有效的新决定时才追加 `CONTEXT.md`；不能写临时进度。

Commit message: `验证：完成管理端 Precision Ops 全站验收`

---

## 规格覆盖矩阵

| 设计规格 | 实施任务 |
|---|---|
| 目标、非目标、数据诚实和最小业务扰动 | 全局约束；Task 6-9；Task 22 |
| 壳层尺寸、合并页头、最近访问、五组导航和迁移兼容 | Task 1-3；各页面迁移任务；Task 22 |
| 色彩、字体、间距、圆角、阴影和按任务密度 | Task 1；Task 4；Task 15-21；Task 22 |
| 三层设计架构和共享组件边界 | Task 1；Task 3-5 |
| 版本化保存视图、本地回退和选择清理 | Task 5；Task 8-9 |
| Dashboard、TaskMonitor、VideoList、ImageManage | Task 6-10 |
| 首次加载、后台刷新、空态、零结果、错误和危险操作 | Task 6-9；Task 11-13；Task 15-22 |
| 响应式、键盘、焦点、对比和 reduced motion | Task 1；Task 3-5；Task 10；Task 14；Task 22 |
| 阶段二 7 个资源集合页 | Task 11-14 |
| 阶段三 14 个表单、编辑器与工具视图 | Task 15-21 |
| 量化验收、自动验证、浏览器验收、风险和回滚 | 每任务独立提交；Task 10；Task 14；Task 22 |

自检结论：规格中的每个实现要求均有任务承接，没有需要另拆独立子系统的遗漏；外部案例调研只作为设计依据，不在实现阶段引入外部依赖或复制外部组件源码。

---

## 执行完成定义

- Task 1-22 按顺序完成，每个任务的定向测试、全量测试、构建和提交均有证据。
- 四个样板页满足规格中的首屏数量与操作次数指标。
- 25 个 Vue 视图均通过 375/768/1024/1440px、键盘、长文本和状态验收。
- 普通业务页不再使用 `hideShellPageHeader`；独立工具仍不渲染 Layout。
- Git diff 不包含路由目标、权限、API、数据库、依赖、Go 或 Android 行为变更。
- `CONTEXT.md`、`plan.md`、中文文案和中文提交信息无乱码。
