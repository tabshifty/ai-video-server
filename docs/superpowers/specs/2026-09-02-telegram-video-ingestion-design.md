# Telegram 指定群组视频采集与历史回填设计

## 1. 目标与范围

本设计为 OrbStack 上运行的视频库增加 Telegram 指定群组采集能力。采集器使用有权限访问群组的个人 Telegram 账号，完成以下工作：

- 管理多个可监控的 Telegram 群组；新增群组后自动开始全量历史回填，并持续监听新消息。
- 历史消息和实时消息都只筛选视频媒体，统一按 `short` 类型进入现有视频库。
- 下载任务持久化、可限流、可重试、可从进程重启处继续。
- 复用现有 `UploadService`、PostgreSQL 视频记录、Redis/Asynq 转码队列和 FFmpeg worker。
- 记录 Telegram 来源关系、消息状态和失败原因，保证同一消息与同一内容不会重复入库。

本期不包含公开群组搜索、Bot API 采集、自动内容审核、自动 hashtag 标签化，也不要求首期提供管理端可视化页面。群组管理必须有受保护的管理 API，后续页面复用同一服务即可。

## 2. 总体架构

```text
Telegram 个人账号（持久化 MTProto session）
                 │
                 ▼
      telegram-ingestor（独立 Go 进程/容器）
      ├─ 实时更新监听
      ├─ 历史分页扫描
      ├─ 来源校验与消息登记
      ├─ telegram-realtime / telegram-backfill 消费者
      └─ 有界并发下载、哈希与导入
                 │
                 ├─ PostgreSQL：来源、消息、游标、关联和状态
                 ├─ Redis/Asynq：采集任务与控制任务
                 └─ 共享 UploadTempDir / StorageRoot
                                  │
                                  ▼
              UploadService（固定 type=short）
                                  │
                                  ▼
              现有 transcode 队列 → worker + FFmpeg
```

`telegram-ingestor` 是 Telegram session 的唯一持有者，避免多个进程同时刷新 session 或竞争下载引用。它既负责生产采集任务，也消费采集队列；API 服务只负责管理来源记录和投递控制任务，现有 `worker` 继续只处理转码、缩略图和已有刮削任务。

OrbStack 部署时，`api`、`worker` 和 `telegram-ingestor` 必须把宿主机目录挂载到相同的容器路径，至少包括：

- `/data/storage`：最终媒体和缩略图，对应 `STORAGE_ROOT`。
- `/data/tmp/uploads`：原始临时文件，对应 `UPLOAD_TEMP_DIR`。
- `/data/telegram-session`：个人账号 session，只挂载给 `telegram-ingestor`。

FFmpeg 和 `ffprobe` 只需安装在 `worker` 镜像中；采集器不自行转码。

## 3. 群组来源管理

新增 `telegram_sources` 表作为来源权威，不再用单个 `TELEGRAM_CHAT_ID` 环境变量。建议字段包括：

- `id`：项目内部 UUID。
- `chat_id`：Telegram 规范化后的整数 ID，唯一。
- `chat_ref`：管理员输入的用户名、邀请链接或 ID，保留用于审计。
- `title`、`username`：最近一次 Telegram 校验得到的显示信息。
- `enabled`：是否继续监听和处理。
- `sync_status`：`pending`、`backfilling`、`live`、`paused` 或 `error`。
- `history_cursor_message_id`：已登记的历史分页边界；保存具体消息 ID，不保存易漂移的数字页码。
- `last_error`、`next_retry_at`、`backfill_completed_at`、创建和更新时间。

管理 API 通过现有管理员鉴权提供：

- 添加来源：写入 `pending`，并投递来源校验/启动回填控制任务。
- 暂停、恢复来源：保留游标和消息记录；暂停期间不消费该来源的新下载任务。
- 手动重新回填：从当前游标继续，不删除已完成记录。
- 查询来源和回填进度：返回状态、已发现/已完成/失败数量和最近错误。

