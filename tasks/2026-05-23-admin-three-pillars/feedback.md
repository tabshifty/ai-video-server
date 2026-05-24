# Feedback：admin 重设计 Phase 4 · code review 反馈

- 日期：2026-05-24
- 关联 commit：`6fb14062 完成 admin 三巨头 IA 改造`
- 来源：grill-with-docs / code-review skill（high effort，3 angles × 6 candidates，1-vote verify）
- 性质：归档反馈；本次会话用户已要求把反馈记入此文件，落地动作另行决定

## 1. 优先级分布

| 等级 | 数量 | 摘要 |
|------|------|------|
| **必修 P0**（数据可达性 + 数据丢失） | 3 | VideoList 操作列丢失风险 / ImageManage 批量操作部分失败 + UI 不刷新 / VideoUpload type 切换静默丢文件 |
| **必修 P1**（状态一致性 + 资源泄漏） | 2 | ImageManage 视图切换选中泄漏 → 误删 / VideoList drawer 缺 destroy-on-close → blob 泄漏 + 异步竞态 |
| **强烈建议**（UX 一致性 + dirty 守卫） | 4 | BulkActionBar 失 loading 反馈 / chip 与输入框重复 / 默认 chip 语义不一致 / 字幕 dirty 不纳入 snapshot |
| **可选** | 1 | VideoUpload 双「开始上传」按钮 |

## 2. 详细反馈

### P0-1 / VideoList 操作列可能永久消失（localStorage required 列未强制注入）

- **位置**：`admin-web/src/views/VideoList.vue:128` `readStoredColumns`
- **症状**：反序列化 localStorage 时只 filter 已知 key、**不**强制注入 required 列（如 'operations'）；若 stored 数组缺 'operations'，`isColumnVisible('operations')` 返回 false，整张表格的操作列永久消失
- **失败路径**：用户取消全部勾选后通过 `onColumnVisibilityChange` 强制注入 required 是正确的；但若 localStorage 被外部工具改、后续新增 required key 的迁移路径漏掉、或用户从前一版本（无 required 概念）升级过来，stored 数组可能不含 'operations' → 操作列消失 → 用户无法触发详情/重转码/删除，被锁死
- **修复建议**：`readStoredColumns` 末尾 `return Array.from(new Set([...parsed.filter(known), ...required]))`，与 `onColumnVisibilityChange` 行为对称
- **作用域归属**：Phase 4 引入的列设置 localStorage 机制；本期落地路径未考虑 stored 数据迁移与外部篡改场景

### P0-2 / ImageManage 批量操作 Promise.all 无 try/catch

- **位置**：`admin-web/src/views/ImageManage.vue:609` `bulkToggleActive` + `:627` `doBulkDelete`
- **症状**：`await Promise.all(targets.map(api))` 无 try/catch；任一条 reject 即整批 reject、success toast 不显示、`load()` 不刷新、UnhandledPromiseRejection 报到控制台
- **失败路径**：用户勾选 20 张图片点「批量删除」其中 1 张被外键引用后端 409 → Promise.all 立即 reject → 「批量删除完成」toast 不显示 → await load() 不执行 → 已成功删的 19 张仍在表格显示 stale → 用户重新点删除触发 404 噪音；BulkActionBar 选中状态也不会被清空
- **修复建议**：try/catch 包裹 + 使用 `Promise.allSettled` 区分 fulfilled / rejected → `ElMessage.success(`完成 ${ok}/${total}，失败 ${fail}`)`；总是 `await load()` 与 `selectedImageRows.value = []`
- **作用域归属**：Phase 4 引入的 BulkActionBar 接入路径；Phase 3 单条 doDelete / toggleActive 已经有 try/catch + extractErrorMessage 范式，本期 bulk 路径漏抄

### P0-3 / VideoUpload type 切换静默 truncate uploadFileList 跨 step 不可见

