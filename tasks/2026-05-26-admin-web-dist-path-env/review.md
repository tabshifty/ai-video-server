# Review：admin-web dist 路径环境变量

- 日期：2026-05-26
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`
- 关联 ADR：`docs/adr/0005-home-deployment-architecture.md`

## 0. 准入条件（任一未满足 → 回到 implement）

- [ ] `internal/config/config.go` 含 `AdminWebDistPath` 字段，`Load()` 中通过 `getEnv("ADMIN_WEB_DIST_PATH", ...)` 装填
- [ ] `internal/handlers/router.go` 不再出现 hard-coded `filepath.Join("admin-web", "dist")` 字面量（或仅作为默认值出现在 config 层）
- [ ] `.env.example` 含 `ADMIN_WEB_DIST_PATH=` 一行 + 注释
- [ ] `go build ./...` BUILD SUCCESSFUL
- [ ] `go test ./internal/config ./internal/handlers -count=1` 全绿
- [ ] 新增至少两个单测：env 设 → 用 env 值；env 未设或目录不存在 → /admin 未挂或返回 404
- [ ] `git diff` 范围仅在 implement §3 的文件清单内（含可接受的 main.go 单点改动用于注入 cfg 到 router）
- [ ] `plan.md` 追加本任务进度
- [ ] PRD §6.2 两条手动场景跑通

## 1. 手测脚本（按 PRD §6.2 验收映射）

### 1.1 默认行为不破坏 dev 体验

1. 工作目录 = repo 根，没有设 `ADMIN_WEB_DIST_PATH`，先 `cd admin-web && npm run build`
2. `go run main.go -mode server`
3. 浏览器访问 `http://127.0.0.1:8080/admin/`
   - [ ] 返回 admin SPA 首页（与改动前一致）
4. 删掉 `admin-web/dist`，重启 server
   - [ ] `/admin/` 返回 404（NoRoute 兜底；与改动前一致）

### 1.2 env 接管路径

```bash
mkdir -p /tmp/admin-dist-test/assets
echo '<html><body>hello-from-env</body></html>' > /tmp/admin-dist-test/index.html
echo '/* fake */' > /tmp/admin-dist-test/assets/index.css
ADMIN_WEB_DIST_PATH=/tmp/admin-dist-test go run main.go -mode server
```

1. 访问 `http://127.0.0.1:8080/admin/`
   - [ ] body 含 `hello-from-env`
2. 访问 `http://127.0.0.1:8080/admin/assets/index.css`
   - [ ] 返回 `/* fake */`
3. Ctrl-C 关闭 server，清理 `/tmp/admin-dist-test`

### 1.3 env 指向不存在目录

```bash
ADMIN_WEB_DIST_PATH=/tmp/this-does-not-exist go run main.go -mode server
```

1. 访问 `/admin/`
   - [ ] 返回 404；server 进程**不崩**

## 2. 代码审查清单

### 2.1 [[家用部署机绝对路径契约]] 对齐

- [ ] env 值期望写绝对路径，但代码不强制（dev 默认就是相对）
- [ ] 文档（`docs/家用部署机.md` §2 第 2 步）已经把 `ADMIN_WEB_DIST_PATH=/Users/<你>/deploy/ai-video-server/current/admin-web-dist` 写出，本任务实现侧与之 1:1 对齐

### 2.2 dev-up.sh 不破坏

- [ ] `bash scripts/dev-up.sh` 仍正常拉起；不设 env 时 server 仍从 `admin-web/dist` 加载（dev-up.sh 的 frontend build 行为不变）
- [ ] dev-up.sh 不需要任何改动

### 2.3 [[migration 前向兼容契约]] 无关

- [ ] 本任务无 SQL 改动，与 migration 契约无关

### 2.4 Karpathy 规则 (`.codex/skills/karpathy-guidelines` 等价)

- [ ] 改动范围最小：只一个字段 + 一处 router 读取 + 一处 env 文档 + 单测
- [ ] 没有引入"未来可能用到"的字段或配置项
- [ ] 没有为兼容旧行为额外加 feature flag

### 2.5 测试有效性

- [ ] 单测在改动**前**会失败（红灯先确认）—— review 时要求贴一段"红灯证据"或在 plan.md 注明红灯日期 / 命令 / 输出摘要
- [ ] 单测不依赖外部网络、不依赖 docker 容器、不依赖 Postgres / Redis
- [ ] `t.Setenv` 用法正确，测试结束后 env 自动还原

## 3. 兼容回归

- [ ] 现有 `/admin/*` 路由集合不变（仍是 GET `/admin`、GET `/admin/`、`/admin/assets/*` 静态、NoRoute 兜底）
- [ ] 现有 admin-web 任何前端代码不需要改（路径变化只发生在 Go server 启动期）
- [ ] swagger `/docs/swagger/` 路径与本任务无交集，未被影响

## 4. 文档同步

- [ ] PRD `./prd.md` 与 implement `./implement.md` 的非目标 / 决策表与最终实现一致；如有偏离需在 plan.md 注明
- [ ] `docs/家用部署机.md` §2 第 2 步对 env 名称的引用与代码一致；不引用任何已删除字段
- [ ] **不**需要 sync CONTEXT.md（PRD §4 已说明本任务不引入新术语）

## 5. 完成标记

满足全部准入条件 + 手测全过 + review 清单全勾后：

- [ ] 在本目录新增 `DONE.md`，按 [[tasks 任务三段执行流]] 约定记录完成日期、关联 commit、验证摘要
- [ ] 提交一个**单 commit**，标题用中文，正文里点名"本 task 是 [[家用部署机]] 启用的前置条件之一"
