# Feedback：admin 重设计 Phase 2 · code review 反馈

- 日期：2026-05-24
- 关联 commit：`359a6489 完成 admin 简单视图重排`
- 来源：grill-with-docs / code-review skill（high effort，3 angles × 6 candidates，1-vote verify）
- 性质：归档反馈，不强制触发返工；本次会话用户已要求把反馈记入此文件

## 1. 优先级分布

| 等级 | 数量 | 摘要 |
|------|------|------|
| **必修 P0**（安全 + 数据正确性） | 3 | TaskMonitor 状态筛选前后端契约偏差 / UserManage 用公共注册端点 + 非原子 / Dashboard echarts 实例泄漏 |
| **必修 P1**（UX + 监控信号） | 3 | 双 PageHeader 标题 / TaskMonitor 删失成功率信号 / IPTVManage 双 el-upload 共享 v-model |
| **强烈建议** | 3 | UserManage 部分提交错文案 / ActorManage 乐观更新不 reload / ActorManage 性别迁移兼容 |
| **可选** | 1 | TaskMonitor 请求竞态 |

## 2. 详细反馈

### P0-1 / TaskMonitor 状态筛选前后端契约偏差

- **位置**：`admin-web/src/views/TaskMonitor.vue:152` ⇄ `internal/handlers/admin.go:732 AdminTasks`
- **症状**：状态筛选 chip 发送 `status` 查询参数，但后端 `AdminTasks` 只读 `page` / `page_size`，整个 status 字段被服务端静默忽略；UI 显示已筛选，实际返回全集。
- **失败路径**：运维点「失败」chip 期望只看失败任务排查异常 → 后端返回 page+page_size 切片下的 pending/running/success/failed 混合列表 → 运维误判系统只有一两条失败、错过待处理事故；或在批量重试时把成功任务也一起拉进重试目标。
- **修复建议**：后端 `repo.AdminListTranscodingTasks(ctx, status, page, pageSize)` 加 status 参数；前端契约不变。前端独立修复无效。
- **作用域归属**：Phase 2 视觉重排引入「状态筛选 chip」是产品决策升级，后端没跟上；修复跨前后端，需 Go 侧改 admin handler + repo SQL。
- **可附属测试**：`internal/repository/*_test.go` 加 `AdminListTranscodingTasks_with_status_filter` 表驱动用例覆盖 pending/running/failed/success/all 五挡。

### P0-2 / UserManage 使用公共注册端点 + 二步非原子

- **位置**：`admin-web/src/api/admin.js:46`（新引入）、`admin-web/src/views/UserManage.vue:96` saveUser
- **症状**：`registerAdminUser` POST 到 `/auth/register`；`internal/handlers/router.go:87` 该路由挂在 `v1.Group("/auth")` **无 AuthMiddleware**——public 路由；且 `/auth/register` 总是返回 role='user' + 一对 JWT，UI 立即丢弃 token、再调 `updateUserRole` 升 admin，二步非原子。
- **失败路径**：
  - (a) 攻击面：任何匿名 curl 都能 POST `/auth/register` 创建账号；admin 按钮只是 UI 包装一条公开能力当作管理能力来设计。
  - (b) 部分提交：admin 在创建对话框选 `role=admin` → registerAdminUser 成功（user 已落库且为 'user' 角色）→ updateUserRole 因网络抖动 / 403 / 500 失败 → catch 提示「创建用户失败」但用户已建好停留在 'user' 角色，admin 重试触发 409 duplicate-username，孤儿账号需 DBA 手动清理。
- **修复建议**：
  - 推动后端新增 `POST /admin/users` 原子创建（认证 + 指定角色 + 不返回访问 token）
  - 前端 `registerAdminUser` 切到该新端点；删除 saveUser 的二步 updateUserRole 链
  - 公共 `/auth/register` 是否封闭、是否限速、是否仅开放给特定 invite token，属于产品安全策略另议
- **作用域归属**：本期 PR 引入了新 API 调用点；属于 Phase 2 引入的安全 + 数据正确性风险。

### P0-3 / Dashboard echarts 实例 v-if 重挂载后写到 detach 旧节点

