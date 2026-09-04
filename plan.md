# plan.md

## 2026-09-04 16:11 +0800
- 进度：完成采集器常驻控制面接线。`-mode run` 通过运行时门控重建已授权 MTProto 客户端及其采集/订阅/Asynq 实例，不再创建 stdin 登录输入器；私有控制 HTTP、立即及 30 秒心跳独立常驻，未授权 session 时页面仍可查询并启动手机号或二维码授权。补充真实监听器集成测试，覆盖控制状态响应、初始心跳和取消后的优雅关闭。
- 影响文件：`cmd/telegram-ingestor/main.go`、新增 `cmd/telegram-ingestor/control_service.go`、`cmd/telegram-ingestor/heartbeat.go`、`cmd/telegram-ingestor/control_service_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./cmd/telegram-ingestor -count=1`、`GOPROXY=off go test -race ./cmd/telegram-ingestor -count=1`、`GOPROXY=off go vet ./cmd/telegram-ingestor`、`GOPROXY=off go build ./cmd/telegram-ingestor`、`git diff --check` 通过；`golangci-lint` 未安装。构建产生的临时根目录二进制已移入系统废纸篓，未纳入工作区或提交。

## 2026-09-04 16:05 +0800
- 进度：续接 Telegram 管理控制面。新增采集器内进程控制服务与心跳的定向测试已通过，下一步把授权服务、来源预览、运行时门控和仅内部可达的控制 HTTP 接入 `-mode run`；采集连接恢复后将重建消费服务与订阅，未授权状态仍保留页面管理能力。
- 影响文件：新增 `cmd/telegram-ingestor/control_service.go`、`cmd/telegram-ingestor/heartbeat.go`、`cmd/telegram-ingestor/control_service_test.go`；预计修改 `cmd/telegram-ingestor/main.go`、后续 API/管理端文件、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./cmd/telegram-ingestor -run 'Test(IngestorControlService|WriteTelegramHeartbeat)' -count=1` 通过；待继续运行采集器包测试、race、vet 和构建。

## 2026-09-04 15:56 +0800
- 进度：完成来源预解析确认票据层。控制协议的预览响应新增加入/审批标记且不再包含原始 `chat_ref`；`SourcePreviewService` 仅保存操作者、过期时间、Telegram 返回的展示结果和输入 SHA-256 指纹，确认时校验操作者/引用/类型/canonical ID，并且只有返回可保存 chat ID 后才消费票据。
- 影响文件：`internal/telegram/control.go`、`internal/telegram/source_preview.go`、`internal/telegram/source_preview_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、对应 `-race`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；待精确提交后将服务接入采集器控制 HTTP。

## 2026-09-04 15:52 +0800
- 进度：完成控制通道配置与采集器运行时门控子步骤。配置新增地址、API 访问 URL 和必填共享令牌，Compose 与示例环境同步注入；新运行时状态机在未授权后停驻等待页面授权恢复、在授权开始时取消并等待旧连接释放 session、恢复后以新客户端重连，控制操作仅在已授权运行态可用。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`cmd/telegram-ingestor/runtime.go`、`cmd/telegram-ingestor/runtime_test.go`、`.env.example`、`deploy/docker-compose.telegram.yml`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/config ./cmd/telegram-ingestor -count=1`、对应 `-race`、`GOPROXY=off go vet ./internal/config ./cmd/telegram-ingestor` 通过；待检查 Compose 解析、精确提交后实现控制服务与心跳入口接线。

## 2026-09-04 15:42 +0800
- 进度：提交 `8436e85` 后进入采集器常驻控制阶段。将增加 `TELEGRAM_CONTROL_ADDR`、`TELEGRAM_CONTROL_URL`、`TELEGRAM_CONTROL_TOKEN`，在独立运行时管理器中把未授权空闲、授权暂停、session 重载和已授权采集连接收敛为可测试状态；随后启动仅内部可达的控制 HTTP 并按 30 秒写入心跳。
- 影响文件：预计修改 `internal/config/`、`cmd/telegram-ingestor/`、`internal/telegram/control.go` 及测试，后续同步 Compose、`.env.example`、运维文档、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：先为配置与运行时状态机建立无网络红灯测试；阶段性运行 `GOPROXY=off go test ./internal/config ./cmd/telegram-ingestor -count=1`、对应 `-race`、`go vet` 与 `go build ./cmd/telegram-ingestor`。

## 2026-09-04 15:38 +0800
- 进度：完成第二阶段采集器运行时基础的客户端子步骤。`GotdClient.RunAuthorized` 仅验证现有 persistent session 的授权状态，未授权时返回可识别错误且不进入 stdin 登录；来源解析已拆为不入群的 `PreviewChat` 与确认后才可能执行邀请导入的 `ConfirmChat`，保留 `ResolveChat` 的既有兼容语义。私密邀请的审批加入请求不会自动入群或生成来源。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/gotd_client.go`、`internal/telegram/client_test.go`、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、`GOPROXY=off go test -race ./internal/telegram -count=1`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；待精确提交后进入常驻控制 HTTP、授权维护 gate 与心跳接线。

## 2026-09-04 15:32 +0800
- 进度：授权服务层已提交 `1cb50d0`，开始第二阶段采集器运行时基础：为 gotd 客户端增加仅使用现有 session 的无交互运行入口，并将来源引用拆成不入群的预解析和确认后入群两步；随后把该能力接入常驻控制 HTTP、授权维护 gate 和心跳。旧 `-mode run` 不再依赖 stdin，页面化授权成为正常路径。
- 影响文件：预计修改 `internal/telegram/client.go`、`internal/telegram/gotd_client.go` 及测试，后续修改 `cmd/telegram-ingestor/`、`internal/config/`、`.env.example`、Compose、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：先建立无网络的客户端状态/引用解析测试，阶段性运行 `GOPROXY=off go test ./internal/telegram`、`go vet` 和采集器构建；真实 session、私密邀请和在线运行仍在部署环境验证。

