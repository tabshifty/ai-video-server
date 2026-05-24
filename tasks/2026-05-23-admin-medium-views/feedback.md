# Feedback：admin 重设计 Phase 3 · code review 反馈

- 日期：2026-05-24
- 关联 commit：`9fd68a5d 重排 admin 中等视图并归档截图`
- 来源：grill-with-docs / code-review skill（high effort，3 angles × 6 candidates，1-vote verify）
- 落地 followup commit：见本任务下一个「修复」commit
- 性质：归档反馈 + 已落地清单

## 1. 优先级与落地状态总览

| 等级 | 数量 | 摘要 | 状态 |
|------|------|------|------|
| **必修 P0** | 3 | AVManualScrape source-summary 删除 / TvSeriesManage el-collapse 拆为永远展开 + 死代码 / ImageCollectionManage drawer dirty 双弹 + 假弹 | ✅ 全部落地 |
| **必修 P1** | 3 | TvSeriesManage el-scrollbar 删除 / ImageCollectionManage v-if + v-loading 空白态 / ScrapePreview Enter 不触发查询（顺带 AVManualScrape 同款补齐） | ✅ 全部落地 |
| **强烈建议** | 3 | audit 范围与 PRD 不齐 / SectionCard v-loading directive 不工作 / 三层嵌套视觉违和 | ✅ 全部落地 |
| **可选** | 1 | ImageCollectionManage cover_image_id snapshot 漏字段 | ✅ 顺手落地 |

## 2. 详细反馈与落地

### P0-1 / AVManualScrape source-summary 删除（产品功能回归）

- **位置**：`admin-web/src/views/AVManualScrape.vue:274` 附近
- **症状**：源站点推荐 / 实际使用 / 当前分类三张 el-tag 反馈块整段被删，但 `recommendedSource` / `usedSource` 两个 ref + doPreview 内对它们的赋值仍保留——产品功能被静默删除并留下死代码
- **失败路径**：运维输入欧美 AV 标题查询时无法看到「自动推荐 vs 实际命中」站点对比，无法定位策略漂移；这是 CONTEXT.md「欧美 AV / 刮削确认门控」反馈链上的关键 UI 回归
- **修复**：在 Toolbar 之后插入 `<div v-if="recommendedSource || usedSource" class="source-summary">` 三张 el-tag 反馈块；scoped CSS 新增 `.source-summary` 内嵌（flex + 小卡背景），形成「查询后可见的简洁状态条」

### P0-2 / TvSeriesManage el-collapse 拆为永远展开 SectionCard + 死代码

- **位置**：`admin-web/src/views/TvSeriesManage.vue:42, 168, 178, 217`（activeSeasons 死代码）+ `:388-510` 区域（嵌套渲染）
- **症状**：原 `<el-collapse v-model="activeSeasons">` 改成纯 SectionCard 堆叠后所有季和分集**永远展开**；20 季 × 20 集的长寿剧首屏挂载 400+ 表单项；`activeSeasons` ref 与 4 处赋值成为僵尸死代码
- **失败路径**：选中一部多季剧 → 一次性渲染所有季 + 每季所有分集（含 el-input / el-select / el-form-item）→ 首屏严重卡顿 + 用户失去「折叠已编辑季 / 聚焦当前季」交互能力
- **修复**：
  - 季 SectionCard 改用 `collapsible :default-expanded="seasonIndex === 0" dense`，第一季默认展开，其余默认折叠
  - 集 SectionCard 改用 `collapsible :default-expanded="false" dense`，全部默认折叠
  - 删除 `activeSeasons` ref 与所有 3 处赋值（line 42 / 168 / 178 / 217）

### P0-3 / ImageCollectionManage drawer dirty 双弹 + 假弹（drawer 试点核心契约）

- **位置**：`admin-web/src/views/ImageCollectionManage.vue:184` requestEditDrawerClose + `:210` save + `:190` handleEditDrawerBeforeClose
- **症状**：(a) 点取消 → `requestEditDrawerClose` 弹 ElMessageBox → 用户「确认丢弃」→ `editDrawerVisible.value = false` → el-drawer v-model 翻转触发 `before-close` → `handleEditDrawerBeforeClose` 看到 dirty 仍 true → **再弹一次**「未保存的修改将会丢失」。(b) save 成功后直接 `editDrawerVisible.value = false`，snapshot 仍是旧表单 form 已被新值填，dirty=true → before-close **假弹**「未保存」即便刚成功保存
- **修复**：cancel 与 save 路径都在 set visible=false **之前**显式调 `captureFormSnapshot()` 复位 snapshot，让 v-model flip 时 dirty=false，before-close 守卫直接走 `done()` 分支不弹

### P1-1 / TvSeriesManage el-scrollbar 删除（独立滚动丢失）

- **位置**：`admin-web/src/views/TvSeriesManage.vue:347, 351, 376`
- **症状**：原侧栏系列列表用 `<el-scrollbar height="calc(100vh - 320px)">` 独立可滚动容器；改造后改成 `.series-list-shell { min-height: 20rem }` 平铺无 max-height 也无 overflow
- **失败路径**：系列数 30+ 时左侧列表撑长整页，与右侧编辑器之间不再各自滚动；触屏端尤其明显
- **修复**：恢复 `<el-scrollbar v-loading="loadingList" class="series-list-shell">` 包裹；scoped CSS 给 `.series-list-shell` 加 `max-height: calc(100vh - 24rem)`

