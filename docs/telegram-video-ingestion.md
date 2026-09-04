# Telegram 视频采集部署与管理

本文说明如何在 OrbStack 或兼容 Docker Compose 的环境中部署 Telegram 视频采集，并通过管理端完成账号授权、来源添加和采集控制。采集器使用个人 Telegram 账号的 MTProto session，只采集管理员明确添加且该账号有权访问的群组或频道，并将视频固定导入短视频库。

日常管理入口是 `https://<服务地址>/admin/telegram`。不需要也不应再通过采集器终端执行登录、直接调用来源创建接口或手工修改 Telegram 媒体状态。

## 职责边界

- 管理端负责 Telegram 采集账号授权、来源预解析与确认、暂停、恢复、重新回填、进度查看和审计查询。
- 部署人员负责准备 Telegram 应用凭据、导入归属用户、环境变量、migration 和容器生命周期。
- `telegram-ingestor` 是唯一读取和保存 MTProto session 的进程；API 只通过 Docker 内网控制通道转发管理请求。
- 管理端不显示或修改 API hash、控制令牌、导入归属用户和容器配置，也不启动或停止容器。

## 部署前检查

- 安装 OrbStack 或 Docker Compose v2，并确认 `docker compose version` 可用。
- 在 <https://my.telegram.org> 创建应用，取得 API ID 和 API hash。
- 准备一个有权读取目标群组历史消息的个人账号。不要使用他人的 session，也不要采集、保存或传播无权处理的内容。
- 确认数据库中已有接收导入视频的用户；登录接口返回的 `data.user_id` 可直接作为 `TELEGRAM_IMPORT_USER_ID`。
- 同一份 session 只能由一个 `telegram-ingestor` 进程使用。不要同时在宿主机和容器中启动采集器。

## 配置目录和环境变量

从仓库根目录准备配置：

```bash
cp .env.example .env
mkdir -p ./storage ./tmp/uploads ./telegram-session
chmod 700 ./telegram-session
```

生产部署应把以下宿主路径改成绝对路径，并保证 UID/GID `10001:10001` 对三个目录有读写权限。Linux 可执行 `sudo chown -R 10001:10001 <目录>`；OrbStack for macOS 通常会自动映射当前用户的 bind mount 权限。

```dotenv
VIDEO_STORAGE_HOST_PATH=/绝对路径/storage
VIDEO_UPLOAD_TEMP_HOST_PATH=/绝对路径/tmp/uploads
TELEGRAM_SESSION_HOST_PATH=/绝对路径/telegram-session
```

编辑 `.env`，至少填写：

```dotenv
JWT_SECRET=<高强度随机值>
PASSWORD_VAULT_KEY=<部署后保持不变的高强度随机值>
TELEGRAM_API_ID=<my.telegram.org 提供的整数>
TELEGRAM_API_HASH=<my.telegram.org 提供的值>
TELEGRAM_PHONE=<带国家区号的手机号>
TELEGRAM_IMPORT_USER_ID=<现有用户 UUID>
TELEGRAM_CONTROL_TOKEN=<高强度随机值>
```

`TELEGRAM_PHONE` 是首次手机号授权接收验证码的账号。`TELEGRAM_CONTROL_TOKEN` 必须由 API 和 `telegram-ingestor` 使用同一值；Compose 默认将其限制在内部网络，不能映射 `TELEGRAM_CONTROL_ADDR` 到宿主机，也不能把令牌发给浏览器。

Compose 会把 API、worker 和采集器的持久媒体与上传暂存目录统一映射为 `/data/storage` 和 `/data/tmp/uploads`。`TELEGRAM_SESSION_PATH` 固定为 `/data/telegram-session/account.session`，该目录只挂载到以非 root 用户运行的采集器。

不要提交 `.env` 或 session 文件。备份 session 前先停止采集器，并按账号凭据的安全等级加密保存备份。

## 应用 migration 并启动服务

