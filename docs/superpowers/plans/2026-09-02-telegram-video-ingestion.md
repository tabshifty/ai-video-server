# Telegram 视频采集与历史回填实施计划

> 面向实现者：按任务顺序执行，每个任务都要完成测试、定向验证和独立中文提交。实现遵循仓库本地流程，不改变现有 Android 工程；本计划不调用外部工作流。

**目标：** 在 OrbStack 中增加一个使用个人 Telegram MTProto 会话的独立采集器，支持多个指定群组的全量历史回填和实时监听，将视频统一按 `short` 导入现有视频库。

**架构：** `telegram-ingestor` 独占 Telegram session，同时生产并消费 `telegram-control`、`telegram-realtime` 和 `telegram-backfill` 队列。PostgreSQL 保存来源、消息、游标和导入状态；采集器通过安全的本地文件导入入口复用现有 `UploadService`，再投递既有 `transcode` 任务给 worker。

**技术栈：** Go 1.22、PostgreSQL migrations、pgx/v5、Redis/Asynq、Telegram MTProto（`github.com/gotd/td`）、现有 FFmpeg/ffprobe worker、OrbStack Docker Compose。

## 全局约束

- Telegram 只接受个人账号能访问的指定群组，不实现公开搜索和 Bot API 历史采集。
- 所有 Telegram 视频固定使用视频类型 `short`，不触发电影、电视剧或 AV 刮削。
- 消息键 `(source_id, message_id)`、Telegram `document_id` 快速键、下载后 `SHA-256 + 文件大小` 必须同时保留，document ID 不能替代文件哈希。
- `file_reference` 只在即时下载前刷新获取，不作为持久化标识。
- API、worker、采集器必须使用相同容器路径挂载 `STORAGE_ROOT` 和 `UPLOAD_TEMP_DIR`；session 目录只挂载给采集器。
- FFmpeg 和 `ffprobe` 只安装在 worker 镜像；采集器不转码。
- 队列是调度层，PostgreSQL 状态是最终依据；视频创建与转码投递之间必须可由 reconciliation 补偿。
- 所有 Markdown、管理 API 可见文案和 Git 提交信息使用无乱码中文。
- migration 编号从当前 `0037` 后的 `0038` 开始，必须同时提供可回滚 down migration。
- 不修改 Android 代码，因此不递增 Android 手机端或 TV 端版本号。
- 计划和实现进度追加到根级 `plan.md`；功能完成后将长期约定追加到 `CONTEXT.md`。

---

### Task 1: 建立 Telegram 来源与媒体状态模型

**Files:**

- Create: `migrations/0038_telegram_video_ingestion.up.sql`
- Create: `migrations/0038_telegram_video_ingestion.down.sql`
- Create: `internal/models/telegram.go`
- Create: `internal/repository/telegram_repository.go`
- Create: `internal/repository/telegram_repository_test.go`
- Modify: `internal/repository/migrations_test.go`

**Interfaces:**

- Produces `models.TelegramSource`、`models.TelegramMedia`，以及 `VideoRepository` 上的来源/媒体读写方法，供后续服务和 handler 使用。
- `TelegramSource` 至少暴露 `ID`、`ChatID`、`ChatRef`、`Title`、`Username`、`Enabled`、`SyncStatus`、`HistoryCursorMessageID`、`LastError`、`NextRetryAt`、`BackfillCompletedAt`、`CreatedAt`、`UpdatedAt`。
- `TelegramMedia` 至少暴露 `ID`、`SourceID`、`ChatID`、`MessageID`、`DocumentID`、`DocumentDCID`、`Filename`、`MIMEType`、`FileSize`、`Caption`、`MessageURL`、`MessageCreatedAt`、`ProcessingStatus`、`TranscodeStatus`、`VideoID`、`SHA256`、`TempPath`、`Attempts`、`NextRetryAt`、`LastError`、`CreatedAt`、`UpdatedAt`。

- [ ] **Step 1: 先写迁移契约测试**

