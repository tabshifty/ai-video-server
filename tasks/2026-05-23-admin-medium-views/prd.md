# PRD：admin 重设计 Phase 3 · 中等视图重排

- 日期：2026-05-24
- 阶段：Phase 3 / 4（中等视图重排）
- 前置任务：
  - `tasks/2026-05-23-admin-shell-redesign/DONE.md`（Phase 1 已 Done，commits `12bf7c93` + `5ae43fa5` + `acbac757`）
  - `tasks/2026-05-23-admin-simple-views/DONE.md`（Phase 2 已 Done，commits `359a6489` + `5181045f` + `71ba36c6`）
- 范围：4 个中等复杂度视图按 Phase 1 设计系统重排 + ImageCollectionManage 上传 dialog → drawer 改造（drawer 模式首次试点）+ 视图层硬编码 hex 全部清零

## 1. 用户故事

作为管理员，希望：

1. **ScrapePreview / AVManualScrape**——「候选预览 + 确认」流程视觉清晰、卡片节奏统一；候选列表 / 候选详情 / 编辑确认三栏被 SectionCard 包裹后层次分明
2. **TvSeriesManage**——剧、季、集三层嵌套表单用 SectionCard 包裹后视觉层级清楚，不再是"裸 el-card + 自定义 .section-head"拼装
3. **ImageCollectionManage**——上传不再弹出居中大 dialog 把屏幕占满，改为右侧 drawer 一边看列表一边填上传字段；drawer 关闭时若已填入字段会弹「未保存确认」
4. **路由元数据 hideShellPageHeader**——4 视图也用 Phase 2 5181045f 引入的机制，让 Layout 顶栏不再重复显示视图标题
5. **整站冷蓝**——AVManualScrape / ScrapePreview 内残留的玫红 hex 全部清零

## 2. 作用域

| 改 | 不改 |
|----|------|
| `admin-web/src/views/ImageCollectionManage.vue`（上传 dialog → drawer + PageHeader + Toolbar + EmptyState） | 路由 URL（不重命名） |
| `admin-web/src/views/ScrapePreview.vue`（PageHeader + Toolbar + SectionCard×3 + EmptyState + hex 清 4 处） | API 接口 / 后端契约（包括刮削确认 / 弃刮 / bypass_cache 全部不动） |
| `admin-web/src/views/AVManualScrape.vue`（PageHeader + Toolbar + SectionCard×N + EmptyState + hex 清 1 处） | Pinia store 形状 |
| `admin-web/src/views/TvSeriesManage.vue`（PageHeader + Toolbar + SectionCard 三层嵌套） | 嵌套表单 IA（保留剧 / 季 / 集 inline 形态） |
| `admin-web/src/router/index.js`（4 路由 meta 加 hideShellPageHeader=true） | 三巨头域视图（VideoList / ImageManage / VideoUpload） |
| `admin-web/src/assets/themeTokens.spec.js`（视图层 audit 扩到 10 文件） | Phase 2 已 Done 的 7 视图 + UploadProgress.vue |
| `CONTEXT.md`（「admin 设计系统术语」段新增 2 条） | 「管理端手动刮削术语」/「AV 上传分类术语」既有段落（不改） |
| `tasks/2026-05-23-admin-medium-views/screenshots/` 9 张 PNG | `Layout.vue` / 7 个 base wrapper（Phase 1 已固化） |

## 3. 非目标

- N1：本期**不动**三巨头（`VideoList` / `ImageManage` / `VideoUpload`），留给 Phase 4
- N2：本期**不改** AVManualScrape / ScrapePreview 的 IA（候选列表 + 详情 + 确认三栏 inline 流程保持原貌）
- N3：本期**不改** TvSeriesManage 的 IA（剧 / 季 / 集嵌套 inline 表单保持原貌，不引入 drawer / 拖拽 / 排序新功能）
- N4：本期**不改** API 接口、Pinia store 形状、路由 URL
- N5：本期**不抽** 4 视图的共享业务逻辑——视觉重排不是重构
- N6：本期**不引入新** wrapper 组件
- N7：本期**不做** `#app :where(...)` 全局 specificity 重构（Phase 1 / Phase 2 遗留）
- N8：本期**不引入** Percy / Chromatic / Playwright 视觉差异 diff
- N9：本期**不接入** dark mode UI 切换
- N10：本期**不写** 视图渲染单测；靠源文 audit + 手测
- N11：本期**不改** ImageCollectionManage 的接口契约——drawer 替换 dialog 后调 `uploadImageCollection` / `createImageCollection` 等 API 不变；仅外壳形态从 dialog 改 drawer

## 4. Phase 3 关键决策结论（grill-with-docs 结果）

