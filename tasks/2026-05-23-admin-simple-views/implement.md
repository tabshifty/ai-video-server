# Implement：admin 重设计 Phase 2 · 简单视图重排

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Phase 1：`tasks/2026-05-23-admin-shell-redesign/DONE.md`
- 任务三段执行流：当前阶段 = Implement

## 1. 实施顺序（TDD 友好）

1. **扩展 spec 先**：在 `assets/themeTokens.spec.js` 新增视图层 audit 断言（先红）
2. **`Dashboard.vue` 第一**：跑 spec 红，重排后绿
3. **`SystemSettings.vue`** / **`UserManage.vue`** / **`TaskMonitor.vue`** / **`IPTVManage.vue`** / **`CollectionManage.vue`** / **`ActorManage.vue`** 依次推进
4. **`UploadProgress.vue` 1 处 hex 清理**跟随首个用到它的视图一并提交
5. **手测 H21–H27** + 截图归档
6. **`npm run build` + `npm test`** 全绿
7. **追加 `plan.md` + commit**

## 2. 通用重排配方（每视图按此模板）

```vue
<script setup>
// 原 imports 保留
// 新增（按需）：
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'         // 仅当有筛选/操作按钮时
import StatCard from '../components/base/StatCard.vue'       // 仅当有指标块时
import SectionCard from '../components/base/SectionCard.vue' // 适度接入
import EmptyState from '../components/base/EmptyState.vue'   // 全部接入
// ...原有 logic 保留不动
</script>

<template>
  <Layout>
    <div class="page-shell">
      <PageHeader title="<视图中文名>">
        <template #actions>
          <!-- 顶部级别 actions，例如 Dashboard 的「刷新」 -->
        </template>
      </PageHeader>

      <Toolbar v-if="<有筛选/操作按钮>">
        <template #filters><!-- chip / select / search --></template>
        <template #actions><!-- 主操作按钮 --></template>
      </Toolbar>

      <!-- 指标块（仅 Dashboard / TaskMonitor / IPTVManage） -->
      <div v-if="<有指标>" class="stats-grid">
        <StatCard v-for="..." :label="..." :value="..." />
      </div>

      <SectionCard v-if="<有分区>">
        <template #title>...</template>
        <!-- body slot 是原 .content-card / .table-panel 的内容 -->
      </SectionCard>

      <!-- 列表 / 表格 / 表单 内容 -->
      <template v-if="<有数据>">...</template>
      <EmptyState v-else title="..." description="..." />
    </div>
  </Layout>
</template>

<style scoped>
/* 1. 删除原 .page-header / .section-head / .metric-card 自定义样式（已被 wrapper 接管） */
/* 2. 删除/替换所有硬编码 hex：
     - #7f1d1d / #881337 / #be123c → var(--primary) 或 var(--primary-strong)
     - #6b7280 / #4b5563 / #9ca3af → var(--text-secondary) 或 var(--text-muted)
     - #e5e7eb / #cad8f5 → var(--line-soft) 或 var(--primary-soft)
   */
/* 3. 保留视图特有的布局 / 间距 / 组件细节调整 */
</style>
```

## 3. 每视图具体改动点

### 3.1 `views/Dashboard.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | 替换 `.page-header` → `<PageHeader title="系统仪表盘" />`；删除「管理员工作台」副标题 |
| 8 张卡 | 替换 `<article class="content-card metric-card">` → `<StatCard :label="..." :value="..." />`；`metric-card--accent` 用户总数那张用 `<StatCard variant="accent" />`（若 StatCard 支持，否则保留默认） |
| 图表 | 包入 `<SectionCard><template #title>近 7 天上传趋势</template>...</SectionCard>` |
| ECharts 主题 | `lineStyle.color: '#e11d48'` → `getComputedStyle(document.documentElement).getPropertyValue('--primary').trim()`；`itemStyle.color: '#2563eb'` 同理；`areaStyle.colorStops` 用 token 派生（`color-mix` 或预先计算的 `--primary-soft`）；`axisLine.lineStyle.color: '#fda4af'` 改 `--line-soft`；`axisLabel.color: '#6b7280'` 改 `--text-secondary`；`splitLine.lineStyle.color: 'rgba(136,19,55,0.08)'` 改 `--line-soft` |
| 失败态 | 加载失败时显示 `<EmptyState title="加载失败" description="..." />` 含「重试」按钮（用 `<template #action>`） |
| Scoped CSS | 删除 `.metric-card` / `.metric-card__label` / `.metric-card__value` / `.metric-card--accent` / `.trend-panel` 全部规则；保留 `.trend-chart` 高度声明；删除 `border-color: #cad8f5` |