- **位置**：`admin-web/src/views/VideoUpload.vue:96` watcher
- **症状**：`form.type` 切换到 'movie' 时 watcher 静默执行 `uploadFileList.value = [uploadFileList.value[0]]`；除 `ElMessage.warning` 之外无其它信号；wizard step 0 用 v-show 隐藏，step 1 不显示文件清单 → 用户无视觉反馈
- **失败路径**：用户 step 0 选 5 个短视频 → 进入 step 1 切类型到 'movie' → 静默丢 4 个文件 → ElMessage.warning 一闪而过 → 用户继续 step 2 → 「开始上传」只上传 1 个文件，其余 4 个完全丢失。原单页大表能直接看到 selectedFiles 列表当下被截断，新 wizard 用 v-show 隐藏 step 0 + step 1 没有显示文件清单
- **修复建议**：watcher 内若要 truncate，弹 `ElMessageBox.confirm('切换电影类型只允许 1 个文件，确认丢弃其余 N 个？', '确认丢弃文件')` 让用户显式确认；或者 step 1 切电影类型时阻止「下一步」并弹回 step 0 让用户先调整文件
- **作用域归属**：Phase 4 拆 wizard 后 v-show 隐藏 step 0 文件清单是新引入的可见性丢失场景；watcher 本身是 Phase 4 之前就存在的逻辑，但 wizard 架构放大了静默后果

### P1-1 / ImageManage 网格/列表视图切换选中状态泄漏

- **位置**：`admin-web/src/views/ImageManage.vue` `viewMode` + `selectedImageRows` + `toggleGridSelection` + `onImageSelectionChange`
- **症状**：网格视图勾选和列表 el-table 选择没有双向同步；切换 viewMode 后两端状态分裂——BulkActionBar 计数与可见选中行不一致
- **失败路径**：grid 视图勾 3 张卡片 → BulkActionBar 显示「已选 3 项」→ 切到 list → el-table 内部 selection 为空 + 表头 checkbox 未勾 + 行无高亮 → BulkActionBar 仍显示 3 → 用户点「批量删除」会删除「看不到勾选的 3 张」与所见不符；反方向亦同样
- **修复建议**：viewMode 切换时把 selectedImageRows 写回 el-table 的 `clearSelection` / `toggleRowSelection` 实现双向同步；或最简方案是切换时主动 `selectedImageRows.value = []` 强制重置（用户切换视图本身就是上下文变化、清空选择不算粗鲁）
- **作用域归属**：Phase 4 引入的 viewMode 切换 + BulkActionBar 接入交集；首次试点 BulkActionBar 与视图切换组合的状态一致性漏洞

### P1-2 / VideoList drawer 缺 destroy-on-close → blob 泄漏 + 异步竞态

- **位置**：`admin-web/src/views/VideoList.vue` `<el-drawer>` 编辑入口
- **症状**：旧 `<el-dialog destroy-on-close>` 在关闭时强制卸载内部组件（subtitle 表单 / 播放预览 iframe / blob URL）；新 `<el-drawer>` 默认未设 `destroy-on-close`，关闭后内部树仍挂载；in-flight async 请求可能写入下一次打开的 detail
- **失败路径**：用户点开视频 A → loadSubtitles(A) + refreshPlayURL(A) 并发 → 立刻关 drawer + 打开视频 B → A 的 subtitleItems 写回到现在显示 B 的 ref → drawer 内出现「B 的标题 + A 的字幕列表」混搭；A 的 blob URL 用 `URL.createObjectURL` 创建若未在切换时 `revokeObjectURL`，每次切换泄漏一份
- **修复建议**：`<el-drawer destroy-on-close>` 或在 `handleDetailClose` 内显式 cancel 所有 in-flight 请求（AbortController）+ `URL.revokeObjectURL(playURL.value)`；optional 给 loadSubtitles / refreshPlayURL 加 video_id 比对 guard
- **作用域归属**：Phase 4 dialog → drawer 改造时未把 destroy-on-close 等价语义带过来；与 Phase 3 ImageCollectionManage drawer pattern 也漏了这条

### P2-1 / VideoList BulkActionBar 失去 loading 状态