来源校验由采集器使用个人账号完成。校验成功后写回规范化 `chat_id` 和群组信息；账号无权访问、群组已删除或引用无效时，来源进入 `error`，不应把它当成空群组继续推进游标。

## 4. 消息记录与三层幂等

新增 `telegram_media` 表，每条群组消息一行，建议字段包括：

- `source_id`、`chat_id`、`message_id`：来源消息唯一键，唯一约束为 `(source_id, message_id)`。
- `telegram_document_id`、`document_dc_id`：Telegram 媒体标识及数据中心信息，可为空并建立索引。
- `filename`、`mime_type`、`file_size`、`caption`、`message_url`、`message_created_at`。
- `processing_status`：`discovered`、`queued`、`downloading`、`importing`、`imported`、`duplicate`、`failed` 或 `skipped`。
- `transcode_status`：`not_required`、`pending`、`enqueued`、`ready` 或 `failed`。
- `video_id`：已关联的视频库记录，可为空。
- `sha256`、`temp_path`、尝试次数、下次重试时间和最后错误。

幂等顺序固定为：

1. `(source_id, message_id)` 防止同一群组消息重复登记。
2. `telegram_document_id` 只用于下载前的快速命中：如果已经有成功导入的消息关联到视频，则当前消息直接关联同一个 `video_id`，标记为 `duplicate`，不下载。
3. 下载后计算 `SHA-256 + 文件大小`，调用视频库的内容查重作为最终依据。它能覆盖重新上传导致 Telegram document ID 改变的情况。

Telegram 的 `file_reference` 会过期，不能作为持久化唯一键。需要下载时必须重新读取消息并获得最新引用。`document_id` 也不能替代文件哈希；同一内容重新上传可能产生新的 document ID，跨群组的并发导入也必须由最终哈希约束兜底。

同一个视频出现在多个群组时，只创建一条视频库记录，每条 Telegram 消息仍保留来源关系。重复消息不覆盖首次入库的标题和描述。

## 5. 队列与执行流

采集任务使用独立 Asynq 队列，不占用现有 `transcode` 队列：

- `telegram-realtime`：实时消息，权重高，优先处理。
- `telegram-backfill`：历史回填，权重低，后台慢速运行。
- `telegram-control`：来源校验、暂停/恢复和回填控制。

任务 payload 只携带 `telegram_media_id` 或 `telegram_source_id`，不把临时文件路径作为唯一事实；消费者每次从数据库读取最新状态。采集器下载并发默认 1，允许通过配置提高到 2，但必须有最大活跃任务数和磁盘空间保护。

单条消息的执行流如下：

1. 消费者以行锁或等价的 claim 逻辑取得 `discovered/queued/failed` 且已到重试时间的记录。
2. 重新读取 Telegram 消息，确认仍是视频并取得最新文件引用；媒体消失时记录可重试或最终失败原因。
3. 再次按 document ID 查找已经成功导入的记录，命中则建立关联并清理临时文件。
4. 下载到 `UPLOAD_TEMP_DIR` 下的随机临时路径，使用文件大小校验和流式 SHA-256；下载未完成不得进入导入步骤。
5. 先用不改变既有视频路径的预检查做内容查重。已有相同视频时只关联 `video_id`，不把临时路径写入既有视频记录。
6. 新内容调用现有 `UploadService` 的本地文件导入入口，固定 `type=short`。标题优先使用 caption，没有 caption 时使用去掉扩展名的原始文件名。
7. 描述保存完整 caption，并追加群组、消息链接和发送时间；原始文件名、Telegram 标识等放入结构化 metadata。首期不把 hashtag 自动转成标签。
8. 新视频创建成功后投递现有 `TranscodePayload`；转码完成由现有 worker 生成 MP4、缩略图和媒体元数据。
9. 根据导入结果更新 `video_id`、`processing_status` 和 `transcode_status`，最后清理不再需要的临时文件。