| # | 决策点 | 选项 |
|---|--------|------|
| Q1 | 历史欠债清理 | 拆 2 commits：补归档 Phase 1 任务文档（`acbac757`）+ 标记 Phase 2 DONE 含 Phase 3/4 skeleton 入 git（`71ba36c6`）——已落地 |
| Q2 | Phase 3 真实 IA 改造范围（PRD skeleton 与代码现状校准） | 4 视图全部「视觉重排 + SectionCard 分区」；**仅** ImageCollectionManage 是唯一 dialog → drawer 目标；TvSeriesManage 无 dialog 也不强改 drawer |
| Q3 | 改造顺序 + 提交策略 | **ImageCollectionManage 第一**（drawer 试点首次落地、给 Phase 4 三巨头铺路）→ ScrapePreview → AVManualScrape → TvSeriesManage（按 el-* 密度升序）；单 commit；如有 followup 走「实现 + followup」二次 commit |
| Q4 | 验收颗粒度 | H31–H34 单视图（外观锚点 + 功能不破二维）+ H35 drawer 通用约束 + H36 路由 meta + H37 全局自动化 |

## 5. 视图重排映射表（校准后）

| # | 视图 | 必接入 wrapper | IA 改造 | hex 清零（行：值） |
|---|------|---------------|---------|------------------|
| 1 | `ImageCollectionManage.vue` (685) | `PageHeader` + `Toolbar`（创建 + 搜索）+ `EmptyState` + 上传 dialog 改 `<el-drawer>` 560px + drawer 内 `SectionCard` ×3（批量选图 / 默认元数据 / 上传队列）+ drawer 底部 fixed `Toolbar`（取消 / 保存） | **唯一 IA 改造**：上传 dialog → drawer | 0 处 |
| 2 | `ScrapePreview.vue` (651) | `PageHeader` + `Toolbar`（查询）+ `SectionCard` ×3（查询表单 / 候选列表+详情 / 编辑确认）+ `EmptyState` | 不动 IA | `500: #7f1d1d` / `541: #be123c` / `546: #be123c` / `552: #881337` → `--primary` / `--primary-strong` / `--primary-soft` 派生 |
| 3 | `AVManualScrape.vue` (549) | `PageHeader` + `Toolbar`（站点切换 + 查询）+ `SectionCard` ×N（手动预览 / 候选列表 / 候选详情 / 确认表单 / 待确认列表）+ `EmptyState` | 不动 IA | `434: #7f1d1d` → `var(--primary)` |
| 4 | `TvSeriesManage.vue` (716) | `PageHeader` + `Toolbar`（创建剧 + 搜索）+ `SectionCard` 三层嵌套（剧基础 / 季列表 / 每季内含集列表）+ `EmptyState` | 不动 IA（保留 inline 嵌套表单） | 0 处 |

## 6. 改造顺序

1. **`ImageCollectionManage.vue`** —— drawer 首次试点；跑通后 Phase 4 三巨头复用零成本
2. `ScrapePreview.vue` —— 视觉 + hex 清 4 处；SectionCard ×3 包裹
3. `AVManualScrape.vue` —— 视觉 + hex 清 1 处；与 ScrapePreview 同构，复用模式
4. `TvSeriesManage.vue` —— 深嵌套 SectionCard；放最后让前 3 视图已验证 SectionCard 形态

## 7. CONTEXT.md 拟新增的 2 条术语

| 术语 | 一句话定义 |
|------|----------|
| `admin 编辑入口 Drawer` | admin 端 ImageCollectionManage（Phase 3 首次落地）与 Phase 4 三巨头的「编辑 / 上传」操作统一改右侧 drawer：宽 560px（`< 1024px` 全屏自适应），底部固定 Toolbar 含取消 / 保存；dirty 状态关闭弹「未保存确认」；drawer 内不嵌套 dialog；drawer 内复用 SectionCard 折叠分区。 |
| `admin 路由 hideShellPageHeader 标记` | Phase 2 `5181045f` 引入的路由 meta 标记：视图自身渲染 PageHeader 时，对应 route 必须显式 `meta.hideShellPageHeader = true`，Layout 顶栏不再渲染相同标题避免双 H1；Phase 2 已覆盖 7 视图，Phase 3 同步落 4 视图，Phase 4 三巨头一并按此约定。 |

## 8. 验收（H31–H37）

### H31 / ImageCollectionManage —— drawer 首次试点

**外观锚点**：
- ✓ 顶部 `PageHeader`「图片合集」，无副标题；Layout 顶栏不再显示同名标题（`route.meta.hideShellPageHeader=true` 生效）
- ✓ `Toolbar` 含「创建合集」按钮 + 搜索框
- ✓ 合集列表无数据时 `EmptyState`
- ✓ 点「上传图片」打开右侧 `<el-drawer>`：宽 560px（`< 1024px` 全屏自适应）
- ✓ drawer 内 `SectionCard` ×3 折叠区块：批量选图（默认展开）/ 默认元数据（默认展开）/ 上传队列（执行后展开）
- ✓ drawer 底部固定 `Toolbar`：左「取消」、右「保存」(primary)，随内容滚动始终可见

