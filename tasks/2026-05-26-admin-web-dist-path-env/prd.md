# PRD：admin-web dist 路径支持 `ADMIN_WEB_DIST_PATH` 环境变量

- 日期：2026-05-26
- 目标端：后端 `internal/handlers/router.go` 静态资源挂载逻辑
- 范围：让 Go server 的 `/admin/*` 静态目录可由环境变量定位绝对路径，不再硬依赖 CWD 下的 `admin-web/dist`

## 1. 用户故事

作为 [[家用部署机]] 的运维者，我希望：

1. Go server 以**绝对路径**定位 admin-web 构建产物 —— 因为部署机的工作目录会随 push symlink-swap，相对路径 `admin-web/dist` 解析到 CWD 下，CWD 一变 `/admin` 就 404。
2. 不破坏 dev 模式：开发者跑 `dev-up.sh` 时，没有设环境变量也能找到 `admin-web/dist`（保持当前默认行为）。

## 2. 作用域

- **覆盖**：
  - `internal/handlers/router.go:221` 的 `adminDist := filepath.Join("admin-web", "dist")` 改为优先读 `ADMIN_WEB_DIST_PATH`，未设时回退原默认。
  - `internal/config/config.go` 增加对应 `AdminWebDistPath` 字段，归属现有 `Config` 结构，沿用 `getEnv` 模式。
  - Router 通过 `Config` 注入或保留 `os.Getenv` 现读 —— 二选一（实施阶段定，PRD 不锁）。
  - `.env.example` 新增 `ADMIN_WEB_DIST_PATH=` 一行（值留空 = 用默认）。
  - 单测：覆盖"env 未设 → 回退默认"、"env 设了 → 用 env 值"两条分支。
- **不覆盖**：
  - `STORAGE_ROOT` / `POSTER_STORAGE_PATH` 改造 —— 这些字段已经是 env-driven 的，无需改代码（[[家用部署机绝对路径契约]] 只要求部署机的 `.env` 配绝对值，代码本身已支持）。
  - admin-web 自身的 build 流程（vite 配置 / `npm run build` 行为）。
  - dev-up.sh 任何改动（dev 模式继续走默认相对路径）。
  - Go server 整体启动流程、其它静态资源（swagger 等）。

## 3. 非目标（明确不做）

- **N1**：不引入"必须用绝对路径"的运行时校验。允许使用者传相对路径（虽然不推荐），代码不主动拒绝。dev 默认值本身就是相对路径。
- **N2**：不重构 router 静态挂载结构（不抽 `mountAdminStatic` 子函数等）。只改路径来源一处，最小改动。
- **N3**：不为 dist 缺失情形改动现有 `os.Stat ... IsDir()` 兜底。挂载条件不变。
- **N4**：不增加 dev/prod mode flag。env-driven 即可。
- **N5**：不改 Config 中其它字段；不顺手补 `POSTER_STORAGE_PATH` 等已有字段的文档（独立 task 处理）。
- **N6**：不在 `Config.Validate()` 之类的地方加新校验（项目目前没有这种统一校验入口）。

## 4. 术语对应

本 task 不引入新的 CONTEXT 术语；它是 [[家用部署机绝对路径契约]] 的**实现侧支撑**，沉淀在 [[家用部署机]] 的 ADR-0005 与 docs/家用部署机.md 中已经说明。实施完成后无需 sync CONTEXT.md。

## 5. 关键决策

| # | 决策点 | 选择 | 理由 |
|---|---|---|---|
| D1 | env var 名 | `ADMIN_WEB_DIST_PATH` | 与 docs/家用部署机.md §2 已经写出的 env 名一致 |
| D2 | 默认值 | 维持 `filepath.Join("admin-web", "dist")`（相对） | 不破坏 dev-up.sh / 开发者 go run 体验 |
| D3 | 路径来源传递 | 优先经 `Config.AdminWebDistPath`，router 从 Config 读 | 与现有 `StorageRoot` 等字段一致，单测可注入 mock Config；如果 `Config` 在 router 路径上不可达，退而求其次用 `os.Getenv` 现读 |
| D4 | 失败兜底 | 不变：env 设了但目录不存在时 `os.Stat ... IsDir()` 失败，等同于"无 dist"分支，Go server 不挂 /admin | 与今天一致，不引入新错误行为 |
| D5 | env 值末尾斜杠 | 不规范化（让 `filepath.Join` 自然处理）| YAGNI |

## 6. 验收

### 6.1 自动化

- `go test ./internal/handlers -run TestAdminDist` 至少覆盖两个用例：
  - env 未设 + `admin-web/dist` 不存在 → /admin 路由未注册
  - env 设到一个临时存在的 dir → /admin/assets 与 /admin 路由按该目录服务
- `go test ./internal/config -run TestAdminWebDistPath` 覆盖 env→Config 字段读取（如果走 D3 主路径）。
- `go build ./...` 通过。

### 6.2 手动

- 在开发机：`go run main.go -mode server`（无 env），访问 `http://127.0.0.1:8080/admin/` —— 行为与本 task 前一致（dist 存在则正常挂，不存在则不挂）。
- 在开发机：`ADMIN_WEB_DIST_PATH=/tmp/empty go run main.go -mode server`，先 `mkdir -p /tmp/empty/assets && echo hello > /tmp/empty/index.html`，访问 `/admin/` 返回 hello。
- 不需要真上家用部署机验证。家用部署机首次 push 时自然会走这条路径。

## 7. 风险与回滚

- 风险：路径解析在 Windows 上的兼容性 —— 本项目只跑 macOS / Linux，不考虑。
- 回滚：单 commit revert 即可，无 schema 改动、无 client 变化。
