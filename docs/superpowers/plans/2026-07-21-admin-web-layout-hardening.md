# 管理端布局健壮性修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复管理端在四档视口、真实动态数据和 Drawer 场景下的内容挤压、局部溢出、触控目标和焦点反馈缺陷，同时保持 Precision Ops 高密度工作流与既有业务语义。

**Architecture:** 先在共享基础组件收敛批量操作条与命令面板，再在页面局部为 Teleport Drawer 声明 form 密度、为 Drawer 宽表建立独立横滚容器，最后收敛任务表列宽和几个由动态文本或固定字段触发的最小宽度。所有回归测试沿用现有 Vitest 源码审计层；不新增运行时依赖、测试框架或通用布局抽象。

**Tech Stack:** Vue 3 SFC、Element Plus、Vitest、Vite。

## 全局约束

- 只修改 `admin-web/` 布局、对应现有 Vitest 源码审计、`CONTEXT.md` 和 `plan.md`；不修改 API、路由、权限、payload、数据模型或危险操作语义。
- 保留桌面 compact / monitor / form 密度；仅 1024px 以下让拥有 `data-density` 的目标消费既有 44px 触控规则。
- `375 / 768 / 1024 / 1440px` 页面级 `scrollWidth === clientWidth`；宽表只允许 `.table-wrap` 或 `.drawer-table-wrap` 内横向滚动。
- 不用 `overflow-x: hidden` 截断问题，不把表格改为卡片，不新增远程字体、UI 库或依赖。
- 所有 Markdown、测试名称、注释和面向用户文案使用正确中文；不写入测试账号、密码或 token。
- 每项代码任务先运行指定的红灯测试，再写最小实现并运行绿灯测试；完成一个任务后单独提交中文 commit。
- `CONTEXT.md` 只保留长期术语，进度和验证流水仅追加至 `plan.md`；既有未跟踪 `.superpowers/` 不纳入任何提交。

## 文件与职责

| 路径 | 职责 |
|---|---|
| `admin-web/src/components/base/BulkActionBar.vue` | 选中资源后的共享浮动批量命令布局。 |
| `admin-web/src/components/base/CommandPalette.vue` | 全局快捷跳转对话框的焦点与滚动边界。 |
| `admin-web/src/components/base/precisionOpsComponents.spec.js` | 共享组件的源码布局契约。 |
| `admin-web/src/views/VideoList.vue` | 视频详情、批量编辑和字幕管理 Drawer。 |
| `admin-web/src/views/ImageManage.vue` | 图片上传、详情和筛选 Drawer。 |
| `admin-web/src/views/ImageCollectionManage.vue` | 图片合集编辑与关联图片 Drawer。 |
| `admin-web/src/views/ActorManage.vue` | 演员列表动态别名列与编辑 Drawer。 |
| `admin-web/src/views/CollectionManage.vue` | 合集编辑 Drawer 的 form 密度声明。 |
| `admin-web/src/views/UserManage.vue` | 用户编辑 Drawer 的 form 密度声明。 |
| `admin-web/src/views/PendingDeleteShorts.vue` | 1024–1199px 待删除工作台详情布局。 |
| `admin-web/src/views/TaskMonitor.vue` | 九列任务表在 1440px 的总宽度预算。 |
| `admin-web/src/views/videoListPage.spec.js` | 视频页 Drawer、季/集字段、候选长文本与批量操作的结构断言。 |
| `admin-web/src/views/imageManagePage.spec.js` | 图片上传 Drawer 宽表及 Drawer 密度结构断言。 |
| `admin-web/src/views/precisionOpsRollout.spec.js` | 集合页、图片合集 Drawer 和待删除工作台的跨页布局断言。 |
| `admin-web/src/views/taskMonitorPage.spec.js` | 任务表列完整性及 1440px 列宽预算断言。 |
| `CONTEXT.md` | 已提交的“管理端布局缺陷”术语；只在出现新增长期约定时追加。 |
| `plan.md` | 反向时间顺序记录红灯、实现、浏览器验收、全量验证和提交范围。 |

---

### Task 1: 共享批量操作条与命令面板边界

**Files:**
- Modify: `admin-web/src/components/base/BulkActionBar.vue:35-73`
- Modify: `admin-web/src/components/base/CommandPalette.vue:174-255`
- Modify: `admin-web/src/components/base/precisionOpsComponents.spec.js:1-60,末尾共享组件断言`

