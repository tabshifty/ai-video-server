# 家用部署机架构：bare-repo post-receive + launchd 硬切重启 + fail-open 构建

家庭日常使用的 Android TV / phone 客户端需要一台 24×7 在线的服务器；开发动作在另一台机器上发生。为了在"开发机 push 完代码立即生效"这一诉求下尽量保留简单性与可维护性，本决策一次性确定 [[家用部署机]] 的完整架构姿态：开发机直接 SSH push 到部署机本地的 bare git repo，部署机的 `post-receive` hook 跑构建并 [[家用部署机硬切重启]] launchd 看管的 Go server / worker；构建失败走 [[家用部署机 fail-open 构建契约]]，旧进程保持运行；rollback 完全手动。

走 a-1（bare repo 本地接 push） 而不是 GitHub webhook / cron polling / GitHub Actions SSH 的核心理由是**零外网依赖** —— 家庭网络在 NAT 后面，部署机大概率没有公网 IP；开发机与部署机同 LAN，直接 SSH push 是最短链路，push 即触发、可直接 tail 部署机 log 调试。GitHub 仍是 source of truth：post-receive 内部顺手 `git push --mirror github` 把每个 commit 同步到云端镜像，但部署不依赖云端可达性。

supervisor 选 launchd 而不是 Docker compose / `brew services` / tmux+nohup，理由是 launchd 是 macOS 原生 boot 启动 + crash 自动拉起 + 标准日志路径，不需要为 Go server / worker 引入 Dockerfile（项目当前没有），也不需要长期占着一个 tmux 会话。Postgres / Redis / 翻译服务 (`llama-server` 或外部端点) 归 [[家用部署机数据层独立生命周期]]，由 docker restart policy / launchd 各自管，**不跟随代码 push 重启**。

重启策略选硬切（kill old → start new，1–3 秒中断窗口）而不是蓝绿反代或 socket activation，是因为家庭用户 N≈2–5、push 多在深夜低峰期；为了零中断引入 nginx/Caddy 反代或改 `gin.Run()` 支持 socket inheritance，代价远超收益。客户端如需在重启窗口保持播放，由客户端侧自行重试，而非部署机做无缝接管 —— 这一边界写入 [[家用部署机硬切重启]]。

构建走 fail-open：post-receive 阶段 git ref 已经原子写入，构建/重启失败不再回滚 ref。go build / npm build 出错时旧进程继续跑，hook 把错误日志 tee 到 `deploy.log` 同时回流到 push 终端的 stderr，开发机推完立即看到。不允许把构建前置到 pre-receive 追"失败拒绝 push"，那会显著拖慢 push 体验并且破坏 [[家用部署机 fail-open 构建契约]]。

push 触发的动作按 [[家用部署机 push 分桶]] 规则筛：命中 `main.go` / `cmd/**` / `internal/**` / `pkg/**` / `go.mod` / `go.sum` / `migrations/**` 才走 Go server + worker 重建；命中 `admin-web/**`（不含 dist 自身）才走 `npm run build`；命中 `android-app/**` / `android-tv-app/**` / `docs/**` / `tasks/**` / `plan.md` / `CONTEXT.md` 等纯客户端或文档路径，hook 直接退出 0，不浪费时间编译。新增顶层目录时必须显式归桶，否则按"跳过"处理。

部署机的工作目录会随 push symlink-swap（macOS 不能覆写正在执行的 mach-O 二进制，必须 rename / symlink swap），任何运行期资产路径必须遵守 [[家用部署机绝对路径契约]]：`STORAGE_ROOT=/Volumes/large/ai-video-server/storage` 固定指向外挂盘，admin-web dist 走新增 `ADMIN_WEB_DIST_PATH` 环境变量。dev 模式仍允许相对路径默认值。

rollback 不做自动健康探活回滚（hook 启动后只 `curl /healthz` 探活一次确认存活，**探不到就报错走 fail-open**，不自动恢复旧 binary）。`scripts/rollback.sh <sha>` 提供手动快速回滚通道：切 symlink + 重启 launchd。保留最近 3 个二进制 (Q5d N=3) 支撑这条路径。**自动 rollback 不做**：健康探活 + 自动回滚要求每个 migration 必须 down 兼容并能被 hook 调用，复杂度远超家庭场景需要；与其追自动，不如把代码可回退的责任落在 [[migration 前向兼容契约]] 上，由 ADR-0006 单独约束。