在 `internal/repository/migrations_test.go` 增加 `TestTelegramVideoIngestionMigration`，读取 `0038` 两个 SQL 文件并断言：两张表存在；`telegram_sources.chat_id` 唯一；`telegram_media(source_id,message_id)` 唯一；媒体外键删除行为为 `ON DELETE CASCADE`；`video_id` 为 `ON DELETE SET NULL`；状态 CHECK 包含设计文档中的 source、processing 和 transcode 状态；document ID、处理状态和重试时间建立索引。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/repository -run 'TestTelegramVideoIngestionMigration$' -count=1
```

预期：FAIL，提示缺少 `0038_telegram_video_ingestion.up.sql` 或断言的表/约束。

- [ ] **Step 3: 写 migration 和模型**

迁移使用 `CREATE TABLE IF NOT EXISTS`，source 状态限定为 `pending/backfilling/live/paused/error`，media 处理状态限定为 `discovered/queued/downloading/importing/imported/duplicate/failed/skipped`，转码状态限定为 `not_required/pending/enqueued/ready/failed`。`file_size`、`attempts`、消息 ID 必须非负；`sha256` 使用 64 字符字段。down migration 先删媒体表再删来源表。

模型字段使用现有 `models` 包的 UUID、`time.Time` 和 JSON tag 约定；可空时间、document ID、video ID 使用指针，避免把数据库 NULL 误读成有效值。

- [ ] **Step 4: 实现仓储边界**

在 `telegram_repository.go` 实现以下方法，所有 SQL 集中在 repository 包：

```go
func (r *VideoRepository) CreateTelegramSource(ctx context.Context, source models.TelegramSource) error
func (r *VideoRepository) ListTelegramSources(ctx context.Context, enabledOnly bool) ([]models.TelegramSource, error)
func (r *VideoRepository) GetTelegramSource(ctx context.Context, sourceID uuid.UUID) (models.TelegramSource, error)
func (r *VideoRepository) UpdateTelegramSource(ctx context.Context, sourceID uuid.UUID, patch TelegramSourcePatch) error
func (r *VideoRepository) InsertTelegramMedia(ctx context.Context, media models.TelegramMedia) (models.TelegramMedia, bool, error)
func (r *VideoRepository) GetTelegramMedia(ctx context.Context, mediaID uuid.UUID) (models.TelegramMedia, error)
func (r *VideoRepository) FindTelegramMediaByDocumentID(ctx context.Context, documentID int64) (models.TelegramMedia, bool, error)
func (r *VideoRepository) ClaimTelegramMedia(ctx context.Context, mediaID uuid.UUID, allowed []string) (models.TelegramMedia, bool, error)
func (r *VideoRepository) UpdateTelegramMedia(ctx context.Context, mediaID uuid.UUID, patch TelegramMediaPatch) error
func (r *VideoRepository) ListTelegramMediaNeedingTranscode(ctx context.Context, limit int) ([]models.TelegramMedia, error)
func (r *VideoRepository) CountTelegramMediaByStatus(ctx context.Context, sourceID uuid.UUID) (map[string]int, error)
```

`InsertTelegramMedia` 通过唯一约束返回已有记录而不是覆盖最终状态；`ClaimTelegramMedia` 使用事务、行锁和状态条件避免两个下载 worker 同时处理一条消息；游标更新只能与消息登记处于同一持久化边界。

- [ ] **Step 5: 运行通过测试并提交**

运行：

```bash
go test ./internal/repository -run 'TestTelegram|Test.*Migration' -count=1
go vet ./internal/repository
```

预期：测试 PASS，`go vet` 无输出。提交：

```bash
git add migrations/0038_telegram_video_ingestion.up.sql migrations/0038_telegram_video_ingestion.down.sql internal/models/telegram.go internal/repository/telegram_repository.go internal/repository/telegram_repository_test.go internal/repository/migrations_test.go
git commit -m "新增 Telegram 采集状态模型"
```

### Task 2: 增加配置和采集队列原语

**Files:**

- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `.env.example`
- Create: `internal/queue/telegram_tasks.go`
- Create: `internal/queue/telegram_tasks_test.go`

**Interfaces:**

- `config.Config` 新增 Telegram API ID/hash、电话、session 路径、导入归属用户 UUID、三个队列名、下载并发和最大活跃任务字段。
- 新增 `func (c Config) ValidateTelegramIngestor() error`，只由采集器入口调用，不影响 server/worker 使用现有最小环境变量启动。
- 新增 `queue.TelegramTaskEnqueuer`：

```go
type TelegramTaskEnqueuer struct {
    client *asynq.Client
    realtimeQueue string
    backfillQueue string
    controlQueue string
}
func NewTelegramTaskEnqueuer(redisAddr, redisPassword, realtimeQueue, backfillQueue, controlQueue string) *TelegramTaskEnqueuer
func (e *TelegramTaskEnqueuer) Close() error
func (e *TelegramTaskEnqueuer) EnqueueSourceSync(sourceID uuid.UUID) error
func (e *TelegramTaskEnqueuer) EnqueueRealtime(mediaID uuid.UUID) error
func (e *TelegramTaskEnqueuer) EnqueueBackfill(mediaID uuid.UUID) error
func (e *TelegramTaskEnqueuer) EnqueueReconcile() error
```

- [ ] **Step 1: 写配置与队列失败测试**

增加 `TestLoadIncludesTelegramConfig`、`TestValidateTelegramIngestorRequiresCredentials`，以及队列选项测试，验证默认队列名分别为 `telegram-realtime`、`telegram-backfill`、`telegram-control`，并发默认为 1、最大活跃任务为正数；校验缺少 API ID、API hash、电话、session 路径或归属用户 UUID 时返回明确错误。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/config ./internal/queue -run 'Test(LoadIncludesTelegram|ValidateTelegram|TelegramTask)' -count=1
```

