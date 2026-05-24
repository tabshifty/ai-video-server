# PRD：admin 重设计 Phase 4 · 三巨头 IA 改造

- 日期：2026-05-24
- 阶段：Phase 4 / 4（三巨头 IA 改造 + 视觉收尾）
- 前置任务：
  - `tasks/2026-05-23-admin-shell-redesign/DONE.md`（Phase 1 已 Done，commits `12bf7c93` + `5ae43fa5` + `acbac757`）
  - `tasks/2026-05-23-admin-simple-views/DONE.md`（Phase 2 已 Done，commits `359a6489` + `5181045f` + `71ba36c6`）
  - `tasks/2026-05-23-admin-medium-views/DONE.md`（Phase 3 已 Done，commits `9fd68a5d` + `3150da71` + `ee3f6712`）
- 范围：VideoList / ImageManage / VideoUpload 三巨头按 Phase 1-3 设计系统重排 + IA 改造（drawer / 列设置 / chip 筛选 / BulkActionBar / 3 步 Wizard）+ ImageManage hex 清零

## 1. 用户故事

作为管理员，希望：

1. **VideoList**——筛选条不再常驻挤占屏幕（折叠成 chip 行 + 「更多筛选」抽屉）；表格列可自选并跨会话持久化；编辑视频改为右侧抽屉，drawer 内含 SectionCard 折叠区块（基础 / 季集 / 演员标签 / 图片合集 / 字幕 / 播放预览）一边看一边改；多选时顶部浮起 BulkActionBar 批量操作
2. **ImageManage**——默认网格视图（图片产品的直觉视图）、可切回列表视图（偏好跨会话持久化）；网格卡片 hover 显示快捷操作；批量选中浮起 BulkActionBar；上传与编辑都从大 dialog 改右侧 drawer；筛选条折叠为 chip 行
3. **VideoUpload**——单页大表拆 3 步 Wizard（选文件 / 基础信息 / 关联与上传）；上一步保留已填字段；上传执行后保留进度与历史不被 step 切换覆盖；类型选择驱动条件字段（AV 地区分类 / 短视频所属合集）
4. **路由元数据 hideShellPageHeader**——3 视图也用 Phase 2-3 引入的机制，让 Layout 顶栏不再重复显示视图标题
5. **ImageManage 视图层 hex 清零**——`#881337` / `#7f1d1d` ×2 共 3 处玫红 hex 全部换 token

## 2. 作用域

| 改 | 不改 |
|----|------|
| `admin-web/src/views/VideoList.vue`（PageHeader + Toolbar + chip 筛选 + 列设置 + drawer + BulkActionBar + EmptyState） | 路由 URL |
| `admin-web/src/views/ImageManage.vue`（PageHeader + Toolbar + chip 筛选 + 视图切换 + 编辑/上传 drawer ×2 + 网格卡片 + BulkActionBar + EmptyState + hex 清 3 处） | Pinia store 形状 |
| `admin-web/src/views/VideoUpload.vue` 改为 step 容器 + 共享 form state | API 接口 / 后端契约 |
| 新增 `admin-web/src/views/VideoUpload/StepFile.vue` / `StepBasic.vue` / `StepRelate.vue` 三个 step 子组件 | 三巨头编辑 drawer 内业务字段（episode / AV / short / movie 各自的 conditional 字段保持原 schema） |
| 新增 `admin-web/src/views/videoUpload.wizard.helpers.js` + `.spec.js`（step 切换 + 字段保留 + 类型条件字段纯函数） | 表格列结构 / 列宽（仅在「列设置」交互层添加显示/隐藏，不删列、不改列宽） |
| `admin-web/src/router/index.js`（3 路由 meta 加 hideShellPageHeader=true） | 表单字段顺序（除了 Wizard 拆 step 引起的逻辑分组）|
| `admin-web/src/assets/themeTokens.spec.js`（视图层 audit 扩到 15 文件，含 3 个 step 子文件） | `Layout.vue` / 7 个 base wrapper（Phase 1-3 已固化） |
| `CONTEXT.md`（「admin 设计系统术语」段新增 5 条） | Phase 1-3 已 Done 的 11 视图 + UploadProgress.vue |
| `tasks/2026-05-23-admin-three-pillars/screenshots/` 11 张 PNG | `BulkActionBar.vue` 接口（Phase 1 已固化的 wrapper API）|