先启动并等待 PostgreSQL、Redis：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d postgres redis
until docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres pg_isready -U video -d video_server; do sleep 2; done
POSTGRES_CONTAINER=video_server_postgres ./scripts/migrate-apply.sh
```

`0038_telegram_video_ingestion.up.sql` 创建来源与媒体状态，`0039_telegram_management.up.sql` 创建账号状态、采集器心跳、短期授权元数据和管理审计。应用完成后可核对：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -c "SELECT version, applied_at FROM schema_migrations WHERE version IN ('0038_telegram_video_ingestion.up.sql', '0039_telegram_management.up.sql') ORDER BY version;"
```

构建并启动三个应用服务：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d --build api worker telegram-ingestor
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml ps
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml logs -f telegram-ingestor
```

首次部署时，采集器没有已授权 session 仍会保持管理控制通道和心跳可用。不要再通过采集器终端执行交互式登录，也不要在采集器运行时覆盖或移走 `account.session`；首次绑定和重新授权都应在管理端完成。

若修改了 Telegram 应用凭据、手机号、导入归属用户、共享控制令牌或宿主目录映射，重新创建相关容器后再检查管理端状态：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml up -d --force-recreate api worker telegram-ingestor
```

## 在管理端绑定采集账号

使用管理员账号登录管理端，打开 `/admin/telegram`。页面顶部会分别显示采集账号、采集器心跳、来源数量和来源异常数。

### 首次绑定：手机号验证码

1. 确认采集器状态可用后，点击“手机号授权”。
2. 输入当前管理员密码，勾选二次确认并继续。该操作会暂停新的 Telegram 读取，现有下载和导入会受控排空。
3. 在配置手机号对应的 Telegram 客户端中取得验证码，并在页面提交。
4. 如果 Telegram 要求两步验证，在页面提交 Telegram 二次验证密码。
5. 授权成功后，临时 session 会经过账号身份校验后原子替换为正式 session，采集器自动恢复读取。

首次账号绑定必须使用手机号验证码；二维码不能代替此步骤。验证码、Telegram 二次验证密码、二维码 token 和当前管理员密码只在当前请求中使用，不会写入数据库、日志、浏览器持久化存储或队列。

系统同一时间只允许一个 Telegram 授权会话。所有管理员都能看到脱敏进度，但只有发起者能提交验证码、二次验证密码或取消授权。取消、失败或过期会丢弃临时 session，并保留既有正式 session；首次绑定失败则恢复未配置状态。

### 已绑定账号：二维码重新授权

1. 点击“二维码重新授权”，完成当前管理员密码和页面二次确认。
2. 使用手机 Telegram 扫描页面中的短期二维码，并在 Telegram 客户端确认登录。
3. 页面自动轮询授权结果；成功后采集器使用新 session 恢复读取。

二维码重新授权只能用于已通过手机号绑定的采集账号。新登录得到的 Telegram 用户 ID 必须与既有绑定一致；账号不一致时会拒绝替换 session，旧 session 保持可用。

## 添加 Telegram 来源

在“Telegram 来源”区域输入 `@username`、Telegram 链接、私密邀请链接或数字 chat ID，然后点击“预解析”。页面会显示 Telegram 返回的标题、用户名、类型和访问状态；只有群组、超级群组和频道可以添加。

确认预览信息后点击“确认添加”。系统重新校验访问权限，以 canonical chat ID 去重并保存展示元数据，然后自动开始从最新消息向旧消息回填历史视频，完成后继续监听实时消息。

私密邀请链接可能使采集账号加入群组或频道。因此私密链接在预解析和最终确认前都会要求当前管理员密码和页面二次确认。原始邀请 token 只用于当前交互，不会保存到数据库、审计、日志或浏览器持久化存储。

## 日常采集控制

来源列表展示每个来源的同步状态、历史游标、完成时间、处理数量和最近的脱敏错误。常用操作如下：