预期：FAIL，因为配置字段和队列类型尚未存在。

- [ ] **Step 3: 实现配置加载和 Asynq payload**

在 `Config` 增加可选 Telegram 字段；`Load` 只解析、不强制校验 Telegram 凭据。`ValidateTelegramIngestor` 使用 `uuid.Parse` 校验归属用户，限制并发至少为 1，最大活跃任务不小于并发。队列方法的 payload 只包含 UUID 字符串：

```go
type TelegramMediaTaskPayload struct { MediaID string `json:"media_id"` }
type TelegramSourceTaskPayload struct { SourceID string `json:"source_id"` }
```

任务常量为 `telegram:source-sync`、`telegram:download`、`telegram:reconcile`；realtime/backfill 通过不同 queue option 调度。每项任务设置 3 次最大重试和 6 小时上限，下载任务的业务退避由数据库 `next_retry_at` 控制。

- [ ] **Step 4: 验证通过并提交**

运行：

```bash
go test ./internal/config ./internal/queue -run 'Test(LoadIncludesTelegram|ValidateTelegram|TelegramTask)' -count=1
go vet ./internal/config ./internal/queue
```

预期：PASS。同步更新 `.env.example` 的中文注释，提交：

```bash
git add internal/config/config.go internal/config/config_test.go .env.example internal/queue/telegram_tasks.go internal/queue/telegram_tasks_test.go
git commit -m "增加 Telegram 配置与采集队列"
```

### Task 3: 修正本地导入的重复路径边界

**Files:**

- Modify: `internal/services/upload.go`
- Modify: `internal/services/upload_test.go`
- Create: `internal/services/import_file.go`
- Create: `internal/services/import_file_test.go`

**Interfaces:**

- Produces `func (s *UploadService) SaveImportedFile(ctx context.Context, in LocalUploadInput, maxVideoSize int64) (UploadResult, error)`。
- `SaveUploadedFile` 保留现有手动上传语义；`SaveImportedFile` 与它共享校验、创建记录和 file hash 逻辑，但重复内容时绝不把临时导入路径写回既有视频的 `original_path`。
- 此任务只抽出可复用的“不刷新重复原片路径”内部选项，Telegram 元数据由 Task 4 提供。

- [ ] **Step 1: 写重复路径回归测试**