## 3. 非目标

- N1：本期**不改** API 接口、Pinia store 形状、路由 URL（VideoUpload 仍是单一 `/upload` 路由，3 步内部切换不引入子路由）
- N2：本期**不引入** vue-grid-layout / interact.js / 拖拽布局（既不做表格列拖拽排序，也不做网格自定义拖拽布局）
- N3：本期**不接入**「VideoUpload 模板保存」（重复上传场景的快捷模板）
- N4：本期**不接入**「图片标签 AI 自动识别」
- N5：本期**不做** VideoList 分屏视图（左列右详情）——drawer 已够用
- N6：本期**不改** VideoList 编辑 drawer 内的业务字段（episode / AV / short / movie 各自的 conditional 字段保持原 schema）
- N7：本期**不引入新** wrapper 组件——坚持 Phase 1 七个
- N8：本期**不做** `#app :where(...)` 全局 specificity 重构（Phase 1-3 遗留）
- N9：本期**不引入** Percy / Chromatic / Playwright 视觉差异 diff
- N10：本期**不接入** dark mode UI 切换
- N11：本期**不写** 视图渲染单测；helpers spec + 源文 audit + 手测三层
- N12：本期**不持久化** BulkActionBar 选中状态（关闭 tab 后状态归零）；批量操作是"当前会话"语义

## 4. Phase 4 关键决策结论（grill-with-docs 结果）

| # | 决策点 | 选项 |
|---|--------|------|
| Q1 | Phase 4 范围与 IA 改造目标确认 | 3 视图 IA 改造 + ImageManage hex 清 3 处 + 复用 Phase 3 ImageCollectionManage drawer 试点 5 项契约（cancel/save 先 captureFormSnapshot / v-loading 挂外层 / dirty 比对含 id 字段 / SectionCard dense 嵌套 / drawer 宽 560px + 全屏自适应）|
| Q2 | 改造顺序 + 提交策略 | **ImageManage 第一**（覆盖最多新模式 BulkActionBar + chip + 视图切换 + 2 drawer + hex 清零）→ VideoList 第二（复用 + 增量列设置）→ VideoUpload 第三（Wizard 拆分独立）；单 commit；followup 模式与 Phase 1-3 对齐 |
| Q3 | 验收颗粒度 | H41–H43 单视图（外观锚点 + 功能不破二维）+ H44 drawer dirty 通用 + H45 BulkActionBar 首次试点 + H46 localStorage 持久化 + H47 路由 meta + H48 全局自动化 |

## 5. 视图重排映射表

| # | 视图 | 必接入 wrapper | IA 改造 | hex 清零 |
|---|------|---------------|---------|---------|
| 1 | `ImageManage.vue` (1045) | PageHeader + Toolbar（chip + 视图切换）+ EmptyState + 编辑 drawer + 上传 drawer + BulkActionBar + 网格卡片 | 4 项新模式：BulkActionBar 首次接入 + chip 筛选 + 视图切换（网格/列表 + localStorage）+ 2 个 dialog → drawer | `919: #881337` → `--primary-strong` / `937 + 986: #7f1d1d` → `--primary` |
| 2 | `VideoList.vue` (1422) | PageHeader + Toolbar（chip + 列设置）+ EmptyState + 编辑 drawer + BulkActionBar | 1 项新模式：**列设置（localStorage）** + 复用 chip / drawer / BulkActionBar | 0 处 |
| 3 | `VideoUpload.vue` (719) → 拆 step 子组件 | PageHeader + 顶部 el-steps + 底部 Toolbar（上一步 / 下一步 / 提交） | **3 步 Wizard 拆分**：StepFile / StepBasic / StepRelate + 共享 form state；上传进度与历史在 step 3 底部不被切换覆盖 | 0 处 |

## 6. 改造顺序

1. **`ImageManage.vue`** —— 最广模式首次落地（BulkActionBar / chip / 视图切换 / 双 drawer / hex 清零 5 项）
2. **`VideoList.vue`** —— 复用 ImageManage 验证过的 pattern + 增量列设置（localStorage 持久化）
3. **`VideoUpload.vue`** —— Wizard 3 步拆分 + step 子组件 + wizard helpers 单测

## 7. CONTEXT.md 拟新增 5 条术语（一并入本期 commit）