- **位置**：`admin-web/src/views/VideoList.vue` BulkActionBar `:actions` 绑定
- **症状**：旧版「批量删除」按钮挂 `:loading="deletingBatch"` 让用户看到忙状态 + 阻止重复点击；新 BulkActionBar 的 actions 数组只暴露 `label / icon / type / onClick`，**没**透传 loading 字段；`deletingBatch` ref 仍在脚本但未喂给 wrapper
- **失败路径**：用户对大量选择点「批量删除」→ 后端请求慢（5-10s）→ 按钮无 loading 反馈用户以为没生效再点 → 触发并发 `batchDeleteAdminVideos` POST → 两轮 success/failure 统计互相覆盖
- **修复建议**：扩 BulkActionBar 的 action descriptor 支持 `:loading` 字段；或在 wrapper 内部按 onClick 返回的 Promise 自动加 loading 守卫；临时方案 actions 改成 computed，按 `deletingBatch.value` 切换 action.disabled
- **作用域归属**：Phase 1 留下的 BulkActionBar wrapper API 不够；本应在 Phase 4 接入时回 Phase 1 followup 补 prop，但实际省略了

### P2-2 / VideoList quick-search v-model + chip 重复显示 + chip × 触发查询

- **位置**：`admin-web/src/views/VideoList.vue` Toolbar quick-search v-model="query.q" + activeFilterChips
- **症状**：activeFilterChips 实时读 query.q；用户敲字时同时显示输入框文字与一条「搜索：xxx」chip + chip × 触发 removeFilter('q') 把输入框清空并意外 load()
- **失败路径**：用户在 quick-search 敲入「ab」→ activeFilterChips 立刻渲染「搜索：ab」chip 与输入框内容重复 → 用户嫌冗余点 × → query.q 清空并立即 load() → 用户其实只是想关 chip 没想清空查询
- **修复建议**：activeFilterChips 排除 q 字段（输入框已经显式可见，无需 chip 重复）；或 chip × 仅清值不主动 load（让用户自己回车确认）
- **作用域归属**：Phase 4 引入的 chip 筛选模式与既有 quick-search 输入框组合的 UX 漏洞

### P2-3 / ImageManage 默认 chip 「启用：仅启用」语义不一致

- **位置**：`admin-web/src/views/ImageManage.vue` activeFilterChips（约 line 89）
- **症状**：`query.active` 默认 '1'（仅启用）；activeFilterChips 用 `query.active !== ''` 判断，初次进入页面就显示一条「启用：仅启用」chip；chip × 把 active 改成 '' 把列表语义从「仅启用」切到「全部」，与默认行为不一致
- **失败路径**：新管理员第一次打开图片管理 → 一进页面就看到「启用：仅启用」chip → 以为「这是我刚才设的筛选」点 × 清掉 → 列表语义改变（停用项也展示）；「重置筛选」按钮把 active 改成 '' 也与初始默认不一致
- **修复建议**：activeFilterChips 排除 `active === '1'`（仅默认值时不显示 chip）；或 chip × / resetFilters 把 active 还原到 '1' 而非 ''
- **作用域归属**：Phase 4 chip 筛选首次接入图片管理；首次试点暴露的「默认值如何在 chip 行表现」未定义清晰

### P2-4 / VideoList drawer 字幕变更未纳入 dirty snapshot

- **位置**：`admin-web/src/views/VideoList.vue` `captureDetailSnapshot` / `serializeDetailState`
- **症状**：VideoList 编辑 drawer 的 dirty 守卫复用 Phase 3 captureFormSnapshot 模式，但 snapshot 不含 `subtitleItems` / 字幕上传状态；用户在 drawer 内上传/删除字幕后 cancel 关闭时 confirmDetailClose 判定 not dirty 直接关
- **失败路径**：用户打开视频编辑 drawer → 字幕区上传一条新字幕成功 → subtitleItems 已变 → 用户切换思路想放弃改动点取消 → confirmDetailClose 比较时 subtitleItems 不在比较范围 → 判 not dirty 直接关无确认 → 用户感觉「dirty 守卫失效」
- **修复建议**：`serializeDetailState` 把 subtitleItems 序列化（用 hash 或 length + 各 item id 数组）纳入 dirty 比对；或字幕上传/删除时本地维护 isSubtitleDirty 标志合并入 dirty computed
- **作用域归属**：Phase 4 drawer 改造复用 Phase 3 snapshot 模式但忽略了字幕子表单是次要状态