新增测试：已有相同 SHA-256 的 `uploaded` 视频，调用 `SaveImportedFile` 后返回 `AlreadyExists=true`、返回原视频 ID，数据库中的 `original_path` 保持原值，导入临时文件可由调用者删除；新文件仍创建 `uploaded` 状态并返回 `Enqueue=true`。保留现有 `SaveUploadedFile` 测试，确保手动上传行为不变。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/services -run 'TestSaveImportedFile|Test.*OriginalPath' -count=1
```

预期：FAIL，当前服务没有独立的安全导入入口。

- [ ] **Step 3: 实现共享保存逻辑**

把 `SaveUploadedFile` 的核心逻辑收敛到带布尔选项的内部函数；手动入口传入 `refreshExistingOriginalPath=true`，新入口传入 `false`。两条路径都必须校验 server SHA-256 与输入 hash、大小和 `short` 类型；只在真正创建新视频时写入 `original_path`。

- [ ] **Step 4: 验证通过并提交**

运行：

```bash
go test ./internal/services -run 'TestSaveImportedFile|Test.*OriginalPath|TestUpload' -count=1
go vet ./internal/services
```

预期：PASS。提交：

```bash
git add internal/services/upload.go internal/services/upload_test.go internal/services/import_file.go internal/services/import_file_test.go
git commit -m "修复外部导入重复路径处理"
```

### Task 4: 封装 Telegram MTProto 客户端和元数据规则

**Files:**

- Modify: `go.mod`
- Modify: `go.sum`
- Create: `internal/telegram/client.go`
- Create: `internal/telegram/gotd_client.go`
- Create: `internal/telegram/client_test.go`
- Create: `internal/services/telegram_metadata.go`
- Create: `internal/services/telegram_metadata_test.go`

**Interfaces:**

- `internal/telegram.Client` 只暴露采集器需要的能力，具体 gotd 类型不得泄漏到 services：

```go
type Message struct {
    ChatID int64
    MessageID int64
    DocumentID int64
    DocumentDCID int
    Filename string
    MIMEType string
    Size int64
    Caption string
    MessageURL string
    SentAt time.Time
}
type Chat struct { ID int64; Title string; Username string }
type Client interface {
    ResolveChat(ctx context.Context, ref string) (Chat, error)
    History(ctx context.Context, chatID, offsetID int64, limit int, fn func(Message) error) error
    Subscribe(ctx context.Context, fn func(Message) error) error
    RefreshMessage(ctx context.Context, chatID, messageID int64) (Message, error)
    Download(ctx context.Context, message Message, dst io.Writer) error
}
```

- `BuildTelegramVideoMetadata` 的输入是 `Message`、群组标题和导入时间，输出标题、描述和 JSON metadata：caption 非空时作为标题；否则使用去扩展名文件名；描述包含完整 caption、群组、消息链接和发送时间；metadata 包含原始文件名、chat ID、message ID、document ID、发送时间。

- [ ] **Step 1: 写 fake client 与元数据失败测试**

在 `client_test.go` 使用内存 fake 验证 `History` 的 offset/limit 传递和 `RefreshMessage` 的调用契约；在 metadata 测试覆盖 caption 标题、文件名回退、caption 原文保留、消息链接和时间格式。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/telegram ./internal/services -run 'Test(Client|BuildTelegramVideoMetadata)' -count=1
```

预期：FAIL，因为客户端接口、gotd adapter 和元数据构造器尚未存在。

- [ ] **Step 3: 固定 MTProto 依赖并实现 adapter**

使用 `github.com/gotd/td` 的 Go 1.22 兼容稳定版本并写入 `go.mod/go.sum`。`gotdClient` 负责 API ID/hash 初始化、session 存储、个人账号登录、消息分页、更新订阅和下载；错误类型保留 FloodWait 等待秒数，供 Task 5 退避。adapter 将 Telegram media document 的 MIME、文件名、大小和 document ID 规范化为本包的 `Message`。

- [ ] **Step 4: 验证通过并提交**

运行：

```bash
go test ./internal/telegram ./internal/services -run 'Test(Client|BuildTelegramVideoMetadata)' -count=1
go vet ./internal/telegram ./internal/services
```

预期：PASS。提交：

```bash
git add go.mod go.sum internal/telegram internal/services/telegram_metadata.go internal/services/telegram_metadata_test.go
git commit -m "封装 Telegram 客户端与视频元数据"
```

### Task 5: 实现来源同步、下载处理和转码补偿

**Files:**

- Create: `internal/services/telegram_ingestion.go`
- Create: `internal/services/telegram_ingestion_test.go`
- Create: `internal/queue/telegram_processor.go`
- Create: `internal/queue/telegram_processor_test.go`
- Create: `cmd/telegram-ingestor/main.go`
- Modify: `main.go` only where shared constructors or queue types must remain compatible

**Interfaces:**

- `TelegramIngestionService` 构造函数接收 `telegram.Client`、来源媒体仓储接口、`UploadService`、`TelegramTaskEnqueuer`、现有转码 `queue.Enqueuer`、共享路径、最大视频大小和系统归属用户 UUID。
- 服务公开：

```go
func (s *TelegramIngestionService) SyncSource(ctx context.Context, sourceID uuid.UUID) error
func (s *TelegramIngestionService) HandleMessage(ctx context.Context, sourceID uuid.UUID, message telegram.Message, priority MessagePriority) error
func (s *TelegramIngestionService) ProcessMedia(ctx context.Context, mediaID uuid.UUID) error
func (s *TelegramIngestionService) ReconcileTranscodes(ctx context.Context, limit int) error
```