**Interfaces:**
- Consumes: `BulkActionBar` 既有 `count: Number`、`actions: Array` props，以及 `CommandPalette` 既有快捷键、焦点陷阱和 `select(item)` 路由语义。
- Produces: 在 `max-width: 63.9375rem` 下可换行的 `.bulk-action-bar` / `.bulk-action-bar__actions`，以及 `.command-palette__search:focus-within` 与 `.command-palette__list` 的可回归样式契约。

- [ ] **Step 1: 写入会失败的共享组件源码测试**

在 `precisionOpsComponents.spec.js` 顶部读取两个 SFC 源文件：

```js
const bulkActionBarSource = readFileSync(new URL('./BulkActionBar.vue', import.meta.url), 'utf8')
const commandPaletteSource = readFileSync(new URL('./CommandPalette.vue', import.meta.url), 'utf8')
const bulkActionBarStyle = extractBlock(bulkActionBarSource, 'style')
const commandPaletteStyle = extractBlock(commandPaletteSource, 'style')
```

在 `describe('Precision Ops base components', ...)` 末尾添加一个测试，先取得 `@media (max-width: 63.9375rem)` 之后的 BulkActionBar CSS，并断言：

```js
expect(findRule(mobileBulkStyle, '.bulk-action-bar')).toContain('flex-direction: column')
expect(findRule(mobileBulkStyle, '.bulk-action-bar')).toContain('align-items: stretch')
expect(findRule(mobileBulkStyle, '.bulk-action-bar__actions')).toContain('width: 100%')
expect(findRule(mobileBulkStyle, '.bulk-action-bar__actions')).toContain('flex-wrap: wrap')
expect(commandPaletteStyle).toMatch(
  /\.command-palette__search:focus-within\s*\{[^}]*box-shadow:\s*0 0 0 3px var\(--line-focus\);/s
)
expect(findRule(commandPaletteStyle, '.command-palette__list')).toContain('overscroll-behavior: contain')
```

测试中先确认媒体查询起点存在，避免错误地把整个样式文件作为窄屏规则读取。

- [ ] **Step 2: 运行红灯测试并确认失败原因**

Run:

```bash
cd admin-web && npm test -- src/components/base/precisionOpsComponents.spec.js
```

Expected: FAIL；缺少 BulkActionBar 的窄屏纵向/换行规则、命令面板 `:focus-within` 焦点环和结果列表的 `overscroll-behavior`。

- [ ] **Step 3: 写入最小共享样式实现**

在 `BulkActionBar.vue` 的 `<style scoped>` 末尾添加：

```css
@media (max-width: 63.9375rem) {
  .bulk-action-bar {
    align-items: stretch;
    flex-direction: column;
  }

  .bulk-action-bar__actions {
    width: 100%;
    flex-wrap: wrap;
  }
}
```

在 `CommandPalette.vue` 中保留 `.command-palette__input { outline: 0; }`，并紧随 `.command-palette__search` 添加：

```css
.command-palette__search:focus-within {
  box-shadow: 0 0 0 3px var(--line-focus);
}
```

在 `.command-palette__list` 声明中加入：

```css
overscroll-behavior: contain;
```

不得修改 `open`、`close`、`onPanelKeydown`、焦点 trap、快捷键监听、路由跳转或动画。

- [ ] **Step 4: 运行绿灯测试和定向构建检查**

Run:

```bash
cd admin-web && npm test -- src/components/base/precisionOpsComponents.spec.js
```

Expected: PASS；共享组件原有断言和新增窄屏、焦点、滚动链断言全部通过。

- [ ] **Step 5: 记录并提交 Task 1**

在 `plan.md` 顶部追加红灯结果、修改文件和绿灯命令。然后：

```bash
git add admin-web/src/components/base/BulkActionBar.vue admin-web/src/components/base/CommandPalette.vue admin-web/src/components/base/precisionOpsComponents.spec.js plan.md
git commit -m "修复管理端共享交互布局"
```

Expected: 只提交本任务四个文件；`.superpowers/` 保持未跟踪。

### Task 2: Drawer 密度、宽表与视频详情收纳