| 术语 | 一句话定义 |
|------|----------|
| `admin 表格列设置` | VideoList 顶部「列设置」按钮，管理员勾选显示/隐藏列；偏好写入 `localStorage` key `admin-videolist-columns`；窗口 < 1280px 自动隐藏次要列但仍尊重用户既有显式偏好；不删列、不改列宽 |
| `admin 视图模式切换` | ImageManage 顶部「视图切换」开关（网格/列表）；默认网格；偏好写入 `localStorage` key `admin-imagemanage-view`；跨会话持久化；切换不破坏分页与筛选状态 |
| `admin 上传向导三步` | VideoUpload 拆 3 步：选文件 / 基础信息 / 关联与上传；上一步保留字段；第 3 步上传进度与历史不被 step 切换覆盖；类型选择驱动条件字段（AV 地区分类 / 短视频所属合集） |
| `admin 批量操作浮条` | 表格/网格选中 ≥ 1 项时顶部浮起 BulkActionBar；含选中数量 + 主操作（批量删除 / 批量改状态 / 批量打标）；取消选择即消失；不跨会话持久化选中 |
| `admin 筛选条折叠` | 表格/列表视图筛选条默认折叠为「已生效筛选 chip 行」+「更多筛选」按钮抽屉；常用筛选始终可见，长尾筛选按需展开；与 VideoList / ImageManage 配合实现「头部干净 + 高级筛选可达」 |

## 8. 验收（H41–H48）

### H41 / ImageManage —— 最广模式覆盖（4 项新模式首次落地）

**外观锚点**：
- ✓ 顶部 `PageHeader`「图片管理」，Layout 顶栏不再显示同名标题
- ✓ `Toolbar` 含「上传图片」+ chip 筛选（状态 / 演员 / 图片合集）+ 视图切换开关（网格/列表）
- ✓ 默认进入网格视图（除非用户既有 localStorage 偏好为列表）
- ✓ 切到列表视图 → 刷新页面后保持列表视图
- ✓ 列表无数据时 `EmptyState`
- ✓ 网格卡片 hover 显示快捷操作（编辑 / 删除 / 切换启用），左上角 selection checkbox
- ✓ 编辑入口 + 上传入口都从 dialog 改 drawer（560px / 全屏自适应 / 底部固定 Toolbar / dirty 守卫）
- ✓ Scoped CSS 不含 `#881337` / `#7f1d1d`

**功能不破**：
- ✓ 图片 CRUD + 启用切换 + 演员/合集绑定全可用
- ✓ 上传 drawer 内多文件批量选图 + 进度反馈 + 上传完成回填表格
- ✓ 网格/列表视图切换不破坏分页与筛选状态

### H42 / VideoList —— 最复杂表格 + 列设置首次落地

**外观锚点**：
- ✓ 顶部 `PageHeader`「视频管理」，Layout 顶栏不再显示同名标题
- ✓ `Toolbar` 含 chip 筛选（状态 / 类型 / 等）+「更多筛选」抽屉触发 + 列设置按钮
- ✓ 表格默认仅显示用户配置或系统默认列
- ✓ 窗口 < 1280px 自动隐藏次要列但仍尊重用户既有显式偏好
- ✓ 行操作 hover 显示快捷操作 + 「⋯」更多菜单
- ✓ 表格无数据时 `EmptyState`
- ✓ 编辑入口 dialog → drawer 560px，drawer 内 SectionCard 分区（基础信息 / 季集 / 演员标签 / 图片合集 / 字幕 / 播放预览）

**功能不破**：
- ✓ 表格分页 / 排序 / 搜索 / 筛选全可用
- ✓ 编辑 drawer 内的 episode / AV / short / movie 各类型 conditional 字段保持原 schema
- ✓ 播放预览仍可用
- ✓ 字幕管理仍可用

### H43 / VideoUpload —— 3 步 Wizard 拆分

**外观锚点**：
- ✓ 顶部 `PageHeader`「上传中心」+ `<el-steps>` 3 步（选文件 / 基础信息 / 关联与上传）
- ✓ Step 1 显示文件拖拽 + 多选 + 已选清单（缩略 + 大小 + 移除）
- ✓ Step 2 显示类型选择 → 类型 = AV 时显示「AV 地区分类」+ 标题 + 描述
- ✓ Step 3 显示标签 / 演员 / 图片合集 / 所属合集（仅 short）+「开始上传」按钮 + 进度条 + 历史结果表
- ✓ 底部固定 `Toolbar` 含下一步 / 上一步 / 提交按钮

