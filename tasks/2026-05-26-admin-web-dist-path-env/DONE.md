# DONE：admin-web dist 路径环境变量

- 完成日期：2026-05-26
- 关联 commit：见仓库 `git log` 与 `git diff` 范围
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`
- 关联 Review：`./review.md`
- 关联 ADR：`docs/adr/0005-home-deployment-architecture.md`

## 实施摘要

按 TDD 红→绿完成 [[家用部署机]] ADR-0005 点名的代码前置条件：`internal/handlers/router.go` 的 admin-web dist 路径从硬编码 `filepath.Join("admin-web", "dist")` 改为读 `ADMIN_WEB_DIST_PATH` 环境变量，默认值 `admin-web/dist` 保持现状以不破坏 dev-up.sh 体验。

实施走 PRD §5 D3 主路径（不走 os.Getenv 退路）：

- `internal/config/config.go` `Config` 加 `AdminWebDistPath` 字段，`Load()` 用 `getEnv("ADMIN_WEB_DIST_PATH", "admin-web/dist")` 装填。
- `internal/handlers/router.go` 抽出 `mountAdminStatic(r *gin.Engine, adminDist string)` 私有 helper，原 inline 21 行挪进去；`API` struct 加 `adminWebDistPath` 字段；`NewAPI` 新增第 23 个 positional 参数；`Register` 调用 `mountAdminStatic(r, a.adminWebDistPath)`。
- `main.go` 唯一的 `handlers.NewAPI` 调用点传 `cfg.AdminWebDistPath`。
- `.env.example` 新增 `ADMIN_WEB_DIST_PATH=` 注释行（留空 = 默认；家用部署机请填绝对路径）。

## 验证

### 自动化（review.md §0 准入）

- ✅ 红灯证据（实施前）：
  - `internal/handlers/admin_static_test.go:27,55` → `undefined: mountAdminStatic`
  - `internal/config/config_admin_dist_test.go:17,18,32,33` → `cfg.AdminWebDistPath undefined (type Config has no field or method AdminWebDistPath)`
- ✅ 绿灯：`go test ./internal/config ./internal/handlers -run 'TestLoadAdminWebDistPath|TestMountAdminStatic' -count=1` 全过
- ✅ 全套：`go test ./... -count=1` 全绿（不影响其它包）
- ✅ `go build ./...` BUILD SUCCESSFUL（含 main.go 新增参数）
- ✅ `go vet ./internal/handlers ./internal/config` 无 warning
- ✅ `git diff --stat` 范围：4 个生产代码文件 + 2 个新测试 + 1 个 task 完成标记，与 implement.md §3.1 / §3.2 一致

### 单测覆盖（PRD §6.1 / review.md §1）

- ✅ `TestLoadAdminWebDistPathFromEnv` — env 设了 → `Config.AdminWebDistPath` 等于该值
- ✅ `TestLoadAdminWebDistPathDefaultsToRelative` — 未设 → 默认 `admin-web/dist`
- ✅ `TestMountAdminStaticServesIndexFromGivenDir` — 给临时 dir → /admin/ 返回该 dir index.html、/admin/assets/* 返回该 dir 资源
- ✅ `TestMountAdminStaticSkipsWhenDirMissing` — 目录不存在 → /admin/ 不挂、返回非 200

### 手测（review.md §1.1 / §1.2 / §1.3）

- ⚠ **未在本机做端到端启动手测**。原因：实施时开发机 dev-up.sh 已经把旧 server 拉起来占着 :8080（pid 46690）、Postgres / Redis 容器跑了 12 小时，强制 kick 重启会冲击你正在跑的 dev 工作流。本改动属 "plain wire + helper extraction" —— `mountAdminStatic` 行为由单测两条覆盖（dir 存在 / 不存在），`Config.Load` 由两条覆盖（env 设 / 默认），`main.go` 的 cfg → NewAPI 接线由 `go build` + `go vet` 静态保证，没有运行期才会暴露的语义。
- 📌 **延后到 [[家用部署机]] 首次启用时执行**：那时部署机会用新二进制 + `ADMIN_WEB_DIST_PATH` 绝对路径首次启动，review.md §1.1 / §1.2 / §1.3 三场景自然覆盖。届时若有偏差，按 [[家用部署机 fail-open 构建契约]] 旧 binary 继续跑、修复后重新 push 即可。

### 代码审查（review.md §2）

- ✅ §2.1 [[家用部署机绝对路径契约]]：默认值仍是相对路径不破坏 dev；docs/家用部署机.md §2 第 2 步示例值与本实现 1:1 对齐
- ✅ §2.2 dev-up.sh 不破坏：未触碰 dev-up.sh；不设 env 时 server 仍走 `admin-web/dist` 与旧行为完全一致
- ✅ §2.3 与 [[migration 前向兼容契约]] 无关：本任务无 SQL 改动
- ✅ §2.4 Karpathy：改动最小（1 字段 + 1 函数提取 + 1 参数 + 1 env 文档 + 4 测试用例）；没引未来才用的字段；没加 feature flag
- ✅ §2.5 测试有效性：红灯先于实现确认；`t.Setenv` 用法正确；单测不依赖 docker/postgres/redis/网络

### 兼容回归（review.md §3）

- ✅ `/admin/*` 路由集合不变（mountAdminStatic 内部逻辑与原 inline 完全等价，包括 NoRoute 兜底）
- ✅ admin-web 前端代码不变
- ✅ swagger 路径不变

### 文档同步（review.md §4）

- ✅ PRD / implement / review 与最终实现一致，无偏离
- ✅ docs/家用部署机.md §2 第 2 步对 `ADMIN_WEB_DIST_PATH` 的引用与代码一致
- ✅ CONTEXT.md 不需要 sync（PRD §4 明确本任务不引入新术语）

## 后续

[[家用部署机]] 启用还差**一步**：按 `docs/家用部署机.md` §2 在部署机做一次性安装（外挂盘挂载 + bare repo 初始化 + launchd plist + post-receive hook + 首次 push 测试）。这是机器特定操作，不在 repo 内可做。