**Files:**
- Modify: `admin-web/src/views/VideoList.vue:1506-2038,2211-2242,2279-2308`
- Modify: `admin-web/src/views/ImageManage.vue:1251-1367,1556-1750`
- Modify: `admin-web/src/views/ImageCollectionManage.vue:554-680,684-905`
- Modify: `admin-web/src/views/ActorManage.vue:437-535`
- Modify: `admin-web/src/views/CollectionManage.vue:280-321`
- Modify: `admin-web/src/views/UserManage.vue:196-237`
- Modify: `admin-web/src/views/videoListPage.spec.js:540-650`
- Modify: `admin-web/src/views/imageManagePage.spec.js:620-700`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js:1000-1080`

**Interfaces:**
- Consumes: `data-density="form"` 已有全局 Element Plus form 密度规则；VideoList 的 `detailDrawerSize`、`subtitleItems` 和候选数据；ImageManage 的 `uploadSummary`；图片合集的既有 Drawer 关闭守卫和选择状态。
- Produces: 表单型 Drawer 的显式 form 密度、`.drawer-table-wrap` 局部横滚边界、视频季/集窄屏网格、视频候选/图片合集元信息的长文本收敛。

- [ ] **Step 1: 写入会失败的视频、图片和集合页测试**

在 `videoListPage.spec.js` 的视频 Drawer 测试中加入以下断言：

```js
expect(template).toMatch(/<el-drawer\b(?=[^>]*v-model="detailVisible")(?=[^>]*data-density="form")[^>]*>/)
expect(template).toContain('<div class="drawer-table-wrap">\n            <el-table :data="subtitleItems"')
expect(findRule(style.slice(style.indexOf('@media (max-width: 63.9375rem)')), '.episode-fields')).toContain('grid-template-columns: repeat(2, minmax(0, 1fr))')
expect(findRule(style.slice(style.indexOf('@media (max-width: 63.9375rem)')), '.episode-fields :deep(.el-input-number)')).toContain('width: 100%')
expect(findRule(style.slice(style.indexOf('@media (max-width: 63.9375rem)')), '.tv-pending-candidate')).toContain('flex-wrap: wrap')
expect(findRule(style.slice(style.indexOf('@media (max-width: 63.9375rem)')), '.tv-pending-candidate > span')).toContain('overflow-wrap: anywhere')
expect(findRule(style, '.drawer-table-wrap')).toContain('overflow-x: auto')
```

在 `imageManagePage.spec.js` 的上传/Drawer 测试中断言：

```js
expect(template).toMatch(/<el-drawer\b(?=[^>]*v-model="uploadDialogVisible")(?=[^>]*data-density="form")[^>]*>/)
expect(template).toContain('<div class="drawer-table-wrap">\n          <el-table :data="uploadSummary.items || []"')
expect(findRule(style, '.drawer-table-wrap')).toContain('min-width: 0')
expect(findRule(style, '.drawer-table-wrap')).toContain('overflow-x: auto')
```

在 `precisionOpsRollout.spec.js` 增加集合 Drawer 断言：Actor、Collection、User、ImageCollection 的 CRUD/Edit Drawer opening tag 都带 `data-density="form"`；图片合集排序输入包含 `style="width: min(200px, 100%)"`；`.drawer-cover-meta` 包含 `min-width: 0`，且 `.drawer-cover-title`、`.drawer-cover-desc`、`.drawer-cover-note` 都包含 `overflow-wrap: anywhere`。

- [ ] **Step 2: 运行红灯测试并确认缺少布局契约**

Run:

```bash
cd admin-web && npm test -- src/views/videoListPage.spec.js src/views/imageManagePage.spec.js src/views/precisionOpsRollout.spec.js
```

Expected: FAIL；现有 Drawer 缺少 form 密度属性、两处宽表没有容器、季/集和长文本窄屏规则不存在，图片合集排序仍为固定 200px。

- [ ] **Step 3: 为 Teleport Drawer 声明 form 密度并包裹宽表**

在下列 `<el-drawer>` opening tag 增加 `data-density="form"`，属性与既有 `class`、`v-model`、关闭守卫保持不变：

```text
ActorManage.vue: 编辑/创建演员 Drawer
CollectionManage.vue: 编辑/创建合集 Drawer
UserManage.vue: 编辑用户 Drawer
ImageManage.vue: 新增图片、图片详情、更多筛选 Drawer
VideoList.vue: 批量编辑、视频详情、更多筛选 Drawer
ImageCollectionManage.vue: 编辑图片合集、合集关联图片 Drawer
```

在 `VideoList.vue` 的现有字幕表开始标签 `<el-table :data="subtitleItems" border size="small" v-loading="subtitleLoading">` 之前插入 `<div class="drawer-table-wrap">`，并在该同一张表的 `</el-table>` 之后插入 `</div>`。表内现有来源、语言、名称、格式、默认、附加信息、操作列及其全部 slot、按钮和 handler 必须逐字保持不变。

在 `ImageManage.vue` 中仅将上传结果 `<el-table :data="uploadSummary.items || []" ...>` 包成同名容器，保留既有五列和 `uploadSummary` 的显示条件。

在两个页面各自的 scoped style 加入：

```css
.drawer-table-wrap {
  width: 100%;
  min-width: 0;
  overflow-x: auto;
}
```

- [ ] **Step 4: 收纳视频详情和图片合集动态内容**

在 `VideoList.vue` 的 `@media (max-width: 63.9375rem)` 中加入：

```css
.episode-fields {
  display: grid;
  width: 100%;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.episode-fields :deep(.el-input-number) {
  width: 100%;
}

.tv-pending-candidate {
  flex-wrap: wrap;
}

.tv-pending-candidate > span {
  min-width: 0;
  overflow-wrap: anywhere;
}
```

把 `ImageCollectionManage.vue` 的排序控件从 `style="width: 200px"` 改为：

```vue
<el-input-number v-model="form.sort_order" :step="1" :precision="0" style="width: min(200px, 100%)" />
```

并在图片合集 scoped style 中写入：

```css
.drawer-cover-meta {
  min-width: 0;
}

.drawer-cover-title,
.drawer-cover-desc,
.drawer-cover-note {
  min-width: 0;
  overflow-wrap: anywhere;
}
```

不得修改季/集的 v-model、候选的 key/text/handler、图片合集 Drawer size、`before-close`、封面逻辑或分页。

- [ ] **Step 5: 运行绿灯测试和 SFC 编译检查**

Run:

```bash
cd admin-web && npm test -- src/views/videoListPage.spec.js src/views/imageManagePage.spec.js src/views/precisionOpsRollout.spec.js
```

Expected: PASS；所有既有页面行为测试与新增 Drawer 布局契约均通过。

- [ ] **Step 6: 记录并提交 Task 2**

在 `plan.md` 顶部追加红灯、实现、绿灯结果。然后：

```bash
git add admin-web/src/views/VideoList.vue admin-web/src/views/ImageManage.vue admin-web/src/views/ImageCollectionManage.vue admin-web/src/views/ActorManage.vue admin-web/src/views/CollectionManage.vue admin-web/src/views/UserManage.vue admin-web/src/views/videoListPage.spec.js admin-web/src/views/imageManagePage.spec.js admin-web/src/views/precisionOpsRollout.spec.js plan.md
git commit -m "收纳管理端 Drawer 内容布局"
```

Expected: 仅暂存任务列出的页面、测试和本任务账本记录。

### Task 3: 集合文本、监控列宽与待删除中等宽度布局

**Files:**
- Modify: `admin-web/src/views/ActorManage.vue:395-420,537-620`
- Modify: `admin-web/src/views/TaskMonitor.vue:250-293`
- Modify: `admin-web/src/views/PendingDeleteShorts.vue:400-690`
- Modify: `admin-web/src/views/taskMonitorPage.spec.js:250-360`
- Modify: `admin-web/src/views/precisionOpsRollout.spec.js:560-630,1030-1060`

**Interfaces:**
- Consumes: Actor `row.aliases` 既有数组数据、TaskMonitor 的九列和格式化 helper、PendingDeleteShorts 既有队列/播放器/操作函数。
- Produces: `actorAliases(row)` 的显示格式、紧凑 Tooltip 触发器、1440px 预算内的任务表列宽，以及仅在 1024–1199px 生效的待删除详情纵向布局。

- [ ] **Step 1: 写入会失败的集合/监控布局测试**

在 `taskMonitorPage.spec.js` 的“保留完整任务列”测试中追加精确列预算：

```js
expect(template).toContain('label="任务" min-width="210"')
expect(template).toContain('label="状态" width="96"')
expect(template).toContain('label="进度" min-width="150"')
expect(template).toContain('label="剩余时间" width="88"')
expect(template).toContain('label="已耗时" width="88"')
expect(template).toContain('prop="retry_count" label="重试" width="60"')
expect(template).toContain('label="错误" min-width="160"')
expect(template).toContain('label="开始时间" width="145"')
expect(template).toContain('label="进度更新时间" width="145"')
```

在 `precisionOpsRollout.spec.js` 的 `extractStyle` helper 后新增：

```js
function findRule(styleSource, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = styleSource.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}
```

在阶段二资源名称测试中加入演员别名断言：

```js
const actor = readView('ActorManage.vue')
const actorStyle = extractStyle(actor)
expect(actor).toContain('function actorAliases(row)')
expect(actor).toContain('<el-tooltip :content="actorAliases(row)" placement="top">')
expect(actor).toContain('class="actor-aliases" tabindex="0" :aria-label="`演员别名：${actorAliases(row)}`"')
expect(findRule(actorStyle, '.actor-aliases')).toContain('text-overflow: ellipsis')
expect(findRule(actorStyle, '.actor-aliases')).toContain('white-space: nowrap')
expect(findRule(actorStyle, '.actor-aliases:focus-visible')).toContain('outline: 2px solid var(--line-focus)')
```

同一测试读取 PendingDeleteShorts style，并断言介质查询包含 `min-width: 64rem`、`max-width: 74.9375rem`，其中 `.pending-delete-detail` 为 `flex-direction: column`，`.pending-delete-detail__controls` 为 `width: 100%` 和 `justify-content: flex-start`。

- [ ] **Step 2: 运行红灯测试并确认旧尺寸不满足预算**

Run:

```bash
cd admin-web && npm test -- src/views/taskMonitorPage.spec.js src/views/precisionOpsRollout.spec.js
```

Expected: FAIL；TaskMonitor 仍为旧列宽，Actor 没有别名 Tooltip，PendingDeleteShorts 没有 1024–1199px 的局部详情布局规则。

- [ ] **Step 3: 收敛任务表列宽并实现演员别名显示 helper**

在 `TaskMonitor.vue` 保留九列顺序、标签、slot、状态组件、进度条和日期格式，只替换列参数为：

```vue
<el-table-column prop="video_title" label="任务" min-width="210">
<el-table-column prop="status" label="状态" width="96">
<el-table-column label="进度" min-width="150">
<el-table-column label="剩余时间" width="88">
<el-table-column label="已耗时" width="88">
<el-table-column prop="retry_count" label="重试" width="60" />
<el-table-column prop="error" label="错误" min-width="160">
<el-table-column label="开始时间" width="145">
<el-table-column label="进度更新时间" width="145">
```

在 `ActorManage.vue` 的 `sourceLabel` 一类展示 helper 附近新增：

```js
function actorAliases(row) {
  const aliases = Array.isArray(row?.aliases)
    ? row.aliases.map((item) => String(item || '').trim()).filter(Boolean)
    : []
  return aliases.length > 0 ? aliases.join(' / ') : '暂无'
}
```

将别名列模板改为：

```vue
<el-tooltip :content="actorAliases(row)" placement="top">
  <span class="actor-aliases" tabindex="0" :aria-label="`演员别名：${actorAliases(row)}`">
    {{ actorAliases(row) }}
  </span>
</el-tooltip>
```

在 Actor scoped style 添加：

```css
.actor-aliases {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actor-aliases:focus-visible {
  outline: 2px solid var(--line-focus);
  outline-offset: 2px;
}
```

- [ ] **Step 4: 只在中等桌面宽度堆叠待删除详情**

在 `PendingDeleteShorts.vue` 现有 `@media (max-width: 1023px)` 之前增加：

```css
@media (min-width: 64rem) and (max-width: 74.9375rem) {
  .pending-delete-detail {
    flex-direction: column;
    gap: var(--space-3);
  }

  .pending-delete-detail__controls {
    width: 100%;
    justify-content: flex-start;
  }
}
```

不得改动 1023px 以下单列 workbench、1200px 及以上横向详情、队列/播放器高度、选择、保留、最终删除或分页。

- [ ] **Step 5: 运行绿灯测试**

Run:

```bash
cd admin-web && npm test -- src/views/taskMonitorPage.spec.js src/views/precisionOpsRollout.spec.js
```

Expected: PASS；九列仍完整，收紧预算已锁定，别名完整值仍可访问，待删除工作台仅在指定中等宽度堆叠详情。

- [ ] **Step 6: 记录并提交 Task 3**

在 `plan.md` 顶部追加红灯、实现、绿灯结果。然后：

```bash
git add admin-web/src/views/ActorManage.vue admin-web/src/views/TaskMonitor.vue admin-web/src/views/PendingDeleteShorts.vue admin-web/src/views/taskMonitorPage.spec.js admin-web/src/views/precisionOpsRollout.spec.js plan.md
git commit -m "收敛管理端集合页面布局"
```

Expected: 只提交本任务的三页、两份测试和账本记录。

### Task 4: 全量验证、真实浏览器验收与交付记录

**Files:**
- Modify: `plan.md: 顶部追加最终验证记录`
- Modify: `CONTEXT.md: 仅当本轮得到新的长期维护约定时追加；否则不修改`

**Interfaces:**
- Consumes: Tasks 1–3 已通过的源码契约；开发代理和受控测试管理员浏览器会话。
- Produces: 四档可复现的浏览器验证结果、干净差异检查、完整管理端测试/构建证据和最终提交。

- [ ] **Step 1: 运行全量自动验证**

Run:

```bash
cd admin-web && npm test
cd admin-web && npm run build
```

Expected: 两条命令退出码均为 0；仅可保留当前 Vite bundle 体积 warning，不接受测试失败、SFC 编译错误或新增 warning。

- [ ] **Step 2: 执行四视口只读浏览器验收**

建立本地开发服务器和受控测试管理员会话；不得把凭据写入 shell 命令、浏览器日志、截图文件名、Markdown 或 Git。对 `375x812`、`768x900`、`1024x900`、`1440x900` 逐路由运行：

```js
({
  path: location.pathname,
  scrollWidth: document.documentElement.scrollWidth,
  clientWidth: document.documentElement.clientWidth
})
```

Expected: 每条普通页面均 `scrollWidth === clientWidth`。表格必须通过 `.table-wrap` 或 `.drawer-table-wrap` 访问横向内容，不得让 `html`、`body` 或 Drawer body 横滚。

重点人工流程：

```text
375px：选中视频和图片，确认三项批量操作完整可点；打开命令面板，Tab 到搜索框后焦点环可见，滚动结果到边界不带动页面；打开视频详情，检查季/集、候选长标题和字幕表；打开图片上传 Drawer，检查结果表只在自身横滚；检查图片合集编辑排序输入、关联图片 Drawer 的长无空格文本。
768px：确认所有 form Drawer 内按钮、输入和选择器至少 44px，关闭按钮与 footer 均可键盘到达。
1024px：确认待删除工作台保持队列/播放器双栏，但详情标题、事实和操作按纵向排布且无挤压。
1440px：确认任务监控的九列均同时可见，且视频、图片、IPTV、安装包、演员、压缩包导入和图像工作台仍保持已确认的高密度布局。
```

危险确认对话框只按 Escape 或取消关闭；不得执行发布、删除、停用、上传、生成或其它直接写操作。

- [ ] **Step 3: 执行静态与编码检查**

Run:

```bash
git diff --check
replacement_status=0
replacement_output="$(LC_ALL=C rg -n $'\xEF\xBF\xBD' CONTEXT.md plan.md docs/superpowers/specs docs/superpowers/plans admin-web/src 2>&1)" || replacement_status=$?
test "$replacement_status" -eq 1
test -z "$replacement_output"
git status --short
```

Expected: `git diff --check` 无输出；替换字符扫描无输出且 `replacement_status` 为 1；状态中除本任务文件和既有 `.superpowers/` 外没有意外文件。

- [ ] **Step 4: 记录最终证据并提交 Task 4**

在 `plan.md` 顶部追加：完整 `npm test` / `npm run build` 结果、四视口结果、浏览器重点流程、静态检查、纳入文件和明确未纳入的 `.superpowers/`。如果没有新长期术语或跨任务约定，不修改 `CONTEXT.md`。

```bash
git add plan.md
git commit -m "完成管理端布局健壮性验证"
git status --short
```

Expected: 最终提交仅含 `plan.md`；工作区仅保留既有未跟踪 `.superpowers/`，无本任务未提交文件。

## 计划自检

- 规格 3.1 的批量操作条、命令面板、Drawer form 密度和 Drawer 宽表容器由 Task 1–2 覆盖。
- 规格 3.2 的九列任务表预算与演员别名由 Task 3 覆盖。
- 规格 3.3 的待删除中等宽度布局、视频季/集与候选动态文本由 Task 2–3 覆盖。
- 规格 3.4 的图片合集输入和动态文本由 Task 2 覆盖。
- 规格 4 的自动验证、四视口浏览器验收、静态检查与提交范围由 Task 4 覆盖。
- 已检查本文件不存在 `TODO`、`TBD`、未定义函数、未定义接口或“类似前一任务”式占位步骤；所有生产改动均有先行红灯断言和精确验证命令。