- `TelegramProcessor.Register(mux *asynq.ServeMux)` 注册 source-sync、download 和 reconcile handler；payload 错误、未知 UUID 和不可重试业务错误必须返回可观察错误。

- [ ] **Step 1: 写编排失败测试**

测试以下场景：同消息重复登记不产生第二条记录；成功 document ID 命中直接标记 duplicate 且不调用 Download；不同 document ID 下载后相同 SHA-256 关联已有 video；新文件调用 `SaveImportedFile(type="short")` 并投递转码；caption 元数据传入描述；FloodWait 写入 `next_retry_at`；转码投递失败保留 `transcode_status=pending`，reconcile 可再次投递；历史游标重复扫描一页不会漏消息。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/services ./internal/queue -run 'TestTelegram(Ingestion|Processor)' -count=1
```

预期：FAIL，因为编排服务和 processor 尚未实现。

- [ ] **Step 3: 实现消息登记和历史回填**

实时更新先调用 `InsertTelegramMedia`，再投递 realtime 任务；历史 `SyncSource` 使用消息 ID 边界从新到旧分页，每页先幂等登记再推进 `history_cursor_message_id`，重复页安全。来源状态按 `pending → backfilling → live` 迁移，暂停来源不投递下载任务；实时监听在回填期间持续登记，切换 live 前执行边界重叠扫描。

- [ ] **Step 4: 实现下载、三层去重和导入**

处理任务先 claim 数据库记录，再 RefreshMessage 获取最新 `file_reference`，检查 document ID，下载到随机临时路径并流式计算 SHA-256。命中内容哈希时只写来源关联并清理临时文件；新内容调用 `SaveImportedFile`，固定 `Type: "short"`、系统归属用户 UUID、metadata 和标题/描述规则。创建成功后调用现有 `EnqueueTranscode`，状态写为 pending/enqueued；任何中断保留失败原因和重试时间。

- [ ] **Step 5: 实现 Asynq processor、reconcile 和登录入口**

processor 的 realtime/backfill server 使用队列权重 10:1，下载并发由配置限制；周期任务查询 `queued/downloading/importing` 超时记录和 `transcode_status=pending` 记录补建任务。`cmd/telegram-ingestor` 支持 `-mode login` 一次性输入验证码/二次验证并写 session，支持 `-mode run` 启动监听、回填调度和 Asynq consumer；启动前调用 `ValidateTelegramIngestor` 并校验系统归属用户存在。

- [ ] **Step 6: 验证通过并提交**

运行：

```bash
go test ./internal/services ./internal/queue ./internal/telegram ./cmd/telegram-ingestor -count=1
go vet ./internal/services ./internal/queue ./internal/telegram ./cmd/telegram-ingestor
go build ./cmd/telegram-ingestor
```

预期：全部 PASS，命令构建成功。提交：

```bash
git add internal/services/telegram_ingestion.go internal/services/telegram_ingestion_test.go internal/queue/telegram_processor.go internal/queue/telegram_processor_test.go cmd/telegram-ingestor/main.go main.go
git commit -m "实现 Telegram 视频采集与回填"
```

### Task 6: 增加管理员来源管理 API

**Files:**

- Create: `internal/services/telegram_source.go`
- Create: `internal/services/telegram_source_test.go`
- Create: `internal/handlers/admin_telegram.go`
- Create: `internal/handlers/admin_telegram_test.go`
- Modify: `internal/handlers/router.go`
- Modify: `main.go` to construct and inject the source service

**Interfaces:**

- `TelegramSourceService` 暴露 `List`、`Add`、`Pause`、`Resume`、`StartBackfill`、`GetProgress`，新增来源时只接收 `chat_ref`，不让客户端伪造规范化 chat ID。
- 管理 API 使用现有 admin middleware：

```text
GET  /api/v1/admin/telegram/sources
POST /api/v1/admin/telegram/sources                 {"chat_ref":"..."}
POST /api/v1/admin/telegram/sources/:id/pause
POST /api/v1/admin/telegram/sources/:id/resume
POST /api/v1/admin/telegram/sources/:id/backfill
GET  /api/v1/admin/telegram/sources/:id/progress
```

- [ ] **Step 1: 写 service/handler 失败测试**

服务测试覆盖空引用、重复来源、暂停/恢复保留游标、添加成功后投递 source-sync；handler 测试覆盖未登录 401、非 admin 403、合法 JSON 201/200、错误响应使用现有 `{code,msg,data}` 包装。测试不得调用真实 Telegram 网络。

- [ ] **Step 2: 运行失败测试**

运行：

```bash
go test ./internal/services ./internal/handlers -run 'TestAdminTelegram|TestTelegramSource' -count=1
```

预期：FAIL，因为 service、handler 和路由尚未存在。

- [ ] **Step 3: 实现 service、handler 和路由**

service 负责规范化输入、调用 repository、投递 control 任务；Telegram 账号权限校验在采集器处理 source-sync 时完成，失败写回 source `error`。handler 只解析参数、调用 service 和返回现有响应 envelope，不在 HTTP 请求中执行历史扫描。路由放入现有 `/api/v1/admin` 分组，继续使用 `AuthMiddleware` 和 `AdminRequired`。

- [ ] **Step 4: 验证通过并提交**

运行：

```bash
go test ./internal/services ./internal/handlers -run 'TestAdminTelegram|TestTelegramSource' -count=1
go vet ./internal/services ./internal/handlers
```

预期：PASS。提交：

```bash
git add internal/services/telegram_source.go internal/services/telegram_source_test.go internal/handlers/admin_telegram.go internal/handlers/admin_telegram_test.go internal/handlers/router.go internal/handlers/router_test.go main.go
git commit -m "增加 Telegram 群组管理接口"
```

### Task 7: OrbStack 部署、运行手册和最终验收

**Files:**

- Create: `deploy/video-server.Dockerfile`
- Create: `deploy/docker-compose.telegram.yml`
- Create: `docs/telegram-video-ingestion.md`
- Modify: `.env.example`
- Modify: `CONTEXT.md`
- Modify: `plan.md`

**Interfaces:**

- Compose 提供 `api`、`worker`、`telegram-ingestor` 三个应用服务，并复用现有 `postgres`、`redis`；应用服务通过同一构建镜像的不同 command 运行。
- 运行手册必须给出登录、添加群组、观察回填进度、暂停/恢复、清理失败任务和回滚 migration 的实际命令。

- [ ] **Step 1: 写部署验收脚本/文档检查**

增加文档静态检查，断言 compose 中三个应用服务的 storage/temp 挂载路径完全相同、session 只挂载给采集器、worker 镜像安装 `ffmpeg` 和 `ffprobe`、采集器 command 包含 `-mode run`，并断言 `.env.example` 含全部 Telegram 配置名。

- [ ] **Step 2: 实现镜像与 compose**

多阶段 Dockerfile 使用 Go 1.22 构建二进制，运行层安装 `ca-certificates`、`ffmpeg` 和 `ffprobe`；compose 将宿主目录显式映射到 `/data/storage`、`/data/tmp/uploads`、`/data/telegram-session`，为采集器设置非 root 用户和 session 文件权限。API/worker 仍通过 `APP_MODE=server/worker` 启动，采集器通过构建出的专用二进制启动。

- [ ] **Step 3: 补充长期约定和运行手册**

在 `CONTEXT.md` 追加 Telegram 来源、三层幂等、游标推进、队列与数据库状态边界、重复导入原片路径保护和共享容器路径约定；在 `plan.md` 追加本次计划、每个实现提交和验证结果。运行手册全部使用中文并明确个人账号 session、版权权限和 FloodWait 注意事项。

- [ ] **Step 4: 执行定向、全量和部署验证**

依次运行：

```bash
go test ./internal/config ./internal/models ./internal/repository ./internal/services ./internal/queue ./internal/telegram ./internal/handlers ./cmd/telegram-ingestor -count=1
go test ./... -count=1
go vet ./...
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml config
```

然后在测试群组执行一次登录、添加来源、小批量回填、重启采集器、恢复回填和实时视频验证；确认视频按 `short` 出现在短视频库、重复消息不下载、相同哈希不创建第二条视频、转码 worker 能读取共享路径。部署失败时只回滚本任务新增提交和 `0038` migration，不触碰既有用户视频。

- [ ] **Step 5: 最终提交**

确认 `git diff --check`、UTF-8 乱码扫描、工作区只含本任务预期改动后提交：

```bash
git add deploy/video-server.Dockerfile deploy/docker-compose.telegram.yml docs/telegram-video-ingestion.md .env.example CONTEXT.md plan.md
git commit -m "补充 Telegram 采集部署与运维文档"
```
