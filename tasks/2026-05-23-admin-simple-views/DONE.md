# Phase 2 完成记录 · admin 简单视图重排

- 完成日期：2026-05-24
- 关联提交：
  - `359a6489` 完成 admin 简单视图重排（首版实现：7 视图按 Phase 1 设计系统重排 + 视图层硬编码 hex 清零）
  - `5181045f` 修复 admin 简单视图反馈问题（落地 feedback.md 的 P0–P3 共 9 项 + 1 小项）

## 验证摘要

### 自动化必跑
- `cd admin-web && npm test` —— 全绿（67 + 新增 phase 2 views hex audit 8 用例 = 75 用例）
- `cd admin-web && npm run build` —— 通过
- Go 端新增定向测试：`internal/handlers/admin_create_user_test.go`（管理端原子建用户）+ `internal/repository/admin_repository_test.go`（任务状态筛选契约）双绿

### 7 视图 H21–H27 手测覆盖（外观锚点 + 功能不破二维）
- H21 Dashboard：PageHeader + 8 StatCard + SectionCard 包 ECharts + EmptyState；stats 数字、resize、加载失败重试均验证
- H22 SystemSettings：PageHeader + N SectionCard 分组；表单保存 / 读取 / 校验仍生效
- H23 UserManage：PageHeader + Toolbar（刷新 + 添加用户）+ EmptyState；新建 / 改角色 / 删除已切到管理端原子接口
- H24 TaskMonitor：PageHeader + Toolbar（刷新 + 4 档状态筛选 chip）+ 4 StatCard（队列 / 处理中 / 已完成 / 失败）+ EmptyState；状态筛选已与后端契约对齐
- H25 IPTVManage：PageHeader + StatCard×N + SectionCard（M3U 源）+ Toolbar（只剩唯一上传入口）+ EmptyState
- H26 CollectionManage：PageHeader + Toolbar + EmptyState；CRUD + 视频绑定仍生效
- H27 ActorManage：PageHeader + Toolbar + EmptyState；历史 Chinese 性别值兼容、停用切换后强制 reload

### H28 自动化全局验证
- npm run build / npm test 全绿
- `themeTokens.spec.js` 视图层 audit 8 文件全部命中：Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage / UploadProgress 共 8 文件零玫红 hex

### 截图归档（`screenshots/`）
- before / after 14 张 PNG（7 视图 × before/after），统一 1440px 桌面分辨率
- 5181045f 重新生成 6 张 after 截图（Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage 7 视图反馈修复后的最新形态），before 保留对照

## Feedback 落地清单（`feedback.md` 共 10 项）

| 等级 | 反馈 | 落地动作 | 状态 |
|------|------|---------|------|
| P0-1 | TaskMonitor 状态筛选前后端契约偏差 | `internal/handlers/admin.go` + `internal/repository/admin_repository.go` 加 status 参数；`internal/repository/admin_repository_test.go` 表驱动覆盖五挡 | ✅ |
| P0-2 | UserManage 用公共 `/auth/register` + 非原子 | 后端新增 `POST /admin/users` 原子创建端点；前端 `registerAdminUser` 切到新端点；`admin_create_user_test.go` 定向测试 | ✅ |
| P0-3 | Dashboard echarts 实例 v-if 重挂载泄漏 | `Dashboard.vue` `renderChart` 进入时若 DOM 不一致则 `chart.dispose(); chart = null` 重 init | ✅ |
| P1-1 | shell 顶栏与视图体内 PageHeader 双显标题 | `Layout.vue` 新增 `showShellPageHeader` computed + `route.meta?.hideShellPageHeader` 标记；`router/index.js` 给视图路由设置 `meta.hideShellPageHeader=true`；视图保留 PageHeader 作为唯一标题源 | ✅ |
| P1-2 | TaskMonitor 已完成 / 总量 StatCard 丢失 | 4 档 StatCard 复原（队列 / 处理中 / 已完成 / 失败） | ✅ |
| P1-3 | IPTVManage 双 el-upload 共享 v-model | 移除 Toolbar 内 `<el-upload>`，Toolbar 上传按钮改为聚焦到 SectionCard 内的唯一上传区域 | ✅ |
| P2-1 | UserManage saveUser 部分提交错文案 | 切到原子接口后单步完成，错误文案区分天然消除 | ✅（随 P0-2 同步） |
| P2-2 | ActorManage toggleActive 不 reload | `toggleActive` 末尾 `await load()` 强对齐服务端真相 | ✅ |
| P2-3 | ActorManage 性别 Chinese→English 无迁移 | `openEdit` 时 normalize（`'男' → 'male'` / `'女' → 'female'` / `'未知' → 'unknown'`）；列表渲染保留兜底 | ✅ |
| P3-1 | TaskMonitor setStatus 与 5s interval 竞态 | 留作后续 polish（轻微 UX 抖动，未阻塞功能） | ⏸️ 延后 |

### 小项归档
- `Dashboard.vue` resolveColor 在 theme.css 加载前空字符串：随 P0-3 修 chart 实例时一并解决（`nextTick` 后调用 `renderChart`）
- `ActorManage.vue:358` 切换按钮 `:type="row.active ? 'warning' : 'success'"` 语义色：未恢复（功能不影响，视觉次要）
- `EmptyState` 嵌进 dialog 内 `min-height` 占位较大：未调整（暂未出现实际挤压）

## CONTEXT.md 沉淀
- 本期未追加新术语（沿用 Phase 1 已有的 6 条「admin 设计系统术语」+ 既有「管理端手动刮削术语」/「AV 上传分类术语」等）；feedback.md 的修复路径已落进 plan.md 与提交记录，不需要新术语

## plan.md 追加条目
- 2026-05-24 Phase 2 实施 + 截图归档记录（359a6489）
- 2026-05-24 Phase 2 反馈修复记录（5181045f）

## 已知遗留（移交 Phase 3 或单独 polish 任务）

- **P3-1**（TaskMonitor 状态切换与自动刷新计时器并发）：留作单独 polish；引入 AbortController / request seq 守卫；不阻塞 Phase 3
- **`#app :where(...)` 全局 specificity 收口**（Phase 1 遗留）：仍未处理；Phase 3 / Phase 4 视图若被压制则就地用更高 specificity 的 scoped 选择器解决，全局重构留单独 follow-up
- **小项 ActorManage 切换按钮语义色 + EmptyState dialog min-height**：观察 Phase 3 / Phase 4 是否出现实际触发场景再决定是否复原 / 调整