**功能不破**：
- ✓ Step 1 必填验证（必选 ≥ 1 文件）
- ✓ Step 2 类型 = 视频时跳过 AV 字段
- ✓ Step 1 ⇄ Step 2 ⇄ Step 3 切换字段保留
- ✓ Step 3 开始上传 → 进度反馈 → 历史结果 → 切回 step 1/2 再回 step 3 进度与历史不丢失
- ✓ 取消上传 / 清空已选 / 清空记录 三按钮在 step 3 底部

### H44 / admin 编辑入口 Drawer dirty 通用契约（复用 Phase 3 试点）

- ✓ VideoList 编辑 drawer + ImageManage 编辑 / 上传 drawer 在 dirty 状态关闭都弹「未保存」
- ✓ save 成功后立即关闭不弹假确认（snapshot 已对齐到 form）
- ✓ cancel 确认丢弃后不二次弹同款确认（snapshot 已复位）

### H45 / admin 批量操作浮条 BulkActionBar 首次试点

- ✓ ImageManage / VideoList 表格/网格选中 ≥ 1 项时 BulkActionBar 浮起
- ✓ 顶部对齐居中，含「已选 N 项」+ 主操作（批量删除 / 批量改状态 / 批量打标）
- ✓ 取消选择即消失
- ✓ 关闭浏览器 tab 不持久化选择状态

### H46 / localStorage 持久化 —— 两条 key 验证

- ✓ VideoList 列设置写入 `localStorage` key `admin-videolist-columns`，刷新页面后保持
- ✓ ImageManage 视图模式写入 `localStorage` key `admin-imagemanage-view`，刷新页面后保持
- ✓ 隐私模式 / 配额溢出时 `localStorage` 读写失败不破坏视图（兜底默认值）

### H47 / admin 路由 hideShellPageHeader 标记 —— 3 视图同步

- ✓ `router/index.js` 内 `/videos` / `/images` / `/upload` 3 个路由 meta 都含 `hideShellPageHeader: true`
- ✓ Layout shell 顶栏不显示同名标题，视图体内 PageHeader 是唯一标题源
- ✓ 与 Phase 2-3 已落的 11 视图保持一致行为

### H48 / 全局自动化

- ✓ `cd admin-web && npm run build` 通过
- ✓ `cd admin-web && npm test` 全绿（79 + 新增 ≥ 3 = ≥ 82 用例）
- ✓ `themeTokens.spec.js` 视图层 audit 扩到 15 文件（12 + 3 巨头），全部零玫红 hex
- ✓ `videoUpload.wizard.spec.js` step 切换 + 字段保留 + 类型条件字段单测全绿
- ✓ `CONTEXT.md`「admin 设计系统术语」段新增 5 条
- ✓ `plan.md` 顶部追加 Phase 4 实施条目

## 9. 截图归档（11 张）

`tasks/2026-05-23-admin-three-pillars/screenshots/`：

| 文件 | 说明 |
|------|------|
| `before/after-image-manage.png`、`after-image-manage-grid.png`、`after-image-manage-drawer.png`、`after-image-manage-bulk.png` | ImageManage 形态全套（4 张 after 涵盖列表/网格/drawer/批量选） |
| `before/after-video-list.png`、`after-video-list-drawer.png`、`after-video-list-columns.png` | VideoList（列表 / drawer / 列设置 popover） |
| `before/after-video-upload.png` | VideoUpload Wizard |

11 张统一 1440px 桌面分辨率；目视判断「IA 改造可见」。

## 10. Done definition

- [ ] H41–H48 全部通过
- [ ] 11 张截图归档完成
- [ ] `CONTEXT.md`「admin 设计系统术语」新增 5 条术语（Phase 1 6 条 + Phase 3 2 条 + Phase 4 5 条 = 13 条收口）
- [ ] `plan.md` 顶部追加 Phase 4 实施条目
- [ ] 全部改动落在 1 commit（中文 subject + body）；如有 code review followup 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-three-pillars/DONE.md` 含完成日期 + commit hash + 验证摘要 + Phase 4 落地后 admin 重设计全 4 阶段收尾标记
