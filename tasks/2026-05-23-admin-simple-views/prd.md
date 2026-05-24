# PRD：admin 重设计 Phase 2 · 简单视图重排

- 日期：2026-05-24
- 阶段：Phase 2 / 4（简单视图重排）
- 前置任务：`tasks/2026-05-23-admin-shell-redesign/DONE.md`（Phase 1 已 Done，commit `12bf7c93` + `5ae43fa5`）
- 范围：7 个简单视图按 Phase 1 设计系统重排 + 视图层硬编码 hex 全部清零

## 1. 用户故事

作为管理员，在 Phase 1 设计系统地基铺好后，希望：

1. **进入仪表盘**就能看到 8 个统一字号 / 字体 / token 的 StatCard 指标卡片，不再是裸 `.metric-card` 自定义 div + 玫红色调（旧 `#e11d48`）
2. **系统设置 / 用户管理 / 任务监控 / IPTV / 合集 / 演员**6 个视图的顶部头部统一用 `PageHeader`、操作行统一用 `Toolbar`、空数据统一用 `EmptyState`，无残留「管理员工作台」副标题
3. **整站冷蓝**——任何视图的 scoped CSS 不再出现 `#7f1d1d` / `#881337` / `#be123c` 玫红 hex，与 Phase 1 已清洁的外壳保持一致
4. **共享 `UploadProgress.vue`** 也清零硬编码 hex，避免被 VideoUpload / VideoList 等下游视图复用时反弹
5. **功能零回归**——任何按钮 / 表单 / 表格 / dialog / 自动刷新行为不变，只是视觉切到新设计系统

## 2. 作用域

| 改 | 不改 |
|----|------|
| `admin-web/src/views/Dashboard.vue` 全文重排（PageHeader + StatCard×8 + SectionCard + EmptyState + ECharts 主色 token 化） | 路由 URL |
| `admin-web/src/views/SystemSettings.vue` 全文重排（PageHeader + SectionCard×N 表单分组） | Pinia store 形状 |
| `admin-web/src/views/UserManage.vue` 全文重排（PageHeader + Toolbar + EmptyState） | API 接口 / 后端契约 |
| `admin-web/src/views/TaskMonitor.vue` 全文重排（PageHeader + Toolbar + StatCard×3 + EmptyState） | 编辑 / 上传 / 创建 dialog 形态（Phase 4 才改 drawer） |
| `admin-web/src/views/IPTVManage.vue` 全文重排（PageHeader + StatCard×N + SectionCard + Toolbar + EmptyState） | 表格列结构 / 列宽 |
| `admin-web/src/views/CollectionManage.vue` 全文重排（PageHeader + Toolbar + EmptyState） | 表单字段顺序 |
| `admin-web/src/views/ActorManage.vue` 全文重排（PageHeader + Toolbar + EmptyState） | ECharts 图表类型与数据维度 |
| `admin-web/src/components/UploadProgress.vue` scoped CSS 清 `#7f1d1d` 1 处 → `var(--primary)` | Phase 3 / Phase 4 域内视图 |
| `admin-web/src/assets/themeTokens.spec.js` 扩展视图层 audit | `Layout.vue` / 7 个 base wrapper / Login（Phase 1 已落） |
| `tasks/2026-05-23-admin-simple-views/screenshots/` 14 张 PNG | `#app :where(...)` specificity 全局收口（留单独 follow-up） |

## 3. 非目标

- N1：本期**不动**三巨头（`VideoList` / `ImageManage` / `VideoUpload`），留给 Phase 4
- N2：本期**不动** Phase 3 视图（`ScrapePreview` / `AVManualScrape` / `TvSeriesManage` / `ImageCollectionManage`）
- N3：本期**不改** API 接口、Pinia store 形状、路由 URL
- N4：本期**不改** ECharts 图表类型与数据维度（仅更新主题色 token）
- N5：本期**不抽** 7 视图的共享业务逻辑——视觉重排不是重构
- N6：本期**不引入新** wrapper 组件（不新加 `PageActions` / `Breadcrumb` / `KbdLabel` 等）；如某视图发现 Phase 1 七个 wrapper 不够，回退到 Phase 1 followup commit 而非新增组件
- N7：本期**不做** `#app :where(...)` 全局 specificity 重构（涉及 13 处 `#app` 前缀的级联调整，留单独 follow-up）；如某视图被压制，就地用更高 specificity 的 scoped 选择器解决
- N8：本期**不引入** Percy / Chromatic / Playwright 视觉差异 diff
- N9：本期**不接入** dark mode UI 切换
- N10：本期**不动** 编辑 / 创建 / 上传 dialog 的形态（drawer 改造是 Phase 4 + 部分 Phase 3 视图的范围）
- N11：本期**不写** 视图渲染单测（成本远高于收益）；靠源文 audit + 手测