历史回填使用 Telegram 消息 ID 作为边界，按从新到旧分页读取。每页消息先幂等写入 `telegram_media`，再投递 backfill 任务；游标只推进到已经持久化登记的边界。进程中断后重复扫描一页是允许的，由唯一约束消除重复。回填期间实时监听继续登记新消息，切换到 `live` 前保留边界重叠检查，避免扫描与实时更新之间出现空窗。

## 6. 现有上传链路的适配边界

现有 `UploadService.SaveUploadedFile` 已经负责哈希校验、视频记录、原始路径和重复判断，`short` 返回 `Enqueue=true`，但它不是队列事务的一部分。Telegram 适配层必须遵守两条约定：

- 对已存在内容先走只读预检查，不能把即将删除的临时路径写回已有视频的 `original_path`。必要时为本地导入增加“重复时不刷新原片路径”的选项，避免重复导入后留下悬空路径。
- 视频记录创建和 Asynq 投递之间采用可补偿状态，而不是假设两者原子完成。`transcode_status=pending` 的记录由采集器启动时和周期性 reconciliation 查询并重新投递；重复转码任务必须可安全跳过。

采集器使用配置的系统归属用户 UUID 写入现有 `videos.user_id`，启动时校验该 UUID 对应用户存在。不能使用随机 UUID 或把 Telegram 账号凭据写入视频记录。

## 7. 重试、限流与失败处理

- Telegram `FloodWait`：读取服务端等待秒数，暂停对应任务或来源，到期后再重试；不得固定高频轮询。
- 网络断开、临时文件 IO 和 Asynq 投递失败：指数退避并限制最大退避时间，保留消息记录和错误。
- 权限撤销、媒体不存在、文件类型不再匹配：记录明确原因；可判定为永久失败时停止无限重试，但来源本身不能因此静默完成。
- 下载成功但导入或转码投递失败：保留临时文件路径和状态，由补偿任务继续；超过保留期限后才由清理任务删除，并保留失败记录。
- 采集器重启：先恢复未完成的 `queued/downloading/importing` 记录，再继续来源游标，不重新下载已标记 `imported/duplicate` 的消息。

数据库是状态最终依据，Redis/Asynq 只负责调度。队列丢失时可根据数据库的 `queued`、`transcode_status=pending` 和 `next_retry_at` 重建任务。

## 8. 登录、配置与安全

首次部署使用一次性登录命令完成个人账号验证码和二次验证，生成持久化 session；之后只从 `/data/telegram-session` 读取 session，不把验证码、二次验证密码或 API hash 写入日志。建议配置项包括：

- Telegram API ID、API hash、电话号码和 session 路径。
- 系统归属用户 UUID。
- 采集并发、单文件大小上限、重试和临时文件保留时间。
- PostgreSQL、Redis、`STORAGE_ROOT` 和 `UPLOAD_TEMP_DIR`。

session 文件和含密钥的环境文件只允许服务用户读取；管理 API 必须沿用现有管理员鉴权。只采集账号有权访问且确认可以保存的群组内容，部署者负责内容版权、隐私和平台条款合规。

## 9. 验证与上线顺序

验证重点包括：

- migration 唯一约束、来源状态和历史游标在重启后的恢复。
- 同消息重复事件、同 document ID 跨群组、同 SHA-256 不同 document ID 的三层去重。
- “视频记录已创建但转码投递失败”的补偿，以及重复任务不会破坏视频状态。
- Telegram file reference 过期后的重新读取、FloodWait 和网络中断退避。
- caption/文件名标题规则、来源 metadata 和短视频播放链路。
- OrbStack 中 API、worker、采集器对共享路径的可见性，以及 worker 内 `ffmpeg/ffprobe` 可用性。

上线分为四步：先迁移并部署采集器登录能力；再只对测试群组执行小批量回填；确认队列、磁盘和转码稳定后启动全部历史回填；最后开启多个来源和实时监听。历史回填不删除 Telegram 原消息，也不改变现有视频记录的首次标题和描述。