- **位置**：`admin-web/src/views/Dashboard.vue:158` `renderChart()` + `chartRef` 在 `v-if="trendPoints.length"` 内
- **症状**：echarts `chart` 在模块作用域常驻；`<div ref="chartRef">` 被 v-if 卸载后再次有数据时是全新 DOM 节点，`renderChart()` 由于 `if (!chart)` 早绿就跳过 `echarts.init`，`setOption` 写到已 detach 的旧 canvas。
- **失败路径**：用户首次进 Dashboard 后端返回 errorMessage → `v-else` 整片连同 chartRef 卸载 → 用户点重试成功 → chartRef 重新挂载为新 DOM 节点但 `chart` 仍指向旧的 detached canvas → setOption 不报错也不渲染，趋势图永远空白直到整页刷新；resize 事件继续 fire 到已 detach 实例。
- **修复建议**：
  - `renderChart()` 进入时若 `chartRef.value` 的 `dom === chart?.getDom?.()` 不一致则 `chart.dispose(); chart = null` 再 init
  - 或：把 chart 实例提升到 ref + watch chartRef 变化做 dispose + re-init
  - 同时验证 `resolveColor` 读 token 时机：若 `getComputedStyle(...).getPropertyValue('--primary')` 在 theme.css 加载前返回空字符串，echarts 会渲染默认黑色——`renderChart` 在 nextTick 后调即可避免

### P1-1 / shell 顶栏与视图体内 PageHeader 双显标题

- **位置**：`admin-web/src/components/Layout.vue:267` `<PageHeader class="shell-page-header" :title="pageTitle" />` ⇄ 7 视图各自的 `<PageHeader title="..." />`
- **症状**：DOM 上出现两份 H1，屏幕阅读器重复播报「系统仪表盘 系统仪表盘」；视觉上同一标题占两行顶部空间。
- **失败路径**：进入 `/dashboard` → shell 顶栏渲染「系统仪表盘」+ 下面 main 区域 PageHeader 又渲染「系统仪表盘」；7 个 view 都中招。
- **修复建议**（二选一）：
  - (a) 每个 view 删除 `<PageHeader />` 调用，把右侧 actions 也提到 shell 层（slot 暴露）；shell PageHeader 唯一负责
  - (b) shell 顶栏改成不显示标题、只显示品牌或路径面包屑，让视图内 PageHeader 唯一负责
- **作用域归属**：Phase 1 落地 shell PageHeader + Phase 2 落地视图 PageHeader 两次叠加触发；属于 Phase 2 暴露的 Phase 1 与 Phase 2 协调缺陷。

### P1-2 / TaskMonitor 删除「已完成」与「任务总量」StatCard 丢失监控信号

- **位置**：`admin-web/src/views/TaskMonitor.vue:159`
- **症状**：顶部 StatCard 重排后只剩「队列 / 处理中 / 失败」三档；旧版「已完成」与「任务总量」两档被删，但后端 `total_count` / success 统计仍可读。
- **失败路径**：运维巡检转码流水线健康度：旧 UI 一眼能看到「失败 / 已完成 / 总量」三维指标推断成功率；新 UI 只有失败 + 队列长度，无法判断「失败 5 条是否意味成功 5000 条」还是「失败 5 条 总共 5 条」；某段时间 worker 卡死、success 统计不增时，新 UI 看到的只有「队列长度上涨」，但看不到成功数停滞，隐藏关键运营信号。
- **修复建议**：StatCard 由 3 档恢复到 4 档（队列 / 处理中 / 已完成 / 失败），或加一个独立的「总任务量」与「成功率」双行卡片。

### P1-3 / IPTVManage 双 el-upload 共享 v-model:file-list

- **位置**：`admin-web/src/views/IPTVManage.vue:178`
- **症状**：Toolbar 内「上传 M3U」`<el-upload>` 与 SectionCard 内「本地 M3U 文件」`<el-upload>` 两个组件共享 `v-model:file-list="uploadFiles"`；Toolbar 的 upload 没有「上传并替换」触发器，选完文件不真发请求。
- **失败路径**：用户在 Toolbar 点「上传 M3U」选文件 A → file-list 写进共享 `uploadFiles` → 用户期待上传，但 Toolbar 那个 el-upload 不调 `uploadPlaylist()`，文件只暂存在 v-model；真正触发上传的按钮在 SectionCard 内部，用户不知道；同时 SectionCard 拖拽区域显示「文件 A」让用户更困惑。
- **修复建议**（二选一）：
  - (a) Toolbar 的「上传 M3U」改成跳转到 SectionCard 内置上传区（scrollIntoView + focus），不在 Toolbar 内放 el-upload
  - (b) Toolbar 内的 el-upload 自己挂 `:on-success` 直接调 `uploadPlaylist`，并使用独立 file-list ref，与 SectionCard 区分

### P2-1 / UserManage saveUser 部分提交错文案