## 4. Phase 2 关键决策结论（grill-with-docs 结果）

| # | 决策点 | 选项 |
|---|--------|------|
| Q1 | Phase 2 是否一并清 scoped CSS 硬编码 hex | **是**：7 视图 + `UploadProgress.vue` 共 8 文件、15 处硬编码 hex 一并清零 |
| Q2 | 每视图 wrapper 接入范围 | **按视图实际形态匹配**：PageHeader 全 7 视图；Toolbar 仅有「筛选 / 操作按钮」时接入；StatCard 仅 Dashboard / TaskMonitor / IPTVManage；SectionCard 适度接入；EmptyState 全 7 视图；BulkActionBar / CommandPalette 不接入 |
| Q3 | 改造顺序 + 提交策略 | **Dashboard 第一**（验证 StatCard 真实场景）→ 按 el-* 密度升序到 ActorManage；**单 commit**；如有 followup code review 则同 Phase 1 走「实现 + followup」二次 commit |
| Q4 | 验收颗粒度 | **每视图一条 H 用例（H21–H27）**，按「外观锚点 + 功能不破」二维；**加一条 H28** 做全局 spec + build 验证 |

## 5. 视图重排映射表

| # | 视图 | 必接入 wrapper | scoped CSS 必清 hex（行 / 色值） |
|---|------|---------------|---------------------------------|
| 1 | `Dashboard.vue` | `PageHeader` + `StatCard` ×8（替换 `.metric-card`）+ `SectionCard`（包 ECharts）+ `EmptyState`（stats 加载失败） | `175: #cad8f5`（→ `--line-soft` 或 `--primary-soft`） |
| 2 | `SystemSettings.vue` | `PageHeader` + `SectionCard` ×N（按表单分组：基础 / 转码 / 刮削 / 安全等） | `102: #7f1d1d` / `123: #9f1239` / `130: #6b7280` / `143: #6b7280` / `156: #e5e7eb`（→ `--primary` / `--primary-strong` / `--text-secondary` / `--text-secondary` / `--line-soft`） |
| 3 | `UserManage.vue` | `PageHeader` + `Toolbar`（刷新 + 添加用户）+ `EmptyState` | `114: #6b7280` / `120: #7f1d1d`（→ `--text-secondary` / `--primary`） |
| 4 | `TaskMonitor.vue` | `PageHeader` + `Toolbar`（刷新 + 状态筛选 chip）+ `StatCard` ×3（队列 / 失败 / 处理中）+ `EmptyState` | 0 处（已在 Phase 1 commit 中预清） |
| 5 | `IPTVManage.vue` | `PageHeader` + `StatCard` ×N（频道总数 / 分组数等）+ `SectionCard`（M3U 源配置）+ `Toolbar`（刷新 / 上传 / 远程拉取）+ `EmptyState` | `267 / 289 / 304: #7f1d1d` ×3（→ `--primary`） |
| 6 | `CollectionManage.vue` | `PageHeader` + `Toolbar`（创建 + 搜索）+ `EmptyState` | 0 处 |
| 7 | `ActorManage.vue` | `PageHeader` + `Toolbar`（搜索 + 创建演员）+ `EmptyState` | `482: #9ca3af` / `536: #6b7280` / `542: #4b5563`（→ `--text-muted` / `--text-secondary` / `--text-primary`） |
| 共享 | `components/UploadProgress.vue` | — | `28: #7f1d1d`（→ `--primary`） |

## 6. 改造顺序

1. `Dashboard.vue`（先攻；StatCard 8 倍验证）
2. `SystemSettings.vue`（5 个 el-* / SectionCard 分组）
3. `UserManage.vue`（7 个 el-* / Toolbar 双按钮）
4. `TaskMonitor.vue`（13 个 el-* / Toolbar + StatCard 混合 + 自动刷新计时器兼容）
5. `IPTVManage.vue`（17 个 el-* / 全 wrapper 大集合）
6. `CollectionManage.vue`（33 个 el-*）
7. `ActorManage.vue`（50 个 el-* / 大头）

`UploadProgress.vue` 的 1 处 hex 清理跟随 `Dashboard.vue` 或 `UserManage.vue`（哪个先用到就清）一并提交，不单独提。

## 7. 验收（H21–H28）

### H21（Dashboard）