### P3-1 / VideoUpload step 2 双「开始上传」按钮

- **位置**：`admin-web/src/views/VideoUpload.vue` wizard footer + StepRelate.vue 内嵌「开始上传」按钮
- **症状**：step 2 同时存在两个「开始上传」按钮（卡内 + sticky footer），各自调用 submit() / emit('submit')；UX 冗余 + 增加双击竞态面
- **失败路径**：用户在 step 2 看到两个相同 label「开始上传」按钮 → UX 困惑；快速连点两个按钮 → submit() 内 uploading.value 守卫多数能挡住，但视觉上「按钮没反应」让用户多次点击。同样 cancel-upload / clear-selected-files / clear-upload-records 三按钮只在 StepRelate 内，footer 无上传中操作
- **修复建议**：StepRelate 内移除「开始上传」按钮，所有 submit 操作集中到 wizard footer；或 wizard footer 在 step 2 时换成「取消上传 / 清空已选」等
- **作用域归属**：Phase 4 拆 step 时主容器 footer 和 step 子组件 footer 没有清晰职责划分

## 3. 处理建议

按 CONTEXT.md「tasks 任务三段执行流」：

- 本任务尚未生成 `DONE.md`，因此本 feedback 可在任务标记 DONE 前直接驱动一次「实现 + followup」二次 commit（与 Phase 1-3 模式一致）
- 若决定本期不返工，可分散处理：
  - **P0-1 / P0-2 / P0-3** 是数据可达性 / 数据丢失类问题，强烈建议本期 followup 落地
  - **P1-1 / P1-2** 是状态一致性 / 资源泄漏，本期 followup 与 P0 一起做最经济
  - **P2 / P3** 可延后到单独 polish 任务

## 4. 已知不在 P0–P3 列表但仍值得记的小项

- VideoList 操作列仍是 3 个明文按钮（详情 / 重转码 / 删除）+ MoreFilled 图标被 `import` 但仅用于 BulkActionBar 的「取消选择」action，并未实现「⋯」更多菜单——PRD 第 8 节 H42 承诺的「行操作 hover 显示快捷操作 +「⋯」更多菜单」未落地；字幕管理 / 改类型仍只能进 drawer 才能访问，BulkActionBar 之外无其它访问路径退化
- `StepFile.vue` 声明 `emits: ['update:fileList']` 并绑定 `@update:file-list` 到 `<el-upload>`，但 Element Plus el-upload 不会 emit 该事件，父组件 `v-model:file-list` 更新通道是死代码；实际数据回流靠 `file-change` emit + 父侧 `onFileChange` 显式 assign。功能正常但 API 契约误导
- `StepBasic` / `StepRelate` 直接 v-model 父组件 reactive `form` 的字段（如 `form.type` / `form.title` / `form.tags`）；通过 Proxy 写传播工作正常但破坏单向数据流契约，将来若要把 form 抽 readonly 包装就会立刻报错
- VideoUpload sticky wizard footer 与 `.upload-page padding-bottom: var(--space-1)` 比例不当，结果区 SectionCard 底部可能被 footer 遮挡

## 5. 后续建议

- Phase 1 BulkActionBar wrapper API 需要扩 `loading` / `disabled` action 属性，supports actions[].loading / disabled / busy ref 三种形式之一；本期 P2-1 修复需要扩 wrapper（回 Phase 1 followup commit）
- drawer 改造的 dirty snapshot 应该包含**所有**会被用户编辑的字段（含字幕 / 演员 / 合集 id 数组等次要 state），形成 Phase 3-4 编辑入口 Drawer 契约的完整版；建议在 CONTEXT.md「admin 编辑入口 Drawer」段补「snapshot 必须覆盖所有可见可编辑字段」一条
- BulkActionBar + 视图切换的组合在本期首次接入暴露状态不一致问题；如未来其它视图复用此组合，应在 wrapper 文档强调「切换上下文必须清空选中」契约