| 操作 | 行为 |
| --- | --- |
| 暂停 | 停止后续采集，保留来源、历史游标、媒体状态和已导入视频。 |
| 恢复 | 从当前历史游标继续同步，不重复扫描已完成的历史页。 |
| 恢复失败来源 | 仅用于来源处于异常状态时；清除来源级错误并重新调度，不重置历史游标。先修复权限、网络或磁盘原因。 |
| 重新回填 | 清零历史游标和回填完成时间，从最新消息重新扫描历史记录。已有消息、Telegram document ID 和文件哈希仍会去重。 |

进度中的“已完成”包含已导入和重复内容，“处理中”包含已发现、已入队、下载中和导入中状态。来源进入 `pending`、`backfilling`、`live` 是正常生命周期；`paused` 是管理员主动停用，`error` 需要检查页面错误摘要和采集器日志。

“管理审计”记录管理员执行的授权和来源操作。审计为只读追加记录，保留操作者、动作、目标、结果和安全摘要，不包含验证码、密码、二维码 token、API hash、私密邀请 token 或完整堆栈。

不要通过直接 SQL 修改 `telegram_media`、手动把失败媒体改为队列状态、删除来源行或清空 Redis 来处理日常故障。PostgreSQL 中的来源、游标和媒体状态才是权威；Redis 只负责唤醒可幂等处理的任务。当前页面不提供删除来源、单条媒体跳过或手工编辑媒体状态的操作。

## 故障排查

| 现象 | 处理方式 |
| --- | --- |
| 页面显示采集器未运行或无响应 | 检查 `docker compose ... ps` 和 `logs -f telegram-ingestor`，确认 PostgreSQL、Redis、`TELEGRAM_CONTROL_TOKEN` 和内部控制地址配置一致。不要暴露控制端口。 |
| 首次绑定或重新授权失败 | 在页面查看脱敏错误。验证码失效、账号不一致或授权过期时取消或等待当前会话结束后重新开始；不要手工替换 session。 |
| 无法预解析来源 | 确认采集账号已授权、采集器为运行中、账号拥有该群组或频道的访问权限，并检查 Telegram 的加入审批或邀请状态。 |
| 来源处于异常 | 先修复页面错误摘要所指向的网络、权限、磁盘或 Telegram 限流原因，再使用“恢复失败来源”。 |
| 遇到 FloodWait | 保持等待 Telegram 要求的退避时间。不要反复重启、重复登录或盲目提高 `TELEGRAM_DOWNLOAD_CONCURRENCY`。 |

初次部署建议保持下载并发 `1`、采集总任务数 `2`。只有在磁盘、带宽和账号限流均稳定后才逐步调整，且总任务数不得小于下载并发。

## 回滚 migration

回滚会删除 Telegram 管理元数据、来源和媒体采集状态，但不会删除 `videos` 表记录和已有视频文件。仅在明确需要回退到不含 Telegram 功能的服务端版本时执行；先备份数据库、停止三个应用服务，并依次回滚 `0039`、`0038`：

```bash
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml stop telegram-ingestor worker api
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  pg_dump -U video -d video_server -Fc > video_server_before_telegram_rollback.dump
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 < migrations/0039_telegram_management.down.sql
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 < migrations/0038_telegram_video_ingestion.down.sql
docker compose -f docker-compose.yml -f deploy/docker-compose.telegram.yml exec -T postgres \
  psql -U video -d video_server -v ON_ERROR_STOP=1 -c "DELETE FROM schema_migrations WHERE version IN ('0039_telegram_management.up.sql', '0038_telegram_video_ingestion.up.sql');"
```

确认回滚后再切回不依赖这些表的服务端提交。不要只回滚 migration 而继续运行含 Telegram 仓储查询的新二进制。

## 安全与合规

- session 等同个人账号长期凭据。限制目录权限、备份访问者和日志读取权限；怀疑泄露时应在 Telegram 客户端撤销会话，然后通过管理端重新授权。
- 只添加已获授权的来源，并遵守群组规则、内容版权、隐私和当地法律。技术上的可访问不代表拥有保存或传播权限。
- 采集器控制令牌、Telegram API hash、验证码、Telegram 二次验证密码和私密邀请链接都属于敏感信息，不应进入截图、工单、日志或聊天记录。