**外观锚点**：
- ✓ 顶部 `PageHeader` 标题「系统仪表盘」，无副标题
- ✓ 8 个 `StatCard` 替换原 `.metric-card`，数字使用 `tabular-nums`
- ✓ 1 个 `SectionCard` 包裹 ECharts 趋势图
- ✓ ECharts 主线色 / 区域填充用 `--primary` 派生（不是 `#e11d48`）
- ✓ Scoped CSS 不含 `#cad8f5` 等硬编码

**功能不破**：
- ✓ `stats` API 返回的数字正确显示
- ✓ ECharts `resize` 跟随窗口宽度
- ✓ 加载失败显示 `EmptyState`，含可点击的「重试」按钮

### H22（SystemSettings）

**外观锚点**：
- ✓ 顶部 `PageHeader` 标题「系统设置」
- ✓ 表单按 N 个 `SectionCard` 分组承载（基础 / 转码 / 刮削 / 安全等）
- ✓ Scoped CSS 不含 `#7f1d1d` / `#9f1239` / `#6b7280` / `#e5e7eb`

**功能不破**：
- ✓ 表单保存 / 读取 / 校验仍生效

### H23（UserManage）

**外观锚点**：
- ✓ 顶部 `PageHeader` + `Toolbar`（刷新 + 添加用户两按钮）
- ✓ 表格无数据时显示 `EmptyState`
- ✓ Scoped CSS 不含 `#6b7280` / `#7f1d1d`

**功能不破**：
- ✓ 表格分页正常
- ✓ 编辑用户 dialog 仍弹出（保留 dialog 形态，Phase 4 才改 drawer）
- ✓ 创建 / 改角色 / 删除全可用

### H24（TaskMonitor）

**外观锚点**：
- ✓ 顶部 `PageHeader` + `Toolbar`（刷新 + 状态筛选 chip）
- ✓ 顶部 3 个 `StatCard`（队列 / 失败 / 处理中三档）
- ✓ 任务列表无数据时显示 `EmptyState`

**功能不破**：
- ✓ 自动刷新计时器仍生效
- ✓ 状态筛选 chip 切换正确过滤列表

### H25（IPTVManage）

**外观锚点**：
- ✓ 顶部 `PageHeader` + `Toolbar`（刷新 + 上传 M3U + 远程拉取）
- ✓ N 个 `StatCard`（频道总数 / 分组数等）
- ✓ `SectionCard` 承载 M3U 源配置
- ✓ 频道列表无数据时 `EmptyState`
- ✓ Scoped CSS 不含 `#7f1d1d`

**功能不破**：
- ✓ 上传 M3U 文件 / 远程拉取 / 手动刷新解析仍生效
- ✓ 频道列表显示与分组正确

### H26（CollectionManage）

**外观锚点**：
- ✓ 顶部 `PageHeader` + `Toolbar`（创建合集 + 搜索框）
- ✓ 合集列表无数据时 `EmptyState`

**功能不破**：
- ✓ 合集 CRUD + 视频绑定仍生效

### H27（ActorManage）

**外观锚点**：
- ✓ 顶部 `PageHeader` + `Toolbar`（搜索 + 创建演员）
- ✓ 演员列表无数据时 `EmptyState`
- ✓ Scoped CSS 不含 `#9ca3af` / `#6b7280` / `#4b5563`

**功能不破**：
- ✓ 演员 CRUD + 搜索 + 头像上传仍生效

### H28（自动化全局验证）

- ✓ `cd admin-web && npm test` 全绿（≥ 69 用例）
- ✓ `cd admin-web && npm run build` 通过
- ✓ `themeTokens.spec.js` 视图层 audit 命中：7 视图 + `UploadProgress.vue` 共 8 文件零玫红 hex（`#881337` / `#be123c` / `#7f1d1d`）

## 8. Done definition

- [ ] 第 7 节 H21–H27 全部手测通过
- [ ] H28 自动化全过
- [ ] 14 张截图（7 视图 × before/after）归档于 `tasks/2026-05-23-admin-simple-views/screenshots/`，统一 1440px 分辨率
- [ ] `plan.md` 顶部已追加 Phase 2 实施条目
- [ ] 全部改动落在 1 commit（中文 subject + 中文 body）；如出现 code review followup 则同 Phase 1 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-simple-views/DONE.md` 含完成日期 + commit hash + 验证摘要

## 9. 自动化必跑

```bash
cd admin-web
npm run build
npm test
```

不新增 spec 文件；仅扩展 `themeTokens.spec.js` 增加视图层 audit。