## 2026-09-04 15:26 +0800
- 进度：完成页面化 Telegram 授权服务的首个可提交层：单例授权服务将账号元数据、审计、管理员所有权和取消/过期清理与短期 session 分离；首次绑定仅允许手机号，二维码仅允许已有绑定且成功后校验 Telegram 用户 ID。新增 gotd 临时 session adapter，验证码、二次验证密码、二维码 token 均不进入仓储、审计或日志；临时文件在同目录经原子替换提升为正式 session。二维码图片过期独立于授权会话总时限，避免 token 刷新导致授权提前过期。
- 影响文件：新增 `internal/telegram/authorization_service.go`、`internal/telegram/authorization_service_test.go`、`internal/telegram/authorization_session_gotd.go`、`internal/telegram/authorization_session_gotd_test.go`；修改 `internal/telegram/auth_state.go`、`internal/telegram/auth_state_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：`GOPROXY=off go test ./internal/telegram -count=1`、`GOPROXY=off go test -race ./internal/telegram -count=1`、`GOPROXY=off go vet ./internal/telegram`、`git diff --check` 通过；二维码登录、手机号验证码和 session 切换仍需在具备真实 Telegram 凭据的部署环境验收。待提交本层，继续采集器常驻控制端口与运行时 gate。

## 2026-09-04 14:54 +0800
- 进度：已提交 Telegram 管理状态与控制协议基线 `ec96d66`。继续实现页面化授权运行时：采集器常驻内部控制端口、手机号验证码/二次验证、二维码短期图像、临时 session 校验和原子替换，以及 30 秒心跳；之后接入管理 API、来源预解析确认、失败恢复和 Vue 工作区。
- 影响文件：预计新增/修改 `internal/telegram/` 授权服务与 gotd adapter、`cmd/telegram-ingestor/`、`internal/config/`、`internal/handlers/`、`internal/services/`、`admin-web/src/`、`deploy/docker-compose.telegram.yml`、`.env.example`、ADR、`CONTEXT.md`、`plan.md`；继续排除用户既有 `docs/examples/`。
- 验证：每层先补定向单测，阶段性运行 Telegram 相关 Go 测试与 `go vet`；前端阶段运行 `npm run build`；收尾运行差异检查、乱码扫描。真实 Telegram 授权、二维码扫描、私密邀请和 Compose 端到端仍需具备凭据的部署环境验收。

## 2026-09-04 继续 +0800
- 进度：接续 Telegram 管理台控制平面实现。已确认上轮红灯测试进程退出，开始实现 `0039` 管理状态 schema、账号/心跳/授权/审计模型与仓储；随后接入授权状态机、内部控制 HTTP 和管理 API/页面。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`internal/models/telegram_management.go`、`internal/repository/telegram_management_repository.go`、`internal/telegram/auth_state.go`、`internal/telegram/control.go` 及后续管理服务、页面和部署配置；保留用户既有 `docs/examples/` 不纳入。
- 验证：已确认 `TestTelegramManagementMigration` 与 `TestAuthorizationStateMachine|TestControlHandler` 按预期红灯；实现后分阶段运行 Go 定向测试、`go vet`、管理端构建、差异检查和乱码扫描。真实 Telegram 凭据与 Docker 端到端验收仍待部署环境。

## 2026-09-03 13:05 +0800
- 进度：开始实现 Telegram 管理台控制平面。第一阶段锁定 `0039` 账号状态、采集器心跳、短期授权元数据和追加式审计表，并建立内部控制协议、常量时间 token 校验、敏感字段脱敏和单例授权状态机的红灯测试；不改变既有 `0038` 来源/媒体表及采集队列语义。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`internal/models/telegram_management.go`、`internal/repository/telegram_management_repository.go`、`internal/telegram/control.go`、`internal/telegram/auth_state.go` 及对应测试；保留用户既有 `docs/examples/` 不纳入。
- 验证：先运行新增定向测试确认因实现缺失而失败；随后按阶段运行 Go 定向测试、`go vet`、管理端构建和差异/乱码检查。真实 Telegram 手机号、二维码、私密邀请和重启恢复继续留待具备凭据与 Docker 的部署机验收。

## 2026-09-03 12:20 +0800
- 进度：用户确认采用 Telegram 管理台全部推荐方案，开始进入实现。新增能力将以 `telegram-ingestor` 独占 session、API 仅经 Docker 内网控制通道转发为边界；引入前向兼容的账号状态/采集器心跳/审计 migration，保留现有来源 API 与队列语义。首次以手机号验证码绑定，二维码仅用于已绑定账号重新授权；高风险操作使用当前管理员密码确认与一次性短期凭据；来源采用预解析后确认、canonical chat ID 去重和来源级失败恢复。
- 影响文件：预计新增 `migrations/0039_telegram_management.*.sql`、`docs/adr/0026-telegram-ingestor-control-plane.md`、Telegram 控制/管理 service 与测试；修改 `internal/telegram/`、`cmd/telegram-ingestor/`、`internal/handlers/`、`internal/repository/`、`internal/config/`、`main.go`、`deploy/docker-compose.telegram.yml`、`admin-web/src/`、`.env.example`、`CONTEXT.md`、`plan.md`。保留用户既有 `docs/examples/` 不纳入。
- 验证：先建立 migration、控制协议、权限/脱敏和页面轮询的红灯测试；实现后运行相关 Go 测试、`go vet`、管理端 `npm run build` 与差异/乱码检查。真实手机号、二维码、私密群组、回填及重启恢复仅能在具备 Telegram 凭据与 Docker 的部署环境验收。

## 2026-09-03 11:57 +0800
- 进度：`grill-with-docs` 第五轮全部采用推荐方案。沿用现有 `admin` 权限；首次绑定、重新授权和私密邀请解析需要管理员密码确认及页面二次确认；授权会话全局单例且发起者独占秘密提交；二维码只返回短期图像；重新授权有限排空现有下载/导入；账号、采集器、来源状态分层；首期限定群组/超级群组/频道、预解析短期有效、无自定义别名；审计保留一年；普通状态 5 秒、授权状态 2 秒轮询且不重入。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅记录领域术语和设计共识，未修改运行时代码、接口或管理端页面。
- 验证：待最后一轮确认页面字段、接口响应、危险操作文案和端到端验收清单；用户确认共享理解前不进入实现。

> 2026-08-11 压缩整理版：按用户要求删除纯部署、推送、重启、健康检查和镜像同步流水；将同一事项的访谈、开始、红灯、实现、复核、待提交等过程记录合并为最终有效结论。历史精确差异与验证细节以 Git 提交、`CONTEXT.md`、ADR 和 `tasks/*/DONE.md` 为准。后续仍按反向时间顺序在顶部追加计划与进度。

## 2026-09-03 11:36 +0800
- 进度：`grill-with-docs` Telegram 管理台设计访谈完成前三轮。用户确认采用完整页面化管理、单个系统 Telegram 采集账号、固定导入归属用户、添加来源前确认并自动历史回填、暂停/恢复/重新回填、失败恢复和轮询状态；新增二维码授权，但首次绑定仍要求手机号验证码，二维码只用于重新授权或恢复。页面不编辑部署级 API ID/hash、手机号或服务生命周期；验证码和二次验证密码属于短期授权输入，不持久化。二维码重新授权必须校验首次绑定的 Telegram 用户 ID；新 session 仅在校验成功后替换旧 session。来源先预解析再确认加入、按 canonical chat ID 去重，采集器运行状态通过心跳观测，页面不启动容器。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅记录领域术语与设计共识，未修改运行时代码、接口或管理端页面。
- 验证：已核对现有 Vue 管理端、管理员认证、Telegram gotd 客户端和二维码 API 能力；待继续完成登录状态机、来源生命周期、错误恢复和页面验收边界访谈，用户确认共享理解前不进入实现。

## 2026-09-03 11:52 +0800
- 进度：`grill-with-docs` 第四轮完成。用户确认 API 通过仅内部可达的采集器控制通道转发授权交互；新增单例 Telegram 账号状态元数据与追加式管理审计；私密邀请 token 解析后丢弃；采集器以 30 秒心跳、90 秒超时提供结构化运行状态；管理台 PC 优先并保证窄视口无溢出，不展示原始采集器日志。
- 影响文件：`CONTEXT.md`、`plan.md`；当前仅补充领域语言和设计进度，未修改运行时代码、接口或管理端页面。
- 验证：已核对现有 Compose session 挂载边界、Telegram 客户端 QR 能力和管理端路由/日志现状；待继续确认控制通道认证、管理员权限、审计保留和异常恢复细节，用户确认共享理解前不进入实现。

## 2026-09-02 21:47 +0800
- 进度：Task 7 自动实现与仓库内验收完成。部署契约覆盖统一镜像、三服务共享路径、session 隔离、非 root 运行、容器服务名连接、固定容器端口、敏感构建上下文排除和手册必备操作；最终包级测试还发现并修复 Task 1 测试行扫描器缺少 Go `int` 目标支持的问题，生产仓储逻辑未改。
- 影响文件：`.dockerignore`、`.gitignore`、`.env.example`、`deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`、`CONTEXT.md`、`internal/repository/telegram_repository_test.go`、`plan.md`；提交时明确排除用户既有 `docs/examples/`。
- 验证：`go test ./deploy ./internal/repository -count=1`、计划包集与全仓测试在跳过既有 `TestParseTVAPKMetadataParsesReleaseAPK` 后通过；`go vet ./...`、两个 `CGO_ENABLED=0 GOOS=linux go build`、`git diff --check`、U+FFFD 扫描通过。未跳过时计划包集和 `go test ./... -count=1` 仅因既有 TV APK 实际 version code `144`、断言仍为 `121` 失败。当前机器没有 `docker` 命令，故 `docker compose ... config`、镜像实际构建和测试群组登录/回填/重启/实时采集需在 OrbStack 部署机按运行手册执行，未冒充已验证。

## 2026-09-02 18:41 +0800
- 进度：Task 7 核心实现完成。新增 Go 1.22 多阶段镜像和三服务 Compose，统一共享媒体/暂存路径并隔离个人 session；Compose 同步透传既有 API/worker 的外部服务配置，避免容器部署丢失原有能力；补齐中文运维手册和 Telegram 长期架构约定。Dockerfile 只复制 Go 构建必需源码，`.dockerignore` 进一步阻止本地 `.env`、session 和媒体目录进入构建上下文。
- 影响文件：`.dockerignore`、`.gitignore`、`deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`、`.env.example`、`CONTEXT.md`、`internal/repository/telegram_repository_test.go`、`plan.md`；保留用户既有 `docs/examples/` 未跟踪内容。
- 验证：`go test ./deploy -count=1`、`git diff --check` 通过，目标文件 U+FFFD 扫描无输出；本机缺少 `docker` 命令，Compose CLI 解析和真实 Telegram 测试待最终验收时如实记录。

## 2026-09-02 18:27 +0800
- 进度：Task 7 部署契约红灯已建立，覆盖三服务统一镜像/共享挂载、session 仅由非 root 采集器挂载、`ffmpeg`/`ffprobe`、采集器 `-mode run` 和 Telegram 环境变量完整性。
- 影响文件：`deploy/telegram_deploy_test.go`、`plan.md`。
- 验证：`go test ./deploy -count=1` 按预期失败，原因仅为 `deploy/docker-compose.telegram.yml`、`deploy/video-server.Dockerfile` 和三个宿主挂载变量尚未实现。

## 2026-09-02 18:24 +0800
- 进度：进入 Task 7。开始建立 OrbStack 部署契约测试，并补齐统一应用镜像、三服务 Compose、Telegram 一次性登录与日常运维手册；三个应用服务共享媒体和上传暂存容器路径，个人 session 仅由非 root 采集器持有。
- 影响文件：预计新增 `deploy/video-server.Dockerfile`、`deploy/docker-compose.telegram.yml`、`deploy/telegram_deploy_test.go`、`docs/telegram-video-ingestion.md`，修改 `.env.example`、`CONTEXT.md`、`plan.md`；保留用户既有 `docs/examples/`，不修改 Android 工程。
- 验证：先运行 `go test ./deploy -count=1` 确认部署契约红灯；实现后执行 Task 7 定向测试、全量 Go 测试、`go vet ./...`、Compose 配置解析、差异及乱码检查。

## 2026-09-02 18:14 +0800
- 进度：Task 6 完成。管理员接口已覆盖来源列表、新增、暂停、恢复、重新回填和进度查询；新增来源统一保存规范化 `chat_ref`，恢复保留游标，重新回填清零游标，控制任务异步投递。
- 影响文件：`internal/services/telegram_source.go`、`internal/services/telegram_source_test.go`、`internal/handlers/admin_telegram.go`、`internal/handlers/admin_telegram_test.go`、`internal/handlers/router.go`、`main.go`、`plan.md`；未纳入用户既有 `docs/examples/`。
- 验证：Task 6 定向测试、handler 全量测试、`go vet` 和主程序构建通过；待提交 `增加 Telegram 群组管理接口`。

## 2026-09-02 18:02 +0800
- 进度：进入 Task 6。按已批准计划实现 Telegram 来源管理服务和管理员 API；HTTP 请求只更新 PostgreSQL 状态并投递 control 任务，群组解析与历史扫描继续由独立采集器执行。
- 影响文件：`internal/services/telegram_source.go`、`internal/services/telegram_source_test.go`、`internal/handlers/admin_telegram.go`、`internal/handlers/admin_telegram_test.go`、`internal/handlers/router.go`、`main.go`、`plan.md`；保留用户既有 `docs/examples/`。
- 验证：`go test ./internal/services ./internal/handlers -run 'TestAdminTelegram|TestTelegramSource' -count=1` 已按预期红灯，待实现后复跑定向测试、`go vet` 和主入口构建。

## 2026-09-02 17:51 +0800
- 进度：Task 5 完成。补齐暂停来源的 claim 前检查、超时下载/导入记录恢复、下载并发限制、转码与下载双重 reconcile，并新增独立 `telegram-ingestor` 的登录/运行入口；实时与回填队列按 `10:1` 权重消费。
- 影响文件：`internal/services/telegram_ingestion.go`、`internal/services/telegram_ingestion_test.go`、`internal/services/upload.go`、`internal/repository/telegram_repository.go`、`internal/queue/telegram_processor.go`、`internal/queue/telegram_processor_test.go`、`cmd/telegram-ingestor/main.go`、`cmd/telegram-ingestor/main_test.go`、`plan.md`；保留用户既有 `docs/examples/` 未跟踪内容。
- 验证：`go test ./internal/services ./internal/queue -run 'TestTelegram(Ingestion|Processor)' -count=1`、`go test ./cmd/telegram-ingestor ./internal/services ./internal/queue ./internal/telegram -run 'TestTelegram' -count=1`、`go vet ./internal/services ./internal/queue ./internal/telegram ./cmd/telegram-ingestor`、`go build ./cmd/telegram-ingestor` 通过；Task 5 包级测试仅因既有 `TestParseTVAPKMetadataParsesReleaseAPK` 断言 `121` 与当前 APK 实际版本 `144` 不一致而失败。
- 提交：待提交 `实现 Telegram 视频采集与回填`，随后进入 Task 6 管理 API。

## 2026-09-02 16:31 +0800
- 进度：进入 Task 5。开始建立 Telegram 来源同步、下载处理、三层幂等和转码补偿的服务/队列处理器边界；独立采集器入口继续保持与现有 `server`/`worker` 入口隔离。
- 影响文件：预计新增 `internal/services/telegram_ingestion.go`、`internal/services/telegram_ingestion_test.go`、`internal/queue/telegram_processor.go`、`internal/queue/telegram_processor_test.go`、`cmd/telegram-ingestor/main.go`，必要时最小修改 `main.go` 与本记录；保留用户既有 `docs/examples/`，不修改 Android 工程。
- 验证：先运行 `go test ./internal/services ./internal/queue -run 'TestTelegram(Ingestion|Processor)' -count=1` 验证红灯，再按 Task 5 完成定向测试、`go vet` 和入口构建。

## 2026-09-02 16:18 +0800
- 进度：Task 4 完成。固定 `github.com/gotd/td v0.99.2`，封装个人账号 session、chat 解析、历史分页、实时订阅、消息刷新、document 流式下载和 FloodWait 错误；元数据构造规则已覆盖 caption 标题与文件名回退。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/client_test.go`、`internal/telegram/gotd_client.go`、`internal/services/telegram_metadata.go`、`internal/services/telegram_metadata_test.go`、`go.mod`、`go.sum`、`plan.md`；未纳入用户既有 `docs/examples/`。
- 验证：`go test ./internal/telegram ./internal/services -count=1`、`go vet ./internal/telegram ./internal/services`、`gofmt -d`、`git diff --check` 通过；准备精确提交 Task 4。

## 2026-09-02 16:00 +0800
- 进度：进入 Task 4。接口、FloodWait 包装类型和 Telegram 元数据测试已建立；待固定 `github.com/gotd/td` 稳定版本并实现个人账号 session、群组解析、历史分页、实时消息订阅、消息刷新和 document 流式下载适配器。
- 影响文件：`internal/telegram/client.go`、`internal/telegram/client_test.go`、`internal/telegram/gotd_client.go`、`internal/services/telegram_metadata.go`、`internal/services/telegram_metadata_test.go`、`go.mod`、`go.sum`、`plan.md`；保留 `docs/examples/` 未跟踪内容，不修改 Android 工程。
- 验证：现有 `go test ./internal/telegram ./internal/services -run 'Test(Client|BuildTelegramVideoMetadata)' -count=1` 待适配器完成后复跑；随后执行 `go vet`、竞态定向测试和提交前差异/乱码检查。

## 2026-09-02 15:47 +0800
- 进度：Task 3 完成。抽取 `UploadService` 的本地文件保存核心；手动上传保留重复时刷新可替换原片路径的既有语义，`SaveImportedFile` 在重复内容时不写入临时导入路径，并覆盖 hash 竞态分支。
- 影响文件：`internal/services/upload.go`、`internal/services/upload_test.go`、`internal/services/import_file.go`、`internal/services/import_file_test.go`、`plan.md`。
- 验证：`go test ./internal/services -run 'TestSaveImportedFile|TestShouldUpdateExistingVideoOriginalPath|TestShouldRefreshExistingVideoOriginalPath|TestPredictedUploadStatus' -count=1`、`go vet ./internal/services`、`git diff --check` 通过；完整 `go test ./internal/services -count=1` 仅受既有 TV APK 测试版本断言 `144 != 121` 失败，未涉及本次代码。

## 2026-09-02 15:41 +0800
- 进度：Task 2 完成。新增 Telegram API ID/hash、电话、session、导入归属用户、实时/历史/控制队列及下载限流配置；`Load` 保持 Telegram 配置可选，`ValidateTelegramIngestor` 单独校验采集器凭据、UUID 和并发边界；新增仅携带 UUID 的 Asynq 采集任务和默认队列。
- 影响文件：`internal/config/config.go`、`internal/config/config_test.go`、`internal/queue/telegram_tasks.go`、`internal/queue/telegram_tasks_test.go`、`.env.example`、`plan.md`。
- 验证：`go test ./internal/config ./internal/queue -run 'Test(LoadIncludesTelegram|LoadDefaultsTelegram|ValidateTelegram|TelegramTask|NewTelegram)' -count=1`、`go test ./internal/config ./internal/queue -count=1`、`go vet ./internal/config ./internal/queue`、`git diff --check` 均通过；待提交 `增加 Telegram 配置与采集队列`。

## 2026-09-02 15:24 +0800
- 进度：根据 `docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md` 开始实现 Telegram 指定群组视频采集、历史回填和管理 API。当前进入 Task 1，先建立可回滚 migration、来源/媒体模型和仓储边界；保持个人 MTProto session 独占、短视频导入、三层幂等及队列/数据库状态分工。
- 影响文件：预计涉及 `migrations/0038_*`、`internal/models/telegram.go`、`internal/repository/telegram_*`、`internal/config`、`internal/queue`、`internal/services`、`internal/telegram`、`internal/handlers`、`cmd/telegram-ingestor`、`deploy`、`docs/telegram-video-ingestion.md`、`CONTEXT.md`、`plan.md`；保留既有 `docs/examples/` 和其它工作区差异，不修改 Android 工程。
- 验证：待按 Task 1 至 Task 7 逐项执行定向测试、`go vet`、构建、迁移/Compose 静态检查及最终全量验证；每个任务完成后独立使用中文提交。

## 2026-09-02 15:32 +0800
- 进度：Task 1 完成。新增 `0038` Telegram 来源/媒体 migration 及可回滚 down migration，建立来源、消息、document ID、SHA-256、导入/转码状态、重试字段和索引；新增模型与 VideoRepository 仓储边界，支持消息幂等登记、同事务游标推进、行锁 claim、部分更新、转码补偿查询和状态统计。
- 影响文件：`migrations/0038_telegram_video_ingestion.up.sql`、`migrations/0038_telegram_video_ingestion.down.sql`、`internal/models/telegram.go`、`internal/repository/telegram_repository.go`、`internal/repository/telegram_repository_test.go`、`internal/repository/migrations_test.go`、`plan.md`。
- 验证：`go test ./internal/repository -run 'TestTelegram|Test.*Migration' -count=1`、`go test ./internal/repository -count=1`、`go vet ./internal/repository`、`git diff --check` 均通过；待提交 `新增 Telegram 采集状态模型`。

## 2026-09-02
- 进度：完成 Telegram 指定群组视频采集与全部历史回填的设计确认，生成实现计划 `docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md`。方案固定多群组数据库来源、实时/历史独立下载队列、`(source_id,message_id)` + Telegram document ID + `SHA-256/文件大小` 三层幂等、短视频导入及转码投递补偿；尚未修改业务代码。
- 影响文件：`docs/superpowers/specs/2026-09-02-telegram-video-ingestion-design.md`、`docs/superpowers/plans/2026-09-02-telegram-video-ingestion.md`、`plan.md`；不修改 Android/TV 版本号。设计提交为 `b187a8e`，实现计划待确认执行方式。
- 验证：设计文档自检通过；计划完成占位词、接口一致性和 `git diff --check` 检查，待用户选择按任务执行方式后进入实现。

## 2026-08-14 09:25 +0800
- 进度：论坛资源改动完成提交前复核，代码、测试、接口契约、ADR 与长期上下文一致；提交仅纳入本任务文件及本任务 `plan.md` 记录，保留既有历史重抓记录和 `docs/examples/` 工作区内容。
- 影响文件：本任务文件同 09:24 记录；不修改 Android/TV 版本号。
- 验证：`git diff --check`、Go 格式检查及本任务中文文件 U+FFFD 扫描通过，暂存后继续核对 staged diff 并提交。

## 2026-08-14 09:24 +0800
- 进度：论坛资源标题搜索、受限帖子展示及保留规则实现完成。管理员列表接口新增最多 200 字符的标题字面子串查询，投影纳入并标记 `restricted + included`；有效 Hermes `discover` 触发的 30 天清理现在保护当前 `pending` 与 `restricted`，其余无资源旧帖规则不变。管理端采用独立搜索草稿和显式提交，刷新/翻页保留已提交查询，清空恢复完整列表；长期契约已同步至 `CONTEXT.md` 与 ADR-0025。
- 影响文件：`internal/models/hermes_forum.go`、`internal/handlers/hermes_forum.go`、`internal/handlers/hermes_forum_test.go`、`internal/services/hermes_forum.go`、`internal/services/hermes_forum_test.go`、`internal/repository/hermes_forum_repository.go`、`internal/repository/hermes_forum_repository_test.go`、`internal/repository/hermes_forum_repository_integration_test.go`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList.vue`、`admin-web/src/views/ForumPostList.spec.js`、`CONTEXT.md`、`docs/adr/0021-admin-forum-resource-list.md`、`docs/adr/0023-hermes-resource-aware-retention.md`、`docs/adr/0025-admin-forum-restricted-search-retention.md`、`plan.md`；不修改 Android/TV 版本号，不纳入既有 `docs/examples/` 与无关 `plan.md` 记录。
- 验证：论坛定向 Go 测试与对应 `-race` 通过，`go vet ./...` 通过；管理端定向 36 项、全量 705 项测试及 `npm run build` 通过；浏览器桌面与 390px 视口验证搜索提交、空态、清空恢复、受限标记、容器滚动及控制台均通过。PostgreSQL 集成测试因未设置 `HERMES_TEST_DATABASE_URL` 按既有机制跳过，`golangci-lint` 本机未安装；`go test ./... -count=1` 的无关 TV APK 固件版本 144/旧断言 121 失败，整包 `-race` 的无关并行 Gin `SetMode` 基线竞态仍存在。待执行提交前差异、乱码与暂存范围检查。

## 2026-08-14 09:12 +0800
- 进度：论坛资源标题搜索与受限帖子展示/保留红灯已建立。Go 定向测试因旧代码缺少搜索参数、受限状态字段、查询构造及超长错误类型而按预期编译失败；管理端定向测试 36 项中 4 项按预期失败，覆盖搜索控件与提交状态、受限标记和搜索空态，API 请求包装既有测试保持通过。
- 影响文件：`internal/handlers/hermes_forum_test.go`、`internal/services/hermes_forum_test.go`、`internal/repository/hermes_forum_repository_test.go`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList.spec.js`、`plan.md`。
- 验证：红灯命令为 `go test ./internal/handlers ./internal/services ./internal/repository -run 'TestAdminForumPosts|TestHermesForumService(ListsAdminReadModel|RejectsLongAdminSearchQuery)|TestBuildAdminForumPostListSQL|TestDiscoverForumPostsCleanup' -count=1` 与 `npm test -- --run src/api/admin.spec.js src/views/ForumPostList.spec.js`；待实现后复跑转绿。

## 2026-08-14 09:08 +0800
- 进度：用户确认 PC 管理端论坛资源完整方案，进入实现。列表新增服务端标题字面子串搜索并展示 `restricted + included`，标题旁标记“受限”；当前为 `pending` 或 `restricted` 的记录不再参与 30 天无资源清理，豁免仅跟随当前状态，不恢复已删除历史。先补 Handler、Service、Repository 与 Vue 页面红灯测试，再做最小实现。
- 影响文件：预计涉及 `internal/models/hermes_forum.go`、`internal/handlers/hermes_forum*`、`internal/services/hermes_forum*`、`internal/repository/hermes_forum*`、`admin-web/src/api/admin.spec.js`、`admin-web/src/views/ForumPostList*`、`CONTEXT.md`、`docs/adr/0021-admin-forum-resource-list.md`、`docs/adr/0023-hermes-resource-aware-retention.md`、`docs/adr/0025-admin-forum-restricted-search-retention.md`、`plan.md`；不修改 Android/TV 版本号。
- 验证：待执行后端定向红灯与回归、`go test ./... -count=1`、`go test -race` 定向包、`go vet ./...`、管理端定向与全量 Vitest、`npm run build`、页面桌面/窄视口检查、`git diff --check` 与中文乱码扫描。

## 2026-08-13 17:24 +0800
- 进度：中国行政区划地图生产解析故障修复完成。Go 管理端静态服务现在于 SPA 回退前原样提供 `/admin/china-map/**` 下的 `.json` 与 `.geojson`，设置正确 MIME 和每次复验缓存策略；缺失、越界或非允许扩展统一 404，不再返回 `index.html`。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`CONTEXT.md`、`plan.md`；无关 Hermes 记录与 `docs/examples/115LocalNatManager/` 不纳入提交。
- 验证：静态路由红灯已复现并转绿；`go test ./internal/handlers -count=1`、定向 `-race`、`go vet ./...`、`go build ./...`、管理端 46 个文件 704/704、`npm run build`、`npm run china-map:validate`、差异/乱码/控制字符检查通过。`go test ./... -count=1` 仅有既有 TV release APK 元数据漂移失败：产物 `versionCode=144`，测试仍期望 `121`，未修改无关 TV 产物或测试。待提交、推送部署机并按生产 JSON/GeoJSON 响应和浏览器地图加载复验。

## 2026-08-13 17:19 +0800
- 进度：修复中国行政区划地图部署后数据解析失败。生产取证确认 `/admin/china-map/catalog.json` 与 `manifest.json` 被 Go 管理端静态服务错误回退为 `index.html`；先补后端红灯测试，再为地图 JSON/GeoJSON 增加受路径边界保护的静态文件服务，缺失文件保持 404，不得伪装成 SPA 页面。
- 影响文件：`internal/handlers/router.go`、`internal/handlers/admin_static_test.go`、`CONTEXT.md`、`plan.md`。
- 验证：待执行后端定向测试、全量 Go 测试与 `go vet ./...`、管理端构建/地图校验、生产响应 Content-Type/正文及浏览器地图加载复验；完成后提交并推送部署机。

## 2026-08-13 15:32 +0800
- 进度：中国行政区划地图完成提交前收口。启动数据明确为目录与清单必需、南海诸岛附图可选，附图加载失败不阻塞主地图；地图信息弹层补齐打开后聚焦、Tab 焦点约束、Esc 关闭和触发按钮焦点恢复。无关 Hermes 记录与 `docs/examples/115LocalNatManager/` 保留在工作区，不纳入本次提交。
- 影响文件：`admin-web/src/views/ToolboxChinaMap.vue`、`admin-web/src/views/chinaMap.data.js`、对应测试、地图功能既有代码与快照、`CONTEXT.md`、`plan.md`。
- 验证：定向回归 19/19、全量 Vitest 46 个文件 704/704、`npm run build`、`npm run china-map:validate` 和浏览器控制台检查通过；1024×768 实测画布与视口同为 1024×768、画布像素非空、悬浮控件无重叠，弹层焦点闭环正常；模拟 `CN-63` 图层首次网络失败时保留全国地图，点击重试后恢复“中国 / 青海省”。构建仅有项目既有的大包体积警告。

## 2026-08-13 15:22 +0800
- 进度：中国行政区划地图完成代码、数据和浏览器收尾。新页面作为 `/toolbox/china-map` 管理员认证独立工作区接入工具箱，画布覆盖完整视口；支持全国、省、地、县可变层级钻取、末级高亮、全局搜索、URL/历史恢复、缩放复位、错误重试、数据说明及不可交互南海诸岛附图。正式提交固定 OSM 时间点的 34 个省级范围与港澳台分层快照，并补齐生成、完整性校验、语义色门禁、首屏返回状态和窄屏按钮可访问性。
- 影响文件：`admin-web/package*.json`、`admin-web/src/App.vue`、`admin-web/src/assets/theme.css`、`admin-web/src/router/*`、`admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxChinaMap.vue`、地图 helper/loader 与测试、`admin-web/scripts/china-map/*`、`admin-web/public/china-map/*`、`docs/adr/0024-admin-china-administrative-map-data.md`、`CONTEXT.md`、`plan.md`。
- 验证：定向 Precision Ops 门禁 227/227、地图与工具箱回归 20/20、全量 Vitest 702/702、`npm run build`、`npm run china-map:validate`、`git diff --check`、U+FFFD 与控制字符扫描均通过；数据门禁为 3631 个区域、34 个省级范围、3245 个末级区域、386 个图层、3630 个面要素。Chrome 验收覆盖 1440×900 与 375×812：画布非空且等于视口，控件无重叠，搜索“乌丘”可定位并恢复完整路径，URL/浏览器后退正常，控制台无报错；无关 Hermes 记录与 `docs/examples/` 不纳入本次提交。最终全量复验与独立审阅待提交前完成。

## 2026-08-13 14:57 +0800
- 进度：中国行政区划地图主体与正式数据已完成，收尾复现全量测试剩余的 5 个 Precision Ops 门禁失败：新视图尚未登记到独立页面边界、工具箱数量仍锁定为 5，以及地图视图存在直接颜色、3 处常驻面板阴影和 2 处直接百分比圆角。按现有门禁修复页面和接入清单，不放宽审计规则。
- 影响文件：`admin-web/src/assets/theme.css`、`admin-web/src/views/ToolboxChinaMap.vue`、`admin-web/src/views/precisionOpsRollout.spec.js`、`plan.md`。
- 验证：定向 `npm test -- --run src/views/precisionOpsAudit.spec.js src/views/precisionOpsRollout.spec.js` 已按预期出现 5 项红灯；待重跑定向与全量测试、管理端构建、地图数据校验和浏览器多视口验收。

## 2026-08-13 10:18 +0800
- 进度：用户确认中国行政区划地图完整方案，进入管理端实现。采用现有 Vue 3、Vue Router、ECharts 与 Element Plus；新增 `/toolbox/china-map` 认证独立路由和工具箱新标签页入口，页面保持无 shell 且地图占满视口。实现按静态索引与父级拆分 GeoJSON 契约分离层级、搜索、URL 恢复、缓存和错误处理，并先补红灯测试；正式数据只接受许可与来源可追溯的快照，不把临时无许可证样本纳入仓库。
- 影响文件：预计涉及 `admin-web/src/views/Toolbox.vue`、`admin-web/src/views/ToolboxChinaMap.vue`、地图 helper/测试、`admin-web/src/router/*`、`admin-web/public/china-map/*`、数据构建脚本与说明、`CONTEXT.md`、ADR-0024、`plan.md`。
- 验证：待执行管理端定向 Vitest、全量 `npm test`、`npm run build`、数据完整性校验、`git diff --check`、中文乱码扫描，以及 1024/1440/窄视口浏览器截图核验。

## 2026-08-13 09:42 +0800
- 进度：开始处理 Hermes 权限误判修复后的历史重抓。目标为 `source=sehuatang` 且 `inspection_status=restricted` 的最终帖子；先只读统计并按旧权限误判、已有资源和真实登录/验证受限分类，再备份数据库行、资源行、Hermes 任务定义和相关状态，暂停新帖与历史补漏调度后逐条执行 `discover -> cdp_fetch -> enrich_thread -> inspection`。仅新发现资源的帖子允许产生通知，避免重复投递。
- 影响文件：`plan.md`；运行时目标为远程 Hermes Forum API、PostgreSQL、`~/.hermes/cron/jobs.json` 与 `~/.hermes/state/`，不修改 Android/TV 版本号。
- 验证：待完成 restricted 统计、备份完整性、重抓结果、通知队列清理及调度恢复核对。

## 2026-08-13 09:50 +0800
- 进度：按用户确认改用状态机恢复方案：将 `source=sehuatang` 的 150 条 `restricted` 终态重置为 `pending`，保留帖子身份和观察时间，清空旧检查结论与指纹；恢复两个论坛 cron，让其按既有 `discover -> inspection` 流程重抓。由于历史补漏游标可能已越过旧页，重置后核对并在必要时把明确候选加入持久队列。
- 影响文件：`plan.md`；运行时目标为远程 `collected_forum_posts` / `collected_forum_post_resources`、Hermes 任务定义和补漏队列。
- 验证：待执行事务行数与约束核对、cron 入队/检查进度、资源结果、重复通知防护及任务恢复状态检查。

## 2026-08-13 09:31 +0800
- 进度：修复 Hermes 将附件提示“阅读权限: 10”误判为整页权限门禁的问题。外部活动脚本 `~/.hermes/scripts/watch_sehuatang_forum95.py` 仅移除 `is_blocked_or_permission_text()` 中的裸“阅读权限”词项，保留明确的“需要阅读权限”“阅读权限高于”“本帖隐藏的内容需要”及登录/验证门禁；已生成可回滚备份 `~/.hermes/scripts/watch_sehuatang_forum95.py.pre-attachment-permission-fix-20260813-0920`。
- 影响文件：外部 Hermes 脚本及其回滚备份；仓库仅追加 `CONTEXT.md`、`docs/adr/0020-hermes-forum-post-ingestion.md`、`plan.md`，既有 `docs/adr/0024-admin-china-administrative-map-data.md` 与 `docs/examples/` 不纳入。
- 验证：临时回归 `/tmp/hermes_attachment_regression_test.py` 修改前按预期红灯，修改后通过；Hermes venv `py_compile` 通过；真实 CDP 页面和 `/tmp/hermes-2526602-cdp.json` 均确认正文含“阅读权限: 10”时 `cdp_fetch()` 为 `ok=true`，`extract_attachments()` 返回 `mod=attachment` 链接；09:21 新帖任务保持 `last_status=ok`。只读数据库核对显示 `tid=2526602` 当前历史终态仍为 `restricted/included`、附件数 0；根据检查结果不可覆盖约束，本次不直接改写记录、不补发通知，待明确授权后另行走可回滚修复流程。

## 2026-08-11 16:42 +0800
- 进度：`grill-with-docs` 数据治理轮全部采用推荐方案。行政边界以固定版本 OpenStreetMap 快照为几何来源，结合可追溯公开中文行政区目录离线规范化并生成分层 GeoJSON；生成数据遵守 ODbL 归属和相同方式共享要求，页面运行时不访问第三方地图服务。区域统一使用仓库稳定 `region_code` 作为“区划标识”，来源行政编码另存且允许为空。数据构建对完整省级范围与港澳台、父子闭合、标识唯一、中文名、有效几何和末级可达实行失败即停止门禁；快照只允许人工触发、审核差异后随管理端版本发布。新增 ADR-0024 固化该难以逆转的数据选择。
- 影响文件：`CONTEXT.md`、`docs/adr/0024-admin-china-administrative-map-data.md`、`plan.md`；生产代码仍未修改。
- 验证：待用户确认完整共同理解后执行文档差异、乱码与术语一致性检查并提交；确认前不进入实现。

## 2026-08-11 15:53 +0800
- 进度：`grill-with-docs` 第三轮全部采用推荐方案。静态行政区划按父级编码拆分并逐层懒加载；切层时保留旧图和轻量加载反馈，数据失败保留上一层并提供重试/返回；县级浮层只显示名称、完整路径和编码；搜索支持即时候选、键盘选择、回车定位和无结果反馈；全国视图包含不参与钻取的南海诸岛附图；层级动画为 200–260ms 且尊重减弱动效；数据来源、版本和许可放入按需地图信息浮层；缓存仅使用静态资源 HTTP 缓存与会话内存。
- 影响文件：`CONTEXT.md`、`plan.md`；生产代码仍未修改。
- 验证：数据源核查确认现成数据均有关键缺口：无许可证、离线再分发条款不明确，或虽开放许可但中国县级数据代表年份为 2017、英文名称且无行政区编码。需要最后确认采用可追溯数据构建方案及编码口径。

## 2026-08-11 15:37 +0800
- 进度：`grill-with-docs` 第二轮全部采用推荐方案。县级点击仅高亮并展示完整行政路径；直辖市、港澳台等按真实父子关系使用可变层级；面包屑可跳祖先、地图返回逐级回退、浏览器历史与 URL 行政区编码同步，退出工具箱独立；全局搜索覆盖省/地/县并以完整路径消歧；标签只显示当前层级且碰撞时隐藏、悬停可查；地图沿用管理端中性语义色与暖金选中态，支持滚轮缩放、拖拽、按钮缩放和复位；正式支持不低于 1024px 的 PC，窄窗口只保证基本可用。
- 影响文件：`CONTEXT.md`、`plan.md`；仍未修改管理端生产代码、依赖、路由或 App 版本。
- 验证：行政区划数据来源事实核查重新执行中；待完成加载、异常、数据版本和验收边界的最后一轮访谈。

## 2026-08-11 09:58 +0800
- 进度：`grill-with-docs` 第一轮确认地图定位。页面是只读的完整中国行政区划浏览工具，包含台湾，不承载业务统计、热力、选择、绑定或编辑；按全国省级轮廓 → 地级层级 → 县级行政区逐级进入，“县级行政区”包含市辖区、县级市、自治县、旗等同级单位。行政区划 GeoJSON 随前端版本静态打包，入口位于工具箱并以新标签页打开，地图浮层提供返回、当前位置、搜索、缩放和复位。
- 影响文件：`CONTEXT.md`、`plan.md`；尚未修改 `admin-web` 代码、路由、依赖或 App 版本。
- 验证：已核对管理端现有 ECharts 5.6 依赖与无 shell 工具页路由模式；行政区划数据来源、许可和体积正在核查，待完成后继续访谈与文档静态检查。

## 2026-08-11 09:51 +0800
- 进度：开始规划 PC 管理端中国行政区划地图页面。已确认页面采用 `admin 全屏地图工作区`：无后台 shell，地图画布占满浏览器视口，返回、搜索和层级操作只能以地图浮层呈现；尚未决定地图用途、县级交互、数据新鲜度和入口位置。
- 影响文件：`CONTEXT.md`、`plan.md`；尚未修改 `admin-web` 代码、路由或依赖。
- 验证：待完成需求访谈后执行文档差异、乱码和结构检查。

## 2026-08-11 09:18 +0800
- 进度：完成 `plan.md` 压缩清理，并明确今后纯部署等非开发操作无需写入开发账本。1133 条、5707 行历史流水按事项和月份归并；纯部署记录删除，已完成事项只保留最终结论，未发现仍待验收的任务目录。
- 影响文件：`AGENTS.md`、`plan.md`；既有未跟踪 `docs/examples/` 不纳入。
- 验证：`git diff --check -- AGENTS.md plan.md` 通过；U+FFFD 与 C0/DEL 控制字符扫描无输出；确认 `tasks/` 下 13 个任务目录均有 `DONE.md`；`plan.md` 压缩为 88 行，工作区范围仅含 `AGENTS.md`、`plan.md`，既有未跟踪 `docs/examples/` 未纳入。

## 当前状态

- 当前没有进行中的仓库实现任务；`tasks/` 下现有 13 个任务目录均已有 `DONE.md`。
- 最近一次功能提交为 `3262790`（修复 Hermes 编辑后资源漏检）。
- 工作区既有未跟踪 `docs/examples/` 不属于本次清理范围。

## 2026-08-09：Hermes 论坛采集可靠性

- 完成编辑后补资源的延迟定稿：标题含 `115` 或 `ed2k` 且首次无资源时保持 `pending`，按 30 分钟、1 小时、2 小时、之后每 6 小时复查；普通空帖立即定稿，历史最终态不追溯重开。7 项隔离回归、脚本编译和真实调度验证通过（`3262790`）。
- 补录漏通知帖子 `tid=3680933`，重新检查后以 `included` 入库 1 个 115 附件并完成单次通知，任务和队列恢复正常（`6fcf4c2`）。
- 修复历史补漏游标停滞：公开列表页的通用权限文案不再误判为访问受限，真实抓取失败不推进游标，空页或无待处理项正常推进；第 1991 页 30 条候选完成处理（`4646a37`）。
- 修复新帖监控遗漏：候选、检查和通知改为可恢复队列，只有明确投递成功才出队；抓取或投递失败非零退出并保留重试，故障恢复和幂等回归通过（`417a6cf`）。

## 2026-08-08：Hermes 队列与保留策略

- 历史补漏改为持久化稳定候选队列，检查成功后才移除，队列清空后才翻页；既有 102 条 `pending` 完成恢复处理，真实调度验证队列可跨批次续跑（`fce95b0`）。
- 论坛记录采用资源感知保留策略：超过 30 天且无附件、无 ED2K 才清理；30 天内记录和有资源的旧记录保留，管理端读取使用同一边界。仓储单测、可选 PostgreSQL 集成测试和 `go vet` 通过（`15e3cc0`）。

## 2026-08-07：TV App 精简

- 新版 TV App 完整移除 IPTV 客户端入口、路由、页面、状态、客户端 API/Repository 和专属资源；后端、管理端、频道数据、旧 APK、通用 LibVLC 与双 ABI 分发继续保留。TV 版本升至 `0.1.145 (145)`，617 项单测和 Debug 双 ABI 构建通过（`87ff990`）。

## 2026-08-06：手机 App 精简

- 手机端移除电影、电视剧入口与专属实现，只允许 `short`、`av` 内容进入列表、详情和播放器；混合来源支持跨页补读，异常旧入口提示不可用并回退。TV 扫码授权、短视频投屏和遥控链路保留。手机版本升至 `0.1.10 (11)`，165 项单测和 Debug 构建通过（`509ba89`）。

## 2026-08-04：论坛资源管理

- 新增管理端论坛资源列表与详情能力，统一展示帖子、附件和 ED2K 资源，并补齐后端读取模型、前端交互、ADR 和领域术语；管理端构建与 Go 定向验证通过（`09875fb`）。
- 完成 Hermes 两阶段论坛采集接口与脚本切换：以 `source + tid` 去重，区分 `dedupe_only`、`pending` 和最终检查态，支持多附件/多 ED2K 与机器 Token 鉴权；历史基线成功导入且重复导入幂等（`0663a8d`、`e07924b`）。

## 2026-08-02：文件大小与管理端标签

- 转码完成后记录主播放文件大小，新增历史回填命令并保持未知值为 `NULL`；服务层、仓储和命令测试通过（`c40e134`、`a82e3f5`）。
- 修复管理端视频标签远程候选，统一已有标签选择与新标签输入语义，管理端构建通过（`5a122cf`）。

## 2026-08-01：转码与资源链路修复

- 修复转码封面流误选、异常媒体流探测、进度假失败、AVC 硬编超长边和源文件兼容等问题；补充 ffmpeg 计划测试与长期兼容约束。
- 收口视频、图片、演员、合集等管理端资源链路的状态反馈、分页、预览与危险操作保护；定向测试和管理端构建通过。

## 2026-07：管理端、TV 与 115 能力

- 完成管理端 Precision Ops 分阶段改造：统一 compact 页面壳、加载/错误/空态、Drawer、分页、状态指示器、危险确认和键盘可达性；修复旧响应覆盖、查询身份错配、窄屏溢出、长文本裁剪等问题。各阶段均以定向测试、全量测试和生产构建收口。
- 完成 TV 首页、海报墙、短视频、长视频详情/播放器、焦点、字幕音轨、软刷新和状态反馈的多轮优化；相关 App 更新均同步递增版本并完成 TV 单测或构建验证。
- 完成 ED2K 下载工作台与 aMule 退役，保留浏览器端 ED2K 链接生成器和已入库成品；数据库历史表保持前向兼容，不再保留运行时代码与部署资产（`3c8fdd6`）。
- 完成 115 开放平台 OAuth/PKCE、令牌刷新、云下载与本地暂存链路设计及实现边界，保持“云下载完成”与“本地落盘/媒体入库”语义分离。
- 完成压缩包导入的编码纠偏、多分组、批量编辑/处理、稳定源路径、删除和去重兼容等能力；后端、管理端和相关文档验证通过。

## 2026-06：播放器、压缩包导入与管理端工具

- TV 长视频播放器完成软准备、软刷新、重试取消、焦点守卫、LibVLC 迁移和音轨/字幕偏好恢复；对应任务均已通过自动化与用户验收并写入 `DONE.md`。
- TV 短视频迁入独立全屏播放页，后续补齐单条失败非阻塞切换、播放提示、返回焦点和导航语义；相关 TV 版本、测试和构建同步更新。
- 压缩包导入完成从上传、扫描、批次详情、分组、编码纠偏到视频/图片处理的完整工作流；错误文本统一为合法 UTF-8，ZIP 编码模式与 RAR/7z 外部解压器边界明确。
- 管理端工具箱完成自适应布局和密码管理简单 CRUD；不引入乐观锁、导入导出、密码生成或额外验证。
- 修复 TV release APK 元数据测试随固件版本漂移、视频稳定源路径、WebP/压缩包处理和管理端页面交互等回归。

## 2026-05：Android/TV 主功能与任务闭环

- 建立独立 Android TV 工程，完成扫码配对、首页分类、搜索、海报墙、长视频/电视剧/IPTV 播放、字幕音轨、断点续播、焦点与遥控交互，并持续按功能更新版本。
- 手机端完成短视频、AV、电视剧、图集、演员详情、搜索、喜欢/收藏、统一播放器和 TV 授权协作等主链路；后续精简边界见 2026-08 记录。
- 后端完成电影/电视剧/AV 转码码率、字幕抽取、播放源、历史记录、刮削与媒体关联等多轮修复。
- 管理端完成演员、合集、图片、视频状态、手动刮削、APK 分发和工具页等能力，并统一分页、反馈与中文文案。
- `tasks/2026-05-23-*`、`tasks/2026-05-24-*`、`tasks/2026-05-25-*`、`tasks/2026-05-26-*` 均已通过对应 `review.md` 和用户验收，完成标记见各目录 `DONE.md`。

## 2026-04：项目基础能力

- 建立 Go 服务、PostgreSQL migration、管理端 Web、上传/分片/秒传、异步刮削、转码任务、签名播放与 Swagger/OpenAPI 基础链路。
- 建立视频、图片、演员、标签、合集、电视剧和 AV 领域模型及管理能力；补齐删除级联、并发竞态、空值扫描、源文件保留和失败恢复。
- 创建手机 Android 首版，完成服务发现、登录、分类、短视频竖滑播放、AV 详情、图集和统一长视频播放器。
- 统一中文 Markdown/界面文案、UTF-8 无乱码、开发脚本、环境文件加载和本地启动约定。

## 历史追溯规则

- 精确实现差异：`git log -- plan.md` 定位对应提交，再用 `git show <commit>` 查看。
- 长期术语与兼容约束：查阅 `CONTEXT.md`。
- 架构取舍：查阅 `docs/adr/`。
- 有 PRD/实现/评审流程的任务：查阅 `tasks/<任务>/prd.md`、`implement.md`、`review.md` 和 `DONE.md`。
- 本文件不再记录纯部署流水；只有部署暴露新的长期兼容约束或未解决阻塞时，才记录结论。