**功能不破**：
- ✓ 合集 CRUD + 视频绑定 + 取消绑定全可用
- ✓ drawer 内上传 / 进度反馈 / 上传完成回填 / 默认演员 / 默认合集字段功能与原 dialog 完全等价
- ✓ drawer dirty 关闭（点取消 / Esc / 点击遮罩）弹「未保存确认」，确认后才关
- ✓ drawer 内不嵌套 dialog（如演员选择 / 合集选择走 popover 或 inline list）

### H32 / ScrapePreview

**外观锚点**：
- ✓ `PageHeader`「通用刮削」+ `Toolbar`
- ✓ `SectionCard` ×3 包裹：查询表单 / 候选列表 + 候选详情 / 编辑确认
- ✓ 候选列表无数据时 `EmptyState`
- ✓ Scoped CSS 不含 `#7f1d1d` / `#be123c` / `#881337`

**功能不破**：
- ✓ 查询预览 / 选择候选 / 编辑 metadata / 提交确认 / `bypass_cache` 全可用
- ✓ 电影 / 剧集两种类型切换正确，季 / 集绑定正确

### H33 / AVManualScrape

**外观锚点**：
- ✓ `PageHeader`「AV 手动刮削」+ `Toolbar`（站点切换 + 查询）
- ✓ `SectionCard` ×N 包裹：手动预览 / 候选列表 / 候选详情 / 待确认列表
- ✓ `EmptyState` 在候选无数据时
- ✓ Scoped CSS 不含 `#7f1d1d`

**功能不破**：
- ✓ 多站点切换（JavDB / JavBus / JavLibrary / ThePornDB / MDCX）查询正确
- ✓ 候选选择 + 确认刮削 / 弃刮全可用
- ✓ `av_scrape_pending` 待确认列表加载正确

### H34 / TvSeriesManage

**外观锚点**：
- ✓ `PageHeader`「电视剧管理」+ `Toolbar`（创建剧 + 搜索）
- ✓ `SectionCard` 三层嵌套：剧基础信息（外层）/ 季列表（中层，每季一个 SectionCard）/ 集列表（内层，每集 inline form）
- ✓ 列表无剧时 `EmptyState`

**功能不破**：
- ✓ 剧 CRUD / 季 CRUD / 集 CRUD 全可用
- ✓ 嵌套表单字段保存与读取正确

### H35 / admin 编辑入口 Drawer dirty 关闭契约（跨视图通用）

- ✓ ImageCollectionManage 上传 drawer 在任一字段已填后点 Esc / 关闭按钮 / 点击遮罩 → 弹「未保存确认」 → 「确认丢弃」才关
- ✓ drawer 未 dirty 时直接关闭无确认提示
- ✓ drawer 内字段保存后 dirty 复位
- ✓ drawer 关闭后 dirty 标志彻底复位（下次打开是干净状态）

### H36 / admin 路由 hideShellPageHeader 标记

- ✓ `router/index.js` 内 `/image-collections` / `/scrape` / `/av-scrape` / `/tv-series` 4 个路由 meta 都含 `hideShellPageHeader: true`
- ✓ 进入这 4 个路由时 Layout 顶栏不显示标题，视图体内 PageHeader 是唯一标题源
- ✓ 与 Phase 2 已落的 7 视图保持一致行为

### H37 / 全局自动化

- ✓ `cd admin-web && npm run build` 通过
- ✓ `cd admin-web && npm test` 全绿（75 + 扩展 audit ≥ 77 用例）
- ✓ `themeTokens.spec.js` 视图层 audit 扩到 10 文件（7 Phase 2 + UploadProgress + ScrapePreview + AVManualScrape），全部零玫红 hex
- ✓ `CONTEXT.md`「admin 设计系统术语」段新增 2 条
- ✓ `plan.md` 顶部追加 Phase 3 实施条目

## 9. 截图归档

`tasks/2026-05-23-admin-medium-views/screenshots/` 共 9 张：

| 文件 | 说明 |
|------|------|
| `before-image-collection.png` / `after-image-collection.png` | ImageCollectionManage 列表态 |
| `after-image-collection-drawer.png` | 上传 drawer 打开形态（首次试点专属） |
| `before-scrape-preview.png` / `after-scrape-preview.png` | ScrapePreview 三栏 |
| `before-av-manual-scrape.png` / `after-av-manual-scrape.png` | AVManualScrape |
| `before-tv-series-manage.png` / `after-tv-series-manage.png` | TvSeriesManage 嵌套 |

统一 1440px 桌面分辨率；不做像素 diff，目视判断「视觉切到新设计系统 + drawer 改造可见」。

## 10. Done definition

- [ ] H31–H37 全部通过
- [ ] 9 张截图归档完成
- [ ] `CONTEXT.md` 「admin 设计系统术语」新增 2 条术语
- [ ] `plan.md` 顶部追加 Phase 3 实施条目
- [ ] 全部改动落在 1 commit（中文 subject + body）；如出现 code review followup 则同 Phase 1 / Phase 2 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-medium-views/DONE.md` 含完成日期 + commit hash + 验证摘要