- **位置**：`admin-web/src/views/UserManage.vue:96` saveUser
- **症状**：与 P0-2 相关。即便不解决「公共注册」问题，当前 saveUser catch 块对「user 已创建但角色升级失败」与「user 未创建」无差别提示「创建用户失败」。
- **修复建议**：finally 内 `await load()` 同步表格；try/catch 拆为两个嵌套——内层 updateUserRole catch 区分文案「用户已创建但角色未生效，请在列表中找到 X 并手动改角色」。

### P2-2 / ActorManage toggleActive 不 reload

- **位置**：`admin-web/src/views/ActorManage.vue:296`
- **症状**：调用 `updateAdminActor` 成功后只乐观本地翻转 `row.active = !row.active`，不触发 `load()`；服务端规范化或字段写入差异时 UI 与 DB 状态不同步。
- **失败路径**：admin 点「停用」→ updateAdminActor 成功 → UI 翻 false + toast「演员已停用」→ 服务端规范化保留旧值（如必填字段缺失时强制不变）→ admin 刷新页面看到状态回弹「演员已启用」与刚才 toast 完全相反；`updated_at` 列也变 stale。
- **修复建议**：toggleActive 末尾 `await load()` 强对齐服务端真相，移除乐观更新代码。

### P2-3 / ActorManage 性别枚举 Chinese→English 无迁移

- **位置**：`admin-web/src/views/ActorManage.vue:414`
- **症状**：el-select option `value="male" / "female" / "unknown"`，但既存 actor 行 `row.gender` 可能存的是历史 Chinese 值（`'男'` / `'女'` / `'未知'`）；form.gender 被赋为「男」时 select 找不到 matching option，dialog 显示为空。
- **失败路径**：数据库既存演员 gender=`男` → openEdit 把 form.gender 设为「男」→ el-select 找不到匹配 → select 显示空但 form.gender 仍是「男」→ admin 没注意改了别的字段保存 → updateAdminActor 把 gender=「男」回写。表格 `prop="gender"` 列混杂显示「男」/「female」/「unknown」。
- **修复建议**：openEdit 时 normalize（`'男' → 'male'` / `'女' → 'female'` / `'未知' → 'unknown'`）；可选地写一次性 Go 迁移脚本把 DB 内 Chinese 性别值全部 normalize 到英文，避免新旧并存。

### P3-1 / TaskMonitor 请求竞态

- **位置**：`admin-web/src/views/TaskMonitor.vue:122` setStatus
- **症状**：setStatus(filter) 立即触发 load()，5 秒 setInterval 也并行 fire；两条 load 没有 abort controller / request-id 守卫，后到的请求可能覆盖先到。
- **失败路径**：admin 点「失败」chip T=0 触发 load1（耗时 1.2s）；T=0.3s 5 秒间隔 timer fire load2；如 load2 早返、load1 迟到，list.value 先被 load2 写、再被 load1 写——若 load1 还带的是上一秒未合并的 query 副本，列表可能短暂闪回上一筛选；loading flag 来回翻转造成 spinner 闪烁。
- **修复建议**：load 内用 AbortController 取消旧请求 或 用 request seq 比较丢弃过期返回；setStatus 后重置 interval 计时器。

## 3. 处理建议

按 CONTEXT.md「tasks 任务三段执行流」：

- 本任务尚未生成 `DONE.md`，因此本 feedback 可在任务标记 DONE 前直接驱动一次「实现 + followup」二次 commit（与 Phase 1 模式一致）
- 若决定本期不返工，可在 Phase 3 / Phase 4 中分散处理：
  - P0-1 / P0-2 涉及后端，需要单独安排 Go 侧任务
  - P0-3 / P1-1 / P1-2 / P1-3 是纯前端 UX 修复，可在 Phase 3 简单视图复用模式相近的修复中顺带
  - P2 / P3 可延后到 Phase 4 或单独 polish 任务

## 4. 已知不在 P0–P3 列表但仍值得记的小项

- `Dashboard.vue:49` resolveColor 在 theme.css 加载前可能返回空字符串（与 P0-3 同一 echarts 链路，修 chart 实例泄漏时一并处理）
- `ActorManage.vue:358` 演员状态切换按钮丢失 `:type="row.active ? 'warning' : 'success'"` 语义色提示；候选卡片头像 / external_id / aliases 简化为纯文字，识别候选准确率下降
- `EmptyState` 嵌进 dialog 内时 `min-height` 占位较大，可能挤压中等屏 dialog 表单底部按钮；ActorManage 创建对话框的候选区受影响
- `themeTokens.spec.js` 已扩展视图层 audit，本期 8 文件 hex 全部清零，spec 绿