## 考虑过的替代方案

- **GitHub webhook → 部署机 HTTPS 端口**：开发机 push GitHub，GitHub 调部署机的 webhook 触发 deploy。需要公网入站（端口转发 / Cloudflare Tunnel / frp），家庭路由器配置复杂；GitHub webhook 失败要去 GitHub 后台看 deliveries，而不是直接 tail 本地日志。失去了"push 即看日志"的体验。
- **Cron / launchd 轮询 `git fetch`**：部署机每分钟拉一次远端。零网络要求但延迟最差，且轮询窗口内推多个 commit 会被合并处理（也许是好事），但失去 push 即触发的即时性。如果将来确实想加云端 mirror 又想保留 push 即触发，可保留 a-1 主链路 + 加一层 cron 兜底，不互斥。
- **GitHub Actions → SSH 远程触发**：CI runner 跑构建，构建产物 scp 到部署机。runner 跑 build 比家庭 Mac 快也更隔离，但需要 SSH 入站或 Tailscale / 反向隧道；Actions cloud log 与部署机本地日志分离；任何 Actions 故障都阻塞 deploy。复杂度对家庭场景过剩。
- **Docker compose + restart policy 全面 Docker 化**：把 Go server / worker / admin-web 都封进 image，用 `restart: unless-stopped`。优点是供给链可重现、跨机迁移容易。缺点是需要新增 Dockerfile + 镜像 build 时间 + 容器内的 ffmpeg / LibVLC 兼容性问题；macOS 上 Docker Desktop 资源开销也比裸跑 Go binary 大。可作为未来跨机部署的迁移目标，但不在第一版。
- **蓝绿 + Caddy/nginx 反代**：起新版在 8081，健康通过后反代切端口，零中断。需要新增反代层、配置健康检查、配置 graceful drain；当前架构里没有反代。家庭场景的 1–3 秒中断不值这套复杂度。
- **launchd socket activation**：launchd 持有 8080 socket，旧进程退出时新进程接住。理论零中断。需要改 `internal/handlers/router.go` 的 `gin.Run()` 走 `net.FileListener` 接 socket inherit；macOS launchd 的 socket activation 文档少、踩坑面大；得不偿失。
- **pre-receive 阶段做构建以失败拒绝 push**：push 会阻塞到 build 完成（30s–2min）；任何 build error 都让你的 `git push` 终端报错。看似严谨实际反人类，开发机 push 频次会被这层 latency 显著降低。fail-open + 日志回流是更优工程权衡。
- **一键安装脚本** `scripts/install-deploy.sh`：把外挂盘检测、brew 装 Go/Node、初始化 bare repo、写 launchd plist、装 newsyslog 规则塞进一个 shell。Q8a 已驳回 —— 部署机一次性安装不值得为它写一份脚本，半失败时调试痛苦；纯文档 `docs/家用部署机.md` 配两个 helper 脚本是更现实的形态。

## 关联

- `docs/家用部署机.md` — 一次性安装步骤、日常运维命令、故障速查表
- `docs/adr/0006-migration-forward-compatibility.md` — 与本 ADR 是兄弟决策；本 ADR 的"手动 rollback only" 由 0006 的前向兼容契约支撑
- `scripts/migrate-apply.sh` — 从 `dev-up.sh` apply_migrations 抽出的独立脚本，hook 与 dev-up 共用
- `scripts/rollback.sh` — 手动 rollback 工具，依赖 [[家用部署机]] 目录布局已就位
- `CONTEXT.md` 新术语：[[家用部署机]] / [[dev 模式（dev-up.sh）]] / [[家用部署机硬切重启]] / [[家用部署机数据层独立生命周期]] / [[家用部署机 push 分桶]] / [[家用部署机 fail-open 构建契约]] / [[家用部署机重启可恢复边界]] / [[家用部署机绝对路径契约]]
- 后续待落实的代码改动（独立 task）：`internal/handlers/router.go` 增加 `ADMIN_WEB_DIST_PATH` 环境变量支持；`internal/config` 中相关媒体路径在部署模式下要求绝对路径