### 3.2 `views/SystemSettings.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | 替换头部 → `<PageHeader title="系统设置" />` |
| 表单分组 | 按现有 `el-form` 字段语义切分到 N 个 `<SectionCard>`：基础 / 转码 / 刮削 / 安全等 |
| Scoped CSS | 删 `#7f1d1d` / `#9f1239` / `#6b7280` ×2 / `#e5e7eb`，全部换 token |

### 3.3 `views/UserManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="用户管理" />` |
| 操作行 | `<Toolbar><template #actions><el-button @click="refresh">刷新</el-button><el-button type="primary" @click="onCreate">添加用户</el-button></template></Toolbar>` |
| 表格 | 列结构保留；`v-if="users.length"` 否则 `<EmptyState title="暂无用户" />` |
| 编辑 dialog | **保留**（Phase 4 才改 drawer） |
| Scoped CSS | 删 `#6b7280` / `#7f1d1d`，全部换 token |

### 3.4 `views/TaskMonitor.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="任务监控" />` |
| 操作行 | `<Toolbar><template #filters><!-- 状态 chip 切换 --></template><template #actions><el-button @click="refresh">刷新</el-button></template></Toolbar>` |
| 指标 | 顶部 `<StatCard label="队列长度" :value="stats.pending" />` / `<StatCard label="处理中" :value="stats.running" />` / `<StatCard label="失败" :value="stats.failed" />` |
| 列表 | 无任务时 `<EmptyState title="暂无任务" />` |
| 自动刷新 | 既有 `setInterval` / `setTimeout` 计时器**不动**，仅外层 wrapper 改 |

### 3.5 `views/IPTVManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="IPTV 管理" />` |
| 操作行 | `<Toolbar><template #actions><el-button @click="refresh">刷新</el-button><el-button @click="onUploadM3u">上传 M3U</el-button><el-button @click="onPullRemote">远程拉取</el-button></template></Toolbar>` |
| 指标 | `<StatCard label="频道总数" :value="..." />` / `<StatCard label="分组数" :value="..." />`（按既有 `stat-value` 元素一一对应迁移） |
| M3U 源 | 包入 `<SectionCard><template #title>M3U 源</template>...</SectionCard>` |
| 频道列表 | 无频道时 `<EmptyState title="暂无频道" description="上传 M3U 文件或填入远程地址刷新" />` |
| Scoped CSS | 删 `#7f1d1d` ×3，全部换 `var(--primary)` |

### 3.6 `views/CollectionManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="合集管理" />` |
| 操作行 | `<Toolbar><template #filters><el-input v-model="search" /></template><template #actions><el-button type="primary" @click="onCreate">创建合集</el-button></template></Toolbar>` |
| 列表 | 无合集时 `<EmptyState title="暂无合集" />` |
| 编辑 dialog | **保留** |

### 3.7 `views/ActorManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="演员管理" />` |
| 操作行 | `<Toolbar><template #filters><el-input v-model="search" /></template><template #actions><el-button type="primary" @click="onCreate">创建演员</el-button></template></Toolbar>` |
| 列表 | 无演员时 `<EmptyState title="暂无演员" />` |
| 编辑 dialog | **保留** |
| Scoped CSS | 删 `#9ca3af` / `#6b7280` / `#4b5563`，全部换 token |

