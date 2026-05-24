# Phase 3 完成记录 · admin 中等视图重排

- 完成日期：2026-05-24
- 关联提交：
  - `9fd68a5d` 重排 admin 中等视图并归档截图（首版实现：4 视图 + drawer 试点 + router meta + audit 扩展）
  - `3150da71` 修复 admin 中等视图重排 code review 反馈（落地 feedback.md 的 P0–P3 共 10 项）

## 验证摘要

### 自动化必跑
- `cd admin-web && npm test` —— 全绿（12 文件 / 79 用例，含 12 视图层 audit）
- `cd admin-web && npm run build` —— 通过（vite 6.4.2，2280 modules，dist 402KB CSS / 2.31MB JS）

### 4 视图 H31–H34 手测覆盖
- H31 ImageCollectionManage：PageHeader + Toolbar + EmptyState + 上传 dialog → drawer（560px / 全屏自适应 / SectionCard×3 / 底部固定 Toolbar / dirty 守卫）；合集 CRUD + 视频绑定 + 取消绑定全可用
- H32 ScrapePreview：PageHeader + Toolbar + SectionCard×3（查询 / 候选列表+详情 / 编辑确认）+ EmptyState；查询预览 + 候选选择 + 编辑 metadata + 提交确认 + bypass_cache 全可用；scoped CSS 无玫红 hex
- H33 AVManualScrape：PageHeader + Toolbar（站点切换 + 查询）+ SectionCard×N + EmptyState；source-summary 三 tag 反馈条恢复；多站点切换 + 候选选择 + 确认刮削 / 弃刮全可用
- H34 TvSeriesManage：PageHeader + Toolbar + el-scrollbar 包裹系列列表 + SectionCard 三层嵌套（剧外层 + 季/集 dense collapsible）；剧/季/集 CRUD 全可用

### H35 / drawer dirty 通用契约（ImageCollectionManage 试点）
- drawer dirty 时 cancel / Esc / 遮罩点击都弹「未保存」确认
- 未 dirty 时直接关闭无确认
- save 成功后 captureFormSnapshot 复位，关闭不再假弹
- cancel 确认丢弃后 captureFormSnapshot 复位，避免 before-close 二次弹同款确认
- drawer 关闭后下次打开是干净状态

### H36 / 路由 hideShellPageHeader 标记
- `/image-collections` / `/scrape` / `/av-scrape` / `/tv-series` 4 个 route meta 均含 `hideShellPageHeader: true`
- Layout shell 顶栏不再渲染同名标题，视图体内 PageHeader 是唯一标题源

### H37 / 全局自动化
- npm test / npm build 全绿
- `themeTokens.spec.js` 视图层 audit 扩到 12 文件（8 Phase 2 + 4 Phase 3）全部命中
- `CONTEXT.md`「admin 设计系统术语」段新增 2 条术语
- `plan.md` 顶部追加 Phase 3 实施 + followup 两条条目

### 截图归档（`screenshots/`）
- 9 张 PNG 已归档：
  - `before/after-image-collection.png`、`after-image-collection-drawer.png`（drawer 试点专属）
  - `before/after-scrape-preview.png`
  - `before/after-av-manual-scrape.png`
  - `before/after-tv-series-manage.png`

## CONTEXT.md 沉淀
- 「admin 设计系统术语」段新增 2 条：
  - `admin 编辑入口 Drawer`（ImageCollectionManage 试点 + Phase 4 三巨头复用契约）
  - `admin 路由 hideShellPageHeader 标记`（视图自渲染 PageHeader 时必须配套 meta 防双 H1）

## Feedback 落地清单（`feedback.md` 共 10 项）

| 等级 | 反馈 | 状态 |
|------|------|------|
| P0-1 | AVManualScrape source-summary 删除（产品功能回归） | ✅ 三 el-tag 反馈条恢复 + 死代码语义化 |
| P0-2 | TvSeriesManage el-collapse → 永远展开 SectionCard + 死代码 | ✅ 季/集 SectionCard collapsible + 第一季默认展开 + activeSeasons 死代码删除 |
| P0-3 | ImageCollectionManage drawer dirty 双弹 + 假弹 | ✅ cancel + save 路径都先 captureFormSnapshot 再 set visible=false |
| P1-1 | TvSeriesManage el-scrollbar 删除 | ✅ 恢复 `<el-scrollbar>` + max-height 实现独立滚动 |
| P1-2 | ImageCollectionManage v-if+v-loading 空白态 | ✅ v-loading 挂到外层 SectionCard，依赖 SectionCard 加 position:relative |
| P1-3 | ScrapePreview Enter 键失效（顺带 AVManualScrape） | ✅ 4 处 `@keyup.enter="doPreview"` |
| P2-1 | themeTokens audit 范围与 PRD 不齐 | ✅ VIEW_HEX_AUDIT_TARGETS 扩到 12 文件 |
| P2-2 | SectionCard v-loading directive 不工作 | ✅ `.section-card` 加 `position: relative` |
| P2-3 | 三层嵌套 SectionCard 视觉违和 | ✅ 新增 dense prop（is-dense class 减重）+ defaultExpanded prop |
| P3-1 | ImageCollectionManage cover_image_id snapshot 漏字段 | ✅ captureFormSnapshot + dirty 比对补字段 |

## plan.md 追加条目
- 2026-05-24 Phase 3 实施 + 截图归档（`9fd68a5d`）
- 2026-05-24 Phase 3 反馈修复（`3150da71`）

## 已知遗留（移交 Phase 4 或单独 polish）

- **Phase 2 P3-1** TaskMonitor 状态切换与自动刷新计时器并发：留作单独 polish
- **Phase 1-3 遗留** `#app :where(...)` 全局 specificity 收口：仍未处理；Phase 4 若被压制再就地解决
- **Phase 4 drawer 复用契约**：ImageCollectionManage drawer 验证过的 pattern（cancel + save 先 captureFormSnapshot、v-loading 挂外层、dirty 守卫含 id 类字段、dense 嵌套 SectionCard）直接给 VideoList / ImageManage / VideoUpload 三巨头复用