### P1-2 / ImageCollectionManage v-if + v-loading 空白态

- **位置**：`admin-web/src/views/ImageCollectionManage.vue:466`
- **症状**：`<el-table v-if="list.length" v-loading="loading">` 配合 `<EmptyState v-else-if="!loading">` 在 `list=[] && loading=true` 时两个分支都不渲染，整个面板一片空白
- **失败路径**：首次进入页面 / 切换筛选条件清空 list 再请求时，loading 阶段表格被 `v-if` 摘掉、EmptyState 又因 `!loading=false` 不渲染 → 用户以为页面卡死
- **修复**：把 `v-loading="loading"` 挂在外层 `<SectionCard>` 上（依赖 SectionCard 已加 `position: relative` 让 v-loading 遮罩正确定位）；`<el-table v-if="list.length">` 与 `<EmptyState v-else-if="!loading">` 不变；loading 期间整个 SectionCard 显示遮罩 spinner

### P1-3 / ScrapePreview + AVManualScrape Enter 不触发查询（一致性回归）

- **位置**：`admin-web/src/views/ScrapePreview.vue:298, 301` + `AVManualScrape.vue:255, 258`
- **症状**：旧版查询按钮在 `<el-form inline>` 内由 form submit 默认接管 Enter，重构后按钮被挪到 Toolbar `#actions` 与 form 分离，且未补 `@keyup.enter`；而 ImageCollectionManage 同期重排却保留了 `@keyup.enter="load"`，跨视图键盘体验不一致
- **修复**：给「视频ID」/「标题」两个 el-input 都加 `@keyup.enter="doPreview"`（ScrapePreview + AVManualScrape 各 2 处），与 ImageCollectionManage 形成键盘体验闭环

### P2-1 / themeTokens audit 范围与 PRD 不齐

- **位置**：`admin-web/src/assets/themeTokens.spec.js:5-16`
- **症状**：PRD 第 1.1 节明确「10 文件 audit」实际只追加了 ScrapePreview + AVManualScrape；TvSeriesManage / ImageCollectionManage 本期改造但未纳入 audit，spec 与改造范围不齐
- **修复**：VIEW_HEX_AUDIT_TARGETS 扩展到 12 项（8 + Phase 3 4 项全部），audit 关闭范围漏洞

### P2-2 / SectionCard v-loading directive 不工作（compound 组件根没 position relative）

- **位置**：`admin-web/src/components/base/SectionCard.vue`
- **症状**：`<SectionCard v-loading="...">` 把 v-loading 挂在 Vue compound 组件上；SectionCard 根 `<section>` 没 `position: relative`，Element Plus v-loading 遮罩落到祖先 positioned 节点（可能是 Layout shell）
- **修复**：`.section-card` 添加 `position: relative`，让 v-loading mask 正确绑定到本身

### P2-3 / TvSeriesManage 三层嵌套 SectionCard 视觉违和

- **位置**：`admin-web/src/views/TvSeriesManage.vue:392, 446, 479`
- **症状**：剧外层 SectionCard → 季中层 SectionCard → 集内层 SectionCard 三层都默认带 border + padding + shadow，累积视觉边框 + 缩进与 Phase 1 Modern Minimal「elevation 表达层次」克制原则相悖；1366px 笔记本明显挤压
- **修复**：SectionCard 新增 `dense` prop——`is-dense` class 减小 gap / padding，去掉 shadow，title 字号收为 small 灰；季 + 集嵌套 SectionCard 都传 `dense`；外层剧 SectionCard 维持默认形态

### P3-1 / ImageCollectionManage cover_image_id snapshot 漏字段

- **位置**：`admin-web/src/views/ImageCollectionManage.vue:122` captureFormSnapshot + `:52` editDrawerDirty
- **症状**：snapshot 不含 `cover_image_id`，dirty 比对也不含，可能导致 cover 被旧值 clobber 或 dirty 守卫误判
- **修复**：captureFormSnapshot 补 `cover_image_id` 字段（用 `?? null` 兼容 undefined）；editDrawerDirty computed 增加 `(form.cover_image_id ?? null) !== (editDrawerSnapshot.value.cover_image_id ?? null)` 比对

## 3. 落地后验证

- `cd admin-web && npm test`：79 / 79 全绿（含新增 4 个视图层 audit，共 12 audit 用例）
- `cd admin-web && npm run build`：通过（dist/assets/index-*.css 401.96 kB / dist/assets/index-*.js 2.31 MB）
- `git diff --check`：无空白污染
- 乱码扫描：无残留

## 4. 已知遗留（不进入本 followup）

- 无；本期 P0-P3 + 可选项全部落地
- Phase 2 遗留 P3-1（TaskMonitor 请求竞态）仍未处理，留单独 polish
- Phase 1 / Phase 2 遗留 `#app :where(...)` 全局 specificity 收口仍未处理

## 5. 后续建议

- Phase 4 三巨头 drawer 改造可直接复用 ImageCollectionManage 验证过的 pattern：
  - cancel + save 路径都先 `captureFormSnapshot()` 再 set visible=false
  - 外层容器挂 `v-loading` 而非内层 `el-table`
  - dirty 比对要把所有 form 字段（含 id 类引用字段）纳入 snapshot
- SectionCard 的 `dense` prop 可在 Phase 4 嵌套场景复用
- 若未来发现 SectionCard 在更多深嵌套场景表现不佳，考虑追加 `flat` 形态（无 border / 无 padding / 无 shadow，纯 grouping 容器）