### 3.8 `components/UploadProgress.vue`

| 改动 | 描述 |
|------|------|
| Scoped CSS | line 28 `color: #7f1d1d` → `color: var(--primary)` |

## 4. spec 扩展

### `admin-web/src/assets/themeTokens.spec.js`

新增 audit 块（追加到现有 `describe('theme tokens')` 之后）：

```js
import { readFileSync } from 'node:fs'
import { describe, it, expect } from 'vitest'

const ROSE_HEX = [/#881337/i, /#be123c/i, /#7f1d1d/i]
const PHASE2_VIEWS = [
  'src/views/Dashboard.vue',
  'src/views/SystemSettings.vue',
  'src/views/UserManage.vue',
  'src/views/TaskMonitor.vue',
  'src/views/IPTVManage.vue',
  'src/views/CollectionManage.vue',
  'src/views/ActorManage.vue',
  'src/components/UploadProgress.vue',
]

describe('phase 2 views hex audit', () => {
  PHASE2_VIEWS.forEach((path) => {
    it(`${path} 不含玫红 hex 字面量`, () => {
      const source = readFileSync(path, 'utf8')
      ROSE_HEX.forEach((pattern) => {
        expect(source).not.toMatch(pattern)
      })
    })
  })
})
```

预期：
- **实施前**：8 个用例 ≥ 6 个红（仅 `Dashboard.vue` 与 `CollectionManage.vue` / `TaskMonitor.vue` 绿）
- **实施后**：8 个用例全绿

## 5. plan.md 追加模板

```markdown
## 2026-05-24 · admin Phase 2：简单视图重排

- 摘要：admin-web 全面重设计第 2 阶段——7 个简单视图（Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage）按 Phase 1 设计系统接入 PageHeader / Toolbar / StatCard / SectionCard / EmptyState 五个共享 wrapper；同时清零 8 文件 15 处硬编码玫红/灰 hex 字面量；ECharts 主色 token 化；UploadProgress 共享组件顺手清 1 处 hex。
- 受影响文件：admin-web/src/views/{Dashboard,SystemSettings,UserManage,TaskMonitor,IPTVManage,CollectionManage,ActorManage}.vue、admin-web/src/components/UploadProgress.vue、admin-web/src/assets/themeTokens.spec.js（扩 audit）、tasks/2026-05-23-admin-simple-views/{prd.md,implement.md,review.md,screenshots/,DONE.md}、plan.md
- 验证：cd admin-web && npm run build 通过；cd admin-web && npm test 全绿（67 + 8 audit = 75 用例）；H21–H27 手测通过；14 张截图归档。
```

## 6. 风险与回退

- **风险 1**：`StatCard` 接入 Dashboard 时发现 `metric-card--accent` 的「accent 变体」需求未在 Phase 1 wrapper API 中暴露——对策：先用 `<StatCard variant="accent" />` 试，如果 wrapper 不支持则降级用默认形态 + 内部说明（不让单点视觉差异阻塞整体推进）。如需新增 prop，回 Phase 1 followup commit 补
- **风险 2**：`SectionCard` 包 ECharts 时，`SectionCard` 的 padding 可能挤压 ECharts 容器导致 `chart.resize()` 计算出错——对策：观察后用 SectionCard 的 body slot 内 `padding: 0`（如果 API 暴露）或 wrapper 调整
- **风险 3**：ECharts 主色用 `getComputedStyle` 读 token 在 `theme.css` 加载前可能空字符串——对策：在 `renderChart()` 内 `nextTick` 后读取；或读取后 fallback 到 `'#2563eb'`
- **风险 4**：`Toolbar` 的 `#filters` 和 `#actions` 双 slot 间距如果不符合 Phase 1 设计，需要观察实际效果；不达预期则回 Phase 1 followup
- **回退**：所有改动是「替换 wrapper + 删除 scoped CSS 硬编码」纯替换改动；恢复原文件即可回到 Phase 2 实施前（保留 Phase 1 全部成果）
